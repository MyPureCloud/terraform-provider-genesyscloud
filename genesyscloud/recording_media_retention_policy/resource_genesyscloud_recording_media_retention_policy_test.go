package recording_media_retention_policy

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	authRole "terraform-provider-genesyscloud/genesyscloud/auth_role"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingLanguage "terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingEmailDomain "terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	userRoles "terraform-provider-genesyscloud/genesyscloud/user_roles"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	integration "terraform-provider-genesyscloud/genesyscloud/integration"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_recording_media_retention_policy_test.go contains all of the test cases for running the resource
tests for integration_credentials.
*/

type Mediapolicies struct {
	CallPolicy    Callmediapolicy
	ChatPolicy    Chatmediapolicy
	EmailPolicy   Emailmediapolicy
	MessagePolicy Messagemediapolicy
}

type Chatmediapolicy struct {
	Actions    Policyactions
	Conditions Chatmediapolicyconditions
}

type Callmediapolicy struct {
	Actions    Policyactions
	Conditions Callmediapolicyconditions
}

type Emailmediapolicy struct {
	Actions    Policyactions
	Conditions Emailmediapolicyconditions
}

type Messagemediapolicy struct {
	Actions    Policyactions
	Conditions Messagemediapolicyconditions
}

type Policyactions struct {
	RetainRecording                bool
	DeleteRecording                bool
	AlwaysDelete                   bool
	AssignEvaluations              []Evaluationassignment
	AssignMeteredEvaluations       []Meteredevaluationassignment
	AssignMeteredAssignmentByAgent []Meteredassignmentbyagent
	AssignCalibrations             []Calibrationassignment
	AssignSurveys                  []Surveyassignment
	RetentionDuration              Retentionduration
	InitiateScreenRecording        Initiatescreenrecording
	MediaTranscriptions            []Mediatranscription
	IntegrationExport              Integrationexport
}

type Evaluationassignment struct {
	EvaluationForm Evaluationform
	User           User
}

type Meteredevaluationassignment struct {
	Evaluators           []User
	MaxNumberEvaluations int
	EvaluationForm       Evaluationform
	AssignToActiveUser   bool
	TimeInterval         EvalTimeinterval
}

type Meteredassignmentbyagent struct {
	Evaluators           []User
	MaxNumberEvaluations int
	EvaluationForm       Evaluationform
	TimeInterval         AgentTimeinterval
	TimeZone             string
}

type Calibrationassignment struct {
	Calibrator      User
	Evaluators      []User
	EvaluationForm  Evaluationform
	ExpertEvaluator User
}

type Surveyassignment struct {
	SurveyForm         Publishedsurveyformreference
	Flow               Domainentityref
	InviteTimeInterval string
	SendingUser        string
	SendingDomain      string
}

type Retentionduration struct {
	ArchiveRetention Archiveretention
	DeleteRetention  Deleteretention
}

type Initiatescreenrecording struct {
	RecordACW        bool
	ArchiveRetention Archiveretention
	DeleteRetention  Deleteretention
}

type Mediatranscription struct {
	DisplayName           string
	TranscriptionProvider string
	IntegrationId         string
}

type Integrationexport struct {
	Integration                  Domainentityref
	ShouldExportScreenRecordings bool
}

type Evaluationform struct {
	Id             string
	Name           string
	ModifiedDate   time.Time
	Published      bool
	ContextId      string
	QuestionGroups []Evaluationquestiongroup
	SelfUri        string
}

type User struct {
	Id string
}

type Servicelevel struct {
	Percentage float64
	DurationMs int
}

type Mediasetting struct {
	AlertingTimeoutSeconds int
	ServiceLevel           Servicelevel
}

type Queue struct {
	Id string
}

type Publishedsurveyformreference struct {
	Id        string
	Name      string
	ContextId string
	SelfUri   string
}

type Domainentityref struct {
	Id      string
	Name    string
	SelfUri string
}

type Evaluationquestiongroup struct {
	Id                      string
	Name                    string
	VarType                 string
	DefaultAnswersToHighest bool
	DefaultAnswersToNA      bool
	NaEnabled               bool
	Weight                  float32
	ManualWeight            bool
	Questions               []Evaluationquestion
	VisibilityCondition     Visibilitycondition
}

type Evaluationquestion struct {
	Id                  string
	Text                string
	HelpText            string
	VarType             string
	NaEnabled           bool
	CommentsRequired    bool
	VisibilityCondition Visibilitycondition
	AnswerOptions       []Answeroption
	IsKill              bool
	IsCritical          bool
}

type Visibilitycondition struct {
	CombiningOperation string
	Predicates         []string
}

type Answeroption struct {
	Id    string
	Text  string
	Value int
}

type Inboundroute struct {
	Id        string
	Name      string
	Pattern   string
	Queue     Domainentityref
	Priority  int
	Skills    []Domainentityref
	Language  Domainentityref
	FromName  string
	FromEmail string
	Flow      Domainentityref
	AutoBcc   []Emailaddress
	SpamFlow  Domainentityref
}

type Emailaddress struct {
	Email string
	Name  string
}

type AgentTimeinterval struct {
	Months int
	Weeks  int
	Days   int
}

type EvalTimeinterval struct {
	Days  int
	Hours int
}

type Policyerrors struct {
	PolicyErrorMessages []Policyerrormessage
}

type Policyerrormessage struct {
	StatusCode        int
	UserMessage       []string
	UserParamsMessage string
	ErrorCode         string
	CorrelationId     string
	UserParams        []Userparam
	InsertDate        time.Time
}

type Userparam struct {
	Key   string
	Value string
}

type Archiveretention struct {
	Days          int
	StorageMedium string
}

type Deleteretention struct {
	Days int
}
type Division struct {
	Id      string
	Name    string
	SelfUri string
}

type Chat struct {
	JabberId string
}

type Contact struct {
	Address     string
	MediaType   string
	VarType     string
	Extension   string
	CountryCode string
	Integration string
}

type Userimage struct {
	Resolution string
	ImageUri   string
}

type Biography struct {
	Biography string
	Interests []string
	Hobbies   []string
	Spouse    string
	Education []Education
}

type Employerinfo struct {
	OfficialName string
	EmployeeId   string
	EmployeeType string
	DateHire     string
}

type Oauthlasttokenissued struct {
	DateIssued time.Time
}

type Routingrule struct {
	Operator    string
	Threshold   int
	WaitSeconds float64
}

type Bullseye struct {
	Rings []Ring
}

type Acwsettings struct {
	WrapupPrompt string
	TimeoutMs    int
}

type Queuemessagingaddresses struct {
	SmsAddress Domainentityref
}

type Queueemailaddress struct {
	Domain Domainentityref
	Route  Inboundroute
}

type Wrapupcode struct {
	Id string
}

type Language struct {
	Id string
}

type Timeslot struct {
	StartTime string
	StopTime  string
	Day       int
}

type Timeallowed struct {
	TimeSlots  []Timeslot
	TimeZoneId string
	Empty      bool
}

type Durationcondition struct {
	DurationTarget   string
	DurationOperator string
	DurationRange    string
	DurationMode     string
}

type Education struct {
	School       string
	FieldOfStudy string
	Notes        string
	DateStart    time.Time
	DateEnd      time.Time
}

type Ring struct {
	ExpansionCriteria []Expansioncriterium
	Actions           Actions
}

