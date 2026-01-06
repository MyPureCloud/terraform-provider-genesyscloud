package user

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_language"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_utilization_label"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Ensure test resources are initialized for Framework tests
func init() {
	if frameworkResources == nil || frameworkDataSources == nil {
		initTestResources()
	}
}

// getFrameworkProviderFactories returns provider factories for Framework testing.
// This creates a muxed provider that includes:
//   - Framework resources: genesyscloud_user (for creating test users)
//   - Framework data sources: genesyscloud_user (for testing data source lookups)
//   - SDKv2 resources: Any dependencies needed (e.g., auth_division if needed)
//
// The muxed provider allows tests to use both Framework and SDKv2 resources together.
func getFrameworkProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"genesyscloud": func() (tfprotov6.ProviderServer, error) {
			// Define Framework resources for testing
			frameworkResources := map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			}

			// Define Framework data sources for testing
			frameworkDataSources := map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			}

			// Create muxed provider that includes both Framework and SDKv2 resources
			// This allows the test to use SDKv2 dependencies (if any) alongside Framework user resource/data source
			muxFactory := provider.NewMuxedProvider(
				"test",
				map[string]*schema.Resource{}, // SDKv2 resources (add dependencies here if needed)
				map[string]*schema.Resource{}, // SDKv2 data sources (add dependencies here if needed)
				frameworkResources,
				frameworkDataSources,
			)

			serverFactory, err := muxFactory()
			if err != nil {
				return nil, err
			}

			return serverFactory(), nil
		},
	}
}

