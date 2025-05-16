package user

import (
	"context"
	"fmt"
	location "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routinglanguage "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingSkill "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill"
	routingUtilization "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_utilization"
	routingUtilizationLabel "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_utilization_label"
	"log"
	"strconv"
	"strings"

	extensionPool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func TestAccResourceUserBasic(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel1 = "test-user1"
		userResourceLabel2 = "test-user2"
		email1             = "terraform-" + uuid.NewString() + "@user.com"
		email2             = "terraform-" + uuid.NewString() + "@user.com"
		email3             = "terraform-" + uuid.NewString() + "@user.com"
		userName1          = "John Terraform"
		userName2          = "Jim Terraform"
		stateActive        = "active"
		stateInactive      = "inactive"
		title1             = "Senior Director"
		title2             = "Project Manager"
		department1        = "Development"
		department2        = "Project Management"
		profileSkill1      = "Java"
		profileSkill2      = "Go"
		cert1              = "AWS Dev"
		cert2              = "AWS Architect"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateUserResource(
					userResourceLabel1,
					email1,
					userName1,
					util.NullValue, // Defaults to active
					strconv.Quote(title1),
					strconv.Quote(department1),
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
					"",             // No profile skills
					"",             // No certs
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "state", stateActive),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "title", title1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "department", department1),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "password.%"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "manager", ""),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "acd_auto_answer", "false"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "profile_skills.%"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "certifications.%"),
					provider.TestDefaultHomeDivision(ResourceType+"."+userResourceLabel1),
				),
			},
			{
				// Update
				Config: GenerateUserResource(
					userResourceLabel1,
					email2,
					userName2,
					strconv.Quote(stateInactive),
					strconv.Quote(title2),
					strconv.Quote(department2),
					util.NullValue, // No manager
					util.TrueValue, // AcdAutoAnswer
					strconv.Quote(profileSkill1),
					strconv.Quote(cert1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "state", stateInactive),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "title", title2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "department", department2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "manager", ""),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "acd_auto_answer", "true"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "profile_skills.0", profileSkill1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "certifications.0", cert1),
					provider.TestDefaultHomeDivision(ResourceType+"."+userResourceLabel1),
				),
			},
			{
				// Create another user and set manager as existing user
				Config: GenerateUserResource(
					userResourceLabel1,
					email2,
					userName2,
					strconv.Quote(stateInactive),
					strconv.Quote(title2),
					strconv.Quote(department2),
					util.NullValue,  // No manager
					util.FalseValue, // AcdAutoAnswer
					strconv.Quote(profileSkill2),
					strconv.Quote(cert2),
				) + GenerateUserResource(
					userResourceLabel2,
					email3,
					userName1,
					util.NullValue, // Active
					strconv.Quote(title1),
					strconv.Quote(department1),
					ResourceType+"."+userResourceLabel1+".id",
					util.TrueValue, // AcdAutoAnswer
					strconv.Quote(profileSkill1),
					strconv.Quote(cert1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel2, "email", email3),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel2, "name", userName1),
					resource.TestCheckResourceAttrPair(ResourceType+"."+userResourceLabel2, "manager", ResourceType+"."+userResourceLabel1, "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "manager", ""),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "profile_skills.0", profileSkill2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "certifications.0", cert2),
				),
			},
			{
				// Remove manager and update profile skills/certs
				Config: GenerateUserResource(
					userResourceLabel2,
					email3,
					userName1,
					util.NullValue, // Active
					strconv.Quote(title1),
					strconv.Quote(department1),
					util.NullValue,
					util.FalseValue, // AcdAutoAnswer
					"",              // No profile skills
					"",              // No certs
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel2, "email", email3),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel2, "name", userName1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel2, "manager", ""),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel2, "profile_skills.%"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel2, "certifications.%"),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + userResourceLabel2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserVoicemailUserpolicies(t *testing.T) {
	var (
		userResourceLabel1     = "test-user1"
		email1                 = "terraform-" + uuid.NewString() + "@user.com"
		email2                 = "terraform-" + uuid.NewString() + "@user.com"
		userName1              = "John Terraform"
		userName2              = "Jim Terraform"
		timeoutSeconds1        = 550
		timeoutSeconds2        = 450
		sendEmailNotification1 = true
		sendEmailNotification2 = false
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName1,
					GenerateVoicemailUserpolicies(timeoutSeconds1, sendEmailNotification1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "voicemail_userpolicies.0.alert_timeout_seconds", strconv.Itoa(timeoutSeconds1)),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "voicemail_userpolicies.0.send_email_notifications", strconv.FormatBool(sendEmailNotification1)),
					provider.TestDefaultHomeDivision(ResourceType+"."+userResourceLabel1),
				),
			},
			{
				// Update
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email2,
					userName2,
					GenerateVoicemailUserpolicies(timeoutSeconds2, sendEmailNotification2),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "voicemail_userpolicies.0.alert_timeout_seconds", strconv.Itoa(timeoutSeconds2)),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "voicemail_userpolicies.0.send_email_notifications", strconv.FormatBool(sendEmailNotification2)),
					provider.TestDefaultHomeDivision(ResourceType+"."+userResourceLabel1),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + userResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func generateUserWithCustomAttrs(resourceLabel string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, ResourceType, resourceLabel, email, name, strings.Join(attrs, "\n"))
}

