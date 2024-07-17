package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	chunksProcess "terraform-provider-genesyscloud/genesyscloud/util/chunks"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
	"github.com/nyaruka/phonenumbers"
)

func buildSdkAddresses(d *schema.ResourceData) (*[]platformclientv2.Contact, diag.Diagnostics) {
	if addresses := d.Get("addresses").([]interface{}); addresses != nil {
		sdkAddresses := make([]platformclientv2.Contact, 0)
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
	return nil, nil
}

func executeUpdateUser(ctx context.Context, d *schema.ResourceData, proxy *userProxy, updateUser platformclientv2.Updateuser) diag.Diagnostics {

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		currentUser, proxyResponse, errGet := proxy.getUserById(ctx, d.Id(), nil, "")

		if errGet != nil {
			return proxyResponse, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to read user %s error: %s", d.Id(), errGet), proxyResponse)
		}
		updateUser.Version = currentUser.Version
		_, proxyPatchResponse, patchErr := proxy.updateUser(ctx, d.Id(), &updateUser)

		if patchErr != nil {
			return proxyPatchResponse, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Faild to update user %s | Error: %s.", *updateUser.Name, patchErr), proxyPatchResponse)
		}
		return proxyPatchResponse, nil
	})

	if diagErr != nil {
		return diagErr
	}

	return nil
}

func executeAllUpdates(d *schema.ResourceData, proxy *userProxy, sdkConfig *platformclientv2.Configuration, updateObjectDivision bool) diag.Diagnostics {

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

	return nil
}

func updateUserSkills(d *schema.ResourceData, proxy *userProxy) diag.Diagnostics {
	transformFunc := func(configSkill interface{}) platformclientv2.Userroutingskillpost {
		skillMap := configSkill.(map[string]interface{})
		skillID := skillMap["skill_id"].(string)
		skillProf := skillMap["proficiency"].(float64)

		return platformclientv2.Userroutingskillpost{
			Id:          &skillID,
			Proficiency: &skillProf,
		}
	}

	chunkProcessor := func(chunk []platformclientv2.Userroutingskillpost) diag.Diagnostics {
		diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			_, resp, err := proxy.userApi.PatchUserRoutingskillsBulk(d.Id(), chunk)
			if err != nil {
				return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update skills for user %s error: %s", d.Id(), err), resp)
			}
			return nil, nil
		})
		if diagErr != nil {
			return diagErr
		}
		return nil
	}

	if d.HasChange("routing_skills") {
		if skillsConfig := d.Get("routing_skills"); skillsConfig != nil {
			skillsList := skillsConfig.(*schema.Set).List()
			chunks := chunksProcess.ChunkItems(skillsList, transformFunc, 50)
			return chunksProcess.ProcessChunks(chunks, chunkProcessor)
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
							return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to remove language from user %s error: %s", d.Id(), err), resp)
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
					return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update profile skills for user %s error: %s", d.Id(), err), resp)
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

func updateUserRoutingUtilization(d *schema.ResourceData, proxy *userProxy) diag.Diagnostics {
	if d.HasChange("routing_utilization") {
		if utilConfig := d.Get("routing_utilization").([]interface{}); utilConfig != nil {
			var err error

			log.Printf("Updating user utilization for user %s", d.Id())

			if len(utilConfig) > 0 && utilConfig[0] != nil { // Specified but empty utilization list will reset to org-wide defaults
				// Update settings
				allSettings := utilConfig[0].(map[string]interface{})
				labelUtilizations := allSettings["label_utilizations"].([]interface{})

				if labelUtilizations != nil && len(labelUtilizations) > 0 {
					apiClient := &proxy.routingApi.Configuration.APIClient

					path := fmt.Sprintf("%s/api/v2/routing/users/%s/utilization", proxy.routingApi.Configuration.BasePath, d.Id())
					headerParams := util.BuildHeaderParams(proxy.routingApi)
					requestPayload := make(map[string]interface{})
					requestPayload["utilization"] = buildMediaTypeUtilizations(allSettings)
					requestPayload["labelUtilizations"] = util.BuildLabelUtilizationsRequest(labelUtilizations)
					_, err = apiClient.CallAPI(path, "PUT", requestPayload, headerParams, nil, nil, "", nil)
				} else {
					sdkSettings := make(map[string]platformclientv2.Mediautilization)
					for sdkType, schemaType := range util.GetUtilizationMediaTypes() {
						if mediaSettings, ok := allSettings[schemaType]; ok && len(mediaSettings.([]interface{})) > 0 {
							sdkSettings[sdkType] = util.BuildSdkMediaUtilization(mediaSettings.([]interface{}))
						}
					}

					_, _, err = proxy.userApi.PutRoutingUserUtilization(d.Id(), platformclientv2.Utilizationrequest{
						Utilization: &sdkSettings,
					})
				}

				if err != nil {
					return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to update Routing Utilization for user %s", d.Id()), err)
				}
			} else {
				// Reset to org-wide defaults
				resp, err := proxy.userApi.DeleteRoutingUserUtilization(d.Id())
				if err != nil {
					return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete routing utilization for user %s error: %s", d.Id(), err), resp)
				}
			}

			log.Printf("Updated user utilization for user %s", d.Id())
		}
	}
	return nil
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
				return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update languages for user %s error: %s", userID, err), resp)
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
			return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to query languages for user %s error: %s", userID, err), resp)
		}
		if langs == nil || langs.Entities == nil || len(*langs.Entities) == 0 {
			return sdkLanguages, nil
		}

		sdkLanguages = append(sdkLanguages, *langs.Entities...)
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
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to search for user %s error: %s", email, getErr), resp)
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
			return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to read user %s error: %s", d.Id(), err), proxyResp)
		}

		_, proxyPatchResponse, patchErr := proxy.patchUserWithState(ctx, d.Id(), &platformclientv2.Updateuser{
			State:   &state,
			Version: currentUser.Version,
		})

		if patchErr != nil {
			return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Faild to restored deleted user %s | Error: %s.", email, patchErr), proxyPatchResponse)
		}

		return nil, updateUser(ctx, d, meta)
	})
}

