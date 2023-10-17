package recording_media_retention_policy

import (
	"fmt"
	"math/rand"
	"strconv"
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
		policyResource     = "recording-media-retention-policy"
		policyDataResource = "recording-media-retention-policy-data"

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
						TimeInterval: Timeinterval{
							Months: 1,
							Weeks:  1,
							Days:   1,
							Hours:  1,
						},
					},
				},
				AssignMeteredAssignmentByAgent: []Meteredassignmentbyagent{
					{
						Evaluators:           []User{{}},
						MaxNumberEvaluations: 1,
						TimeInterval: Timeinterval{
							Months: 1,
							Weeks:  1,
							Days:   1,
							Hours:  1,
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
		domainRes = "routing-domain1"
		domainId  = "terraform" + strconv.Itoa(rand.Intn(1000)) + ".com"
	)

	_, err := gcloud.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	CleanupRoutingEmailDomains()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: gcloud.GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					falseValue, // Subdomain
					nullValue,
				) + gcloud.GenerateRoutingQueueResourceBasic(queueResource1, queueName, "") +
					gcloud.GenerateAuthRoleResource(
						roleResource1,
						roleName1,
						roleDesc1,
						gcloud.GenerateRolePermissions(permissions...),
						gcloud.GenerateRolePermPolicy(qualityDomain, evaluationEntityType, strconv.Quote(editAction)),
						gcloud.GenerateRolePermPolicy(qualityDomain, calibrationEntityType, strconv.Quote(addAction)),
					) +
					gcloud.GenerateUserRoles(
						userRoleResource1,
						userResource1,
						gcloud.GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
					) +
					gcloud.GenerateUserWithCustomAttrs(userResource1, userEmail, userName) +
					gcloud.GenerateEvaluationFormResource(evaluationFormResource1, &evaluationFormResourceBody) +
					gcloud.GenerateSurveyFormResource(surveyFormResource1, &surveyFormResourceBody) +
					integration.GenerateIntegrationResource(integrationResource1, strconv.Quote(integrationIntendedState), strconv.Quote(integrationType), "") +
					gcloud.GenerateRoutingLanguageResource(languageResource1, languageName) +
					gcloud.GenerateRoutingWrapupcodeResource(wrapupCodeResource1, wrapupCodeName) +
					gcloud.GenerateFlowResource(
						flowResource1,
						filePath1,
						"",
						false,
						gcloud.GenerateSubstitutionsMap(map[string]string{
							"flow_name":            flowName,
							"default_language":     "en-us",
							"greeting":             "Archy says hi!!!",
							"menu_disconnect_name": "Disconnect",
						}),
					) +
					generateMediaRetentionPolicyResource(
						policyResource, &mediaRetentionChatPolicy,
					) +
					generateRecordingMediaRetentionPolicyDataSource(
						policyDataResource,
						policyName,
						"genesyscloud_recording_media_retention_policy."+policyResource,
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_recording_media_retention_policy."+policyDataResource, "id", "genesyscloud_recording_media_retention_policy."+policyResource, "id"),
				),
			},
		},
	})
}

func generateRecordingMediaRetentionPolicyDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOn string,
) string {
	return fmt.Sprintf(`data "genesyscloud_recording_media_retention_policy" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, resourceID, name, dependsOn)
}