func TestAccResourceUserAddresses(t *testing.T) {
	t.Parallel()
	var (
		addrUserResourceLabel1      = "test-user-addr1"
		addrUserResourceLabel2      = "test-user-addr2"
		addrUserResourceLabel3      = "test-user-addr3"
		addrUserResourceLabel4      = "test-user-addr4"
		addrUserName1               = "Nancy Terraform"
		addrUserName2               = "Oliver Tofu"
		addrUserName3               = "Tony Bee"
		addrUserName4               = "Scott Crav"
		addrEmail1                  = "terraform1-" + uuid.NewString() + "@user.com"
		addrEmail2                  = "terraform2-" + uuid.NewString() + "@user.com"
		addrEmail3                  = "terraform3-" + uuid.NewString() + "@user.com"
		addrEmail4                  = "terraform4-" + uuid.NewString() + "@user.com"
		addrPhone1                  = "+13174269078"
		addrPhone2                  = "+441434634996"
		addrPhone3                  = "+13174222323"
		addrPhone4_1                = "+13174265397"
		addrPhone4_2                = "+13172347890"
		addrPhoneExt1               = "1234"
		addrPhoneExt2               = "1345"
		addrPhoneExt3_1             = "3456"
		addrPhoneExt3_2             = "7270"
		addrPhoneExt4_1             = "2345"
		addrPhoneExt4_2             = "9000"
		addrPhoneExt4_3             = "98765"
		phoneMediaType              = "PHONE"
		smsMediaType                = "SMS"
		addrTypeWork                = "WORK"
		addrTypeWork2               = "WORK2"
		addrTypeWork3               = "WORK3"
		addrTypeHome                = "HOME"
		addrTypeMobile              = "MOBILE"
		extensionPoolResourceLabel1 = "test-extensionpool1" + uuid.NewString()
		extensionPoolStartNumber1   = "8000"
		extensionPoolEndNumber1     = "8999"
	)

	extensionPoolResource := extensionPool.ExtensionPoolStruct{
		ResourceLabel: extensionPoolResourceLabel1,
		StartNumber:   extensionPoolStartNumber1,
		EndNumber:     extensionPoolEndNumber1,
		Description:   util.NullValue, // No description
	}

	err := extensionPool.DeleteExtensionPoolWithNumber(extensionPoolStartNumber1)
	if err != nil {
		log.Fatalf("%s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateUserWithCustomAttrs(
					addrUserResourceLabel1,
					addrEmail1,
					addrUserName1,
					generateUserAddresses(
						generateUserPhoneAddress(
							strconv.Quote(addrPhone1),
							util.NullValue, // Default to type PHONE
							util.NullValue, // Default to type WORK
							util.NullValue, // No extension
						),
						generateUserEmailAddress(
							strconv.Quote(addrEmail2),
							strconv.Quote(addrTypeHome),
						),
					),
				) + extensionPool.GenerateExtensionPoolResource(&extensionPoolResource),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "email", addrEmail1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "name", addrUserName1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.number", addrPhone1),
					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.extension"),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.type", addrTypeWork),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.other_emails.0.address", addrEmail2),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.other_emails.0.type", addrTypeHome),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + addrUserResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update phone number and other email attributes
				Config: generateUserWithCustomAttrs(
					addrUserResourceLabel1,
					addrEmail1,
					addrUserName1,
					generateUserAddresses(
						generateUserPhoneAddress(
							strconv.Quote(addrPhone2),
							strconv.Quote(smsMediaType),
							strconv.Quote(addrTypeHome),
							strconv.Quote(addrPhoneExt1),
						),
						generateUserEmailAddress(
							strconv.Quote(addrEmail3),
							strconv.Quote(addrTypeWork),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "email", addrEmail1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "name", addrUserName1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.number", addrPhone2),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.media_type", smsMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.type", addrTypeHome),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.extension", addrPhoneExt1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.other_emails.0.address", addrEmail3),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.other_emails.0.type", addrTypeWork),
				),
			},
			{
				// Add a user with only extension
				Config: generateUserWithCustomAttrs(
					addrUserResourceLabel2,
					addrEmail2,
					addrUserName2,
					generateUserAddresses(
						generateUserPhoneAddress(
							util.NullValue,
							strconv.Quote(phoneMediaType),
							strconv.Quote(addrTypeHome),
							strconv.Quote(addrPhoneExt2),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel2, "email", addrEmail2),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel2, "name", addrUserName2),
					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel2, "addresses.0.phone_numbers.0.number"),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel2, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel2, "addresses.0.phone_numbers.0.type", addrTypeHome),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel2, "addresses.0.phone_numbers.0.extension", addrPhoneExt2),
					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel2, "addresses.0.other_emails.0.address"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel2, "addresses.0.other_emails.0.type"),
				),
			},
			{
				// Add a user checking DEVTOOLING-739
				Config: generateUserWithCustomAttrs(
					addrUserResourceLabel3,
					addrEmail3,
					addrUserName3,
					generateUserAddresses(
						generateUserPhoneAddress(
							strconv.Quote(addrPhone3),
							strconv.Quote(phoneMediaType),
							strconv.Quote(addrTypeMobile),
							strconv.Quote(addrPhoneExt3_1),
						),
						generateUserPhoneAddress(
							util.NullValue,
							strconv.Quote(phoneMediaType),
							strconv.Quote(addrTypeWork3),
							strconv.Quote(addrPhoneExt3_2),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					// Basic resource attributes
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel3, "email", addrEmail3),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel3, "name", addrUserName3),

					// Check for the first phone address (with phone + extension)
					resource.TestCheckTypeSetElemNestedAttrs(
						ResourceType+"."+addrUserResourceLabel3,
						"addresses.0.phone_numbers.*",
						map[string]string{
							"number":     addrPhone3,
							"media_type": phoneMediaType,
							"type":       addrTypeMobile,
							"extension":  addrPhoneExt3_1,
						},
					),

					// Check for the second phone address (with only extension)
					resource.TestCheckTypeSetElemNestedAttrs(
						ResourceType+"."+addrUserResourceLabel3,
						"addresses.0.phone_numbers.*",
						map[string]string{
							"media_type": phoneMediaType,
							"type":       addrTypeWork3,
							"extension":  addrPhoneExt3_2,
						},
					),

					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel3, "addresses.0.other_emails.0.address"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel3, "addresses.0.other_emails.0.type"),
				),
			},
			{
				// Add a user multiple phone addresses
				Config: generateUserWithCustomAttrs(
					addrUserResourceLabel4,
					addrEmail4,
					addrUserName4,
					generateUserAddresses(
						generateUserPhoneAddress(
							util.NullValue,
							strconv.Quote(phoneMediaType),
							strconv.Quote(addrTypeHome),
							strconv.Quote(addrPhoneExt4_1),
						),
						generateUserPhoneAddress(
							strconv.Quote(addrPhone4_1),
							util.NullValue, // Default to type PHONE
							util.NullValue, // Default to type WORK
							util.NullValue, // No extension
						),
						generateUserPhoneAddress(
							strconv.Quote(addrPhone4_2),
							strconv.Quote(smsMediaType),
							strconv.Quote(addrTypeWork2),
							strconv.Quote(addrPhoneExt4_2),
						),
						generateUserPhoneAddress(
							util.NullValue,
							strconv.Quote(phoneMediaType),
							strconv.Quote(addrTypeWork3),
							strconv.Quote(addrPhoneExt4_3),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					// Basic resource attributes
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel4, "email", addrEmail4),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel4, "name", addrUserName4),

					// Check for the first phone address (with only extension)
					resource.TestCheckTypeSetElemNestedAttrs(
						ResourceType+"."+addrUserResourceLabel4,
						"addresses.0.phone_numbers.*",
						map[string]string{
							"media_type": phoneMediaType,
							"type":       addrTypeHome,
							"extension":  addrPhoneExt4_1,
						},
					),

					// Check for the second phone address (with only number)
					resource.TestCheckTypeSetElemNestedAttrs(
						ResourceType+"."+addrUserResourceLabel4,
						"addresses.0.phone_numbers.*",
						map[string]string{
							"number":     addrPhone4_1,
							"media_type": phoneMediaType,
							"type":       addrTypeWork,
						},
					),

					// Check for the third phone address (with both number and extension)
					resource.TestCheckTypeSetElemNestedAttrs(
						ResourceType+"."+addrUserResourceLabel4,
						"addresses.0.phone_numbers.*",
						map[string]string{
							"number":     addrPhone4_2,
							"media_type": smsMediaType,
							"type":       addrTypeWork2,
							"extension":  addrPhoneExt4_2,
						},
					),

					// Check for the fourth phone address (with only extension)
					resource.TestCheckTypeSetElemNestedAttrs(
						ResourceType+"."+addrUserResourceLabel4,
						"addresses.0.phone_numbers.*",
						map[string]string{
							"media_type": phoneMediaType,
							"type":       addrTypeWork3,
							"extension":  addrPhoneExt4_3,
						},
					),

					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel4, "addresses.0.other_emails.0.address"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel4, "addresses.0.other_emails.0.type"),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserPhone(t *testing.T) {
	var (
		addrUserResourceLabel1      = "test-user-addr"
		addrUserName                = "Nancy Terraform"
		addrEmail1                  = "terraform-" + uuid.NewString() + "@user.com"
		addrPhone1                  = "+13173271898"
		addrPhone2                  = "+13173271899"
		addrExt1                    = "3532"
		phoneMediaType              = "PHONE"
		addrTypeWork                = "WORK"
		extensionPoolResourceLabel1 = "test-extensionpool" + uuid.NewString()
		extensionPoolStartNumber1   = "3000"
		extensionPoolEndNumber1     = "4000"
	)

	extensionPoolResource := extensionPool.ExtensionPoolStruct{
		ResourceLabel: extensionPoolResourceLabel1,
		StartNumber:   extensionPoolStartNumber1,
		EndNumber:     extensionPoolEndNumber1,
		Description:   util.NullValue, // No description
	}

	extensionPool.DeleteExtensionPoolWithNumber(extensionPoolStartNumber1)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateUserWithCustomAttrs(
					addrUserResourceLabel1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							util.NullValue,            // number
							util.NullValue,            // Default to type PHONE
							util.NullValue,            // Default to type WORK
							strconv.Quote(addrPhone1), // extension
						),
					),
				) + extensionPool.GenerateExtensionPoolResource(&extensionPoolResource),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "email", addrEmail1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "name", addrUserName),
					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.number"),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.extension", addrPhone1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.type", addrTypeWork),
				),
			},
			{
				Config: generateUserWithCustomAttrs(
					addrUserResourceLabel1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							util.NullValue,            // number
							util.NullValue,            // Default to type PHONE
							util.NullValue,            // Default to type WORK
							strconv.Quote(addrPhone2), // extension
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "email", addrEmail1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "name", addrUserName),
					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.number"),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.extension", addrPhone2),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.type", addrTypeWork),
				),
			},
			{
				Config: generateUserWithCustomAttrs(
					addrUserResourceLabel1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							strconv.Quote(addrPhone2), // number
							util.NullValue,            // Default to type PHONE
							util.NullValue,            // Default to type WORK
							util.NullValue,            // extension
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "email", addrEmail1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "name", addrUserName),
					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.extension"),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.number", addrPhone2),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.type", addrTypeWork),
				),
			},
			{
				Config: generateUserWithCustomAttrs(
					addrUserResourceLabel1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							strconv.Quote(addrPhone2), // number
							util.NullValue,            // Default to type PHONE
							util.NullValue,            // Default to type WORK
							strconv.Quote(addrExt1),   // extension
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "email", addrEmail1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "name", addrUserName),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.number", addrPhone2),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.extension", addrExt1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.type", addrTypeWork),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserSkills(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel1  = "test-user"
		email1              = "terraform-" + uuid.NewString() + "@user.com"
		userName1           = "Skill Terraform"
		skillResourceLabel1 = "test-skill-1"
		skillResourceLabel2 = "test-skill-2"
		skillName1          = "skill1-" + uuid.NewString()
		skillName2          = "skill2-" + uuid.NewString()
		proficiency1        = "1.5"
		proficiency2        = "2.5"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user with 1 skill
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName1,
					generateUserRoutingSkill("genesyscloud_routing_skill."+skillResourceLabel1+".id", proficiency1),
				) + routingSkill.GenerateRoutingSkillResource(skillResourceLabel1, skillName1),
				Check: resource.ComposeTestCheckFunc(
					validateUserSkill(ResourceType+"."+userResourceLabel1, "genesyscloud_routing_skill."+skillResourceLabel1, proficiency1),
				),
			},
			{
				// Create another skill and add to the user
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName1,
					generateUserRoutingSkill("genesyscloud_routing_skill."+skillResourceLabel1+".id", proficiency1),
					generateUserRoutingSkill("genesyscloud_routing_skill."+skillResourceLabel2+".id", proficiency2),
				) + routingSkill.GenerateRoutingSkillResource(
					skillResourceLabel1,
					skillName1,
				) + routingSkill.GenerateRoutingSkillResource(
					skillResourceLabel2,
					skillName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserSkill(ResourceType+"."+userResourceLabel1, "genesyscloud_routing_skill."+skillResourceLabel1, proficiency1),
					validateUserSkill(ResourceType+"."+userResourceLabel1, "genesyscloud_routing_skill."+skillResourceLabel2, proficiency2),
				),
			},
			{
				// Remove a skill from the user and modify proficiency
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName1,
					generateUserRoutingSkill("genesyscloud_routing_skill."+skillResourceLabel2+".id", proficiency1),
				) + routingSkill.GenerateRoutingSkillResource(
					skillResourceLabel2,
					skillName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserSkill(ResourceType+"."+userResourceLabel1, "genesyscloud_routing_skill."+skillResourceLabel2, proficiency1),
				),
			},
			{
				// Remove all skills from the user
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName1,
					"routing_skills = []",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "skills.%"),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper updation
						return nil
					},
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(45 * time.Second)
			return testVerifyUsersDestroyed(state)
		},
	})
}

