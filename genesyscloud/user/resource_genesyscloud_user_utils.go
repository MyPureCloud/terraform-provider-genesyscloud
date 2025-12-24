package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/mail"
	"os"
	"sort"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	chunksProcess "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/chunks"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	pfdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// SDK v2 diagnostics imported as 'diag' in this file

var (
	// Map of SDK media type name to schema media type name
	utilizationMediaTypes = map[string]string{
		"call":     "call",
		"callback": "callback",
		"chat":     "chat",
		"email":    "email",
		"message":  "message",
	}
)

type mediaUtilization struct {
	MaximumCapacity         int32    `json:"maximumCapacity"`
	InterruptableMediaTypes []string `json:"interruptableMediaTypes"`
	IncludeNonAcd           bool     `json:"includeNonAcd"`
}

type labelUtilization struct {
	MaximumCapacity      int32    `json:"maximumCapacity"`
	InterruptingLabelIds []string `json:"interruptingLabelIds"`
}

// Helper function to get media utilization attribute types
func getMediaUtilizationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"maximum_capacity":          types.Int64Type,
		"include_non_acd":           types.BoolType,
		"interruptible_media_types": types.SetType{ElemType: types.StringType},
	}
}

// Element type definitions - define once, reuse everywhere
func routingSkillsElementType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"skill_id":    types.StringType,
			"proficiency": types.Float64Type,
		},
	}
}

func routingLanguagesElementType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"language_id": types.StringType,
			"proficiency": types.Int64Type,
		},
	}
}

func locationsElementType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"location_id": types.StringType,
			"notes":       types.StringType,
		},
	}
}

func employerInfoElementType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"official_name": types.StringType,
			"employee_id":   types.StringType,
			"employee_type": types.StringType,
			"date_hire":     types.StringType,
		},
	}
}

func voicemailUserpoliciesElementType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"alert_timeout_seconds":    types.Int64Type,
			"send_email_notifications": types.BoolType,
		},
	}
}

type FrameworkRetryWrapper struct {
	resourceType string
}

func newFrameworkRetryWrapper(resourceType string) *FrameworkRetryWrapper {
	return &FrameworkRetryWrapper{
		resourceType: resourceType,
	}
}

// Helper function to get label utilization attribute types
func getLabelUtilizationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"label_id":               types.StringType,
		"maximum_capacity":       types.Int64Type,
		"interrupting_label_ids": types.SetType{ElemType: types.StringType},
	}
}

// readUser reads the user data from the API and populates the model with all attributes.
// This helper is called from Create, Read, and Update to ensure consistent state handling across all CRUD operations.
// Plugin Framework requires this helper pattern because Create/Read/Update methods receive different request/response types
// (CreateRequest/CreateResponse vs ReadRequest/ReadResponse), preventing direct method calls between CRUD operations.
// It fetches user details, voicemail policies, routing utilization, and flattens all complex attributes into Framework types.

func readUser(ctx context.Context, model *UserFrameworkResourceModel, proxy *userProxy, diagnostics *pfdiag.Diagnostics, isImport ...bool) {
	log.Printf("Reading user %s", model.Id.ValueString())

	//TODO
	// Consistency checker removed – not required in Plugin Framework.
	// PF automatically ensures plan/state alignment and drift handling through validators,
	// plan modifiers, and read-after-write patterns.
	// The SDKv2 consistency checker depended on internal ResourceData diffs and reflection,
	// which are unsupported and unnecessary in PF.
	// In the Terraform Plugin Framework, schema attribute validators, plan modifiers,
	// and typed Plan/State models replace the need for an SDKv2-style consistency checker.
	// See PF docs: Validation (https://developer.hashicorp.com/terraform/plugin/framework/validation)
	// and Handling Data – Plan Modifiers (https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/list-nested)
	// for supporting details.
	// cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceUser(), constants.ConsistencyChecks(), ResourceType)

	// Determine if this is an import operation
	importMode := len(isImport) > 0 && isImport[0]

	retryDiags := util.PFWithRetriesForRead(ctx, func() (bool, error) {

		// Fetch user from API
		currentUser, proxyResponse, getUserErr := proxy.getUserById(ctx, model.Id.ValueString(), []string{
			// Expands
			"skills",
			"languages",
			"locations",
			"profileSkills",
			"certifications",
			"employerInfo",
		}, "")

		if getUserErr != nil {
			if util.IsStatus404(proxyResponse) {
				log.Printf("While calling getUserById received 404 error")
				return true, fmt.Errorf("API Error: 404")
			}
			log.Printf("While calling getUserById received NonRetryableError error")
			return false, fmt.Errorf("Failed to read user %s | error: %s", model.Id.ValueString(), getUserErr)
		}

		// Set basic user attributes
		setBasicUserAttributes(model, currentUser)
		// Set manager
		setManagerAttribute(model, currentUser)

		if b, err := json.Marshal(currentUser.Addresses); err == nil {
			log.Printf("[INV][SDKv2] ReadUser API.Addresses=%s", string(b))
		}

		var addressDiags pfdiag.Diagnostics
		model.Addresses, addressDiags = flattenUserAddresses(ctx, currentUser.Addresses, proxy)
		if addressDiags.HasError() {
			return false, fmt.Errorf("Failed to flatten addresses: %v", addressDiags)
		}

		// Handle managed attributes with consistent pattern
		handleManagedRoutingSkills(model, currentUser, &addressDiags)
		handleManagedRoutingLanguages(model, currentUser, &addressDiags)
		handleManagedLocations(model, currentUser, &addressDiags)

		// Flatten profile skills and certifications (always managed)
		model.ProfileSkills = flattenUserData(currentUser.ProfileSkills)
		model.Certifications = flattenUserData(currentUser.Certifications)

		// Handle employer info
		handleManagedEmployerInfo(model, currentUser, &addressDiags)

		// Get and handle voicemail userpolicies
		if !handleVoicemailUserpolicies(ctx, model, proxy, &addressDiags, importMode) {
			// Function returned false = error occurred
			// Check if it's retryable (you'd need to inspect tempDiags or modify the function)
			// For now, treat as retryable (matching SDKv2 behavior)
			return true, fmt.Errorf("Failed to read voicemail userpolicies")
		}

		// Get routing utilization
		if !handleRoutingUtilization(ctx, model, proxy, &addressDiags) {
			// Function returned false = error occurred
			// SDKv2 treated this as NonRetryableError
			return false, fmt.Errorf("Failed to read routing utilization")
		}

		log.Printf("Read user %s %s", model.Id.ValueString(), *currentUser.Email)
		return false, nil
	})

	if retryDiags.HasError() {
		*diagnostics = append(*diagnostics, retryDiags...)
		return
	}

	if len(retryDiags) > 0 {
		for _, d := range retryDiags {
			if d.Severity() == pfdiag.SeverityWarning {
				*diagnostics = append(*diagnostics, retryDiags...)
				return
			}
		}
	}

	// Log final state
	logFinalState(model)
}

// Helper function to set basic user attributes
func setBasicUserAttributes(model *UserFrameworkResourceModel, currentUser *platformclientv2.User) {
	resourcedata.SetPFNillableValueString(&model.Name, currentUser.Name)
	resourcedata.SetPFNillableValueString(&model.Email, currentUser.Email)
	resourcedata.SetPFNillableValueString(&model.DivisionId, currentUser.Division.Id)
	resourcedata.SetPFNillableValueString(&model.State, currentUser.State)
	resourcedata.SetPFNillableValueString(&model.Department, currentUser.Department)
	resourcedata.SetPFNillableValueString(&model.Title, currentUser.Title)
	resourcedata.SetPFNillableValueBool(&model.AcdAutoAnswer, currentUser.AcdAutoAnswer)
}

// Helper function to set manager attribute
func setManagerAttribute(model *UserFrameworkResourceModel, currentUser *platformclientv2.User) {
	model.Manager = types.StringNull()
	if currentUser.Manager != nil {
		if manager := *currentUser.Manager; manager != nil && manager.Id != nil {
			model.Manager = types.StringValue(*manager.Id)
		}
	}
}

// Helper function to handle routing skills
func handleManagedRoutingSkills(model *UserFrameworkResourceModel, currentUser *platformclientv2.User, diagnostics *pfdiag.Diagnostics) {
	isManaged := !model.RoutingSkills.IsNull() && !model.RoutingSkills.IsUnknown()

	if isManaged {
		var skillsDiags pfdiag.Diagnostics
		model.RoutingSkills, skillsDiags = flattenUserSkills(currentUser.Skills)
		*diagnostics = append(*diagnostics, skillsDiags...)
	} else {
		model.RoutingSkills = types.SetNull(routingSkillsElementType())
	}
}

// Helper function to handle routing languages
func handleManagedRoutingLanguages(model *UserFrameworkResourceModel, currentUser *platformclientv2.User, diagnostics *pfdiag.Diagnostics) {
	isManaged := !model.RoutingLanguages.IsNull() && !model.RoutingLanguages.IsUnknown()

	if isManaged {
		var languagesDiags pfdiag.Diagnostics
		model.RoutingLanguages, languagesDiags = flattenUserLanguages(currentUser.Languages)
		*diagnostics = append(*diagnostics, languagesDiags...)
	} else {
		model.RoutingLanguages = types.SetNull(routingLanguagesElementType())
	}
}

// Helper function to handle locations
func handleManagedLocations(model *UserFrameworkResourceModel, currentUser *platformclientv2.User, diagnostics *pfdiag.Diagnostics) {
	isManaged := !model.Locations.IsNull() && !model.Locations.IsUnknown()

	if isManaged {
		var locationsDiags pfdiag.Diagnostics
		model.Locations, locationsDiags = flattenUserLocations(currentUser.Locations)
		*diagnostics = append(*diagnostics, locationsDiags...)
	} else {
		model.Locations = types.SetNull(locationsElementType())
		log.Printf("[INV][PF] readUser(): locations unmanaged, keeping null in state")
	}
}

// Helper function to handle employer info
func handleManagedEmployerInfo(model *UserFrameworkResourceModel, currentUser *platformclientv2.User, diagnostics *pfdiag.Diagnostics) {
	isManaged := !model.EmployerInfo.IsNull() && !model.EmployerInfo.IsUnknown()

	if isManaged {
		var employerInfoDiags pfdiag.Diagnostics
		model.EmployerInfo, employerInfoDiags = flattenUserEmployerInfo(currentUser.EmployerInfo)
		*diagnostics = append(*diagnostics, employerInfoDiags...)
	} else {
		model.EmployerInfo = types.ListNull(employerInfoElementType())
		log.Printf("[INV][PF] readUser(): employer_info unmanaged, keeping null in state")
	}
}

// Helper function to handle voicemail userpolicies - returns false if should abort
func handleVoicemailUserpolicies(ctx context.Context, model *UserFrameworkResourceModel, proxy *userProxy, diagnostics *pfdiag.Diagnostics, isImport ...bool) bool {
	currentVoicemailUserpolicies, apiResp, voicemailErr := proxy.getVoicemailUserpoliciesById(ctx, model.Id.ValueString())

	if voicemailErr != nil {
		*diagnostics = append(*diagnostics, util.BuildFrameworkAPIDiagnosticError(
			ResourceType,
			fmt.Sprintf("Failed to read voicemail userpolicies %s error: %s", model.Id.ValueString(), voicemailErr),
			apiResp,
		)...)
		log.Printf("Error while reading getVoicemailUserpoliciesById in User %s", model.Id.ValueString())
		return false
	}

	// Determine if this is import mode
	importMode := len(isImport) > 0 && isImport[0]
	isManaged := !model.VoicemailUserpolicies.IsNull() && !model.VoicemailUserpolicies.IsUnknown()

	// During import, always populate from API. Otherwise, only if managed.
	if importMode || isManaged {
		var voicemailDiags pfdiag.Diagnostics
		model.VoicemailUserpolicies, voicemailDiags = flattenVoicemailUserpolicies(currentVoicemailUserpolicies)
		*diagnostics = append(*diagnostics, voicemailDiags...)
	} else {
		model.VoicemailUserpolicies = types.ListNull(voicemailUserpoliciesElementType())
		log.Printf("[INV][PF] readUser(): voicemail_userpolicies unmanaged, keeping null in state")
	}

	return true
}

