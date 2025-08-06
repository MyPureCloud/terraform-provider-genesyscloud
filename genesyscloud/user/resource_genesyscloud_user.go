package user

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

type agentUtilizationWithLabels struct {
	Utilization       map[string]mediaUtilization `json:"utilization"`
	LabelUtilizations map[string]labelUtilization `json:"labelUtilizations"`
	Level             string                      `json:"level"`
}

func GetAllUsers(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := GetUserProxy(sdkConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	users, proxyResponse, err := proxy.GetAllUser(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of users error: %s", err), proxyResponse)
	}

	// Add resources to metamap
	for _, user := range *users {
		if user.Id == nil || user.Email == nil {
			continue
		}
		hashedUniqueFields, err := util.QuickHashFields(user.Name, user.Department, user.PrimaryContactInfo, user.Addresses)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		resources[*user.Id] = &resourceExporter.ResourceMeta{BlockLabel: *user.Email, BlockHash: hashedUniqueFields}
	}

	return resources, nil
}

func createUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetUserProxy(sdkConfig)

	email := d.Get("email").(string)
	divisionID := d.Get("division_id").(string)

	addresses, addrErr := buildSdkAddresses(d)
	if addrErr != nil {
		return addrErr
	}

	// Check for a deleted user before creating
	id, _ := getDeletedUserId(email, proxy)
	if id != nil {
		d.SetId(*id)
		return restoreDeletedUser(ctx, d, meta, proxy)
	}

	createUser := platformclientv2.Createuser{
		Name:       platformclientv2.String(d.Get("name").(string)),
		State:      platformclientv2.String(d.Get("state").(string)),
		Title:      platformclientv2.String(d.Get("title").(string)),
		Department: platformclientv2.String(d.Get("department").(string)),
		Email:      &email,
		Addresses:  addresses,
	}

	// Optional attribute that should not be empty strings
	if divisionID != "" {
		createUser.DivisionId = &divisionID
	}

	log.Printf("Creating user %s", email)

	userResponse, proxyPostResponse, postErr := proxy.createUser(ctx, &createUser)
	if postErr != nil {
		if proxyPostResponse != nil && proxyPostResponse.Error != nil && (*proxyPostResponse.Error).Code == "general.conflict" {
			// Check for a deleted user
			id, diagErr := getDeletedUserId(email, proxy)
			if diagErr != nil {
				return diagErr
			}
			if id != nil {
				d.SetId(*id)
				return restoreDeletedUser(ctx, d, meta, proxy)
			}
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create user %s error: %s", email, postErr), proxyPostResponse)
	}

	d.SetId(*userResponse.Id)

	// Set attributes that can only be modified in a patch
	if d.HasChanges("manager", "locations", "acd_auto_answer", "profile_skills", "certifications", "employer_info") {
		log.Printf("Updating additional attributes for user %s", email)

		_, proxyPatchResponse, patchErr := proxy.patchUserWithState(ctx, *userResponse.Id, &platformclientv2.Updateuser{
			Manager:        platformclientv2.String(d.Get("manager").(string)),
			AcdAutoAnswer:  platformclientv2.Bool(d.Get("acd_auto_answer").(bool)),
			Locations:      buildSdkLocations(d),
			Certifications: buildSdkCertifications(d),
			EmployerInfo:   buildSdkEmployerInfo(d),
			Version:        userResponse.Version,
		})

		if patchErr != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update user %s error: %s", d.Id(), patchErr), proxyPatchResponse)
		}
	}

	diagErr := executeAllUpdates(ctx, d, proxy, sdkConfig, false)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Created user %s %s", email, *userResponse.Id)
	return readUser(ctx, d, meta)
}

func readUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetUserProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceUser(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading user %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {

		currentUser, proxyResponse, errGet := proxy.getUserById(ctx, d.Id(), []string{
			// Expands
			"skills",
			"languages",
			"locations",
			"profileSkills",
			"certifications",
			"employerInfo"},
			"")

		if errGet != nil {
			readErr := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read user %s | error: %s", d.Id(), errGet), proxyResponse)
			if util.IsStatus404(proxyResponse) {
				return retry.RetryableError(readErr)
			}
			return retry.NonRetryableError(readErr)
		}

		// Required attributes
		resourcedata.SetNillableValue(d, "name", currentUser.Name)
		resourcedata.SetNillableValue(d, "email", currentUser.Email)
		resourcedata.SetNillableValue(d, "division_id", currentUser.Division.Id)
		resourcedata.SetNillableValue(d, "state", currentUser.State)
		resourcedata.SetNillableValue(d, "department", currentUser.Department)
		resourcedata.SetNillableValue(d, "title", currentUser.Title)
		resourcedata.SetNillableValue(d, "acd_auto_answer", currentUser.AcdAutoAnswer)

		_ = d.Set("manager", nil)
		if currentUser.Manager != nil {
			_ = d.Set("manager", *(*currentUser.Manager).Id)
		}

		_ = d.Set("addresses", flattenUserAddresses(ctx, currentUser.Addresses, proxy))
		_ = d.Set("routing_skills", flattenUserSkills(currentUser.Skills))
		_ = d.Set("routing_languages", flattenUserLanguages(currentUser.Languages))
		_ = d.Set("locations", flattenUserLocations(currentUser.Locations))
		_ = d.Set("profile_skills", flattenUserData(currentUser.ProfileSkills))
		_ = d.Set("certifications", flattenUserData(currentUser.Certifications))
		_ = d.Set("employer_info", flattenUserEmployerInfo(currentUser.EmployerInfo))

		//Get attributes from Voicemail/Userpolicies resource
		currentVoicemailUserpolicies, resp, err := proxy.getVoicemailUserpoliciesById(ctx, d.Id())
		if err != nil {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read voicemail userpolicies %s error: %s", d.Id(), err), resp))
		}

		_ = d.Set("voicemail_userpolicies", flattenVoicemailUserpolicies(d, currentVoicemailUserpolicies))

		if diagErr := readUserRoutingUtilization(d, proxy); diagErr != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		log.Printf("Read user %s %s", d.Id(), *currentUser.Email)

		return cc.CheckState(d)
	})
}

func updateUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetUserProxy(sdkConfig)

	addresses, err := buildSdkAddresses(d)
	if err != nil {
		return err
	}

	email := d.Get("email").(string)

	log.Printf("Updating user %s", email)

	// If state changes, it is the only modifiable field, so it must be updated separately
	if d.HasChange("state") {
		log.Printf("Updating state for user %s", email)
		updateUserRequestBody := platformclientv2.Updateuser{
			State: platformclientv2.String(d.Get("state").(string)),
		}
		diagErr := executeUpdateUser(ctx, d, proxy, updateUserRequestBody)
		if diagErr != nil {
			return diagErr
		}
	}

	updateUserRequestBody := platformclientv2.Updateuser{
		Name:           platformclientv2.String(d.Get("name").(string)),
		Department:     platformclientv2.String(d.Get("department").(string)),
		Title:          platformclientv2.String(d.Get("title").(string)),
		Manager:        platformclientv2.String(d.Get("manager").(string)),
		AcdAutoAnswer:  platformclientv2.Bool(d.Get("acd_auto_answer").(bool)),
		Email:          &email,
		Addresses:      addresses,
		Locations:      buildSdkLocations(d),
		Certifications: buildSdkCertifications(d),
		EmployerInfo:   buildSdkEmployerInfo(d),
	}
	diagErr := executeUpdateUser(ctx, d, proxy, updateUserRequestBody)
	if diagErr != nil {
		return diagErr
	}

	diagErr = executeAllUpdates(ctx, d, proxy, sdkConfig, true)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating user %s", email)
	return readUser(ctx, d, meta)
}

func deleteUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetUserProxy(sdkConfig)

	email := d.Get("email").(string)

	log.Printf("Deleting user %s", email)
	err := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		_, proxyDelResponse, err := proxy.deleteUser(ctx, d.Id())
		if err != nil {
			return proxyDelResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete user %s error: %s", d.Id(), err), proxyDelResponse)
		}
		log.Printf("Deleted user %s", email)
		return nil, nil
	})
	if err != nil {
		return err
	}

	// Verify user in deleted state and search index has been updated
	return util.WithRetries(ctx, 3*time.Minute, func() *retry.RetryError {
		id, err := getDeletedUserId(email, proxy)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("error searching for deleted user %s: %v", email, err))
		}
		if id == nil {
			return retry.RetryableError(fmt.Errorf("user %s not yet in deleted state", email))
		}
		return nil
	})
}