func TestAccResourceUserLanguages(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel1 = "test-user"
		email1             = "terraform-" + uuid.NewString() + "@user.com"
		userName1          = "Lang Terraform"
		langResourceLabel1 = "test-lang-1"
		langResourceLabel2 = "test-lang-2"
		langName1          = "lang1-" + uuid.NewString()
		langName2          = "lang2-" + uuid.NewString()
		proficiency1       = "1"
		proficiency2       = "2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user with 1 language
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName1,
					generateUserRoutingLang("genesyscloud_routing_language."+langResourceLabel1+".id", proficiency1),
				) + routinglanguage.GenerateRoutingLanguageResource(langResourceLabel1, langName1),
				Check: resource.ComposeTestCheckFunc(
					validateUserLanguage(ResourceType+"."+userResourceLabel1, "genesyscloud_routing_language."+langResourceLabel1, proficiency1),
				),
			},
			{
				// Create another language and add to the user
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName1,
					generateUserRoutingLang("genesyscloud_routing_language."+langResourceLabel1+".id", proficiency1),
					generateUserRoutingLang("genesyscloud_routing_language."+langResourceLabel2+".id", proficiency2),
				) + routinglanguage.GenerateRoutingLanguageResource(
					langResourceLabel1,
					langName1,
				) + routinglanguage.GenerateRoutingLanguageResource(
					langResourceLabel2,
					langName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserLanguage(ResourceType+"."+userResourceLabel1, "genesyscloud_routing_language."+langResourceLabel1, proficiency1),
					validateUserLanguage(ResourceType+"."+userResourceLabel1, "genesyscloud_routing_language."+langResourceLabel2, proficiency2),
				),
			},
			{
				// Remove a language from the user and modify proficiency
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName1,
					generateUserRoutingLang("genesyscloud_routing_language."+langResourceLabel2+".id", proficiency1),
				) + routinglanguage.GenerateRoutingLanguageResource(
					langResourceLabel2,
					langName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserLanguage(ResourceType+"."+userResourceLabel1, "genesyscloud_routing_language."+langResourceLabel2, proficiency1),
				),
			},
			{
				// Remove all languages from the user
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName1,
					"routing_languages = []",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_languages.%"),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper deletion
						return nil
					},
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserLocations(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel1 = "test-user-loc"
		email              = "terraform-" + uuid.NewString() + "@user.com"
		userName           = "Loki Terraform"
		locResourceLabel1  = "test-location1"
		locResourceLabel2  = "test-location2"
		locName1           = "Terraform location" + uuid.NewString()
		locName2           = "Terraform location" + uuid.NewString()
		locNotes1          = "First floor"
		locNotes2          = "Second floor"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user with a location
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email,
					userName,
					generateUserLocation(
						"genesyscloud_location."+locResourceLabel1+".id",
						strconv.Quote(locNotes1),
					),
				) + location.GenerateLocationResourceBasic(locResourceLabel1, locName1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email),
					resource.TestCheckResourceAttrPair(ResourceType+"."+userResourceLabel1, "locations.0.location_id", "genesyscloud_location."+locResourceLabel1, "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "locations.0.notes", locNotes1),
				),
			},
			{
				// Update with a new location
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email,
					userName,
					generateUserLocation(
						"genesyscloud_location."+locResourceLabel2+".id",
						strconv.Quote(locNotes2),
					),
				) + location.GenerateLocationResourceBasic(locResourceLabel2, locName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email),
					resource.TestCheckResourceAttrPair(ResourceType+"."+userResourceLabel1, "locations.0.location_id", "genesyscloud_location."+locResourceLabel2, "id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "locations.0.notes", locNotes2),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserEmployerInfo(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel1 = "test-user-info"
		userName           = "Info Terraform"
		email1             = "terraform-" + uuid.NewString() + "@user.com"
		empTypeFull        = "Full-time"
		empTypePart        = "Part-time"
		hireDate1          = "2010-05-06"
		hireDate2          = "1999-10-25"
		empID1             = "12345"
		empID2             = "abcde"
		offName1           = "John Smith"
		offName2           = "Johnny"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName,
					generateUserEmployerInfo(
						strconv.Quote(offName1), // Only set official name
						util.NullValue,
						util.NullValue,
						util.NullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.0.official_name", offName1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.0.employee_id", ""),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.0.employee_type", ""),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.0.date_hire", ""),
				),
			},
			{
				// Update with other attributes
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName,
					generateUserEmployerInfo(
						util.NullValue,
						strconv.Quote(empID1),
						strconv.Quote(empTypeFull),
						strconv.Quote(hireDate1),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.0.official_name", ""),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.0.employee_id", empID1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.0.employee_type", empTypeFull),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.0.date_hire", hireDate1),
				),
			},
			{
				// Update all attributes
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName,
					generateUserEmployerInfo(
						strconv.Quote(offName2),
						strconv.Quote(empID2),
						strconv.Quote(empTypePart),
						strconv.Quote(hireDate2),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.0.official_name", offName2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.0.employee_id", empID2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.0.employee_type", empTypePart),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.0.date_hire", hireDate2),
				),
			},
			{
				// Remove all employer info attributes
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName,
					"employer_info = []",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "employer_info.%"),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserroutingUtilBasic(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel1 = "test-user-util"
		userName           = "Terraform Util"
		email1             = "terraform-" + uuid.NewString() + "@user.com"
		maxCapacity0       = "0"
		maxCapacity1       = "10"
		maxCapacity2       = "12"
		utilTypeCall       = "call"
		utilTypeEmail      = "email"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with utilization settings
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName,
					generateUserRoutingUtil(
						routingUtilization.GenerateRoutingUtilMediaType("call", maxCapacity1, util.FalseValue),
						routingUtilization.GenerateRoutingUtilMediaType("callback", maxCapacity1, util.FalseValue),
						routingUtilization.GenerateRoutingUtilMediaType("chat", maxCapacity1, util.FalseValue),
						routingUtilization.GenerateRoutingUtilMediaType("email", maxCapacity1, util.FalseValue),
						routingUtilization.GenerateRoutingUtilMediaType("message", maxCapacity1, util.FalseValue),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserUtilizationLevel(ResourceType+"."+userResourceLabel1, "Agent"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.interruptible_media_types.%"),
				),
			},
			{
				// Update utilization settings and set different org-level settings
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName,
					generateUserRoutingUtil(
						routingUtilization.GenerateRoutingUtilMediaType("call", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeEmail)),
						routingUtilization.GenerateRoutingUtilMediaType("callback", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						routingUtilization.GenerateRoutingUtilMediaType("chat", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						routingUtilization.GenerateRoutingUtilMediaType("email", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						routingUtilization.GenerateRoutingUtilMediaType("message", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserUtilizationLevel(ResourceType+"."+userResourceLabel1, "Agent"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.interruptible_media_types", utilTypeCall),
				),
			},
			{
				// Ensure max capacity can be set to 0
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName,
					generateUserRoutingUtil(
						routingUtilization.GenerateRoutingUtilMediaType("call", maxCapacity0, util.TrueValue, strconv.Quote(utilTypeEmail)),
						routingUtilization.GenerateRoutingUtilMediaType("callback", maxCapacity0, util.TrueValue, strconv.Quote(utilTypeCall)),
						routingUtilization.GenerateRoutingUtilMediaType("chat", maxCapacity0, util.TrueValue, strconv.Quote(utilTypeCall)),
						routingUtilization.GenerateRoutingUtilMediaType("email", maxCapacity0, util.TrueValue, strconv.Quote(utilTypeCall)),
						routingUtilization.GenerateRoutingUtilMediaType("message", maxCapacity0, util.TrueValue, strconv.Quote(utilTypeCall)),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserUtilizationLevel(ResourceType+"."+userResourceLabel1, "Agent"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.interruptible_media_types", utilTypeCall),
				),
			},
			{
				// Reset to org-level settings by specifying empty routing utilization attribute
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName,
					"routing_utilization = []",
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserUtilizationLevel(ResourceType+"."+userResourceLabel1, "Organization"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.%"),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserroutingUtilWithLabels(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel1 = "test-user-util"
		userName           = "Terraform Util"
		email1             = "terraform-" + uuid.NewString() + "@user.com"
		maxCapacity0       = "0"
		maxCapacity1       = "10"
		maxCapacity2       = "12"
		utilTypeCall       = "call"
		utilTypeEmail      = "email"

		redLabelResourceLabel   = "label_red"
		blueLabelResourceLabel  = "label_blue"
		greenLabelResourceLabel = "label_green"
		redLabelName            = "Terraform Red " + uuid.NewString()
		blueLabelName           = "Terraform Blue " + uuid.NewString()
		greenLabelName          = "Terraform Green " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with utilization settings
				Config: routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(redLabelResourceLabel, redLabelName, "") +
					routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(blueLabelResourceLabel, blueLabelName, redLabelResourceLabel) +
					routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(greenLabelResourceLabel, greenLabelName, blueLabelResourceLabel) +
					generateUserWithCustomAttrs(
						userResourceLabel1,
						email1,
						userName,
						generateUserRoutingUtil(
							routingUtilization.GenerateRoutingUtilMediaType("call", maxCapacity1, util.FalseValue),
							routingUtilization.GenerateRoutingUtilMediaType("callback", maxCapacity1, util.FalseValue),
							routingUtilization.GenerateRoutingUtilMediaType("chat", maxCapacity1, util.FalseValue),
							routingUtilization.GenerateRoutingUtilMediaType("email", maxCapacity1, util.FalseValue),
							routingUtilization.GenerateRoutingUtilMediaType("message", maxCapacity1, util.FalseValue),
							routingUtilizationLabel.GenerateLabelUtilization(redLabelResourceLabel, maxCapacity1),
							routingUtilizationLabel.GenerateLabelUtilization(blueLabelResourceLabel, maxCapacity1, redLabelResourceLabel),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					validateUserUtilizationLevel(ResourceType+"."+userResourceLabel1, "Agent"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel1, "routing_utilization.0.label_utilizations.0.label_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.label_utilizations.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel1, "routing_utilization.0.label_utilizations.1.label_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.label_utilizations.1.maximum_capacity", maxCapacity1),
				),
			},
			{
				// Update utilization settings and set different org-level settings
				Config: routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(redLabelResourceLabel, redLabelName, "") +
					routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(blueLabelResourceLabel, blueLabelName, redLabelResourceLabel) +
					routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(greenLabelResourceLabel, greenLabelName, blueLabelResourceLabel) +
					generateUserWithCustomAttrs(
						userResourceLabel1,
						email1,
						userName,
						generateUserRoutingUtil(
							routingUtilization.GenerateRoutingUtilMediaType("call", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeEmail)),
							routingUtilization.GenerateRoutingUtilMediaType("callback", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
							routingUtilization.GenerateRoutingUtilMediaType("chat", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
							routingUtilization.GenerateRoutingUtilMediaType("email", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
							routingUtilization.GenerateRoutingUtilMediaType("message", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
							routingUtilizationLabel.GenerateLabelUtilization(redLabelResourceLabel, maxCapacity2),
							routingUtilizationLabel.GenerateLabelUtilization(blueLabelResourceLabel, maxCapacity2, redLabelResourceLabel),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					validateUserUtilizationLevel(ResourceType+"."+userResourceLabel1, "Agent"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel1, "routing_utilization.0.label_utilizations.0.label_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.label_utilizations.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel1, "routing_utilization.0.label_utilizations.1.label_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.label_utilizations.1.maximum_capacity", maxCapacity2),
				),
			},
			{
				// Ensure max capacity can be set to 0
				Config: routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(redLabelResourceLabel, redLabelName, "") +
					routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(blueLabelResourceLabel, blueLabelName, redLabelResourceLabel) +
					routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(greenLabelResourceLabel, greenLabelName, blueLabelResourceLabel) +
					generateUserWithCustomAttrs(
						userResourceLabel1,
						email1,
						userName,
						generateUserRoutingUtil(
							routingUtilization.GenerateRoutingUtilMediaType("call", maxCapacity0, util.TrueValue, strconv.Quote(utilTypeEmail)),
							routingUtilization.GenerateRoutingUtilMediaType("callback", maxCapacity0, util.TrueValue, strconv.Quote(utilTypeCall)),
							routingUtilization.GenerateRoutingUtilMediaType("chat", maxCapacity0, util.TrueValue, strconv.Quote(utilTypeCall)),
							routingUtilization.GenerateRoutingUtilMediaType("email", maxCapacity0, util.TrueValue, strconv.Quote(utilTypeCall)),
							routingUtilization.GenerateRoutingUtilMediaType("message", maxCapacity0, util.TrueValue, strconv.Quote(utilTypeCall)),
							routingUtilizationLabel.GenerateLabelUtilization(redLabelResourceLabel, maxCapacity0),
							routingUtilizationLabel.GenerateLabelUtilization(blueLabelResourceLabel, maxCapacity0, redLabelResourceLabel),
						),
					),
				Check: resource.ComposeTestCheckFunc(
					validateUserUtilizationLevel(ResourceType+"."+userResourceLabel1, "Agent"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray(ResourceType+"."+userResourceLabel1, "routing_utilization.0.message.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel1, "routing_utilization.0.label_utilizations.0.label_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.label_utilizations.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel1, "routing_utilization.0.label_utilizations.1.label_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.0.label_utilizations.1.maximum_capacity", maxCapacity0),
				),
			},
			{
				// Reset to org-level settings by specifying empty routing utilization attribute
				Config: generateUserWithCustomAttrs(
					userResourceLabel1,
					email1,
					userName,
					"routing_utilization = []",
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserUtilizationLevel(ResourceType+"."+userResourceLabel1, "Organization"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel1, "routing_utilization.%"),
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(45 * time.Second)
			return testVerifyUsersDestroyed(state)
		},
	})
}

func TestAccResourceUserRestore(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel1 = "test-user"
		email1             = "terraform-" + uuid.NewString() + "@user.com"
		userName1          = "Terraform Restore1"
		userName2          = "Terraform Restore2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create a basic user
				Config: GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName1),
				),
			},
			{
				Config: GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName1,
				),
				Destroy: true, // Delete the user
				Check:   testVerifyUsersDestroyed,
			},
			{
				// Restore the same user email but set a different name
				Config: GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName2),
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(45 * time.Second)
			return testVerifyUsersDestroyed(state)
		},
	})
}

func TestAccResourceUserCreateWhenDestroyed(t *testing.T) {
	var (
		userResourceLabel1 = "test-user"
		email1             = "terraform-" + uuid.NewString() + "@user.com"
		userName1          = "Terraform Existing"
		userName2          = "Terraform Create"
		stateActive        = "active"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create a basic user
				Config: GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName1),
				),
			},
			{
				Config: GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName1,
				),
				Destroy: true, // Delete the user
				Check:   testVerifyUsersDestroyed,
			},
			{
				// Restore the same user email but set a different name
				Config: GenerateBasicUserResource(
					userResourceLabel1,
					email1,
					userName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "email", email1),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "name", userName2),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel1, "state", stateActive),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func testVerifyUsersDestroyed(state *terraform.State) error {
	usersAPI := platformclientv2.NewUsersApi()

	diagErr := util.WithRetries(context.Background(), 20*time.Second, func() *retry.RetryError {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != ResourceType {
				continue
			}
			err := checkUserDeleted(rs.Primary.ID)(state)
			if err != nil {
				continue
			}
			_, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "", "")

			if err != nil {
				if util.IsStatus404(resp) {
					continue
				}
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Unexpected error: %s", err), resp))
			}

			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("User (%s) still exists", rs.Primary.ID), resp))
		}
		return nil
	})

	if diagErr != nil {
		return fmt.Errorf(fmt.Sprintf("%v", diagErr))
	}

	// Success. All users destroyed
	return nil
}