func TestAccFrameworkResourceUserBasic(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-framework"
		email1            = "terraform-framework-" + uuid.NewString() + "@user.com"
		email2            = "terraform-framework-" + uuid.NewString() + "@user.com"
		userName1         = "John Framework"
		userName2         = "Jane Framework"
		stateActive       = "active"
		stateInactive     = "inactive"
		title1            = "Senior Developer"
		title2            = "Project Lead"
		department1       = "Engineering"
		department2       = "Product"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create basic user
				Config: generateFrameworkUserResource(
					userResourceLabel,
					email1,
					userName1,
					util.NullValue, // Defaults to active
					strconv.Quote(title1),
					strconv.Quote(department1),
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "state", stateActive),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "title", title1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "department", department1),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "manager"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "acd_auto_answer", "false"),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "id"),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "division_id"),
				),
			},
			{
				// Update user attributes
				Config: generateFrameworkUserResource(
					userResourceLabel,
					email2,
					userName2,
					strconv.Quote(stateInactive),
					strconv.Quote(title2),
					strconv.Quote(department2),
					util.NullValue, // No manager
					util.TrueValue, // AcdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "state", stateInactive),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "title", title2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "department", department2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "acd_auto_answer", "true"),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "id"),
				),
			},
			{
				// Import state verification
				ResourceName:            ResourceType + "." + userResourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"}, // Password not returned by API
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserWithProfileSkillsAndCertifications(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-profile"
		email             = "terraform-profile-" + uuid.NewString() + "@user.com"
		userName          = "Profile User"
		profileSkill1     = "Java"
		profileSkill2     = "Go"
		cert1             = "AWS Developer"
		cert2             = "AWS Architect"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create user with profile skills and certifications
				Config: generateFrameworkUserWithProfileAttrs(
					userResourceLabel,
					email,
					userName,
					generateProfileSkills(profileSkill1),
					generateCertifications(cert1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "profile_skills.0", profileSkill1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "certifications.0", cert1),
				),
			},
			{
				// Update profile skills and certifications
				Config: generateFrameworkUserWithProfileAttrs(
					userResourceLabel,
					email,
					userName,
					generateProfileSkills(profileSkill2),
					generateCertifications(cert2),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "profile_skills.0", profileSkill2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "certifications.0", cert2),
				),
			},
			{
				// Remove profile skills and certifications
				Config: generateFrameworkUserWithProfileAttrs(
					userResourceLabel,
					email,
					userName,
					"profile_skills = []", // Explicitly empty array
					"certifications = []", // Explicitly empty array
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "profile_skills.#", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "certifications.#", "0"),
				),
			},
			{
				// Import state verification
				ResourceName:            ResourceType + "." + userResourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserAddresses(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-addresses"
		email1            = "terraform-addr-" + uuid.NewString() + "@user.com"
		email2            = "terraform-other-" + uuid.NewString() + "@user.com"
		userName          = "Address User"
		phone1            = "+13173271898" // E.164 format matching SDK
		phone2            = "+13173271899" // E.164 format matching SDK
		phoneExt1         = "3532"         // Extension matching SDK
		phoneExt2         = "5678"
		phoneMediaType    = "PHONE"
		smsMediaType      = "SMS"
		addrTypeWork      = "WORK"
		addrTypeHome      = "HOME"
		addrTypeMobile    = "MOBILE"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create user with phone number and other email
				Config: generateFrameworkUserWithAddresses(
					userResourceLabel,
					email1,
					userName,
					generateFrameworkUserPhoneAddress(
						strconv.Quote(phone1),
						util.NullValue, // Default to PHONE
						util.NullValue, // Default to WORK
						util.NullValue, // No extension
					),
					generateFrameworkUserEmailAddress(
						strconv.Quote(email2),
						strconv.Quote(addrTypeHome),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.number", phone1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.type", addrTypeWork),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.address", email2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.type", addrTypeHome),
				),
			},
			{
				// Update phone number attributes and other email
				Config: generateFrameworkUserWithAddresses(
					userResourceLabel,
					email1,
					userName,
					generateFrameworkUserPhoneAddress(
						strconv.Quote(phone2),
						strconv.Quote(smsMediaType),
						strconv.Quote(addrTypeMobile),
						strconv.Quote(phoneExt1),
					),
					generateFrameworkUserEmailAddress(
						strconv.Quote(email1), // Use primary email as other email
						strconv.Quote(addrTypeWork),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.number", phone2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.media_type", smsMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.type", addrTypeMobile),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension", phoneExt1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.address", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.type", addrTypeWork),
				),
			},
			{
				// Test multiple phone numbers while maintaining other_emails
				Config: generateFrameworkUserWithAddressesAndMultiplePhones(
					userResourceLabel,
					email1,
					userName,
					generateFrameworkUserPhoneAddress(
						strconv.Quote(phone1),
						strconv.Quote(phoneMediaType),
						strconv.Quote(addrTypeWork),
						util.NullValue,
					),
					generateFrameworkUserPhoneAddress(
						strconv.Quote(phone2),
						strconv.Quote(smsMediaType),
						strconv.Quote(addrTypeMobile),
						strconv.Quote(phoneExt2),
					),
					generateFrameworkUserEmailAddress(
						strconv.Quote(email1),
						strconv.Quote(addrTypeWork),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					// Check for first phone number
					resource.TestCheckTypeSetElemNestedAttrs(
						ResourceType+"."+userResourceLabel,
						"addresses.0.phone_numbers.*",
						map[string]string{
							"number":     phone1,
							"media_type": phoneMediaType,
							"type":       addrTypeWork,
						},
					),
					// Check for second phone number
					resource.TestCheckTypeSetElemNestedAttrs(
						ResourceType+"."+userResourceLabel,
						"addresses.0.phone_numbers.*",
						map[string]string{
							"number":     phone2,
							"media_type": smsMediaType,
							"type":       addrTypeMobile,
							"extension":  phoneExt2,
						},
					),
					// Verify other_emails is still present
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.address", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.type", addrTypeWork),
				),
			},
			{
				// Test extension-only phone number (SDK edge case) - maintain other_emails
				Config: generateFrameworkUserWithAddresses(
					userResourceLabel,
					email1,
					userName,
					generateFrameworkUserPhoneAddress(
						util.NullValue,        // No number
						util.NullValue,        // Default to PHONE
						util.NullValue,        // Default to WORK
						strconv.Quote(phone1), // Extension using phone1 value
					),
					generateFrameworkUserEmailAddress(
						strconv.Quote(email1),
						strconv.Quote(addrTypeWork),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.number"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension", phone1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.type", addrTypeWork),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.address", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.type", addrTypeWork),
				),
			},
			{
				// Test extension-only to different extension-only (SDK edge case) - maintain other_emails
				Config: generateFrameworkUserWithAddresses(
					userResourceLabel,
					email1,
					userName,
					generateFrameworkUserPhoneAddress(
						util.NullValue,        // No number
						util.NullValue,        // Default to PHONE
						util.NullValue,        // Default to WORK
						strconv.Quote(phone2), // Different extension using phone2 value
					),
					generateFrameworkUserEmailAddress(
						strconv.Quote(email1),
						strconv.Quote(addrTypeWork),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.number"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension", phone2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.type", addrTypeWork),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.address", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.type", addrTypeWork),
				),
			},
			{
				// Test extension-only to number without extension (SDK edge case) - maintain other_emails
				Config: generateFrameworkUserWithAddresses(
					userResourceLabel,
					email1,
					userName,
					generateFrameworkUserPhoneAddress(
						strconv.Quote(phone2), // Number
						util.NullValue,        // Default to PHONE
						util.NullValue,        // Default to WORK
						util.NullValue,        // No extension
					),
					generateFrameworkUserEmailAddress(
						strconv.Quote(email1),
						strconv.Quote(addrTypeWork),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.number", phone2),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.type", addrTypeWork),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.address", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.type", addrTypeWork),
				),
			},
			{
				// Test number without extension to number with extension (SDK edge case) - maintain other_emails
				Config: generateFrameworkUserWithAddresses(
					userResourceLabel,
					email1,
					userName,
					generateFrameworkUserPhoneAddress(
						strconv.Quote(phone2),    // Same number
						util.NullValue,           // Default to PHONE
						util.NullValue,           // Default to WORK
						strconv.Quote(phoneExt1), // Add extension
					),
					generateFrameworkUserEmailAddress(
						strconv.Quote(email1),
						strconv.Quote(addrTypeWork),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.number", phone2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension", phoneExt1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.type", addrTypeWork),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.address", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.type", addrTypeWork),
				),
			},
			{
				// Test E.164 format validation with different phone number - maintain other_emails
				Config: generateFrameworkUserWithAddresses(
					userResourceLabel,
					email1,
					userName,
					generateFrameworkUserPhoneAddress(
						strconv.Quote(phone1), // E.164 formatted number
						util.NullValue,        // Default to PHONE
						util.NullValue,        // Default to WORK
						util.NullValue,        // No extension
					),
					generateFrameworkUserEmailAddress(
						strconv.Quote(email1),
						strconv.Quote(addrTypeWork),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.number", phone1),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.type", addrTypeWork),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.address", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.other_emails.0.type", addrTypeWork),
				),
			},

			// TODO
			// (ADDRESSES-DELETION-ASYMMETRY): This test step expects addresses to be fully removed
			// when omitted from config (addresses.# = 0). However, the Genesys Cloud API exhibits
			// asymmetric deletion behavior: when an empty Addresses array is sent via PATCH,
			// phone_numbers (PHONE/SMS media types) ARE deleted, but other_emails (EMAIL media type)
			// are NOT deleted. This may be intentional API behavior rather than a bug.
			//
			// Why this worked in SDK v2 but fails in Plugin Framework:
			// - SDK v2: After update, the Read function populated other_emails back into state from the API.
			//   The state management silently accepted this mismatch between config (no addresses) and
			//   state (has other_emails). No drift was detected, but state was inconsistent with config.
			// - Plugin Framework: PF has stricter state consistency checks. After update, when Read tries
			//   to populate other_emails from API but config says addresses should be null, PF detects
			//   the inconsistency and throws error: "Provider produced inconsistent result after apply -
			//   block count changed from 0 to 1". This is actually BETTER behavior - it catches the problem
			//   instead of silently accepting inconsistent state.
			//
			// Root cause: API's PATCH /api/v2/users/{userId} endpoint treats different media types
			// differently when processing an empty Addresses array. This asymmetry prevents clean deletion.
			//
			// Resolution options:
			// 1. Confirm with Genesys Cloud if this is intentional API behavior or a bug
			// 2. If API behavior won't change: Implement explicit deletion logic in updateUser() to send
			//    separate delete requests for EMAIL media types when addresses block is removed
			// 3. Alternative approach: Use UseStateForUnknown() plan modifier on addresses block
			//    - When addresses omitted from config → maintain prior state (don't attempt deletion)
			//    - Aligns with schema description: "If not set, this resource will not manage addresses"
			//    - Avoids the API asymmetry issue entirely
			//    - See commented-out Check block below for this behavior
			//    - Note: Adding UseStateForUnknown() alone doesn't fully solve this because we also
			//      need to handle the case where user explicitly wants to remove addresses (empty block)
			//
			// Current status: Test may fail until resolution approach is decided and implemented.
			// The commented Check block below shows the "maintain state" behavior as an alternative.

			// Test omitting addresses from config - currently expects removal but may need to maintain state
			// This tests the behavior where addresses block is omitted from config.

			// ------------------------------------------------------------
			// Update the user by removing all addresses (DEVTOOLING-1238)
			// ------------------------------------------------------------
			/*{
				Config: generateFrameworkUserResource(
					userResourceLabel,
					email1,
					userName,
					util.NullValue, // Active
					util.NullValue, // No title
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					// Verify addresses are removed when omitted from config
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.#", "0"),
				),
			},*/
			{
				// Import state verification
				ResourceName:            ResourceType + "." + userResourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserAddressWithExtensionPool(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel         = "test-user-extension-pool"
		email                     = "terraform-ext-pool-" + uuid.NewString() + "@user.com"
		userName                  = "Extension Pool User"
		extensionPoolLabel1       = "test-extension-pool-1"
		extensionPoolLabel2       = "test-extension-pool-2"
		extensionPoolStartNumber1 = "21000"
		extensionPoolEndNumber1   = "21001"
		extensionPoolStartNumber2 = "21002"
		extensionPoolEndNumber2   = "21003"
		extension1                = "21000"
		extension2                = "21002"
		phoneMediaType            = "PHONE"
		addrTypeWork              = "WORK"
	)

	// Clean up extension pools before test (if they exist)
	// Note: Cleanup is handled in CheckDestroy function

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			map[string]*schema.Resource{
				"genesyscloud_telephony_providers_edges_extension_pool": telephony_providers_edges_extension_pool.ResourceTelephonyExtensionPool(),
			},
			nil, // SDKv2 data sources
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create user with extension pool
				Config: generateFrameworkUserWithExtensionPool(
					userResourceLabel,
					email,
					userName,
					extensionPoolLabel1,
					extensionPoolStartNumber1,
					extensionPoolEndNumber1,
					extension1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension", extension1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.type", addrTypeWork),
					// TODO: Commented out - SDKv2 used hashing for Set identity, not possible in PF. Will revisit.
					// resource.TestCheckResourceAttrPair(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension_pool_id", "genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolLabel1, "id"),

				),
			},
			{
				// Update to different extension pool
				Config: generateFrameworkUserWithExtensionPool(
					userResourceLabel,
					email,
					userName,
					extensionPoolLabel2,
					extensionPoolStartNumber2,
					extensionPoolEndNumber2,
					extension2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension", extension2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.type", addrTypeWork),
					// TODO: Commented out - SDKv2 used hashing for Set identity, not possible in PF. Will revisit.
					// resource.TestCheckResourceAttrPair(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension_pool_id", "genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolLabel2, "id"),

				),
			},
			{
				// ------------------------------------------------------------
				// Remove addresses with extension pool (DEVTOOLING-1238)
				// ------------------------------------------------------------
				Config: generateFrameworkUserResource(
					userResourceLabel,
					email,
					userName,
					util.NullValue, // Default state
					util.NullValue, // No title
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.#", "0"),
				),
			},
			{
				// Import state verification
				ResourceName:            ResourceType + "." + userResourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserSkillsAndLanguages(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel   = "test-user-skills"
		email               = "terraform-skills-" + uuid.NewString() + "@user.com"
		userName            = "Skills User"
		skillResourceLabel1 = "test-skill-1"
		skillResourceLabel2 = "test-skill-2"
		langResourceLabel1  = "test-lang-1"
		langResourceLabel2  = "test-lang-2"
		skillName1          = "skill1-" + uuid.NewString()
		skillName2          = "skill2-" + uuid.NewString()
		langName1           = "lang1-" + uuid.NewString()
		langName2           = "lang2-" + uuid.NewString()
		proficiency1        = "1.5"
		proficiency2        = "2.5"
		proficiency3        = "3"
		proficiency4        = "4"
	)

	// Import routing skill and language packages for resource generation
	skillResource1 := fmt.Sprintf(`resource "genesyscloud_routing_skill" "%s" {
		name = "%s"
	}
`, skillResourceLabel1, skillName1)

	skillResource2 := fmt.Sprintf(`resource "genesyscloud_routing_skill" "%s" {
		name = "%s"
	}
`, skillResourceLabel2, skillName2)

	langResource1 := fmt.Sprintf(`resource "genesyscloud_routing_language" "%s" {
		name = "%s"
	}
`, langResourceLabel1, langName1)

	langResource2 := fmt.Sprintf(`resource "genesyscloud_routing_language" "%s" {
		name = "%s"
	}
`, langResourceLabel2, langName2)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			map[string]*schema.Resource{
				"genesyscloud_routing_skill": routing_skill.ResourceRoutingSkill(),
			},
			map[string]*schema.Resource{
				"genesyscloud_routing_skill": routing_skill.DataSourceRoutingSkill(),
			},
			map[string]func() frameworkresource.Resource{
				ResourceType:                    NewUserFrameworkResource,
				"genesyscloud_routing_language": routing_language.NewFrameworkRoutingLanguageResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType:                    NewUserFrameworkDataSource,
				"genesyscloud_routing_language": routing_language.NewFrameworkRoutingLanguageDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create user with routing skills and languages
				Config: skillResource1 + skillResource2 + langResource1 + langResource2 +
					generateFrameworkUserWithSkillsAndLanguages(
						userResourceLabel,
						email,
						userName,
						generateFrameworkUserRoutingSkill(
							fmt.Sprintf("genesyscloud_routing_skill.%s.id", skillResourceLabel1),
							proficiency1,
						),
						generateFrameworkUserRoutingLanguage(
							fmt.Sprintf("genesyscloud_routing_language.%s.id", langResourceLabel1),
							proficiency3,
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_skills.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_languages.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						ResourceType+"."+userResourceLabel,
						"routing_skills.*",
						map[string]string{
							"proficiency": proficiency1,
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						ResourceType+"."+userResourceLabel,
						"routing_languages.*",
						map[string]string{
							"proficiency": proficiency3,
						},
					),
				),
			},
			{
				// Update skills and languages
				Config: skillResource1 + skillResource2 + langResource1 + langResource2 +
					generateFrameworkUserWithMultipleSkillsAndLanguages(
						userResourceLabel,
						email,
						userName,
						generateFrameworkUserRoutingSkill(
							fmt.Sprintf("genesyscloud_routing_skill.%s.id", skillResourceLabel1),
							proficiency2,
						),
						generateFrameworkUserRoutingSkill(
							fmt.Sprintf("genesyscloud_routing_skill.%s.id", skillResourceLabel2),
							proficiency1,
						),
						generateFrameworkUserRoutingLanguage(
							fmt.Sprintf("genesyscloud_routing_language.%s.id", langResourceLabel1),
							proficiency4,
						),
						generateFrameworkUserRoutingLanguage(
							fmt.Sprintf("genesyscloud_routing_language.%s.id", langResourceLabel2),
							proficiency3,
						),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_skills.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_languages.#", "2"),
				),
			},
			{
				// Remove all skills and languages
				Config: generateFrameworkUserResource(
					userResourceLabel,
					email,
					userName,
					util.NullValue, // Active
					util.NullValue, // No title
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_skills.#", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_languages.#", "0"),
				),
			},
			{
				// Import state verification
				ResourceName:            ResourceType + "." + userResourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserEmployerInfo(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-employer"
		email             = "terraform-employer-" + uuid.NewString() + "@user.com"
		userName          = "Employer User"
		officialName1     = "John Doe Official"
		officialName2     = "Jane Smith Official"
		employeeId1       = "EMP001"
		employeeId2       = "EMP002"
		employeeType1     = "Full-time"
		employeeType2     = "Part-time"
		dateHire1         = "2023-01-15"
		dateHire2         = "2023-06-01"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create user with employer info
				Config: generateFrameworkUserWithEmployerInfo(
					userResourceLabel,
					email,
					userName,
					generateFrameworkUserEmployerInfo(
						strconv.Quote(officialName1),
						strconv.Quote(employeeId1),
						strconv.Quote(employeeType1),
						strconv.Quote(dateHire1),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "employer_info.0.official_name", officialName1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "employer_info.0.employee_id", employeeId1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "employer_info.0.employee_type", employeeType1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "employer_info.0.date_hire", dateHire1),
				),
			},
			{
				// Update employer info
				Config: generateFrameworkUserWithEmployerInfo(
					userResourceLabel,
					email,
					userName,
					generateFrameworkUserEmployerInfo(
						strconv.Quote(officialName2),
						strconv.Quote(employeeId2),
						strconv.Quote(employeeType2),
						strconv.Quote(dateHire2),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "employer_info.0.official_name", officialName2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "employer_info.0.employee_id", employeeId2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "employer_info.0.employee_type", employeeType2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "employer_info.0.date_hire", dateHire2),
				),
			},
			{
				// Remove employer info
				Config: generateFrameworkUserResource(
					userResourceLabel,
					email,
					userName,
					util.NullValue, // Active
					util.NullValue, // No title
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "employer_info.#", "0"),
				),
			},
			{
				// Import state verification
				ResourceName:            ResourceType + "." + userResourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserVoicemailPolicies(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel      = "test-user-voicemail"
		email                  = "terraform-voicemail-" + uuid.NewString() + "@user.com"
		userName               = "Voicemail User"
		timeoutSeconds1        = 550
		timeoutSeconds2        = 450
		sendEmailNotification1 = true
		sendEmailNotification2 = false
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create user with voicemail policies
				Config: generateFrameworkUserWithVoicemailPolicies(
					userResourceLabel,
					email,
					userName,
					generateFrameworkVoicemailUserpolicies(timeoutSeconds1, sendEmailNotification1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "voicemail_userpolicies.0.alert_timeout_seconds", strconv.Itoa(timeoutSeconds1)),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "voicemail_userpolicies.0.send_email_notifications", strconv.FormatBool(sendEmailNotification1)),
				),
			},
			{
				// Update voicemail policies
				Config: generateFrameworkUserWithVoicemailPolicies(
					userResourceLabel,
					email,
					userName,
					generateFrameworkVoicemailUserpolicies(timeoutSeconds2, sendEmailNotification2),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "voicemail_userpolicies.0.alert_timeout_seconds", strconv.Itoa(timeoutSeconds2)),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "voicemail_userpolicies.0.send_email_notifications", strconv.FormatBool(sendEmailNotification2)),
				),
			},
			{
				// Import state verification
				ResourceName:            ResourceType + "." + userResourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserValidation(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-validation"
		email             = "terraform-validation-" + uuid.NewString() + "@user.com"
		userName          = "Validation User"
		invalidPhone      = "invalid-phone"
		validPhone        = "+13174269078"
		invalidDate       = "invalid-date"
		validDate         = "2023-01-15"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Test invalid phone number validation
				Config: generateFrameworkUserWithAddresses(
					userResourceLabel,
					email,
					userName,
					generateFrameworkUserPhoneAddress(
						strconv.Quote(invalidPhone),
						util.NullValue, // Default to PHONE
						util.NullValue, // Default to WORK
						util.NullValue, // No extension
					),
					"", // No email address
				),
				ExpectError: regexp.MustCompile("Phone Number Validation Error"),
			},
			{
				// Test valid phone number passes validation
				Config: generateFrameworkUserWithAddresses(
					userResourceLabel,
					email,
					userName,
					generateFrameworkUserPhoneAddress(
						strconv.Quote(validPhone),
						util.NullValue, // Default to PHONE
						util.NullValue, // Default to WORK
						util.NullValue, // No extension
					),
					"", // No email address
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.number", validPhone),
				),
			},
			{
				// Test invalid date format in employer info
				Config: generateFrameworkUserWithEmployerInfo(
					userResourceLabel,
					email,
					userName,
					generateFrameworkUserEmployerInfo(
						strconv.Quote("Official Name"),
						strconv.Quote("EMP001"),
						strconv.Quote("Full-time"),
						strconv.Quote(invalidDate),
					),
				),
				ExpectError: regexp.MustCompile("Expected the date in ISO-8601 format"),
			},
			{
				// Test valid date format passes validation
				Config: generateFrameworkUserWithEmployerInfo(
					userResourceLabel,
					email,
					userName,
					generateFrameworkUserEmployerInfo(
						strconv.Quote("Official Name"),
						strconv.Quote("EMP001"),
						strconv.Quote("Full-time"),
						strconv.Quote(validDate),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "employer_info.0.date_hire", validDate),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserSkillProficiencyValidation(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel  = "test-user-skill-validation"
		email              = "terraform-skill-val-" + uuid.NewString() + "@user.com"
		userName           = "Skill Validation User"
		skillResourceLabel = "test-skill-validation"
		skillName          = "validation-skill-" + uuid.NewString()
		invalidProficiency = "10.0" // Out of range (0-5)
		validProficiency   = "3.5"
	)

	skillResource := fmt.Sprintf(`resource "genesyscloud_routing_skill" "%s" {
		name = "%s"
	}
`, skillResourceLabel, skillName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			map[string]*schema.Resource{
				"genesyscloud_routing_skill": routing_skill.ResourceRoutingSkill(),
			},
			map[string]*schema.Resource{
				"genesyscloud_routing_skill": routing_skill.DataSourceRoutingSkill(),
			},
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Test invalid proficiency validation
				Config: skillResource + generateFrameworkUserWithSkillsAndLanguages(
					userResourceLabel,
					email,
					userName,
					generateFrameworkUserRoutingSkill(
						fmt.Sprintf("genesyscloud_routing_skill.%s.id", skillResourceLabel),
						invalidProficiency,
					),
					"", // No language
				),
				ExpectError: regexp.MustCompile(`value must be between 0.000000 and 5.000000, got: 10.000000`),
			},
			{
				// Test valid proficiency passes validation
				Config: skillResource + generateFrameworkUserWithSkillsAndLanguages(
					userResourceLabel,
					email,
					userName,
					generateFrameworkUserRoutingSkill(
						fmt.Sprintf("genesyscloud_routing_skill.%s.id", skillResourceLabel),
						validProficiency,
					),
					"", // No language
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						ResourceType+"."+userResourceLabel,
						"routing_skills.*",
						map[string]string{
							"proficiency": validProficiency,
						},
					),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserDeletedUserRestoration(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-restoration"
		email             = "terraform-restore-" + uuid.NewString() + "@user.com"
		userName          = "Restoration User"
		title1            = "Original Title"
		title2            = "Updated Title"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create user
				Config: generateFrameworkUserResource(
					userResourceLabel,
					email,
					userName,
					util.NullValue, // Active
					strconv.Quote(title1),
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "title", title1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "state", "active"),
				),
			},
			{
				// Set user to inactive (simulating deletion)
				Config: generateFrameworkUserResource(
					userResourceLabel,
					email,
					userName,
					strconv.Quote("inactive"),
					strconv.Quote(title1),
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "state", "inactive"),
				),
			},
			{
				// Restore user to active and update attributes
				Config: generateFrameworkUserResource(
					userResourceLabel,
					email,
					userName,
					strconv.Quote("active"),
					strconv.Quote(title2),
					util.NullValue, // No department
					util.NullValue, // No manager
					util.TrueValue, // AcdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "state", "active"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "title", title2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "acd_auto_answer", "true"),
				),
			},
			{
				// Import state verification
				ResourceName:            ResourceType + "." + userResourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserConcurrentModification(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel1 = "test-user-concurrent-1"
		userResourceLabel2 = "test-user-concurrent-2"
		email1             = "terraform-concurrent-1-" + uuid.NewString() + "@user.com"
		email2             = "terraform-concurrent-2-" + uuid.NewString() + "@user.com"
		userName1          = "Concurrent User 1"
		userName2          = "Concurrent User 2"
		title              = "Concurrent Test"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create multiple users simultaneously to test concurrent handling
				Config: generateFrameworkUserResource(
					userResourceLabel1,
					email1,
					userName1,
					util.NullValue, // Active
					strconv.Quote(title),
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				) + generateFrameworkUserResource(
					userResourceLabel2,
					email2,
					userName2,
					util.NullValue, // Active
					strconv.Quote(title),
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel2, "email", email2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel2, "name", userName2),
				),
			},
			{
				// Set one user as manager of the other
				Config: generateFrameworkUserResource(
					userResourceLabel1,
					email1,
					userName1,
					util.NullValue, // Active
					strconv.Quote(title),
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				) + generateFrameworkUserResource(
					userResourceLabel2,
					email2,
					userName2,
					util.NullValue, // Active
					strconv.Quote(title),
					util.NullValue, // No department
					ResourceType+"."+userResourceLabel1+".id", // Manager
					util.NullValue, // Default acdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ResourceType+"."+userResourceLabel2, "manager", ResourceType+"."+userResourceLabel1, "id"),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserAPIErrorHandling(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-api-error"
		invalidEmail      = "invalid-email-format" // Invalid email format
		userName          = "API Error User"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Test API error handling with invalid email
				Config: generateFrameworkUserResource(
					userResourceLabel,
					invalidEmail,
					userName,
					util.NullValue, // Active
					util.NullValue, // No title
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				),
				ExpectError: regexp.MustCompile("Failed to create user|Invalid email format|Bad Request"),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserRoutingUtilizationValidation(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-routing-util"
		email             = "terraform-routing-" + uuid.NewString() + "@user.com"
		userName          = "Routing Util User"
		invalidCapacity   = "30" // Out of range (0-25)
		validCapacity     = "15"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Test invalid capacity validation
				Config: generateFrameworkUserWithRoutingUtilization(
					userResourceLabel,
					email,
					userName,
					generateFrameworkRoutingUtilizationCall(invalidCapacity, "false"),
				),
				ExpectError: regexp.MustCompile(`(?s)value must be.*between 0 and 25.*got.*30`),
			},
			{
				// Test valid capacity passes validation
				Config: generateFrameworkUserWithRoutingUtilization(
					userResourceLabel,
					email,
					userName,
					generateFrameworkRoutingUtilizationCall(validCapacity, "false"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.maximum_capacity", validCapacity),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserRoutingUtilizationBasic(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-routing-util-basic"
		email             = "terraform-routing-basic-" + uuid.NewString() + "@user.com"
		userName          = "Basic Routing User"
		maxCapacity0      = "0"
		maxCapacity1      = "10"
		maxCapacity2      = "12"
		utilTypeCall      = "call"
		utilTypeEmail     = "email"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create with utilization settings - matches SDK TestAccResourceUserroutingUtilBasic step 1
				Config: generateFrameworkUserWithRoutingUtilization(
					userResourceLabel,
					email,
					userName,
					generateFrameworkRoutingUtilizationAllMediaTypes(maxCapacity1, "false"),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFrameworkUserUtilizationLevel(ResourceType+"."+userResourceLabel, "Agent"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.include_non_acd", "false"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.interruptible_media_types.#"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.include_non_acd", "false"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.interruptible_media_types.#"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.include_non_acd", "false"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.interruptible_media_types.#"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.include_non_acd", "false"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.interruptible_media_types.#"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.include_non_acd", "false"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.interruptible_media_types.#"),
				),
			},
			{
				// Update utilization settings and set different org-level settings - matches SDK step 2
				Config: generateFrameworkUserWithRoutingUtilization(
					userResourceLabel,
					email,
					userName,
					generateFrameworkRoutingUtilizationWithSpecificInterruptible(maxCapacity2, "true", utilTypeEmail, utilTypeCall, utilTypeCall, utilTypeCall, utilTypeCall),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFrameworkUserUtilizationLevel(ResourceType+"."+userResourceLabel, "Agent"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.interruptible_media_types", utilTypeCall),
				),
			},
			{
				// Ensure max capacity can be set to 0 - matches SDK step 3
				Config: generateFrameworkUserWithRoutingUtilization(
					userResourceLabel,
					email,
					userName,
					generateFrameworkRoutingUtilizationWithSpecificInterruptible(maxCapacity0, "true", utilTypeEmail, utilTypeCall, utilTypeCall, utilTypeCall, utilTypeCall),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFrameworkUserUtilizationLevel(ResourceType+"."+userResourceLabel, "Agent"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.interruptible_media_types", utilTypeCall),
				),
			},
			{
				// Reset to org-level settings by specifying empty routing utilization attribute - matches SDK step 4
				Config: generateFrameworkUserResource(
					userResourceLabel,
					email,
					userName,
					util.NullValue, // Active
					util.NullValue, // No title
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					validateFrameworkUserUtilizationLevel(ResourceType+"."+userResourceLabel, "Organization"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.#", "0"),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserRoutingUtilizationWithLabels(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel       = "test-user-routing-util-labels"
		email                   = "terraform-routing-labels-" + uuid.NewString() + "@user.com"
		userName                = "Labels Routing User"
		maxCapacity0            = "0"
		maxCapacity1            = "10"
		maxCapacity2            = "12"
		utilTypeCall            = "call"
		utilTypeEmail           = "email"
		redLabelResourceLabel   = "label_red"
		blueLabelResourceLabel  = "label_blue"
		greenLabelResourceLabel = "label_green"
		redLabelName            = "Terraform Red " + uuid.NewString()
		blueLabelName           = "Terraform Blue " + uuid.NewString()
		greenLabelName          = "Terraform Green " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			map[string]*schema.Resource{
				"genesyscloud_routing_utilization_label": routing_utilization_label.ResourceRoutingUtilizationLabel(),
			},
			map[string]*schema.Resource{
				"genesyscloud_routing_utilization_label": routing_utilization_label.DataSourceRoutingUtilizationLabel(),
			},
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create with utilization settings - matches SDK TestAccResourceUserroutingUtilWithLabels step 1
				Config: generateFrameworkRoutingUtilizationLabelResource(redLabelResourceLabel, redLabelName, "") +
					generateFrameworkRoutingUtilizationLabelResource(blueLabelResourceLabel, blueLabelName, redLabelResourceLabel) +
					generateFrameworkRoutingUtilizationLabelResource(greenLabelResourceLabel, greenLabelName, blueLabelResourceLabel) +
					generateFrameworkUserWithRoutingUtilizationAndLabels(
						userResourceLabel,
						email,
						userName,
						generateFrameworkRoutingUtilizationAllMediaTypes(maxCapacity1, "false"),
						generateFrameworkLabelUtilization(redLabelResourceLabel, maxCapacity1, ""),
						generateFrameworkLabelUtilization(blueLabelResourceLabel, maxCapacity1, redLabelResourceLabel),
					),
				Check: resource.ComposeTestCheckFunc(
					validateFrameworkUserUtilizationLevel(ResourceType+"."+userResourceLabel, "Agent"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.include_non_acd", "false"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.interruptible_media_types.#"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.include_non_acd", "false"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.interruptible_media_types.#"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.include_non_acd", "false"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.interruptible_media_types.#"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.include_non_acd", "false"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.interruptible_media_types.#"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.include_non_acd", "false"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.interruptible_media_types.#"),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "routing_utilization.0.label_utilizations.0.label_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.label_utilizations.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "routing_utilization.0.label_utilizations.1.label_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.label_utilizations.1.maximum_capacity", maxCapacity1),
				),
			},
			{
				// Update utilization settings and set different org-level settings - matches SDK step 2
				Config: generateFrameworkRoutingUtilizationLabelResource(redLabelResourceLabel, redLabelName, "") +
					generateFrameworkRoutingUtilizationLabelResource(blueLabelResourceLabel, blueLabelName, redLabelResourceLabel) +
					generateFrameworkRoutingUtilizationLabelResource(greenLabelResourceLabel, greenLabelName, blueLabelResourceLabel) +
					generateFrameworkUserWithRoutingUtilizationAndLabels(
						userResourceLabel,
						email,
						userName,
						generateFrameworkRoutingUtilizationWithSpecificInterruptible(maxCapacity2, "true", utilTypeEmail, utilTypeCall, utilTypeCall, utilTypeCall, utilTypeCall),
						generateFrameworkLabelUtilization(redLabelResourceLabel, maxCapacity2, ""),
						generateFrameworkLabelUtilization(blueLabelResourceLabel, maxCapacity2, redLabelResourceLabel),
					),
				Check: resource.ComposeTestCheckFunc(
					validateFrameworkUserUtilizationLevel(ResourceType+"."+userResourceLabel, "Agent"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "routing_utilization.0.label_utilizations.0.label_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.label_utilizations.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "routing_utilization.0.label_utilizations.1.label_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.label_utilizations.1.maximum_capacity", maxCapacity2),
				),
			},
			{
				// Ensure max capacity can be set to 0 - matches SDK step 3
				Config: generateFrameworkRoutingUtilizationLabelResource(redLabelResourceLabel, redLabelName, "") +
					generateFrameworkRoutingUtilizationLabelResource(blueLabelResourceLabel, blueLabelName, redLabelResourceLabel) +
					generateFrameworkRoutingUtilizationLabelResource(greenLabelResourceLabel, greenLabelName, blueLabelResourceLabel) +
					generateFrameworkUserWithRoutingUtilizationAndLabels(
						userResourceLabel,
						email,
						userName,
						generateFrameworkRoutingUtilizationWithSpecificInterruptible(maxCapacity0, "true", utilTypeEmail, utilTypeCall, utilTypeCall, utilTypeCall, utilTypeCall),
						generateFrameworkLabelUtilization(redLabelResourceLabel, maxCapacity0, ""),
						generateFrameworkLabelUtilization(blueLabelResourceLabel, maxCapacity0, redLabelResourceLabel),
					),
				Check: resource.ComposeTestCheckFunc(
					validateFrameworkUserUtilizationLevel(ResourceType+"."+userResourceLabel, "Agent"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.include_non_acd", "true"),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel, "routing_utilization.0.message.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "routing_utilization.0.label_utilizations.0.label_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.label_utilizations.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "routing_utilization.0.label_utilizations.1.label_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.0.label_utilizations.1.maximum_capacity", maxCapacity0),
				),
			},
			{
				// Reset to org-level settings by specifying empty routing utilization attribute - matches SDK step 4
				Config: generateFrameworkUserResource(
					userResourceLabel,
					email,
					userName,
					util.NullValue, // Active
					util.NullValue, // No title
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				),
				Check: resource.ComposeTestCheckFunc(
					validateFrameworkUserUtilizationLevel(ResourceType+"."+userResourceLabel, "Organization"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "routing_utilization.#", "0"),
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(45 * time.Second)
			return testVerifyUsersDestroyed(state)
		},
	})
}

func TestAccFrameworkResourceUserProfileSkills(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-profile-skills"
		email             = "terraform-profile-skills-" + uuid.NewString() + "@user.com"
		userName          = "Profile Skills User"
		skill1            = "Java"
		skill2            = "Python"
		skill3            = "JavaScript"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Test profile skills creation
				Config: generateFrameworkUserWithProfileSkills(
					userResourceLabel,
					email,
					userName,
					generateFrameworkProfileSkills(skill1, skill2),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "profile_skills.#", "2"),
				),
			},
			{
				// Test profile skills update
				Config: generateFrameworkUserWithProfileSkills(
					userResourceLabel,
					email,
					userName,
					generateFrameworkProfileSkills(skill1, skill2, skill3),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "profile_skills.#", "3"),
				),
			},
			{
				// Test profile skills removal
				Config: generateFrameworkUserWithProfileSkills(
					userResourceLabel,
					email,
					userName,
					"", // No profile skills
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "profile_skills.#", "0"),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkResourceUserPassword(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-password"
		email             = "terraform-framework-" + uuid.NewString() + "@user.com"
		userName          = "Password Test User"
		initialPassword   = "myInitialPassword123!@#"
		updatedPassword   = "myUpdatedPassword456!@#"

		// Track password updates
		passwordUpdateCalled bool
		lastPasswordUpdate   string
	)

	err := setUserTestsActiveEnvVar()
	if err != nil {
		t.Logf("failed to set env var: %s", err.Error())
	}
	defer func() {
		if err = unsetUserTestsActiveEnvVar(); err != nil {
			t.Logf("failed to unset env var: %s", err.Error())
		}
	}()

	// Reset tracking variables
	passwordUpdateCalled = false
	lastPasswordUpdate = ""

	// Get the authorized SDK configuration
	sdkConfig, err := provider.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	// Create our mock proxy with the authorized configuration
	userProxyInstance := newUserProxy(sdkConfig)
	originalUpdatePassword := userProxyInstance.updatePasswordAttr

	userProxyInstance.updatePasswordAttr = func(ctx context.Context, p *userProxy, id string, password string) (*platformclientv2.APIResponse, error) {
		passwordUpdateCalled = true
		lastPasswordUpdate = password
		return originalUpdatePassword(ctx, p, id, password)
	}

	// Initialize internal proxy
	internalProxy = userProxyInstance
	defer func() {
		internalProxy = nil
	}()

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create user with initial password
				PreConfig: func() {
					// Reset for next test
					passwordUpdateCalled = false
					lastPasswordUpdate = ""
				},
				Config: generateFrameworkUserWithPassword(
					userResourceLabel,
					email,
					userName,
					strconv.Quote(initialPassword),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "id"),
					func(state *terraform.State) error {
						if !passwordUpdateCalled {
							return fmt.Errorf("expected password update to be called for initial password")
						}
						if lastPasswordUpdate != initialPassword {
							return fmt.Errorf("expected password to be %s, got %s", initialPassword, lastPasswordUpdate)
						}
						return nil
					},
				),
			},
			{
				PreConfig: func() {
					// Reset for next test
					passwordUpdateCalled = false
					lastPasswordUpdate = ""
				},
				// Update with new password
				Config: generateFrameworkUserWithPassword(
					userResourceLabel,
					email,
					userName,
					strconv.Quote(updatedPassword),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "id"),
					func(state *terraform.State) error {
						if !passwordUpdateCalled {
							return fmt.Errorf("expected password update to be called for password update")
						}
						if lastPasswordUpdate != updatedPassword {
							return fmt.Errorf("expected password to be %s, got %s", updatedPassword, lastPasswordUpdate)
						}
						return nil
					},
				),
			},
			{
				PreConfig: func() {
					// Reset for next test
					passwordUpdateCalled = false
					lastPasswordUpdate = ""
				},
				Config: generateFrameworkUserWithPassword(
					userResourceLabel,
					email,
					userName,
					`""`, // Empty password
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "id"),
					func(state *terraform.State) error {
						if passwordUpdateCalled {
							return fmt.Errorf("expected password update to not be called for empty password")
						}
						return nil
					},
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func validateFrameworkUserUtilizationLevel(userResourcePath string, level string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		userResource, ok := state.RootModule().Resources[userResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find user %s in state", userResourcePath)
		}
		userID := userResource.Primary.ID

		usersAPI := platformclientv2.NewUsersApi()
		util, _, err := usersAPI.GetRoutingUserUtilization(userID)
		if err != nil {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}

		if *util.Level != level {
			return fmt.Errorf("Unexpected utilization level for user %s: %s", userID, *util.Level)
		}

		return nil
	}
}

func testVerifyUsersDestroyed(state *terraform.State) error {
	usersAPI := platformclientv2.NewUsersApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		// Add retry logic for eventual consistency
		maxRetries := 10
		for i := 0; i < maxRetries; i++ {
			user, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "", "")
			if err != nil {
				if util.IsStatus404(resp) {
					// User not found as expected (hard deleted)
					break
				}
				// Unexpected error
				if i == maxRetries-1 {
					return fmt.Errorf("Unexpected error checking user %s: %s", rs.Primary.ID, err)
				}
			} else if user != nil {
				if user.State != nil && *user.State == "deleted" {
					// User soft deleted as expected
					break
				}
				// User still exists and is not deleted
				if i == maxRetries-1 {
					userState := "unknown"
					if user.State != nil {
						userState = *user.State
					}
					return fmt.Errorf("User (%s) still exists with state: %s", rs.Primary.ID, userState)
				}
			}

			// Wait before retrying
			if i < maxRetries-1 {
				time.Sleep(2 * time.Second)
			}
		}
	}
	return nil
}

func generateFrameworkUserWithRoutingUtilizationAndLabels(resourceLabel, email, name, routingUtil string, labelUtilizations ...string) string {
	labelUtilConfig := ""
	if len(labelUtilizations) > 0 {
		labelUtilConfig = strings.Join(labelUtilizations, "\n")
	}

	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		routing_utilization {
			%s
			%s
		}
	}`, ResourceType, resourceLabel, email, name, routingUtil, labelUtilConfig)
}

func generateFrameworkLabelUtilization(labelResourceLabel, maxCapacity, interruptingLabelResourceLabel string) string {
	interruptingConfig := ""
	if interruptingLabelResourceLabel != "" {
		interruptingConfig = fmt.Sprintf(`interrupting_label_ids = [genesyscloud_routing_utilization_label.%s.id]`, interruptingLabelResourceLabel)
	}

	return fmt.Sprintf(`label_utilizations {
		label_id = genesyscloud_routing_utilization_label.%s.id
		maximum_capacity = %s
		%s
	}`, labelResourceLabel, maxCapacity, interruptingConfig)
}

func generateFrameworkRoutingUtilMediaType(
	mediaType string,
	maxCapacity string,
	includeNonAcd string,
	interruptTypes ...string) string {

	interruptConfig := ""
	if len(interruptTypes) > 0 {
		interruptConfig = fmt.Sprintf(`interruptible_media_types = [%s]`, strings.Join(interruptTypes, ","))
	}

	return fmt.Sprintf(`%s {
		maximum_capacity = %s
		include_non_acd = %s
		%s
	}`, mediaType, maxCapacity, includeNonAcd, interruptConfig)
}

func generateFrameworkRoutingUtilizationLabelResource(resourceLabel, name, dependsOnResource string) string {
	dependsOn := ""
	if dependsOnResource != "" {
		dependsOn = fmt.Sprintf("depends_on = [genesyscloud_routing_utilization_label.%s]", dependsOnResource)
	}

	return fmt.Sprintf(`resource "genesyscloud_routing_utilization_label" "%s" {
		name = "%s"
		%s
	}
	`, resourceLabel, name, dependsOn)
}

func generateFrameworkUserWithRoutingUtilizationComplete(
	resourceLabel string,
	email string,
	name string,
	callConfig string,
	callbackConfig string,
	chatConfig string,
	emailConfig string,
	messageConfig string,
	labelUtilizations ...string) string {

	labelUtilConfig := ""
	if len(labelUtilizations) > 0 {
		labelUtilConfig = fmt.Sprintf("label_utilizations = [\n%s\n]", strings.Join(labelUtilizations, ",\n"))
	}

	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		routing_utilization = [
			{
				%s
				%s
				%s
				%s
				%s
				%s
			}
		]
	}`, ResourceType, resourceLabel, email, name, callConfig, callbackConfig, chatConfig, emailConfig, messageConfig, labelUtilConfig)
}

func generateFrameworkUserWithPasswordComplete(
	resourceLabel string,
	email string,
	name string,
	password string,
	state string,
	title string,
	department string) string {

	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		password = %s
		%s
		%s
		%s
	}`, ResourceType, resourceLabel, email, name, password,
		generateOptionalAttr("state", state),
		generateOptionalAttr("title", title),
		generateOptionalAttr("department", department))
}

func generateFrameworkUserWithExtensionPoolComplete(
	userResourceLabel string,
	email string,
	name string,
	extensionPoolLabel string,
	startNumber string,
	endNumber string,
	extension string,
	phoneNumber string,
	mediaType string,
	addressType string) string {

	phoneConfig := generateFrameworkUserPhoneAddressWithExtensionPoolComplete(
		phoneNumber, extension, mediaType, addressType, extensionPoolLabel)

	return fmt.Sprintf(`
resource "genesyscloud_telephony_providers_edges_extension_pool" "%s" {
	start_number = "%s"
	end_number = "%s"
	description = "Test extension pool for user integration"
}

resource "%s" "%s" {
	email = "%s"
	name = "%s"
	addresses = [
		{
			phone_numbers = [
				%s
			]
		}
	]
}
`, extensionPoolLabel, startNumber, endNumber, ResourceType, userResourceLabel, email, name, phoneConfig)
}

func generateFrameworkUserPhoneAddressWithExtensionPoolComplete(
	phoneNumber string,
	extension string,
	mediaType string,
	addressType string,
	extensionPoolLabel string) string {

	numberConfig := ""
	if phoneNumber != util.NullValue && phoneNumber != "" {
		numberConfig = fmt.Sprintf(`number = %s`, phoneNumber)
	}

	return fmt.Sprintf(`{
		%s
		extension = "%s"
		media_type = "%s"
		type = "%s"
		extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.%s.id
	}`, numberConfig, extension, mediaType, addressType, extensionPoolLabel)
}

func validateFrameworkUserRoutingUtilizationSettings(
	userResourcePath string,
	mediaType string,
	expectedCapacity string,
	expectedIncludeNonAcd string) resource.TestCheckFunc {

	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(userResourcePath, fmt.Sprintf("routing_utilization.0.%s.0.maximum_capacity", mediaType), expectedCapacity),
		resource.TestCheckResourceAttr(userResourcePath, fmt.Sprintf("routing_utilization.0.%s.0.include_non_acd", mediaType), expectedIncludeNonAcd),
	)
}

func validateFrameworkUserPhoneNumberEdgeCases(
	userResourcePath string,
	phoneIndex string,
	expectedNumber string,
	expectedExtension string,
	expectedMediaType string,
	expectedType string) resource.TestCheckFunc {

	checks := []resource.TestCheckFunc{}

	if expectedNumber != "" {
		checks = append(checks, resource.TestCheckResourceAttr(userResourcePath, fmt.Sprintf("addresses.0.phone_numbers.%s.number", phoneIndex), expectedNumber))
	} else {
		checks = append(checks, resource.TestCheckNoResourceAttr(userResourcePath, fmt.Sprintf("addresses.0.phone_numbers.%s.number", phoneIndex)))
	}

	if expectedExtension != "" {
		checks = append(checks, resource.TestCheckResourceAttr(userResourcePath, fmt.Sprintf("addresses.0.phone_numbers.%s.extension", phoneIndex), expectedExtension))
	} else {
		checks = append(checks, resource.TestCheckNoResourceAttr(userResourcePath, fmt.Sprintf("addresses.0.phone_numbers.%s.extension", phoneIndex)))
	}

	checks = append(checks, resource.TestCheckResourceAttr(userResourcePath, fmt.Sprintf("addresses.0.phone_numbers.%s.media_type", phoneIndex), expectedMediaType))
	checks = append(checks, resource.TestCheckResourceAttr(userResourcePath, fmt.Sprintf("addresses.0.phone_numbers.%s.type", phoneIndex), expectedType))

	return resource.ComposeTestCheckFunc(checks...)
}

func validateFrameworkUserExtensionPoolIntegration(
	userResourcePath string,
	extensionPoolResourcePath string,
	expectedExtension string) resource.TestCheckFunc {

	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(userResourcePath, "addresses.0.phone_numbers.0.extension", expectedExtension),
		resource.TestCheckResourceAttrPair(userResourcePath, "addresses.0.phone_numbers.0.extension_pool_id", extensionPoolResourcePath, "id"),
	)
}

func validateFrameworkUserPasswordUpdate(
	userResourcePath string,
	passwordUpdateTracker *bool,
	expectedPasswordValue *string) resource.TestCheckFunc {

	return func(state *terraform.State) error {
		if passwordUpdateTracker != nil && expectedPasswordValue != nil {
			if *expectedPasswordValue == "" && *passwordUpdateTracker {
				return fmt.Errorf("expected password update to not be called for empty password")
			}
			if *expectedPasswordValue != "" && !*passwordUpdateTracker {
				return fmt.Errorf("expected password update to be called for non-empty password")
			}
		}
		return nil
	}
}

// Helper function to generate Framework user resource with password
func generateFrameworkUserWithPassword(
	resourceLabel string,
	email string,
	name string,
	password string,
) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		password = %s
	}
	`, ResourceType, resourceLabel, email, name, password)
}

