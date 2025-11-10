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
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	chunksProcess "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/chunks"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	frameworkdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

// buildSdkAddresses converts Framework types.List addresses to SDK Contact slice
func buildSdkAddresses(addresses types.List) (*[]platformclientv2.Contact, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics
	sdkAddresses := make([]platformclientv2.Contact, 0)

	if addresses.IsNull() || addresses.IsUnknown() {
		return &sdkAddresses, nil
	}

	addressElements := addresses.Elements()
	if len(addressElements) == 0 {
		return &sdkAddresses, nil
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
			emailContacts, emailDiags := buildSdkEmails(otherEmailsSet)
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
			phoneContacts, phoneDiags := buildSdkPhoneNumbers(phoneNumbersSet)
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

		// Validate email type (matches SDK behavior)
		if emailTypeValidationDiags := validateEmailType(emailType); emailTypeValidationDiags.HasError() {
			diagnostics.Append(emailTypeValidationDiags...)
			continue // Skip this email if validation fails
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

	for i, phoneElement := range phoneElements {
		phoneObj, ok := phoneElement.(types.Object)
		if !ok {
			diagnostics.AddError("Invalid Phone Number Type", "Expected phone number to be an object")
			continue
		}

		phoneAttrs := phoneObj.Attributes()

		var phoneNumber, phoneMediaType, phoneType, phoneExt, poolID string
		var numberIsNull, extIsNull, poolIsNull bool = true, true, true

		if numberAttr, exists := phoneAttrs["number"]; exists && !numberAttr.IsNull() {
			phoneNumber = numberAttr.(types.String).ValueString()
			numberIsNull = false
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
			extIsNull = false
		}

		if poolAttr, exists := phoneAttrs["extension_pool_id"]; exists && !poolAttr.IsNull() {
			poolID = poolAttr.(types.String).ValueString()
			poolIsNull = false
		}

		log.Printf("[INV] phone[%d](plan): number=%q (null=%v) ext=%q (null=%v) pool=%q (null=%v) type=%q media=%q",
			i, phoneNumber, numberIsNull, phoneExt, extIsNull, poolID, poolIsNull, phoneType, phoneMediaType)

		// Validate phone number format (matches SDK behavior)
		if phoneValidationDiags := validatePhoneNumberE164(phoneNumber); phoneValidationDiags.HasError() {
			diagnostics.Append(phoneValidationDiags...)
			continue // Skip this phone number if validation fails
		}

		// Validate media type (matches SDK behavior)
		if mediaTypeValidationDiags := validatePhoneMediaType(phoneMediaType); mediaTypeValidationDiags.HasError() {
			diagnostics.Append(mediaTypeValidationDiags...)
			continue // Skip this phone number if validation fails
		}

		// Validate phone type (matches SDK behavior)
		if typeValidationDiags := validatePhoneType(phoneType); typeValidationDiags.HasError() {
			diagnostics.Append(typeValidationDiags...)
			continue // Skip this phone number if validation fails
		}

		// Convert Go strings to pointers for SDK
		var numPtr *string
		if phoneNumber != "" {
			numPtr = &phoneNumber
		}
		var extPtr *string
		if phoneExt != "" {
			extPtr = &phoneExt
		}
		var poolPtr *string
		if poolID != "" {
			poolPtr = &poolID
		}
		var typePtr *string
		if phoneType != "" {
			typePtr = &phoneType
		}
		var mediaPtr *string
		if phoneMediaType != "" {
			mediaPtr = &phoneMediaType
		}

		// Canonicalize: "" -> nil (ensures we never send empty strings to API)
		numPtr = canonPtrStr(numPtr)
		extPtr = canonPtrStr(extPtr)
		poolPtr = canonPtrStr(poolPtr)
		typePtr = canonPtrStr(typePtr)
		mediaPtr = canonPtrStr(mediaPtr)

		log.Printf("[INV] phone[%d](after canonicalization): number=%q ext=%q pool=%q type=%q media=%q",
			i, invStr(numPtr), invStr(extPtr), invStr(poolPtr), invStr(typePtr), invStr(mediaPtr))

		// Build SDK contact using canonicalized pointers (nil means absent)
		contact := platformclientv2.Contact{
			MediaType: mediaPtr,
			VarType:   typePtr,
			Address:   numPtr, // <- nil for extension-only or truly absent
			Extension: extPtr, // <- nil if not set
		}

		log.Printf("[INV] phone(SDK object): media=%q type=%q number=%q ext=%q",
			invStr(contact.MediaType), invStr(contact.VarType), invStr(contact.Address), invStr(contact.Extension))

		sdkContacts = append(sdkContacts, contact)
	}

	invDumpSDKPhones("buildSdkPhoneNumbersFromFramework result", sdkContacts)
	return sdkContacts, diagnostics
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
func flattenUserAddressesForFramework(ctx context.Context, addresses *[]platformclientv2.Contact, proxy *userProxy, configuredAddresses types.List) (types.List, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics

	// DEBUG: Add unique identifier for this call
	callID := fmt.Sprintf("%d", time.Now().UnixNano()%10000)
	log.Printf("DEBUG: [%s] flattenUserAddressesForFramework called", callID)

	// Log incoming contacts at entry with identity strings
	if addresses != nil {
		log.Printf("[INV] FLATTEN: incoming SDK addresses contacts=%d", len(*addresses))
		for i, c := range *addresses {
			log.Printf("[INV]   in[%d] %s (display=%q)", i, invSDKPhoneIdentity(c), invStr(c.Display))
		}
	}

	// DEBUG: Log what the API returns
	if addresses != nil {
		log.Printf("DEBUG: [%s] API returned %d addresses", callID, len(*addresses))
		log.Printf("[INV] flatten: API returned addresses JSON: %s", mustJSON(addresses))
		for i, addr := range *addresses {
			var mediaType, address, varType, extension, display string
			if addr.MediaType != nil {
				mediaType = *addr.MediaType
			}
			if addr.Address != nil {
				address = *addr.Address
			}
			if addr.VarType != nil {
				varType = *addr.VarType
			}
			if addr.Extension != nil {
				extension = *addr.Extension
			}
			if addr.Display != nil {
				display = *addr.Display
			}
			log.Printf("DEBUG: [%s] Address[%d]: MediaType='%s', Address='%s', VarType='%s', Extension='%s', Display='%s'",
				callID, i, mediaType, address, varType, extension, display)

			// Log pattern matching for debugging
			if mediaType == "PHONE" || mediaType == "SMS" {
				addressNil := address == ""
				extensionNil := extension == ""
				displayNil := display == ""
				log.Printf("DEBUG: [%s] Address[%d] Pattern Analysis: Address=nil:%v, Extension=nil:%v, Display=nil:%v",
					callID, i, addressNil, extensionNil, displayNil)
			}
		}
	} else {
		log.Printf("DEBUG: [%s] API returned nil addresses", callID)
	}

	if addresses == nil || len(*addresses) == 0 {
		// Return empty list for ListNestedAttribute schema
		return types.ListValueMust(
			types.ObjectType{
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
			},
			[]attr.Value{},
		), diagnostics
	}

	emailElements := make([]attr.Value, 0)
	phoneElements := make([]attr.Value, 0)

	// Check if other_emails was configured in the original plan/config
	// Default to true for backward compatibility (e.g., during import or when config context is unavailable)
	otherEmailsConfigured := true

	if !configuredAddresses.IsNull() && !configuredAddresses.IsUnknown() {
		configElements := configuredAddresses.Elements()
		if len(configElements) > 0 {
			if configObj, ok := configElements[0].(types.Object); ok {
				configAttrs := configObj.Attributes()
				if otherEmailsAttr, exists := configAttrs["other_emails"]; exists {
					log.Printf("DEBUG: other_emails attribute exists in config, IsNull: %v", otherEmailsAttr.IsNull())
					if !otherEmailsAttr.IsNull() {
						otherEmailsConfigured = true
						log.Printf("DEBUG: other_emails was configured in plan")
					} else {
						otherEmailsConfigured = false
						log.Printf("DEBUG: other_emails exists but is null in plan")
					}
				} else {
					otherEmailsConfigured = false
					log.Printf("DEBUG: other_emails attribute does NOT exist in plan")
				}
			}
		} else {
			// No address configuration at all - default to not configured for other_emails
			otherEmailsConfigured = false
			log.Printf("DEBUG: No addresses configured - other_emails will be null")
		}
	} else {
		// Config context not available (e.g., import) - use API response (backward compatibility)
		log.Printf("DEBUG: Config context not available - using API response for other_emails (backward compatibility)")
	}

	for i, address := range *addresses {
		if address.MediaType != nil {
			if *address.MediaType == "SMS" || *address.MediaType == "PHONE" {
				// Initialize phone attributes following SDK pattern exactly
				phoneNumber := map[string]interface{}{
					"number":            nil,
					"media_type":        *address.MediaType,
					"type":              nil,
					"extension":         nil,
					"extension_pool_id": "",
				}

				// SDK Pattern matching - exact same logic as SDK with proper if/else if structure
				if address.Address != nil {
					// Pattern 1: Address field contains phone number
					utilE164 := util.NewUtilE164Service()
					phoneNumber["number"] = utilE164.FormatAsCalculatedE164Number(strings.Trim(*address.Address, "()"))
					log.Printf("DEBUG: [%s] Address[%d] SDK Pattern 1 - Address: '%s'", callID, i, *address.Address)
				} else if address.Extension != nil && address.Display != nil {
					if *address.Extension == *address.Display {
						// Pattern 2: Extension == Display (true extension with extension pool)
						extensionNum := strings.Trim(*address.Extension, "()")
						phoneNumber["extension"] = extensionNum
						phoneNumber["extension_pool_id"] = fetchExtensionPoolId(ctx, extensionNum, proxy)
						log.Printf("DEBUG: [%s] Address[%d] SDK Pattern 2 - Extension==Display: '%s'", callID, i, extensionNum)
					} else {
						// Pattern 3: Extension != Display (phone number + extension)
						phoneNumber["extension"] = *address.Extension
						utilE164 := util.NewUtilE164Service()
						phoneNumber["number"] = utilE164.FormatAsCalculatedE164Number(strings.Trim(*address.Display, "()"))
						log.Printf("DEBUG: [%s] Address[%d] SDK Pattern 3 - Extension!=Display: number='%s', ext='%s'", callID, i, *address.Display, *address.Extension)
					}
				} else if address.Address == nil && address.Extension == nil && address.Display != nil {
					// Pattern 4: Extension-only case (Address and Extension nil, Display has value)
					phoneNumber["extension"] = strings.Trim(*address.Display, "()")
					log.Printf("DEBUG: [%s] Address[%d] SDK Pattern 4 - Extension-only: '%s'", callID, i, *address.Display)
				}

				// Set type from API response
				if address.VarType != nil {
					phoneNumber["type"] = *address.VarType
				}

				// Canonicalize SDK values: "" -> nil (ensures state never contains empty strings)
				// Intentionally ignore address.Display - it's decoration, not identity
				var numStr, extStr, poolStr, typeStr *string
				if phoneNumber["number"] != nil {
					s := phoneNumber["number"].(string)
					numStr = &s
				}
				if phoneNumber["extension"] != nil {
					s := phoneNumber["extension"].(string)
					extStr = &s
				}
				if phoneNumber["extension_pool_id"] != nil && phoneNumber["extension_pool_id"] != "" {
					s := phoneNumber["extension_pool_id"].(string)
					poolStr = &s
				}
				if phoneNumber["type"] != nil {
					s := phoneNumber["type"].(string)
					typeStr = &s
				}

				// Canonicalize: "" -> nil
				numStr = canonPtrStr(numStr)
				extStr = canonPtrStr(extStr)
				poolStr = canonPtrStr(poolStr)
				typeStr = canonPtrStr(typeStr)
				mediaStr := canonPtrStr(address.MediaType)

				log.Printf("[INV] phone[%d](after canonicalization): number=%q ext=%q pool=%q type=%q media=%q",
					i, invStr(numStr), invStr(extStr), invStr(poolStr), invStr(typeStr), invStr(mediaStr))

				// Map to Framework tri-state strings (NULL for nil/empty)
				fNumber := fStringFromPtr(numStr)
				fExtension := fStringFromPtr(extStr)
				fPool := fStringFromPtr(poolStr)
				fType := fStringFromPtr(typeStr)
				fMedia := fStringFromPtr(mediaStr)

				// Extension-only normalization check
				if fNumber.IsNull() && !fExtension.IsNull() {
					log.Printf("[INV] Normalize EXT-ONLY: unset number for type=%s ext=%s",
						invTriString(fType), invTriString(fExtension))
				}

				// Build the phone object attribute map
				phoneAttrs := map[string]attr.Value{
					"media_type":        fMedia,
					"type":              fType,
					"number":            fNumber,
					"extension":         fExtension,
					"extension_pool_id": fPool,
				}

				// Log tri-state conversion for each Framework field
				log.Printf("[INV]   conv -> Framework tri-state: media=%s type=%s number=%s ext=%s pool=%s",
					invTriString(phoneAttrs["media_type"].(types.String)),
					invTriString(phoneAttrs["type"].(types.String)),
					invTriString(phoneAttrs["number"].(types.String)),
					invTriString(phoneAttrs["extension"].(types.String)),
					invTriString(phoneAttrs["extension_pool_id"].(types.String)))

				// Log phone identity for debugging
				log.Printf("[INV] FLATTEN phone identity -> %s",
					phoneIdentity(fMedia, fType, fNumber, fExtension))

				// Create phone object
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
					log.Printf("DEBUG: [%s] Address[%d] Added phone element (total: %d)", callID, i, len(phoneElements))
				}

			} else if *address.MediaType == "EMAIL" {
				log.Printf("DEBUG: Processing EMAIL address '%s' with type '%s', otherEmailsConfigured: %v", *address.Address, *address.VarType, otherEmailsConfigured)
				// Only process EMAIL addresses if other_emails was configured
				if otherEmailsConfigured {
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
						log.Printf("DEBUG: Added EMAIL to emailElements, total count: %d", len(emailElements))
					}
				} else {
					log.Printf("DEBUG: Skipping EMAIL address '%s' because other_emails not configured", *address.Address)
				}
			} else {
				log.Printf("Unknown address media type %s", *address.MediaType)
			}
		}
	}

	// Create sets for emails and phone numbers
	var emailSet types.Set
	var emailSetDiags frameworkdiag.Diagnostics

	if otherEmailsConfigured {
		// If other_emails was configured, create set with actual elements (may be empty)
		emailSet, emailSetDiags = types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"address": types.StringType,
				"type":    types.StringType,
			},
		}, emailElements)
		log.Printf("DEBUG: Created other_emails set with %d elements (configured), diagnostics: %v", len(emailElements), emailSetDiags)
		if emailSetDiags.HasError() {
			log.Printf("DEBUG: ERROR creating email set: %v", emailSetDiags)
		}
	} else {
		// If other_emails was not configured, return null set
		emailSet = types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"address": types.StringType,
				"type":    types.StringType,
			},
		})
		log.Printf("DEBUG: Created null other_emails set (not configured)")
	}
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

	// Define consistent address object type for ListNestedAttribute schema compatibility
	addressObjectType := types.ObjectType{
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
	}

	addressObj, addressObjDiags := types.ObjectValue(addressObjectType.AttrTypes, addressAttrs)
	diagnostics.Append(addressObjDiags...)

	// DEBUG: Log final counts
	log.Printf("DEBUG: [%s] Final result - phoneElements: %d, emailElements: %d", callID, len(phoneElements), len(emailElements))

	// Create the list with the single address object - compatible with ListNestedAttribute
	addressList, listDiags := types.ListValue(addressObjectType, []attr.Value{addressObj})
	diagnostics.Append(listDiags...)

	log.Printf("DEBUG: Returning address list with %d elements", len([]attr.Value{addressObj}))

	// Log flatten result with detailed phone set information
	log.Printf("[INV] FLATTEN result: phoneSet.elems=%d emailSet.elems=%d", len(phoneElements), len(emailElements))

	// Dump each phone object as JSON for full visibility
	for i, pe := range phoneElements {
		log.Printf("[INV]   phoneSet[%d]=%s", i, invMustJSON(pe))
	}

	// Log phone identities for easy comparison
	for i, pe := range phoneElements {
		obj := pe.(types.Object)
		m := obj.Attributes()
		log.Printf("[INV] FLATTEN phone[%d] -> %s",
			i,
			phoneIdentity(
				m["media_type"].(types.String),
				m["type"].(types.String),
				m["number"].(types.String),
				m["extension"].(types.String),
			),
		)
	}

	// Dump the final addresses list value that will be returned
	log.Printf("[INV] FLATTEN result (Framework addresses list)=%s", invMustJSON(addressList))

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

	// Handle null utilization
	if utilization == nil {
		log.Printf("DEBUG: Utilization is nil, returning null")
		log.Printf("[INV] RU FLATTEN: utilization is nil, returning null parent")
		return types.ListNull(routingUtilizationObjectType), diagnostics
	}

	log.Printf("DEBUG: Flattening utilization of type: %T", utilization)
	log.Printf("[INV] RU FLATTEN: building parent block")

	// Try to convert to the expected type
	agentUtil, ok := utilization.(*platformclientv2.Agentmaxutilizationresponse)
	if !ok {
		log.Printf("DEBUG: Unexpected utilization type: %T", utilization)
		log.Printf("[INV] RU FLATTEN: unexpected type, returning null parent")
		return types.ListNull(routingUtilizationObjectType), diagnostics
	}

	// Check if utilization is at organization level (no user-specific settings)
	if agentUtil.Utilization == nil {
		log.Printf("DEBUG: No user-specific utilization settings (org level)")
		log.Printf("[INV] RU FLATTEN: org level (no user settings), returning null parent")
		return types.ListNull(routingUtilizationObjectType), diagnostics
	}

	// Build the utilization object
	utilizationAttrs := map[string]attr.Value{
		"call":               types.ListNull(mediaTypeObjectType),
		"callback":           types.ListNull(mediaTypeObjectType),
		"chat":               types.ListNull(mediaTypeObjectType),
		"email":              types.ListNull(mediaTypeObjectType),
		"message":            types.ListNull(mediaTypeObjectType),
		"label_utilizations": types.ListNull(labelUtilizationObjectType),
	}

	// Handle call utilization if present
	if agentUtil.Utilization != nil {
		if callUtil, exists := (*agentUtil.Utilization)["call"]; exists {
			callAttrs := map[string]attr.Value{
				"maximum_capacity":          types.Int64Null(),
				"include_non_acd":           types.BoolNull(),
				"interruptible_media_types": types.SetNull(types.StringType),
			}

			if callUtil.MaximumCapacity != nil {
				callAttrs["maximum_capacity"] = types.Int64Value(int64(*callUtil.MaximumCapacity))
			}
			if callUtil.IncludeNonAcd != nil {
				callAttrs["include_non_acd"] = types.BoolValue(*callUtil.IncludeNonAcd)
			}
			if callUtil.InterruptableMediaTypes != nil && len(*callUtil.InterruptableMediaTypes) > 0 {
				interruptibleTypes := make([]attr.Value, len(*callUtil.InterruptableMediaTypes))
				for i, mediaType := range *callUtil.InterruptableMediaTypes {
					interruptibleTypes[i] = types.StringValue(mediaType)
				}
				callAttrs["interruptible_media_types"], _ = types.SetValue(types.StringType, interruptibleTypes)
				log.Printf("DEBUG: Setting interruptible_media_types: %v", *callUtil.InterruptableMediaTypes)
			} else {
				// Keep as null if not configured or empty - this preserves the original plan state
				log.Printf("DEBUG: Keeping interruptible_media_types as null (empty or not configured)")
			}

			callObj, _ := types.ObjectValue(mediaTypeObjectType.AttrTypes, callAttrs)
			utilizationAttrs["call"], _ = types.ListValue(mediaTypeObjectType, []attr.Value{callObj})

			log.Printf("DEBUG: Flattened call utilization - MaxCapacity: %v, IncludeNonAcd: %v",
				callUtil.MaximumCapacity, callUtil.IncludeNonAcd)
		}
	}

	utilizationObj, _ := types.ObjectValue(routingUtilizationObjectType.AttrTypes, utilizationAttrs)

	// Log media type presence
	log.Printf("[INV] RU FLATTEN: call=%s callback=%s chat=%s email=%s message=%s",
		invPresence(utilizationAttrs["call"].(types.List)),
		invPresence(utilizationAttrs["callback"].(types.List)),
		invPresence(utilizationAttrs["chat"].(types.List)),
		invPresence(utilizationAttrs["email"].(types.List)),
		invPresence(utilizationAttrs["message"].(types.List)))

	result, diags := types.ListValue(routingUtilizationObjectType, []attr.Value{utilizationObj})
	diagnostics.Append(diags...)

	log.Printf("[INV] RU FLATTEN result (parent present=%t) JSON=%s",
		!result.IsNull() && !result.IsUnknown(),
		invMustJSON(result))

	return result, diagnostics
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
		// Handle both framework ([]string) and SDK v2 (*schema.Set) types
		switch v := types.(type) {
		case []string:
			// Framework type - already a string slice
			interruptableMediaTypes = &v
		case *schema.Set:
			// SDK v2 type - convert from Set
			interruptableMediaTypes = lists.SetToStringList(v)
		case []interface{}:
			// Convert interface slice to string slice
			stringList := lists.InterfaceListToStrings(v)
			interruptableMediaTypes = &stringList
		}
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

		var interruptingLabelIds []string
		if ids, ok := labelUtilizationMap["interrupting_label_ids"]; ok && ids != nil {
			// Handle both framework ([]string) and SDK v2 (*schema.Set) types
			switch v := ids.(type) {
			case []string:
				// Framework type - already a string slice
				interruptingLabelIds = v
			case *schema.Set:
				// SDK v2 type - convert from Set
				if converted := lists.SetToStringList(v); converted != nil {
					interruptingLabelIds = *converted
				}
			case []interface{}:
				// Convert interface slice to string slice
				interruptingLabelIds = lists.InterfaceListToStrings(v)
			}
		}

		request[labelUtilizationMap["label_id"].(string)] = labelUtilization{
			MaximumCapacity:      int32(labelUtilizationMap["maximum_capacity"].(int)),
			InterruptingLabelIds: interruptingLabelIds,
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

// Framework Diagnostic Utility Functions
// These functions provide Framework-specific diagnostic conversion and error handling
// while preserving identical business logic to SDK equivalents

// convertSDKDiagnosticsToFramework converts SDK v2 diagnostics to Framework diagnostics
// while preserving identical error messages, severity levels, and diagnostic details
func convertSDKDiagnosticsToFramework(sdkDiags diag.Diagnostics) frameworkdiag.Diagnostics {
	var frameworkDiags frameworkdiag.Diagnostics

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

// buildFrameworkAPIDiagnosticError creates Framework diagnostics from API errors
// using util.BuildAPIDiagnosticError internally to maintain SDK behavior
func buildFrameworkAPIDiagnosticError(resourceType string, summary string, apiResponse *platformclientv2.APIResponse) frameworkdiag.Diagnostics {
	// Use SDK function internally to maintain identical behavior
	sdkDiags := util.BuildAPIDiagnosticError(resourceType, summary, apiResponse)

	// Convert to Framework format
	return convertSDKDiagnosticsToFramework(sdkDiags)
}

// handleFrameworkAPIError handles API errors and returns Framework diagnostics
// while preserving correlation IDs and status codes
func handleFrameworkAPIError(resourceType string, operation string, resourceId string, err error, apiResponse *platformclientv2.APIResponse) frameworkdiag.Diagnostics {
	summary := fmt.Sprintf("Failed to %s %s %s", operation, resourceType, resourceId)
	if err != nil {
		summary = fmt.Sprintf("%s error: %s", summary, err)
	}

	return buildFrameworkAPIDiagnosticError(resourceType, summary, apiResponse)
}

// FrameworkRetryWrapper wraps SDK retry logic for Framework compatibility
type FrameworkRetryWrapper struct {
	resourceType string
}

// newFrameworkRetryWrapper creates a retry wrapper for Framework operations
// using identical SDK retry logic
func newFrameworkRetryWrapper(resourceType string) *FrameworkRetryWrapper {
	return &FrameworkRetryWrapper{
		resourceType: resourceType,
	}
}

// executeWithRetry implements retry logic for Framework operations with SDK backoff patterns
func (w *FrameworkRetryWrapper) executeWithRetry(operation func() (*platformclientv2.APIResponse, error), errorMessage string) frameworkdiag.Diagnostics {
	// Use SDK retry logic internally to maintain identical behavior
	sdkDiags := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		apiResponse, err := operation()
		if err != nil {
			return apiResponse, util.BuildAPIDiagnosticError(w.resourceType, errorMessage, apiResponse)
		}
		return apiResponse, nil
	})

	// Convert to Framework format if there are errors
	if sdkDiags.HasError() {
		return convertSDKDiagnosticsToFramework(sdkDiags)
	}

	return frameworkdiag.Diagnostics{}
}

