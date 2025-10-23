package user

import (
	"context"
	"fmt"
	"log"
	"net/mail"
	"os"
	"sort"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	chunksProcess "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/chunks"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	frameworkdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// buildSdkAddresses removed - SDKv2 function no longer needed

// buildSdkAddressesFromFramework converts Framework types.List addresses to SDK Contact slice
func buildSdkAddressesFromFramework(addresses types.List) (*[]platformclientv2.Contact, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics
	sdkAddresses := make([]platformclientv2.Contact, 0)

	if addresses.IsNull() || addresses.IsUnknown() {
		return &sdkAddresses, diagnostics
	}

	addressElements := addresses.Elements()
	if len(addressElements) == 0 {
		return &sdkAddresses, diagnostics
	}

	// Get the first (and only) address element since MaxItems is 1
	addressObj, ok := addressElements[0].(types.Object)
	if !ok {
		diagnostics.AddError("Invalid Address Type", "Expected address to be an object")
		return nil, diagnostics
	}

	addressAttrs := addressObj.Attributes()

	// Process other_emails
	if otherEmailsAttr, exists := addressAttrs["other_emails"]; exists && !otherEmailsAttr.IsNull() {
		otherEmailsSet, ok := otherEmailsAttr.(types.Set)
		if ok {
			emailContacts, emailDiags := buildSdkEmailsFromFramework(otherEmailsSet)
			diagnostics.Append(emailDiags...)
			if !diagnostics.HasError() {
				sdkAddresses = append(sdkAddresses, emailContacts...)
			}
		}
	}

	// Process phone_numbers
	if phoneNumbersAttr, exists := addressAttrs["phone_numbers"]; exists && !phoneNumbersAttr.IsNull() {
		phoneNumbersSet, ok := phoneNumbersAttr.(types.Set)
		if ok {
			phoneContacts, phoneDiags := buildSdkPhoneNumbersFromFramework(phoneNumbersSet)
			diagnostics.Append(phoneDiags...)
			if !diagnostics.HasError() {
				sdkAddresses = append(sdkAddresses, phoneContacts...)
			}
		}
	}

	return &sdkAddresses, diagnostics
}

// buildSdkEmailsFromFramework converts Framework types.Set emails to SDK Contact slice
func buildSdkEmailsFromFramework(configEmails types.Set) ([]platformclientv2.Contact, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics
	emailElements := configEmails.Elements()
	sdkContacts := make([]platformclientv2.Contact, 0, len(emailElements))

	for _, emailElement := range emailElements {
		emailObj, ok := emailElement.(types.Object)
		if !ok {
			diagnostics.AddError("Invalid Email Type", "Expected email to be an object")
			continue
		}

		emailAttrs := emailObj.Attributes()

		var emailAddress, emailType string

		if addressAttr, exists := emailAttrs["address"]; exists && !addressAttr.IsNull() {
			emailAddress = addressAttr.(types.String).ValueString()
		}

		if typeAttr, exists := emailAttrs["type"]; exists && !typeAttr.IsNull() {
			emailType = typeAttr.(types.String).ValueString()
		} else {
			emailType = "WORK" // Default value
		}

		contactTypeEmail := "EMAIL"
		sdkContacts = append(sdkContacts, platformclientv2.Contact{
			Address:   &emailAddress,
			MediaType: &contactTypeEmail,
			VarType:   &emailType,
		})
	}

	return sdkContacts, diagnostics
}

// buildSdkPhoneNumbersFromFramework converts Framework types.Set phone numbers to SDK Contact slice
func buildSdkPhoneNumbersFromFramework(configPhoneNumbers types.Set) ([]platformclientv2.Contact, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics
	phoneElements := configPhoneNumbers.Elements()
	sdkContacts := make([]platformclientv2.Contact, 0, len(phoneElements))

	for _, phoneElement := range phoneElements {
		phoneObj, ok := phoneElement.(types.Object)
		if !ok {
			diagnostics.AddError("Invalid Phone Number Type", "Expected phone number to be an object")
			continue
		}

		phoneAttrs := phoneObj.Attributes()

		var phoneNumber, phoneMediaType, phoneType, phoneExt string

		if numberAttr, exists := phoneAttrs["number"]; exists && !numberAttr.IsNull() {
			phoneNumber = numberAttr.(types.String).ValueString()
		}

		if mediaTypeAttr, exists := phoneAttrs["media_type"]; exists && !mediaTypeAttr.IsNull() {
			phoneMediaType = mediaTypeAttr.(types.String).ValueString()
		} else {
			phoneMediaType = "PHONE" // Default value
		}

		if typeAttr, exists := phoneAttrs["type"]; exists && !typeAttr.IsNull() {
			phoneType = typeAttr.(types.String).ValueString()
		} else {
			phoneType = "WORK" // Default value
		}

		if extAttr, exists := phoneAttrs["extension"]; exists && !extAttr.IsNull() {
			phoneExt = extAttr.(types.String).ValueString()
		}

		// Note: extension_pool_id is handled by the API and not sent in the request

		contact := platformclientv2.Contact{
			MediaType: &phoneMediaType,
			VarType:   &phoneType,
		}

		if phoneNumber != "" {
			contact.Address = &phoneNumber
		}
		if phoneExt != "" {
			contact.Extension = &phoneExt
		}

		sdkContacts = append(sdkContacts, contact)
	}

	return sdkContacts, diagnostics
}

// executeUpdateUser removed - SDKv2 function no longer needed

// executeAllUpdates removed - SDKv2 function no longer needed

// updateUserSkills removed - SDKv2 function no longer needed

// updateUserLanguages removed - SDKv2 function no longer needed

// updateUserProfileSkills removed - SDKv2 function no longer needed

// updateUserVoicemailPolicies removed - SDKv2 function no longer needed

// updateUserRoutingUtilization removed - SDKv2 function no longer needed

// updatePassword removed - SDKv2 function no longer needed

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

// restoreDeletedUser removed - SDKv2 function no longer needed

// readUserRoutingUtilization removed - SDKv2 function no longer needed

// phoneNumberHash removed - SDKv2 function no longer needed

// buildSdkEmails removed - SDKv2 function replaced with buildSdkEmailsFromFramework

// buildSdkPhoneNumbers removed - SDKv2 function replaced with buildSdkPhoneNumbersFromFramework

// buildSdkLocations removed - SDKv2 function replaced with buildSdkLocationsFromFramework

// buildSdkEmployerInfo removed - SDKv2 function replaced with buildSdkEmployerInfoFromFramework

// buildSdkCertifications removed - SDKv2 function replaced with buildSdkCertificationsFromFramework

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

// flattenUserAddresses removed - SDKv2 function replaced with flattenUserAddressesForFramework

// flattenUserAddressesForFramework converts SDK Contact slice to Framework types.List
func flattenUserAddressesForFramework(ctx context.Context, addresses *[]platformclientv2.Contact, proxy *userProxy) (types.List, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics

	if addresses == nil || len(*addresses) == 0 {
		// If no addresses from API, return empty address object to match plan structure
		emptyPhoneSet, phoneSetDiags := types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"number":            types.StringType,
				"media_type":        types.StringType,
				"type":              types.StringType,
				"extension":         types.StringType,
				"extension_pool_id": types.StringType,
			},
		}, []attr.Value{})
		diagnostics.Append(phoneSetDiags...)

		emptyEmailSet, emailSetDiags := types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"address": types.StringType,
				"type":    types.StringType,
			},
		}, []attr.Value{})
		diagnostics.Append(emailSetDiags...)

		addressObj, addressObjDiags := types.ObjectValue(map[string]attr.Type{
			"other_emails": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"address": types.StringType,
						"type":    types.StringType,
					},
				},
			},
			"phone_numbers": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"number":            types.StringType,
						"media_type":        types.StringType,
						"type":              types.StringType,
						"extension":         types.StringType,
						"extension_pool_id": types.StringType,
					},
				},
			},
		}, map[string]attr.Value{
			"other_emails":  emptyEmailSet,
			"phone_numbers": emptyPhoneSet,
		})
		diagnostics.Append(addressObjDiags...)

		addressList, listDiags := types.ListValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"other_emails": types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"address": types.StringType,
							"type":    types.StringType,
						},
					},
				},
				"phone_numbers": types.SetType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"number":            types.StringType,
							"media_type":        types.StringType,
							"type":              types.StringType,
							"extension":         types.StringType,
							"extension_pool_id": types.StringType,
						},
					},
				},
			},
		}, []attr.Value{addressObj})
		diagnostics.Append(listDiags...)

		return addressList, diagnostics
	}

	emailElements := make([]attr.Value, 0)
	phoneElements := make([]attr.Value, 0)

	for _, address := range *addresses {
		if address.MediaType != nil {
			if *address.MediaType == "SMS" || *address.MediaType == "PHONE" {
				// Create phone attributes with exact values expected by the plan
				phoneAttrs := map[string]attr.Value{
					"number":            types.StringNull(),
					"media_type":        types.StringValue("PHONE"), // Schema default
					"type":              types.StringValue("WORK"),  // Schema default
					"extension":         types.StringNull(),
					"extension_pool_id": types.StringNull(),
				}

				// Set the phone number from API
				if address.Address != nil {
					phoneAttrs["number"] = types.StringValue(*address.Address)
				}

				// Override media_type if API provides different value
				if address.MediaType != nil && *address.MediaType != "PHONE" {
					phoneAttrs["media_type"] = types.StringValue(*address.MediaType)
				}

				// Override type if API provides different value
				if address.VarType != nil && *address.VarType != "WORK" {
					phoneAttrs["type"] = types.StringValue(*address.VarType)
				}

				// Handle extension only if provided by API
				if address.Extension != nil {
					if address.Display != nil {
						if *address.Extension == *address.Display {
							extensionNum := strings.Trim(*address.Extension, "()")
							phoneAttrs["extension"] = types.StringValue(extensionNum)
							extensionPoolId := fetchExtensionPoolId(ctx, extensionNum, proxy)
							if extensionPoolId != "" {
								phoneAttrs["extension_pool_id"] = types.StringValue(extensionPoolId)
							}
						} else {
							phoneAttrs["extension"] = types.StringValue(*address.Extension)
							phoneAttrs["number"] = types.StringValue(*address.Display)
						}
					} else {
						phoneAttrs["extension"] = types.StringValue(*address.Extension)
					}
				} else if address.Address == nil && address.Display != nil {
					phoneAttrs["extension"] = types.StringValue(*address.Display)
				}
				// Otherwise extension remains null as initialized

				// Create phone object with attributes in schema order
				phoneObj, objDiags := types.ObjectValue(map[string]attr.Type{
					"number":            types.StringType,
					"media_type":        types.StringType,
					"type":              types.StringType,
					"extension":         types.StringType,
					"extension_pool_id": types.StringType,
				}, phoneAttrs)
				diagnostics.Append(objDiags...)

				if !diagnostics.HasError() {
					phoneElements = append(phoneElements, phoneObj)
				}

			} else if *address.MediaType == "EMAIL" {
				emailAttrs := map[string]attr.Value{
					"address": types.StringValue(*address.Address),
					"type":    types.StringValue(*address.VarType),
				}

				emailObj, objDiags := types.ObjectValue(map[string]attr.Type{
					"address": types.StringType,
					"type":    types.StringType,
				}, emailAttrs)
				diagnostics.Append(objDiags...)

				if !diagnostics.HasError() {
					emailElements = append(emailElements, emailObj)
				}
			}
		}
	}

	// Create sets for emails and phone numbers
	emailSet, emailSetDiags := types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"address": types.StringType,
			"type":    types.StringType,
		},
	}, emailElements)
	diagnostics.Append(emailSetDiags...)

	phoneSet, phoneSetDiags := types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"number":            types.StringType,
			"media_type":        types.StringType,
			"type":              types.StringType,
			"extension":         types.StringType,
			"extension_pool_id": types.StringType,
		},
	}, phoneElements)
	diagnostics.Append(phoneSetDiags...)

	// Create the address object
	addressAttrs := map[string]attr.Value{
		"other_emails":  emailSet,
		"phone_numbers": phoneSet,
	}

	addressObj, addressObjDiags := types.ObjectValue(map[string]attr.Type{
		"other_emails": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"address": types.StringType,
					"type":    types.StringType,
				},
			},
		},
		"phone_numbers": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"number":            types.StringType,
					"media_type":        types.StringType,
					"type":              types.StringType,
					"extension":         types.StringType,
					"extension_pool_id": types.StringType,
				},
			},
		},
	}, addressAttrs)
	diagnostics.Append(addressObjDiags...)

	// Create the list with the single address object
	addressList, listDiags := types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"other_emails": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"address": types.StringType,
						"type":    types.StringType,
					},
				},
			},
			"phone_numbers": types.SetType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"number":            types.StringType,
						"media_type":        types.StringType,
						"type":              types.StringType,
						"extension":         types.StringType,
						"extension_pool_id": types.StringType,
					},
				},
			},
		},
	}, []attr.Value{addressObj})
	diagnostics.Append(listDiags...)

	return addressList, diagnostics
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

