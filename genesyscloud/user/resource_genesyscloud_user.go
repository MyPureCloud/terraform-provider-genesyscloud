package user

import (
	"context"
	"fmt"
	"log"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

const falseValue = "false"
const trueValue = "true"
const nullValue = "null"

var (
	utilizationMediaTypes = map[string]string{
		"call":     "call",
		"callback": "callback",
		"chat":     "chat",
		"email":    "email",
		"message":  "message",
	}

	contactTypeEmail = "EMAIL"
)

// getAllUsers is used to retrieve all of the users currently in the Genesys Cloud platform
func getAllUsers(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	userProxy := getUserProxy(sdkConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	users, err := userProxy.getAllUsers(ctx)

	if err != nil {
		return nil, diag.Errorf("%v", err)
	}

	for _, user := range *users {
		resources[*user.Id] = &resourceExporter.ResourceMeta{Name: *user.Email}
	}

	return resources, nil
}

func createUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	email := d.Get("email").(string)
	name := d.Get("name").(string)
	password := d.Get("password").(string)
	state := d.Get("state").(string)
	divisionID := d.Get("division_id").(string)
	department := d.Get("department").(string)
	title := d.Get("title").(string)
	manager := d.Get("manager").(string)
	acdAutoAnswer := d.Get("acd_auto_answer").(bool)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	addresses, addrErr := buildSdkAddresses(d)
	if addrErr != nil {
		return addrErr
	}

	createUser := platformclientv2.Createuser{
		Email:      &email,
		Name:       &name,
		State:      &state,
		Addresses:  addresses,
		Department: &department,
		Title:      &title,
	}

	// Optional attributes that should not be empty strings
	if password != "" {
		createUser.Password = &password
	}
	if divisionID != "" {
		createUser.DivisionId = &divisionID
	}

	log.Printf("Creating user %s", email)
	user, resp, err := usersAPI.PostUsers(createUser)
	if err != nil {
		if resp != nil && resp.Error != nil && (*resp.Error).Code == "general.conflict" {
			// Check for a deleted user
			id, diagErr := getDeletedUserId(email, usersAPI)
			if diagErr != nil {
				return diagErr
			}
			if id != nil {
				d.SetId(*id)
				return restoreDeletedUser(ctx, d, meta, usersAPI)
			}
		}
		return diag.Errorf("Failed to create user %s: %s", email, err)
	}

	d.SetId(*user.Id)

	// Set attributes that can only be modified in a patch
	if d.HasChanges(
		"manager",
		"locations",
		"acd_auto_answer",
		"profile_skills",
		"certifications",
		"employer_info") {
		log.Printf("Updating additional attributes for user %s", email)
		_, _, patchErr := usersAPI.PatchUser(d.Id(), platformclientv2.Updateuser{
			Manager:        &manager,
			Locations:      buildSdkLocations(d),
			AcdAutoAnswer:  &acdAutoAnswer,
			Certifications: buildSdkCertifications(d),
			EmployerInfo:   buildSdkEmployerInfo(d),
			Version:        user.Version,
		})
		if patchErr != nil {
			return diag.Errorf("Failed to update user %s: %v", d.Id(), patchErr)
		}
	}

	diagErr := updateUserSkills(ctx, sdkConfig, d)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserLanguages(ctx, sdkConfig, d)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserProfileSkills(ctx, sdkConfig, d)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserRoutingUtilization(ctx, sdkConfig, d)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Created user %s %s", email, *user.Id)
	return readUser(ctx, d, meta)
}

func readUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	log.Printf("Reading user %s", d.Id())
	return gcloud.WithRetriesForRead(ctx, d, func() *resource.RetryError {
		currentUser, resp, getErr := usersAPI.GetUser(d.Id(), []string{
			// Expands
			"skills",
			"languages",
			"locations",
			"profileSkills",
			"certifications",
			"employerInfo",
		}, "", "")

		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read user %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read user %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceUser())

		// Required attributes
		d.Set("name", *currentUser.Name)
		d.Set("email", *currentUser.Email)
		d.Set("division_id", *currentUser.Division.Id)
		d.Set("state", *currentUser.State)

		if currentUser.Department != nil {
			d.Set("department", *currentUser.Department)
		} else {
			d.Set("department", nil)
		}

		if currentUser.Title != nil {
			d.Set("title", *currentUser.Title)
		} else {
			d.Set("title", nil)
		}

		if currentUser.Manager != nil {
			d.Set("manager", *(*currentUser.Manager).Id)
		} else {
			d.Set("manager", nil)
		}

		if currentUser.AcdAutoAnswer != nil {
			d.Set("acd_auto_answer", *currentUser.AcdAutoAnswer)
		} else {
			d.Set("acd_auto_answer", nil)
		}

		d.Set("addresses", flattenUserAddresses(d, currentUser.Addresses))
		d.Set("routing_skills", flattenUserSkills(currentUser.Skills))
		d.Set("routing_languages", flattenUserLanguages(currentUser.Languages))
		d.Set("locations", flattenUserLocations(currentUser.Locations))
		d.Set("profile_skills", flattenUserProfileSkills(currentUser.ProfileSkills))
		d.Set("certifications", flattenUserCertifications(currentUser.Certifications))
		d.Set("employer_info", flattenUserEmployerInfo(currentUser.EmployerInfo))

		if diagErr := readUserRoutingUtilization(d, usersAPI); diagErr != nil {
			return resource.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		log.Printf("Read user %s %s", d.Id(), *currentUser.Email)
		return cc.CheckState()
	})
}

func updateUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	email := d.Get("email").(string)
	state := d.Get("state").(string)
	department := d.Get("department").(string)
	title := d.Get("title").(string)
	manager := d.Get("manager").(string)
	acdAutoAnswer := d.Get("acd_auto_answer").(bool)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	addresses, err := buildSdkAddresses(d)
	if err != nil {
		return err
	}

	log.Printf("Updating user %s", email)

	// If state changes, it is the only modifiable field, so it must be updated separately
	if d.HasChange("state") {
		log.Printf("Updating state for user %s", email)
		patchErr := patchUser(d.Id(), platformclientv2.Updateuser{
			State: &state,
		}, usersAPI)
		if patchErr != nil {
			return patchErr
		}
	}

	patchErr := patchUser(d.Id(), platformclientv2.Updateuser{
		Name:           &name,
		Email:          &email,
		Department:     &department,
		Title:          &title,
		Manager:        &manager,
		Addresses:      addresses,
		Locations:      buildSdkLocations(d),
		AcdAutoAnswer:  &acdAutoAnswer,
		Certifications: buildSdkCertifications(d),
		EmployerInfo:   buildSdkEmployerInfo(d),
	}, usersAPI)
	if patchErr != nil {
		return patchErr
	}

	diagErr := gcloud.UpdateObjectDivision(d, "USER", sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserSkills(ctx, sdkConfig, d)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserLanguages(ctx, sdkConfig, d)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserProfileSkills(ctx, sdkConfig, d)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserRoutingUtilization(ctx, sdkConfig, d)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating user %s", email)
	return readUser(ctx, d, meta)
}

func deleteUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	email := d.Get("email").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	log.Printf("Deleting user %s", email)
	err := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		_, resp, err := usersAPI.DeleteUser(d.Id())
		if err != nil {
			time.Sleep(5 * time.Second)
			return resp, diag.Errorf("Failed to delete user %s: %s", email, err)
		}
		log.Printf("Deleted user %s", email)
		return nil, nil
	})
	if err != nil {
		return err
	}

	// Verify user in deleted state and search index has been updated
	return gcloud.WithRetries(ctx, 180*time.Second, func() *resource.RetryError {
		id, err := getDeletedUserId(email, usersAPI)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error searching for deleted user %s: %v", email, err))
		}
		if id == nil {
			return resource.RetryableError(fmt.Errorf("User %s not yet in deleted state", email))
		}
		return nil
	})
}

func patchUser(id string, update platformclientv2.Updateuser, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	return patchUserWithState(id, "", update, usersAPI)
}

func patchUserWithState(id string, state string, update platformclientv2.Updateuser, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	return gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		currentUser, _, getErr := usersAPI.GetUser(id, nil, "", state)
		if getErr != nil {
			return nil, diag.Errorf("Failed to read user %s: %s", id, getErr)
		}

		update.Version = currentUser.Version
		_, resp, patchErr := usersAPI.PatchUser(id, update)
		if patchErr != nil {
			return resp, diag.Errorf("Failed to update user %s: %v", id, patchErr)
		}
		return nil, nil
	})
}

