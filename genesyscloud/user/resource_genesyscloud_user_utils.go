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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
	"github.com/nyaruka/phonenumbers"
)

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

func buildSdkAddresses(d *schema.ResourceData) (*[]platformclientv2.Contact, diag.Diagnostics) {
	sdkAddresses := make([]platformclientv2.Contact, 0)
	if addresses := d.Get("addresses").([]interface{}); addresses != nil {
		var otherEmails *schema.Set
		var phoneNumbers *schema.Set
		if len(addresses) > 0 {
			if addressMap, ok := addresses[0].(map[string]interface{}); ok {
				otherEmails = addressMap["other_emails"].(*schema.Set)
				phoneNumbers = addressMap["phone_numbers"].(*schema.Set)
			} else {
				return nil, nil
			}
		}

		if otherEmails != nil {
			sdkAddresses = append(sdkAddresses, buildSdkEmails(otherEmails)...)
		}
		if phoneNumbers != nil {
			sdkNums, err := buildSdkPhoneNumbers(phoneNumbers)
			if err != nil {
				return nil, err
			}
			sdkAddresses = append(sdkAddresses, sdkNums...)
		}
		return &sdkAddresses, nil
	}
	return &sdkAddresses, nil
}

func executeUpdateUser(ctx context.Context, d *schema.ResourceData, proxy *userProxy, updateUser platformclientv2.Updateuser) diag.Diagnostics {
	return util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		currentUser, proxyResponse, errGet := proxy.getUserById(ctx, d.Id(), nil, "")
		if errGet != nil {
			return proxyResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read user %s error: %s", d.Id(), errGet), proxyResponse)
		}

		updateUser.Version = currentUser.Version

		_, proxyPatchResponse, patchErr := proxy.updateUser(ctx, d.Id(), &updateUser)
		if patchErr != nil {
			return proxyPatchResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Faild to update user %s | Error: %s.", d.Id(), patchErr), proxyPatchResponse)
		}
		return proxyPatchResponse, nil
	})
}

func executeAllUpdates(ctx context.Context, d *schema.ResourceData, proxy *userProxy, sdkConfig *platformclientv2.Configuration, updateObjectDivision bool) diag.Diagnostics {

	if updateObjectDivision {
		diagErr := util.UpdateObjectDivision(d, "USER", sdkConfig)
		if diagErr != nil {
			return diagErr
		}
	}

	diagErr := updateUserSkills(d, proxy)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserLanguages(d, proxy)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserProfileSkills(d, proxy)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserRoutingUtilization(d, proxy)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserVoicemailPolicies(d, proxy)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updatePassword(ctx, d, proxy)
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func updateUserSkills(d *schema.ResourceData, proxy *userProxy) diag.Diagnostics {
	if d.HasChange("routing_skills") {
		if skillsConfig := d.Get("routing_skills"); skillsConfig != nil {
			log.Printf("Updating skills for user %s", d.Get("email"))
			newSkillProfs := make(map[string]float64)
			skillList := skillsConfig.(*schema.Set).List()
			newSkillIds := make([]string, len(skillList))
			for i, skill := range skillList {
				skillMap := skill.(map[string]interface{})
				newSkillIds[i] = skillMap["skill_id"].(string)
				newSkillProfs[newSkillIds[i]] = skillMap["proficiency"].(float64)
			}

			oldSdkSkills, err := getUserRoutingSkills(d.Id(), proxy)
			if err != nil {
				return err
			}

			oldSkillIds := make([]string, len(oldSdkSkills))
			oldSkillProfs := make(map[string]float64)
			for i, skill := range oldSdkSkills {
				oldSkillIds[i] = *skill.Id
				oldSkillProfs[oldSkillIds[i]] = *skill.Proficiency
			}

			if len(oldSkillIds) > 0 {
				skillsToRemove := lists.SliceDifference(oldSkillIds, newSkillIds)
				for _, skillId := range skillsToRemove {
					diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
						resp, err := proxy.userApi.DeleteUserRoutingskill(d.Id(), skillId)
						if err != nil {
							return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove skill from user %s error: %s", d.Id(), err), resp)
						}
						return nil, nil
					})
					if diagErr != nil {
						return diagErr
					}
				}
			}

			if len(newSkillIds) > 0 {
				// skills to add
				skillsToAddOrUpdate := lists.SliceDifference(newSkillIds, oldSkillIds)
				// Check for existing proficiencies to update which can be done with the same API
				for langID, newNum := range newSkillProfs {
					if oldNum, found := oldSkillProfs[langID]; found {
						if newNum != oldNum {
							skillsToAddOrUpdate = append(skillsToAddOrUpdate, langID)
						}
					}
				}

				if diagErr := updateUserRoutingSkills(d.Id(), skillsToAddOrUpdate, newSkillProfs, proxy); diagErr != nil {
					return diagErr
				}
			}

		}
	}
	return nil
}