// SDKv2 flattening functions removed - replaced with Framework equivalents:
// - flattenUserSkills -> flattenUserSkillsForFramework
// - flattenUserLanguages -> flattenUserLanguagesForFramework
// - flattenUserLocations -> flattenUserLocationsForFramework
// - flattenUserData -> flattenUserDataForFramework

// flattenUserSkillsForFramework converts SDK routing skills to Framework types.Set
func flattenUserSkillsForFramework(skills *[]platformclientv2.Userroutingskill) (types.Set, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics

	if skills == nil || len(*skills) == 0 {
		return types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"skill_id":    types.StringType,
				"proficiency": types.Float64Type,
			},
		}), diagnostics
	}

	skillElements := make([]attr.Value, 0, len(*skills))

	for _, sdkSkill := range *skills {
		skillAttrs := map[string]attr.Value{
			"skill_id":    types.StringValue(*sdkSkill.Id),
			"proficiency": types.Float64Value(*sdkSkill.Proficiency),
		}

		skillObj, objDiags := types.ObjectValue(map[string]attr.Type{
			"skill_id":    types.StringType,
			"proficiency": types.Float64Type,
		}, skillAttrs)
		diagnostics.Append(objDiags...)

		if !diagnostics.HasError() {
			skillElements = append(skillElements, skillObj)
		}
	}

	skillSet, setDiags := types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"skill_id":    types.StringType,
			"proficiency": types.Float64Type,
		},
	}, skillElements)
	diagnostics.Append(setDiags...)

	return skillSet, diagnostics
}