func readUserRoutingUtilization(d *schema.ResourceData, proxy *userProxy) diag.Diagnostics {
	log.Printf("Getting user utilization")

	apiClient := &proxy.routingApi.Configuration.APIClient

	path := fmt.Sprintf("%s/api/v2/routing/users/%s/utilization", proxy.routingApi.Configuration.BasePath, d.Id())
	headerParams := util.BuildHeaderParams(proxy.routingApi)
	response, err := apiClient.CallAPI(path, "GET", nil, headerParams, nil, nil, "", nil)

	if err != nil {
		if util.IsStatus404(response) {
			d.SetId("") // User doesn't exist
			return nil
		}
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to read routing utilization for user %s error: %s", d.Id(), err), response)
	}

	agentUtilization := &AgentUtilizationWithLabels{}
	json.Unmarshal(response.RawBody, &agentUtilization)

	if agentUtilization == nil {
		d.Set("routing_utilization", nil)
	} else if agentUtilization.Level == "Organization" {
		// If the settings are org-wide, set to empty to indicate no settings on the user
		d.Set("routing_utilization", []interface{}{})
	} else {
		allSettings := map[string]interface{}{}

		if agentUtilization.Utilization != nil {
			for sdkType, schemaType := range util.GetUtilizationMediaTypes() {
				if mediaSettings, ok := agentUtilization.Utilization[sdkType]; ok {
					allSettings[schemaType] = util.FlattenUtilizationSetting(mediaSettings)
				}
			}
		}

		if agentUtilization.LabelUtilizations != nil {
			utilConfig := d.Get("routing_utilization").([]interface{})
			if utilConfig != nil && len(utilConfig) > 0 && utilConfig[0] != nil {
				originalSettings := utilConfig[0].(map[string]interface{})
				originalLabelUtilizations := originalSettings["label_utilizations"].([]interface{})

				// Only add to the state the configured labels, in the configured order, but not any extras, to help terraform with matching new and old state.
				filteredLabelUtilizations := util.FilterAndFlattenLabelUtilizations(agentUtilization.LabelUtilizations, originalLabelUtilizations)

				allSettings["label_utilizations"] = filteredLabelUtilizations
			} else {
				allSettings["label_utilizations"] = make([]interface{}, 0)
			}
		}

		d.Set("routing_utilization", []interface{}{allSettings})
	}

	return nil
}