// withRetriesForReadFramework implements retry logic for Framework read operations
// with SDK backoff patterns
func withRetriesForReadFramework(ctx context.Context, retryFunc func() *retry.RetryError) frameworkdiag.Diagnostics {
	// Use SDK timeout values (5 minutes for read operations)
	timeout := 5 * time.Minute

	err := retry.RetryContext(ctx, timeout, retryFunc)
	if err != nil {
		var frameworkDiags frameworkdiag.Diagnostics
		if strings.Contains(fmt.Sprintf("%v", err), "API Error: 404") {
			// Handle 404 errors gracefully - this matches SDK behavior
			frameworkDiags.AddError("Resource Not Found", "The requested resource was not found")
		} else if util.IsTimeoutError(err) {
			frameworkDiags.AddError("Operation Timeout", fmt.Sprintf("Operation timed out after %v", timeout))
		} else {
			frameworkDiags.AddError("Operation Failed", fmt.Sprintf("Operation failed: %v", err))
		}
		return frameworkDiags
	}

	return frameworkdiag.Diagnostics{}
}

// ensureIdenticalTimeoutBehavior ensures Framework operations use identical SDK timeout values
// (5min read, 3min delete, 2min other)
func ensureIdenticalTimeoutBehavior(operationType string) time.Duration {
	switch strings.ToLower(operationType) {
	case "read":
		return 5 * time.Minute
	case "delete":
		return 3 * time.Minute
	default:
		return 2 * time.Minute
	}
}