type Expansioncriterium struct {
	VarType   string
	Threshold float64
}

type Skillstoremove struct {
	Name    string
	Id      string
	SelfUri string
}

type Actions struct {
	SkillsToRemove []Skillstoremove
}

type Callmediapolicyconditions struct {
	ForUsers    []User
	DateRanges  []string
	ForQueues   []Queue
	WrapupCodes []Wrapupcode
	Languages   []Language
	TimeAllowed Timeallowed
	Directions  []string
	Duration    Durationcondition
}

type Chatmediapolicyconditions struct {
	ForUsers    []User
	DateRanges  []string
	ForQueues   []Queue
	WrapupCodes []Wrapupcode
	Languages   []Language
	TimeAllowed Timeallowed
	Duration    Durationcondition
}

type Emailmediapolicyconditions struct {
	ForUsers    []User
	DateRanges  []string
	ForQueues   []Queue
	WrapupCodes []Wrapupcode
	Languages   []Language
	TimeAllowed Timeallowed
}

type Messagemediapolicyconditions struct {
	ForUsers    []User
	DateRanges  []string
	ForQueues   []Queue
	WrapupCodes []Wrapupcode
	Languages   []Language
	TimeAllowed Timeallowed
}

type Policyconditions struct {
	ForUsers    []User
	Directions  []string
	DateRanges  []string
	MediaTypes  []string
	ForQueues   []Queue
	Duration    Durationcondition
	WrapupCodes []Wrapupcode
	TimeAllowed Timeallowed
}

type Policycreate struct {
	Name          string
	Order         int
	Description   string
	Enabled       bool
	MediaPolicies Mediapolicies
	Conditions    Policyconditions
	Actions       Policyactions
	PolicyErrors  Policyerrors
	SelfUri       string
}

var (
	policyResource1          = "test-media-retention-policy-1"
	queueResource1           = "test-queue-1"
	queueName                = "terraform-queue" + uuid.NewString()
	userResource1            = "test-user-1"
	userName                 = "terraform-user" + uuid.NewString()
	userEmail                = "notanemail@example.com" + uuid.NewString()
	userRoleResource1        = "test-user-role-2"
	questionGroupName        = "terraform-question-group" + uuid.NewString()
	evaluationFormResource1  = "test-evaluation-form-1"
	evaluationFormName       = "terraform-evaluation-form" + uuid.NewString()
	surveyFormResource1      = "test-survey-form-1"
	surveyFormName           = "terraform-survey-form" + uuid.NewString()
	integrationResource1     = "test-integration-resource-1"
	integrationType          = "purecloud-data-actions"
	integrationIntendedState = "ENABLED"
	flowResource1            = "test-flow-resource-1"
	flowName                 = "terraform-flow" + uuid.NewString()
	filePath1                = "../../examples/resources/genesyscloud_flow/inboundcall_flow_example_substitutions.yaml"
	languageResource1        = "test-language-1"
	languageName             = "terraform-language" + uuid.NewString()
	wrapupCodeResource1      = "test-wrapup-code-1"
	wrapupCodeName           = "terraform-wrapup-code" + uuid.NewString()
	roleResource1            = "auth-role1"
	roleName1                = "terraform-role" + uuid.NewString()
	roleDesc1                = "Terraform test role"
	perm1                    = "evaluation"
	perm2                    = "calibration"
	qualityDomain            = "quality"
	evaluationEntityType     = "evaluation"
	calibrationEntityType    = "calibration"
	editAction               = "edit"
	addAction                = "add"
	roleActions              = make([]string, 0)
	permissions              = make([]string, 0)

	questionGroupBody1 = []gcloud.EvaluationFormQuestionGroupStruct{
		{
			Name: questionGroupName,
			Questions: []gcloud.EvaluationFormQuestionStruct{
				{
					Text: "question-1",
					AnswerOptions: []gcloud.AnswerOptionStruct{
						{
							Text: "yes",
						},
						{
							Text: "yes",
						},
					},
				},
			},
		},
	}
	evaluationFormResourceBody = gcloud.EvaluationFormStruct{
		Name:           evaluationFormName,
		Published:      true,
		QuestionGroups: questionGroupBody1,
	}

	questionGroupBody2 = []gcloud.SurveyFormQuestionGroupStruct{
		{
			Name: questionGroupName,
			Questions: []gcloud.SurveyFormQuestionStruct{
				{
					Text:                  "question-1",
					VarType:               "freeTextQuestion",
					MaxResponseCharacters: 1000,
				},
			},
		},
	}
	surveyFormResourceBody = gcloud.SurveyFormStruct{
		Name:           surveyFormName,
		Language:       "en-US",
		Published:      true,
		QuestionGroups: questionGroupBody2,
	}
)