func getDeletedUserId(email string, usersAPI *platformclientv2.UsersApi) (*string, diag.Diagnostics) {
	exactType := "EXACT"
	results, _, getErr := usersAPI.PostUsersSearch(platformclientv2.Usersearchrequest{
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
		return nil, diag.Errorf("Failed to search for user %s: %s", email, getErr)
	}
	if results.Results != nil && len(*results.Results) > 0 {
		// User found
		return (*results.Results)[0].Id, nil
	}
	return nil, nil
}

func restoreDeletedUser(ctx context.Context, d *schema.ResourceData, meta interface{}, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	email := d.Get("email").(string)
	state := d.Get("state").(string)

	log.Printf("Restoring deleted user %s", email)
	patchErr := patchUserWithState(d.Id(), "deleted", platformclientv2.Updateuser{
		State: &state,
	}, usersAPI)
	if patchErr != nil {
		return patchErr
	}
	return updateUser(ctx, d, meta)
}

func readUserRoutingUtilization(d *schema.ResourceData, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	settings, resp, getErr := usersAPI.GetRoutingUserUtilization(d.Id())
	if getErr != nil {
		if gcloud.IsStatus404(resp) {
			d.SetId("") // User doesn't exist
			return nil
		}
		return diag.Errorf("Failed to read Routing Utilization for user %s: %s", d.Id(), getErr)
	}

	if settings != nil && settings.Utilization != nil {
		// If the settings are org-wide, set to empty to indicate no settings on the user
		if settings.Level != nil && *settings.Level == "Organization" {
			d.Set("routing_utilization", []interface{}{})
		} else {
			allSettings := map[string]interface{}{}
			for sdkType, schemaType := range utilizationMediaTypes {
				if mediaSettings, ok := (*settings.Utilization)[sdkType]; ok {
					allSettings[schemaType] = flattenUtilizationSetting(mediaSettings)
				}
			}
			d.Set("routing_utilization", []interface{}{allSettings})
		}
	} else {
		d.Set("routing_utilization", nil)
	}

	return nil
}

func updateUserSkills(ctx context.Context, sdkConfig *platformclientv2.Configuration, d *schema.ResourceData) diag.Diagnostics {
	userProxy := getUserProxy(sdkConfig)
	if d.HasChange("routing_skills") {
		if skillsConfig := d.Get("routing_skills"); skillsConfig != nil {
			sdkSkills := make([]platformclientv2.Userroutingskillpost, 0)

			skillsList := skillsConfig.(*schema.Set).List()
			for _, configSkill := range skillsList {
				skillMap := configSkill.(map[string]interface{})
				skillID := skillMap["skill_id"].(string)
				skillProf := skillMap["proficiency"].(float64)

				sdkSkills = append(sdkSkills, platformclientv2.Userroutingskillpost{
					Id:          &skillID,
					Proficiency: &skillProf,
				})
			}

			return gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
				resp, err := userProxy.updateUserRoutingSkills(ctx, d.Id(), sdkSkills)
				if err != nil {
					return resp, diag.Errorf("Failed to update skills for user %s: %s", d.Id(), err)
				}
				return nil, nil
			})
		}
	}
	return nil
}

func updateUserLanguages(ctx context.Context, sdkConfig *platformclientv2.Configuration, d *schema.ResourceData) diag.Diagnostics {
	userProxy := getUserProxy(sdkConfig)
	if d.HasChange("routing_languages") {
		if languages := d.Get("routing_languages"); languages != nil {
			log.Printf("Updating languages for user %s", d.Get("email"))
			newLangProfs := make(map[string]int)
			langList := languages.(*schema.Set).List()
			newLangIds := make([]string, len(langList))
			for i, lang := range langList {
				langMap := lang.(map[string]interface{})
				newLangIds[i] = langMap["language_id"].(string)
				newLangProfs[newLangIds[i]] = langMap["proficiency"].(int)
			}

			oldSdkLangs, err := userProxy.getUserRoutingLanguages(ctx, d.Id())
			if err != nil {
				return diag.Errorf("%s", err)
			}

			oldLangIds := make([]string, len(*oldSdkLangs))
			oldLangProfs := make(map[string]int)
			for i, lang := range *oldSdkLangs {
				oldLangIds[i] = *lang.Id
				oldLangProfs[oldLangIds[i]] = int(*lang.Proficiency)
			}

			if len(oldLangIds) > 0 {
				langsToRemove := lists.SliceDifference(oldLangIds, newLangIds)
				for _, langID := range langsToRemove {
					diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
						resp, err := userProxy.deleteUserRoutinglanguage(ctx, d.Id(), langID)
						if err != nil {
							return resp, diag.Errorf("Failed to remove language from user %s: %s", d.Id(), err)
						}
						return nil, nil
					})
					if diagErr != nil {
						return diagErr
					}
				}
			}

			if len(newLangIds) > 0 {
				// Languages to add
				langsToAddOrUpdate := lists.SliceDifference(newLangIds, oldLangIds)

				// Check for existing proficiencies to update which can be done with the same API
				for langID, newNum := range newLangProfs {
					if oldNum, found := oldLangProfs[langID]; found {
						if newNum != oldNum {
							langsToAddOrUpdate = append(langsToAddOrUpdate, langID)
						}
					}
				}

				if diagErr := updateUserRoutingLanguages(ctx, sdkConfig, d.Id(), langsToAddOrUpdate, newLangProfs); diagErr != nil {
					return diagErr
				}
			}
			log.Printf("Languages updated for user %s", d.Get("email"))
		}
	}
	return nil
}