// handleFrameworkTimeout handles timeout errors with SDK-equivalent messages
func handleFrameworkTimeout(resourceType string, operation string, resourceId string, timeout time.Duration) frameworkdiag.Diagnostics {
	var frameworkDiags frameworkdiag.Diagnostics
	summary := fmt.Sprintf("%s %s Timeout", strings.Title(operation), resourceType)
	detail := fmt.Sprintf("Operation timed out after %v for resource %s", timeout, resourceId)
	frameworkDiags.AddError(summary, detail)
	return frameworkDiags
}

// handleFrameworkValidationError handles validation errors preserving SDK error messages
func handleFrameworkValidationError(resourceType string, field string, value interface{}, validationMessage string) frameworkdiag.Diagnostics {
	var frameworkDiags frameworkdiag.Diagnostics
	summary := fmt.Sprintf("%s Validation Error", resourceType)
	detail := fmt.Sprintf("Validation failed for field '%s' with value '%v': %s", field, value, validationMessage)
	frameworkDiags.AddError(summary, detail)
	return frameworkDiags
}

// handleFramework404Error handles 404 errors with SDK behavior patterns
func handleFramework404Error(resourceType string, resourceId string) frameworkdiag.Diagnostics {
	var frameworkDiags frameworkdiag.Diagnostics
	summary := fmt.Sprintf("%s Not Found", resourceType)
	detail := fmt.Sprintf("The %s with ID '%s' was not found", strings.ToLower(resourceType), resourceId)
	frameworkDiags.AddError(summary, detail)
	return frameworkDiags
}

// Framework validation functions for addresses - integrated into build functions
// Validation is performed during the buildSdkAddressesFromFramework process to match SDK behavior

// validatePhoneNumberE164 validates that a phone number is in E.164 format
func validatePhoneNumberE164(phoneNumber string) frameworkdiag.Diagnostics {
	var diagnostics frameworkdiag.Diagnostics

	// Skip validation in BCP mode (matches SDK behavior)
	if feature_toggles.BcpModeEnabledExists() {
		return diagnostics
	}

	if phoneNumber == "" {
		return diagnostics // Empty phone numbers are allowed
	}

	utilE164 := util.NewUtilE164Service()
	validNum, diags := utilE164.IsValidE164Number(phoneNumber)
	if diags != nil {
		diagnostics.AddError(
			"Phone Number Validation Error",
			fmt.Sprintf("Failed to validate phone number: %v", diags),
		)
		return diagnostics
	}

	if !validNum {
		diagnostics.AddError(
			"Phone Number Validation Error",
			fmt.Sprintf("Phone number must be in E.164 format: %s", phoneNumber),
		)
	}

	return diagnostics
}

// validatePhoneMediaType validates phone media type values
func validatePhoneMediaType(mediaType string) frameworkdiag.Diagnostics {
	var diagnostics frameworkdiag.Diagnostics

	if mediaType == "" {
		return diagnostics // Empty values use defaults
	}

	validMediaTypes := []string{"PHONE", "SMS"}
	for _, valid := range validMediaTypes {
		if mediaType == valid {
			return diagnostics
		}
	}

	diagnostics.AddError(
		"Phone Media Type Validation Error",
		fmt.Sprintf("Media type must be one of %v, got: %s", validMediaTypes, mediaType),
	)
	return diagnostics
}

// validatePhoneType validates phone type values
func validatePhoneType(phoneType string) frameworkdiag.Diagnostics {
	var diagnostics frameworkdiag.Diagnostics

	if phoneType == "" {
		return diagnostics // Empty values use defaults
	}

	validTypes := []string{"WORK", "WORK2", "WORK3", "WORK4", "HOME", "MOBILE", "OTHER"}
	for _, valid := range validTypes {
		if phoneType == valid {
			return diagnostics
		}
	}

	diagnostics.AddError(
		"Phone Type Validation Error",
		fmt.Sprintf("Type must be one of %v, got: %s", validTypes, phoneType),
	)
	return diagnostics
}

// validateEmailType validates email type values
func validateEmailType(emailType string) frameworkdiag.Diagnostics {
	var diagnostics frameworkdiag.Diagnostics

	if emailType == "" {
		return diagnostics // Empty values use defaults
	}

	validTypes := []string{"WORK", "HOME"}
	for _, valid := range validTypes {
		if emailType == valid {
			return diagnostics
		}
	}

	diagnostics.AddError(
		"Email Type Validation Error",
		fmt.Sprintf("Type must be one of %v, got: %s", validTypes, emailType),
	)
	return diagnostics
}

// Migrated Build Functions from Framework Resource File
// These functions were moved from UserFrameworkResource methods to standalone functions
// to follow SDK organizational patterns while preserving identical business logic

// buildSdkLocations converts Framework locations to SDK locations
// Migrated from (r *UserFrameworkResource) buildSdkLocations method
func buildSdkLocations(ctx context.Context, locations types.Set) *[]platformclientv2.Location {
	if locations.IsNull() || locations.IsUnknown() {
		return nil
	}

	sdkLocations := make([]platformclientv2.Location, 0)
	for _, locVal := range locations.Elements() {
		locObj, ok := locVal.(basetypes.ObjectValue)
		if !ok {
			continue
		}
		locAttrs := locObj.Attributes()

		locID := locAttrs["location_id"].(basetypes.StringValue).ValueString()
		locNotes := locAttrs["notes"].(basetypes.StringValue).ValueString()

		sdkLocations = append(sdkLocations, platformclientv2.Location{
			Id:    &locID,
			Notes: &locNotes,
		})
	}
	return &sdkLocations
}

// buildSdkCertifications converts Framework certifications to SDK certifications
// Migrated from (r *UserFrameworkResource) buildSdkCertifications method
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

// buildSdkProfileSkills converts Framework profile skills to SDK profile skills
// Migrated from (r *UserFrameworkResource) buildSdkProfileSkills method
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

// buildSdkEmployerInfo converts Framework employer info to SDK employer info
// Migrated from (r *UserFrameworkResource) buildSdkEmployerInfo method
func buildSdkEmployerInfo(ctx context.Context, employerInfo types.List) *platformclientv2.Employerinfo {
	if employerInfo.IsNull() || employerInfo.IsUnknown() {
		// Return empty employer info object with nil pointers to clear existing data (matches SDK pattern)
		return &platformclientv2.Employerinfo{}
	}

	elements := employerInfo.Elements()
	if len(elements) == 0 {
		return nil
	}

	// Get the first (and only) element since MaxItems is 1
	empInfoObj, ok := elements[0].(basetypes.ObjectValue)
	if !ok {
		return nil
	}

	empAttrs := empInfoObj.Attributes()
	sdkEmployerInfo := &platformclientv2.Employerinfo{}

	if officialName, exists := empAttrs["official_name"]; exists && !officialName.IsNull() {
		if nameVal, ok := officialName.(basetypes.StringValue); ok && nameVal.ValueString() != "" {
			sdkEmployerInfo.OfficialName = platformclientv2.String(nameVal.ValueString())
		}
	}

	if employeeId, exists := empAttrs["employee_id"]; exists && !employeeId.IsNull() {
		if idVal, ok := employeeId.(basetypes.StringValue); ok && idVal.ValueString() != "" {
			sdkEmployerInfo.EmployeeId = platformclientv2.String(idVal.ValueString())
		}
	}

	if employeeType, exists := empAttrs["employee_type"]; exists && !employeeType.IsNull() {
		if typeVal, ok := employeeType.(basetypes.StringValue); ok && typeVal.ValueString() != "" {
			sdkEmployerInfo.EmployeeType = platformclientv2.String(typeVal.ValueString())
		}
	}

	if dateHire, exists := empAttrs["date_hire"]; exists && !dateHire.IsNull() {
		if dateVal, ok := dateHire.(basetypes.StringValue); ok && dateVal.ValueString() != "" {
			dateStr := dateVal.ValueString()
			sdkEmployerInfo.DateHire = &dateStr
		}
	}

	return sdkEmployerInfo
}

// flattenEmployerInfo converts SDK employer info to Framework employer info
// Migrated from (r *UserFrameworkResource) flattenEmployerInfo method
func flattenEmployerInfo(ctx context.Context, employerInfo *platformclientv2.Employerinfo) types.List {
	if employerInfo == nil {
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"official_name": types.StringType,
				"employee_id":   types.StringType,
				"employee_type": types.StringType,
				"date_hire":     types.StringType,
			},
		})
	}

	// Check if all employer info fields are empty/nil (cleared by API)
	isEmpty := (employerInfo.OfficialName == nil || *employerInfo.OfficialName == "") &&
		(employerInfo.EmployeeId == nil || *employerInfo.EmployeeId == "") &&
		(employerInfo.EmployeeType == nil || *employerInfo.EmployeeType == "") &&
		(employerInfo.DateHire == nil || *employerInfo.DateHire == "")

	if isEmpty {
		// Return null list when all fields are empty (matches test expectation)
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"official_name": types.StringType,
				"employee_id":   types.StringType,
				"employee_type": types.StringType,
				"date_hire":     types.StringType,
			},
		})
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

	empInfoObj, _ := types.ObjectValue(map[string]attr.Type{
		"official_name": types.StringType,
		"employee_id":   types.StringType,
		"employee_type": types.StringType,
		"date_hire":     types.StringType,
	}, empInfoAttrs)

	empInfoList, _ := types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"official_name": types.StringType,
			"employee_id":   types.StringType,
			"employee_type": types.StringType,
			"date_hire":     types.StringType,
		},
	}, []attr.Value{empInfoObj})

	return empInfoList
}