func TestAccResourceMediaRetentionPolicyBasic(t *testing.T) {
	permissions = append(permissions, strconv.Quote(perm1))
	permissions = append(permissions, strconv.Quote(perm2))
	roleActions = append(roleActions, strconv.Quote(editAction))
	roleActions = append(roleActions, strconv.Quote(addAction))

	basePolicy := Policycreate{
		Name:        "terraform-media-retention-policy" + uuid.NewString(),
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

	mediaRetentionCallPolicy := basePolicy
	mediaRetentionCallPolicy.MediaPolicies = Mediapolicies{
		CallPolicy: Callmediapolicy{
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
			Conditions: Callmediapolicyconditions{
				DateRanges: []string{
					"2022-05-12T04:00:00.000Z/2022-05-13T04:00:00.000Z",
				},
				Directions: []string{
					"INBOUND",
				},
				ForQueues: []Queue{{}},
				ForUsers:  []User{{}},
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

	mediaRetentionMessagePolicy := basePolicy
	mediaRetentionMessagePolicy.MediaPolicies = Mediapolicies{
		MessagePolicy: Messagemediapolicy{
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
						Evaluators: []User{
							{},
						},
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
			Conditions: Messagemediapolicyconditions{
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

	mediaRetentionEmailPolicy := basePolicy
	mediaRetentionEmailPolicy.MediaPolicies = Mediapolicies{
		EmailPolicy: Emailmediapolicy{
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
			Conditions: Emailmediapolicyconditions{
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
		domainId  = fmt.Sprintf("terraformmedia%v.com", time.Now().Unix())
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
					domainRes,
					domainId,
					util.FalseValue, // Subdomain
					util.NullValue,
				) + routingQueue.GenerateRoutingQueueResourceBasic(queueResource1, queueName, "") +
					authRole.GenerateAuthRoleResource(
						roleResource1,
						roleName1,
						roleDesc1,
						authRole.GenerateRolePermissions(permissions...),
						authRole.GenerateRolePermPolicy(qualityDomain, evaluationEntityType, strconv.Quote(editAction)),
						authRole.GenerateRolePermPolicy(qualityDomain, calibrationEntityType, strconv.Quote(addAction)),
					) +
					userRoles.GenerateUserRoles(
						userRoleResource1,
						userResource1,
						GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
					) +
					generateUserWithCustomAttrs(userResource1, userEmail, userName) +
					gcloud.GenerateEvaluationFormResource(evaluationFormResource1, &evaluationFormResourceBody) +
					gcloud.GenerateSurveyFormResource(surveyFormResource1, &surveyFormResourceBody) +
					integration.GenerateIntegrationResource(integrationResource1, strconv.Quote(integrationIntendedState), strconv.Quote(integrationType), "") +
					routingLanguage.GenerateRoutingLanguageResource(languageResource1, languageName) +
					gcloud.GenerateRoutingWrapupcodeResource(wrapupCodeResource1, wrapupCodeName) +
					architect_flow.GenerateFlowResource(
						flowResource1,
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
					generateMediaRetentionPolicyResource(policyResource1, &mediaRetentionCallPolicy),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "name", mediaRetentionCallPolicy.Name),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "description", mediaRetentionCallPolicy.Description),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "enabled", util.TrueValue),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.retain_recording", strconv.FormatBool(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.RetainRecording)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.delete_recording", strconv.FormatBool(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.DeleteRecording)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.always_delete", strconv.FormatBool(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AlwaysDelete)),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_evaluations.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignEvaluations))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_evaluations.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_evaluations.0.user_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredEvaluations))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.0.max_number_evaluations", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredEvaluations[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.0.evaluator_ids.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredEvaluations[0].Evaluators))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.0.evaluator_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.0.time_interval.0.days", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredEvaluations[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredEvaluations))),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredAssignmentByAgent))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.0.max_number_evaluations", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredAssignmentByAgent[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluator_ids.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredAssignmentByAgent[0].Evaluators))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluator_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.0.time_interval.0.days", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredAssignmentByAgent[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredAssignmentByAgent))),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_calibrations.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignCalibrations))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_calibrations.0.calibrator_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_calibrations.0.evaluator_ids.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignCalibrations[0].Evaluators))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_calibrations.0.evaluator_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_calibrations.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_calibrations.0.expert_evaluator_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_surveys.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignSurveys))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_surveys.0.survey_form_name", surveyFormName),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_surveys.0.flow_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.retention_duration.0.archive_retention.0.storage_medium", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.RetentionDuration.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.retention_duration.0.delete_retention.0.days", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.RetentionDuration.DeleteRetention.Days)),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.initiate_screen_recording.0.archive_retention.0.storage_medium", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.InitiateScreenRecording.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.initiate_screen_recording.0.delete_retention.0.days", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.InitiateScreenRecording.DeleteRetention.Days)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.initiate_screen_recording.0.record_acw", strconv.FormatBool(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.InitiateScreenRecording.RecordACW)),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.for_queue_ids.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.ForQueues))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.for_queue_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.for_user_ids.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.ForUsers))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.for_user_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.wrapup_code_ids.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.WrapupCodes))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.wrapup_code_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.language_ids.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.Languages))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.language_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.time_allowed.0.time_zone_id", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.TimeAllowed.TimeZoneId),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.time_allowed.0.time_slots.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.TimeAllowed.TimeSlots))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.time_allowed.0.time_slots.0.start_time", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.TimeAllowed.TimeSlots[0].StartTime),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.date_ranges.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.DateRanges))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.date_ranges.0", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.DateRanges[0])),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.directions.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.Directions))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.directions.0", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.Directions[0])),
				),
			},
			{

				Config: routingEmailDomain.GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue, // Subdomain
					util.NullValue,
				) + routingQueue.GenerateRoutingQueueResourceBasic(queueResource1, queueName, "") +
					authRole.GenerateAuthRoleResource(
						roleResource1,
						roleName1,
						roleDesc1,
						authRole.GenerateRolePermissions(permissions...),
						authRole.GenerateRolePermPolicy(qualityDomain, evaluationEntityType, strconv.Quote(editAction)),
						authRole.GenerateRolePermPolicy(qualityDomain, calibrationEntityType, strconv.Quote(addAction)),
					) +
					userRoles.GenerateUserRoles(
						userRoleResource1,
						userResource1,
						GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
					) +
					generateUserWithCustomAttrs(userResource1, userEmail, userName) +
					gcloud.GenerateEvaluationFormResource(evaluationFormResource1, &evaluationFormResourceBody) +
					gcloud.GenerateSurveyFormResource(surveyFormResource1, &surveyFormResourceBody) +
					integration.GenerateIntegrationResource(integrationResource1, strconv.Quote(integrationIntendedState), strconv.Quote(integrationType), "") +
					routingLanguage.GenerateRoutingLanguageResource(languageResource1, languageName) +
					gcloud.GenerateRoutingWrapupcodeResource(wrapupCodeResource1, wrapupCodeName) +
					architect_flow.GenerateFlowResource(
						flowResource1,
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
					generateMediaRetentionPolicyResource(policyResource1, &mediaRetentionChatPolicy),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "name", mediaRetentionCallPolicy.Name),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "description", mediaRetentionCallPolicy.Description),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "enabled", util.TrueValue),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.retain_recording", strconv.FormatBool(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.RetainRecording)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.delete_recording", strconv.FormatBool(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.DeleteRecording)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.always_delete", strconv.FormatBool(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AlwaysDelete)),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_evaluations.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignEvaluations))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_evaluations.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_evaluations.0.user_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredEvaluations))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.0.max_number_evaluations", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredEvaluations[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.0.evaluator_ids.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredEvaluations[0].Evaluators))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.0.evaluator_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.0.time_interval.0.days", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredEvaluations[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredEvaluations))),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredAssignmentByAgent))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.0.max_number_evaluations", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredAssignmentByAgent[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluator_ids.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredAssignmentByAgent[0].Evaluators))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluator_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.0.time_interval.0.days", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredAssignmentByAgent[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredAssignmentByAgent))),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_calibrations.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignCalibrations))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_calibrations.0.calibrator_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_calibrations.0.evaluator_ids.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignCalibrations[0].Evaluators))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_calibrations.0.evaluator_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_calibrations.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_calibrations.0.expert_evaluator_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_surveys.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignSurveys))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_surveys.0.survey_form_name", surveyFormName),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_surveys.0.flow_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.retention_duration.0.archive_retention.0.storage_medium", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.RetentionDuration.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.retention_duration.0.delete_retention.0.days", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.RetentionDuration.DeleteRetention.Days)),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.initiate_screen_recording.0.archive_retention.0.storage_medium", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.InitiateScreenRecording.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.initiate_screen_recording.0.delete_retention.0.days", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.InitiateScreenRecording.DeleteRetention.Days)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.initiate_screen_recording.0.record_acw", strconv.FormatBool(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.InitiateScreenRecording.RecordACW)),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.for_queue_ids.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.ForQueues))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.for_queue_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.for_user_ids.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.ForUsers))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.for_user_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.wrapup_code_ids.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.WrapupCodes))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.wrapup_code_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.language_ids.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.Languages))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.language_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.time_allowed.0.time_zone_id", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.TimeAllowed.TimeZoneId),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.time_allowed.0.time_slots.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.TimeAllowed.TimeSlots))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.time_allowed.0.time_slots.0.start_time", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.TimeAllowed.TimeSlots[0].StartTime),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.date_ranges.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.DateRanges))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.date_ranges.0", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.DateRanges[0])),
				),
			},
			{

				Config: routingEmailDomain.GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue, // Subdomain
					util.NullValue,
				) + routingQueue.GenerateRoutingQueueResourceBasic(queueResource1, queueName, "") +
					authRole.GenerateAuthRoleResource(
						roleResource1,
						roleName1,
						roleDesc1,
						authRole.GenerateRolePermissions(permissions...),
						authRole.GenerateRolePermPolicy(qualityDomain, evaluationEntityType, strconv.Quote(editAction)),
						authRole.GenerateRolePermPolicy(qualityDomain, calibrationEntityType, strconv.Quote(addAction)),
					) +
					userRoles.GenerateUserRoles(
						userRoleResource1,
						userResource1,
						GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
					) +
					generateUserWithCustomAttrs(userResource1, userEmail, userName) +
					gcloud.GenerateEvaluationFormResource(evaluationFormResource1, &evaluationFormResourceBody) +
					gcloud.GenerateSurveyFormResource(surveyFormResource1, &surveyFormResourceBody) +
					integration.GenerateIntegrationResource(integrationResource1, strconv.Quote(integrationIntendedState), strconv.Quote(integrationType), "") +
					routingLanguage.GenerateRoutingLanguageResource(languageResource1, languageName) +
					gcloud.GenerateRoutingWrapupcodeResource(wrapupCodeResource1, wrapupCodeName) +
					architect_flow.GenerateFlowResource(
						flowResource1,
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
					generateMediaRetentionPolicyResource(policyResource1, &mediaRetentionMessagePolicy),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "name", mediaRetentionCallPolicy.Name),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "description", mediaRetentionCallPolicy.Description),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "enabled", util.TrueValue),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.retain_recording", strconv.FormatBool(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.RetainRecording)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.delete_recording", strconv.FormatBool(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.DeleteRecording)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.always_delete", strconv.FormatBool(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AlwaysDelete)),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_evaluations.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignEvaluations))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_evaluations.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_evaluations.0.user_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredEvaluations))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.0.max_number_evaluations", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredEvaluations[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.0.evaluator_ids.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredEvaluations[0].Evaluators))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.0.evaluator_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.0.time_interval.0.days", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredEvaluations[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredEvaluations))),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredAssignmentByAgent))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.0.max_number_evaluations", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredAssignmentByAgent[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluator_ids.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredAssignmentByAgent[0].Evaluators))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluator_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.0.time_interval.0.days", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredAssignmentByAgent[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredAssignmentByAgent))),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_calibrations.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignCalibrations))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_calibrations.0.calibrator_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_calibrations.0.evaluator_ids.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignCalibrations[0].Evaluators))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_calibrations.0.evaluator_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_calibrations.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_calibrations.0.expert_evaluator_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_surveys.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignSurveys))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_surveys.0.survey_form_name", surveyFormName),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_surveys.0.flow_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.retention_duration.0.archive_retention.0.storage_medium", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.RetentionDuration.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.retention_duration.0.delete_retention.0.days", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.RetentionDuration.DeleteRetention.Days)),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.initiate_screen_recording.0.archive_retention.0.storage_medium", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.InitiateScreenRecording.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.initiate_screen_recording.0.delete_retention.0.days", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.InitiateScreenRecording.DeleteRetention.Days)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.initiate_screen_recording.0.record_acw", strconv.FormatBool(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.InitiateScreenRecording.RecordACW)),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.for_queue_ids.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.ForQueues))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.for_queue_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.for_user_ids.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.ForUsers))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.for_user_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.wrapup_code_ids.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.WrapupCodes))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.wrapup_code_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.language_ids.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.Languages))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.language_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.time_allowed.0.time_zone_id", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.TimeAllowed.TimeZoneId),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.time_allowed.0.time_slots.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.TimeAllowed.TimeSlots))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.time_allowed.0.time_slots.0.start_time", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.TimeAllowed.TimeSlots[0].StartTime),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.date_ranges.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.DateRanges))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.date_ranges.0", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.DateRanges[0])),
				),
			},
			{

				Config: routingEmailDomain.GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue, // Subdomain
					util.NullValue,
				) + routingQueue.GenerateRoutingQueueResourceBasic(queueResource1, queueName, "") +
					authRole.GenerateAuthRoleResource(
						roleResource1,
						roleName1,
						roleDesc1,
						authRole.GenerateRolePermissions(permissions...),
						authRole.GenerateRolePermPolicy(qualityDomain, evaluationEntityType, strconv.Quote(editAction)),
						authRole.GenerateRolePermPolicy(qualityDomain, calibrationEntityType, strconv.Quote(addAction)),
					) +
					userRoles.GenerateUserRoles(
						userRoleResource1,
						userResource1,
						GenerateResourceRoles("genesyscloud_auth_role."+roleResource1+".id"),
					) +
					generateUserWithCustomAttrs(userResource1, userEmail, userName) +
					gcloud.GenerateEvaluationFormResource(evaluationFormResource1, &evaluationFormResourceBody) +
					gcloud.GenerateSurveyFormResource(surveyFormResource1, &surveyFormResourceBody) +
					integration.GenerateIntegrationResource(integrationResource1, strconv.Quote(integrationIntendedState), strconv.Quote(integrationType), "") +
					routingLanguage.GenerateRoutingLanguageResource(languageResource1, languageName) +
					gcloud.GenerateRoutingWrapupcodeResource(wrapupCodeResource1, wrapupCodeName) +
					architect_flow.GenerateFlowResource(
						flowResource1,
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
					generateMediaRetentionPolicyResource(policyResource1, &mediaRetentionEmailPolicy),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "name", mediaRetentionCallPolicy.Name),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "description", mediaRetentionCallPolicy.Description),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "enabled", util.TrueValue),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.retain_recording", strconv.FormatBool(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.RetainRecording)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.delete_recording", strconv.FormatBool(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.DeleteRecording)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.always_delete", strconv.FormatBool(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AlwaysDelete)),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_evaluations.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignEvaluations))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_evaluations.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_evaluations.0.user_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredEvaluations))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.0.max_number_evaluations", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredEvaluations[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.0.evaluator_ids.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredEvaluations[0].Evaluators))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.0.evaluator_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.0.time_interval.0.days", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredEvaluations[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredEvaluations))),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredAssignmentByAgent))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.0.max_number_evaluations", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredAssignmentByAgent[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluator_ids.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredAssignmentByAgent[0].Evaluators))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluator_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.0.time_interval.0.days", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredAssignmentByAgent[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredAssignmentByAgent))),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_calibrations.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignCalibrations))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_calibrations.0.calibrator_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_calibrations.0.evaluator_ids.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignCalibrations[0].Evaluators))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_calibrations.0.evaluator_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_calibrations.0.evaluation_form_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_calibrations.0.expert_evaluator_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_surveys.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignSurveys))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_surveys.0.survey_form_name", surveyFormName),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_surveys.0.flow_id", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.retention_duration.0.archive_retention.0.storage_medium", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.RetentionDuration.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.retention_duration.0.delete_retention.0.days", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.RetentionDuration.DeleteRetention.Days)),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.initiate_screen_recording.0.archive_retention.0.storage_medium", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.InitiateScreenRecording.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.initiate_screen_recording.0.delete_retention.0.days", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.InitiateScreenRecording.DeleteRetention.Days)),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.initiate_screen_recording.0.record_acw", strconv.FormatBool(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.InitiateScreenRecording.RecordACW)),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.for_queue_ids.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.ForQueues))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.for_queue_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.for_user_ids.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.ForUsers))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.for_user_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.wrapup_code_ids.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.WrapupCodes))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.wrapup_code_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.language_ids.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.Languages))),
					resource.TestMatchResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.language_ids.0", regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.time_allowed.0.time_zone_id", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.TimeAllowed.TimeZoneId),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.time_allowed.0.time_slots.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.TimeAllowed.TimeSlots))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.time_allowed.0.time_slots.0.start_time", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.TimeAllowed.TimeSlots[0].StartTime),

					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.date_ranges.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.DateRanges))),
					resource.TestCheckResourceAttr("genesyscloud_recording_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.date_ranges.0", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.DateRanges[0])),
				),
			},
			{

				ResourceName:            "genesyscloud_recording_media_retention_policy." + policyResource1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"order"},
			},
		},
		CheckDestroy: testVerifyMediaRetentionPolicyDestroyed,
	})
}