// flattenUserLanguagesForFramework converts SDK routing languages to Framework types.Set
func flattenUserLanguagesForFramework(languages *[]platformclientv2.Userroutinglanguage) (types.Set, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics

	if languages == nil || len(*languages) == 0 {
		return types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"language_id": types.StringType,
				"proficiency": types.Int64Type,
			},
		}), diagnostics
	}

	languageElements := make([]attr.Value, 0, len(*languages))

	for _, sdkLang := range *languages {
		languageAttrs := map[string]attr.Value{
			"language_id": types.StringValue(*sdkLang.Id),
			"proficiency": types.Int64Value(int64(*sdkLang.Proficiency)),
		}

		languageObj, objDiags := types.ObjectValue(map[string]attr.Type{
			"language_id": types.StringType,
			"proficiency": types.Int64Type,
		}, languageAttrs)
		diagnostics.Append(objDiags...)

		if !diagnostics.HasError() {
			languageElements = append(languageElements, languageObj)
		}
	}

	languageSet, setDiags := types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"language_id": types.StringType,
			"proficiency": types.Int64Type,
		},
	}, languageElements)
	diagnostics.Append(setDiags...)

	return languageSet, diagnostics
}

// buildSdkSkillsFromFramework converts Framework routing skills to SDK skills
func buildSdkSkillsFromFramework(routingSkills types.Set) ([]platformclientv2.Userroutingskillpost, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if routingSkills.IsNull() || routingSkills.IsUnknown() {
		return nil, diagnostics
	}

	skillElements := routingSkills.Elements()
	sdkSkills := make([]platformclientv2.Userroutingskillpost, 0, len(skillElements))

	for _, skillElement := range skillElements {
		skillObj, ok := skillElement.(types.Object)
		if !ok {
			diagnostics = append(diagnostics, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid Skill Type",
				Detail:   "Expected skill to be an object",
			})
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

		sdkSkills = append(sdkSkills, platformclientv2.Userroutingskillpost{
			Id:          &skillId,
			Proficiency: &proficiency,
		})
	}

	return sdkSkills, diagnostics
}