// Helper function to generate Framework user with extension pool integration
// TODO: In SDKv2, hashing was used for Set identity mapping which excluded extension_pool_id.
// This is not possible in Plugin Framework. We need to implement separate logic or a compatible
// approach. For now, commenting out extension_pool_id - will revisit this logic later.
func generateFrameworkUserWithExtensionPool(userResourceLabel, email, name, extensionPoolLabel, startNumber, endNumber, extension string) string {
	return fmt.Sprintf(`
resource "genesyscloud_telephony_providers_edges_extension_pool" "%s" {
	start_number = "%s"
	end_number = "%s"
	description = "Test extension pool for user integration"
}

resource "%s" "%s" {
	email = "%s"
	name = "%s"
	addresses {
		phone_numbers {
			extension = "%s"
			media_type = "PHONE"
			type = "WORK"
			# TODO: Temporarily commented - extension_pool_id causes Set identity mismatch in PF
			# extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.<extensionPoolLabel>.id
		}
	}
}
`, extensionPoolLabel, startNumber, endNumber, ResourceType, userResourceLabel, email, name, extension)
}

// Helper function to generate Framework user phone address with extension pool
func generateFrameworkUserPhoneAddressWithExtensionPool(extension, extensionPoolLabel string) string {
	return fmt.Sprintf(`{
		extension = "%s"
		media_type = "PHONE"
		type = "WORK"
		extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.%s.id
	}`, extension, extensionPoolLabel)
}