func updateUserLanguages(d *schema.ResourceData, proxy *userProxy) diag.Diagnostics {
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

			oldSdkLangs, err := getUserRoutingLanguages(d.Id(), proxy)
			if err != nil {
				return err
			}

			oldLangIds := make([]string, len(oldSdkLangs))
			oldLangProfs := make(map[string]int)
			for i, lang := range oldSdkLangs {
				oldLangIds[i] = *lang.Id
				oldLangProfs[oldLangIds[i]] = int(*lang.Proficiency)
			}

			if len(oldLangIds) > 0 {
				langsToRemove := lists.SliceDifference(oldLangIds, newLangIds)
				for _, langID := range langsToRemove {
					diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
						resp, err := proxy.userApi.DeleteUserRoutinglanguage(d.Id(), langID)
						if err != nil {
							return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove language from user %s error: %s", d.Id(), err), resp)
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
				if diagErr := updateUserRoutingLanguages(d.Id(), langsToAddOrUpdate, newLangProfs, proxy); diagErr != nil {
					return diagErr
				}
			}
			log.Printf("Languages updated for user %s", d.Get("email"))
		}
	}
	return nil
}

func updateUserProfileSkills(d *schema.ResourceData, proxy *userProxy) diag.Diagnostics {
	if d.HasChange("profile_skills") {
		if profileSkills := d.Get("profile_skills"); profileSkills != nil {
			profileSkills := lists.SetToStringList(profileSkills.(*schema.Set))
			diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
				_, resp, err := proxy.userApi.PutUserProfileskills(d.Id(), *profileSkills)
				if err != nil {
					return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update profile skills for user %s error: %s", d.Id(), err), resp)
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

func updateUserVoicemailPolicies(d *schema.ResourceData, proxy *userProxy) diag.Diagnostics {
	if !d.HasChange("voicemail_userpolicies") {
		return nil
	}

	voicemailUserpolicies := d.Get("voicemail_userpolicies").([]interface{})
	reqBody := buildVoicemailUserpoliciesRequest(voicemailUserpolicies)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		_, proxyPutResponse, putErr := proxy.voicemailApi.PatchVoicemailUserpolicy(d.Id(), reqBody)
		if putErr != nil {
			return proxyPutResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update voicemail userpolicices for user %s error: %s", d.Id(), putErr), proxyPutResponse)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func updateUserRoutingUtilization(d *schema.ResourceData, proxy *userProxy) diag.Diagnostics {
	if d.HasChange("routing_utilization") {
		if utilConfig := d.Get("routing_utilization").([]interface{}); utilConfig != nil {
			var err error

			log.Printf("Updating user utilization for user %s", d.Id())

			if len(utilConfig) > 0 && utilConfig[0] != nil { // Specified but empty utilization list will reset to org-wide defaults
				// Update settings
				allSettings := utilConfig[0].(map[string]interface{})
				labelUtilizations := allSettings["label_utilizations"].([]interface{})

				if len(labelUtilizations) > 0 {
					apiClient := &proxy.routingApi.Configuration.APIClient

					path := fmt.Sprintf("%s/api/v2/routing/users/%s/utilization", proxy.routingApi.Configuration.BasePath, d.Id())
					headerParams := buildHeaderParams(proxy.routingApi)
					requestPayload := make(map[string]interface{})
					requestPayload["utilization"] = buildMediaTypeUtilizations(allSettings)
					requestPayload["labelUtilizations"] = buildLabelUtilizationsRequest(labelUtilizations)
					_, err = apiClient.CallAPI(path, "PUT", requestPayload, headerParams, nil, nil, "", nil, "")
				} else {
					sdkSettings := make(map[string]platformclientv2.Mediautilization)
					for sdkType, schemaType := range getUtilizationMediaTypes() {
						if mediaSettings, ok := allSettings[schemaType]; ok && len(mediaSettings.([]interface{})) > 0 {
							sdkSettings[sdkType] = buildSdkMediaUtilization(mediaSettings.([]interface{}))
						}
					}

					_, _, err = proxy.userApi.PutRoutingUserUtilization(d.Id(), platformclientv2.Utilizationrequest{
						Utilization: &sdkSettings,
					})
				}

				if err != nil {
					return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Routing Utilization for user %s", d.Id()), err)
				}
			} else {
				// Reset to org-wide defaults
				resp, err := proxy.userApi.DeleteRoutingUserUtilization(d.Id())
				if err != nil {
					return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete routing utilization for user %s error: %s", d.Id(), err), resp)
				}
			}

			log.Printf("Updated user utilization for user %s", d.Id())
		}
	}
	return nil
}

func updatePassword(ctx context.Context, d *schema.ResourceData, proxy *userProxy) diag.Diagnostics {
	if !d.HasChange("password") {
		return nil
	}

	password := d.Get("password").(string)

	if password == "" {
		return nil // Skip password update if empty
	}

	_, err := proxy.updatePassword(ctx, d.Id(), password)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to update password for user %s", d.Id()), err)
	}

	return nil
}

func updateUserRoutingSkills(userID string, skillsToUpdate []string, skillProfs map[string]float64, proxy *userProxy) diag.Diagnostics {
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

	chunkProcessor := func(chunk []platformclientv2.Userroutingskillpost) diag.Diagnostics {
		diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			_, resp, err := proxy.userApi.PatchUserRoutingskillsBulk(userID, chunk)
			if err != nil {
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update languages for user %s error: %s", userID, err), resp)
			}
			return nil, nil
		})
		if diagErr != nil {
			return diagErr
		}
		return nil
	}

	// Generic Function call which takes in the chunks and the processing function
	return chunksProcess.ProcessChunks(chunks, chunkProcessor)
}

func updateUserRoutingLanguages(userID string, langsToUpdate []string, langProfs map[string]int, proxy *userProxy) diag.Diagnostics {
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

	chunkProcessor := func(chunk []platformclientv2.Userroutinglanguagepost) diag.Diagnostics {
		diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			_, resp, err := proxy.userApi.PatchUserRoutinglanguagesBulk(userID, chunk)
			if err != nil {
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update languages for user %s error: %s", userID, err), resp)
			}
			return nil, nil
		})
		if diagErr != nil {
			return diagErr
		}
		return nil
	}

	// Genric Function call which takes in the chunks and the processing function
	return chunksProcess.ProcessChunks(chunks, chunkProcessor)
}

