package user

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
	"github.com/nyaruka/phonenumbers"
	"log"
	"sort"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
)

func getSdkUtilizationTypes() []string {
	types := make([]string, 0, len(utilizationMediaTypes))
	for t := range utilizationMediaTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
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