// flattenUserData converts SDK string slice to Framework Set
// Migrated from (r *UserFrameworkResource) flattenUserData method
func flattenUserData(ctx context.Context, userDataSlice *[]string) types.Set {
	elements := make([]attr.Value, 0)

	if userDataSlice != nil {
		for _, item := range *userDataSlice {
			elements = append(elements, types.StringValue(item))
		}
	}

	setVal, _ := types.SetValue(types.StringType, elements)
	return setVal
}

// executeUpdateUser executes a user update with retry logic
// Migrated from (r *UserFrameworkResource) executeUpdateUser method
// Following SDK pattern: mirrors SDK executeUpdateUser logic for Framework interface
func executeUpdateUser(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, updateUser platformclientv2.Updateuser) frameworkdiag.Diagnostics {
	// Use SDK-aligned retry logic with version mismatch handling
	retryWrapper := newFrameworkRetryWrapper(ResourceType)

	return retryWrapper.executeWithRetry(func() (*platformclientv2.APIResponse, error) {
		// Get current user version (matches SDK pattern)
		currentUser, proxyResponse, errGet := proxy.getUserById(ctx, plan.Id.ValueString(), nil, "")
		if errGet != nil {
			return proxyResponse, errGet
		}

		updateUser.Version = currentUser.Version

		updatedUser, proxyPatchResponse, patchErr := proxy.updateUser(ctx, plan.Id.ValueString(), &updateUser)
		if proxyPatchResponse != nil {
			log.Printf("[INV] UPDATE patch status: %v", proxyPatchResponse.StatusCode)
		}
		if patchErr == nil && updatedUser != nil {
			// Log the immediate PATCH response
			log.Printf("[INV] UPDATE PATCH response.Addresses=%s", invMustJSON(updatedUser.Addresses))
			if updatedUser.Addresses != nil {
				invDumpSDKPhones("UPDATE PATCH response (SDK)", *updatedUser.Addresses)
			}
		}
		return proxyPatchResponse, patchErr
	}, fmt.Sprintf("Failed to update user %s", plan.Id.ValueString()))
}

// executeAllUpdates executes all additional updates for the user
// Migrated from (r *UserFrameworkResource) executeAllUpdates method
// Following SDK pattern: mirrors SDK executeAllUpdates logic for Framework interface
func executeAllUpdates(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, sdkConfig *platformclientv2.Configuration, updateObjectDivision bool, state ...*UserFrameworkResourceModel) frameworkdiag.Diagnostics {
	var diagnostics frameworkdiag.Diagnostics

	if updateObjectDivision {
		// Implement Framework-compatible division update logic (Task 6.1, 6.2, 6.3, 6.4, 6.5)
		var currentState *UserFrameworkResourceModel
		if len(state) > 0 {
			currentState = state[0]
		}
		diagErr := updateUserDivision(ctx, plan, currentState, sdkConfig)
		if diagErr.HasError() {
			// Task 8.4: Return diagnostic messages without stopping other updates when division update fails
			for _, divisionDiag := range diagErr {
				diagnostics.AddWarning(
					"Division Update Failed",
					fmt.Sprintf("Division update failed for user %s but other updates will continue. %s: %s", plan.Id.ValueString(), divisionDiag.Summary(), divisionDiag.Detail()),
				)
			}
			log.Printf("Division update failed for user %s, continuing with other updates", plan.Id.ValueString())
		} else {
			log.Printf("Division update completed successfully for user %s", plan.Id.ValueString())
		}
	}

	// Update user skills - using standalone function from utils file
	diagErr := updateUserSkills(ctx, plan, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	// Update user languages - using standalone function from utils file
	diagErr = updateUserLanguages(ctx, plan, proxy)
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
	diagErr = updateUserVoicemailPolicies(ctx, plan, proxy)
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

// updateUserDivision handles division updates for Framework users
// Migrated from (r *UserFrameworkResource) updateUserDivision method
// Following SDK pattern: mirrors SDK updateUserDivision logic for Framework interface
func updateUserDivision(ctx context.Context, plan *UserFrameworkResourceModel, state *UserFrameworkResourceModel, sdkConfig *platformclientv2.Configuration) frameworkdiag.Diagnostics {
	var diagnostics frameworkdiag.Diagnostics

	// Task 8.4: Validate user ID is present
	if plan.Id.IsNull() || plan.Id.ValueString() == "" {
		// Use SDK-aligned error handling for validation errors
		frameworkDiags := handleFrameworkValidationError(ResourceType, "user_id", plan.Id, "user ID is missing or empty")
		diagnostics.Append(frameworkDiags...)
		return diagnostics
	}

	// Check if division_id has changed between plan and state (Task 6.1)
	if state != nil && plan.DivisionId.Equal(state.DivisionId) {
		// No change in division_id, skip division update logic without errors (Task 6.5)
		log.Printf("Division ID unchanged for user %s, skipping division update", plan.Id.ValueString())
		return diagnostics
	}

	divisionID := plan.DivisionId.ValueString()

	// If division_id is empty, default to home division (Task 6.2)
	if divisionID == "" {
		homeDivision, diagErr := util.GetHomeDivisionID()
		if diagErr != nil {
			// Use SDK-aligned diagnostic conversion for home division errors
			frameworkDiags := convertSDKDiagnosticsToFramework(diagErr)
			diagnostics.Append(frameworkDiags...)
			return diagnostics
		}
		divisionID = homeDivision
		log.Printf("Using home division %s for user %s", divisionID, plan.Id.ValueString())
	}

	// Task 8.4: Validate division ID format
	if divisionID == "" {
		// Use SDK-aligned error handling for validation errors
		frameworkDiags := handleFrameworkValidationError(ResourceType, "division_id", divisionID, "division ID is empty after resolution")
		diagnostics.Append(frameworkDiags...)
		return diagnostics
	}

	// Convert Framework types to format expected by division update utilities (Task 6.2)
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Updating division for USER %s to %s", plan.Id.ValueString(), divisionID)

	// Call the division update API (Task 6.1)
	_, divErr := authAPI.PostAuthorizationDivisionObject(divisionID, "USER", []string{plan.Id.ValueString()})
	if divErr != nil {
		// Use SDK-aligned error handling for division update errors
		frameworkDiags := handleFrameworkAPIError(ResourceType, "update user division", plan.Id.ValueString(), divErr, nil)
		diagnostics.Append(frameworkDiags...)
		return diagnostics
	}

	log.Printf("Successfully updated division for USER %s to %s", plan.Id.ValueString(), divisionID)
	return diagnostics
}

// updateUserSkills updates user routing skills
// Following SDK pattern: mirrors SDK updateUserSkills logic for Framework interface
func updateUserSkills(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) frameworkdiag.Diagnostics {
	var diagnostics frameworkdiag.Diagnostics

	if plan.RoutingSkills.IsNull() || plan.RoutingSkills.IsUnknown() {
		// Skills removed from configuration - remove all existing skills
		log.Printf("DEBUG: Skills removed from configuration - removing all existing skills")
		oldSdkSkills, err := getUserRoutingSkills(plan.Id.ValueString(), proxy)
		if err != nil {
			// Use SDK-aligned diagnostic conversion
			frameworkDiags := convertSDKDiagnosticsToFramework(err)
			diagnostics.Append(frameworkDiags...)
			return diagnostics
		}

		// Remove all existing skills
		if len(oldSdkSkills) > 0 {
			for _, skill := range oldSdkSkills {
				if skill.Id != nil {
					// Use SDK-aligned retry logic for skill removal
					diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
						resp, err := proxy.userApi.DeleteUserRoutingskill(plan.Id.ValueString(), *skill.Id)
						if err != nil {
							return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove skill from user %s error: %s", plan.Id.ValueString(), err), resp)
						}
						return nil, nil
					})
					if diagErr != nil {
						// Use SDK-aligned diagnostic conversion
						frameworkDiags := convertSDKDiagnosticsToFramework(diagErr)
						diagnostics.Append(frameworkDiags...)
						return diagnostics
					}
				}
			}
		}
		return diagnostics
	}

	// Convert Framework skills to SDK format
	newSkills, skillDiags := buildSdkSkillsFromFramework(plan.RoutingSkills)
	if len(skillDiags) > 0 {
		// Use SDK-aligned diagnostic conversion
		frameworkDiags := convertSDKDiagnosticsToFramework(skillDiags)
		diagnostics.Append(frameworkDiags...)
		return diagnostics
	}

	// Get current skills
	oldSdkSkills, err := getUserRoutingSkills(plan.Id.ValueString(), proxy)
	if err != nil {
		// Use SDK-aligned diagnostic conversion
		frameworkDiags := convertSDKDiagnosticsToFramework(err)
		diagnostics.Append(frameworkDiags...)
		return diagnostics
	}

	// Build maps for comparison
	newSkillProfs := make(map[string]float64)
	newSkillIds := make([]string, len(newSkills))
	for i, skill := range newSkills {
		newSkillIds[i] = *skill.Id
		newSkillProfs[*skill.Id] = *skill.Proficiency
	}

	oldSkillIds := make([]string, len(oldSdkSkills))
	oldSkillProfs := make(map[string]float64)
	for i, skill := range oldSdkSkills {
		oldSkillIds[i] = *skill.Id
		oldSkillProfs[*skill.Id] = *skill.Proficiency
	}

	// Remove skills that are no longer needed
	if len(oldSkillIds) > 0 {
		skillsToRemove := lists.SliceDifference(oldSkillIds, newSkillIds)
		for _, skillId := range skillsToRemove {
			diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
				resp, err := proxy.userApi.DeleteUserRoutingskill(plan.Id.ValueString(), skillId)
				if err != nil {
					return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove skill from user %s error: %s", plan.Id.ValueString(), err), resp)
				}
				return nil, nil
			})
			if diagErr != nil {
				// Convert SDK diagnostics to Framework diagnostics
				frameworkDiags := convertSDKDiagnosticsToFramework(diagErr)
				diagnostics.Append(frameworkDiags...)
				return diagnostics
			}
		}
	}

	// Add or update skills
	if len(newSkillIds) > 0 {
		skillsToAddOrUpdate := lists.SliceDifference(newSkillIds, oldSkillIds)
		// Check for existing proficiencies to update
		for skillID, newProf := range newSkillProfs {
			if oldProf, found := oldSkillProfs[skillID]; found {
				if newProf != oldProf {
					skillsToAddOrUpdate = append(skillsToAddOrUpdate, skillID)
				}
			}
		}

		if len(skillsToAddOrUpdate) > 0 {
			if diagErr := updateUserRoutingSkills(plan.Id.ValueString(), skillsToAddOrUpdate, newSkillProfs, proxy); diagErr != nil {
				// Convert SDK diagnostics to Framework diagnostics
				frameworkDiags := convertSDKDiagnosticsToFramework(diagErr)
				diagnostics.Append(frameworkDiags...)
				return diagnostics
			}
		}
	}

	return diagnostics
}