// Helper function to generate Framework user data source configuration
func generateUserDataSource(
	resourceLabel string,
	email string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string,
) string {
	return fmt.Sprintf(`data "%s" "%s" {
		email = %s
		name = %s
		depends_on = [%s]
	}`, ResourceType, resourceLabel, email, name, dependsOnResource)
}

func GenerateVoicemailUserpolicies(timeout int, sendEmailNotifications bool) string {
	return fmt.Sprintf(`voicemail_userpolicies {
		alert_timeout_seconds = %d
		send_email_notifications = %t
	}
	`, timeout, sendEmailNotifications)
}

func generateFrameworkUserResource(
	resourceLabel string,
	email string,
	name string,
	state string,
	title string,
	department string,
	manager string,
	acdAutoAnswer string,
) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		%s
		%s
		%s
		%s
		%s
	}
	`, ResourceType, resourceLabel, email, name,
		generateOptionalAttr("state", state),
		generateOptionalAttr("title", title),
		generateOptionalAttr("department", department),
		generateOptionalAttr("manager", manager),
		generateOptionalAttr("acd_auto_answer", acdAutoAnswer))
}

func generateFrameworkUserWithProfileAttrs(
	resourceLabel string,
	email string,
	name string,
	profileSkills string,
	certifications string,
) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		%s
		%s
	}
	`, ResourceType, resourceLabel, email, name, profileSkills, certifications)
}