func testVerifyMediaRetentionPolicyDestroyed(state *terraform.State) error {
	recordingAPI := platformclientv2.NewRecordingApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_recording_media_retention_policy" {
			continue
		}

		form, resp, err := recordingAPI.GetRecordingMediaretentionpolicy(rs.Primary.ID)
		if form != nil {
			continue
		}

		if form != nil {
			return fmt.Errorf("Policy (%s) still exists", rs.Primary.ID)
		}

		if util.IsStatus404(resp) {
			// Policy not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All Media retention policies destroyed
	return nil
}

func generateMediaRetentionPolicyResource(resourceID string, mediaRetentionPolicy *Policycreate) string {
	policy := fmt.Sprintf(`resource "genesyscloud_recording_media_retention_policy" "%s" {
		name = "%s"
		order = %v
        description = "%s"
        enabled = %v
		%s
        %s
        %s
		%s
		%s
	}
	`, resourceID,
		mediaRetentionPolicy.Name,
		mediaRetentionPolicy.Order,
		mediaRetentionPolicy.Description,
		mediaRetentionPolicy.Enabled,
		generateMediaPolicies(&mediaRetentionPolicy.MediaPolicies),
		generateConditions(&mediaRetentionPolicy.Conditions),
		generatePolicyActions(&mediaRetentionPolicy.Actions),
		generatePolicyErrors(&mediaRetentionPolicy.PolicyErrors),
		generateMediaRetentionPolicyLifeCycle(),
	)

	return policy
}