// updateUserLanguages updates user routing languages
// Following SDK pattern: mirrors SDK updateUserLanguages logic for Framework interface
func updateUserLanguages(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) frameworkdiag.Diagnostics {
	var diagnostics frameworkdiag.Diagnostics

	if plan.RoutingLanguages.IsNull() || plan.RoutingLanguages.IsUnknown() {
		// Languages removed from configuration - remove all existing languages
		log.Printf("DEBUG: Languages removed from configuration - removing all existing languages")
		oldSdkLanguages, err := getUserRoutingLanguages(plan.Id.ValueString(), proxy)
		if err != nil {
			frameworkDiags := convertSDKDiagnosticsToFramework(err)
			diagnostics.Append(frameworkDiags...)
			return diagnostics
		}

		// Remove all existing languages
		if len(oldSdkLanguages) > 0 {
			for _, language := range oldSdkLanguages {
				if language.Id != nil {
					diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
						resp, err := proxy.userApi.DeleteUserRoutinglanguage(plan.Id.ValueString(), *language.Id)
						if err != nil {
							return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove language from user %s error: %s", plan.Id.ValueString(), err), resp)
						}
						return nil, nil
					})
					if diagErr != nil {
						frameworkDiags := convertSDKDiagnosticsToFramework(diagErr)
						diagnostics.Append(frameworkDiags...)
						return diagnostics
					}
				}
			}
		}
		return diagnostics
	}

	// Convert Framework languages to SDK format
	newLanguages, langDiags := buildSdkLanguagesFromFramework(plan.RoutingLanguages)
	if len(langDiags) > 0 {
		// Convert SDK diagnostics to Framework diagnostics
		frameworkDiags := convertSDKDiagnosticsToFramework(langDiags)
		diagnostics.Append(frameworkDiags...)
		return diagnostics
	}

	// Get current languages
	oldSdkLanguages, err := getUserRoutingLanguages(plan.Id.ValueString(), proxy)
	if err != nil {
		// Convert SDK diagnostics to Framework diagnostics
		frameworkDiags := convertSDKDiagnosticsToFramework(err)
		diagnostics.Append(frameworkDiags...)
		return diagnostics
	}

	// Build maps for comparison
	newLangProfs := make(map[string]int)
	newLangIds := make([]string, len(newLanguages))
	for i, lang := range newLanguages {
		newLangIds[i] = *lang.Id
		newLangProfs[*lang.Id] = int(*lang.Proficiency)
	}

	oldLangIds := make([]string, len(oldSdkLanguages))
	oldLangProfs := make(map[string]int)
	for i, lang := range oldSdkLanguages {
		oldLangIds[i] = *lang.Id
		oldLangProfs[*lang.Id] = int(*lang.Proficiency)
	}

	// Remove languages that are no longer needed
	if len(oldLangIds) > 0 {
		langsToRemove := lists.SliceDifference(oldLangIds, newLangIds)
		for _, langID := range langsToRemove {
			diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
				resp, err := proxy.userApi.DeleteUserRoutinglanguage(plan.Id.ValueString(), langID)
				if err != nil {
					return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove language from user %s error: %s", plan.Id.ValueString(), err), resp)
				}
				return nil, nil
			})
			if diagErr != nil {
				// Convert SDK diagnostics to Framework diagnostics
				frameworkDiags := convertSDKDiagnosticsToFramework(diagErr)
				diagnostics.Append(frameworkDiags...)
				return diagnostics
			}
		}
	}

	// Add or update languages
	if len(newLangIds) > 0 {
		langsToAddOrUpdate := lists.SliceDifference(newLangIds, oldLangIds)
		// Check for existing proficiencies to update
		for langID, newProf := range newLangProfs {
			if oldProf, found := oldLangProfs[langID]; found {
				if newProf != oldProf {
					langsToAddOrUpdate = append(langsToAddOrUpdate, langID)
				}
			}
		}

		if len(langsToAddOrUpdate) > 0 {
			if diagErr := updateUserRoutingLanguages(plan.Id.ValueString(), langsToAddOrUpdate, newLangProfs, proxy); diagErr != nil {
				// Convert SDK diagnostics to Framework diagnostics
				frameworkDiags := convertSDKDiagnosticsToFramework(diagErr)
				diagnostics.Append(frameworkDiags...)
				return diagnostics
			}
		}
	}

	return diagnostics
}

// updateUserProfileSkills updates user profile skills
// Following SDK pattern: mirrors SDK updateUserProfileSkills logic for Framework interface
func updateUserProfileSkills(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) frameworkdiag.Diagnostics {
	var diagnostics frameworkdiag.Diagnostics

	// Check if profile skills are configured in the plan
	if plan.ProfileSkills.IsNull() || plan.ProfileSkills.IsUnknown() {
		// Profile skills not configured - skip update
		return diagnostics
	}

	// Convert Framework Set to string slice
	var profileSkillsSlice []string
	diags := plan.ProfileSkills.ElementsAs(ctx, &profileSkillsSlice, false)
	if diags.HasError() {
		// Task 8.2: Add appropriate diagnostic messages for profile skills conversion errors
		diagnostics.Append(diags...)
		diagnostics.AddError(
			"Profile Skills Conversion Error",
			fmt.Sprintf("Failed to convert profile skills for user %s", plan.Id.ValueString()),
		)
		return diagnostics
	}

	log.Printf("Updating profile skills for user %s with %d skills", plan.Id.ValueString(), len(profileSkillsSlice))

	// Use SDK-aligned retry logic with version mismatch handling
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		_, resp, err := proxy.userApi.PutUserProfileskills(plan.Id.ValueString(), profileSkillsSlice)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update profile skills for user %s error: %s", plan.Id.ValueString(), err), resp)
		}
		return nil, nil
	})

	if diagErr != nil {
		// Use SDK-aligned diagnostic conversion
		frameworkDiags := convertSDKDiagnosticsToFramework(diagErr)
		diagnostics.Append(frameworkDiags...)
	} else {
		log.Printf("Successfully updated profile skills for user %s", plan.Id.ValueString())
	}

	return diagnostics
}

// updateUserRoutingUtilization updates user routing utilization
// Following SDK pattern: mirrors SDK updateUserRoutingUtilization logic for Framework interface
func updateUserRoutingUtilization(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) frameworkdiag.Diagnostics {
	var diagnostics frameworkdiag.Diagnostics

	if plan.RoutingUtilization.IsNull() || plan.RoutingUtilization.IsUnknown() {
		// Routing utilization not configured - skip update
		return diagnostics
	}

	log.Printf("DEBUG: Updating user utilization for user %s", plan.Id.ValueString())

	// Extract routing utilization from plan
	var routingUtilizations []RoutingUtilizationModel
	diags := plan.RoutingUtilization.ElementsAs(ctx, &routingUtilizations, false)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return diagnostics
	}

	if len(routingUtilizations) == 0 {
		// Empty utilization list - reset to org defaults (Task 3.3)
		log.Printf("Resetting routing utilization to org defaults for user %s", plan.Id.ValueString())
		log.Printf("[INV] RU PLAN -> DELETE (empty list, resetting to org defaults)")
		_, err := proxy.userApi.DeleteRoutingUserUtilization(plan.Id.ValueString())
		if err != nil {
			diagnostics.AddError(
				"Failed to reset routing utilization",
				fmt.Sprintf("Failed to reset routing utilization for user %s: %s", plan.Id.ValueString(), err),
			)
		}
		log.Printf("[INV] RU write sequence: deleted=true posted=false")
		return diagnostics
	}

	// Process the utilization configuration
	utilization := routingUtilizations[0]

	// Log plan state before write
	log.Printf("[INV] RU PLAN -> payload: has_user_block=%t media_keys=%v",
		len(routingUtilizations) > 0, getSdkUtilizationTypes())
	log.Printf("[INV] RU PLAN payload JSON: %s", invMustJSON(utilization))

	// Task 3.1: Detect presence of label utilizations
	hasLabelUtilizations := !utilization.LabelUtilizations.IsNull() && len(utilization.LabelUtilizations.Elements()) > 0

	var err error
	if hasLabelUtilizations {
		// Task 3.4: Use direct API call for label utilizations
		log.Printf("Label utilizations detected - using direct API call for user %s", plan.Id.ValueString())

		// Build complex payload with both utilization and labelUtilizations sections
		apiClient := &proxy.routingApi.Configuration.APIClient
		path := fmt.Sprintf("%s/api/v2/routing/users/%s/utilization", proxy.routingApi.Configuration.BasePath, plan.Id.ValueString())
		headerParams := buildHeaderParams(proxy.routingApi)

		requestPayload := make(map[string]interface{})

		// Build media type utilizations using existing function
		allSettings := convertFrameworkToSDKUtilization(ctx, utilization, &diagnostics)
		if diagnostics.HasError() {
			return diagnostics
		}
		requestPayload["utilization"] = buildMediaTypeUtilizations(allSettings)

		// Build label utilizations using Framework-compatible function (Task 4.2)
		var labelUtilizations []LabelUtilizationModel
		labelDiags := utilization.LabelUtilizations.ElementsAs(ctx, &labelUtilizations, false)
		if labelDiags.HasError() {
			// Task 8.3: Return clear error messages for invalid label utilization configurations
			diagnostics.Append(labelDiags...)
			diagnostics.AddError(
				"Label Utilization Processing Error",
				fmt.Sprintf("Failed to process label utilizations for user %s. Ensure all label utilization configurations are valid.", plan.Id.ValueString()),
			)
			return diagnostics
		}

		labelUtilizationsRequest, labelRequestDiags := buildLabelUtilizationsRequestFromFramework(ctx, labelUtilizations)
		if labelRequestDiags.HasError() {
			// Task 8.3: Return clear error messages for invalid label utilization configurations
			diagnostics.Append(labelRequestDiags...)
			diagnostics.AddError(
				"Label Utilization Validation Error",
				fmt.Sprintf("Invalid label utilization configuration for user %s. Please check the label utilization settings and try again.", plan.Id.ValueString()),
			)
			return diagnostics
		}
		requestPayload["labelUtilizations"] = labelUtilizationsRequest

		_, err = apiClient.CallAPI(path, "PUT", requestPayload, headerParams, nil, nil, "", nil, "")
	} else {
		// Task 3.2: Use SDK client for media types only
		log.Printf("No label utilizations - using SDK client for user %s", plan.Id.ValueString())

		// Build simple Utilizationrequest payload
		sdkSettings := make(map[string]platformclientv2.Mediautilization)

		// Process each media type using existing SDK pattern
		allSettings := convertFrameworkToSDKUtilization(ctx, utilization, &diagnostics)
		if diagnostics.HasError() {
			return diagnostics
		}

		for sdkType, schemaType := range getUtilizationMediaTypes() {
			if mediaSettings, ok := allSettings[schemaType]; ok && len(mediaSettings.([]interface{})) > 0 {
				sdkSettings[sdkType] = buildSdkMediaUtilization(mediaSettings.([]interface{}))
			}
		}

		// Task 3.5: Use SDK client PutRoutingUserUtilization method
		_, _, err = proxy.userApi.PutRoutingUserUtilization(plan.Id.ValueString(), platformclientv2.Utilizationrequest{
			Utilization: &sdkSettings,
		})
	}

	if err != nil {
		log.Printf("DEBUG: Failed to update routing utilization for user %s: %s", plan.Id.ValueString(), err)
		log.Printf("[INV] RU write sequence: deleted=%t posted=false (error)", hasLabelUtilizations)
		// Use SDK-aligned error handling for routing utilization errors
		frameworkDiags := handleFrameworkAPIError(ResourceType, "update routing utilization", plan.Id.ValueString(), err, nil)
		diagnostics.Append(frameworkDiags...)
		return diagnostics
	}

	log.Printf("DEBUG: Successfully updated user utilization for user %s", plan.Id.ValueString())
	log.Printf("[INV] RU write sequence: deleted=%t posted=true", hasLabelUtilizations)

	// Add a small delay to allow the API to process the routing utilization update
	// This matches the pattern used in other parts of the codebase for async operations
	time.Sleep(2 * time.Second)
	log.Printf("DEBUG: Waited for routing utilization to take effect for user %s", plan.Id.ValueString())

	return diagnostics
}