func generateProfileSkills(skills ...string) string {
	if len(skills) == 0 {
		return ""
	}
	skillsStr := ""
	for _, skill := range skills {
		skillsStr += fmt.Sprintf(`"%s",`, skill)
	}
	return fmt.Sprintf("profile_skills = [%s]", skillsStr[:len(skillsStr)-1]) // Remove trailing comma
}

func generateCertifications(certs ...string) string {
	if len(certs) == 0 {
		return ""
	}
	certsStr := ""
	for _, cert := range certs {
		certsStr += fmt.Sprintf(`"%s",`, cert)
	}
	return fmt.Sprintf("certifications = [%s]", certsStr[:len(certsStr)-1]) // Remove trailing comma
}

func generateOptionalAttr(attrName string, value string) string {
	if value == util.NullValue || value == "" {
		return ""
	}
	return fmt.Sprintf("%s = %s", attrName, value)
}

func generateFrameworkUserWithAddressesAndMultiplePhones(
	resourceLabel string,
	email string,
	name string,
	phoneAddress1 string,
	phoneAddress2 string,
	emailAddress string) string {

	phoneBlocks := fmt.Sprintf(`
			phone_numbers {
				%s
			}
			phone_numbers {
				%s
			}`, phoneAddress1, phoneAddress2)

	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		addresses {%s
			%s
		}
	}`, ResourceType, resourceLabel, email, name, phoneBlocks, emailAddress)
}

func generateFrameworkUserWithAddresses(resourceLabel, email, name, phoneAddress, emailAddress string) string {
	phoneBlock := ""
	if phoneAddress != "" {
		phoneBlock = fmt.Sprintf(`
			phone_numbers {
				%s
			}`, phoneAddress)
	}

	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		addresses {%s
			%s
		}
	}`, ResourceType, resourceLabel, email, name, phoneBlock, emailAddress)
}