// buildSdkLanguagesFromFramework converts Framework routing languages to SDK languages
func buildSdkLanguagesFromFramework(routingLanguages types.Set) ([]platformclientv2.Userroutinglanguagepost, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if routingLanguages.IsNull() || routingLanguages.IsUnknown() {
		return nil, diagnostics
	}

	languageElements := routingLanguages.Elements()
	sdkLanguages := make([]platformclientv2.Userroutinglanguagepost, 0, len(languageElements))

	for _, languageElement := range languageElements {
		languageObj, ok := languageElement.(types.Object)
		if !ok {
			diagnostics = append(diagnostics, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid Language Type",
				Detail:   "Expected language to be an object",
			})
			continue
		}

		languageAttrs := languageObj.Attributes()

		var languageId string
		var proficiency float64

		if languageIdAttr, exists := languageAttrs["language_id"]; exists && !languageIdAttr.IsNull() {
			languageId = languageIdAttr.(types.String).ValueString()
		}

		if proficiencyAttr, exists := languageAttrs["proficiency"]; exists && !proficiencyAttr.IsNull() {
			proficiency = float64(proficiencyAttr.(types.Int64).ValueInt64())
		}

		sdkLanguages = append(sdkLanguages, platformclientv2.Userroutinglanguagepost{
			Id:          &languageId,
			Proficiency: &proficiency,
		})
	}

	return sdkLanguages, diagnostics
}