// updateUserVoicemailPolicies updates user voicemail policies
// Following SDK pattern: mirrors SDK updateUserVoicemailPolicies logic for Framework interface
func updateUserVoicemailPolicies(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) frameworkdiag.Diagnostics {
	var diagnostics frameworkdiag.Diagnostics

	if plan.VoicemailUserpolicies.IsNull() || len(plan.VoicemailUserpolicies.Elements()) == 0 {
		// No voicemail policies configured - skip update
		return diagnostics
	}

	// Extract voicemail policies from plan
	var voicemailPolicies []VoicemailUserpoliciesModel
	diags := plan.VoicemailUserpolicies.ElementsAs(ctx, &voicemailPolicies, false)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return diagnostics
	}

	if len(voicemailPolicies) == 0 {
		return diagnostics
	}

	// Build request body (matches SDK pattern)
	voicemailPolicy := voicemailPolicies[0]
	reqBody := platformclientv2.Voicemailuserpolicy{}

	log.Printf("DEBUG: Building voicemail update request - AlertTimeoutSeconds: %v, SendEmailNotifications: %v",
		voicemailPolicy.AlertTimeoutSeconds, voicemailPolicy.SendEmailNotifications)
	log.Printf("[INV] VM PLAN -> payload: alert_timeout_seconds=%v email_notif=%v",
		voicemailPolicy.AlertTimeoutSeconds, voicemailPolicy.SendEmailNotifications)

	if !voicemailPolicy.SendEmailNotifications.IsNull() {
		reqBody.SendEmailNotifications = voicemailPolicy.SendEmailNotifications.ValueBoolPointer()
		log.Printf("DEBUG: Setting SendEmailNotifications to %t", *reqBody.SendEmailNotifications)
	}

	if !voicemailPolicy.AlertTimeoutSeconds.IsNull() {
		alertTimeout := int(voicemailPolicy.AlertTimeoutSeconds.ValueInt64())
		if alertTimeout > 0 {
			reqBody.AlertTimeoutSeconds = &alertTimeout
			log.Printf("DEBUG: Setting AlertTimeoutSeconds to %d", *reqBody.AlertTimeoutSeconds)
		} else {
			log.Printf("DEBUG: AlertTimeoutSeconds is %d (<=0), not setting in request", alertTimeout)
		}
	} else {
		log.Printf("DEBUG: AlertTimeoutSeconds is null, not setting in request")
	}

	// Update voicemail policies with retry logic (matches SDK pattern)
	log.Printf("DEBUG: Updating voicemail policies for user %s", plan.Id.ValueString())
	log.Printf("DEBUG: Request body - AlertTimeoutSeconds: %v, SendEmailNotifications: %v",
		reqBody.AlertTimeoutSeconds, reqBody.SendEmailNotifications)
	log.Printf("[INV] VM payload JSON: %s", invMustJSON(reqBody))

	// Use SDK-aligned retry logic with version mismatch handling
	retryWrapper := newFrameworkRetryWrapper(ResourceType)
	frameworkDiags := retryWrapper.executeWithRetry(func() (*platformclientv2.APIResponse, error) {
		_, resp, err := proxy.voicemailApi.PatchVoicemailUserpolicy(plan.Id.ValueString(), reqBody)
		return resp, err
	}, fmt.Sprintf("Failed to update voicemail userpolicies for user %s", plan.Id.ValueString()))

	if frameworkDiags.HasError() {
		diagnostics.Append(frameworkDiags...)
	} else {
		log.Printf("DEBUG: Successfully updated voicemail policies for user %s", plan.Id.ValueString())
		// Add a small delay to allow the API to process the voicemail policies update
		// This matches the pattern used in other parts of the codebase for async operations
		log.Printf("DEBUG: Voicemail policies updated successfully, waiting for API to process...")
		time.Sleep(2 * time.Second)
	}

	return diagnostics
}

// updatePassword updates user password
// Following SDK pattern: mirrors SDK updatePassword logic for Framework interface
func updatePassword(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) frameworkdiag.Diagnostics {
	if plan.Password.IsNull() || plan.Password.IsUnknown() || plan.Password.ValueString() == "" {
		return frameworkdiag.Diagnostics{}
	}

	password := plan.Password.ValueString()
	_, err := proxy.updatePassword(ctx, plan.Id.ValueString(), password)
	if err != nil {
		// Use SDK-aligned error handling for password update errors
		return handleFrameworkAPIError(ResourceType, "update password", plan.Id.ValueString(), err, nil)
	}

	return frameworkdiag.Diagnostics{}
}

// Helper functions for update operations

// convertFrameworkToSDKUtilization converts Framework routing utilization to SDK format for compatibility with existing functions
func convertFrameworkToSDKUtilization(ctx context.Context, utilization RoutingUtilizationModel, diagnostics *frameworkdiag.Diagnostics) map[string]interface{} {
	allSettings := make(map[string]interface{})

	// Map Framework fields to SDK schema types (matches getUtilizationMediaTypes)
	mediaTypeFields := map[string]types.List{
		"call":     utilization.Call,
		"callback": utilization.Callback,
		"chat":     utilization.Chat,
		"email":    utilization.Email,
		"message":  utilization.Message,
	}

	// Convert each media type from Framework to SDK format
	for schemaType, mediaTypeList := range mediaTypeFields {
		if !mediaTypeList.IsNull() && len(mediaTypeList.Elements()) > 0 {
			var mediaSettings []MediaUtilizationModel
			mediaDiags := mediaTypeList.ElementsAs(ctx, &mediaSettings, false)
			if mediaDiags.HasError() {
				diagnostics.Append(mediaDiags...)
				continue
			}

			if len(mediaSettings) > 0 {
				mediaSetting := mediaSettings[0]

				// Convert to SDK format (matches SDK schema.ResourceData format)
				settingsMap := make(map[string]interface{})
				settingsMap["maximum_capacity"] = int(mediaSetting.MaximumCapacity.ValueInt64())
				settingsMap["include_non_acd"] = mediaSetting.IncludeNonAcd.ValueBool()

				// Convert interruptible media types to SDK format
				if !mediaSetting.InterruptibleMediaTypes.IsNull() {
					var mediaTypes []string
					diags := mediaSetting.InterruptibleMediaTypes.ElementsAs(ctx, &mediaTypes, false)
					if !diags.HasError() {
						// Convert to *schema.Set format expected by existing functions
						// For now, we'll use a simple slice - the existing functions will handle conversion
						settingsMap["interruptible_media_types"] = mediaTypes
					}
				}

				allSettings[schemaType] = []interface{}{settingsMap}
			}
		}
	}

	// Add label utilizations if present
	if !utilization.LabelUtilizations.IsNull() && len(utilization.LabelUtilizations.Elements()) > 0 {
		var labelUtilizations []LabelUtilizationModel
		labelDiags := utilization.LabelUtilizations.ElementsAs(ctx, &labelUtilizations, false)
		if !labelDiags.HasError() {
			// Convert to SDK format
			labelUtilizationsSlice := make([]interface{}, 0, len(labelUtilizations))
			for _, labelUtil := range labelUtilizations {
				labelMap := make(map[string]interface{})
				labelMap["label_id"] = labelUtil.LabelId.ValueString()
				labelMap["maximum_capacity"] = int(labelUtil.MaximumCapacity.ValueInt64())

				if !labelUtil.InterruptingLabelIds.IsNull() {
					var interruptingIds []string
					diags := labelUtil.InterruptingLabelIds.ElementsAs(ctx, &interruptingIds, false)
					if !diags.HasError() {
						labelMap["interrupting_label_ids"] = interruptingIds
					}
				}

				labelUtilizationsSlice = append(labelUtilizationsSlice, labelMap)
			}
			allSettings["label_utilizations"] = labelUtilizationsSlice
		}
	}

	return allSettings
}

// buildLabelUtilizationsRequestFromFramework builds label utilizations request from Framework types
// This processes label_id, maximum_capacity, and interrupting_label_ids for each label (Task 4.1)
func buildLabelUtilizationsRequestFromFramework(ctx context.Context, labelUtilizations []LabelUtilizationModel) (map[string]labelUtilization, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics
	request := make(map[string]labelUtilization)

	for i, labelUtil := range labelUtilizations {
		// Task 8.3: Validate label_id is not empty
		labelId := labelUtil.LabelId.ValueString()
		if labelId == "" {
			diagnostics.AddError(
				"Invalid Label Utilization Configuration",
				fmt.Sprintf("Label utilization at index %d has empty label_id. Each label utilization must have a valid label_id.", i),
			)
			continue
		}

		// Task 8.3: Validate maximum_capacity is within valid range
		maxCapacity := int32(labelUtil.MaximumCapacity.ValueInt64())
		if maxCapacity < 0 || maxCapacity > 25 {
			diagnostics.AddError(
				"Invalid Label Utilization Configuration",
				fmt.Sprintf("Label utilization for label_id '%s' has invalid maximum_capacity %d. Maximum capacity must be between 0 and 25.", labelId, maxCapacity),
			)
			continue
		}

		// Task 8.3: Process interrupting_label_ids with validation
		var interruptingIds []string
		if !labelUtil.InterruptingLabelIds.IsNull() && !labelUtil.InterruptingLabelIds.IsUnknown() {
			diags := labelUtil.InterruptingLabelIds.ElementsAs(ctx, &interruptingIds, false)
			if diags.HasError() {
				// Task 8.3: Return clear error messages for invalid label utilization configurations
				diagnostics.Append(diags...)
				diagnostics.AddError(
					"Invalid Label Utilization Configuration",
					fmt.Sprintf("Failed to process interrupting_label_ids for label_id '%s'. Ensure all interrupting label IDs are valid strings.", labelId),
				)
				continue
			}

			// Task 8.3: Validate interrupting label IDs are not empty
			for j, interruptingId := range interruptingIds {
				if interruptingId == "" {
					diagnostics.AddError(
						"Invalid Label Utilization Configuration",
						fmt.Sprintf("Label utilization for label_id '%s' has empty interrupting_label_id at index %d. All interrupting label IDs must be non-empty strings.", labelId, j),
					)
				}
			}
		}

		// Task 8.3: Check for duplicate label IDs
		if _, exists := request[labelId]; exists {
			diagnostics.AddError(
				"Invalid Label Utilization Configuration",
				fmt.Sprintf("Duplicate label_id '%s' found in label utilizations. Each label_id must be unique.", labelId),
			)
			continue
		}

		request[labelId] = labelUtilization{
			MaximumCapacity:      maxCapacity,
			InterruptingLabelIds: interruptingIds,
		}
	}

	return request, diagnostics
}

