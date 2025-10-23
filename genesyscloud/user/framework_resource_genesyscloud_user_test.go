package user

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Ensure test resources are initialized for Framework tests
func init() {
	if frameworkResources == nil || frameworkDataSources == nil {
		initTestResources()
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

// Helper function to generate basic Framework user resource configuration
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

// Helper function to generate user with profile attributes
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

// Helper function to generate profile skills
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

// Helper function to generate certifications
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

// Helper function to generate optional attributes
func generateOptionalAttr(attrName string, value string) string {
	if value == util.NullValue || value == "" {
		return ""
	}
	return fmt.Sprintf("%s = %s", attrName, value)
}
func TestAccFrameworkResourceUserAddresses(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel = "test-user-addresses"
		email1            = "terraform-addr-" + uuid.NewString() + "@user.com"
		email2            = "terraform-other-" + uuid.NewString() + "@user.com"
		userName          = "Address User"
		phone1            = "+13174269078"
		phone2            = "+441434634996"
		phoneExt1         = "1234"
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
				// Test multiple phone numbers
				Config: generateFrameworkUserWithMultiplePhones(
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
				),
			},
			{
				// Remove all addresses
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
	}`, skillResourceLabel1, skillName1)

	skillResource2 := fmt.Sprintf(`resource "genesyscloud_routing_skill" "%s" {
		name = "%s"
	}`, skillResourceLabel2, skillName2)

	langResource1 := fmt.Sprintf(`resource "genesyscloud_routing_language" "%s" {
		name = "%s"
	}`, langResourceLabel1, langName1)

	langResource2 := fmt.Sprintf(`resource "genesyscloud_routing_language" "%s" {
		name = "%s"
	}`, langResourceLabel2, langName2)

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

// Helper functions for Framework tests

func generateFrameworkUserWithAddresses(resourceLabel, email, name, phoneAddress, emailAddress string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		addresses = [
			{
				%s
				%s
			}
		]
	}`, ResourceType, resourceLabel, email, name, phoneAddress, emailAddress)
}

func generateFrameworkUserWithMultiplePhones(resourceLabel, email, name string, phoneAddresses ...string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		addresses = [
			{
				%s
			}
		]
	}`, ResourceType, resourceLabel, email, name, strings.Join(phoneAddresses, "\n"))
}

func generateFrameworkUserPhoneAddress(phoneNum, phoneMediaType, phoneType, extension string) string {
	return fmt.Sprintf(`phone_numbers = [
		{
			number = %s
			media_type = %s
			type = %s
			extension = %s
		}
	]`, phoneNum, phoneMediaType, phoneType, extension)
}

func generateFrameworkUserEmailAddress(emailAddress, emailType string) string {
	return fmt.Sprintf(`other_emails = [
		{
			address = %s
			type = %s
		}
	]`, emailAddress, emailType)
}

func generateFrameworkUserWithSkillsAndLanguages(resourceLabel, email, name, skill, language string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		%s
		%s
	}`, ResourceType, resourceLabel, email, name, skill, language)
}

func generateFrameworkUserWithMultipleSkillsAndLanguages(resourceLabel, email, name string, skillsAndLanguages ...string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		%s
	}`, ResourceType, resourceLabel, email, name, strings.Join(skillsAndLanguages, "\n"))
}

func generateFrameworkUserRoutingSkill(skillID, proficiency string) string {
	return fmt.Sprintf(`routing_skills {
		skill_id = %s
		proficiency = %s
	}`, skillID, proficiency)
}

func generateFrameworkUserRoutingLanguage(langID, proficiency string) string {
	return fmt.Sprintf(`routing_languages {
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
	return fmt.Sprintf(`employer_info = [
		{
			official_name = %s
			employee_id = %s
			employee_type = %s
			date_hire = %s
		}
	]`, officialName, employeeId, employeeType, dateHire)
}

func generateFrameworkUserWithVoicemailPolicies(resourceLabel, email, name, voicemailPolicies string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		%s
	}`, ResourceType, resourceLabel, email, name, voicemailPolicies)
}

func generateFrameworkVoicemailUserpolicies(timeoutSeconds int, sendEmailNotifications bool) string {
	return fmt.Sprintf(`voicemail_userpolicies = [
		{
			alert_timeout_seconds = %d
			send_email_notifications = %t
		}
	]`, timeoutSeconds, sendEmailNotifications)
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
				ExpectError: regexp.MustCompile("Phone number must be in E.164 format"),
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
				ExpectError: regexp.MustCompile("Date must be in ISO-8601 format"),
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
	}`, skillResourceLabel, skillName)

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
				ExpectError: regexp.MustCompile("Attribute routing_skills\\[0\\]\\.proficiency value must be between 0\\.000000 and 5\\.000000"),
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
				ExpectError: regexp.MustCompile("Attribute routing_utilization\\[0\\]\\.call\\[0\\]\\.maximum_capacity value must be between 0 and 25"),
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

// Additional helper functions for edge case tests

func generateFrameworkUserWithRoutingUtilization(resourceLabel, email, name, routingUtil string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		email = "%s"
		name = "%s"
		routing_utilization = [
			{
				%s
			}
		]
	}`, ResourceType, resourceLabel, email, name, routingUtil)
}

func generateFrameworkRoutingUtilizationCall(maxCapacity, includeNonAcd string) string {
	return fmt.Sprintf(`call = [
		{
			maximum_capacity = %s
			include_non_acd = %s
		}
	]`, maxCapacity, includeNonAcd)
}

// testVerifyUsersDestroyed verifies that users are properly destroyed after tests
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