func getUserRoutingLanguages(userID string, proxy *userProxy) ([]platformclientv2.Userroutinglanguage, diag.Diagnostics) {
	const maxPageSize = 50

	var sdkLanguages []platformclientv2.Userroutinglanguage
	for pageNum := 1; ; pageNum++ {
		langs, resp, err := proxy.userApi.GetUserRoutinglanguages(userID, maxPageSize, pageNum, "")
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to query languages for user %s error: %s", userID, err), resp)
		}
		if langs == nil || langs.Entities == nil || len(*langs.Entities) == 0 {
			return sdkLanguages, nil
		}

		sdkLanguages = append(sdkLanguages, *langs.Entities...)
	}
}

func getUserRoutingSkills(userID string, proxy *userProxy) ([]platformclientv2.Userroutingskill, diag.Diagnostics) {
	const maxPageSize = 50

	var sdkSkills []platformclientv2.Userroutingskill
	for pageNum := 1; ; pageNum++ {
		skills, resp, err := proxy.userApi.GetUserRoutingskills(userID, maxPageSize, pageNum, "")
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to query languages for user %s error: %s", userID, err), resp)
		}
		if skills == nil || skills.Entities == nil || len(*skills.Entities) == 0 {
			return sdkSkills, nil
		}

		sdkSkills = append(sdkSkills, *skills.Entities...)
	}
}

