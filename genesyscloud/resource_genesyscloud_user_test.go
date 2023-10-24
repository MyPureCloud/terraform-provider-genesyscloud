package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceUserBasic(t *testing.T) {
	t.Parallel()
	var (
		userResource1 = "test-user1"
		userResource2 = "test-user2"
		email1        = "terraform-" + uuid.NewString() + "@example.com"
		email2        = "terraform-" + uuid.NewString() + "@example.com"
		email3        = "terraform-" + uuid.NewString() + "@example.com"
		userName1     = "John Terraform"
		userName2     = "Jim Terraform"
		stateActive   = "active"
		stateInactive = "inactive"
		title1        = "Senior Director"
		title2        = "Project Manager"
		department1   = "Development"
		department2   = "Project Management"
		profileSkill1 = "Java"
		profileSkill2 = "Go"
		cert1         = "AWS Dev"
		cert2         = "AWS Architect"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateUserResource(
					userResource1,
					email1,
					userName1,
					nullValue, // Defaults to active
					strconv.Quote(title1),
					strconv.Quote(department1),
					nullValue, // No manager
					nullValue, // Default acdAutoAnswer
					"",        // No profile skills
					"",        // No certs
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", userName1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "state", stateActive),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "title", title1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "department", department1),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "password.%"),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "manager", ""),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "acd_auto_answer", "false"),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "profile_skills.%"),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "certifications.%"),
					TestDefaultHomeDivision("genesyscloud_user."+userResource1),
				),
			},
			{
				// Update
				Config: GenerateUserResource(
					userResource1,
					email2,
					userName2,
					strconv.Quote(stateInactive),
					strconv.Quote(title2),
					strconv.Quote(department2),
					nullValue, // No manager
					trueValue, // AcdAutoAnswer
					strconv.Quote(profileSkill1),
					strconv.Quote(cert1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", userName2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "state", stateInactive),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "title", title2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "department", department2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "manager", ""),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "acd_auto_answer", "true"),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "profile_skills.0", profileSkill1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "certifications.0", cert1),
					TestDefaultHomeDivision("genesyscloud_user."+userResource1),
				),
			},
			{
				// Create another user and set manager as existing user
				Config: GenerateUserResource(
					userResource1,
					email2,
					userName2,
					strconv.Quote(stateInactive),
					strconv.Quote(title2),
					strconv.Quote(department2),
					nullValue,  // No manager
					falseValue, // AcdAutoAnswer
					strconv.Quote(profileSkill2),
					strconv.Quote(cert2),
				) + GenerateUserResource(
					userResource2,
					email3,
					userName1,
					nullValue, // Active
					strconv.Quote(title1),
					strconv.Quote(department1),
					"genesyscloud_user."+userResource1+".id",
					trueValue, // AcdAutoAnswer
					strconv.Quote(profileSkill1),
					strconv.Quote(cert1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource2, "email", email3),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource2, "name", userName1),
					resource.TestCheckResourceAttrPair("genesyscloud_user."+userResource2, "manager", "genesyscloud_user."+userResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "manager", ""),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "profile_skills.0", profileSkill2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "certifications.0", cert2),
				),
			},
			{
				// Remove manager and update profile skills/certs
				Config: GenerateUserResource(
					userResource2,
					email3,
					userName1,
					nullValue, // Active
					strconv.Quote(title1),
					strconv.Quote(department1),
					nullValue,
					falseValue, // AcdAutoAnswer
					"",         // No profile skills
					"",         // No certs
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource2, "email", email3),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource2, "name", userName1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource2, "manager", ""),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource2, "profile_skills.%"),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource2, "certifications.%"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_user." + userResource2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserAddresses(t *testing.T) {
	t.Parallel()
	var (
		addrUserResource1 = "test-user-addr"
		addrUserName      = "Nancy Terraform"
		addrEmail1        = "terraform-" + uuid.NewString() + "@example.com"
		addrEmail2        = "terraform-" + uuid.NewString() + "@example.com"
		addrEmail3        = "terraform-" + uuid.NewString() + "@example.com"
		addrPhone1        = "+13174269078"
		addrPhone2        = "+441434634996"
		addrPhoneExt      = "1234"
		phoneMediaType    = "PHONE"
		smsMediaType      = "SMS"
		addrTypeWork      = "WORK"
		addrTypeHome      = "HOME"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateUserWithCustomAttrs(
					addrUserResource1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							strconv.Quote(addrPhone1),
							nullValue, // Default to type PHONE
							nullValue, // Default to type WORK
							nullValue, // No extension
						),
						generateUserEmailAddress(
							strconv.Quote(addrEmail2),
							strconv.Quote(addrTypeHome),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "email", addrEmail1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "name", addrUserName),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.number", addrPhone1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.type", addrTypeWork),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.other_emails.0.address", addrEmail2),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.other_emails.0.type", addrTypeHome),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_user." + addrUserResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update phone number and other email attributes
				Config: GenerateUserWithCustomAttrs(
					addrUserResource1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							strconv.Quote(addrPhone2),
							strconv.Quote(smsMediaType),
							strconv.Quote(addrTypeHome),
							strconv.Quote(addrPhoneExt),
						),
						generateUserEmailAddress(
							strconv.Quote(addrEmail3),
							strconv.Quote(addrTypeWork),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "email", addrEmail1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "name", addrUserName),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.number", addrPhone2),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.media_type", smsMediaType),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.type", addrTypeHome),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.extension", addrPhoneExt),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.other_emails.0.address", addrEmail3),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.other_emails.0.type", addrTypeWork),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserPhone(t *testing.T) {
	t.Parallel()
	var (
		addrUserResource1 = "test-user-addr"
		addrUserName      = "Nancy Terraform"
		addrEmail1        = "terraform-" + uuid.NewString() + "@example.com"
		addrPhone1        = "+13173271898"
		addrPhone2        = "+13173271899"
		addrExt1          = "353"
		phoneMediaType    = "PHONE"
		addrTypeWork      = "WORK"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateUserWithCustomAttrs(
					addrUserResource1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							nullValue,                 // number
							nullValue,                 // Default to type PHONE
							nullValue,                 // Default to type WORK
							strconv.Quote(addrPhone1), // extension
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "email", addrEmail1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "name", addrUserName),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.extension", addrPhone1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.type", addrTypeWork),
				),
			},
			{
				Config: GenerateUserWithCustomAttrs(
					addrUserResource1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							nullValue,                 // number
							nullValue,                 // Default to type PHONE
							nullValue,                 // Default to type WORK
							strconv.Quote(addrPhone2), // extension
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "email", addrEmail1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "name", addrUserName),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.extension", addrPhone2),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.type", addrTypeWork),
				),
			},
			{
				Config: GenerateUserWithCustomAttrs(
					addrUserResource1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							strconv.Quote(addrPhone2), // number
							nullValue,                 // Default to type PHONE
							nullValue,                 // Default to type WORK
							nullValue,                 // extension
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "email", addrEmail1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "name", addrUserName),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.number", addrPhone2),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.type", addrTypeWork),
				),
			},
			{
				Config: GenerateUserWithCustomAttrs(
					addrUserResource1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							strconv.Quote(addrPhone2), // number
							nullValue,                 // Default to type PHONE
							nullValue,                 // Default to type WORK
							strconv.Quote(addrExt1),   // extension
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "email", addrEmail1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "name", addrUserName),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.number", addrPhone2),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.extension", addrExt1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.type", addrTypeWork),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserSkills(t *testing.T) {
	t.Parallel()
	var (
		userResource1  = "test-user"
		email1         = "terraform-" + uuid.NewString() + "@example.com"
		userName1      = "Skill Terraform"
		skillResource1 = "test-skill-1"
		skillResource2 = "test-skill-2"
		skillName1     = "skill1-" + uuid.NewString()
		skillName2     = "skill2-" + uuid.NewString()
		proficiency1   = "1.5"
		proficiency2   = "2.5"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user with 1 skill
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					generateUserRoutingSkill("genesyscloud_routing_skill."+skillResource1+".id", proficiency1),
				) + generateRoutingSkillResource(skillResource1, skillName1),
				Check: resource.ComposeTestCheckFunc(
					validateUserSkill("genesyscloud_user."+userResource1, "genesyscloud_routing_skill."+skillResource1, proficiency1),
				),
			},
			{
				// Create another skill and add to the user
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					generateUserRoutingSkill("genesyscloud_routing_skill."+skillResource1+".id", proficiency1),
					generateUserRoutingSkill("genesyscloud_routing_skill."+skillResource2+".id", proficiency2),
				) + generateRoutingSkillResource(
					skillResource1,
					skillName1,
				) + generateRoutingSkillResource(
					skillResource2,
					skillName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserSkill("genesyscloud_user."+userResource1, "genesyscloud_routing_skill."+skillResource1, proficiency1),
					validateUserSkill("genesyscloud_user."+userResource1, "genesyscloud_routing_skill."+skillResource2, proficiency2),
				),
			},
			{
				// Remove a skill from the user and modify proficiency
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					generateUserRoutingSkill("genesyscloud_routing_skill."+skillResource2+".id", proficiency1),
				) + generateRoutingSkillResource(
					skillResource2,
					skillName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserSkill("genesyscloud_user."+userResource1, "genesyscloud_routing_skill."+skillResource2, proficiency1),
				),
			},
			{
				// Remove all skills from the user
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					"routing_skills = []",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "skills.%"),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserLanguages(t *testing.T) {
	t.Parallel()
	var (
		userResource1 = "test-user"
		email1        = "terraform-" + uuid.NewString() + "@example.com"
		userName1     = "Lang Terraform"
		langResource1 = "test-lang-1"
		langResource2 = "test-lang-2"
		langName1     = "lang1-" + uuid.NewString()
		langName2     = "lang2-" + uuid.NewString()
		proficiency1  = "1"
		proficiency2  = "2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user with 1 language
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					generateUserRoutingLang("genesyscloud_routing_language."+langResource1+".id", proficiency1),
				) + GenerateRoutingLanguageResource(langResource1, langName1),
				Check: resource.ComposeTestCheckFunc(
					validateUserLanguage("genesyscloud_user."+userResource1, "genesyscloud_routing_language."+langResource1, proficiency1),
				),
			},
			{
				// Create another language and add to the user
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					generateUserRoutingLang("genesyscloud_routing_language."+langResource1+".id", proficiency1),
					generateUserRoutingLang("genesyscloud_routing_language."+langResource2+".id", proficiency2),
				) + GenerateRoutingLanguageResource(
					langResource1,
					langName1,
				) + GenerateRoutingLanguageResource(
					langResource2,
					langName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserLanguage("genesyscloud_user."+userResource1, "genesyscloud_routing_language."+langResource1, proficiency1),
					validateUserLanguage("genesyscloud_user."+userResource1, "genesyscloud_routing_language."+langResource2, proficiency2),
				),
			},
			{
				// Remove a language from the user and modify proficiency
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					generateUserRoutingLang("genesyscloud_routing_language."+langResource2+".id", proficiency1),
				) + GenerateRoutingLanguageResource(
					langResource2,
					langName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserLanguage("genesyscloud_user."+userResource1, "genesyscloud_routing_language."+langResource2, proficiency1),
				),
			},
			{
				// Remove all languages from the user
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					"routing_languages = []",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "routing_languages.%"),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserLocations(t *testing.T) {
	t.Parallel()
	var (
		userResource1 = "test-user-loc"
		email         = "terraform-" + uuid.NewString() + "@example.com"
		userName      = "Loki Terraform"
		locResource1  = "test-location1"
		locResource2  = "test-location2"
		locName1      = "Terraform location" + uuid.NewString()
		locName2      = "Terraform location" + uuid.NewString()
		locNotes1     = "First floor"
		locNotes2     = "Second floor"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create user with a location
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email,
					userName,
					generateUserLocation(
						"genesyscloud_location."+locResource1+".id",
						strconv.Quote(locNotes1),
					),
				) + GenerateLocationResourceBasic(locResource1, locName1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email),
					resource.TestCheckResourceAttrPair("genesyscloud_user."+userResource1, "locations.0.location_id", "genesyscloud_location."+locResource1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "locations.0.notes", locNotes1),
				),
			},
			{
				// Update with a new location
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email,
					userName,
					generateUserLocation(
						"genesyscloud_location."+locResource2+".id",
						strconv.Quote(locNotes2),
					),
				) + GenerateLocationResourceBasic(locResource2, locName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email),
					resource.TestCheckResourceAttrPair("genesyscloud_user."+userResource1, "locations.0.location_id", "genesyscloud_location."+locResource2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "locations.0.notes", locNotes2),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserEmployerInfo(t *testing.T) {
	t.Parallel()
	var (
		userResource1 = "test-user-info"
		userName      = "Info Terraform"
		email1        = "terraform-" + uuid.NewString() + "@example.com"
		empTypeFull   = "Full-time"
		empTypePart   = "Part-time"
		hireDate1     = "2010-05-06"
		hireDate2     = "1999-10-25"
		empID1        = "12345"
		empID2        = "abcde"
		offName1      = "John Smith"
		offName2      = "Johnny"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName,
					generateUserEmployerInfo(
						strconv.Quote(offName1), // Only set official name
						nullValue,
						nullValue,
						nullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", userName),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "employer_info.0.official_name", offName1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "employer_info.0.employee_id", ""),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "employer_info.0.employee_type", ""),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "employer_info.0.date_hire", ""),
				),
			},
			{
				// Update with other attributes
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName,
					generateUserEmployerInfo(
						nullValue,
						strconv.Quote(empID1),
						strconv.Quote(empTypeFull),
						strconv.Quote(hireDate1),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", userName),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "employer_info.0.official_name", ""),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "employer_info.0.employee_id", empID1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "employer_info.0.employee_type", empTypeFull),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "employer_info.0.date_hire", hireDate1),
				),
			},
			{
				// Update all attributes
				Config: GenerateUserWithCustomAttrs(
					userResource1,
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
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", userName),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "employer_info.0.official_name", offName2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "employer_info.0.employee_id", empID2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "employer_info.0.employee_type", empTypePart),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "employer_info.0.date_hire", hireDate2),
				),
			},
			{
				// Remove all employer info attributes
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName,
					"employer_info = []",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", userName),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "employer_info.%"),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserRoutingUtil(t *testing.T) {
	t.Parallel()
	var (
		userResource1 = "test-user-util"
		userName      = "Terraform Util"
		email1        = "terraform-" + uuid.NewString() + "@example.com"
		maxCapacity0  = "0"
		maxCapacity1  = "10"
		maxCapacity2  = "12"
		utilTypeCall  = "call"
		utilTypeEmail = "email"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with utilization settings
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName,
					generateUserRoutingUtil(
						generateRoutingUtilMediaType("call", maxCapacity1, falseValue),
						generateRoutingUtilMediaType("callback", maxCapacity1, falseValue),
						generateRoutingUtilMediaType("chat", maxCapacity1, falseValue),
						generateRoutingUtilMediaType("email", maxCapacity1, falseValue),
						generateRoutingUtilMediaType("message", maxCapacity1, falseValue),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserUtilizationLevel("genesyscloud_user."+userResource1, "Agent"),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.call.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.call.0.include_non_acd", falseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.call.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.callback.0.include_non_acd", falseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.callback.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.chat.0.include_non_acd", falseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.chat.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.email.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.email.0.include_non_acd", falseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.email.0.interruptible_media_types.%"),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.message.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.message.0.include_non_acd", falseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.message.0.interruptible_media_types.%"),
				),
			},
			{
				// Update utilization settings and set different org-level settings
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName,
					generateUserRoutingUtil(
						generateRoutingUtilMediaType("call", maxCapacity2, trueValue, strconv.Quote(utilTypeEmail)),
						generateRoutingUtilMediaType("callback", maxCapacity2, trueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("chat", maxCapacity2, trueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("email", maxCapacity2, trueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("message", maxCapacity2, trueValue, strconv.Quote(utilTypeCall)),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserUtilizationLevel("genesyscloud_user."+userResource1, "Agent"),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.call.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.call.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_user."+userResource1, "routing_utilization.0.call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.callback.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_user."+userResource1, "routing_utilization.0.callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.chat.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_user."+userResource1, "routing_utilization.0.chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.email.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.email.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_user."+userResource1, "routing_utilization.0.email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.message.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.message.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_user."+userResource1, "routing_utilization.0.message.0.interruptible_media_types", utilTypeCall),
				),
			},
			{
				// Ensure max capacity can be set to 0
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName,
					generateUserRoutingUtil(
						generateRoutingUtilMediaType("call", maxCapacity0, trueValue, strconv.Quote(utilTypeEmail)),
						generateRoutingUtilMediaType("callback", maxCapacity0, trueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("chat", maxCapacity0, trueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("email", maxCapacity0, trueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("message", maxCapacity0, trueValue, strconv.Quote(utilTypeCall)),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserUtilizationLevel("genesyscloud_user."+userResource1, "Agent"),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.call.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.call.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_user."+userResource1, "routing_utilization.0.call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.callback.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.callback.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_user."+userResource1, "routing_utilization.0.callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.chat.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.chat.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_user."+userResource1, "routing_utilization.0.chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.email.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.email.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_user."+userResource1, "routing_utilization.0.email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.message.0.maximum_capacity", maxCapacity0),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.0.message.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_user."+userResource1, "routing_utilization.0.message.0.interruptible_media_types", utilTypeCall),
				),
			},
			{
				// Reset to org-level settings by specifying empty routing utilization attribute
				Config: GenerateUserWithCustomAttrs(
					userResource1,
					email1,
					userName,
					"routing_utilization = []",
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserUtilizationLevel("genesyscloud_user."+userResource1, "Organization"),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "routing_utilization.%"),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserRestore(t *testing.T) {
	t.Parallel()
	var (
		userResource1 = "test-user"
		email1        = "terraform-" + uuid.NewString() + "@example.com"
		userName1     = "Terraform Restore1"
		userName2     = "Terraform Restore2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create a basic user
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", userName1),
				),
			},
			{
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				),
				Destroy: true, // Delete the user
				Check:   testVerifyUsersDestroyed,
			},
			{
				// Restore the same user email but set a different name
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", userName2),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserCreateWhenDestroyed(t *testing.T) {
	t.Parallel()
	var (
		userResource1 = "test-user"
		email1        = "terraform-" + uuid.NewString() + "@example.com"
		userName1     = "Terraform Existing"
		userName2     = "Terraform Create"
		stateActive   = "active"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create a basic user
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", userName1),
				),
			},
			{
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName1,
				),
				Destroy: true, // Delete the user
				Check:   testVerifyUsersDestroyed,
			},
			{
				// Restore the same user email but set a different name
				Config: GenerateBasicUserResource(
					userResource1,
					email1,
					userName2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", userName2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "state", stateActive),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func testVerifyUsersDestroyed(state *terraform.State) error {
	usersAPI := platformclientv2.NewUsersApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_user" {
			continue
		}

		user, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "", "")
		if user != nil {
			return fmt.Errorf("User (%s) still exists", rs.Primary.ID)
		} else if IsStatus404(resp) {
			// User not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All users destroyed
	return nil
}

func validateUserSkill(userResourceName string, skillResourceName string, proficiency string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		userResource, ok := state.RootModule().Resources[userResourceName]
		if !ok {
			return fmt.Errorf("Failed to find user %s in state", userResourceName)
		}
		userID := userResource.Primary.ID

		skillResource, ok := state.RootModule().Resources[skillResourceName]
		if !ok {
			return fmt.Errorf("Failed to find skill %s in state", skillResourceName)
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

func validateUserLanguage(userResourceName string, langResourceName string, proficiency string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		userResource, ok := state.RootModule().Resources[userResourceName]
		if !ok {
			return fmt.Errorf("Failed to find user %s in state", userResourceName)
		}
		userID := userResource.Primary.ID

		langResource, ok := state.RootModule().Resources[langResourceName]
		if !ok {
			return fmt.Errorf("Failed to find language %s in state", langResourceName)
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

func validateUserUtilizationLevel(userResourceName string, level string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		userResource, ok := state.RootModule().Resources[userResourceName]
		if !ok {
			return fmt.Errorf("Failed to find user %s in state", userResourceName)
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

func generateUserPhoneAddress(phoneNum string, phoneMediaType string, phoneType string, extension string) string {
	return fmt.Sprintf(`phone_numbers {
				number = %s
				media_type = %s
				type = %s
				extension = %s
			}
			`, phoneNum, phoneMediaType, phoneType, extension)
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
