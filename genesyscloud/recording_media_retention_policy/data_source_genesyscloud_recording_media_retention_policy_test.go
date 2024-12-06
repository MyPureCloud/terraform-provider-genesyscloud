package recording_media_retention_policy

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	authRole "terraform-provider-genesyscloud/genesyscloud/auth_role"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailDomain "terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingLanguage "terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingWrapupcode "terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	userRoles "terraform-provider-genesyscloud/genesyscloud/user_roles"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	integration "terraform-provider-genesyscloud/genesyscloud/integration"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the Recording media Retention Policy Data Source
*/
func TestAccDataSourceRecordingMediaRetentionPolicy(t *testing.T) {
	var (
		policyResourceLabel     = "recording-media-retention-policy"
		policyDataResourceLabel = "recording-media-retention-policy-data"

		policyName = "terraform-policy-" + uuid.NewString()
	)

	basePolicy := Policycreate{
		Name:        policyName,
		Order:       0,
		Description: "a media retention policy",
		Enabled:     true,
	}

	mediaRetentionChatPolicy := basePolicy
	mediaRetentionChatPolicy.MediaPolicies = Mediapolicies{
		ChatPolicy: Chatmediapolicy{
			Actions: Policyactions{
				RetainRecording: true,
				DeleteRecording: false,
				AlwaysDelete:    false,
				AssignEvaluations: []Evaluationassignment{
					{
						User: User{},
					},
				},
				AssignMeteredEvaluations: []Meteredevaluationassignment{
					{
						Evaluators:           []User{{}},
						MaxNumberEvaluations: 1,
						AssignToActiveUser:   true,
						TimeInterval: EvalTimeinterval{
							Days:  1,
							Hours: 1,
						},
					},
				},
				AssignMeteredAssignmentByAgent: []Meteredassignmentbyagent{
					{
						Evaluators:           []User{{}},
						MaxNumberEvaluations: 1,
						TimeInterval: AgentTimeinterval{
							Months: 1,
							Weeks:  1,
							Days:   1,
						},
						TimeZone: "EST",
					},
				},
				AssignCalibrations: []Calibrationassignment{
					{
						Evaluators: []User{{}},
					},
				},
				AssignSurveys: []Surveyassignment{
					{
						SendingDomain: "genesyscloud_routing_email_domain.routing-domain1.domain_id",
						SurveyForm:    Publishedsurveyformreference{},
					},
				},
				RetentionDuration: Retentionduration{
					ArchiveRetention: Archiveretention{
						Days:          1,
						StorageMedium: "CLOUDARCHIVE",
					},
					DeleteRetention: Deleteretention{
						Days: 3,
					},
				},
				InitiateScreenRecording: Initiatescreenrecording{
					RecordACW: true,
					ArchiveRetention: Archiveretention{
						Days:          1,
						StorageMedium: "CLOUDARCHIVE",
					},
					DeleteRetention: Deleteretention{
						Days: 3,
					},
				},
				IntegrationExport: Integrationexport{
					ShouldExportScreenRecordings: true,
				},
			},
			Conditions: Chatmediapolicyconditions{
				DateRanges: []string{
					"2022-05-12T04:00:00.000Z/2022-05-13T04:00:00.000Z",
				},
				ForUsers:  []User{{}},
				ForQueues: []Queue{{}},
				TimeAllowed: Timeallowed{
					TimeSlots: []Timeslot{
						{
							StartTime: "10:10:10.010",
							StopTime:  "11:11:11.011",
							Day:       3,
						},
					},
					TimeZoneId: "Europe/Paris",
					Empty:      false,
				},
				WrapupCodes: []Wrapupcode{{}},
				Languages:   []Language{{}},
			},
		},
	}

	var (
		domainResourceLabel = "routing-domain1"
		domainId            = "terraformmedia" + strconv.Itoa(rand.Intn(1000)) + ".com"
		divResourceLabel    = "test-division"
		divName             = "terraform-" + uuid.NewString()
	)

	_, err := provider.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	CleanupRoutingEmailDomains()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: routingEmailDomain.GenerateRoutingEmailDomainResource(
					domainResourceLabel,
					domainId,
					util.FalseValue, // Subdomain
					util.NullValue,
				) + routingQueue.GenerateRoutingQueueResourceBasic(queueResourceLabel1, queueName, "") +
					authRole.GenerateAuthRoleResource(
						roleResourceLabel1,
						roleName1,
						roleDesc1,
						authRole.GenerateRolePermissions(permissions...),
						authRole.GenerateRolePermPolicy(qualityDomain, evaluationEntityType, strconv.Quote(editAction)),
						authRole.GenerateRolePermPolicy(qualityDomain, calibrationEntityType, strconv.Quote(addAction)),
					) +
					userRoles.GenerateUserRoles(
						userRoleResourceLabel1,
						userResourceLabel1,
						generateResourceRoles("genesyscloud_auth_role."+roleResourceLabel1+".id"),
					) +
					generateUserWithCustomAttrs(userResourceLabel1, userEmail, userName) +
					gcloud.GenerateEvaluationFormResource(evaluationFormResourceLabel1, &evaluationFormResourceBody) +
					gcloud.GenerateSurveyFormResource(surveyFormResourceLabel1, &surveyFormResourceBody) +
					integration.GenerateIntegrationResource(integrationResourceLabel1, strconv.Quote(integrationIntendedState), strconv.Quote(integrationType), "") +
					routingLanguage.GenerateRoutingLanguageResource(languageResourceLabel1, languageName) +
					authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					routingWrapupcode.GenerateRoutingWrapupcodeResource(wrapupCodeResourceLabel1, wrapupCodeName, "genesyscloud_auth_division."+divResourceLabel+".id") +
					architect_flow.GenerateFlowResource(
						flowResourceLabel1,
						filePath1,
						"",
						false,
						util.GenerateSubstitutionsMap(map[string]string{
							"flow_name":            flowName,
							"default_language":     "en-us",
							"greeting":             "Archy says hi!!!",
							"menu_disconnect_name": "Disconnect",
						}),
					) +
					generateMediaRetentionPolicyResource(
						policyResourceLabel, &mediaRetentionChatPolicy,
					) +
					generateRecordingMediaRetentionPolicyDataSource(
						policyDataResourceLabel,
						policyName,
						"genesyscloud_recording_media_retention_policy."+policyResourceLabel,
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_recording_media_retention_policy."+policyDataResourceLabel, "id", "genesyscloud_recording_media_retention_policy."+policyResourceLabel, "id"),
				),
			},
		},
		CheckDestroy: testVerifyMediaRetentionPolicyDestroyed,
	})
}

func generateRecordingMediaRetentionPolicyDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOn string,
) string {
	return fmt.Sprintf(`data "genesyscloud_recording_media_retention_policy" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, resourceLabel, name, dependsOn)
}

func generateResourceRoles(skillID string, divisionIds ...string) string {
	var divAttr string
	if len(divisionIds) > 0 {
		divAttr = "division_ids = [" + strings.Join(divisionIds, ",") + "]"
	}
	return fmt.Sprintf(`roles {
		role_id = %s
		%s
	}
	`, skillID, divAttr)
}

func generateUserWithCustomAttrs(resourceLabel string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceLabel, email, name, strings.Join(attrs, "\n"))
}
