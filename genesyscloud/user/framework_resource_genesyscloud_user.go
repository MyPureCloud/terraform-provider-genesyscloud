package user

import (
	"context"
	"fmt"
	"log"
	"time"

	pfdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Ensure UserFrameworkResource satisfies various resource interfaces
var (
	_ resource.Resource                = &UserFrameworkResource{}
	_ resource.ResourceWithConfigure   = &UserFrameworkResource{}
	_ resource.ResourceWithImportState = &UserFrameworkResource{}
)

// UserFrameworkResource is the main resource struct that manages Genesys Cloud user lifecycle operations.
// It holds the API client configuration needed to communicate with the Genesys Cloud platform.
// This struct implements the Terraform Framework resource interfaces (Resource, ResourceWithConfigure, ResourceWithImportState).
type UserFrameworkResource struct {
	clientConfig *platformclientv2.Configuration
}

// UserFrameworkResourceModel represents the complete Terraform state for a Genesys Cloud user.
// This model maps directly to the Terraform configuration schema and is used to marshal/unmarshal
// state data during CRUD operations. All fields use Terraform Framework types (types.String, types.Bool, etc.)
// to properly handle null, unknown, and known values in the Terraform lifecycle.
type UserFrameworkResourceModel struct {
	Id                    types.String `tfsdk:"id"`
	Email                 types.String `tfsdk:"email"`
	Name                  types.String `tfsdk:"name"`
	Password              types.String `tfsdk:"password"`
	State                 types.String `tfsdk:"state"`
	DivisionId            types.String `tfsdk:"division_id"`
	Department            types.String `tfsdk:"department"`
	Title                 types.String `tfsdk:"title"`
	Manager               types.String `tfsdk:"manager"`
	AcdAutoAnswer         types.Bool   `tfsdk:"acd_auto_answer"`
	RoutingSkills         types.Set    `tfsdk:"routing_skills"`
	RoutingLanguages      types.Set    `tfsdk:"routing_languages"`
	Locations             types.Set    `tfsdk:"locations"`
	Addresses             types.List   `tfsdk:"addresses"`
	ProfileSkills         types.Set    `tfsdk:"profile_skills"`
	Certifications        types.Set    `tfsdk:"certifications"`
	EmployerInfo          types.List   `tfsdk:"employer_info"`
	RoutingUtilization    types.List   `tfsdk:"routing_utilization"`
	VoicemailUserpolicies types.List   `tfsdk:"voicemail_userpolicies"`
}

// VoicemailUserpoliciesModel represents voicemail configuration settings for a user.
// These settings control voicemail behavior including timeout duration and email notification preferences.
type VoicemailUserpoliciesModel struct {
	AlertTimeoutSeconds    types.Int64 `tfsdk:"alert_timeout_seconds"`
	SendEmailNotifications types.Bool  `tfsdk:"send_email_notifications"`
}

// RoutingUtilizationModel defines the capacity settings for different communication channels.
// This controls how many concurrent interactions a user can handle across various media types
// (call, callback, message, email, chat) and supports label-based routing utilization.
type RoutingUtilizationModel struct {
	Call              types.List `tfsdk:"call"`
	Callback          types.List `tfsdk:"callback"`
	Message           types.List `tfsdk:"message"`
	Email             types.List `tfsdk:"email"`
	Chat              types.List `tfsdk:"chat"`
	LabelUtilizations types.List `tfsdk:"label_utilizations"`
}

// MediaUtilizationModel defines capacity settings for a specific media type (call, chat, email, etc.).
// MaximumCapacity sets the concurrent interaction limit, IncludeNonAcd determines if non-ACD interactions
// count toward capacity, and InterruptibleMediaTypes specifies which media can be interrupted by this type.
type MediaUtilizationModel struct {
	MaximumCapacity         types.Int64 `tfsdk:"maximum_capacity"`
	IncludeNonAcd           types.Bool  `tfsdk:"include_non_acd"`
	InterruptibleMediaTypes types.Set   `tfsdk:"interruptible_media_types"`
}

// LabelUtilizationModel defines capacity settings for label-based routing.
// Labels allow more granular control over routing capacity beyond standard media types,
// enabling custom routing rules based on interaction characteristics.
type LabelUtilizationModel struct {
	LabelId              types.String `tfsdk:"label_id"`
	MaximumCapacity      types.Int64  `tfsdk:"maximum_capacity"`
	InterruptingLabelIds types.Set    `tfsdk:"interrupting_label_ids"`
}

// agentUtilizationWithLabels mirrors the SDK response structure for agent utilization API responses.
// This internal struct is used for JSON unmarshaling when reading utilization data from the API.
// It includes both standard media utilization and label-based utilization configurations.
type agentUtilizationWithLabels struct {
	Utilization       map[string]mediaUtilization `json:"utilization"`
	LabelUtilizations map[string]labelUtilization `json:"labelUtilizations"`
	Level             string                      `json:"level"`
}

// NewUserFrameworkResource is a factory function that creates a new instance of the user resource.
// This function is called by the Terraform provider during initialization to register the resource type.
// It follows the Terraform Framework pattern for resource instantiation.
func NewUserFrameworkResource() resource.Resource {
	return &UserFrameworkResource{}
}

// Metadata sets the resource type name that will be used in Terraform configurations.
// The TypeName is constructed by combining the provider name with "_user" to create
// the full resource identifier (e.g., "genesyscloud_user").
// This method is automatically called by Terraform during provider schema discovery.
func (r *UserFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the complete resource schema including all attributes, their types, and validation rules.
// This tells Terraform what configuration options are available and how to validate user input.
// The schema is automatically invoked during provider initialization and is used for plan/apply operations.
func (r *UserFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = UserResourceSchema()
}

// Configure receives the provider's configured API client and stores it in the resource instance.
// This method establishes the connection between provider-level configuration (API credentials, region, etc.)
// and resource-level operations. It's called automatically by Terraform after provider configuration
// is complete but before any CRUD operations are executed.
func (r *UserFrameworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured
	if req.ProviderData == nil {
		return
	}

	providerMeta, ok := req.ProviderData.(*provider.ProviderMeta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider.ProviderMeta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.clientConfig = providerMeta.ClientConfig
}

// Create implements the resource creation lifecycle for Genesys Cloud users.
// This method handles three scenarios:
// 1. Creating a brand new user via the API
// 2. Restoring a previously deleted user (Genesys Cloud soft-deletes users)
// 3. Handling conflict errors when a user email already exists in deleted state
//
// The creation process involves multiple steps:
// - Check if user exists in deleted state (to avoid conflicts)
// - Create new user OR restore deleted user
// - Update attributes that can only be set via PATCH (manager, locations, etc.)
// - Apply routing skills, languages, and utilization settings
// - Read final state to ensure consistency
//
// This method is called by Terraform when 'terraform apply' detects a new resource in the configuration.
func (r *UserFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserFrameworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := GetUserProxy(r.clientConfig)
	email := plan.Email.ValueString()
	divisionID := plan.DivisionId.ValueString()

	// Build addresses from plan
	addresses, addressDiags := buildSdkAddresses(plan.Addresses)
	resp.Diagnostics.Append(addressDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for a deleted user before creating
	id, diagErr := getDeletedUserId(email, proxy)
	if diagErr.HasError() {
		for _, sdkDiag := range diagErr {
			resp.Diagnostics.AddError(sdkDiag.Summary(), sdkDiag.Detail())
		}
		return
	}

	if id != nil {
		// Found deleted user - restore and configure
		plan.Id = types.StringValue(*id)
		log.Printf("Found deleted user %s, restoring", email)

		restoreDeletedUser(ctx, &plan, proxy, r.clientConfig, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		// Set state and RETURN (matches SDKv2 pattern)
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
		return
	}

	// No deleted user - create new one
	createUser := platformclientv2.Createuser{
		Name:       platformclientv2.String(plan.Name.ValueString()),
		State:      platformclientv2.String(plan.State.ValueString()),
		Title:      platformclientv2.String(plan.Title.ValueString()),
		Department: platformclientv2.String(plan.Department.ValueString()),
		Email:      &email,
		Addresses:  addresses,
	}

	if divisionID != "" {
		createUser.DivisionId = &divisionID
	}

	log.Printf("Creating user %s", email)
	log.Printf("[INV] CREATE payload.Addresses=%s", invMustJSON(createUser.Addresses))
	if createUser.Addresses != nil {
		invDumpSDKPhones("CREATE payload (SDK)", *createUser.Addresses)
	}

	userResponse, proxyPostResponse, postErr := proxy.createUser(ctx, &createUser)
	if postErr != nil {
		if proxyPostResponse != nil && proxyPostResponse.Error != nil && (*proxyPostResponse.Error).Code == "general.conflict" {
			// CREATE failed with conflict - check for deleted user
			log.Printf("[INV] CREATE failed with conflict, checking for deleted user %s", email)

			id, diagErr := getDeletedUserId(email, proxy)
			if diagErr.HasError() {
				resp.Diagnostics.Append(diagErr...)
				return
			}

			if id != nil {
				// Found deleted user after conflict - restore and configure
				plan.Id = types.StringValue(*id)
				log.Printf("[INV] RESTORE (conflict): Found deleted user, restoring")

				restoreDeletedUser(ctx, &plan, proxy, r.clientConfig, &resp.Diagnostics)
				if resp.Diagnostics.HasError() {
					return
				}

				// Set state and RETURN (matches SDKv2 pattern)
				log.Printf("[INV] RESTORE (conflict): Complete for user %s", email)
				resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
				return
			}
		}

		// Other error
		resp.Diagnostics.Append(util.BuildFrameworkAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create user %s error: %s", email, postErr), proxyPostResponse)...)
		return
	}

	// User created successfully
	plan.Id = types.StringValue(*userResponse.Id)

	// Log CREATE echo (server response)
	log.Printf("[INV] CREATE echo (server user.Addresses)=%s", invMustJSON(userResponse.Addresses))
	if userResponse.Addresses != nil {
		for i, c := range *userResponse.Addresses {
			log.Printf("[INV]   echo[%d] %s (display=%q)", i, invSDKPhoneIdentity(c), invStr(c.Display))
		}
		invDumpSDKPhones("CREATE response (SDK)", *userResponse.Addresses)
	}

	// Set attributes that can only be modified in a patch
	// (These can't be set in createUser, must use PATCH)
	if hasChanges(&plan, "manager", "locations", "acd_auto_answer", "profile_skills", "certifications", "employer_info") {
		log.Printf("Updating additional attributes for user %s", email)

		additionalAttrsUpdate := &platformclientv2.Updateuser{
			Manager:        platformclientv2.String(plan.Manager.ValueString()),
			AcdAutoAnswer:  platformclientv2.Bool(plan.AcdAutoAnswer.ValueBool()),
			Locations:      buildSdkLocations(ctx, plan.Locations),
			Certifications: buildSdkCertifications(ctx, plan.Certifications),
			EmployerInfo:   buildSdkEmployerInfo(ctx, plan.EmployerInfo),
			Version:        userResponse.Version,
		}

		_, proxyPatchResponse, patchErr := proxy.patchUserWithState(ctx, *userResponse.Id, additionalAttrsUpdate)
		if patchErr != nil {
			resp.Diagnostics.Append(util.BuildFrameworkAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update user %s error: %s", plan.Id.ValueString(), patchErr), proxyPatchResponse)...)
			return
		}
	}

	// Apply skills, languages, utilization
	frameworkDiags := executeAllUpdates(ctx, &plan, proxy, r.clientConfig, false)
	if frameworkDiags.HasError() {
		resp.Diagnostics.Append(frameworkDiags...)
		return
	}

	log.Printf("Created user %s %s", email, *userResponse.Id)

	// Read back the created user to populate state with server-generated values and ensure consistency.
	// This matches SDKv2 pattern where Create calls Read to get the final state after creation.
	// All attributes including addresses, skills, voicemail policies, and routing utilization are fetched.
	readUser(ctx, &plan, proxy, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Log final state
	log.Printf("[INV] FINAL STATE effective addresses: %s", invMustJSON(plan.Addresses))
	if idents := invPhoneIdentitiesFromFrameworkAddresses(plan.Addresses); len(idents) > 0 {
		log.Printf("[INV] FINAL STATE phone identities: %v", idents)
	}
	log.Printf("[INV] FINAL STATE routing_utilization: %s", invMustJSON(plan.RoutingUtilization))
	log.Printf("[INV] FINAL STATE voicemail_userpolicies: %s", invMustJSON(plan.VoicemailUserpolicies))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data from Genesys Cloud.
// This method is critical for detecting drift between the Terraform configuration and actual infrastructure.
// It implements retry logic to handle Genesys Cloud's eventual consistency model where changes
// may not be immediately visible after creation or updates.
//
// The read operation:
// - Fetches current user state from the API
// - Handles 404 errors as retryable (user may not be indexed yet)
// - Updates Terraform state with current values
// - Logs extensive debugging information for troubleshooting
//
// This method is called during:
// - 'terraform plan' to detect drift
// - 'terraform refresh' to sync state
// - Before update/delete operations to get current state
func (r *UserFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserFrameworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := GetUserProxy(r.clientConfig)

	// Fetch current user state from API to detect drift and refresh Terraform state.
	// This is the primary read operation called during plan, refresh, and before update/delete.
	// Uses the same helper as Create/Update to ensure consistent state representation across all operations.
	readUser(ctx, &state, proxy, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		//TODO
		// If there's a 404 error, the resource was deleted outside Terraform
		// The caller (Read method) should handle removing from state if needed
		return
	}

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update modifies an existing user's attributes to match the desired Terraform configuration.
// This method compares the planned state (desired) with the current state (actual) to determine
// what changes need to be applied. It handles several special cases:
//
// 1. State changes (active/inactive) must be updated separately from other attributes
// 2. Some attributes require specific API endpoints or ordering
// 3. Routing utilization and skills require separate update calls
//
// The update process:
// - Compare plan vs state to identify changes
// - Handle state transitions separately (if state attribute changed)
// - Update basic attributes (name, email, department, etc.) via PATCH
// - Update routing skills, languages, and utilization via dedicated endpoints
// - Read final state to verify changes and detect any drift
//
// This method is called by Terraform when 'terraform apply' detects changes to an existing resource.
func (r *UserFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserFrameworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state UserFrameworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := GetUserProxy(r.clientConfig)

	// Log identity comparison between plan and state
	planIDs := invPhoneIdentitiesFromFrameworkAddresses(plan.Addresses)
	stateIDs := invPhoneIdentitiesFromFrameworkAddresses(state.Addresses)
	log.Printf("[INV] PLAN phone identities:  %v", planIDs)
	log.Printf("[INV] STATE phone identities: %v", stateIDs)

	// INVESTIGATION: Log plan state before update
	log.Printf("DEBUG: [INVESTIGATION] Before update - plan.RoutingUtilization.IsNull(): %v", plan.RoutingUtilization.IsNull())
	log.Printf("[INV] BEFORE UPDATE plan.Addresses isNull=%v isUnknown=%v", plan.Addresses.IsNull(), plan.Addresses.IsUnknown())
	log.Printf("[INV] BEFORE UPDATE plan.Addresses=%s", invMustJSON(plan.Addresses))

	// Call the helper function that contains all update logic
	// This matches SDKv2 pattern where Update calls updateUser()
	updateUser(ctx, &plan, proxy, r.clientConfig, &resp.Diagnostics, &state)
	if resp.Diagnostics.HasError() {
		return
	}

	// INVESTIGATION: Log plan state after update
	log.Printf("DEBUG: [INVESTIGATION] After update - plan.RoutingUtilization.IsNull(): %v", plan.RoutingUtilization.IsNull())
	log.Printf("[INV] AFTER UPDATE plan.Addresses isNull=%v isUnknown=%v", plan.Addresses.IsNull(), plan.Addresses.IsUnknown())
	log.Printf("[INV] AFTER UPDATE plan.Addresses=%s", invMustJSON(plan.Addresses))

	// Log post-update identities
	if idents := invPhoneIdentitiesFromFrameworkAddresses(plan.Addresses); len(idents) > 0 {
		log.Printf("[INV] POST-UPDATE phone identities: %v", idents)
	}

	// Log cty.Value structure for deep comparison
	if rawPlanAttrs, err := plan.Addresses.ToTerraformValue(ctx); err == nil {
		log.Printf("[INV] cty.Plan.Addresses raw=%s", invMustJSON(rawPlanAttrs))
	}
	if rawStateAttrs, err := state.Addresses.ToTerraformValue(ctx); err == nil {
		log.Printf("[INV] cty.State.Addresses raw=%s", invMustJSON(rawStateAttrs))
	}

	// Log final state before Set
	log.Printf("[INV] FINAL STATE effective addresses: %s", invMustJSON(plan.Addresses))
	if idents := invPhoneIdentitiesFromFrameworkAddresses(plan.Addresses); len(idents) > 0 {
		log.Printf("[INV] FINAL STATE phone identities: %v", idents)
	}
	log.Printf("[INV] FINAL STATE routing_utilization: %s", invMustJSON(plan.RoutingUtilization))
	log.Printf("[INV] FINAL STATE voicemail_userpolicies: %s", invMustJSON(plan.VoicemailUserpolicies))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete implements the resource deletion lifecycle for Genesys Cloud users.
// Note: This performs a SOFT DELETE, not a permanent deletion. Genesys Cloud moves users
// to a "deleted" state where they can be restored later. This is why the Create method
// checks for deleted users before creating new ones.
//
// The deletion process:
// - Calls the delete API endpoint with retry logic (handles version mismatch errors)
// - Verifies the user reaches deleted state via search API
// - Uses timeout pattern to wait for eventual consistency
//
// This method is called when:
// - 'terraform destroy' is executed
// - A resource is removed from the Terraform configuration and apply is run
func (r *UserFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserFrameworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := GetUserProxy(r.clientConfig)
	email := state.Email.ValueString()

	log.Printf("Deleting user %s", email)

	err := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		_, proxyDelResponse, err := proxy.deleteUser(ctx, state.Id.ValueString())
		if err != nil {
			return proxyDelResponse, util.BuildFrameworkAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete user %s error: %s", state.Id.ValueString(), err), proxyDelResponse)
		}
		log.Printf("Deleted user %s", email)
		return nil, nil
	})
	if err != nil {
		resp.Diagnostics.Append(err...)
		return
	}

	// Verify user in deleted state and search index has been updated
	verifyDiags := util.PFWithRetries(ctx, 3*time.Minute, func() (bool, error) {
		id, getErr := getDeletedUserId(email, proxy)
		if getErr.HasError() {
			// Non-retryable error
			return false, fmt.Errorf("error searching for deleted user %s: %v", email, getErr)
		}
		if id == nil {
			// Retryable - user not yet in deleted state
			return true, fmt.Errorf("user %s not yet in deleted state", email)
		}
		// Success - user is deleted
		return false, nil
	})

	if verifyDiags.HasError() {
		resp.Diagnostics.Append(verifyDiags...)
		return
	}
}

// GetAllUsers retrieves all users from Genesys Cloud for export operations.
// This function is used by the Terraform exporter tool to generate Terraform configurations
// from existing infrastructure. It returns a map of user IDs to metadata where the email
// is used as the Terraform block label.
//
// This function uses Framework diagnostics and is called by:
// - Resource exporter tools
// - Bulk import utilities
// - State migration scripts
//
// Returns: Map of user IDs to ResourceMeta (containing email as block label)
func GetAllUsers(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, pfdiag.Diagnostics) {
	proxy := GetUserProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	users, resp, err := proxy.GetAllUser(ctx)
	if err != nil {
		return nil, util.BuildFrameworkAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of users error: %s", err), resp)
	}

	for _, user := range *users {
		if user.Id != nil && user.Email != nil {
			hashedUniqueFields, err := util.QuickHashFields(user.Name, user.Department, user.PrimaryContactInfo, user.Addresses)
			if err != nil {
				return nil, util.BuildFrameworkDiagnosticError(ResourceType, fmt.Sprintf("Failed to hash user fields: %s", err), err)
			}
			resources[*user.Id] = &resourceExporter.ResourceMeta{BlockLabel: *user.Email, BlockHash: hashedUniqueFields}
		}
	}

	return resources, nil
}

// GetAllUsersSDK retrieves all users for export operations using SDK v2 diagnostics.
// This is functionally identical to GetAllUsers() but returns SDK-style diagnostics
// instead of Framework diagnostics for backward compatibility with legacy code paths.
//
// This function is used by:
// - Legacy export code that expects SDK diagnostic types
// - Migration utilities transitioning from SDK to Framework
//
// Returns: Map of user IDs to ResourceMeta with SDK-style error handling
func GetAllUsersSDK(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, sdkdiag.Diagnostics) {
	proxy := GetUserProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	users, resp, err := proxy.GetAllUser(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get users: %s", err), resp)
	}

	for _, user := range *users {
		if user.Id != nil && user.Email != nil {
			resources[*user.Id] = &resourceExporter.ResourceMeta{BlockLabel: *user.Email}
		}
	}

	return resources, nil
}

// ImportState enables importing existing Genesys Cloud users into Terraform state.
// This allows you to manage pre-existing users with Terraform without having to recreate them.
// The import uses a passthrough pattern where the user ID provided on the command line
// becomes the resource ID in state.
//
// Usage: terraform import genesyscloud_user.example <user_id>
//
// After import, run 'terraform plan' to see what configuration needs to be added
// to match the imported user's current state.
func (r *UserFrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