// Helper function to handle routing utilization - returns false if should abort
func handleRoutingUtilization(ctx context.Context, model *UserFrameworkResourceModel, proxy *userProxy, diagnostics *pfdiag.Diagnostics) bool {
	apiResponse, diagErr := readUserRoutingUtilization(ctx, model, proxy)

	if diagErr.HasError() {
		*diagnostics = append(*diagnostics, diagErr...)

		if util.IsStatus404(apiResponse) {
			log.Printf("User %s not found (404 from routing utilization)", model.Id.ValueString())
			return false
		}

		*diagnostics = append(*diagnostics, util.BuildFrameworkAPIDiagnosticError(
			ResourceType,
			fmt.Sprintf("Failed to read routing utilization %s", model.Id.ValueString()),
			apiResponse,
		)...)
		log.Printf("Error while reading readUserRoutingUtilization in User %s", model.Id.ValueString())
		return false
	}

	return true
}

// Helper function to log final state
func logFinalState(model *UserFrameworkResourceModel) {
	log.Printf("[INV] FINAL STATE effective addresses: %s", invMustJSON(model.Addresses))

	if idents := invPhoneIdentitiesFromFrameworkAddresses(model.Addresses); len(idents) > 0 {
		log.Printf("[INV] FINAL STATE phone identities: %v", idents)
	}

	log.Printf("[INV] FINAL STATE routing_utilization: %s", invMustJSON(model.RoutingUtilization))
	log.Printf("[INV] FINAL STATE voicemail_userpolicies: %s", invMustJSON(model.VoicemailUserpolicies))
}

// updateUser applies the full Terraform configuration to an existing user and updates all modifiable attributes.
// This helper is called from Create, Update, and restore operations to ensure consistent update handling across all CRUD operations.
// Plugin Framework requires this helper pattern because Create/Update methods receive different request/response types
// (CreateRequest/CreateResponse vs UpdateRequest/UpdateResponse), preventing direct method calls between CRUD operations.
// It builds SDK request payloads, applies state changes separately, updates core attributes, and applies skills/languages/utilization.
func updateUser(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, clientConfig *platformclientv2.Configuration, diagnostics *pfdiag.Diagnostics, state ...*UserFrameworkResourceModel) {
	// Build addresses from plan
	addresses, addressDiags := buildSdkAddresses(ctx, plan.Addresses)
	*diagnostics = append(*diagnostics, addressDiags...)
	if diagnostics.HasError() {
		return
	}

	email := plan.Email.ValueString()
	log.Printf("Updating user %s", email)

	// If state changes, it is the only modifiable field, so it must be updated separately.
	// Note: During restore (Create -> restoreDeletedUser -> updateUser), the state parameter
	// is not passed (currentState = nil), so this state update is skipped. State is already
	// updated in restoreDeletedUser's restore PATCH. This differs from SDK v2 which performs
	// a redundant state update here (d.HasChange("state") compares "" vs config = true).
	var currentState *UserFrameworkResourceModel
	if len(state) > 0 {
		currentState = state[0]
	}

	if currentState != nil && !plan.State.Equal(currentState.State) {
		log.Printf("Updating state for user %s", email)
		updateUserRequestBody := platformclientv2.Updateuser{
			State: platformclientv2.String(plan.State.ValueString()),
		}
		diagErr := executeUpdateUser(ctx, plan, proxy, updateUserRequestBody)
		if diagErr.HasError() {
			*diagnostics = append(*diagnostics, diagErr...)
			return
		}
	}

	// Update all other attributes
	updateUserRequestBody := platformclientv2.Updateuser{
		Name:           platformclientv2.String(plan.Name.ValueString()),
		Department:     platformclientv2.String(plan.Department.ValueString()),
		Title:          platformclientv2.String(plan.Title.ValueString()),
		Manager:        platformclientv2.String(plan.Manager.ValueString()),
		AcdAutoAnswer:  platformclientv2.Bool(plan.AcdAutoAnswer.ValueBool()),
		Email:          &email,
		Addresses:      addresses,
		Locations:      buildSdkLocations(ctx, plan.Locations),
		Certifications: buildSdkCertifications(ctx, plan.Certifications),
		ProfileSkills:  buildSdkProfileSkills(ctx, plan.ProfileSkills),
		EmployerInfo:   buildSdkEmployerInfo(ctx, plan.EmployerInfo),
	}

	log.Printf("[INV] UPDATE payload.Addresses=%s", invMustJSON(updateUserRequestBody.Addresses))
	if updateUserRequestBody.Addresses != nil {
		invDumpSDKPhones("UPDATE payload (SDK)", *updateUserRequestBody.Addresses)
	}

	// PATCH core user attributes (name, email, addresses, locations, etc.)
	// Includes retry logic with version checking for concurrent modification handling
	diagErr := executeUpdateUser(ctx, plan, proxy, updateUserRequestBody)
	if diagErr.HasError() {
		*diagnostics = append(*diagnostics, diagErr...)
		return
	}

	// Apply updates requiring separate API endpoints: division, skills, languages,
	// routing utilization, voicemail policies, and password.
	// currentState is used for change detection (nil during restore path).
	frameworkDiags := executeAllUpdates(ctx, plan, proxy, clientConfig, true, currentState)
	if frameworkDiags.HasError() {
		*diagnostics = append(*diagnostics, frameworkDiags...)
		return
	}

	log.Printf("Finished updating user %s", email)

	// Read back final state
	readUser(ctx, plan, proxy, diagnostics)
}

func readUserRoutingUtilization(ctx context.Context, state *UserFrameworkResourceModel, proxy *userProxy) (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics
	log.Printf("Getting user utilization")
	// Define reusable type definitions
	mediaUtilObjType := types.ObjectType{AttrTypes: getMediaUtilizationAttrTypes()}
	labelUtilObjType := types.ObjectType{AttrTypes: getLabelUtilizationAttrTypes()}

	elemType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"call":               types.ListType{ElemType: mediaUtilObjType},
		"callback":           types.ListType{ElemType: mediaUtilObjType},
		"message":            types.ListType{ElemType: mediaUtilObjType},
		"email":              types.ListType{ElemType: mediaUtilObjType},
		"chat":               types.ListType{ElemType: mediaUtilObjType},
		"label_utilizations": types.ListType{ElemType: labelUtilObjType},
	}}

	// Make API call
	apiClient := &proxy.routingApi.Configuration.APIClient
	path := fmt.Sprintf("%s/api/v2/routing/users/%s/utilization",
		proxy.routingApi.Configuration.BasePath, state.Id.ValueString())

	response, err := apiClient.CallAPI(path, "GET", nil, buildHeaderParams(proxy.routingApi),
		nil, nil, "", nil, "")
	if err != nil {
		// Return the response so caller can check for 404
		return response, util.BuildFrameworkAPIDiagnosticError(ResourceType,
			fmt.Sprintf("Failed to read routing utilization for user %s error: %s", state.Id.ValueString(), err), response)
	}

	// Unmarshal response
	agentUtilization := &agentUtilizationWithLabels{}
	if err = json.Unmarshal(response.RawBody, &agentUtilization); err != nil {
		log.Printf("[WARN] failed to unmarshal json: %s", err.Error())
		diagnostics.AddError("JSON Unmarshal Error",
			fmt.Sprintf("Failed to unmarshal routing utilization: %s", err.Error()))
		return response, diagnostics
	}

	if agentUtilization == nil {
		state.RoutingUtilization = types.ListNull(elemType)
		log.Printf("[INV][PF] readUserRoutingUtilization(): agentUtilization is nil, setting to null")
		return response, diagnostics
	}

	if agentUtilization.Level == "Organization" {
		// If the settings are org-wide, set to empty to indicate no settings on the user
		emptyList, diags := types.ListValue(elemType, []attr.Value{})
		diagnostics.Append(diags...)
		state.RoutingUtilization = emptyList
		log.Printf("[INV][PF] readUserRoutingUtilization(): Level=Organization, setting to empty list")
		return response, diagnostics
	}

	//Look at existing plan/state to know which media types are actually configured
	var existingConfigs []RoutingUtilizationModel
	var existing RoutingUtilizationModel
	hasExisting := false

	if !state.RoutingUtilization.IsNull() && !state.RoutingUtilization.IsUnknown() {
		diags := state.RoutingUtilization.ElementsAs(ctx, &existingConfigs, false)
		diagnostics.Append(diags...)
		if len(existingConfigs) > 0 {
			existing = existingConfigs[0]
			hasExisting = true
		}
	}

	isConfiguredMedia := func(list types.List) bool {
		// If list is null/unknown, user did not configure this media type
		if list.IsNull() || list.IsUnknown() {
			return false
		}
		// We don’t need to check length here – with MaxItems=1 and required fields,
		// a non-null/non-unknown list corresponds to “configured”.
		return true
	}

	// Build the settings object
	allSettingsAttrs := map[string]attr.Value{
		"call":               types.ListNull(mediaUtilObjType),
		"callback":           types.ListNull(mediaUtilObjType),
		"message":            types.ListNull(mediaUtilObjType),
		"email":              types.ListNull(mediaUtilObjType),
		"chat":               types.ListNull(mediaUtilObjType),
		"label_utilizations": types.ListNull(labelUtilObjType),
	}

	// Flatten media utilization settings
	if agentUtilization.Utilization != nil {
		for sdkType, schemaType := range getUtilizationMediaTypes() {
			if mediaSettings, ok := agentUtilization.Utilization[sdkType]; ok {
				// 🔹 NEW: Only populate media blocks that are configured in plan/state
				if hasExisting {
					switch schemaType {
					case "call":
						if !isConfiguredMedia(existing.Call) {
							continue
						}
					case "callback":
						if !isConfiguredMedia(existing.Callback) {
							continue
						}
					case "message":
						if !isConfiguredMedia(existing.Message) {
							continue
						}
					case "email":
						if !isConfiguredMedia(existing.Email) {
							continue
						}
					case "chat":
						if !isConfiguredMedia(existing.Chat) {
							continue
						}
					}
				}

				flattenedMedia, diags := flattenUtilizationSetting(mediaSettings)
				diagnostics.Append(diags...)
				allSettingsAttrs[schemaType] = flattenedMedia
			}
		}
	}

	// Flatten label utilizations with filtering based on current state
	if agentUtilization.LabelUtilizations != nil {
		// Get the current state's label_utilizations to preserve order
		var originalLabelUtilizations types.List
		if !state.RoutingUtilization.IsNull() && !state.RoutingUtilization.IsUnknown() {
			var currentUtilConfig []RoutingUtilizationModel
			diags := state.RoutingUtilization.ElementsAs(ctx, &currentUtilConfig, false)
			diagnostics.Append(diags...)
			if len(currentUtilConfig) > 0 {
				originalLabelUtilizations = currentUtilConfig[0].LabelUtilizations
			}
		}

		if !originalLabelUtilizations.IsNull() && !originalLabelUtilizations.IsUnknown() {
			// Filter and flatten based on original order
			filteredLabels, diags := filterAndFlattenLabelUtilizations(ctx,
				agentUtilization.LabelUtilizations, originalLabelUtilizations)
			diagnostics.Append(diags...)
			allSettingsAttrs["label_utilizations"] = filteredLabels
		} else {
			// No original labels, return empty list
			emptyLabels, diags := types.ListValue(labelUtilObjType, []attr.Value{})
			diagnostics.Append(diags...)
			allSettingsAttrs["label_utilizations"] = emptyLabels
		}
	}

	// Create the settings object
	settingsObj, diags := types.ObjectValue(elemType.AttrTypes, allSettingsAttrs)
	diagnostics.Append(diags...)

	// Create the list with one element
	utilizationList, diags := types.ListValue(elemType, []attr.Value{settingsObj})
	diagnostics.Append(diags...)
	state.RoutingUtilization = utilizationList

	log.Printf("[INV][PF] readUserRoutingUtilization(): isNull=%v isUnknown=%v routingUtilization=%s",
		state.RoutingUtilization.IsNull(), state.RoutingUtilization.IsUnknown(), invMustJSON(state.RoutingUtilization))

	return response, diagnostics
}