func generateMediaRetentionPolicyLifeCycle() string {
	return `
	lifecycle {
		ignore_changes = [
			"media_policies[0].call_policy[0].actions[0].assign_evaluations[0].evaluation_form_id",                  
			"media_policies[0].call_policy[0].actions[0].assign_calibrations[0].evaluation_form_id",              
			"media_policies[0].call_policy[0].actions[0].assign_metered_evaluations[0].evaluation_form_id",       
			"media_policies[0].call_policy[0].actions[0].assign_metered_assignment_by_agent[0].evaluation_form_id",
			"media_policies[0].chat_policy[0].actions[0].assign_evaluations[0].evaluation_form_id",             
			"media_policies[0].chat_policy[0].actions[0].assign_calibrations[0].evaluation_form_id",            
			"media_policies[0].chat_policy[0].actions[0].assign_metered_evaluations[0].evaluation_form_id",      
			"media_policies[0].chat_policy[0].actions[0].assign_metered_assignment_by_agent[0].evaluation_form_id", 
			"media_policies[0].message_policy[0].actions[0].assign_evaluations[0].evaluation_form_id",        
			"media_policies[0].message_policy[0].actions[0].assign_calibrations[0].evaluation_form_id",       
			"media_policies[0].message_policy[0].actions[0].assign_metered_evaluations[0].evaluation_form_id", 
			"media_policies[0].message_policy[0].actions[0].assign_metered_assignment_by_agent[0].evaluation_form_id",
			"media_policies[0].email_policy[0].actions[0].assign_evaluations[0].evaluation_form_id",               
			"media_policies[0].email_policy[0].actions[0].assign_calibrations[0].evaluation_form_id",               
			"media_policies[0].email_policy[0].actions[0].assign_metered_evaluations[0].evaluation_form_id",     
			"media_policies[0].email_policy[0].actions[0].assign_metered_assignment_by_agent[0].evaluation_form_id",
			"actions[0].assign_evaluations[0].evaluation_form_id",                                         
			"actions[0].assign_calibrations[0].evaluation_form_id",                                       
			"actions[0].assign_metered_evaluations[0].evaluation_form_id",                              
			"actions[0].assign_metered_assignment_by_agent[0].evaluation_form_id",                           
		]
	}
	`
}

func generatePolicyErrors(policyErrors *Policyerrors) string {
	if reflect.DeepEqual(*policyErrors, Policyerrors{}) {
		return ""
	}

	policyErrorsString := fmt.Sprintf(`
        policy_errors {
			%s
        }
        `, generatePolicyErrorMessages(&policyErrors.PolicyErrorMessages),
	)

	return policyErrorsString
}

func generatePolicyErrorMessages(messages *[]Policyerrormessage) string {
	if messages == nil || len(*messages) <= 0 {
		return ""
	}

	messagesString := ""
	for _, message := range *messages {
		userMessageString := ""
		for i, v := range message.UserMessage {
			if i > 0 {
				userMessageString += ", "
			}
			userMessageString += v
		}

		messageString := fmt.Sprintf(`
        	policy_error_messages {
        	    status_code = %v
        	    user_message = [%s]
        	    user_params_message = "%s"
        	    error_code = "%s"
        	    correlation_id = "%s"
        	    %s
				insert_date = %v
        	}
        	`, message.StatusCode,
			userMessageString,
			message.UserParamsMessage,
			message.ErrorCode,
			message.CorrelationId,
			generateUserParams(&message.UserParams),
			message.InsertDate,
		)

		messagesString += messageString
	}

	return messagesString
}

func generateUserParams(params *[]Userparam) string {
	if params == nil || len(*params) <= 0 {
		return ""
	}

	paramsString := ""
	for _, param := range *params {
		paramString := fmt.Sprintf(`
        	user_params {
        	    key = "%s"
        	    value = "%s"
        	}
        	`, param.Key,
			param.Value,
		)

		paramsString += paramString
	}

	return paramsString
}

func generateConditions(conditions *Policyconditions) string {
	if reflect.DeepEqual(*conditions, Policyconditions{}) {
		return ""
	}

	dateRangesString := ""
	for i, dateRange := range conditions.DateRanges {
		if i > 0 {
			dateRangesString += ", "
		}

		dateRangesString += strconv.Quote(dateRange)
	}

	directionsString := ""
	for i, direction := range conditions.Directions {
		if i > 0 {
			directionsString += ", "
		}

		directionsString += strconv.Quote(direction)
	}

	mediaTypesString := ""
	for i, mediaType := range conditions.MediaTypes {
		if i > 0 {
			mediaTypesString += ", "
		}

		mediaTypesString += strconv.Quote(mediaType)
	}

	userIdsString := ""
	for i := range conditions.ForUsers {
		if i > 0 {
			userIdsString += ", "
		}
		userIdsString += fmt.Sprintf("genesyscloud_user.%s.id", userResource1)
	}

	queueIdsString := ""
	for i := range conditions.ForQueues {
		if i > 0 {
			queueIdsString += ", "
		}
		queueIdsString += fmt.Sprintf("genesyscloud_routing_queue.%s.id", queueResource1)
	}

	wrapupCodeIdsString := ""
	for i := range conditions.WrapupCodes {
		if i > 0 {
			wrapupCodeIdsString += ", "
		}
		wrapupCodeIdsString += fmt.Sprintf("genesyscloud_routing_wrapupcode.%s.id", wrapupCodeResource1)
	}

	conditionsString := fmt.Sprintf(`
        conditions {
			for_user_ids = [%s]
			directions = [%s]
			date_ranges = [%s]
			media_types = [%s]
            for_queue_ids = [%s]
			%s
			wrapup_code_ids = [%s]
			%s
        }
        `, userIdsString,
		directionsString,
		dateRangesString,
		mediaTypesString,
		queueIdsString,
		generateDuration(&conditions.Duration),
		wrapupCodeIdsString,
		generateTimeAllowed(&conditions.TimeAllowed),
	)

	return conditionsString
}

