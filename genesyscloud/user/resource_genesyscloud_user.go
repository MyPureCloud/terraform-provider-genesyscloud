package user

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
	"github.com/nyaruka/phonenumbers"
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

	users, err := userProxy.getAllUserScripts(ctx)

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

	//Fixme
	user, diagnostics, done := createUser(ctx, d, meta, usersAPI, createUser, email)
	if done {
		return diagnostics
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

	diagErr := updateUserSkills(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserLanguages(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserProfileSkills(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserRoutingUtilization(d, usersAPI)
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

	diagErr = updateUserSkills(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserLanguages(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserProfileSkills(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserRoutingUtilization(d, usersAPI)
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
		if len(configInfo) > 0 {
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

func updateUserSkills(d *schema.ResourceData, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
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
				_, resp, err := usersAPI.PutUserRoutingskillsBulk(d.Id(), sdkSkills)
				if err != nil {
					return resp, diag.Errorf("Failed to update skills for user %s: %s", d.Id(), err)
				}
				return nil, nil
			})
		}
	}
	return nil
}

func updateUserLanguages(d *schema.ResourceData, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
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

			oldSdkLangs, err := getUserRoutingLanguages(d.Id(), usersAPI)
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
					diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
						resp, err := usersAPI.DeleteUserRoutinglanguage(d.Id(), langID)
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
				if diagErr := updateUserRoutingLanguages(d.Id(), langsToAddOrUpdate, newLangProfs, usersAPI); diagErr != nil {
					return diagErr
				}
			}
			log.Printf("Languages updated for user %s", d.Get("email"))
		}
	}
	return nil
}

func getUserRoutingLanguages(userID string, api *platformclientv2.UsersApi) ([]platformclientv2.Userroutinglanguage, diag.Diagnostics) {
	const maxPageSize = 50

	var sdkLanguages []platformclientv2.Userroutinglanguage
	for pageNum := 1; ; pageNum++ {
		langs, _, err := api.GetUserRoutinglanguages(userID, maxPageSize, pageNum, "")
		if err != nil {
			return nil, diag.Errorf("Failed to query languages for user %s: %s", userID, err)
		}
		if langs == nil || langs.Entities == nil || len(*langs.Entities) == 0 {
			return sdkLanguages, nil
		}
		for _, language := range *langs.Entities {
			sdkLanguages = append(sdkLanguages, language)
		}
	}
}

func updateUserRoutingLanguages(
	userID string,
	langsToUpdate []string,
	langProfs map[string]int,
	api *platformclientv2.UsersApi) diag.Diagnostics {
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
				_, resp, err := api.PatchUserRoutinglanguagesBulk(userID, updateChunk)
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

func updateUserProfileSkills(d *schema.ResourceData, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	if d.HasChange("profile_skills") {
		if profileSkills := d.Get("profile_skills"); profileSkills != nil {
			profileSkills := lists.SetToStringList(profileSkills.(*schema.Set))
			diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
				_, resp, err := usersAPI.PutUserProfileskills(d.Id(), *profileSkills)
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

func updateUserRoutingUtilization(d *schema.ResourceData, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
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
				_, _, err := usersAPI.PutRoutingUserUtilization(d.Id(), platformclientv2.Utilization{
					Utilization: &sdkSettings,
				})
				if err != nil {
					return diag.Errorf("Failed to update Routing Utilization for user %s: %s", d.Id(), err)
				}
			} else {
				// Reset to org-wide defaults
				_, err := usersAPI.DeleteRoutingUserUtilization(d.Id())
				if err != nil {
					return diag.Errorf("Failed to delete Routing Utilization for user %s: %s", d.Id(), err)
				}
			}
		}
	}
	return nil
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

func flattenUserProfileSkills(skills *[]string) *schema.Set {
	if skills != nil {
		return lists.StringListToSet(*skills)
	}
	return nil
}

func flattenUserCertifications(certs *[]string) *schema.Set {
	if certs != nil {
		return lists.StringListToSet(*certs)
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

func flattenUtilizationSetting(settings platformclientv2.Mediautilization) []interface{} {
	settingsMap := make(map[string]interface{})
	if settings.MaximumCapacity != nil {
		settingsMap["maximum_capacity"] = *settings.MaximumCapacity
	}
	if settings.InterruptableMediaTypes != nil {
		settingsMap["interruptible_media_types"] = lists.StringListToSet(*settings.InterruptableMediaTypes)
	}
	if settings.IncludeNonAcd != nil {
		settingsMap["include_non_acd"] = *settings.IncludeNonAcd
	}
	return []interface{}{settingsMap}
}

func getSdkUtilizationTypes() []string {
	types := make([]string, 0, len(utilizationMediaTypes))
	for t := range utilizationMediaTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
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