func fetchExtensionPoolId(ctx context.Context, extNum string, proxy *userProxy) string {
	ext, getErr, err := proxy.getTelephonyExtensionPoolByExtension(ctx, extNum)
	if err != nil {
		if getErr != nil {
			log.Printf("Failed to fetch extension pools. Status: %d. Error: %s", getErr.StatusCode, getErr.ErrorMessage)
			return ""
		}
		log.Printf("Failed to fetch extension pool id for extension %s. Error: %s", extNum, err)
		return ""
	}
	return *ext.Id
}

func flattenUserAddresses(ctx context.Context, addresses *[]platformclientv2.Contact, proxy *userProxy) (types.List, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics

	// Define reusable type definitions
	emailObjType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"address": types.StringType,
		"type":    types.StringType,
	}}

	phoneObjType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"number":            types.StringType,
		"media_type":        types.StringType,
		"type":              types.StringType,
		"extension":         types.StringType,
		"extension_pool_id": types.StringType,
	}}

	addressesObjType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"other_emails":  types.SetType{ElemType: emailObjType},
		"phone_numbers": types.SetType{ElemType: phoneObjType},
	}}

	// Return types.ListNull to signal "attribute not set" vs types.ListValue([]) = "explicitly empty"
	if addresses == nil || len(*addresses) == 0 {
		return types.ListNull(addressesObjType), diagnostics
	}

	// Initialize collections for emails and phone numbers
	emailElements := make([]attr.Value, 0)
	phoneElements := make([]attr.Value, 0)
	utilE164 := util.NewUtilE164Service()

	logContact := func(address *platformclientv2.Contact) {
		log.Printf("[INV][PF] flattenUserAddresses(): media=%q type=%q address=%q extension=%q display=%q",
			valueOrEmpty(address.MediaType), valueOrEmpty(address.VarType),
			valueOrEmpty(address.Address), valueOrEmpty(address.Extension), valueOrEmpty(address.Display))
	}

	// Process each contact from API
	for _, address := range *addresses {
		if address.MediaType == nil {
			log.Printf("[INV][PF] flattenUserAddresses(): media=nil hence not processing the contact")
			continue
		}

		logContact(&address)

		switch *address.MediaType {
		case "SMS", "PHONE":
			// Default media_type to PHONE if API didn't set it, matching schema default
			media := "PHONE"
			if address.MediaType != nil && *address.MediaType != "" {
				media = *address.MediaType
			}

			// Initialize phoneNumber similar to SDKv2, but with explicit attr.Value types.
			// TODO Note: extension_pool_id is ALWAYS null in state to keep it out of Set identity
			// and match SDKv2's hash behavior (which ignored extension_pool_id).
			phoneNumber := map[string]attr.Value{
				"media_type":        types.StringValue(media),
				"number":            types.StringNull(),
				"extension":         types.StringNull(),
				"type":              types.StringNull(),
				"extension_pool_id": types.StringNull(), // <- always null in state -- TODO
			}

			//===========================important==========================================
			// ---- PHONE/SMS address patterns (same logic as SDKv2) ----
			// The API can return phone info in 4 distinct ways:
			// 1. Direct phone number
			// 2. Internal extension mapped to an extension pool
			// 3. Phone number + extension pair
			// 4. Standalone extension (not yet mapped)

			// Case 1: Address contains a plain phone number (no extension)
			if address.Address != nil {
				raw := strings.Trim(*address.Address, "()")
				formatted := utilE164.FormatAsCalculatedE164Number(raw)
				if formatted != "" {
					phoneNumber["number"] = types.StringValue(formatted)
				}
			}

			// Case 2: Extension == Display → true internal extension (extension mapped to pool)
			if address.Extension != nil && address.Display != nil && *address.Extension == *address.Display {
				extensionNum := strings.Trim(*address.Extension, "()")
				if extensionNum != "" {
					phoneNumber["extension"] = types.StringValue(extensionNum)
				}
				// In SDKv2, extension_pool_id was ignored by the Set hash.
				// In PF, including it in state would affect Set identity and
				// cause plan/state mismatches. To keep PF behavior closer to SDKv2,
				// we deliberately DO NOT store extension_pool_id in state here.
				//
				// poolId := fetchExtensionPoolId(ctx, extensionNum, proxy)
				// if poolId != "" {
				//     phoneNumber["extension_pool_id"] = types.StringValue(poolId)
				// }
				_ = fetchExtensionPoolId // keep reference if you still use it elsewhere

				// *** KEY FIX: for internal extensions, treat as extension-only identity ***
				// Even if Address is populated (and Case 1 set a number), we force `number`
				// back to null so that:
				//
				//   plan:  extension="9843", number=null
				//   state: extension="9843", number=null
				//
				// and PF can correlate the Set element correctly on create.
				phoneNumber["number"] = types.StringNull()
			}

			// Case 3: Extension ≠ Display → phone number + extension pair
			if address.Extension != nil && address.Display != nil && *address.Extension != *address.Display {
				ext := strings.Trim(*address.Extension, "()")
				if ext != "" {
					phoneNumber["extension"] = types.StringValue(ext)
				}
				raw := strings.Trim(*address.Display, "()")
				formatted := utilE164.FormatAsCalculatedE164Number(raw)
				if formatted != "" {
					phoneNumber["number"] = types.StringValue(formatted)
				}
			}

			// Case 4: Only Display present → unmapped extension (no pool yet)
			if address.Address == nil && address.Extension == nil && address.Display != nil {
				ext := strings.Trim(*address.Display, "()")
				if ext != "" {
					phoneNumber["extension"] = types.StringValue(ext)
				}
			}

			// --- EXTRA NORMALIZATION STEP ---
			// Some orgs/API responses will store an internal extension like "9843"
			// but also echo it back as a phone number in E.164 form, e.g. "+19843".
			// If we keep both:
			//   plan:  extension="9843", number=null
			//   state: extension="9843", number="+19843"
			// the Set element identity will differ and PF can't correlate them.
			//
			// To keep PF behavior aligned with SDKv2 for "extension-only" configs,
			// if `number` is exactly E.164(extension), we treat it as redundant and
			// clear it back to null so identity is extension-only.
			if extVal, ok := phoneNumber["extension"].(types.String); ok && !extVal.IsNull() && !extVal.IsUnknown() {
				if numVal, ok2 := phoneNumber["number"].(types.String); ok2 && !numVal.IsNull() && !numVal.IsUnknown() {
					extStr := strings.TrimSpace(extVal.ValueString())
					numStr := strings.TrimSpace(numVal.ValueString())

					if extStr != "" && numStr != "" {
						// E.164 representation of the extension using same utilE164 helper
						extAsE164 := utilE164.FormatAsCalculatedE164Number(strings.Trim(extStr, "()"))

						if extAsE164 != "" && extAsE164 == numStr {
							// Treat this as "extension-only" identity
							phoneNumber["number"] = types.StringNull()
							log.Printf("[INV][PF] flattenUserAddresses(): normalized E164 duplicate for extension %q; cleared number", extStr)
						}
					}
				}
			}

			// Type: if absent, default to WORK (matches schema default + SDKv2 intent)
			if address.VarType != nil && *address.VarType != "" {
				phoneNumber["type"] = types.StringValue(*address.VarType)
			} else {
				phoneNumber["type"] = types.StringValue("WORK") // <-- default aligns with schema
			}

			phoneObj, objDiags := types.ObjectValue(phoneObjType.AttrTypes, phoneNumber)
			diagnostics.Append(objDiags...)
			if !objDiags.HasError() {
				phoneElements = append(phoneElements, phoneObj)
			}

		case "EMAIL":
			// Handle EMAIL media type
			email := map[string]attr.Value{
				"type":    types.StringValue("WORK"), // <-- default aligns with schema
				"address": types.StringNull(),
			}

			if address.VarType != nil && *address.VarType != "" {
				email["type"] = types.StringValue(*address.VarType)
			}
			if address.Address != nil {
				email["address"] = types.StringValue(*address.Address)
			}

			emailObj, objDiags := types.ObjectValue(emailObjType.AttrTypes, email)
			diagnostics.Append(objDiags...)
			if !objDiags.HasError() {
				emailElements = append(emailElements, emailObj)
			}

		default:
			log.Printf("Unknown address media type %s", *address.MediaType)
		}
	}

	// Create email set
	emailSet, setDiags := types.SetValue(emailObjType, emailElements)
	diagnostics.Append(setDiags...)

	// Create phone number set
	phoneSet, setDiags := types.SetValue(phoneObjType, phoneElements)
	diagnostics.Append(setDiags...)

	// Log final set sizes before return
	log.Printf("[INV][PF] flattenUserAddresses(): phoneSet size=%d emailSet size=%d",
		len(phoneElements), len(emailElements))

	// Create the addresses object containing both sets
	addressesObj, objDiags := types.ObjectValue(addressesObjType.AttrTypes, map[string]attr.Value{
		"other_emails":  emailSet,
		"phone_numbers": phoneSet,
	})
	diagnostics.Append(objDiags...)

	// Return as a list with one element (matching schema: ListNestedBlock with SizeAtMost(1))
	addressesList, listDiags := types.ListValue(addressesObjType, []attr.Value{addressesObj})
	diagnostics.Append(listDiags...)

	return addressesList, diagnostics
}

func flattenUserSkills(skills *[]platformclientv2.Userroutingskill) (types.Set, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics

	// Return an empty set (not null) when there are no skills
	// SetNull means “unknown/not provided.” Here the API told you “there are none,”
	// so it’s a known empty. Returning null can cause diffs later.
	elemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"skill_id":    types.StringType,
			"proficiency": types.Float64Type,
		},
	}

	if skills == nil || len(*skills) == 0 {
		empty, _ := types.SetValue(elemType, []attr.Value{})
		return empty, diagnostics
	}

	skillElements := make([]attr.Value, 0, len(*skills))

	for _, sdkSkill := range *skills {

		id := ""
		if sdkSkill.Id != nil {
			id = *sdkSkill.Id
		}

		prof := 0.0
		if sdkSkill.Proficiency != nil {
			prof = *sdkSkill.Proficiency
		}

		skillAttrs := map[string]attr.Value{
			"skill_id":    types.StringValue(id),
			"proficiency": types.Float64Value(prof),
		}

		skillObj, objDiags := types.ObjectValue(elemType.AttrTypes, skillAttrs)
		diagnostics.Append(objDiags...)

		if !diagnostics.HasError() {
			skillElements = append(skillElements, skillObj)
		}
	}

	skillSet, setDiags := types.SetValue(elemType, skillElements)
	diagnostics.Append(setDiags...)

	log.Printf("[INV][PF] flattenUserSkills(): isNull=%v isUnknown=%v skillSet=%s",
		skillSet.IsNull(), skillSet.IsUnknown(), invMustJSON(skillSet))

	return skillSet, diagnostics
}