func generateTimeAllowed(timeAllowed *Timeallowed) string {
	if reflect.DeepEqual(*timeAllowed, Timeallowed{}) {
		return ""
	}

	timeAllowedString := fmt.Sprintf(`
        time_allowed {
            %s
			time_zone_id = "%s"
			empty = %v
        }
        `, generateTimeSlots(&timeAllowed.TimeSlots),
		timeAllowed.TimeZoneId,
		timeAllowed.Empty,
	)

	return timeAllowedString
}

func generateTimeSlots(slots *[]Timeslot) string {
	if slots == nil || len(*slots) <= 0 {
		return ""
	}

	slotsString := ""
	for _, slot := range *slots {
		slotString := fmt.Sprintf(`
        	time_slots {
        	    start_time = "%s"
        	    stop_time = "%s"
        	    day = %v
        	}
        	`, slot.StartTime,
			slot.StopTime,
			slot.Day,
		)

		slotsString += slotString
	}

	return slotsString
}

func generateDuration(duration *Durationcondition) string {
	if reflect.DeepEqual(*duration, Durationcondition{}) {
		return ""
	}

	durationString := fmt.Sprintf(`
        duration {
            duration_target = "%s"
			duration_operator = "%s"
			duration_range = "%s"
        }
        `, duration.DurationTarget,
		duration.DurationOperator,
		duration.DurationRange,
	)

	return durationString
}

func generateMediaPolicies(mediaPolicies *Mediapolicies) string {
	if reflect.DeepEqual(*mediaPolicies, Mediapolicies{}) {
		return ""
	}

	mediaPoliciesString := fmt.Sprintf(`
        media_policies {
            %s
			%s
			%s
			%s
        }
        `, generateCallPolicy(&mediaPolicies.CallPolicy),
		generateChatPolicy(&mediaPolicies.ChatPolicy),
		generateMessagePolicy(&mediaPolicies.MessagePolicy),
		generateEmailPolicy(&mediaPolicies.EmailPolicy),
	)

	return mediaPoliciesString
}

func generateCallPolicy(callPolicy *Callmediapolicy) string {
	if reflect.DeepEqual(*callPolicy, Callmediapolicy{}) {
		return ""
	}

	callPolicyString := fmt.Sprintf(`
        call_policy {
            %s
            %s
        }
        `, generatePolicyActions(&callPolicy.Actions),
		generateCallMediaPolicyConditions(&callPolicy.Conditions),
	)

	return callPolicyString
}

func generateChatPolicy(chatPolicy *Chatmediapolicy) string {
	if reflect.DeepEqual(*chatPolicy, Chatmediapolicy{}) {
		return ""
	}

	chatPolicyString := fmt.Sprintf(`
        chat_policy {
            %s
            %s
        }
        `, generatePolicyActions(&chatPolicy.Actions),
		generateChatMediaPolicyConditions(&chatPolicy.Conditions),
	)

	return chatPolicyString
}

func generateMessagePolicy(messagePolicy *Messagemediapolicy) string {
	if reflect.DeepEqual(*messagePolicy, Messagemediapolicy{}) {
		return ""
	}

	messagePolicyString := fmt.Sprintf(`
        message_policy {
            %s
            %s
        }
        `, generatePolicyActions(&messagePolicy.Actions),
		generateMessageMediaPolicyConditions(&messagePolicy.Conditions),
	)

	return messagePolicyString
}

func generateEmailPolicy(emailPolicy *Emailmediapolicy) string {
	if reflect.DeepEqual(*emailPolicy, Emailmediapolicy{}) {
		return ""
	}

	emailPolicyString := fmt.Sprintf(`
        email_policy {
            %s
            %s
        }
        `, generatePolicyActions(&emailPolicy.Actions),
		generateEmailMediaPolicyConditions(&emailPolicy.Conditions),
	)

	return emailPolicyString
}

func generateEmailMediaPolicyConditions(conditions *Emailmediapolicyconditions) string {
	if reflect.DeepEqual(*conditions, Emailmediapolicyconditions{}) {
		return ""
	}

	dateRangesString := ""
	for i, dateRange := range conditions.DateRanges {
		if i > 0 {
			dateRangesString += ", "
		}

		dateRangesString += strconv.Quote(dateRange)
	}

	userIdsString := ""
	for i := range conditions.ForUsers {
		if i > 0 {
			userIdsString += ", "
		}
		userIdsString += fmt.Sprintf("genesyscloud_user.%s.id", userResource1)
	}

	queueIdsString := ""
	for i := range conditions.ForQueues {
		if i > 0 {
			queueIdsString += ", "
		}
		queueIdsString += fmt.Sprintf("genesyscloud_routing_queue.%s.id", queueResource1)
	}

	languageIdsString := ""
	for i := range conditions.Languages {
		if i > 0 {
			languageIdsString += ", "
		}
		languageIdsString += fmt.Sprintf("genesyscloud_routing_language.%s.id", languageResource1)
	}

	wrapupCodeIdsString := ""
	for i := range conditions.WrapupCodes {
		if i > 0 {
			wrapupCodeIdsString += ", "
		}
		wrapupCodeIdsString += fmt.Sprintf("genesyscloud_routing_wrapupcode.%s.id", wrapupCodeResource1)
	}

	conditionsString := fmt.Sprintf(`
        conditions {
			for_user_ids = [%s]
			date_ranges = [%s]
            for_queue_ids = [%s]
			wrapup_code_ids = [%s]
			language_ids = [%s]
			%s
        }
        `, userIdsString,
		dateRangesString,
		queueIdsString,
		wrapupCodeIdsString,
		languageIdsString,
		generateTimeAllowed(&conditions.TimeAllowed),
	)

	return conditionsString
}

func generateMessageMediaPolicyConditions(conditions *Messagemediapolicyconditions) string {
	if reflect.DeepEqual(*conditions, Messagemediapolicyconditions{}) {
		return ""
	}

	dateRangesString := ""
	for i, dateRange := range conditions.DateRanges {
		if i > 0 {
			dateRangesString += ", "
		}

		dateRangesString += strconv.Quote(dateRange)
	}

	userIdsString := ""
	for i := range conditions.ForUsers {
		if i > 0 {
			userIdsString += ", "
		}
		userIdsString += fmt.Sprintf("genesyscloud_user.%s.id", userResource1)
	}

	queueIdsString := ""
	for i := range conditions.ForQueues {
		if i > 0 {
			queueIdsString += ", "
		}
		queueIdsString += fmt.Sprintf("genesyscloud_routing_queue.%s.id", queueResource1)
	}

	languageIdsString := ""
	for i := range conditions.Languages {
		if i > 0 {
			languageIdsString += ", "
		}
		languageIdsString += fmt.Sprintf("genesyscloud_routing_language.%s.id", languageResource1)
	}

	wrapupCodeIdsString := ""
	for i := range conditions.WrapupCodes {
		if i > 0 {
			wrapupCodeIdsString += ", "
		}
		wrapupCodeIdsString += fmt.Sprintf("genesyscloud_routing_wrapupcode.%s.id", wrapupCodeResource1)
	}

	conditionsString := fmt.Sprintf(`
        conditions {
			for_user_ids = [%s]
			date_ranges = [%s]
            for_queue_ids = [%s]
			wrapup_code_ids = [%s]
			language_ids=[%s]
			%s
        }
        `, userIdsString,
		dateRangesString,
		queueIdsString,
		wrapupCodeIdsString,
		languageIdsString,
		generateTimeAllowed(&conditions.TimeAllowed),
	)

	return conditionsString
}