func getDeletedUserId(email string, proxy *userProxy) (*string, diag.Diagnostics) {
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
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to search for user %s error: %s", email, getErr), resp)
	}
	if results.Results != nil && len(*results.Results) > 0 {
		// User found
		return (*results.Results)[0].Id, nil
	}
	return nil, nil
}

func restoreDeletedUser(ctx context.Context, d *schema.ResourceData, meta interface{}, proxy *userProxy) diag.Diagnostics {
	email := d.Get("email").(string)
	state := d.Get("state").(string)

	log.Printf("Restoring deleted user %s", email)

	return util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		currentUser, proxyResp, err := proxy.getUserById(ctx, d.Id(), nil, "deleted")
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read user %s error: %s", d.Id(), err), proxyResp)
		}

		_, proxyPatchResponse, patchErr := proxy.patchUserWithState(ctx, d.Id(), &platformclientv2.Updateuser{
			State:   &state,
			Version: currentUser.Version,
		})

		if patchErr != nil {
			return proxyPatchResponse, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Faild to restored deleted user %s | Error: %s.", email, patchErr), proxyPatchResponse)
		}

		return nil, updateUser(ctx, d, meta)
	})
}

func readUserRoutingUtilization(d *schema.ResourceData, proxy *userProxy) diag.Diagnostics {
	log.Printf("Getting user utilization")

	apiClient := &proxy.routingApi.Configuration.APIClient

	path := fmt.Sprintf("%s/api/v2/routing/users/%s/utilization", proxy.routingApi.Configuration.BasePath, d.Id())
	headerParams := buildHeaderParams(proxy.routingApi)
	response, err := apiClient.CallAPI(path, "GET", nil, headerParams, nil, nil, "", nil, "")

	if err != nil {
		if util.IsStatus404(response) {
			d.SetId("") // User doesn't exist
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read routing utilization for user %s error: %s", d.Id(), err), response)
	}

	agentUtilization := &agentUtilizationWithLabels{}
	err = json.Unmarshal(response.RawBody, &agentUtilization)
	if err != nil {
		log.Printf("[WARN] failed to unmarshal json: %s", err.Error())
	}

	if agentUtilization == nil {
		_ = d.Set("routing_utilization", nil)
	} else if agentUtilization.Level == "Organization" {
		// If the settings are org-wide, set to empty to indicate no settings on the user
		_ = d.Set("routing_utilization", []interface{}{})
	} else {
		allSettings := map[string]interface{}{}

		if agentUtilization.Utilization != nil {
			for sdkType, schemaType := range getUtilizationMediaTypes() {
				if mediaSettings, ok := agentUtilization.Utilization[sdkType]; ok {
					allSettings[schemaType] = flattenUtilizationSetting(mediaSettings)
				}
			}
		}

		if agentUtilization.LabelUtilizations != nil {
			utilConfig := d.Get("routing_utilization").([]interface{})
			if len(utilConfig) > 0 && utilConfig[0] != nil {
				originalSettings := utilConfig[0].(map[string]interface{})
				originalLabelUtilizations := originalSettings["label_utilizations"].([]interface{})

				// Only add to the state the configured labels, in the configured order, but not any extras, to help terraform with matching new and old state.
				filteredLabelUtilizations := filterAndFlattenLabelUtilizations(agentUtilization.LabelUtilizations, originalLabelUtilizations)

				allSettings["label_utilizations"] = filteredLabelUtilizations
			} else {
				allSettings["label_utilizations"] = make([]interface{}, 0)
			}
		}

		_ = d.Set("routing_utilization", []interface{}{allSettings})
	}

	return nil
}

func phoneNumberHash(val interface{}) int {
	// Copy map to avoid modifying state
	phoneMap := make(map[string]interface{})
	for k, v := range val.(map[string]interface{}) {
		if k != "extension_pool_id" {
			phoneMap[k] = v
		}
	}
	if num, ok := phoneMap["number"]; ok {
		// Attempt to format phone numbers before hashing
		number, err := phonenumbers.Parse(num.(string), "US")
		if err == nil {
			phoneMap["number"] = phonenumbers.Format(number, phonenumbers.E164)
		}
	}
	return schema.HashResource(phoneNumberResource)(phoneMap)
}

func buildSdkEmails(configEmails *schema.Set) []platformclientv2.Contact {
	emailSlice := configEmails.List()
	sdkContacts := make([]platformclientv2.Contact, len(emailSlice))
	for i, configEmail := range emailSlice {
		emailMap := configEmail.(map[string]interface{})
		emailAddress, _ := emailMap["address"].(string)
		emailType, _ := emailMap["type"].(string)

		sdkContacts[i] = platformclientv2.Contact{
			Address:   &emailAddress,
			MediaType: &contactTypeEmail,
			VarType:   &emailType,
		}
	}
	return sdkContacts
}

func buildSdkPhoneNumbers(configPhoneNumbers *schema.Set) ([]platformclientv2.Contact, diag.Diagnostics) {
	phoneNumberSlice := configPhoneNumbers.List()
	sdkContacts := make([]platformclientv2.Contact, len(phoneNumberSlice))
	for i, configPhone := range phoneNumberSlice {
		phoneMap := configPhone.(map[string]interface{})
		phoneMediaType := phoneMap["media_type"].(string)
		phoneType := phoneMap["type"].(string)

		contact := platformclientv2.Contact{
			MediaType: &phoneMediaType,
			VarType:   &phoneType,
		}

		if phoneNum, ok := phoneMap["number"].(string); ok && phoneNum != "" {
			contact.Address = &phoneNum
		}
		if phoneExt, ok := phoneMap["extension"].(string); ok && phoneExt != "" {
			contact.Extension = &phoneExt
		}

		sdkContacts[i] = contact
	}
	return sdkContacts, nil
}

func buildSdkLocations(d *schema.ResourceData) *[]platformclientv2.Location {
	if locationConfig := d.Get("locations"); locationConfig != nil {
		sdkLocations := make([]platformclientv2.Location, 0)
		locationList := locationConfig.(*schema.Set).List()
		for _, configLoc := range locationList {
			locMap := configLoc.(map[string]interface{})
			locID := locMap["location_id"].(string)
			locNotes := locMap["notes"].(string)

			sdkLocations = append(sdkLocations, platformclientv2.Location{
				Id:    &locID,
				Notes: &locNotes,
			})
		}
		return &sdkLocations
	}
	return nil
}

func buildSdkEmployerInfo(d *schema.ResourceData) *platformclientv2.Employerinfo {
	if configInfo := d.Get("employer_info").([]interface{}); configInfo != nil {
		var sdkInfo platformclientv2.Employerinfo
		if len(configInfo) > 0 && configInfo[0] != nil {
			if _, ok := configInfo[0].(map[string]interface{}); !ok {
				return nil
			}
			infoMap := configInfo[0].(map[string]interface{})
			// Only set non-empty values.
			if offName := infoMap["official_name"].(string); len(offName) > 0 {
				sdkInfo.OfficialName = &offName
			}
			if empID := infoMap["employee_id"].(string); len(empID) > 0 {
				sdkInfo.EmployeeId = &empID
			}
			if empType := infoMap["employee_type"].(string); len(empType) > 0 {
				sdkInfo.EmployeeType = &empType
			}
			if dateHire := infoMap["date_hire"].(string); len(dateHire) > 0 {
				sdkInfo.DateHire = &dateHire
			}
		}
		return &sdkInfo
	}
	return nil
}

func buildSdkCertifications(d *schema.ResourceData) *[]string {
	if certs := d.Get("certifications"); certs != nil {
		return lists.SetToStringList(certs.(*schema.Set))
	}
	return nil
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

func flattenUserAddresses(ctx context.Context, addresses *[]platformclientv2.Contact, proxy *userProxy) []interface{} {
	if addresses == nil || len(*addresses) == 0 {
		return nil
	}

	emailSet := schema.NewSet(schema.HashResource(otherEmailResource), []interface{}{})
	phoneNumSet := schema.NewSet(phoneNumberHash, []interface{}{})

	utilE164 := util.NewUtilE164Service()

	for _, address := range *addresses {
		if address.MediaType != nil {
			if *address.MediaType == "SMS" || *address.MediaType == "PHONE" {
				phoneNumber := make(map[string]interface{})
				phoneNumber["media_type"] = *address.MediaType
				phoneNumber["extension_pool_id"] = ""

				// PHONE and SMS Addresses have four different ways they can return in the API
				// We need to be able to handle them all, and strip off any parentheses that can surround
				// values

				//     	1.) Addresses that return an "address" field are phone numbers without extensions
				if address.Address != nil {
					phoneNumber["number"] = utilE164.FormatAsCalculatedE164Number(strings.Trim(*address.Address, "()"))
				}

				// 		2.) Addresses that return an "extension" field that matches the "display" field are
				//          true internal extensions that have been mapped to an extension pool
				if address.Extension != nil {
					if address.Display != nil {
						if *address.Extension == *address.Display {
							extensionNum := strings.Trim(*address.Extension, "()")
							phoneNumber["extension"] = extensionNum
							phoneNumber["extension_pool_id"] = fetchExtensionPoolId(ctx, extensionNum, proxy)
						}
					}
				}

				// 		3.) Addresses that include both an "extension" and "display" field, but they do not
				//          match indicate that this is a phone number plus an extension
				if address.Extension != nil {
					if address.Display != nil {
						if *address.Extension != *address.Display {
							phoneNumber["extension"] = *address.Extension
							phoneNumber["number"] = utilE164.FormatAsCalculatedE164Number(strings.Trim(*address.Display, "()"))
						}
					}
				}

				// 		4.) Addresses that only include a "display" field (but not "address" or "extension") are
				//          considered an extension that has not been mapped to an internal extension pool yet.
				if address.Address == nil && address.Extension == nil && address.Display != nil {
					phoneNumber["extension"] = strings.Trim(*address.Display, "()")
				}

				if address.VarType != nil {
					phoneNumber["type"] = *address.VarType
				}
				phoneNumSet.Add(phoneNumber)
			} else if *address.MediaType == "EMAIL" {
				email := make(map[string]interface{})
				email["type"] = *address.VarType
				email["address"] = *address.Address
				emailSet.Add(email)
			} else {
				log.Printf("Unknown address media type %s", *address.MediaType)
			}
		}
	}
	return []interface{}{map[string]interface{}{
		"other_emails":  emailSet,
		"phone_numbers": phoneNumSet,
	}}
}

func flattenUserEmployerInfo(empInfo *platformclientv2.Employerinfo) []interface{} {
	if empInfo == nil {
		return nil
	}
	var (
		offName  string
		empID    string
		empType  string
		dateHire string
	)

	if empInfo.OfficialName != nil {
		offName = *empInfo.OfficialName
	}
	if empInfo.EmployeeId != nil {
		empID = *empInfo.EmployeeId
	}
	if empInfo.EmployeeType != nil {
		empType = *empInfo.EmployeeType
	}
	if empInfo.DateHire != nil {
		dateHire = *empInfo.DateHire
	}

	return []interface{}{map[string]interface{}{
		"official_name": offName,
		"employee_id":   empID,
		"employee_type": empType,
		"date_hire":     dateHire,
	}}
}

func flattenUserSkills(skills *[]platformclientv2.Userroutingskill) *schema.Set {
	if skills == nil {
		return nil
	}
	skillSet := schema.NewSet(schema.HashResource(userSkillResource), []interface{}{})
	for _, sdkSkill := range *skills {
		skill := make(map[string]interface{})
		skill["skill_id"] = *sdkSkill.Id
		skill["proficiency"] = *sdkSkill.Proficiency
		skillSet.Add(skill)
	}
	return skillSet
}

func flattenUserLanguages(languages *[]platformclientv2.Userroutinglanguage) *schema.Set {
	if languages == nil {
		return nil
	}
	languageSet := schema.NewSet(schema.HashResource(userLanguageResource), []interface{}{})
	for _, sdkLang := range *languages {
		language := make(map[string]interface{})
		language["language_id"] = *sdkLang.Id
		language["proficiency"] = int(*sdkLang.Proficiency)
		languageSet.Add(language)
	}
	return languageSet
}

func flattenUserLocations(locations *[]platformclientv2.Location) *schema.Set {
	if locations == nil {
		return nil
	}
	locSet := schema.NewSet(schema.HashResource(userLocationResource), []interface{}{})
	for _, sdkLoc := range *locations {
		if sdkLoc.LocationDefinition != nil {
			location := make(map[string]interface{})
			location["location_id"] = *sdkLoc.LocationDefinition.Id
			if sdkLoc.Notes != nil {
				location["notes"] = *sdkLoc.Notes
			}
			locSet.Add(location)
		}
	}
	return locSet
}

func flattenUserData(userDataSlice *[]string) *schema.Set {
	if userDataSlice != nil {
		return lists.StringListToSet(*userDataSlice)
	}
	return nil
}

func buildMediaTypeUtilizations(allUtilizations map[string]interface{}) *map[string]platformclientv2.Mediautilization {
	settings := make(map[string]platformclientv2.Mediautilization)

	for sdkType, schemaType := range getUtilizationMediaTypes() {
		mediaSettings := allUtilizations[schemaType].([]interface{})
		if len(mediaSettings) > 0 {
			settings[sdkType] = buildSdkMediaUtilization(mediaSettings)
		}
	}

	return &settings
}

func emailorNameDisambiguation(searchField string) (string, string) {
	emailField := "email"
	nameField := "name"
	_, err := mail.ParseAddress(searchField)
	if err == nil {
		return searchField, emailField
	}
	return searchField, nameField
}

func getUtilizationMediaTypes() map[string]string {
	return utilizationMediaTypes
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

func flattenUtilizationSetting(settings mediaUtilization) []interface{} {
	settingsMap := make(map[string]interface{})

	settingsMap["maximum_capacity"] = settings.MaximumCapacity
	settingsMap["include_non_acd"] = settings.IncludeNonAcd
	if settings.InterruptableMediaTypes != nil {
		settingsMap["interruptible_media_types"] = lists.StringListToSet(settings.InterruptableMediaTypes)
	}

	return []interface{}{settingsMap}
}

func filterAndFlattenLabelUtilizations(labelUtilizations map[string]labelUtilization, originalLabelUtilizations []interface{}) []interface{} {
	flattenedLabelUtilizations := make([]interface{}, 0)

	for _, originalLabelUtilization := range originalLabelUtilizations {
		originalLabelId := (originalLabelUtilization.(map[string]interface{}))["label_id"].(string)

		for currentLabelId, currentLabelUtilization := range labelUtilizations {
			if currentLabelId == originalLabelId {
				flattenedLabelUtilizations = append(flattenedLabelUtilizations, flattenLabelUtilization(currentLabelId, currentLabelUtilization))
				delete(labelUtilizations, currentLabelId)
				break
			}
		}
	}

	return flattenedLabelUtilizations
}

func flattenLabelUtilization(labelId string, labelUtilization labelUtilization) map[string]interface{} {
	utilizationMap := make(map[string]interface{})

	utilizationMap["label_id"] = labelId
	utilizationMap["maximum_capacity"] = labelUtilization.MaximumCapacity
	if labelUtilization.InterruptingLabelIds != nil {
		utilizationMap["interrupting_label_ids"] = lists.StringListToSet(labelUtilization.InterruptingLabelIds)
	}

	return utilizationMap
}

func buildSdkMediaUtilization(settings []interface{}) platformclientv2.Mediautilization {
	settingsMap := settings[0].(map[string]interface{})

	maxCapacity := settingsMap["maximum_capacity"].(int)
	includeNonAcd := settingsMap["include_non_acd"].(bool)

	// Optional
	interruptableMediaTypes := &[]string{}
	if types, ok := settingsMap["interruptible_media_types"]; ok {
		interruptableMediaTypes = lists.SetToStringList(types.(*schema.Set))
	}

	return platformclientv2.Mediautilization{
		MaximumCapacity:         &maxCapacity,
		IncludeNonAcd:           &includeNonAcd,
		InterruptableMediaTypes: interruptableMediaTypes,
	}
}

func buildLabelUtilizationsRequest(labelUtilizations []interface{}) map[string]labelUtilization {
	request := make(map[string]labelUtilization)
	for _, labelUtilizationValue := range labelUtilizations {
		labelUtilizationMap := labelUtilizationValue.(map[string]interface{})
		interruptingLabelIds := lists.SetToStringList(labelUtilizationMap["interrupting_label_ids"].(*schema.Set))

		request[labelUtilizationMap["label_id"].(string)] = labelUtilization{
			MaximumCapacity:      int32(labelUtilizationMap["maximum_capacity"].(int)),
			InterruptingLabelIds: *interruptingLabelIds,
		}
	}
	return request
}

func getSdkUtilizationTypes() []string {
	types := make([]string, 0, len(utilizationMediaTypes))
	for t := range utilizationMediaTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func buildVoicemailUserpoliciesRequest(voicemailUserpolicies []interface{}) platformclientv2.Voicemailuserpolicy {
	var request platformclientv2.Voicemailuserpolicy
	if extractMap, ok := voicemailUserpolicies[0].(map[string]interface{}); ok {
		sendEmailNotifications := extractMap["send_email_notifications"].(bool)
		request = platformclientv2.Voicemailuserpolicy{
			SendEmailNotifications: &sendEmailNotifications,
		}
		// Optional
		if alertTimeoutSeconds := extractMap["alert_timeout_seconds"].(int); alertTimeoutSeconds > 0 {
			request.AlertTimeoutSeconds = &alertTimeoutSeconds
		}
	}
	return request
}

func flattenVoicemailUserpolicies(d *schema.ResourceData, voicemail *platformclientv2.Voicemailuserpolicy) []interface{} {
	if voicemail == nil {
		return nil
	}

	voicemailUserpolicy := make(map[string]interface{})
	if voicemail.AlertTimeoutSeconds != nil {
		voicemailUserpolicy["alert_timeout_seconds"] = *voicemail.AlertTimeoutSeconds
	}
	if voicemail.SendEmailNotifications != nil {
		voicemailUserpolicy["send_email_notifications"] = *voicemail.SendEmailNotifications
	}

	return []interface{}{voicemailUserpolicy}
}

// GenerateBasicUserResource generates a basic user resource with minimum required fields
func GenerateBasicUserResource(resourceLabel string, email string, name string) string {
	return GenerateUserResource(resourceLabel, email, name, util.NullValue, util.NullValue, util.NullValue, util.NullValue, util.NullValue, "", "")
}

func GenerateUserResource(resourceLabel string, email string, name string, state string, title string, department string, manager string, acdAutoAnswer string, profileSkills string, certifications string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
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
	`, ResourceType, resourceLabel, email, name, state, title, department, manager, acdAutoAnswer, profileSkills, certifications)
}

func GenerateVoicemailUserpolicies(timeout int, sendEmailNotifications bool) string {
	return fmt.Sprintf(`voicemail_userpolicies {
		alert_timeout_seconds = %d
		send_email_notifications = %t
	}
	`, timeout, sendEmailNotifications)
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