func flattenUserLanguages(languages *[]platformclientv2.Userroutinglanguage) (types.Set, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics

	elemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"language_id": types.StringType,
			"proficiency": types.Int64Type,
		},
	}

	if languages == nil || len(*languages) == 0 {
		empty, _ := types.SetValue(elemType, []attr.Value{})
		return empty, diagnostics
	}

	languageElements := make([]attr.Value, 0, len(*languages))

	for _, sdkLang := range *languages {

		id := ""
		if sdkLang.Id != nil {
			id = *sdkLang.Id
		}

		prof := 0.0
		if sdkLang.Proficiency != nil {
			prof = *sdkLang.Proficiency
		}

		languageAttrs := map[string]attr.Value{
			"language_id": types.StringValue(id),
			"proficiency": types.Int64Value(int64(prof)),
		}

		languageObj, objDiags := types.ObjectValue(elemType.AttrTypes, languageAttrs)
		diagnostics.Append(objDiags...)

		if !diagnostics.HasError() {
			languageElements = append(languageElements, languageObj)
		}
	}

	languageSet, setDiags := types.SetValue(elemType, languageElements)
	diagnostics.Append(setDiags...)

	log.Printf("[INV][PF] flattenUserLanguages(): isNull=%v isUnknown=%v languageSet=%s",
		languageSet.IsNull(), languageSet.IsUnknown(), invMustJSON(languageSet))

	return languageSet, diagnostics
}

func flattenUserLocations(locations *[]platformclientv2.Location) (types.Set, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics

	elemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"location_id": types.StringType,
			"notes":       types.StringType,
		},
	}

	if locations == nil || len(*locations) == 0 {
		empty, _ := types.SetValue(elemType, []attr.Value{})
		return empty, diagnostics
	}

	locationElements := make([]attr.Value, 0, len(*locations))

	for _, sdkLoc := range *locations {
		if sdkLoc.LocationDefinition != nil {

			id := ""
			if sdkLoc.LocationDefinition.Id != nil {
				id = *sdkLoc.LocationDefinition.Id
			}

			notes := ""
			if sdkLoc.Notes != nil {
				notes = *sdkLoc.Notes
			}

			locationAttrs := map[string]attr.Value{
				"location_id": types.StringValue(id),
				"notes":       types.StringValue(notes),
			}

			locationObj, objDiags := types.ObjectValue(elemType.AttrTypes, locationAttrs)
			diagnostics.Append(objDiags...)

			if !diagnostics.HasError() {
				locationElements = append(locationElements, locationObj)
			}
		}
	}

	locationSet, setDiags := types.SetValue(elemType, locationElements)
	diagnostics.Append(setDiags...)

	log.Printf("[INV][PF] flattenUserLocations(): isNull=%v isUnknown=%v locationSet=%s",
		locationSet.IsNull(), locationSet.IsUnknown(), invMustJSON(locationSet))

	return locationSet, diagnostics
}

// flattens ProfileSkills, Certifications
func flattenUserData(userDataSlice *[]string) types.Set {
	elements := make([]attr.Value, 0)

	if userDataSlice != nil {
		for _, item := range *userDataSlice {
			elements = append(elements, types.StringValue(item))
		}
	}

	setVal, _ := types.SetValue(types.StringType, elements)

	log.Printf("[INV][PF] flattenUserData(): isNull=%v isUnknown=%v userData(ProfileSkills or Certifications)=%s",
		setVal.IsNull(), setVal.IsUnknown(), invMustJSON(setVal))

	return setVal
}

func flattenUserEmployerInfo(employerInfo *platformclientv2.Employerinfo) (types.List, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics

	elemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"official_name": types.StringType,
			"employee_id":   types.StringType,
			"employee_type": types.StringType,
			"date_hire":     types.StringType,
		},
	}

	if employerInfo == nil {
		//TODO
		// If you want an empty list (not null), then use types.ListValue():
		// empty, _ := types.ListValue(elemType, []attr.Value{})  // Empty list []
		return types.ListNull(elemType), nil
	}

	empInfoAttrs := map[string]attr.Value{
		"official_name": types.StringNull(),
		"employee_id":   types.StringNull(),
		"employee_type": types.StringNull(),
		"date_hire":     types.StringNull(),
	}

	if employerInfo.OfficialName != nil {
		empInfoAttrs["official_name"] = types.StringValue(*employerInfo.OfficialName)
	}

	if employerInfo.EmployeeId != nil {
		empInfoAttrs["employee_id"] = types.StringValue(*employerInfo.EmployeeId)
	}

	if employerInfo.EmployeeType != nil {
		empInfoAttrs["employee_type"] = types.StringValue(*employerInfo.EmployeeType)
	}

	if employerInfo.DateHire != nil {
		empInfoAttrs["date_hire"] = types.StringValue(*employerInfo.DateHire)
	}

	empInfoObj, diagnostics := types.ObjectValue(elemType.AttrTypes, empInfoAttrs)

	empInfoList, diagnostics := types.ListValue(elemType, []attr.Value{empInfoObj})

	log.Printf("[INV][PF] flattenUserEmployerInfo(): isNull=%v isUnknown=%v empInfoList=%s",
		empInfoList.IsNull(), empInfoList.IsUnknown(), invMustJSON(empInfoList))

	return empInfoList, diagnostics
}

func flattenVoicemailUserpolicies(voicemail *platformclientv2.Voicemailuserpolicy) (types.List, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics

	elemType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"alert_timeout_seconds":    types.Int64Type,
			"send_email_notifications": types.BoolType,
		},
	}

	if voicemail == nil {
		voicemailResult := types.ListNull(elemType)
		return voicemailResult, nil
	}

	voicemailAttrs := map[string]attr.Value{
		"alert_timeout_seconds":    types.Int64Null(),
		"send_email_notifications": types.BoolNull(),
	}

	if voicemail.AlertTimeoutSeconds != nil {
		voicemailAttrs["alert_timeout_seconds"] = types.Int64Value(int64(*voicemail.AlertTimeoutSeconds))
	}
	if voicemail.SendEmailNotifications != nil {
		voicemailAttrs["send_email_notifications"] = types.BoolValue(*voicemail.SendEmailNotifications)
	}

	voicemailObj, diags := types.ObjectValue(elemType.AttrTypes, voicemailAttrs)
	diagnostics.Append(diags...)

	voicemailList, diags := types.ListValue(elemType, []attr.Value{voicemailObj})
	diagnostics.Append(diags...)

	log.Printf("[INV][PF] flattenVoicemailUserpolicies(): isNull=%v isUnknown=%v voicemailList=%s",
		voicemailList.IsNull(), voicemailList.IsUnknown(), invMustJSON(voicemailList))

	return voicemailList, diagnostics
}

func flattenUtilizationSetting(mediaSettings mediaUtilization) (types.List, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics

	elemType := types.ObjectType{AttrTypes: getMediaUtilizationAttrTypes()}

	attrs := map[string]attr.Value{
		"maximum_capacity":          types.Int64Value(int64(mediaSettings.MaximumCapacity)),
		"include_non_acd":           types.BoolValue(mediaSettings.IncludeNonAcd),
		"interruptible_media_types": types.SetNull(types.StringType),
	}

	// Handle interruptible media types
	if len(mediaSettings.InterruptableMediaTypes) > 0 {
		interruptibleSet, diags := types.SetValueFrom(context.Background(), types.StringType,
			mediaSettings.InterruptableMediaTypes)
		diagnostics.Append(diags...)
		attrs["interruptible_media_types"] = interruptibleSet
	}

	obj, diags := types.ObjectValue(elemType.AttrTypes, attrs)
	diagnostics.Append(diags...)

	list, diags := types.ListValue(elemType, []attr.Value{obj})
	diagnostics.Append(diags...)

	return list, diagnostics
}

func filterAndFlattenLabelUtilizations(ctx context.Context, apiLabels map[string]labelUtilization, originalLabels types.List) (types.List, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics

	elemType := types.ObjectType{AttrTypes: getLabelUtilizationAttrTypes()}

	var originalLabelModels []LabelUtilizationModel
	diags := originalLabels.ElementsAs(ctx, &originalLabelModels, false)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return types.ListNull(elemType), diagnostics
	}

	var filteredValues []attr.Value
	for _, originalLabel := range originalLabelModels {
		labelId := originalLabel.LabelId.ValueString()

		if apiLabel, ok := apiLabels[labelId]; ok {
			attrs := map[string]attr.Value{
				"label_id":               types.StringValue(labelId),
				"maximum_capacity":       types.Int64Value(int64(apiLabel.MaximumCapacity)),
				"interrupting_label_ids": types.SetNull(types.StringType),
			}

			if len(apiLabel.InterruptingLabelIds) > 0 {
				interruptingSet, diags := types.SetValueFrom(ctx, types.StringType, apiLabel.InterruptingLabelIds)
				diagnostics.Append(diags...)
				attrs["interrupting_label_ids"] = interruptingSet
			}

			obj, diags := types.ObjectValue(elemType.AttrTypes, attrs)
			diagnostics.Append(diags...)
			filteredValues = append(filteredValues, obj)

			// SDKv2 had delte
			// NO delete() needed - map is not used after this function
		}
	}

	list, diags := types.ListValue(elemType, filteredValues)
	diagnostics.Append(diags...)

	return list, diagnostics
}

func buildSdkAddresses(ctx context.Context, addresses types.List) (*[]platformclientv2.Contact, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics
	sdkAddresses := make([]platformclientv2.Contact, 0)

	// Check if addresses is null or unknown
	if addresses.IsNull() || addresses.IsUnknown() {
		return &sdkAddresses, diagnostics
	}

	// Define the model for addresses block
	type AddressesModel struct {
		OtherEmails  types.Set `tfsdk:"other_emails"`
		PhoneNumbers types.Set `tfsdk:"phone_numbers"`
	}

	// Extract addresses into typed model
	var addressesBlocks []AddressesModel
	diags := addresses.ElementsAs(ctx, &addressesBlocks, false)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return &sdkAddresses, diagnostics
	}

	// Check if we have at least one addresses block
	if len(addressesBlocks) == 0 {
		return &sdkAddresses, diagnostics
	}

	// Get the first (and only) addresses block
	addressBlock := addressesBlocks[0]

	// Build emails
	if !addressBlock.OtherEmails.IsNull() && !addressBlock.OtherEmails.IsUnknown() {
		emailContacts, emailDiags := buildSdkEmails(addressBlock.OtherEmails)
		diagnostics.Append(emailDiags...)
		sdkAddresses = append(sdkAddresses, emailContacts...)
	}

	// Build phone numbers
	if !addressBlock.PhoneNumbers.IsNull() && !addressBlock.PhoneNumbers.IsUnknown() {
		phoneContacts, phoneDiags := buildSdkPhoneNumbers(addressBlock.PhoneNumbers)
		diagnostics.Append(phoneDiags...)
		sdkAddresses = append(sdkAddresses, phoneContacts...)
	}

	return &sdkAddresses, diagnostics
}