func updateUserRoutingLanguages(
	ctx context.Context,
	sdkConfig *platformclientv2.Configuration,
	userID string,
	langsToUpdate []string,
	langProfs map[string]int) diag.Diagnostics {
	userProxy := getUserProxy(sdkConfig)

	// Bulk API restricts language adds to 50 per call
	const maxBatchSize = 50
	for i := 0; i < len(langsToUpdate); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(langsToUpdate) {
			end = len(langsToUpdate)
		}
		var updateChunk []platformclientv2.Userroutinglanguagepost
		for _, id := range langsToUpdate[i:end] {
			newProf := float64(langProfs[id])
			tempId := id
			updateChunk = append(updateChunk, platformclientv2.Userroutinglanguagepost{
				Id:          &tempId,
				Proficiency: &newProf,
			})
		}

		if len(updateChunk) > 0 {
			diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
				resp, err := userProxy.updateUserRoutinglanguages(ctx, userID, updateChunk)
				if err != nil {
					return resp, diag.Errorf("Failed to update languages for user %s: %s", userID, err)
				}
				return nil, nil
			})
			if diagErr != nil {
				return diagErr
			}
		}
	}
	return nil
}

func updateUserProfileSkills(ctx context.Context, sdkConfig *platformclientv2.Configuration, d *schema.ResourceData) diag.Diagnostics {
	userProxy := getUserProxy(sdkConfig)
	if d.HasChange("profile_skills") {
		if profileSkills := d.Get("profile_skills"); profileSkills != nil {
			profileSkills := lists.SetToStringList(profileSkills.(*schema.Set))
			diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
				_, resp, err := userProxy.updateUserProfileSkills(ctx, d.Id(), *profileSkills)
				if err != nil {
					return resp, diag.Errorf("Failed to update profile skills for user %s: %s", d.Id(), err)
				}
				return nil, nil
			})
			if diagErr != nil {
				return diagErr
			}
		}
	}
	return nil
}

// Done
func updateUserRoutingUtilization(ctx context.Context, sdkConfig *platformclientv2.Configuration, d *schema.ResourceData) diag.Diagnostics {
	userProxy := getUserProxy(sdkConfig)

	if d.HasChange("routing_utilization") {
		if utilConfig := d.Get("routing_utilization").([]interface{}); utilConfig != nil {
			if len(utilConfig) > 0 { // Specified but empty utilization list will reset to org-wide defaults
				sdkSettings := make(map[string]platformclientv2.Mediautilization)
				allSettings := utilConfig[0].(map[string]interface{})
				for sdkType, schemaType := range utilizationMediaTypes {
					if mediaSettings, ok := allSettings[schemaType]; ok && len(mediaSettings.([]interface{})) > 0 {
						sdkSettings[sdkType] = buildSdkMediaUtilization(mediaSettings.([]interface{}))
					}
				}
				// Update settings
				err := userProxy.updateUserRoutingUtilization(ctx, d.Id(), &platformclientv2.Utilization{
					Utilization: &sdkSettings,
				})
				if err != nil {
					return diag.Errorf("Failed to update Routing Utilization for user %s: %s", d.Id(), err)
				}
			} else {
				// Reset to org-wide defaults
				err := userProxy.deleteRoutingUserUtilization(ctx, d.Id())
				if err != nil {
					return diag.Errorf("Failed to delete Routing Utilization for user %s: %s", d.Id(), err)
				}
			}
		}
	}
	return nil
}

// Basic user with minimum required fields
func GenerateBasicUserResource(resourceID string, email string, name string) string {
	return GenerateUserResource(resourceID, email, name, nullValue, nullValue, nullValue, nullValue, nullValue, "", "")
}

func GenerateUserResource(
	resourceID string,
	email string,
	name string,
	state string,
	title string,
	department string,
	manager string,
	acdAutoAnswer string,
	profileSkills string,
	certifications string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		state = %s
		title = %s
		department = %s
		manager = %s
		acd_auto_answer = %s
		profile_skills = [%s]
		certifications = [%s]
	}
	`, resourceID, email, name, state, title, department, manager, acdAutoAnswer, profileSkills, certifications)
}