func generateChatMediaPolicyConditions(conditions *Chatmediapolicyconditions) string {
	if reflect.DeepEqual(*conditions, Chatmediapolicyconditions{}) {
		return ""
	}

	dateRangesString := ""
	for i, dateRange := range conditions.DateRanges {
		if i > 0 {
			dateRangesString += ", "
		}

		dateRangesString += strconv.Quote(dateRange)
	}

	userIdsString := ""
	for i := range conditions.ForUsers {
		if i > 0 {
			userIdsString += ", "
		}
		userIdsString += fmt.Sprintf("genesyscloud_user.%s.id", userResource1)
	}

	queueIdsString := ""
	for i := range conditions.ForQueues {
		if i > 0 {
			queueIdsString += ", "
		}
		queueIdsString += fmt.Sprintf("genesyscloud_routing_queue.%s.id", queueResource1)
	}

	languageIdsString := ""
	for i := range conditions.Languages {
		if i > 0 {
			languageIdsString += ", "
		}
		languageIdsString += fmt.Sprintf("genesyscloud_routing_language.%s.id", languageResource1)
	}

	wrapupCodeIdsString := ""
	for i := range conditions.WrapupCodes {
		if i > 0 {
			wrapupCodeIdsString += ", "
		}
		wrapupCodeIdsString += fmt.Sprintf("genesyscloud_routing_wrapupcode.%s.id", wrapupCodeResource1)
	}

	conditionsString := fmt.Sprintf(`
        conditions {
			for_user_ids = [%s]
			date_ranges = [%s]
            for_queue_ids = [%s]
			wrapup_code_ids = [%s]
			language_ids = [%s]
			%s
			%s
        }
        `, userIdsString,
		dateRangesString,
		queueIdsString,
		wrapupCodeIdsString,
		languageIdsString,
		generateTimeAllowed(&conditions.TimeAllowed),
		generateDuration(&conditions.Duration),
	)

	return conditionsString
}

func generatePolicyActions(actions *Policyactions) string {
	if reflect.DeepEqual(*actions, Policyactions{}) {
		return ""
	}

	actionsString := fmt.Sprintf(`
        actions {
            retain_recording = %v
            delete_recording = %v
            always_delete = %v
            %s
            %s
            %s
            %s
            %s
            %s
            %s
            %s
            %s
        }
        `, actions.RetainRecording,
		actions.DeleteRecording,
		actions.AlwaysDelete,
		generateAssignEvaluations(),
		generateAssignMeteredEvaluations(&actions.AssignMeteredEvaluations),
		generateAssignMeteredAssignmentByAgent(&actions.AssignMeteredAssignmentByAgent),
		generateAssignCalibrations(&actions.AssignCalibrations),
		generateAssignSurveys(&actions.AssignSurveys),
		generateRetentionDuration(&actions.RetentionDuration),
		generateInitiateScreenRecording(&actions.InitiateScreenRecording),
		generateMediaTranscriptions(&actions.MediaTranscriptions),
		generateIntegrationExport(&actions.IntegrationExport),
	)

	return actionsString
}

func generateIntegrationExport(integrationExport *Integrationexport) string {
	if reflect.DeepEqual(*integrationExport, Integrationexport{}) {
		return ""
	}

	integrationExportString := fmt.Sprintf(`
        integration_export {
			integration_id = genesyscloud_integration.%s.id
            should_export_screen_recordings = %v
        }
        `, integrationResource1,
		integrationExport.ShouldExportScreenRecordings,
	)

	return integrationExportString
}

func generateMediaTranscriptions(transcriptions *[]Mediatranscription) string {
	if *transcriptions == nil || len(*transcriptions) <= 0 {
		return ""
	}

	transcriptionsString := ""
	for _, transcription := range *transcriptions {
		transcriptionString := fmt.Sprintf(`
        	media_transcriptions {
        	    display_name = "%s"
        	    transcription_provider = "%s"
        	    integration_id = genesyscloud_integration.%s.id
        	}
        	`, transcription.DisplayName,
			transcription.TranscriptionProvider,
			integrationResource1,
		)

		transcriptionsString += transcriptionString
	}

	return transcriptionsString
}

func generateInitiateScreenRecording(initiateScreenRecording *Initiatescreenrecording) string {
	if reflect.DeepEqual(*initiateScreenRecording, Initiatescreenrecording{}) {
		return ""
	}

	initiateScreenRecordingString := fmt.Sprintf(`
        initiate_screen_recording {
            record_acw = %v
			%s
			%s
        }
        `, initiateScreenRecording.RecordACW,
		generateArchiveRetention(&initiateScreenRecording.ArchiveRetention),
		generateDeleteRetention(&initiateScreenRecording.DeleteRetention),
	)

	return initiateScreenRecordingString
}

func generateRetentionDuration(retentionDuration *Retentionduration) string {
	if reflect.DeepEqual(*retentionDuration, Retentionduration{}) {
		return ""
	}

	retentionDurationString := fmt.Sprintf(`
        retention_duration {
            %s
            %s
        }
        `, generateArchiveRetention(&retentionDuration.ArchiveRetention),
		generateDeleteRetention(&retentionDuration.DeleteRetention),
	)

	return retentionDurationString
}

func generateArchiveRetention(archiveRetention *Archiveretention) string {
	if reflect.DeepEqual(*archiveRetention, Archiveretention{}) {
		return ""
	}

	archiveRetentionString := fmt.Sprintf(`
        archive_retention {
            days = %v
            storage_medium = "%s"
        }
        `, archiveRetention.Days,
		archiveRetention.StorageMedium,
	)

	return archiveRetentionString
}

func generateDeleteRetention(deleteRetention *Deleteretention) string {
	if reflect.DeepEqual(*deleteRetention, Deleteretention{}) {
		return ""
	}

	deleteRetentionString := fmt.Sprintf(`
        delete_retention {
            days = %v
        }
        `, deleteRetention.Days,
	)

	return deleteRetentionString
}

func generateAssignCalibrations(assignments *[]Calibrationassignment) string {
	if *assignments == nil || len(*assignments) <= 0 {
		return ""
	}

	assignmentsString := ""
	for _, assignment := range *assignments {
		evaluatorIdsString := ""
		for i := range assignment.Evaluators {
			if i > 0 {
				evaluatorIdsString += ", "
			}
			evaluatorIdsString += fmt.Sprintf("genesyscloud_user.%s.id", userResource1)
		}

		assignmentString := fmt.Sprintf(`
        	assign_calibrations {
				calibrator_id = genesyscloud_user.%s.id
				evaluator_ids = [%s]
				evaluation_form_id = genesyscloud_quality_forms_evaluation.%s.id
				expert_evaluator_id = genesyscloud_user.%s.id
        	}
        	`, userResource1,
			evaluatorIdsString,
			evaluationFormResource1,
			userResource1,
		)

		assignmentsString += assignmentString
	}

	return assignmentsString
}