func buildSdkEmails(configEmails types.Set) ([]platformclientv2.Contact, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics

	// Check if set is null or unknown
	if configEmails.IsNull() || configEmails.IsUnknown() {
		return []platformclientv2.Contact{}, diagnostics
	}

	// Define the model for other_emails
	type OtherEmailModel struct {
		Address types.String `tfsdk:"address"`
		Type    types.String `tfsdk:"type"`
	}

	// Extract emails into typed model
	var emails []OtherEmailModel
	diags := configEmails.ElementsAs(context.Background(), &emails, false)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return []platformclientv2.Contact{}, diagnostics
	}

	// Pre-allocate result array
	sdkContacts := make([]platformclientv2.Contact, len(emails))

	// Build contacts
	for i, email := range emails {
		emailAddress := email.Address.ValueString()
		emailType := email.Type.ValueString()

		sdkContacts[i] = platformclientv2.Contact{
			Address:   &emailAddress,
			MediaType: &contactTypeEmail,
			VarType:   &emailType,
		}
	}

	return sdkContacts, diagnostics
}

func buildSdkPhoneNumbers(configPhoneNumbers types.Set) ([]platformclientv2.Contact, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics

	// Check if set is null or unknown
	if configPhoneNumbers.IsNull() || configPhoneNumbers.IsUnknown() {
		return []platformclientv2.Contact{}, diagnostics
	}

	// Define the model for phone_numbers
	type PhoneNumberModel struct {
		Number          types.String `tfsdk:"number"`
		MediaType       types.String `tfsdk:"media_type"`
		Type            types.String `tfsdk:"type"`
		Extension       types.String `tfsdk:"extension"`
		ExtensionPoolId types.String `tfsdk:"extension_pool_id"`
	}

	// Extract phone numbers into typed model
	var phoneNumbers []PhoneNumberModel
	diags := configPhoneNumbers.ElementsAs(context.Background(), &phoneNumbers, false)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return []platformclientv2.Contact{}, diagnostics
	}

	// Pre-allocate result array
	sdkContacts := make([]platformclientv2.Contact, len(phoneNumbers))

	// Build contacts
	for i, phone := range phoneNumbers {
		// Required fields (have defaults in schema)
		phoneMediaType := phone.MediaType.ValueString()
		phoneType := phone.Type.ValueString()

		contact := platformclientv2.Contact{
			MediaType: &phoneMediaType,
			VarType:   &phoneType,
		}

		// Optional field: number
		if !phone.Number.IsNull() && !phone.Number.IsUnknown() {
			phoneNum := phone.Number.ValueString()
			if phoneNum != "" {
				contact.Address = &phoneNum
			}
		}

		// Optional field: extension
		if !phone.Extension.IsNull() && !phone.Extension.IsUnknown() {
			phoneExt := phone.Extension.ValueString()
			if phoneExt != "" {
				contact.Extension = &phoneExt
			}
		}

		sdkContacts[i] = contact
	}

	return sdkContacts, diagnostics
}

func getDeletedUserId(email string, proxy *userProxy) (*string, pfdiag.Diagnostics) {
	var diagnostics pfdiag.Diagnostics

	exactType := "EXACT"
	results, resp, getErr := proxy.userApi.PostUsersSearch(platformclientv2.Usersearchrequest{
		Query: &[]platformclientv2.Usersearchcriteria{
			{
				Fields:  &[]string{"email"},
				Value:   &email,
				VarType: &exactType,
			},
			{
				Fields:  &[]string{"state"},
				Values:  &[]string{"deleted"},
				VarType: &exactType,
			},
		},
	})

	if getErr != nil {
		return nil, util.BuildFrameworkAPIDiagnosticError(ResourceType,
			fmt.Sprintf("Failed to search for user %s error: %s", email, getErr), resp)
	}

	if results.Results != nil && len(*results.Results) > 0 {
		// User found
		return (*results.Results)[0].Id, diagnostics
	}

	return nil, diagnostics
}

func restoreDeletedUser(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, clientConfig *platformclientv2.Configuration, diagnostics *pfdiag.Diagnostics) {
	email := plan.Email.ValueString()
	state := plan.State.ValueString()

	log.Printf("Restoring deleted user %s", email)

	err := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
		// Get current user (with version)
		currentUser, proxyResponse, err := proxy.getUserById(ctx, plan.Id.ValueString(), nil, "deleted")
		if err != nil {
			return nil, util.BuildFrameworkAPIDiagnosticError(ResourceType,
				fmt.Sprintf("Failed to read user %s error: %s", plan.Id.ValueString(), err), proxyResponse)
		}

		// Log current addresses before restore PATCH
		log.Printf("[INV] RESTORE payload for user %s", email)
		log.Printf("[INV] RESTORE payload.State=%s", state)
		log.Printf("[INV] RESTORE current.Addresses=%s", invMustJSON(currentUser.Addresses))
		if currentUser.Addresses != nil {
			invDumpSDKPhones("RESTORE current addresses (before PATCH)", *currentUser.Addresses)
		}

		restoredUser, proxyPatchResponse, patchErr := proxy.patchUserWithState(ctx, plan.Id.ValueString(),
			&platformclientv2.Updateuser{
				State:   &state,
				Version: currentUser.Version,
			})
		if patchErr != nil {
			return proxyPatchResponse, util.BuildFrameworkAPIDiagnosticError(ResourceType,
				fmt.Sprintf("Failed to restore deleted user %s | Error: %s.", email, patchErr), proxyPatchResponse)
		}

		// Log PATCH response
		if proxyPatchResponse != nil {
			log.Printf("[INV] RESTORE PATCH status: %v", proxyPatchResponse.StatusCode)
		}
		if restoredUser != nil {
			log.Printf("[INV] RESTORE PATCH response.Addresses=%s", invMustJSON(restoredUser.Addresses))
			if restoredUser.Addresses != nil {
				invDumpSDKPhones("RESTORE PATCH response (SDK)", *restoredUser.Addresses)
			}
		}

		// Apply full configuration (equivalent to SDKv2's updateUser call)
		// Pass nil for state parameter to updateUser() since this is a restore
		// (no previous state to compare)
		updateUser(ctx, plan, proxy, clientConfig, diagnostics)
		//          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
		//          Matches SDKv2 pattern: restoreDeletedUser calls updateUser

		if diagnostics.HasError() {
			return nil, *diagnostics
		}

		log.Printf("[INV] RESTORE complete for user %s", email)
		log.Printf("[INV] RESTORE final plan.Addresses=%s", invMustJSON(plan.Addresses))

		return nil, nil
	})

	if err != nil {
		*diagnostics = append(*diagnostics, err...)
	}
}