func generateFrameworkUserWithMultiplePhones(resourceLabel, email, name string, phoneAddresses ...string) string {
	phoneBlocks := ""
	for _, phoneAddr := range phoneAddresses {
		phoneBlocks += fmt.Sprintf(`
			phone_numbers {
				%s
			}`, phoneAddr)
	}

	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		addresses {%s
		}
	}`, ResourceType, resourceLabel, email, name, phoneBlocks)
}

func generateFrameworkUserPhoneAddress(phoneNum, phoneMediaType, phoneType, extension string) string {
	return fmt.Sprintf(`number = %s
				media_type = %s
				type = %s
				extension = %s`, phoneNum, phoneMediaType, phoneType, extension)
}

func generateFrameworkUserEmailAddress(emailAddress, emailType string) string {
	return fmt.Sprintf(`other_emails {
				address = %s
				type = %s
			}`, emailAddress, emailType)
}

func generateFrameworkUserWithSkillsAndLanguages(resourceLabel, email, name, skill, language string) string {
	skillsBlock := ""
	if skill != "" {
		skillsBlock = fmt.Sprintf("routing_skills %s", skill)
	}

	languagesBlock := ""
	if language != "" {
		languagesBlock = fmt.Sprintf("routing_languages %s", language)
	}

	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		%s
		%s
	}`, ResourceType, resourceLabel, email, name, skillsBlock, languagesBlock)
}