func generateAssignMeteredAssignmentByAgent(assignments *[]Meteredassignmentbyagent) string {
	if *assignments == nil || len(*assignments) <= 0 {
		return ""
	}

	assignmentsString := ""
	for _, assignment := range *assignments {
		evaluatorIdsString := ""
		for i := range assignment.Evaluators {
			if i > 0 {
				evaluatorIdsString += ", "
			}
			evaluatorIdsString += fmt.Sprintf("genesyscloud_user.%s.id", userResource1)
		}
		assignmentString := fmt.Sprintf(`
        	assign_metered_assignment_by_agent {
        	    evaluator_ids = [%s]
				max_number_evaluations = %v
				evaluation_form_id = genesyscloud_quality_forms_evaluation.%s.id
				%s
				time_zone = "%s"
        	}
        	`, evaluatorIdsString,
			assignment.MaxNumberEvaluations,
			evaluationFormResource1,
			generateAgentTimeInterval(&assignment.TimeInterval),
			assignment.TimeZone,
		)

		assignmentsString += assignmentString
	}

	return assignmentsString
}

func generateAssignMeteredEvaluations(assignments *[]Meteredevaluationassignment) string {
	if *assignments == nil || len(*assignments) <= 0 {
		return ""
	}

	assignmentsString := ""
	for _, assignment := range *assignments {
		evaluatorIdsString := ""
		for i := range assignment.Evaluators {
			if i > 0 {
				evaluatorIdsString += ", "
			}
			evaluatorIdsString += fmt.Sprintf("genesyscloud_user.%s.id", userResource1)
		}
		assignmentString := fmt.Sprintf(`
        	assign_metered_evaluations {
        	    evaluator_ids = [%s]
				max_number_evaluations = %v
				evaluation_form_id = genesyscloud_quality_forms_evaluation.%s.id
				assign_to_active_user = %v
				%s
        	}
        	`, evaluatorIdsString,
			assignment.MaxNumberEvaluations,
			evaluationFormResource1,
			assignment.AssignToActiveUser,
			generateEvalTimeInterval(&assignment.TimeInterval),
		)

		assignmentsString += assignmentString
	}

	return assignmentsString
}

func generateAssignEvaluations() string {
	assignmentString := fmt.Sprintf(`
    	assign_evaluations {
			evaluation_form_id = genesyscloud_quality_forms_evaluation.%s.id
    	    user_id = genesyscloud_user.%s.id
    	}
    	`, evaluationFormResource1,
		userResource1,
	)
	return assignmentString
}

func generateAgentTimeInterval(timeInterval *AgentTimeinterval) string {
	if reflect.DeepEqual(*timeInterval, AgentTimeinterval{}) {
		return ""
	}

	timeIntervalString := fmt.Sprintf(`
        time_interval {
            months = %v
            weeks = %v
            days = %v
        }
        `, timeInterval.Months,
		timeInterval.Weeks,
		timeInterval.Days,
	)

	return timeIntervalString
}

func generateEvalTimeInterval(timeInterval *EvalTimeinterval) string {
	if reflect.DeepEqual(*timeInterval, EvalTimeinterval{}) {
		return ""
	}

	timeIntervalString := fmt.Sprintf(`
        time_interval {
            days = %v
            hours = %v
        }
        `,
		timeInterval.Days,
		timeInterval.Hours,
	)

	return timeIntervalString
}

func generateCallMediaPolicyConditions(conditions *Callmediapolicyconditions) string {
	if reflect.DeepEqual(*conditions, Callmediapolicyconditions{}) {
		return ""
	}

	dateRangesString := ""
	for i, dateRange := range conditions.DateRanges {
		if i > 0 {
			dateRangesString += ", "
		}

		dateRangesString += strconv.Quote(dateRange)
	}

	directionsString := ""
	for i, directions := range conditions.Directions {
		if i > 0 {
			directionsString += ", "
		}

		directionsString += strconv.Quote(directions)
	}

	userIdsString := ""
	for i := range conditions.ForUsers {
		if i > 0 {
			userIdsString += ", "
		}
		userIdsString += fmt.Sprintf("genesyscloud_user.%s.id", userResource1)
	}

	queueIdsString := ""
	for i := range conditions.ForQueues {
		if i > 0 {
			queueIdsString += ", "
		}
		queueIdsString += fmt.Sprintf("genesyscloud_routing_queue.%s.id", queueResource1)
	}

	languageIdsString := ""
	for i := range conditions.Languages {
		if i > 0 {
			languageIdsString += ", "
		}
		languageIdsString += fmt.Sprintf("genesyscloud_routing_language.%s.id", languageResource1)
	}

	wrapupCodeIdsString := ""
	for i := range conditions.WrapupCodes {
		if i > 0 {
			wrapupCodeIdsString += ", "
		}
		wrapupCodeIdsString += fmt.Sprintf("genesyscloud_routing_wrapupcode.%s.id", wrapupCodeResource1)
	}

	conditionsString := fmt.Sprintf(`
        conditions {
			for_user_ids = [%s]
			date_ranges = [%s]
            for_queue_ids = [%s]
			wrapup_code_ids = [%s]
			language_ids=[%s]
			%s
			directions = [%s]
			%s
        }
        `, userIdsString,
		dateRangesString,
		queueIdsString,
		wrapupCodeIdsString,
		languageIdsString,
		generateTimeAllowed(&conditions.TimeAllowed),
		directionsString,
		generateDuration(&conditions.Duration),
	)

	return conditionsString
}

func generateAssignSurveys(assignSurveys *[]Surveyassignment) string {
	if *assignSurveys == nil || len(*assignSurveys) <= 0 {
		return ""
	}

	assignSurveysString := ""

	for _, assignSurvey := range *assignSurveys {
		assignSurveyString := fmt.Sprintf(`
        assign_surveys {
            sending_domain = %s
            survey_form_name = "%s"
            flow_id = genesyscloud_flow.%s.id
        }
        `, assignSurvey.SendingDomain,
			surveyFormName,
			flowResource1,
		)

		assignSurveysString += assignSurveyString
	}

	return assignSurveysString
}

func CleanupRoutingEmailDomains() {
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		routingEmailDomains, _, getErr := routingAPI.GetRoutingEmailDomains(pageSize, pageNum, false, "")
		if getErr != nil {
			log.Printf("failed to get page %v of routing email domains: %v", pageNum, getErr)
			return
		}

		if routingEmailDomains.Entities == nil || len(*routingEmailDomains.Entities) == 0 {
			break
		}

		for _, routingEmailDomain := range *routingEmailDomains.Entities {
			if routingEmailDomain.Name != nil && strings.HasPrefix(*routingEmailDomain.Name, "terraformmedia") {
				_, err := routingAPI.DeleteRoutingEmailDomain(*routingEmailDomain.Id)
				if err != nil {
					log.Printf("Failed to delete routing email domain %s: %s", *routingEmailDomain.Id, err)
					return
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func GenerateResourceRoles(skillID string, divisionIds ...string) string {
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

// TODO Duplicating this code within the function to not break a cyclid dependency
func generateUserWithCustomAttrs(resourceID string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceID, email, name, strings.Join(attrs, "\n"))
}