func executeAllUpdates(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, sdkConfig *platformclientv2.Configuration, updateObjectDivision bool, state ...*UserFrameworkResourceModel) pfdiag.Diagnostics {
	var diagnostics pfdiag.Diagnostics

	var currentState *UserFrameworkResourceModel
	if len(state) > 0 {
		currentState = state[0]
	}

	if updateObjectDivision {
		diagErr := util.UpdateObjectDivisionPF(ctx, plan, currentState, "USER", sdkConfig)
		if diagErr.HasError() {
			diagnostics.Append(diagErr...)
			return diagnostics
		}
	}

	// Update user skills - using standalone function from utils file
	diagErr := updateUserSkills(ctx, plan, currentState, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	// Update user languages - using standalone function from utils file
	diagErr = updateUserLanguages(ctx, plan, currentState, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	// Update profile skills - using standalone function from utils file
	diagErr = updateUserProfileSkills(ctx, plan, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	// Update routing utilization - using standalone function from utils file
	log.Printf("DEBUG: About to update routing utilization for user %s", plan.Id.ValueString())
	diagErr = updateUserRoutingUtilization(ctx, plan, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	// Update voicemail policies - using standalone function from utils file
	diagErr = updateUserVoicemailPolicies(ctx, plan, currentState, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	// Update password - using standalone function from utils file
	diagErr = updatePassword(ctx, plan, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	return diagnostics
}

func updateUserSkills(ctx context.Context, plan *UserFrameworkResourceModel, state *UserFrameworkResourceModel, proxy *userProxy) pfdiag.Diagnostics {
	var diagnostics pfdiag.Diagnostics

	// In Plugin Framework, we don't have d.HasChange() at the helper level
	// Change detection happens at the resource Update() method level
	// This helper is only called when there's a change, so we process unconditionally

	// Was this attribute previously managed?
	wasManaged := state != nil && !state.RoutingSkills.IsNull() && !state.RoutingSkills.IsUnknown()

	// Check if routing_skills is null or unknown (equivalent to SDKv2's skillsConfig == nil)
	isConfigured := !plan.RoutingSkills.IsNull() && !plan.RoutingSkills.IsUnknown()

	// Case 1: never managed and still not configured → do nothing, leave API as-is
	if !wasManaged && !isConfigured {
		// SDKv2 behavior: When skills are nil, return without removing existing skills
		// This matches the nested if structure in SDKv2 where removal only happens inside skillsConfig != nil
		log.Printf("Skills are null/unknown for user %s, skipping skill updates", plan.Id.ValueString())
		return diagnostics
	}

	// Skills are configured or were previously managed - process them (equivalent to skillsConfig != nil)
	log.Printf("Updating skills for user %s (wasManaged=%v, isConfigured=%v)", plan.Email.ValueString(), wasManaged, isConfigured)

	// Build new skills map from Framework types only if configured
	newSkillProfs := make(map[string]float64)
	newSkillIds := []string{}

	if isConfigured {
		skillElements := plan.RoutingSkills.Elements()
		newSkillIds = make([]string, 0, len(skillElements))

		for _, skillElement := range skillElements {
			skillObj, ok := skillElement.(types.Object)
			if !ok {
				continue
			}

			skillAttrs := skillObj.Attributes()
			var skillId string
			var proficiency float64

			if skillIdAttr, exists := skillAttrs["skill_id"]; exists && !skillIdAttr.IsNull() {
				skillId = skillIdAttr.(types.String).ValueString()
			}

			if proficiencyAttr, exists := skillAttrs["proficiency"]; exists && !proficiencyAttr.IsNull() {
				proficiency = proficiencyAttr.(types.Float64).ValueFloat64()
			}

			if skillId == "" {
				continue
			}

			newSkillIds = append(newSkillIds, skillId)
			newSkillProfs[skillId] = proficiency
		}
	}

	// Get current skills from API
	oldSdkSkills, getErr := getUserRoutingSkills(plan.Id.ValueString(), proxy)
	if getErr != nil {
		// Convert SDK diagnostics to Framework diagnostics
		return getErr
	}

	// Build old skills map
	oldSkillIds := make([]string, 0, len(oldSdkSkills))
	oldSkillProfs := make(map[string]float64)
	for _, skill := range oldSdkSkills {
		oldSkillIds = append(oldSkillIds, *skill.Id)
		oldSkillProfs[*skill.Id] = *skill.Proficiency
	}

	// Remove skills that are no longer in configuration
	if len(oldSkillIds) > 0 {
		var skillsToRemove []string

		if !isConfigured {
			// Block removed in config but was managed before → clear everything
			skillsToRemove = oldSkillIds
		} else {
			// Normal diff behavior
			skillsToRemove = lists.SliceDifference(oldSkillIds, newSkillIds)
		}

		for _, skillId := range skillsToRemove {
			diagErr := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
				resp, err := proxy.userApi.DeleteUserRoutingskill(plan.Id.ValueString(), skillId)
				if err != nil {
					return resp, util.BuildFrameworkAPIDiagnosticError(ResourceType,
						fmt.Sprintf("Failed to remove skill from user %s error: %s", plan.Id.ValueString(), err), resp)
				}
				return nil, nil
			})
			if diagErr != nil {
				return diagErr
			}
		}
	}

	// Add or update skills only when attribute is configured in plan
	if isConfigured && len(newSkillIds) > 0 {
		// Skills to add
		skillsToAddOrUpdate := lists.SliceDifference(newSkillIds, oldSkillIds)

		// Check for existing proficiencies to update which can be done with the same API
		for skillID, newProf := range newSkillProfs {
			if oldProf, found := oldSkillProfs[skillID]; found && newProf != oldProf {
				skillsToAddOrUpdate = append(skillsToAddOrUpdate, skillID)
			}
		}

		if len(skillsToAddOrUpdate) > 0 {
			if diagErr := updateUserRoutingSkills(plan.Id.ValueString(), skillsToAddOrUpdate,
				newSkillProfs, proxy); diagErr != nil {
				return diagErr
			}
		}
	}

	return diagnostics
}

func updateUserLanguages(ctx context.Context, plan *UserFrameworkResourceModel, state *UserFrameworkResourceModel, proxy *userProxy) pfdiag.Diagnostics {
	var diagnostics pfdiag.Diagnostics

	// Was this attribute previously managed?
	wasManaged := state != nil && !state.RoutingLanguages.IsNull() && !state.RoutingLanguages.IsUnknown()

	// Check if routing_languages is null or unknown
	isConfigured := !plan.RoutingLanguages.IsNull() && !plan.RoutingLanguages.IsUnknown()

	// Case 1: never managed and still not configured → do nothing, leave API as-is
	if !wasManaged && !isConfigured {
		log.Printf("routing_languages unmanaged for user %s, skipping", plan.Id.ValueString())
		return diagnostics
	}

	// Languages are configured or were previously managed - process them (equivalent to languages != nil block)
	log.Printf("Updating languages for user %s (wasManaged=%v, isConfigured=%v)", plan.Email.ValueString(), wasManaged, isConfigured)

	// Build new languages map from Framework types only if configured
	newLangProfs := make(map[string]int)
	newLangIds := []string{}

	if isConfigured {
		languageElements := plan.RoutingLanguages.Elements()
		newLangIds = make([]string, 0, len(languageElements))

		for _, languageElement := range languageElements {
			languageObj, ok := languageElement.(types.Object)
			if !ok {
				continue
			}

			languageAttrs := languageObj.Attributes()
			var languageId string
			var proficiency int64

			if languageIdAttr, exists := languageAttrs["language_id"]; exists && !languageIdAttr.IsNull() {
				languageId = languageIdAttr.(types.String).ValueString()
			}

			if proficiencyAttr, exists := languageAttrs["proficiency"]; exists && !proficiencyAttr.IsNull() {
				proficiency = proficiencyAttr.(types.Int64).ValueInt64()
			}

			if languageId == "" {
				continue
			}

			newLangIds = append(newLangIds, languageId)
			newLangProfs[languageId] = int(proficiency)
		}
	}

	// Get current languages from API
	oldSdkLangs, getErr := getUserRoutingLanguages(plan.Id.ValueString(), proxy)
	if getErr != nil {
		return getErr
	}

	// Build old languages map
	oldLangIds := make([]string, 0, len(oldSdkLangs))
	oldLangProfs := make(map[string]int)
	for _, lang := range oldSdkLangs {
		oldLangIds = append(oldLangIds, *lang.Id)
		oldLangProfs[*lang.Id] = int(*lang.Proficiency)
	}

	// Remove languages that are no longer in configuration
	if len(oldLangIds) > 0 {
		var langsToRemove []string

		if !isConfigured {
			// Block removed in config but was managed before → clear everything
			langsToRemove = oldLangIds
		} else {
			// Normal diff behavior
			langsToRemove = lists.SliceDifference(oldLangIds, newLangIds)
		}

		for _, langID := range langsToRemove {
			diagErr := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
				resp, err := proxy.userApi.DeleteUserRoutinglanguage(plan.Id.ValueString(), langID)
				if err != nil {
					return resp, util.BuildFrameworkAPIDiagnosticError(ResourceType,
						fmt.Sprintf("Failed to remove language from user %s error: %s", plan.Id.ValueString(), err), resp)
				}
				return nil, nil
			})
			if diagErr != nil {
				return diagErr
			}
		}
	}

	// Add or update languages only when attribute is configured in plan
	if isConfigured && len(newLangIds) > 0 {
		// Languages to add
		langsToAddOrUpdate := lists.SliceDifference(newLangIds, oldLangIds)

		// Check for existing proficiencies to update which can be done with the same API
		for langID, newNum := range newLangProfs {
			if oldNum, found := oldLangProfs[langID]; found && newNum != oldNum {
				langsToAddOrUpdate = append(langsToAddOrUpdate, langID)
			}
		}

		if len(langsToAddOrUpdate) > 0 {
			if diagErr := updateUserRoutingLanguages(plan.Id.ValueString(), langsToAddOrUpdate,
				newLangProfs, proxy); diagErr != nil {
				return diagErr
			}
		}
	}

	log.Printf("Languages updated for user %s", plan.Email.ValueString())
	return diagnostics
}

func updateUserProfileSkills(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) pfdiag.Diagnostics {
	var diagnostics pfdiag.Diagnostics

	if plan.ProfileSkills.IsNull() || plan.ProfileSkills.IsUnknown() {
		return diagnostics
	}

	var profileSkillsSlice []string
	diags := plan.ProfileSkills.ElementsAs(ctx, &profileSkillsSlice, false)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return diagnostics
	}

	diagErr := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
		_, resp, err := proxy.userApi.PutUserProfileskills(plan.Id.ValueString(), profileSkillsSlice)
		if err != nil {
			return resp, util.BuildFrameworkAPIDiagnosticError(ResourceType,
				fmt.Sprintf("Failed to update profile skills for user %s error: %s", plan.Id.ValueString(), err), resp)
		}
		return nil, nil
	})

	if diagErr != nil {
		return diagErr
	}

	return diagnostics
}

func updateUserRoutingUtilization(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) pfdiag.Diagnostics {
	var diagnostics pfdiag.Diagnostics

	// Check if routing_utilization is null or unknown
	if plan.RoutingUtilization.IsNull() || plan.RoutingUtilization.IsUnknown() {
		return diagnostics
	}

	// SDKv2: var err error
	var err error

	log.Printf("Updating user utilization for user %s", plan.Id.ValueString())

	// Extract routing utilization from plan
	var routingUtilizations []RoutingUtilizationModel
	diags := plan.RoutingUtilization.ElementsAs(ctx, &routingUtilizations, false)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return diagnostics
	}

	if len(routingUtilizations) > 0 {
		// Update settings (has configuration)
		utilization := routingUtilizations[0]

		// Check if label utilizations are present
		hasLabelUtilizations := !utilization.LabelUtilizations.IsNull() &&
			!utilization.LabelUtilizations.IsUnknown() &&
			len(utilization.LabelUtilizations.Elements()) > 0

		if hasLabelUtilizations {
			// Path A: Has label utilizations - use direct API call
			apiClient := &proxy.routingApi.Configuration.APIClient

			path := fmt.Sprintf("%s/api/v2/routing/users/%s/utilization",
				proxy.routingApi.Configuration.BasePath, plan.Id.ValueString())

			headerParams := buildHeaderParams(proxy.routingApi)

			requestPayload := make(map[string]interface{})

			requestPayload["utilization"] = buildMediaTypeUtilizations(ctx, utilization)

			var labelUtilizations []LabelUtilizationModel
			labelDiags := utilization.LabelUtilizations.ElementsAs(ctx, &labelUtilizations, false)
			if labelDiags.HasError() {
				diagnostics.Append(labelDiags...)
				return diagnostics
			}
			requestPayload["labelUtilizations"] = buildLabelUtilizationsRequest(ctx, labelUtilizations)

			_, err = apiClient.CallAPI(path, "PUT", requestPayload, headerParams, nil, nil, "", nil, "")
		} else {
			// Path B: No label utilizations - use SDK method
			sdkSettings := make(map[string]platformclientv2.Mediautilization)

			// Build map of media type → Framework field
			mediaTypeFields := map[string]types.List{
				"call":     utilization.Call,
				"callback": utilization.Callback,
				"chat":     utilization.Chat,
				"email":    utilization.Email,
				"message":  utilization.Message,
			}

			for sdkType, schemaType := range getUtilizationMediaTypes() {
				mediaTypeList := mediaTypeFields[schemaType]
				if !mediaTypeList.IsNull() && !mediaTypeList.IsUnknown() && len(mediaTypeList.Elements()) > 0 {
					sdkSettings[sdkType] = buildSdkMediaUtilization(ctx, mediaTypeList)
				}
			}

			_, _, err = proxy.userApi.PutRoutingUserUtilization(plan.Id.ValueString(), platformclientv2.Utilizationrequest{
				Utilization: &sdkSettings,
			})
		}

		if err != nil {
			diagnostics.AddError(
				"Failed to update Routing Utilization",
				fmt.Sprintf("Failed to update Routing Utilization for user %s: %s", plan.Id.ValueString(), err),
			)
			return diagnostics
		}
	} else {
		// Reset to org-wide defaults (empty list)
		resp, err := proxy.userApi.DeleteRoutingUserUtilization(plan.Id.ValueString())
		if err != nil {
			apiDiags := util.BuildAPIDiagnosticError(ResourceType,
				fmt.Sprintf("Failed to delete routing utilization for user %s error: %s", plan.Id.ValueString(), err), resp)
			diagnostics.Append(convertSDKDiagnosticsToFramework(apiDiags)...)
			return diagnostics
		}
	}

	log.Printf("Updated user utilization for user %s", plan.Id.ValueString())

	// SDKv2: return nil
	return diagnostics
}

func updateUserRoutingSkills(userID string, skillsToUpdate []string, skillProfs map[string]float64, proxy *userProxy) pfdiag.Diagnostics {
	// Bulk API restricts skills adds to 50 per call
	const maxBatchSize = 50

	chunkBuild := func(val string) platformclientv2.Userroutingskillpost {
		newProf := skillProfs[val]
		return platformclientv2.Userroutingskillpost{
			Id:          &val,
			Proficiency: &newProf,
		}
	}

	// Generic call to prepare chunks for the Update. Takes in three args
	// 1. skillsToUpdate 2. The Entity prepare func for the update 3. Chunk Size
	chunks := chunksProcess.ChunkItems(skillsToUpdate, chunkBuild, maxBatchSize)
	// Closure to process the chunks

	//TODO
	//later we have to modify the Diagnostics from sdkv2 to PF
	chunkProcessor := func(chunk []platformclientv2.Userroutingskillpost) diag.Diagnostics {
		diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			_, resp, err := proxy.userApi.PatchUserRoutingskillsBulk(userID, chunk)
			if err != nil {
				return resp, util.BuildAPIDiagnosticError(ResourceType,
					fmt.Sprintf("Failed to update skills for user %s error: %s", userID, err), resp)
			}
			return nil, nil
		})
		if diagErr != nil {
			return diagErr
		}
		return nil
	}

	// Generic Function call which takes in the chunks and the processing function
	chunkErr := chunksProcess.ProcessChunks(chunks, chunkProcessor)

	return convertSDKDiagnosticsToFramework(chunkErr)
}

