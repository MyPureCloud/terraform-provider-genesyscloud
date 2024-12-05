package user

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

type agentUtilizationWithLabels struct {
	Utilization       map[string]mediaUtilization `json:"utilization"`
	LabelUtilizations map[string]labelUtilization `json:"labelUtilizations"`
	Level             string                      `json:"level"`
}

func GetAllUsers(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getUserProxy(sdkConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	// Newly created resources often aren't returned unless there's a delay
	time.Sleep(5 * time.Second)

	users, proxyResponse, err := proxy.getAllUser(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of users error: %s", err), proxyResponse)
	}

	// Add resources to metamap
	for _, user := range *users {
		resources[*user.Id] = &resourceExporter.ResourceMeta{BlockLabel: *user.Email}
	}

	return resources, nil
}

func createUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getUserProxy(sdkConfig)

	email := d.Get("email").(string)
	password := d.Get("password").(string)
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

	// Optional attributes that should not be empty strings
	if password != "" {
		createUser.Password = &password
	}

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

	diagErr := executeAllUpdates(d, proxy, sdkConfig, false)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Created user %s %s", email, *userResponse.Id)
	return readUser(ctx, d, meta)
}

func readUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getUserProxy(sdkConfig)
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
			if util.IsStatus404(proxyResponse) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read user %s | error: %s", d.Id(), errGet), proxyResponse))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read user %s | error: %s", d.Id(), errGet), proxyResponse))
		}

		// Required attributes
		resourcedata.SetNillableValue(d, "name", currentUser.Name)
		resourcedata.SetNillableValue(d, "email", currentUser.Email)
		resourcedata.SetNillableValue(d, "division_id", currentUser.Division.Id)
		resourcedata.SetNillableValue(d, "state", currentUser.State)
		resourcedata.SetNillableValue(d, "department", currentUser.Department)
		resourcedata.SetNillableValue(d, "title", currentUser.Title)
		resourcedata.SetNillableValue(d, "acd_auto_answer", currentUser.AcdAutoAnswer)
		if currentUser.Manager != nil {
			d.Set("manager", *(*currentUser.Manager).Id)
		} else {
			d.Set("manager", nil)
		}
		d.Set("addresses", flattenUserAddresses(d, currentUser.Addresses))
		d.Set("routing_skills", flattenUserSkills(currentUser.Skills))
		d.Set("routing_languages", flattenUserLanguages(currentUser.Languages))
		d.Set("locations", flattenUserLocations(currentUser.Locations))
		d.Set("profile_skills", flattenUserData(currentUser.ProfileSkills))
		d.Set("certifications", flattenUserData(currentUser.Certifications))
		d.Set("employer_info", flattenUserEmployerInfo(currentUser.EmployerInfo))

		if diagErr := readUserRoutingUtilization(d, proxy); diagErr != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		log.Printf("Read user %s %s", d.Id(), *currentUser.Email)

		return cc.CheckState(d)
	})
}

func updateUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getUserProxy(sdkConfig)

	addresses, err := buildSdkAddresses(d)
	if err != nil {
		return err
	}

	email := d.Get("email").(string)

	log.Printf("Updating user %s", email)

	// If state changes, it is the only modifiable field, so it must be updated separately
	if d.HasChange("state") {
		log.Printf("Updating state for user %s", email)
		updateUser := platformclientv2.Updateuser{
			State: platformclientv2.String(d.Get("state").(string)),
		}
		diagErr := executeUpdateUser(ctx, d, proxy, updateUser)
		if diagErr != nil {
			return diagErr
		}
	}

	updateUser := platformclientv2.Updateuser{
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
	diagErr := executeUpdateUser(ctx, d, proxy, updateUser)
	if diagErr != nil {
		return diagErr
	}

	diagErr = executeAllUpdates(d, proxy, sdkConfig, true)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating user %s", email)
	return readUser(ctx, d, meta)
}

func deleteUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getUserProxy(sdkConfig)

	email := d.Get("email").(string)

	log.Printf("Deleting user %s", email)
	err := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		_, proxyDelResponse, err := proxy.deleteUser(ctx, d.Id())
		if err != nil {
			time.Sleep(5 * time.Second)
			return proxyDelResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete user %s error: %s", d.Id(), err), proxyDelResponse)
		}
		log.Printf("Deleted user %s", email)
		return nil, nil
	})
	if err != nil {
		return err
	}

	// Verify user in deleted state and search index has been updated
	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		id, err := getDeletedUserId(email, proxy)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error searching for deleted user %s: %v", email, err))
		}
		if id == nil {
			return retry.RetryableError(fmt.Errorf("User %s not yet in deleted state", email))
		}
		return nil
	})
}