// filterAndFlattenLabelUtilizationsFromFramework filters API response to only include originally configured labels in correct order (Task 4.4)
// This prevents drift by excluding extra labels from state identical to SDK behavior
func filterAndFlattenLabelUtilizationsFromFramework(ctx context.Context, labelUtilizations map[string]labelUtilization, originalLabelUtilizations []LabelUtilizationModel) (types.List, frameworkdiag.Diagnostics) {
	var diagnostics frameworkdiag.Diagnostics
	flattenedLabelUtilizations := make([]attr.Value, 0)

	// Process labels in the same order as originally configured (matches SDK behavior)
	for _, originalLabelUtil := range originalLabelUtilizations {
		originalLabelId := originalLabelUtil.LabelId.ValueString()

		// Find matching label in API response
		if currentLabelUtilization, exists := labelUtilizations[originalLabelId]; exists {
			// Convert to Framework types
			labelAttrs := map[string]attr.Value{
				"label_id":         types.StringValue(originalLabelId),
				"maximum_capacity": types.Int64Value(int64(currentLabelUtilization.MaximumCapacity)),
			}

			// Process interrupting_label_ids
			if len(currentLabelUtilization.InterruptingLabelIds) > 0 {
				interruptingIds := make([]attr.Value, 0, len(currentLabelUtilization.InterruptingLabelIds))
				for _, id := range currentLabelUtilization.InterruptingLabelIds {
					interruptingIds = append(interruptingIds, types.StringValue(id))
				}
				labelAttrs["interrupting_label_ids"], _ = types.SetValue(types.StringType, interruptingIds)
			} else {
				labelAttrs["interrupting_label_ids"] = types.SetNull(types.StringType)
			}

			labelObj, objDiags := types.ObjectValue(getLabelUtilizationObjectType().AttrTypes, labelAttrs)
			if objDiags.HasError() {
				// Task 8.5: Return appropriate diagnostic messages with context for utility function failures
				diagnostics.Append(objDiags...)
				diagnostics.AddError(
					"Label Utilization Object Creation Error",
					fmt.Sprintf("Failed to create label utilization object for label_id '%s'", originalLabelId),
				)
				continue
			}

			flattenedLabelUtilizations = append(flattenedLabelUtilizations, labelObj)
		} else {
			// Task 8.5: Log when configured label is not found in API response
			diagnostics.AddWarning(
				"Label Utilization Missing",
				fmt.Sprintf("Configured label_id '%s' not found in API response, excluding from state", originalLabelId),
			)
		}
	}

	// Task 8.5: Return filtered list (only configured labels in configured order) with error handling
	labelUtilizationsList, listDiags := types.ListValue(getLabelUtilizationObjectType(), flattenedLabelUtilizations)
	if listDiags.HasError() {
		diagnostics.Append(listDiags...)
		diagnostics.AddError(
			"Label Utilization List Creation Error",
			"Failed to create label utilization list from filtered objects",
		)
	}

	return labelUtilizationsList, diagnostics
}

// getLabelUtilizationObjectType returns the object type for label utilizations
func getLabelUtilizationObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"label_id":               types.StringType,
			"maximum_capacity":       types.Int64Type,
			"interrupting_label_ids": types.SetType{ElemType: types.StringType},
		},
	}
}

// readUser reads the user data and populates the model
// Migrated from (r *UserFrameworkResource) readUser method
// Following SDK pattern: mirrors SDK readUser logic for Framework interface
func readUser(ctx context.Context, model *UserFrameworkResourceModel, proxy *userProxy, diagnostics *frameworkdiag.Diagnostics) {
	currentUser, proxyResponse, errGet := proxy.getUserById(ctx, model.Id.ValueString(), []string{
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
			// Use SDK-aligned error handling for 404 errors
			frameworkDiags := handleFramework404Error(ResourceType, model.Id.ValueString())
			*diagnostics = append(*diagnostics, frameworkDiags...)
			return
		}
		// Use SDK-aligned error handling for read errors
		frameworkDiags := buildFrameworkAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read user %s error: %s", model.Id.ValueString(), errGet), proxyResponse)
		*diagnostics = append(*diagnostics, frameworkDiags...)
		return
	}

	// Log raw API response for addresses
	log.Printf("[INV] READ raw SDK user.Addresses=%s", invMustJSON(currentUser.Addresses))
	if currentUser.Addresses != nil {
		invDumpSDKPhones("READ raw SDK", *currentUser.Addresses)
	}

	// Set basic attributes
	if currentUser.Name != nil {
		model.Name = types.StringValue(*currentUser.Name)
	}
	if currentUser.Email != nil {
		model.Email = types.StringValue(*currentUser.Email)
	}
	if currentUser.Division != nil && currentUser.Division.Id != nil {
		model.DivisionId = types.StringValue(*currentUser.Division.Id)
	}
	if currentUser.State != nil {
		model.State = types.StringValue(*currentUser.State)
	}
	if currentUser.Department != nil {
		model.Department = types.StringValue(*currentUser.Department)
	}
	if currentUser.Title != nil {
		model.Title = types.StringValue(*currentUser.Title)
	}
	if currentUser.AcdAutoAnswer != nil {
		model.AcdAutoAnswer = types.BoolValue(*currentUser.AcdAutoAnswer)
	}

	// Set manager
	if currentUser.Manager != nil {
		model.Manager = types.StringValue(*(*currentUser.Manager).Id)
	} else {
		model.Manager = types.StringNull()
	}

	// Set addresses - check if addresses were configured in the plan
	if model.Addresses.IsNull() {
		// If addresses were not configured, keep them as null (Framework pattern)
		log.Printf("DEBUG: addresses not configured in plan - keeping as null regardless of API response")
		// model.Addresses remains null
	} else {
		// If addresses were configured, process API response
		log.Printf("DEBUG: addresses configured in plan - processing API response")
		addressesList, addressDiags := flattenUserAddressesForFramework(ctx, currentUser.Addresses, proxy, model.Addresses)
		*diagnostics = append(*diagnostics, addressDiags...)
		if !diagnostics.HasError() {
			model.Addresses = addressesList

			// Log post-flatten identities
			if idents := invPhoneIdentitiesFromFrameworkAddresses(addressesList); len(idents) > 0 {
				log.Printf("[INV] POST-FLAT phone identities: %v", idents)
			}
		}
	}

	// Set routing skills - always process API response (matches SDK behavior)
	skillsSet, skillsDiags := flattenUserSkillsForFramework(currentUser.Skills)
	*diagnostics = append(*diagnostics, skillsDiags...)
	if !diagnostics.HasError() {
		model.RoutingSkills = skillsSet
	}

	// Set routing languages - always process API response (matches SDK behavior)
	languagesSet, languagesDiags := flattenUserLanguagesForFramework(currentUser.Languages)
	*diagnostics = append(*diagnostics, languagesDiags...)
	if !diagnostics.HasError() {
		model.RoutingLanguages = languagesSet
	}

	// Set locations
	locationsSet, locationsDiags := flattenUserLocationsForFramework(currentUser.Locations)
	*diagnostics = append(*diagnostics, locationsDiags...)
	if !diagnostics.HasError() {
		model.Locations = locationsSet
	}

	// Set profile skills
	model.ProfileSkills = flattenUserData(ctx, currentUser.ProfileSkills)

	// Set certifications
	model.Certifications = flattenUserData(ctx, currentUser.Certifications)

	// Set employer info - always process API response
	log.Printf("DEBUG: currentUser.EmployerInfo is nil: %v", currentUser.EmployerInfo == nil)
	if currentUser.EmployerInfo != nil {
		log.Printf("DEBUG: Processing employer info from API")
		model.EmployerInfo = flattenEmployerInfo(ctx, currentUser.EmployerInfo)
	} else {
		log.Printf("DEBUG: No employer info from API - setting to null (matches SDK behavior)")
		// Return null when API returns nil (matches SDK behavior)
		model.EmployerInfo = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"official_name": types.StringType,
				"employee_id":   types.StringType,
				"employee_type": types.StringType,
				"date_hire":     types.StringType,
			},
		})
	}

	// Get voicemail user policies (matches SDK behavior)
	voicemailPolicies, _, err := proxy.getVoicemailUserpoliciesById(ctx, model.Id.ValueString())
	if err != nil {
		// If we can't read voicemail policies, set to null (matches SDK behavior)
		log.Printf("DEBUG: Failed to read voicemail policies (may not be configured): %s", err)
	}

	// Log raw API response
	log.Printf("[INV] VM READ raw: %s", invMustJSON(voicemailPolicies))

	// Log API response values
	if voicemailPolicies != nil {
		log.Printf("[INV] VM API -> alert_timeout_seconds=%v send_email_notifications=%v",
			voicemailPolicies.AlertTimeoutSeconds, voicemailPolicies.SendEmailNotifications)
	} else {
		log.Printf("[INV] VM API -> voicemailPolicies=nil")
	}

	// Set voicemail policies - process API response for import compatibility
	voicemailObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"alert_timeout_seconds":    types.Int64Type,
			"send_email_notifications": types.BoolType,
		},
	}

	if voicemailPolicies != nil {
		// Create voicemail policy object from API response (matches SDK behavior)
		voicemailAttrs := map[string]attr.Value{
			"alert_timeout_seconds":    types.Int64Null(),
			"send_email_notifications": types.BoolNull(),
		}

		if voicemailPolicies.AlertTimeoutSeconds != nil {
			voicemailAttrs["alert_timeout_seconds"] = types.Int64Value(int64(*voicemailPolicies.AlertTimeoutSeconds))
			log.Printf("[INV] VM FLATTEN alert_timeout_seconds=%d", *voicemailPolicies.AlertTimeoutSeconds)
		} else {
			log.Printf("[INV] VM FLATTEN alert_timeout_seconds=NULL")
		}
		if voicemailPolicies.SendEmailNotifications != nil {
			voicemailAttrs["send_email_notifications"] = types.BoolValue(*voicemailPolicies.SendEmailNotifications)
		}

		voicemailObj, objDiags := types.ObjectValue(voicemailObjectType.AttrTypes, voicemailAttrs)
		*diagnostics = append(*diagnostics, objDiags...)

		if !diagnostics.HasError() {
			model.VoicemailUserpolicies = types.ListValueMust(voicemailObjectType, []attr.Value{voicemailObj})
		}
	} else {
		// No voicemail policies from API - return null (matches SDK behavior)
		model.VoicemailUserpolicies = types.ListNull(voicemailObjectType)
	}

	// Read routing utilization from API (following SDK pattern)
	log.Printf("Getting user utilization")

	apiClient := &proxy.routingApi.Configuration.APIClient
	path := fmt.Sprintf("%s/api/v2/routing/users/%s/utilization", proxy.routingApi.Configuration.BasePath, model.Id.ValueString())
	headerParams := buildHeaderParams(proxy.routingApi)
	response, err := apiClient.CallAPI(path, "GET", nil, headerParams, nil, nil, "", nil, "")

	// Log raw API response
	if response != nil && response.RawBody != nil {
		log.Printf("[INV] RU READ raw: %s", string(response.RawBody))
	}

	if err != nil {
		if response != nil && response.StatusCode == 404 {
			// Task 8.1: Set routing_utilization to null when API returns 404 matching SDK behavior
			log.Printf("No routing utilization found for user %s (404), setting to null", model.Id.ValueString())
			model.RoutingUtilization = types.ListNull(getRoutingUtilizationObjectType())
		} else {
			// Task 8.1: Handle other routing utilization errors appropriately
			log.Printf("Failed to read routing utilization for user %s: %s", model.Id.ValueString(), err)
			*diagnostics = append(*diagnostics, frameworkdiag.NewErrorDiagnostic(
				"Failed to read routing utilization",
				fmt.Sprintf("Failed to read routing utilization for user %s: %s", model.Id.ValueString(), err),
			))
			model.RoutingUtilization = types.ListNull(getRoutingUtilizationObjectType())
		}
	} else {
		// Parse response using agentUtilizationWithLabels structure (Task 5.1)
		agentUtilization := &agentUtilizationWithLabels{}

		// Log raw API response status
		log.Printf("[INV] RU API -> status=%d", response.StatusCode)

		if err := json.Unmarshal(response.RawBody, &agentUtilization); err != nil {
			log.Printf("Failed to unmarshal routing utilization for user %s: %s", model.Id.ValueString(), err)
			model.RoutingUtilization = types.ListNull(getRoutingUtilizationObjectType())
		} else {
			// Log parsed API response
			hasCall := false
			hasEmail := false
			hasChat := false
			hasMessage := false
			if agentUtilization != nil && agentUtilization.Utilization != nil {
				_, hasCall = agentUtilization.Utilization["call"]
				_, hasEmail = agentUtilization.Utilization["email"]
				_, hasChat = agentUtilization.Utilization["chat"]
				_, hasMessage = agentUtilization.Utilization["message"]
			}
			labelCount := 0
			if agentUtilization != nil && agentUtilization.LabelUtilizations != nil {
				labelCount = len(agentUtilization.LabelUtilizations)
			}
			level := ""
			if agentUtilization != nil {
				level = agentUtilization.Level
			}
			log.Printf("[INV] RU API parsed: level=%q has_call=%t has_email=%t has_chat=%t has_message=%t label_utilizations=%d",
				level, hasCall, hasEmail, hasChat, hasMessage, labelCount)

			// Convert SDK logic to Framework types
			if agentUtilization == nil {
				log.Printf("No utilization data for user %s, setting to null", model.Id.ValueString())
				model.RoutingUtilization = types.ListNull(getRoutingUtilizationObjectType())
			} else if agentUtilization.Level == "Organization" {
				// Task 5.2: Organization-level settings - set to empty list (matches SDK behavior)
				log.Printf("DEBUG: Organization-level utilization for user %s, setting to empty list", model.Id.ValueString())
				model.RoutingUtilization, _ = types.ListValue(getRoutingUtilizationObjectType(), []attr.Value{})
			} else {
				// Task 5.3: User-specific settings - process all media types and label utilizations exactly as SDK
				log.Printf("User-specific utilization found for user %s, building Framework object", model.Id.ValueString())

				utilizationAttrs := map[string]attr.Value{
					"call":               types.ListNull(getMediaUtilizationObjectType()),
					"callback":           types.ListNull(getMediaUtilizationObjectType()),
					"chat":               types.ListNull(getMediaUtilizationObjectType()),
					"email":              types.ListNull(getMediaUtilizationObjectType()),
					"message":            types.ListNull(getMediaUtilizationObjectType()),
					"label_utilizations": types.ListNull(getLabelUtilizationObjectType()),
				}

				// Process utilization data if present (using agentUtilizationWithLabels structure)
				if len(agentUtilization.Utilization) > 0 {
					// Map SDK media types to Framework schema types (matches SDK getUtilizationMediaTypes)
					for sdkType, schemaType := range getUtilizationMediaTypes() {
						if mediaUtil, exists := agentUtilization.Utilization[sdkType]; exists {
							log.Printf("Processing %s utilization: %+v", schemaType, mediaUtil)

							// Build media utilization object using mediaUtilization struct (Task 5.3)
							mediaAttrs := map[string]attr.Value{
								"maximum_capacity":          types.Int64Null(),
								"include_non_acd":           types.BoolNull(),
								"interruptible_media_types": types.SetNull(types.StringType),
							}

							// Set values from API using mediaUtilization struct fields
							mediaAttrs["maximum_capacity"] = types.Int64Value(int64(mediaUtil.MaximumCapacity))
							mediaAttrs["include_non_acd"] = types.BoolValue(mediaUtil.IncludeNonAcd)

							// Only set interruptible_media_types if it has actual values (matches SDK logic)
							if len(mediaUtil.InterruptableMediaTypes) > 0 {
								var interruptibleTypes []attr.Value
								for _, mediaType := range mediaUtil.InterruptableMediaTypes {
									interruptibleTypes = append(interruptibleTypes, types.StringValue(mediaType))
								}
								mediaAttrs["interruptible_media_types"], _ = types.SetValue(types.StringType, interruptibleTypes)
							}
							// If empty or not present, leave as null (matches SDK behavior)

							mediaObj, objErr := types.ObjectValue(getMediaUtilizationObjectType().AttrTypes, mediaAttrs)
							if objErr != nil {
								log.Printf("Error creating media object for %s: %v", schemaType, objErr)
								continue
							}

							utilizationAttrs[schemaType], _ = types.ListValue(getMediaUtilizationObjectType(), []attr.Value{mediaObj})
							log.Printf("Successfully set %s utilization", schemaType)
						} else {
							log.Printf("No %s utilization data from API, leaving as null", schemaType)
						}
					}
				}

				// Task 5.4: Process label utilizations using agentUtilizationWithLabels structure
				if len(agentUtilization.LabelUtilizations) > 0 {
					log.Printf("Processing label utilizations: %+v", agentUtilization.LabelUtilizations)

					// Use labelUtilization struct directly from agentUtilizationWithLabels (Task 5.4)
					apiLabelUtilizations := agentUtilization.LabelUtilizations

					// Check if label utilizations were configured in the plan to apply filtering
					if !model.RoutingUtilization.IsNull() && len(model.RoutingUtilization.Elements()) > 0 {
						// Get the current plan's label utilizations for filtering
						var currentUtilizations []RoutingUtilizationModel
						if diags := model.RoutingUtilization.ElementsAs(ctx, &currentUtilizations, false); !diags.HasError() && len(currentUtilizations) > 0 {
							currentUtilization := currentUtilizations[0]
							if !currentUtilization.LabelUtilizations.IsNull() && len(currentUtilization.LabelUtilizations.Elements()) > 0 {
								var originalLabelUtilizations []LabelUtilizationModel
								if diags := currentUtilization.LabelUtilizations.ElementsAs(ctx, &originalLabelUtilizations, false); !diags.HasError() {
									// Apply filtering to exclude extra labels from state (matches SDK behavior)
									filteredLabelUtilizations, filterDiags := filterAndFlattenLabelUtilizationsFromFramework(ctx, apiLabelUtilizations, originalLabelUtilizations)
									if !filterDiags.HasError() {
										utilizationAttrs["label_utilizations"] = filteredLabelUtilizations
										log.Printf("Successfully filtered and set label utilizations")
									} else {
										log.Printf("Error filtering label utilizations: %v", filterDiags)
									}
								}
							}
						}
					}
				}

				utilizationObj, _ := types.ObjectValue(getRoutingUtilizationObjectType().AttrTypes, utilizationAttrs)
				result, _ := types.ListValue(getRoutingUtilizationObjectType(), []attr.Value{utilizationObj})
				model.RoutingUtilization = result

				// Log flatten result
				log.Printf("[INV] RU FLATTEN result (parent present=%t) JSON=%s",
					!result.IsNull() && !result.IsUnknown(), invMustJSON(result))

				log.Printf("Successfully set routing utilization for user %s", model.Id.ValueString())
			}
		}
	}

	// INVESTIGATION: Log final state values
	log.Printf("DEBUG: [INVESTIGATION] Final read state for user %s %s", model.Id.ValueString(), *currentUser.Email)
	log.Printf("DEBUG: [INVESTIGATION] Final routing_utilization.IsNull(): %v", model.RoutingUtilization.IsNull())
	log.Printf("DEBUG: [INVESTIGATION] This null value will cause Terraform drift detection")
}