func TestAccResourceUserPassword(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-password"
		email             = "terraform-" + uuid.NewString() + "@user.com"
		userName          = "Password Test User"
		initialPassword   = "myInitialPassword123!@#"
		updatedPassword   = "myUpdatedPassword456!@#"

		// Track password updates
		passwordUpdateCalled bool
		lastPasswordUpdate   string
	)

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
		resp, err := originalUpdatePassword(ctx, p, id, password)
		return resp, err
	}

	// Initialize internal proxy
	internalProxy = userProxyInstance
	defer func() {
		internalProxy = nil
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user with initial password
				PreConfig: func() {
					// Reset for next test
					passwordUpdateCalled = false
					lastPasswordUpdate = ""
				},
				Config: generateUserWithCustomAttrs(
					userResourceLabel,
					email,
					userName,
					fmt.Sprintf(`password = "%s"`, initialPassword),
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
				Config: generateUserWithCustomAttrs(
					userResourceLabel,
					email,
					userName,
					fmt.Sprintf(`password = "%s"`, updatedPassword),
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
				Config: generateUserWithCustomAttrs(
					userResourceLabel,
					email,
					userName,
					`password = ""`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email),
					resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName),
					resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "id"),
					func(state *terraform.State) error {
						if passwordUpdateCalled {
							return fmt.Errorf("expected password update to not be called for password update")
						}
						return nil
					},
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserAddressWithExtensionPool(t *testing.T) {
	var (
		addrUserResourceLabel1      = "test-user-addr-ext-pool"
		addrUserName                = "Tim Cheese"
		addrEmail1                  = "terraform-" + uuid.NewString() + "@user.com"
		addrExt1                    = "4105"
		addrExt2                    = "4225"
		phoneMediaType              = "PHONE"
		addrTypeWork                = "WORK"
		extensionPoolResourceLabel1 = "test-extensionpool" + uuid.NewString()
		extensionPoolStartNumber1   = "4100"
		extensionPoolEndNumber1     = "4199"
		extensionPoolResourceLabel2 = "test2-extensionpool" + uuid.NewString()
		extensionPoolStartNumber2   = "4200"
		extensionPoolEndNumber2     = "4299"
	)

	extensionPoolResource1 := extensionPool.ExtensionPoolStruct{
		ResourceLabel: extensionPoolResourceLabel1,
		StartNumber:   extensionPoolStartNumber1,
		EndNumber:     extensionPoolEndNumber1,
		Description:   util.NullValue, // No description
	}
	extensionPoolResource2 := extensionPool.ExtensionPoolStruct{
		ResourceLabel: extensionPoolResourceLabel2,
		StartNumber:   extensionPoolStartNumber2,
		EndNumber:     extensionPoolEndNumber2,
		Description:   util.NullValue, // No description
	}

	extensionPool.DeleteExtensionPoolWithNumber(extensionPoolStartNumber1)
	extensionPool.DeleteExtensionPoolWithNumber(extensionPoolStartNumber2)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateUserWithCustomAttrs(
					addrUserResourceLabel1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							util.NullValue,          // number
							util.NullValue,          // Default to type PHONE
							util.NullValue,          // Default to type WORK
							strconv.Quote(addrExt1), // extension
							fmt.Sprintf("extension_pool_id = %s.%s.id", extensionPool.ResourceType, extensionPoolResourceLabel1),
						),
					),
				) + extensionPool.GenerateExtensionPoolResource(&extensionPoolResource1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "email", addrEmail1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "name", addrUserName),
					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.number"),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.extension", addrExt1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.type", addrTypeWork),
					resource.TestCheckResourceAttrPair(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.extension_pool_id",
						extensionPool.ResourceType+"."+extensionPoolResourceLabel1, "id"),
				),
			},
			{
				// Update
				Config: generateUserWithCustomAttrs(
					addrUserResourceLabel1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							util.NullValue,
							util.NullValue,          // Default to type PHONE
							util.NullValue,          // Default to type WORK
							strconv.Quote(addrExt2), // extension
							fmt.Sprintf("extension_pool_id = %s.%s.id", extensionPool.ResourceType, extensionPoolResourceLabel2),
						),
					),
				) + extensionPool.GenerateExtensionPoolResource(&extensionPoolResource1) + extensionPool.GenerateExtensionPoolResource(&extensionPoolResource2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "email", addrEmail1),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "name", addrUserName),
					resource.TestCheckNoResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.number"),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.extension", addrExt2),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.type", addrTypeWork),
					resource.TestCheckResourceAttrPair(ResourceType+"."+addrUserResourceLabel1, "addresses.0.phone_numbers.0.extension_pool_id",
						extensionPool.ResourceType+"."+extensionPoolResourceLabel2, "id"),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func checkUserDeleted(id string) resource.TestCheckFunc {
	log.Printf("Fetching user with ID: %s\n", id)
	return func(s *terraform.State) error {
		maxAttempts := 30
		for i := 0; i < maxAttempts; i++ {

			deleted, err := isUserDeleted(id)
			if err != nil {
				return err
			}
			if deleted {
				return nil
			}
			time.Sleep(10 * time.Second)
		}
		return fmt.Errorf("user %s was not deleted properly", id)
	}
}