func updateUserRoutingLanguages(userID string, langsToUpdate []string, langProfs map[string]int, proxy *userProxy) pfdiag.Diagnostics {
	// Bulk API restricts language adds to 50 per call
	const maxBatchSize = 50

	chunkBuild := func(val string) platformclientv2.Userroutinglanguagepost {
		newProf := float64(langProfs[val])
		return platformclientv2.Userroutinglanguagepost{
			Id:          &val,
			Proficiency: &newProf,
		}
	}

	// Generic call to prepare chunks for the Update. Takes in three args
	// 1. langsToUpdate 2. The Entity prepare func for the update 3. Chunk Size
	chunks := chunksProcess.ChunkItems(langsToUpdate, chunkBuild, maxBatchSize)
	// Closure to process the chunks

	//TODO
	//later we have to modify the Diagnostics from sdkv2 to PF
	chunkProcessor := func(chunk []platformclientv2.Userroutinglanguagepost) diag.Diagnostics {
		diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			_, resp, err := proxy.userApi.PatchUserRoutinglanguagesBulk(userID, chunk)
			if err != nil {
				return resp, util.BuildAPIDiagnosticError(ResourceType,
					fmt.Sprintf("Failed to update languages for user %s error: %s", userID, err), resp)
			}
			return nil, nil
		})
		if diagErr != nil {
			return diagErr
		}
		return nil
	}

	// Genric Function call which takes in the chunks and the processing function
	chunkErr := chunksProcess.ProcessChunks(chunks, chunkProcessor)

	return convertSDKDiagnosticsToFramework(chunkErr)
}

func getUserRoutingSkills(userID string, proxy *userProxy) ([]platformclientv2.Userroutingskill, pfdiag.Diagnostics) {
	const maxPageSize = 50

	var sdkSkills []platformclientv2.Userroutingskill
	for pageNum := 1; ; pageNum++ {
		skills, resp, err := proxy.userApi.GetUserRoutingskills(userID, maxPageSize, pageNum, "")
		if err != nil {
			return nil, util.BuildFrameworkAPIDiagnosticError(ResourceType,
				fmt.Sprintf("Failed to query skills for user %s error: %s", userID, err), resp)
		}
		if skills == nil || skills.Entities == nil || len(*skills.Entities) == 0 {
			return sdkSkills, nil
		}

		sdkSkills = append(sdkSkills, *skills.Entities...)
	}
}

func getUserRoutingLanguages(userID string, proxy *userProxy) ([]platformclientv2.Userroutinglanguage, pfdiag.Diagnostics) {
	const maxPageSize = 50

	var sdkLanguages []platformclientv2.Userroutinglanguage
	for pageNum := 1; ; pageNum++ {
		langs, resp, err := proxy.userApi.GetUserRoutinglanguages(userID, maxPageSize, pageNum, "")
		if err != nil {
			return nil, util.BuildFrameworkAPIDiagnosticError(ResourceType,
				fmt.Sprintf("Failed to query languages for user %s error: %s", userID, err), resp)
		}
		if langs == nil || langs.Entities == nil || len(*langs.Entities) == 0 {
			return sdkLanguages, nil
		}

		sdkLanguages = append(sdkLanguages, *langs.Entities...)
	}
}

func getUtilizationMediaTypes() map[string]string {
	return utilizationMediaTypes
}

func buildLabelUtilizationsRequest(ctx context.Context, labelUtilizations []LabelUtilizationModel) map[string]labelUtilization {
	request := make(map[string]labelUtilization)

	for _, labelUtil := range labelUtilizations {
		// Extract label_id
		labelId := labelUtil.LabelId.ValueString()

		// Extract maximum_capacity and convert to int32
		maxCapacity := int32(labelUtil.MaximumCapacity.ValueInt64())

		// Extract interrupting_label_ids
		// SDKv2 does no null checking - directly accesses and converts
		var interruptingIds []string
		labelUtil.InterruptingLabelIds.ElementsAs(ctx, &interruptingIds, false)

		// Build labelUtilization struct and add to map
		request[labelId] = labelUtilization{
			MaximumCapacity:      maxCapacity,
			InterruptingLabelIds: interruptingIds,
		}
	}

	return request
}

func buildMediaTypeUtilizations(ctx context.Context, utilization RoutingUtilizationModel) *map[string]platformclientv2.Mediautilization {
	settings := make(map[string]platformclientv2.Mediautilization)

	// Translation layer: Convert typed struct fields to map for dynamic access
	// This bridges the gap between PF's typed models and SDKv2's dynamic maps
	mediaTypeFields := map[string]types.List{
		"call":     utilization.Call,
		"callback": utilization.Callback,
		"chat":     utilization.Chat,
		"email":    utilization.Email,
		"message":  utilization.Message,
	}

	for sdkType, schemaType := range getUtilizationMediaTypes() {
		mediaTypeList := mediaTypeFields[schemaType]

		if !mediaTypeList.IsNull() && !mediaTypeList.IsUnknown() && len(mediaTypeList.Elements()) > 0 {
			settings[sdkType] = buildSdkMediaUtilization(ctx, mediaTypeList)
		}
	}

	return &settings
}

func buildSdkMediaUtilization(ctx context.Context, settings types.List) platformclientv2.Mediautilization {

	// Extract first element (MaxItems=1 in schema)
	var mediaSettings []MediaUtilizationModel
	settings.ElementsAs(ctx, &mediaSettings, false)

	settingsModel := mediaSettings[0]

	// Extract maximum_capacity (required field)
	maxCapacity := int(settingsModel.MaximumCapacity.ValueInt64())

	// Extract include_non_acd (required field)
	includeNonAcd := settingsModel.IncludeNonAcd.ValueBool()

	// Initialize optional field with empty slice pointer
	interruptableMediaTypes := &[]string{}

	// Extract interruptible_media_types (optional field)
	if !settingsModel.InterruptibleMediaTypes.IsNull() && !settingsModel.InterruptibleMediaTypes.IsUnknown() {
		var mediaTypes []string
		settingsModel.InterruptibleMediaTypes.ElementsAs(ctx, &mediaTypes, false)
		interruptableMediaTypes = &mediaTypes
	}

	// Build and return SDK struct
	return platformclientv2.Mediautilization{
		MaximumCapacity:         &maxCapacity,
		IncludeNonAcd:           &includeNonAcd,
		InterruptableMediaTypes: interruptableMediaTypes,
	}
}

func buildHeaderParams(routingAPI *platformclientv2.RoutingApi) map[string]string {
	headerParams := make(map[string]string)

	for key := range routingAPI.Configuration.DefaultHeader {
		headerParams[key] = routingAPI.Configuration.DefaultHeader[key]
	}

	headerParams["Authorization"] = "Bearer " + routingAPI.Configuration.AccessToken
	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	return headerParams
}

func updateUserVoicemailPolicies(ctx context.Context, plan *UserFrameworkResourceModel, state *UserFrameworkResourceModel, proxy *userProxy) pfdiag.Diagnostics {
	var diagnostics pfdiag.Diagnostics

	wasManaged := state != nil &&
		!state.VoicemailUserpolicies.IsNull() &&
		!state.VoicemailUserpolicies.IsUnknown()

	isConfigured := !plan.VoicemailUserpolicies.IsNull() &&
		!plan.VoicemailUserpolicies.IsUnknown()

	// Case 1: Unmanaged both before and now -> do nothing
	if !wasManaged && !isConfigured {
		log.Printf("voicemail_userpolicies unmanaged for user %s, skipping", plan.Id.ValueString())
		return diagnostics
	}

	// Case 2: Was managed, now block removed -> clear remote policies
	if wasManaged && !isConfigured {
		log.Printf("Clearing voicemail_userpolicies for user %s", plan.Id.ValueString())
		emptyReq := platformclientv2.Voicemailuserpolicy{}

		diagErr := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
			_, proxyPutResponse, putErr := proxy.voicemailApi.PatchVoicemailUserpolicy(plan.Id.ValueString(), emptyReq)
			if putErr != nil {
				return proxyPutResponse, util.BuildFrameworkAPIDiagnosticError(ResourceType,
					fmt.Sprintf("Failed to clear voicemail_userpolicies for user %s: %s",
						plan.Id.ValueString(), putErr),
					proxyPutResponse,
				)
			}
			return nil, nil
		})

		if diagErr != nil {
			diagnostics.Append(diagErr...)
		}
		return diagnostics
	}

	// Case 3: isConfigured -> normal update
	log.Printf("Updating voicemail_userpolicies for user %s", plan.Id.ValueString())

	var voicemailPolicies []VoicemailUserpoliciesModel
	diags := plan.VoicemailUserpolicies.ElementsAs(ctx, &voicemailPolicies, false)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return diagnostics
	}

	if len(voicemailPolicies) == 0 {
		// Should not happen with proper schema, but handle defensively
		// Send empty request to clear settings
		log.Printf("voicemail_userpolicies list empty for user %s, clearing", plan.Id.ValueString())
		emptyReq := platformclientv2.Voicemailuserpolicy{}

		diagErr := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
			_, proxyPutResponse, putErr := proxy.voicemailApi.PatchVoicemailUserpolicy(plan.Id.ValueString(), emptyReq)
			if putErr != nil {
				return proxyPutResponse, util.BuildFrameworkAPIDiagnosticError(ResourceType,
					fmt.Sprintf("Failed to clear voicemail_userpolicies for user %s: %s",
						plan.Id.ValueString(), putErr),
					proxyPutResponse,
				)
			}
			return nil, nil
		})

		if diagErr != nil {
			diagnostics.Append(diagErr...)
		}
		return diagnostics
	}

	reqBody := buildVoicemailUserpoliciesRequest(ctx, voicemailPolicies)

	diagErr := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
		_, proxyPutResponse, putErr := proxy.voicemailApi.PatchVoicemailUserpolicy(plan.Id.ValueString(), reqBody)
		if putErr != nil {
			return proxyPutResponse, util.BuildFrameworkAPIDiagnosticError(ResourceType,
				fmt.Sprintf("Failed to update voicemail_userpolicies for user %s: %s",
					plan.Id.ValueString(), putErr),
				proxyPutResponse,
			)
		}
		return nil, nil
	})

	if diagErr != nil {
		diagnostics.Append(diagErr...)
	}

	return diagnostics
}

func buildVoicemailUserpoliciesRequest(ctx context.Context, voicemailPolicies []VoicemailUserpoliciesModel) platformclientv2.Voicemailuserpolicy {
	var request platformclientv2.Voicemailuserpolicy

	// Extract first element (MaxItems=1 in schema)
	if len(voicemailPolicies) > 0 {
		voicemailPolicy := voicemailPolicies[0]

		// Extract send_email_notifications (required field)
		sendEmailNotifications := voicemailPolicy.SendEmailNotifications.ValueBool()

		request = platformclientv2.Voicemailuserpolicy{
			SendEmailNotifications: &sendEmailNotifications,
		}

		// Extract alert_timeout_seconds (optional field, only if > 0)
		if !voicemailPolicy.AlertTimeoutSeconds.IsNull() && !voicemailPolicy.AlertTimeoutSeconds.IsUnknown() {
			alertTimeoutSeconds := int(voicemailPolicy.AlertTimeoutSeconds.ValueInt64())
			if alertTimeoutSeconds > 0 {
				request.AlertTimeoutSeconds = &alertTimeoutSeconds
			}
		}
	}

	return request
}

func updatePassword(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) pfdiag.Diagnostics {

	// In PF, change detection is handled at resource level
	// Check if password is null, unknown, or empty
	if plan.Password.IsNull() || plan.Password.IsUnknown() {
		return pfdiag.Diagnostics{}
	}

	password := plan.Password.ValueString()

	if password == "" {
		// Skip password update if empty
		return pfdiag.Diagnostics{}
	}

	_, err := proxy.updatePassword(ctx, plan.Id.ValueString(), password)
	if err != nil {
		var diagnostics pfdiag.Diagnostics
		diagnostics.AddError(
			"Failed to update password",
			fmt.Sprintf("Failed to update password for user %s: %s", plan.Id.ValueString(), err),
		)
		return diagnostics
	}

	return pfdiag.Diagnostics{}
}