// flattenUserLocationsForFramework converts SDK locations to Framework types.Set
func flattenUserLocationsForFramework(locations *[]platformclientv2.Location) (types.Set, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics

	if locations == nil || len(*locations) == 0 {
		return types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"location_id": types.StringType,
				"notes":       types.StringType,
			},
		}), diagnostics
	}

	locationElements := make([]attr.Value, 0, len(*locations))

	for _, sdkLoc := range *locations {
		if sdkLoc.LocationDefinition != nil {
			locationAttrs := map[string]attr.Value{
				"location_id": types.StringValue(*sdkLoc.LocationDefinition.Id),
				"notes":       types.StringNull(),
			}

			if sdkLoc.Notes != nil {
				locationAttrs["notes"] = types.StringValue(*sdkLoc.Notes)
			}

			locationObj, objDiags := types.ObjectValue(map[string]attr.Type{
				"location_id": types.StringType,
				"notes":       types.StringType,
			}, locationAttrs)
			diagnostics.Append(objDiags...)

			if !diagnostics.HasError() {
				locationElements = append(locationElements, locationObj)
			}
		}
	}

	locationSet, setDiags := types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"location_id": types.StringType,
			"notes":       types.StringType,
		},
	}, locationElements)
	diagnostics.Append(setDiags...)

	return locationSet, diagnostics
}

// buildSdkRoutingUtilizationFromFramework converts Framework routing utilization to SDK format
func buildSdkRoutingUtilizationFromFramework(routingUtilization types.List) (*platformclientv2.Utilizationrequest, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics

	if routingUtilization.IsNull() || routingUtilization.IsUnknown() {
		return nil, diagnostics
	}

	// For now, return a basic implementation
	// This is a complex structure that would need full implementation in a later task
	utilizationRequest := &platformclientv2.Utilizationrequest{
		Utilization: &map[string]platformclientv2.Mediautilization{},
	}

	return utilizationRequest, diagnostics
}

// flattenUserRoutingUtilizationForFramework converts SDK routing utilization to Framework types.List
func flattenUserRoutingUtilizationForFramework(utilization interface{}) (types.List, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics

	// Define the media type object structure to match the schema
	mediaTypeObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"maximum_capacity":          types.Int64Type,
			"interruptible_media_types": types.SetType{ElemType: types.StringType},
			"include_non_acd":           types.BoolType,
		},
	}

	// Define the label utilization object structure
	labelUtilizationObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"label_id":               types.StringType,
			"maximum_capacity":       types.Int64Type,
			"interrupting_label_ids": types.SetType{ElemType: types.StringType},
		},
	}

	// Define the main routing utilization object structure to match the schema
	routingUtilizationObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"call":               types.ListType{ElemType: mediaTypeObjectType},
			"callback":           types.ListType{ElemType: mediaTypeObjectType},
			"chat":               types.ListType{ElemType: mediaTypeObjectType},
			"email":              types.ListType{ElemType: mediaTypeObjectType},
			"message":            types.ListType{ElemType: mediaTypeObjectType},
			"label_utilizations": types.ListType{ElemType: labelUtilizationObjectType},
		},
	}

	// For now, return null with the proper type structure
	// This would need to handle the agentUtilizationWithLabels structure and convert it properly
	return types.ListNull(routingUtilizationObjectType), diagnostics
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

// flattenVoicemailUserpolicies removed - SDKv2 function no longer needed

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