func generateFrameworkUserWithMultipleSkillsAndLanguages(resourceLabel, email, name string, skillsAndLanguages ...string) string {
	var skills []string
	var languages []string

	// Separate skills from languages based on content
	for _, item := range skillsAndLanguages {
		if strings.Contains(item, "skill_id") {
			skills = append(skills, item)
		} else if strings.Contains(item, "language_id") {
			languages = append(languages, item)
		}
	}

	var blocks []string
	if len(skills) > 0 {
		for _, skill := range skills {
			blocks = append(blocks, fmt.Sprintf("routing_skills %s", skill))
		}
	}
	if len(languages) > 0 {
		for _, language := range languages {
			blocks = append(blocks, fmt.Sprintf("routing_languages %s", language))
		}
	}

	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		%s
	}`, ResourceType, resourceLabel, email, name, strings.Join(blocks, "\n\t\t"))
}

func generateFrameworkUserRoutingSkill(skillID, proficiency string) string {
	return fmt.Sprintf(`{
		skill_id = %s
		proficiency = %s
	}`, skillID, proficiency)
}

func generateFrameworkUserRoutingLanguage(langID, proficiency string) string {
	return fmt.Sprintf(`{
		language_id = %s
		proficiency = %s
	}`, langID, proficiency)
}

func generateFrameworkUserWithEmployerInfo(resourceLabel, email, name, employerInfo string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		%s
	}`, ResourceType, resourceLabel, email, name, employerInfo)
}