// restoreDeletedUser restores a deleted user
// Migrated from (r *UserFrameworkResource) restoreDeletedUser method
// Following SDK pattern: mirrors SDK restoreDeletedUser logic for Framework interface
func restoreDeletedUser(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, diagnostics *frameworkdiag.Diagnostics) {
	email := plan.Email.ValueString()
	state := plan.State.ValueString()

	log.Printf("Restoring deleted user %s", email)

	currentUser, proxyResponse, err := proxy.getUserById(ctx, plan.Id.ValueString(), nil, "deleted")
	if err != nil {
		// Use SDK-aligned error handling for read errors
		frameworkDiags := buildFrameworkAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read user %s error: %s", plan.Id.ValueString(), err), proxyResponse)
		diagnostics.Append(frameworkDiags...)
		return
	}

	// Log current addresses before restore PATCH
	log.Printf("[INV] RESTORE payload for user %s", email)
	log.Printf("[INV] RESTORE payload.State=%s", state)
	log.Printf("[INV] RESTORE current.Addresses=%s", invMustJSON(currentUser.Addresses))
	if currentUser.Addresses != nil {
		invDumpSDKPhones("RESTORE current addresses (before PATCH)", *currentUser.Addresses)
	}

	restoredUser, proxyPatchResponse, patchErr := proxy.patchUserWithState(ctx, plan.Id.ValueString(), &platformclientv2.Updateuser{
		State:   &state,
		Version: currentUser.Version,
	})

	if patchErr != nil {
		// Use SDK-aligned error handling for restore errors
		frameworkDiags := handleFrameworkAPIError(ResourceType, "restore deleted user", email, patchErr, nil)
		diagnostics.Append(frameworkDiags...)
		return
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

	// Read back the restored user to get current state
	log.Printf("[INV] Reading back restored user %s", email)
	readUser(ctx, plan, proxy, diagnostics)
	if diagnostics.HasError() {
		return
	}

	log.Printf("[INV] RESTORE read-back complete for user %s", email)
	log.Printf("[INV] RESTORE final plan.Addresses=%s", invMustJSON(plan.Addresses))

	// After restoring, we need to perform additional updates
	// This will be handled by the calling Create method
}

// hasChanges checks if any of the specified attributes have changes
// Migrated from (r *UserFrameworkResource) hasChanges method
// Following SDK pattern: mirrors SDK hasChanges logic for Framework interface
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

// getRoutingUtilizationObjectType returns the object type for routing utilization
// Moved from Framework resource file to maintain clean separation of concerns
func getRoutingUtilizationObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"call":               types.ListType{ElemType: getMediaUtilizationObjectType()},
			"callback":           types.ListType{ElemType: getMediaUtilizationObjectType()},
			"chat":               types.ListType{ElemType: getMediaUtilizationObjectType()},
			"email":              types.ListType{ElemType: getMediaUtilizationObjectType()},
			"message":            types.ListType{ElemType: getMediaUtilizationObjectType()},
			"label_utilizations": types.ListType{ElemType: getLabelUtilizationObjectType()},
		},
	}
}

// getMediaUtilizationObjectType returns the object type for media utilization
// Moved from Framework resource file to maintain clean separation of concerns
func getMediaUtilizationObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"maximum_capacity":          types.Int64Type,
			"include_non_acd":           types.BoolType,
			"interruptible_media_types": types.SetType{ElemType: types.StringType},
		},
	}
}

// valueOrEmpty returns the string value of a pointer or "<nil>" if nil
func valueOrEmpty(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

// safe returns the string value of a pointer or "<nil>" if nil
func safe(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

// safeLen returns the length of a slice pointer or 0 if nil
func safeLen(contacts *[]platformclientv2.Contact) int {
	if contacts == nil {
		return 0
	}
	return len(*contacts)
}

// mustJSON marshals to JSON for debug; swallow errors in logs
func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<json err: %v>", err)
	}
	return string(b)
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

// invPresence shows presence state for types.List (NULL/UNKNOWN/SET)
func invPresence(v types.List) string {
	if v.IsNull() {
		return "NULL"
	}
	if v.IsUnknown() {
		return "UNKNOWN"
	}
	return "SET"
}

// invSDKPhoneIdentity builds a stable identity string for a phone contact as it appears from the SDK/API
func invSDKPhoneIdentity(c platformclientv2.Contact) string {
	return fmt.Sprintf("media=%s|type=%s|number=%s|ext=%s",
		invStr(c.MediaType), invStr(c.VarType), invStr(c.Address), invStr(c.Extension))
}

// canonPtrStr canonicalizes API/plan strings: treat "" the same as null
func canonPtrStr(p *string) *string {
	if p == nil {
		return nil
	}
	if *p == "" {
		return nil
	}
	return p
}

// fStringFromPtr builds a Framework types.String from a pointer, using NULL for nil/empty
func fStringFromPtr(p *string) types.String {
	if p == nil || *p == "" {
		return types.StringNull()
	}
	return types.StringValue(*p)
}

// phoneIdentity builds a one-liner identity for debug (not used in hashing, just logs)
func phoneIdentity(media, typ, number, ext types.String) string {
	get := func(s types.String) string {
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
	return fmt.Sprintf("media=%s|type=%s|number=%s|ext=%s",
		get(media), get(typ), get(number), get(ext))
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