func phoneNumberHash(val interface{}) int {
	// Copy map to avoid modifying state
	phoneMap := make(map[string]interface{})
	for k, v := range val.(map[string]interface{}) {
		phoneMap[k] = v
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

func getNumbers(d *schema.ResourceData, index int) (bool, bool) {
	isNumber := false
	isExtension := false

	if addresses1 := d.Get("addresses").([]interface{}); addresses1 != nil {
		var phoneNumbers *schema.Set
		if len(addresses1) > 0 {
			addressMap := addresses1[0].(map[string]interface{})
			phoneNumbers = addressMap["phone_numbers"].(*schema.Set)
		}

		if phoneNumbers != nil {
			phoneNumberSlice := phoneNumbers.List()
			for ii, configPhone := range phoneNumberSlice {
				if ii != index {
					continue
				}
				phoneMap := configPhone.(map[string]interface{})
				if phoneNum, ok := phoneMap["number"].(string); ok && phoneNum != "" {
					isNumber = true
				}
				if phoneExt, ok := phoneMap["extension"].(string); ok && phoneExt != "" {
					isExtension = true
				}
				break
			}
		}
	}
	return isNumber, isExtension
}

func flattenUserAddresses(d *schema.ResourceData, addresses *[]platformclientv2.Contact) []interface{} {
	if addresses == nil || len(*addresses) == 0 {
		return nil
	}

	emailSet := schema.NewSet(schema.HashResource(otherEmailResource), []interface{}{})
	phoneNumSet := schema.NewSet(phoneNumberHash, []interface{}{})

	for i, address := range *addresses {
		if address.MediaType != nil {
			if *address.MediaType == "SMS" || *address.MediaType == "PHONE" {
				phoneNumber := make(map[string]interface{})
				phoneNumber["media_type"] = *address.MediaType

				// Strip off any parentheses from phone numbers
				if address.Address != nil {
					phoneNumber["number"] = strings.Trim(*address.Address, "()")
				} else if address.Display != nil {
					// Some numbers are only returned in Display
					isNumber, isExtension := getNumbers(d, i)

					if isNumber && phoneNumber["number"] != "" {
						phoneNumber["number"] = strings.Trim(*address.Display, "()")
					}
					if isExtension {
						phoneNumber["extension"] = strings.Trim(*address.Display, "()")
					}

					if !isNumber && !isExtension {
						if address.Extension == nil {
							phoneNumber["extension"] = strings.Trim(*address.Display, "()")
						} else if phoneNumber["number"] != "" {
							phoneNumber["number"] = strings.Trim(*address.Display, "()")
						}
					}
				}

				if address.Extension != nil {
					phoneNumber["extension"] = *address.Extension
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

	for sdkType, schemaType := range util.GetUtilizationMediaTypes() {
		mediaSettings := allUtilizations[schemaType].([]interface{})
		if mediaSettings != nil && len(mediaSettings) > 0 {
			settings[sdkType] = util.BuildSdkMediaUtilization(mediaSettings)
		}
	}

	return &settings
}

// Basic user with minimum required fields
func GenerateBasicUserResource(resourceID string, email string, name string) string {
	return GenerateUserResource(resourceID, email, name, util.NullValue, util.NullValue, util.NullValue, util.NullValue, util.NullValue, "", "")
}

func GenerateUserResource(resourceID string, email string, name string, state string, title string, department string, manager string, acdAutoAnswer string, profileSkills string, certifications string) string {
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
	`, resourceName, resourceID, email, name, state, title, department, manager, acdAutoAnswer, profileSkills, certifications)
}