func hasChanges(plan *UserFrameworkResourceModel, attributes ...string) bool {
	// For create operations, we consider all non-null values as changes
	for _, attr := range attributes {
		switch attr {
		case "manager":
			if !plan.Manager.IsNull() && !plan.Manager.IsUnknown() && plan.Manager.ValueString() != "" {
				return true
			}
		case "locations":
			if !plan.Locations.IsNull() && !plan.Locations.IsUnknown() {
				elements := plan.Locations.Elements()
				if len(elements) > 0 {
					return true
				}
			}
		case "acd_auto_answer":
			if !plan.AcdAutoAnswer.IsNull() && !plan.AcdAutoAnswer.IsUnknown() {
				return true
			}
		case "profile_skills":
			if !plan.ProfileSkills.IsNull() && !plan.ProfileSkills.IsUnknown() {
				// Profile skills are specified in plan (even if empty), so there's a change
				return true
			}
		case "certifications":
			if !plan.Certifications.IsNull() && !plan.Certifications.IsUnknown() {
				// Certifications are specified in plan (even if empty), so there's a change
				return true
			}
		case "employer_info":
			if !plan.EmployerInfo.IsNull() && !plan.EmployerInfo.IsUnknown() {
				elements := plan.EmployerInfo.Elements()
				if len(elements) > 0 {
					return true
				}
			}
		}
	}
	return false
}

func getSdkUtilizationTypes() []string {
	types := make([]string, 0, len(utilizationMediaTypes))
	for t := range utilizationMediaTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func buildSdkLocations(ctx context.Context, locations types.Set) *[]platformclientv2.Location {
	// Check if locations is null or unknown
	if locations.IsNull() || locations.IsUnknown() {
		return nil
	}

	sdkLocations := make([]platformclientv2.Location, 0)

	// Extract locations from Framework Set
	locationElements := locations.Elements()

	for _, locElement := range locationElements {
		locObj, ok := locElement.(types.Object)
		if !ok {
			continue
		}

		locAttrs := locObj.Attributes()

		var locID string
		if locIDAttr, exists := locAttrs["location_id"]; exists && !locIDAttr.IsNull() {
			locID = locIDAttr.(types.String).ValueString()
		}

		var locNotes string
		if locNotesAttr, exists := locAttrs["notes"]; exists && !locNotesAttr.IsNull() {
			locNotes = locNotesAttr.(types.String).ValueString()
		}

		sdkLocations = append(sdkLocations, platformclientv2.Location{
			Id:    &locID,
			Notes: &locNotes,
		})
	}

	return &sdkLocations
}

func buildSdkCertifications(ctx context.Context, certifications types.Set) *[]string {
	certs := make([]string, 0)

	// If certifications are specified (even if empty), process them
	if !certifications.IsNull() && !certifications.IsUnknown() {
		for _, certVal := range certifications.Elements() {
			certStr, ok := certVal.(basetypes.StringValue)
			if ok {
				certs = append(certs, certStr.ValueString())
			}
		}
	}
	// Always return a slice (even if empty) to allow clearing certifications
	return &certs
}

func buildSdkProfileSkills(ctx context.Context, profileSkills types.Set) *[]string {
	skills := make([]string, 0)

	// If profile skills are specified (even if empty), process them
	if !profileSkills.IsNull() && !profileSkills.IsUnknown() {
		for _, skillVal := range profileSkills.Elements() {
			skillStr, ok := skillVal.(basetypes.StringValue)
			if ok {
				skills = append(skills, skillStr.ValueString())
			}
		}
	}
	// Always return a slice (even if empty) to allow clearing profile skills
	return &skills
}

func buildSdkEmployerInfo(ctx context.Context, employerInfo types.List) *platformclientv2.Employerinfo {

	// Check if employer_info is null or unknown
	if employerInfo.IsNull() || employerInfo.IsUnknown() {
		return nil
	}

	var sdkInfo platformclientv2.Employerinfo

	// Extract employer info from Framework List
	elements := employerInfo.Elements()
	if len(elements) > 0 {
		// Type assert to Object
		empInfoObj, ok := elements[0].(types.Object)
		if !ok {
			return nil
		}

		empAttrs := empInfoObj.Attributes()

		// Extract official_name (optional, only if non-empty)
		if officialName, exists := empAttrs["official_name"]; exists && !officialName.IsNull() {
			if nameVal, ok := officialName.(types.String); ok {
				offName := nameVal.ValueString()
				if len(offName) > 0 {
					sdkInfo.OfficialName = &offName
				}
			}
		}

		// Extract employee_id (optional, only if non-empty)
		if employeeId, exists := empAttrs["employee_id"]; exists && !employeeId.IsNull() {
			if idVal, ok := employeeId.(types.String); ok {
				empID := idVal.ValueString()
				if len(empID) > 0 {
					// SDKv2: sdkInfo.EmployeeId = &empID
					sdkInfo.EmployeeId = &empID
				}
			}
		}

		// Extract employee_type (optional, only if non-empty)
		if employeeType, exists := empAttrs["employee_type"]; exists && !employeeType.IsNull() {
			if typeVal, ok := employeeType.(types.String); ok {
				empType := typeVal.ValueString()
				if len(empType) > 0 {
					sdkInfo.EmployeeType = &empType
				}
			}
		}

		// Extract date_hire (optional, only if non-empty)
		if dateHire, exists := empAttrs["date_hire"]; exists && !dateHire.IsNull() {
			if dateVal, ok := dateHire.(types.String); ok {
				dateHireStr := dateVal.ValueString()
				if len(dateHireStr) > 0 {
					sdkInfo.DateHire = &dateHireStr
				}
			}
		}
	}

	return &sdkInfo
}

func executeUpdateUser(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, updateUser platformclientv2.Updateuser) pfdiag.Diagnostics {

	// Retry on version mismatch errors
	diagErr := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, pfdiag.Diagnostics) {

		currentUser, proxyResponse, errGet := proxy.getUserById(ctx, plan.Id.ValueString(), nil, "")
		if errGet != nil {
			return proxyResponse, util.BuildFrameworkAPIDiagnosticError(ResourceType,
				fmt.Sprintf("Failed to read user %s error: %s", plan.Id.ValueString(), errGet), proxyResponse)
		}

		updateUser.Version = currentUser.Version

		_, proxyPatchResponse, patchErr := proxy.updateUser(ctx, plan.Id.ValueString(), &updateUser)
		if patchErr != nil {
			return proxyPatchResponse, util.BuildFrameworkAPIDiagnosticError(ResourceType,
				fmt.Sprintf("Failed to update user %s | Error: %s.", plan.Id.ValueString(), patchErr), proxyPatchResponse)
		}

		return proxyPatchResponse, nil
	})

	if diagErr != nil {
		return diagErr
	}

	return pfdiag.Diagnostics{}
}

//------------------------------ the function below needed for extra logging, //TODO Remove them later----------------------------

func emailorNameDisambiguation(searchField string) (string, string) {
	emailField := "email"
	nameField := "name"
	_, err := mail.ParseAddress(searchField)
	if err == nil {
		return searchField, emailField
	}
	return searchField, nameField
}

func convertSDKDiagnosticsToFramework(sdkDiags diag.Diagnostics) pfdiag.Diagnostics {
	var frameworkDiags pfdiag.Diagnostics

	for _, sdkDiag := range sdkDiags {
		switch sdkDiag.Severity {
		case diag.Error:
			frameworkDiags.AddError(sdkDiag.Summary, sdkDiag.Detail)
		case diag.Warning:
			frameworkDiags.AddWarning(sdkDiag.Summary, sdkDiag.Detail)
		default:
			// Default to error for unknown severity levels to maintain SDK behavior
			frameworkDiags.AddError(sdkDiag.Summary, sdkDiag.Detail)
		}
	}

	return frameworkDiags
}

// valueOrEmpty returns the string value of a pointer or "<nil>" if nil
func valueOrEmpty(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

// invMustJSON marshals to JSON for debug; swallow errors in logs (consistent naming with log-changes1.md)
func invMustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<json err: %v>", err)
	}
	return string(b)
}

// invStr returns the string value of a pointer or "<nil>" if nil (consistent naming with log-changes1.md)
func invStr(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

// invDumpSDKPhones prints a compact one-line per phone contact (SDK type)
func invDumpSDKPhones(tag string, phones []platformclientv2.Contact) {
	log.Printf("[INV] %s phones(%d):", tag, len(phones))
	for i, c := range phones {
		log.Printf("[INV]   #%d media=%q type=%q number=%q ext=%q display=%q",
			i, invStr(c.MediaType), invStr(c.VarType), invStr(c.Address), invStr(c.Extension), invStr(c.Display))
	}
}

// invTriString shows Framework tri-state for types.String (Null/Unknown/Value + empty?)
func invTriString(s types.String) string {
	switch {
	case s.IsNull():
		return "NULL"
	case s.IsUnknown():
		return "UNKNOWN"
	default:
		v := s.ValueString()
		if v == "" {
			return `""`
		}
		return v
	}
}

// invSDKPhoneIdentity builds a stable identity string for a phone contact as it appears from the SDK/API
func invSDKPhoneIdentity(c platformclientv2.Contact) string {
	return fmt.Sprintf("media=%s|type=%s|number=%s|ext=%s",
		invStr(c.MediaType), invStr(c.VarType), invStr(c.Address), invStr(c.Extension))
}

// invPhoneIdentitiesFromFrameworkAddresses extracts phone identity strings from a Framework addresses value (plan or state)
func invPhoneIdentitiesFromFrameworkAddresses(addrVal types.List) []string {
	ids := []string{}
	if addrVal.IsNull() || addrVal.IsUnknown() {
		return ids
	}

	addrElements := addrVal.Elements()
	if len(addrElements) == 0 {
		return ids
	}

	for _, addrElem := range addrElements {
		addrObj, ok := addrElem.(types.Object)
		if !ok {
			continue
		}

		phonesAttr, exists := addrObj.Attributes()["phone_numbers"]
		if !exists {
			continue
		}

		phonesSet, ok := phonesAttr.(types.Set)
		if !ok || phonesSet.IsNull() || phonesSet.IsUnknown() {
			continue
		}

		phoneElements := phonesSet.Elements()
		for _, phoneElem := range phoneElements {
			phoneObj, ok := phoneElem.(types.Object)
			if !ok {
				continue
			}

			attrs := phoneObj.Attributes()
			media := attrs["media_type"].(types.String)
			typ := attrs["type"].(types.String)
			num := attrs["number"].(types.String)
			ext := attrs["extension"].(types.String)

			id := fmt.Sprintf("media=%s|type=%s|number=%s|ext=%s",
				invTriString(media), invTriString(typ), invTriString(num), invTriString(ext))
			ids = append(ids, id)
		}
	}
	return ids
}

/*
The code below is used for testing purposes. When the env var is set, the singleton pattern will be in effect for the proxy
instance, which will allow us to mock out certain methods.
(See the comments above GetUserProxy to understand why we avoid the singleton proxy outside of testing)
*/

const userTestsActiveEnvVar string = "TF_GC_USER_TESTS_ACTIVE"

func setUserTestsActiveEnvVar() error {
	return os.Setenv(userTestsActiveEnvVar, "true")
}

func unsetUserTestsActiveEnvVar() error {
	return os.Unsetenv(userTestsActiveEnvVar)
}

func userTestsAreActive() bool {
	_, isSet := os.LookupEnv(userTestsActiveEnvVar)
	return isSet
}