func isUserDeleted(id string) (bool, error) {
	sdkConfig, _ := provider.AuthorizeSdk()
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)
	// Attempt to get the user
	_, response, err := usersAPI.GetUser(id, nil, "", "")

	// Check if the user is not found (deleted)
	if response != nil && response.StatusCode == 404 {
		return true, nil // User is deleted
	}

	// Handle other errors
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return false, err
	}

	// If user is found, it means the user is not deleted
	return false, nil
}

func validateUserSkill(userResourcePath string, skillResourcePath string, proficiency string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		userResource, ok := state.RootModule().Resources[userResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find user %s in state", userResourcePath)
		}
		userID := userResource.Primary.ID

		skillResource, ok := state.RootModule().Resources[skillResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find skill %s in state", skillResourcePath)
		}
		skillID := skillResource.Primary.ID

		numSkillsAttr, ok := userResource.Primary.Attributes["routing_skills.#"]
		if !ok {
			return fmt.Errorf("No skills found for user %s in state", userID)
		}

		numSkills, _ := strconv.Atoi(numSkillsAttr)
		for i := 0; i < numSkills; i++ {
			if userResource.Primary.Attributes["routing_skills."+strconv.Itoa(i)+".skill_id"] == skillID {
				if userResource.Primary.Attributes["routing_skills."+strconv.Itoa(i)+".proficiency"] == proficiency {
					// Found skill with correct proficiency
					return nil
				}
				return fmt.Errorf("Skill %s found for user %s with incorrect proficiency", skillID, userID)
			}
		}

		return fmt.Errorf("Skill %s not found for user %s in state", skillID, userID)
	}
}