func generateFrameworkUserEmployerInfo(officialName, employeeId, employeeType, dateHire string) string {
	return fmt.Sprintf(`employer_info {
		official_name = %s
		employee_id = %s
		employee_type = %s
		date_hire = %s
	}`, officialName, employeeId, employeeType, dateHire)
}

func generateFrameworkUserWithVoicemailPolicies(resourceLabel, email, name, voicemailPolicies string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		%s
	}`, ResourceType, resourceLabel, email, name, voicemailPolicies)
}

func generateFrameworkVoicemailUserpolicies(timeoutSeconds int, sendEmailNotifications bool) string {
	return fmt.Sprintf(`voicemail_userpolicies {
		alert_timeout_seconds = %d
		send_email_notifications = %t
	}`, timeoutSeconds, sendEmailNotifications)
}

// Additional helper functions for edge case tests
func generateFrameworkUserWithRoutingUtilization(resourceLabel, email, name, routingUtil string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		routing_utilization {
			%s
		}
	}`, ResourceType, resourceLabel, email, name, routingUtil)
}

func generateFrameworkRoutingUtilizationCall(maxCapacity, includeNonAcd string) string {
	return fmt.Sprintf(`call {
		maximum_capacity = %s
		include_non_acd = %s
	}`, maxCapacity, includeNonAcd)
}

func generateFrameworkRoutingUtilizationAllMediaTypes(maxCapacity, includeNonAcd string) string {
	return fmt.Sprintf(`
		call {
			maximum_capacity = %s
			include_non_acd = %s
		}
		callback {
			maximum_capacity = %s
			include_non_acd = %s
		}
		chat {
			maximum_capacity = %s
			include_non_acd = %s
		}
		email {
			maximum_capacity = %s
			include_non_acd = %s
		}
		message {
			maximum_capacity = %s
			include_non_acd = %s
		}`, maxCapacity, includeNonAcd, maxCapacity, includeNonAcd, maxCapacity, includeNonAcd, maxCapacity, includeNonAcd, maxCapacity, includeNonAcd)
}

func generateFrameworkRoutingUtilizationWithInterruptible(maxCapacity, includeNonAcd, callInterruptible, otherInterruptible string) string {
	return fmt.Sprintf(`
		call {
			maximum_capacity = %s
			include_non_acd = %s
			interruptible_media_types = ["%s"]
		}
		callback {
			maximum_capacity = %s
			include_non_acd = %s
			interruptible_media_types = ["%s"]
		}
		chat {
			maximum_capacity = %s
			include_non_acd = %s
			interruptible_media_types = ["%s"]
		}
		email {
			maximum_capacity = %s
			include_non_acd = %s
			interruptible_media_types = ["%s"]
		}
		message {
			maximum_capacity = %s
			include_non_acd = %s
			interruptible_media_types = ["%s"]
		}`, maxCapacity, includeNonAcd, callInterruptible, maxCapacity, includeNonAcd, otherInterruptible, maxCapacity, includeNonAcd, otherInterruptible, maxCapacity, includeNonAcd, otherInterruptible, maxCapacity, includeNonAcd, otherInterruptible)
}

func generateFrameworkRoutingUtilizationWithSpecificInterruptible(maxCapacity, includeNonAcd, callInterruptible, callbackInterruptible, chatInterruptible, emailInterruptible, messageInterruptible string) string {
	return fmt.Sprintf(`
		call {
			maximum_capacity = %s
			include_non_acd = %s
			interruptible_media_types = ["%s"]
		}
		callback {
			maximum_capacity = %s
			include_non_acd = %s
			interruptible_media_types = ["%s"]
		}
		chat {
			maximum_capacity = %s
			include_non_acd = %s
			interruptible_media_types = ["%s"]
		}
		email {
			maximum_capacity = %s
			include_non_acd = %s
			interruptible_media_types = ["%s"]
		}
		message {
			maximum_capacity = %s
			include_non_acd = %s
			interruptible_media_types = ["%s"]
		}`, maxCapacity, includeNonAcd, callInterruptible, maxCapacity, includeNonAcd, callbackInterruptible, maxCapacity, includeNonAcd, chatInterruptible, maxCapacity, includeNonAcd, emailInterruptible, maxCapacity, includeNonAcd, messageInterruptible)
}

func generateFrameworkUserWithProfileSkills(resourceLabel, email, name, profileSkills string) string {
	profileSkillsConfig := ""
	if profileSkills != "" {
		profileSkillsConfig = fmt.Sprintf("profile_skills = [%s]", profileSkills)
	} else {
		// Explicitly set to empty set to remove all skills
		profileSkillsConfig = "profile_skills = []"
	}

	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		%s
	}`, ResourceType, resourceLabel, email, name, profileSkillsConfig)
}

func generateFrameworkProfileSkills(skills ...string) string {
	var skillStrings []string
	for _, skill := range skills {
		skillStrings = append(skillStrings, fmt.Sprintf(`"%s"`, skill))
	}
	return strings.Join(skillStrings, ", ")
}