func validateUserLanguage(userResourcePath string, langResourcePath string, proficiency string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		userResource, ok := state.RootModule().Resources[userResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find user %s in state", userResourcePath)
		}
		userID := userResource.Primary.ID

		langResource, ok := state.RootModule().Resources[langResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find language %s in state", langResourcePath)
		}
		langID := langResource.Primary.ID

		numLangAttr, ok := userResource.Primary.Attributes["routing_languages.#"]
		if !ok {
			return fmt.Errorf("No languages found for user %s in state", userID)
		}

		numLangs, _ := strconv.Atoi(numLangAttr)
		for i := 0; i < numLangs; i++ {
			if userResource.Primary.Attributes["routing_languages."+strconv.Itoa(i)+".language_id"] == langID {
				if userResource.Primary.Attributes["routing_languages."+strconv.Itoa(i)+".proficiency"] == proficiency {
					// Found language with correct proficiency
					return nil
				}
				return fmt.Errorf("Language %s found for user %s with incorrect proficiency", langID, userID)
			}
		}

		return fmt.Errorf("Language %s not found for user %s in state", langID, userID)
	}
}

func validateUserUtilizationLevel(userResourcePath string, level string) resource.TestCheckFunc {
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

func generateUserAddresses(nestedBlocks ...string) string {
	return fmt.Sprintf(`addresses {
		%s
	}
	`, strings.Join(nestedBlocks, "\n"))
}

func generateUserEmployerInfo(offName string, empID string, empType string, dateHire string) string {
	return fmt.Sprintf(`employer_info {
		official_name = %s
		employee_id = %s
		employee_type = %s
		date_hire = %s
	}
	`, offName, empID, empType, dateHire)
}

func generateUserRoutingUtil(nestedBlocks ...string) string {
	return fmt.Sprintf(`routing_utilization {
		%s
	}
	`, strings.Join(nestedBlocks, "\n"))
}

func generateUserPhoneAddress(phoneNum string, phoneMediaType string, phoneType string, extension string, extras ...string) string {
	return fmt.Sprintf(`phone_numbers {
				number = %s
				media_type = %s
				type = %s
				extension = %s
				%s
			}
			`, phoneNum, phoneMediaType, phoneType, extension, strings.Join(extras, "\n"))
}

func generateUserEmailAddress(emailAddress string, emailType string) string {
	return fmt.Sprintf(`other_emails {
				address = %s
				type = %s
			}
			`, emailAddress, emailType)
}

func generateUserRoutingSkill(skillID string, proficiency string) string {
	return fmt.Sprintf(`routing_skills {
		skill_id = %s
		proficiency = %s
	}
	`, skillID, proficiency)
}

func generateUserRoutingLang(langID string, proficiency string) string {
	return fmt.Sprintf(`routing_languages {
		language_id = %s
		proficiency = %s
	}
	`, langID, proficiency)
}

func generateUserLocation(locResource string, notes string) string {
	return fmt.Sprintf(`locations {
				location_id = %s
				notes = %s
			}
			`, locResource, notes)
}
