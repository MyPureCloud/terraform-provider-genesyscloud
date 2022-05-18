package genesyscloud

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

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
	EvaluationContextId  string
	Evaluators           []User
	MaxNumberEvaluations int
	EvaluationForm       Evaluationform
	AssignToActiveUser   bool
	TimeInterval         Timeinterval
}

type Meteredassignmentbyagent struct {
	EvaluationContextId  string
	Evaluators           []User
	MaxNumberEvaluations int
	EvaluationForm       Evaluationform
	TimeInterval         Timeinterval
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
	Id              string
	Name            string
	Division        Division
	Chat            Chat
	Department      string
	Email           string
	Addresses       []Contact
	Title           string
	Username        string
	Images          []Userimage
	Version         int
	Certifications  []string
	Biography       Biography
	EmployerInfo    Employerinfo
	AcdAutoAnswer   bool
	LastTokenIssued Oauthlasttokenissued
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
	Id                         string
	Name                       string
	Division                   Division
	Description                string
	DateCreated                time.Time
	DateModified               time.Time
	ModifiedBy                 string
	CreatedBy                  string
	MemberCount                int
	UserMemberCount            int
	JoinedMemberCount          int
	MediaSettings              map[string]Mediasetting
	RoutingRules               []Routingrule
	Bullseye                   Bullseye
	AcwSettings                Acwsettings
	SkillEvaluationMethod      string
	QueueFlow                  Domainentityref
	EmailInQueueFlow           Domainentityref
	MessageInQueueFlow         Domainentityref
	WhisperPrompt              Domainentityref
	OnHoldPrompt               Domainentityref
	AutoAnswerOnly             bool
	EnableTranscription        bool
	EnableManualAssignment     bool
	CallingPartyName           string
	CallingPartyNumber         string
	OutboundMessagingAddresses Queuemessagingaddresses
	OutboundEmailAddress       Queueemailaddress
	SelfUri                    string
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

type Timeinterval struct {
	Months int
	Weeks  int
	Days   int
	Hours  int
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
	Id           string
	Name         string
	DateCreated  time.Time
	DateModified time.Time
	ModifiedBy   string
	CreatedBy    string
}

type Language struct {
	Id           string
	Name         string
	DateModified time.Time
	State        string
	Version      string
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
	Id            string
	Name          string
	ModifiedDate  time.Time
	CreatedDate   time.Time
	Order         int
	Description   string
	Enabled       bool
	MediaPolicies Mediapolicies
	Conditions    Policyconditions `json:"conditions,omitempty"`
	Actions       Policyactions
	PolicyErrors  Policyerrors
	SelfUri       string
}

func TestAccResourceMediaRetentionPolicyBasic(t *testing.T) {
	policyResource1 := "test-media-retention-policy-1"

	basePolicy := Policycreate{
		Name:        "terraform-media-retention-policy",
		Order:       1,
		Description: "a media retention policy",
		Enabled:     true,
	}

	mediaRetentionChatPolicy := basePolicy
	mediaRetentionChatPolicy.MediaPolicies = Mediapolicies{
		ChatPolicy: Chatmediapolicy{
			Actions: Policyactions{
				RetainRecording: false,
				DeleteRecording: true,
				AlwaysDelete:    false,
				AssignEvaluations: []Evaluationassignment{
					{
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
						User: User{
							Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
							Name: "Jacob Shaw",
							Division: Division{
								Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								Name:    "New Home",
								SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							},
							Chat: Chat{
								JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
							},
							Email:         "jacob.shaw@genesys.com",
							Username:      "jacob.shaw@genesys.com",
							Version:       21,
							AcdAutoAnswer: false,
						},
					},
				},
				AssignMeteredEvaluations: []Meteredevaluationassignment{
					{
						Evaluators: []User{
							{
								Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
								Name: "Jacob Shaw",
								Division: Division{
									Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
									Name:    "New Home",
									SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								},
								Chat: Chat{
									JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
								},
								Email:         "jacob.shaw@genesys.com",
								Username:      "jacob.shaw@genesys.com",
								Version:       21,
								AcdAutoAnswer: false,
							},
						},
						MaxNumberEvaluations: 1,
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
						AssignToActiveUser: true,
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
						Evaluators: []User{
							{
								Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
								Name: "Jacob Shaw",
								Division: Division{
									Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
									Name:    "New Home",
									SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								},
								Chat: Chat{
									JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
								},
								Email:         "jacob.shaw@genesys.com",
								Username:      "jacob.shaw@genesys.com",
								Version:       21,
								AcdAutoAnswer: false,
							},
						},
						MaxNumberEvaluations: 1,
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
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
						Calibrator: User{
							Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
							Name: "Jacob Shaw",
							Division: Division{
								Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								Name:    "New Home",
								SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							},
							Chat: Chat{
								JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
							},
							Email:         "jacob.shaw@genesys.com",
							Username:      "jacob.shaw@genesys.com",
							Version:       21,
							AcdAutoAnswer: false,
						},
						Evaluators: []User{
							{
								Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
								Name: "Jacob Shaw",
								Division: Division{
									Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
									Name:    "New Home",
									SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								},
								Chat: Chat{
									JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
								},
								Email:         "jacob.shaw@genesys.com",
								Username:      "jacob.shaw@genesys.com",
								Version:       21,
								AcdAutoAnswer: false,
							},
						},
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
						ExpertEvaluator: User{
							Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
							Name: "Jacob Shaw",
							Division: Division{
								Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								Name:    "New Home",
								SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							},
							Chat: Chat{
								JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
							},
							Email:         "jacob.shaw@genesys.com",
							Username:      "jacob.shaw@genesys.com",
							Version:       21,
							AcdAutoAnswer: false,
						},
					},
				},
				AssignSurveys: []Surveyassignment{
					{
						SendingDomain: "surveys.mypurecloud.com",
						SurveyForm: Publishedsurveyformreference{
							Name:      "ronan test",
							ContextId: "714b523e-6fc9-44fa-9eff-cd5f6a13430a",
						},
						Flow: Domainentityref{
							Id:      "fa1da2b1-4710-4a4f-9d13-399ec13e4f53",
							Name:    "SendSurvey",
							SelfUri: "/api/v2/flows/fa1da2b1-4710-4a4f-9d13-399ec13e4f53",
						},
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
			},
			Conditions: Chatmediapolicyconditions{
				DateRanges: []string{
					"2022-05-12T04:00:00.000Z/2022-05-13T04:00:00.000Z",
				},
				ForQueues: []Queue{
					{
						Id:   "c3275a86-3a0c-4a06-beaf-ca1bf096b7b5",
						Name: "Transcription Queue",
						Division: Division{
							Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							Name:    "New Home",
							SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
						},
						EnableManualAssignment: true,
						EnableTranscription:    true,
					},
				},
				ForUsers: []User{
					{
						Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
						Name: "Jacob Shaw",
						Division: Division{
							Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							Name:    "New Home",
							SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
						},
						Chat: Chat{
							JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
						},
						Email:         "jacob.shaw@genesys.com",
						Username:      "jacob.shaw@genesys.com",
						Version:       21,
						AcdAutoAnswer: false,
					},
				},
				WrapupCodes: []Wrapupcode{
					{
						Id:   "f62e36fa-8f8b-4d8b-be31-649f959c62ec",
						Name: "Postman 1",
					},
				},
				Languages: []Language{
					{
						Id:      "13005cea-29c2-458d-9247-90502d4154dc",
						Name:    "English - Spoken",
						State:   "active",
						Version: "1",
					},
				},
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
			},
		},
	}

	mediaRetentionCallPolicy := basePolicy
	mediaRetentionCallPolicy.MediaPolicies = Mediapolicies{
		CallPolicy: Callmediapolicy{
			Actions: Policyactions{
				RetainRecording: false,
				DeleteRecording: true,
				AlwaysDelete:    false,
				AssignEvaluations: []Evaluationassignment{
					{
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
						User: User{
							Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
							Name: "Jacob Shaw",
							Division: Division{
								Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								Name:    "New Home",
								SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							},
							Chat: Chat{
								JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
							},
							Email:         "jacob.shaw@genesys.com",
							Username:      "jacob.shaw@genesys.com",
							Version:       21,
							AcdAutoAnswer: false,
						},
					},
				},
				AssignMeteredEvaluations: []Meteredevaluationassignment{
					{
						Evaluators: []User{
							{
								Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
								Name: "Jacob Shaw",
								Division: Division{
									Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
									Name:    "New Home",
									SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								},
								Chat: Chat{
									JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
								},
								Email:         "jacob.shaw@genesys.com",
								Username:      "jacob.shaw@genesys.com",
								Version:       21,
								AcdAutoAnswer: false,
							},
						},
						MaxNumberEvaluations: 1,
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
						AssignToActiveUser: true,
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
						Evaluators: []User{
							{
								Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
								Name: "Jacob Shaw",
								Division: Division{
									Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
									Name:    "New Home",
									SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								},
								Chat: Chat{
									JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
								},
								Email:         "jacob.shaw@genesys.com",
								Username:      "jacob.shaw@genesys.com",
								Version:       21,
								AcdAutoAnswer: false,
							},
						},
						MaxNumberEvaluations: 1,
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
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
						Calibrator: User{
							Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
							Name: "Jacob Shaw",
							Division: Division{
								Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								Name:    "New Home",
								SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							},
							Chat: Chat{
								JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
							},
							Email:         "jacob.shaw@genesys.com",
							Username:      "jacob.shaw@genesys.com",
							Version:       21,
							AcdAutoAnswer: false,
						},
						Evaluators: []User{
							{
								Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
								Name: "Jacob Shaw",
								Division: Division{
									Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
									Name:    "New Home",
									SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								},
								Chat: Chat{
									JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
								},
								Email:         "jacob.shaw@genesys.com",
								Username:      "jacob.shaw@genesys.com",
								Version:       21,
								AcdAutoAnswer: false,
							},
						},
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
						ExpertEvaluator: User{
							Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
							Name: "Jacob Shaw",
							Division: Division{
								Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								Name:    "New Home",
								SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							},
							Chat: Chat{
								JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
							},
							Email:         "jacob.shaw@genesys.com",
							Username:      "jacob.shaw@genesys.com",
							Version:       21,
							AcdAutoAnswer: false,
						},
					},
				},
				AssignSurveys: []Surveyassignment{
					{
						SendingDomain: "surveys.mypurecloud.com",
						SurveyForm: Publishedsurveyformreference{
							Name:      "ronan test",
							ContextId: "714b523e-6fc9-44fa-9eff-cd5f6a13430a",
						},
						Flow: Domainentityref{
							Id:      "fa1da2b1-4710-4a4f-9d13-399ec13e4f53",
							Name:    "SendSurvey",
							SelfUri: "/api/v2/flows/fa1da2b1-4710-4a4f-9d13-399ec13e4f53",
						},
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
			},
			Conditions: Callmediapolicyconditions{
				DateRanges: []string{
					"2022-05-12T04:00:00.000Z/2022-05-13T04:00:00.000Z",
				},
				Directions: []string{
					"INBOUND",
				},
				ForQueues: []Queue{
					{
						Id:   "c3275a86-3a0c-4a06-beaf-ca1bf096b7b5",
						Name: "Transcription Queue",
						Division: Division{
							Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							Name:    "New Home",
							SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
						},
						EnableManualAssignment: true,
						EnableTranscription:    true,
					},
				},
				ForUsers: []User{
					{
						Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
						Name: "Jacob Shaw",
						Division: Division{
							Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							Name:    "New Home",
							SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
						},
						Chat: Chat{
							JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
						},
						Email:         "jacob.shaw@genesys.com",
						Username:      "jacob.shaw@genesys.com",
						Version:       21,
						AcdAutoAnswer: false,
					},
				},
				WrapupCodes: []Wrapupcode{
					{
						Id:   "f62e36fa-8f8b-4d8b-be31-649f959c62ec",
						Name: "Postman 1",
					},
				},
				Languages: []Language{
					{
						Id:      "13005cea-29c2-458d-9247-90502d4154dc",
						Name:    "English - Spoken",
						State:   "active",
						Version: "1",
					},
				},
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
			},
		},
	}

	mediaRetentionMessagePolicy := basePolicy
	mediaRetentionMessagePolicy.MediaPolicies = Mediapolicies{
		MessagePolicy: Messagemediapolicy{
			Actions: Policyactions{
				RetainRecording: false,
				DeleteRecording: true,
				AlwaysDelete:    false,
				AssignEvaluations: []Evaluationassignment{
					{
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
						User: User{
							Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
							Name: "Jacob Shaw",
							Division: Division{
								Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								Name:    "New Home",
								SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							},
							Chat: Chat{
								JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
							},
							Email:         "jacob.shaw@genesys.com",
							Username:      "jacob.shaw@genesys.com",
							Version:       21,
							AcdAutoAnswer: false,
						},
					},
				},
				AssignMeteredEvaluations: []Meteredevaluationassignment{
					{
						Evaluators: []User{
							{
								Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
								Name: "Jacob Shaw",
								Division: Division{
									Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
									Name:    "New Home",
									SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								},
								Chat: Chat{
									JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
								},
								Email:         "jacob.shaw@genesys.com",
								Username:      "jacob.shaw@genesys.com",
								Version:       21,
								AcdAutoAnswer: false,
							},
						},
						MaxNumberEvaluations: 1,
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
						AssignToActiveUser: true,
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
						Evaluators: []User{
							{
								Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
								Name: "Jacob Shaw",
								Division: Division{
									Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
									Name:    "New Home",
									SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								},
								Chat: Chat{
									JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
								},
								Email:         "jacob.shaw@genesys.com",
								Username:      "jacob.shaw@genesys.com",
								Version:       21,
								AcdAutoAnswer: false,
							},
						},
						MaxNumberEvaluations: 1,
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
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
						Calibrator: User{
							Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
							Name: "Jacob Shaw",
							Division: Division{
								Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								Name:    "New Home",
								SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							},
							Chat: Chat{
								JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
							},
							Email:         "jacob.shaw@genesys.com",
							Username:      "jacob.shaw@genesys.com",
							Version:       21,
							AcdAutoAnswer: false,
						},
						Evaluators: []User{
							{
								Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
								Name: "Jacob Shaw",
								Division: Division{
									Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
									Name:    "New Home",
									SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								},
								Chat: Chat{
									JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
								},
								Email:         "jacob.shaw@genesys.com",
								Username:      "jacob.shaw@genesys.com",
								Version:       21,
								AcdAutoAnswer: false,
							},
						},
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
						ExpertEvaluator: User{
							Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
							Name: "Jacob Shaw",
							Division: Division{
								Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								Name:    "New Home",
								SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							},
							Chat: Chat{
								JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
							},
							Email:         "jacob.shaw@genesys.com",
							Username:      "jacob.shaw@genesys.com",
							Version:       21,
							AcdAutoAnswer: false,
						},
					},
				},
				AssignSurveys: []Surveyassignment{
					{
						SendingDomain: "surveys.mypurecloud.com",
						SurveyForm: Publishedsurveyformreference{
							Name:      "ronan test",
							ContextId: "714b523e-6fc9-44fa-9eff-cd5f6a13430a",
						},
						Flow: Domainentityref{
							Id:      "fa1da2b1-4710-4a4f-9d13-399ec13e4f53",
							Name:    "SendSurvey",
							SelfUri: "/api/v2/flows/fa1da2b1-4710-4a4f-9d13-399ec13e4f53",
						},
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
			},
			Conditions: Messagemediapolicyconditions{
				DateRanges: []string{
					"2022-05-12T04:00:00.000Z/2022-05-13T04:00:00.000Z",
				},
				ForQueues: []Queue{
					{
						Id:   "c3275a86-3a0c-4a06-beaf-ca1bf096b7b5",
						Name: "Transcription Queue",
						Division: Division{
							Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							Name:    "New Home",
							SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
						},
						EnableManualAssignment: true,
						EnableTranscription:    true,
					},
				},
				ForUsers: []User{
					{
						Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
						Name: "Jacob Shaw",
						Division: Division{
							Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							Name:    "New Home",
							SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
						},
						Chat: Chat{
							JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
						},
						Email:         "jacob.shaw@genesys.com",
						Username:      "jacob.shaw@genesys.com",
						Version:       21,
						AcdAutoAnswer: false,
					},
				},
				WrapupCodes: []Wrapupcode{
					{
						Id:   "f62e36fa-8f8b-4d8b-be31-649f959c62ec",
						Name: "Postman 1",
					},
				},
				Languages: []Language{
					{
						Id:      "13005cea-29c2-458d-9247-90502d4154dc",
						Name:    "English - Spoken",
						State:   "active",
						Version: "1",
					},
				},
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
			},
		},
	}

	mediaRetentionEmailPolicy := basePolicy
	mediaRetentionEmailPolicy.MediaPolicies = Mediapolicies{
		EmailPolicy: Emailmediapolicy{
			Actions: Policyactions{
				RetainRecording: false,
				DeleteRecording: true,
				AlwaysDelete:    false,
				AssignEvaluations: []Evaluationassignment{
					{
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
						User: User{
							Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
							Name: "Jacob Shaw",
							Division: Division{
								Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								Name:    "New Home",
								SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							},
							Chat: Chat{
								JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
							},
							Email:         "jacob.shaw@genesys.com",
							Username:      "jacob.shaw@genesys.com",
							Version:       21,
							AcdAutoAnswer: false,
						},
					},
				},
				AssignMeteredEvaluations: []Meteredevaluationassignment{
					{
						Evaluators: []User{
							{
								Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
								Name: "Jacob Shaw",
								Division: Division{
									Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
									Name:    "New Home",
									SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								},
								Chat: Chat{
									JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
								},
								Email:         "jacob.shaw@genesys.com",
								Username:      "jacob.shaw@genesys.com",
								Version:       21,
								AcdAutoAnswer: false,
							},
						},
						MaxNumberEvaluations: 1,
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
						AssignToActiveUser: true,
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
						Evaluators: []User{
							{
								Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
								Name: "Jacob Shaw",
								Division: Division{
									Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
									Name:    "New Home",
									SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								},
								Chat: Chat{
									JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
								},
								Email:         "jacob.shaw@genesys.com",
								Username:      "jacob.shaw@genesys.com",
								Version:       21,
								AcdAutoAnswer: false,
							},
						},
						MaxNumberEvaluations: 1,
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
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
						Calibrator: User{
							Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
							Name: "Jacob Shaw",
							Division: Division{
								Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								Name:    "New Home",
								SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							},
							Chat: Chat{
								JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
							},
							Email:         "jacob.shaw@genesys.com",
							Username:      "jacob.shaw@genesys.com",
							Version:       21,
							AcdAutoAnswer: false,
						},
						Evaluators: []User{
							{
								Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
								Name: "Jacob Shaw",
								Division: Division{
									Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
									Name:    "New Home",
									SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								},
								Chat: Chat{
									JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
								},
								Email:         "jacob.shaw@genesys.com",
								Username:      "jacob.shaw@genesys.com",
								Version:       21,
								AcdAutoAnswer: false,
							},
						},
						EvaluationForm: Evaluationform{
							QuestionGroups: []Evaluationquestiongroup{
								{
									Name:                    "Test group",
									DefaultAnswersToHighest: true,
									DefaultAnswersToNA:      false,
									NaEnabled:               true,
									Weight:                  100,
									ManualWeight:            true,
								},
							},
							Name:      "Jacob test",
							Published: true,
							ContextId: "92ca7b82-04d6-4e3c-8bdd-6c1e143e22fd",
						},
						ExpertEvaluator: User{
							Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
							Name: "Jacob Shaw",
							Division: Division{
								Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
								Name:    "New Home",
								SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							},
							Chat: Chat{
								JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
							},
							Email:         "jacob.shaw@genesys.com",
							Username:      "jacob.shaw@genesys.com",
							Version:       21,
							AcdAutoAnswer: false,
						},
					},
				},
				AssignSurveys: []Surveyassignment{
					{
						SendingDomain: "surveys.mypurecloud.com",
						SurveyForm: Publishedsurveyformreference{
							Name:      "ronan test",
							ContextId: "714b523e-6fc9-44fa-9eff-cd5f6a13430a",
						},
						Flow: Domainentityref{
							Id:      "fa1da2b1-4710-4a4f-9d13-399ec13e4f53",
							Name:    "SendSurvey",
							SelfUri: "/api/v2/flows/fa1da2b1-4710-4a4f-9d13-399ec13e4f53",
						},
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
			},
			Conditions: Emailmediapolicyconditions{
				DateRanges: []string{
					"2022-05-12T04:00:00.000Z/2022-05-13T04:00:00.000Z",
				},
				ForQueues: []Queue{
					{
						Id:   "c3275a86-3a0c-4a06-beaf-ca1bf096b7b5",
						Name: "Transcription Queue",
						Division: Division{
							Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							Name:    "New Home",
							SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
						},
						EnableManualAssignment: true,
						EnableTranscription:    true,
					},
				},
				ForUsers: []User{
					{
						Id:   "fec34eaa-818e-4617-b317-9360a00089fc",
						Name: "Jacob Shaw",
						Division: Division{
							Id:      "4a7e4f4b-e734-40b6-838e-557c5eedd49a",
							Name:    "New Home",
							SelfUri: "/api/v2/authorization/divisions/4a7e4f4b-e734-40b6-838e-557c5eedd49a",
						},
						Chat: Chat{
							JabberId: "609ad2b4c1464a1b08dc70b3@hollywoo.orgspan.com",
						},
						Email:         "jacob.shaw@genesys.com",
						Username:      "jacob.shaw@genesys.com",
						Version:       21,
						AcdAutoAnswer: false,
					},
				},
				WrapupCodes: []Wrapupcode{
					{
						Id:   "f62e36fa-8f8b-4d8b-be31-649f959c62ec",
						Name: "Postman 1",
					},
				},
				Languages: []Language{
					{
						Id:      "13005cea-29c2-458d-9247-90502d4154dc",
						Name:    "English - Spoken",
						State:   "active",
						Version: "1",
					},
				},
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
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateMediaRetentionPolicyResource(policyResource1, &mediaRetentionCallPolicy),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "name", mediaRetentionCallPolicy.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "description", mediaRetentionCallPolicy.Description),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "enabled", trueValue),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.retain_recording", strconv.FormatBool(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.RetainRecording)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.delete_recording", strconv.FormatBool(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.DeleteRecording)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.always_delete", strconv.FormatBool(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AlwaysDelete)),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_evaluations.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignEvaluations))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_evaluations.0.evaluation_form.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignEvaluations[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_evaluations.0.evaluation_form.0.question_groups.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignEvaluations[0].EvaluationForm.QuestionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_evaluations.0.evaluation_form.0.question_groups.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignEvaluations[0].EvaluationForm.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_evaluations.0.user.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignEvaluations[0].User.Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredEvaluations))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.0.max_number_evaluations", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredEvaluations[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.0.evaluators.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredEvaluations[0].Evaluators))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.0.evaluators.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredEvaluations[0].Evaluators[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.0.evaluation_form.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredEvaluations[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.0.time_interval.0.days", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredEvaluations[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredEvaluations))),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredAssignmentByAgent))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.0.max_number_evaluations", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredAssignmentByAgent[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluators.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredAssignmentByAgent[0].Evaluators))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluators.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredAssignmentByAgent[0].Evaluators[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluation_form.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredAssignmentByAgent[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.0.time_interval.0.days", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredAssignmentByAgent[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignMeteredAssignmentByAgent))),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_calibrations.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignCalibrations))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_calibrations.0.calibrator.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignCalibrations[0].Calibrator.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_calibrations.0.evaluators.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignCalibrations[0].Evaluators))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_calibrations.0.evaluators.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignCalibrations[0].Evaluators[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_calibrations.0.evaluation_form.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignCalibrations[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_calibrations.0.expert_evaluator.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignCalibrations[0].ExpertEvaluator.Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_surveys.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignSurveys))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_surveys.0.survey_form.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignSurveys[0].SurveyForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_surveys.0.survey_form.0.context_id", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignSurveys[0].SurveyForm.ContextId),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_surveys.0.flow.0.id", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignSurveys[0].Flow.Id),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_surveys.0.flow.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignSurveys[0].Flow.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.assign_surveys.0.flow.0.self_uri", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.AssignSurveys[0].Flow.SelfUri),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.retention_duration.0.archive_retention.0.storage_medium", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.RetentionDuration.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.retention_duration.0.delete_retention.0.days", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.RetentionDuration.DeleteRetention.Days)),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.initiate_screen_recording.0.archive_retention.0.storage_medium", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.InitiateScreenRecording.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.initiate_screen_recording.0.delete_retention.0.days", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.InitiateScreenRecording.DeleteRetention.Days)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.actions.0.initiate_screen_recording.0.record_acw", strconv.FormatBool(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Actions.InitiateScreenRecording.RecordACW)),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.for_queues.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.ForQueues))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.for_queues.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.ForQueues[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.for_queues.0.division.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.ForQueues[0].Division.Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.for_users.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.ForUsers))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.for_users.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.ForUsers[0].Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.wrapup_codes.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.WrapupCodes))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.wrapup_codes.0.name", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.WrapupCodes[0].Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.languages.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.Languages))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.languages.0.id", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.Languages[0].Id),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.time_allowed.0.time_zone_id", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.TimeAllowed.TimeZoneId),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.time_allowed.0.time_slots.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.TimeAllowed.TimeSlots))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.time_allowed.0.time_slots.0.start_time", mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.TimeAllowed.TimeSlots[0].StartTime),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.date_ranges.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.DateRanges))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.date_ranges.0", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.DateRanges[0])),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.directions.#", fmt.Sprint(len(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.Directions))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.call_policy.0.conditions.0.directions.0", fmt.Sprint(mediaRetentionCallPolicy.MediaPolicies.CallPolicy.Conditions.Directions[0])),
				),
			},
			{

				Config: generateMediaRetentionPolicyResource(policyResource1, &mediaRetentionChatPolicy),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "name", mediaRetentionCallPolicy.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "description", mediaRetentionCallPolicy.Description),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "enabled", trueValue),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.retain_recording", strconv.FormatBool(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.RetainRecording)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.delete_recording", strconv.FormatBool(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.DeleteRecording)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.always_delete", strconv.FormatBool(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AlwaysDelete)),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_evaluations.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignEvaluations))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_evaluations.0.evaluation_form.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignEvaluations[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_evaluations.0.evaluation_form.0.question_groups.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignEvaluations[0].EvaluationForm.QuestionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_evaluations.0.evaluation_form.0.question_groups.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignEvaluations[0].EvaluationForm.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_evaluations.0.user.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignEvaluations[0].User.Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredEvaluations))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.0.max_number_evaluations", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredEvaluations[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.0.evaluators.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredEvaluations[0].Evaluators))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.0.evaluators.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredEvaluations[0].Evaluators[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.0.evaluation_form.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredEvaluations[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.0.time_interval.0.days", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredEvaluations[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredEvaluations))),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredAssignmentByAgent))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.0.max_number_evaluations", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredAssignmentByAgent[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluators.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredAssignmentByAgent[0].Evaluators))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluators.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredAssignmentByAgent[0].Evaluators[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluation_form.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredAssignmentByAgent[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.0.time_interval.0.days", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredAssignmentByAgent[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignMeteredAssignmentByAgent))),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_calibrations.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignCalibrations))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_calibrations.0.calibrator.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignCalibrations[0].Calibrator.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_calibrations.0.evaluators.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignCalibrations[0].Evaluators))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_calibrations.0.evaluators.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignCalibrations[0].Evaluators[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_calibrations.0.evaluation_form.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignCalibrations[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_calibrations.0.expert_evaluator.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignCalibrations[0].ExpertEvaluator.Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_surveys.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignSurveys))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_surveys.0.survey_form.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignSurveys[0].SurveyForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_surveys.0.survey_form.0.context_id", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignSurveys[0].SurveyForm.ContextId),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_surveys.0.flow.0.id", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignSurveys[0].Flow.Id),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_surveys.0.flow.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignSurveys[0].Flow.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.assign_surveys.0.flow.0.self_uri", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.AssignSurveys[0].Flow.SelfUri),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.retention_duration.0.archive_retention.0.storage_medium", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.RetentionDuration.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.retention_duration.0.delete_retention.0.days", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.RetentionDuration.DeleteRetention.Days)),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.initiate_screen_recording.0.archive_retention.0.storage_medium", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.InitiateScreenRecording.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.initiate_screen_recording.0.delete_retention.0.days", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.InitiateScreenRecording.DeleteRetention.Days)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.actions.0.initiate_screen_recording.0.record_acw", strconv.FormatBool(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Actions.InitiateScreenRecording.RecordACW)),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.for_queues.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.ForQueues))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.for_queues.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.ForQueues[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.for_queues.0.division.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.ForQueues[0].Division.Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.for_users.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.ForUsers))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.for_users.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.ForUsers[0].Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.wrapup_codes.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.WrapupCodes))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.wrapup_codes.0.name", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.WrapupCodes[0].Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.languages.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.Languages))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.languages.0.id", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.Languages[0].Id),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.time_allowed.0.time_zone_id", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.TimeAllowed.TimeZoneId),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.time_allowed.0.time_slots.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.TimeAllowed.TimeSlots))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.time_allowed.0.time_slots.0.start_time", mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.TimeAllowed.TimeSlots[0].StartTime),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.date_ranges.#", fmt.Sprint(len(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.DateRanges))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.chat_policy.0.conditions.0.date_ranges.0", fmt.Sprint(mediaRetentionChatPolicy.MediaPolicies.ChatPolicy.Conditions.DateRanges[0])),
				),
			},
			{

				Config: generateMediaRetentionPolicyResource(policyResource1, &mediaRetentionMessagePolicy),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "name", mediaRetentionCallPolicy.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "description", mediaRetentionCallPolicy.Description),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "enabled", trueValue),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.retain_recording", strconv.FormatBool(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.RetainRecording)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.delete_recording", strconv.FormatBool(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.DeleteRecording)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.always_delete", strconv.FormatBool(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AlwaysDelete)),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_evaluations.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignEvaluations))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_evaluations.0.evaluation_form.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignEvaluations[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_evaluations.0.evaluation_form.0.question_groups.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignEvaluations[0].EvaluationForm.QuestionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_evaluations.0.evaluation_form.0.question_groups.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignEvaluations[0].EvaluationForm.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_evaluations.0.user.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignEvaluations[0].User.Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredEvaluations))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.0.max_number_evaluations", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredEvaluations[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.0.evaluators.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredEvaluations[0].Evaluators))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.0.evaluators.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredEvaluations[0].Evaluators[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.0.evaluation_form.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredEvaluations[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.0.time_interval.0.days", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredEvaluations[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredEvaluations))),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredAssignmentByAgent))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.0.max_number_evaluations", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredAssignmentByAgent[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluators.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredAssignmentByAgent[0].Evaluators))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluators.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredAssignmentByAgent[0].Evaluators[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluation_form.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredAssignmentByAgent[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.0.time_interval.0.days", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredAssignmentByAgent[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignMeteredAssignmentByAgent))),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_calibrations.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignCalibrations))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_calibrations.0.calibrator.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignCalibrations[0].Calibrator.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_calibrations.0.evaluators.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignCalibrations[0].Evaluators))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_calibrations.0.evaluators.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignCalibrations[0].Evaluators[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_calibrations.0.evaluation_form.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignCalibrations[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_calibrations.0.expert_evaluator.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignCalibrations[0].ExpertEvaluator.Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_surveys.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignSurveys))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_surveys.0.survey_form.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignSurveys[0].SurveyForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_surveys.0.survey_form.0.context_id", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignSurveys[0].SurveyForm.ContextId),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_surveys.0.flow.0.id", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignSurveys[0].Flow.Id),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_surveys.0.flow.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignSurveys[0].Flow.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.assign_surveys.0.flow.0.self_uri", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.AssignSurveys[0].Flow.SelfUri),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.retention_duration.0.archive_retention.0.storage_medium", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.RetentionDuration.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.retention_duration.0.delete_retention.0.days", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.RetentionDuration.DeleteRetention.Days)),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.initiate_screen_recording.0.archive_retention.0.storage_medium", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.InitiateScreenRecording.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.initiate_screen_recording.0.delete_retention.0.days", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.InitiateScreenRecording.DeleteRetention.Days)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.actions.0.initiate_screen_recording.0.record_acw", strconv.FormatBool(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Actions.InitiateScreenRecording.RecordACW)),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.for_queues.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.ForQueues))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.for_queues.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.ForQueues[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.for_queues.0.division.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.ForQueues[0].Division.Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.for_users.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.ForUsers))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.for_users.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.ForUsers[0].Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.wrapup_codes.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.WrapupCodes))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.wrapup_codes.0.name", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.WrapupCodes[0].Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.languages.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.Languages))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.languages.0.id", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.Languages[0].Id),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.time_allowed.0.time_zone_id", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.TimeAllowed.TimeZoneId),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.time_allowed.0.time_slots.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.TimeAllowed.TimeSlots))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.time_allowed.0.time_slots.0.start_time", mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.TimeAllowed.TimeSlots[0].StartTime),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.date_ranges.#", fmt.Sprint(len(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.DateRanges))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.message_policy.0.conditions.0.date_ranges.0", fmt.Sprint(mediaRetentionMessagePolicy.MediaPolicies.MessagePolicy.Conditions.DateRanges[0])),
				),
			},
			{

				Config: generateMediaRetentionPolicyResource(policyResource1, &mediaRetentionEmailPolicy),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "name", mediaRetentionCallPolicy.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "description", mediaRetentionCallPolicy.Description),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "enabled", trueValue),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.retain_recording", strconv.FormatBool(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.RetainRecording)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.delete_recording", strconv.FormatBool(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.DeleteRecording)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.always_delete", strconv.FormatBool(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AlwaysDelete)),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_evaluations.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignEvaluations))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_evaluations.0.evaluation_form.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignEvaluations[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_evaluations.0.evaluation_form.0.question_groups.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignEvaluations[0].EvaluationForm.QuestionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_evaluations.0.evaluation_form.0.question_groups.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignEvaluations[0].EvaluationForm.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_evaluations.0.user.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignEvaluations[0].User.Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredEvaluations))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.0.max_number_evaluations", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredEvaluations[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.0.evaluators.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredEvaluations[0].Evaluators))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.0.evaluators.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredEvaluations[0].Evaluators[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.0.evaluation_form.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredEvaluations[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.0.time_interval.0.days", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredEvaluations[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_evaluations.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredEvaluations))),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredAssignmentByAgent))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.0.max_number_evaluations", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredAssignmentByAgent[0].MaxNumberEvaluations)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluators.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredAssignmentByAgent[0].Evaluators))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluators.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredAssignmentByAgent[0].Evaluators[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.0.evaluation_form.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredAssignmentByAgent[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.0.time_interval.0.days", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredAssignmentByAgent[0].TimeInterval.Days)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_metered_assignment_by_agent.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignMeteredAssignmentByAgent))),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_calibrations.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignCalibrations))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_calibrations.0.calibrator.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignCalibrations[0].Calibrator.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_calibrations.0.evaluators.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignCalibrations[0].Evaluators))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_calibrations.0.evaluators.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignCalibrations[0].Evaluators[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_calibrations.0.evaluation_form.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignCalibrations[0].EvaluationForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_calibrations.0.expert_evaluator.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignCalibrations[0].ExpertEvaluator.Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_surveys.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignSurveys))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_surveys.0.survey_form.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignSurveys[0].SurveyForm.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_surveys.0.survey_form.0.context_id", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignSurveys[0].SurveyForm.ContextId),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_surveys.0.flow.0.id", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignSurveys[0].Flow.Id),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_surveys.0.flow.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignSurveys[0].Flow.Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.assign_surveys.0.flow.0.self_uri", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.AssignSurveys[0].Flow.SelfUri),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.retention_duration.0.archive_retention.0.storage_medium", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.RetentionDuration.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.retention_duration.0.delete_retention.0.days", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.RetentionDuration.DeleteRetention.Days)),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.initiate_screen_recording.0.archive_retention.0.storage_medium", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.InitiateScreenRecording.ArchiveRetention.StorageMedium),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.initiate_screen_recording.0.delete_retention.0.days", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.InitiateScreenRecording.DeleteRetention.Days)),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.actions.0.initiate_screen_recording.0.record_acw", strconv.FormatBool(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Actions.InitiateScreenRecording.RecordACW)),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.for_queues.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.ForQueues))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.for_queues.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.ForQueues[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.for_queues.0.division.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.ForQueues[0].Division.Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.for_users.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.ForUsers))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.for_users.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.ForUsers[0].Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.wrapup_codes.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.WrapupCodes))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.wrapup_codes.0.name", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.WrapupCodes[0].Name),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.languages.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.Languages))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.languages.0.id", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.Languages[0].Id),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.time_allowed.0.time_zone_id", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.TimeAllowed.TimeZoneId),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.time_allowed.0.time_slots.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.TimeAllowed.TimeSlots))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.time_allowed.0.time_slots.0.start_time", mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.TimeAllowed.TimeSlots[0].StartTime),

					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.date_ranges.#", fmt.Sprint(len(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.DateRanges))),
					resource.TestCheckResourceAttr("genesyscloud_media_retention_policy."+policyResource1, "media_policies.0.email_policy.0.conditions.0.date_ranges.0", fmt.Sprint(mediaRetentionEmailPolicy.MediaPolicies.EmailPolicy.Conditions.DateRanges[0])),
				),
			},
			{

				ResourceName:            "genesyscloud_media_retention_policy." + policyResource1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"order", "created_date"},
			},
		},
		CheckDestroy: testVerifyEvaluationFormDestroyed,
	})
}

func generateMediaRetentionPolicyResource(resourceID string, mediaRetentionPolicy *Policycreate) string {
	policy := fmt.Sprintf(`resource "genesyscloud_media_retention_policy" "%s" {
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
		generateLifeCycle(),
	)
	return policy
}

func generateLifeCycle() string {
	return `
	lifecycle {
		ignore_changes = [
			media_policies[0].call_policy[0].actions[0].assign_metered_evaluations[0].evaluators[0].id,
			media_policies[0].call_policy[0].actions[0].assign_metered_evaluations[0].evaluators[0].division[0].name,
			media_policies[0].call_policy[0].actions[0].assign_evaluations[0].user[0].id,
			media_policies[0].call_policy[0].actions[0].assign_evaluations[0].user[0].division[0].name,
			media_policies[0].call_policy[0].actions[0].assign_evaluations[0].user[0].division[0].name,
			media_policies[0].call_policy[0].actions[0].assign_metered_assignment_by_agent[0].evaluators[0].id,
			media_policies[0].call_policy[0].actions[0].assign_metered_assignment_by_agent[0].evaluators[0].division[0].name,
			media_policies[0].call_policy[0].actions[0].assign_calibrations[0].calibrator[0].id,
			media_policies[0].call_policy[0].actions[0].assign_calibrations[0].calibrator[0].division[0].name,
			media_policies[0].call_policy[0].actions[0].assign_calibrations[0].evaluators[0].id,
			media_policies[0].call_policy[0].actions[0].assign_calibrations[0].evaluators[0].division[0].name,
			media_policies[0].call_policy[0].actions[0].assign_calibrations[0].expert_evaluator[0].id,
			media_policies[0].call_policy[0].actions[0].assign_calibrations[0].expert_evaluator[0].division[0].name,
			media_policies[0].call_policy[0].conditions[0].for_users[0].id,
			media_policies[0].call_policy[0].conditions[0].for_users[0].division[0].name,
			media_policies[0].call_policy[0].conditions[0].languages[0].name,
			media_policies[0].call_policy[0].conditions[0].languages[0].state,
			media_policies[0].call_policy[0].conditions[0].languages[0].version,
			media_policies[0].chat_policy[0].actions[0].assign_metered_evaluations[0].evaluators[0].id,
			media_policies[0].chat_policy[0].actions[0].assign_metered_evaluations[0].evaluators[0].division[0].name,
			media_policies[0].chat_policy[0].actions[0].assign_evaluations[0].user[0].id,
			media_policies[0].chat_policy[0].actions[0].assign_evaluations[0].user[0].division[0].name,
			media_policies[0].chat_policy[0].actions[0].assign_evaluations[0].user[0].division[0].name,
			media_policies[0].chat_policy[0].actions[0].assign_metered_assignment_by_agent[0].evaluators[0].id,
			media_policies[0].chat_policy[0].actions[0].assign_metered_assignment_by_agent[0].evaluators[0].division[0].name,
			media_policies[0].chat_policy[0].actions[0].assign_calibrations[0].calibrator[0].id,
			media_policies[0].chat_policy[0].actions[0].assign_calibrations[0].calibrator[0].division[0].name,
			media_policies[0].chat_policy[0].actions[0].assign_calibrations[0].evaluators[0].id,
			media_policies[0].chat_policy[0].actions[0].assign_calibrations[0].evaluators[0].division[0].name,
			media_policies[0].chat_policy[0].actions[0].assign_calibrations[0].expert_evaluator[0].id,
			media_policies[0].chat_policy[0].actions[0].assign_calibrations[0].expert_evaluator[0].division[0].name,
			media_policies[0].chat_policy[0].conditions[0].for_users[0].id,
			media_policies[0].chat_policy[0].conditions[0].for_users[0].division[0].name,
			media_policies[0].chat_policy[0].conditions[0].languages[0].name,
			media_policies[0].chat_policy[0].conditions[0].languages[0].state,
			media_policies[0].chat_policy[0].conditions[0].languages[0].version,
			media_policies[0].message_policy[0].actions[0].assign_metered_evaluations[0].evaluators[0].id,
			media_policies[0].message_policy[0].actions[0].assign_metered_evaluations[0].evaluators[0].division[0].name,
			media_policies[0].message_policy[0].actions[0].assign_evaluations[0].user[0].id,
			media_policies[0].message_policy[0].actions[0].assign_evaluations[0].user[0].division[0].name,
			media_policies[0].message_policy[0].actions[0].assign_evaluations[0].user[0].division[0].name,
			media_policies[0].message_policy[0].actions[0].assign_metered_assignment_by_agent[0].evaluators[0].id,
			media_policies[0].message_policy[0].actions[0].assign_metered_assignment_by_agent[0].evaluators[0].division[0].name,
			media_policies[0].message_policy[0].actions[0].assign_calibrations[0].calibrator[0].id,
			media_policies[0].message_policy[0].actions[0].assign_calibrations[0].calibrator[0].division[0].name,
			media_policies[0].message_policy[0].actions[0].assign_calibrations[0].evaluators[0].id,
			media_policies[0].message_policy[0].actions[0].assign_calibrations[0].evaluators[0].division[0].name,
			media_policies[0].message_policy[0].actions[0].assign_calibrations[0].expert_evaluator[0].id,
			media_policies[0].message_policy[0].actions[0].assign_calibrations[0].expert_evaluator[0].division[0].name,
			media_policies[0].message_policy[0].conditions[0].for_users[0].id,
			media_policies[0].message_policy[0].conditions[0].for_users[0].division[0].name,
			media_policies[0].message_policy[0].conditions[0].languages[0].name,
			media_policies[0].message_policy[0].conditions[0].languages[0].state,
			media_policies[0].message_policy[0].conditions[0].languages[0].version,
			media_policies[0].email_policy[0].actions[0].assign_metered_evaluations[0].evaluators[0].id,
			media_policies[0].email_policy[0].actions[0].assign_metered_evaluations[0].evaluators[0].division[0].name,
			media_policies[0].email_policy[0].actions[0].assign_evaluations[0].user[0].id,
			media_policies[0].email_policy[0].actions[0].assign_evaluations[0].user[0].division[0].name,
			media_policies[0].email_policy[0].actions[0].assign_evaluations[0].user[0].division[0].name,
			media_policies[0].email_policy[0].actions[0].assign_metered_assignment_by_agent[0].evaluators[0].id,
			media_policies[0].email_policy[0].actions[0].assign_metered_assignment_by_agent[0].evaluators[0].division[0].name,
			media_policies[0].email_policy[0].actions[0].assign_calibrations[0].calibrator[0].id,
			media_policies[0].email_policy[0].actions[0].assign_calibrations[0].calibrator[0].division[0].name,
			media_policies[0].email_policy[0].actions[0].assign_calibrations[0].evaluators[0].id,
			media_policies[0].email_policy[0].actions[0].assign_calibrations[0].evaluators[0].division[0].name,
			media_policies[0].email_policy[0].actions[0].assign_calibrations[0].expert_evaluator[0].id,
			media_policies[0].email_policy[0].actions[0].assign_calibrations[0].expert_evaluator[0].division[0].name,
			media_policies[0].email_policy[0].conditions[0].for_users[0].id,
			media_policies[0].email_policy[0].conditions[0].for_users[0].division[0].name,
			media_policies[0].email_policy[0].conditions[0].languages[0].name,
			media_policies[0].email_policy[0].conditions[0].languages[0].state,
			media_policies[0].email_policy[0].conditions[0].languages[0].version,
			conditions[0],
			conditions[1]
		]
	}
	`
}

// media_policies[0].call_policy[0].actions[0].assign_calibrations[1].calibrator[0].id,
// media_policies[0].call_policy[0].actions[0].assign_calibrations[1].calibrator[0].division[0].name,
// media_policies[0].call_policy[0].actions[0].assign_calibrations[1].evaluators[0].id,
// media_policies[0].call_policy[0].actions[0].assign_calibrations[1].evaluators[0].division[0].name,
// media_policies[0].call_policy[0].actions[0].assign_calibrations[1].expert_evaluator[0].id,
// media_policies[0].call_policy[0].actions[0].assign_calibrations[1].expert_evaluator[0].division[0].name,

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

	conditionsString := fmt.Sprintf(`
        conditions {
			%s
			directions = [%s]
			date_ranges = [%s]
			media_types = [%s]
            %s
			%s
			%s
			%s
        }
        `, generateUsers("for_users", &conditions.ForUsers),
		directionsString,
		dateRangesString,
		mediaTypesString,
		generateQueues(&conditions.ForQueues),
		generateDuration(&conditions.Duration),
		generateWrapupCodes(&conditions.WrapupCodes),
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

func generateWrapupCodes(codes *[]Wrapupcode) string {
	if codes == nil || len(*codes) <= 0 {
		return ""
	}

	codesString := ""
	for _, code := range *codes {
		codeString := fmt.Sprintf(`
        	wrapup_codes {
        	    id = "%s"
        	    name = "%s"
        	    modified_by = "%s"
        	    created_by = "%s"
        	}
        	`, code.Id,
			code.Name,
			code.ModifiedBy,
			code.CreatedBy,
		)

		codesString += codeString
	}

	return codesString
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

	conditionsString := fmt.Sprintf(`
        conditions {
			%s
			date_ranges = [%s]
            %s
			%s
			%s
			%s
        }
        `, generateUsers("for_users", &conditions.ForUsers),
		dateRangesString,
		generateQueues(&conditions.ForQueues),
		generateWrapupCodes(&conditions.WrapupCodes),
		generateLanguages(&conditions.Languages),
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

	conditionsString := fmt.Sprintf(`
        conditions {
			%s
			date_ranges = [%s]
            %s
			%s
			%s
			%s
        }
        `, generateUsers("for_users", &conditions.ForUsers),
		dateRangesString,
		generateQueues(&conditions.ForQueues),
		generateWrapupCodes(&conditions.WrapupCodes),
		generateLanguages(&conditions.Languages),
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

	conditionsString := fmt.Sprintf(`
        conditions {
			%s
			date_ranges = [%s]
            %s
			%s
			%s
			%s
			%s
        }
        `, generateUsers("for_users", &conditions.ForUsers),
		dateRangesString,
		generateQueues(&conditions.ForQueues),
		generateWrapupCodes(&conditions.WrapupCodes),
		generateLanguages(&conditions.Languages),
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
		generateAssignEvaluations(&actions.AssignEvaluations),
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
			%s
            should_export_screen_recordings = %v
        }
        `, generateDomainEntityRef("integration", &integrationExport.Integration),
		integrationExport.ShouldExportScreenRecordings,
	)

	return integrationExportString
}

func generateDomainEntityRef(propertyName string, ref *Domainentityref) string {
	if reflect.DeepEqual(*ref, Domainentityref{}) {
		return ""
	}

	domainEntityRefString := fmt.Sprintf(`
        %s {
			id = "%s""
			name = "%s"
			self_uri = "%s"
        }
        `, propertyName,
		ref.Id,
		ref.Name,
		ref.SelfUri,
	)

	return domainEntityRefString
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
        	    integration_id = "%s"
        	}
        	`, transcription.DisplayName,
			transcription.TranscriptionProvider,
			transcription.IntegrationId,
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
		assignmentString := fmt.Sprintf(`
        	assign_calibrations {
				%s
				%s
				%s
				%s
        	}
        	`, generateUser("calibrator", &assignment.Calibrator),
			generateUsers("evaluators", &assignment.Evaluators),
			generateEvaluationForm(&assignment.EvaluationForm),
			generateUser("expert_evaluator", &assignment.ExpertEvaluator),
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
		assignmentString := fmt.Sprintf(`
        	assign_metered_assignment_by_agent {
				evaluation_context_id = "%s"
        	    %s
				max_number_evaluations = %v
        	    %s
				%s
				time_zone = "%s"
        	}
        	`, assignment.EvaluationContextId,
			generateUsers("evaluators", &assignment.Evaluators),
			assignment.MaxNumberEvaluations,
			generateEvaluationForm(&assignment.EvaluationForm),
			generateTimeInterval(&assignment.TimeInterval),
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
		assignmentString := fmt.Sprintf(`
        	assign_metered_evaluations {
				evaluation_context_id = "%s"
        	    %s
				max_number_evaluations = %v
        	    %s
				assign_to_active_user = %v
				%s
        	}
        	`, assignment.EvaluationContextId,
			generateUsers("evaluators", &assignment.Evaluators),
			assignment.MaxNumberEvaluations,
			generateEvaluationForm(&assignment.EvaluationForm),
			assignment.AssignToActiveUser,
			generateTimeInterval(&assignment.TimeInterval),
		)

		assignmentsString += assignmentString
	}

	return assignmentsString
}

func generateAssignEvaluations(assignments *[]Evaluationassignment) string {
	if *assignments == nil || len(*assignments) <= 0 {
		return ""
	}

	assignmentsString := ""
	for _, assignment := range *assignments {
		assignmentString := fmt.Sprintf(`
        	assign_evaluations {
        	    %s
        	    %s
        	}
        	`, generateEvaluationForm(&assignment.EvaluationForm),
			generateUser("user", &assignment.User),
		)

		assignmentsString += assignmentString
	}

	return assignmentsString
}

func generateTimeInterval(timeInterval *Timeinterval) string {
	if reflect.DeepEqual(*timeInterval, Timeinterval{}) {
		return ""
	}

	timeIntervalString := fmt.Sprintf(`
        time_interval {
            months = %v
            weeks = %v
            days = %v
            hours = %v
        }
        `, timeInterval.Months,
		timeInterval.Weeks,
		timeInterval.Days,
		timeInterval.Hours,
	)

	return timeIntervalString
}

func generateEvaluationForm(evaluationForm *Evaluationform) string {
	if reflect.DeepEqual(*evaluationForm, Evaluationform{}) {
		return ""
	}

	evaluationFormString := fmt.Sprintf(`
        evaluation_form {
            id = "%s"
            name = "%s"
            published = %v
            context_id = "%s"
			%s
        }
        `, evaluationForm.Id,
		evaluationForm.Name,
		evaluationForm.Published,
		evaluationForm.ContextId,
		generateQuestionGroups(&evaluationForm.QuestionGroups),
	)

	return evaluationFormString
}

func generateQuestionGroups(groups *[]Evaluationquestiongroup) string {
	if *groups == nil || len(*groups) <= 0 {
		return ""
	}

	groupsString := ""
	for _, group := range *groups {
		groupString := fmt.Sprintf(`
        	question_groups {
        	    id = "%s"
        	    name = "%s"
        	    type = "%s"
        	    default_answers_to_highest = %v
        	    default_answers_to_na = %v
        	    na_enabled = %v
				weight = %v
				manual_weight = %v
				%s
				%s
        	}
        	`, group.Id,
			group.Name,
			group.VarType,
			group.DefaultAnswersToHighest,
			group.DefaultAnswersToNA,
			group.NaEnabled,
			group.Weight,
			group.ManualWeight,
			generateQuestions(&group.Questions),
			generateVisibilityCondition(&group.VisibilityCondition),
		)

		groupsString += groupString
	}

	return groupsString
}

func generateQuestions(questions *[]Evaluationquestion) string {
	if *questions == nil || len(*questions) <= 0 {
		return ""
	}

	questionsString := ""
	for _, question := range *questions {
		questionString := fmt.Sprintf(`
        	questions {
        	    id = "%s"
        	    text = "%s"
        	    help_text = "%s"
        	    type = "%s"
        	    na_enabled = %v
        	    comments_required = %v
				%s
				%s
				is_kill = %v
        	    is_critical = %v
        	}
        	`, question.Id,
			question.Text,
			question.HelpText,
			question.VarType,
			question.NaEnabled,
			question.CommentsRequired,
			generateVisibilityCondition(&question.VisibilityCondition),
			generateAnswerOptions(&question.AnswerOptions),
			question.IsKill,
			question.IsCritical,
		)

		questionsString += questionString
	}

	return questionsString
}

func generateAnswerOptions(options *[]Answeroption) string {
	if *options == nil || len(*options) <= 0 {
		return ""
	}

	optionsString := ""
	for _, option := range *options {
		optionString := fmt.Sprintf(`
        	answer_options {
        	    id = "%s"
        	    text = "%s"
        	    value = %v
        	}
        	`, option.Id,
			option.Text,
			option.Value,
		)

		optionsString += optionString
	}

	return optionsString
}

func generateVisibilityCondition(visibilityCondition *Visibilitycondition) string {
	if reflect.DeepEqual(*visibilityCondition, Visibilitycondition{}) {
		return ""
	}

	predicatesString := ""
	for i, v := range visibilityCondition.Predicates {
		if i > 0 {
			predicatesString += ", "
		}
		predicatesString += strconv.Quote(v)
	}

	visibilityConditionString := fmt.Sprintf(`
        visibility_condition {
            combining_operation = "%s"
			predicates = [%s]
        }
        `, visibilityCondition.CombiningOperation,
		predicatesString,
	)

	return visibilityConditionString
}

func generateUser(propertyName string, user *User) string {
	if reflect.DeepEqual(*user, User{}) {
		return ""
	}

	certificationsString := ""
	for i, v := range user.Certifications {
		if i > 0 {
			certificationsString += ", "
		}
		certificationsString += strconv.Quote(v)
	}

	userString := fmt.Sprintf(`
		%s {
			id = "%s"
			name = "%s"
			%s
			%s
			department = "%s"
			email = "%s"
			%s
			title = "%s"
			username = "%s"
			%s
			version = %v
			certifications = [%s]
			%s
			%s
			acd_auto_answer = %v
			%s
		}
	`, propertyName,
		user.Id,
		user.Name,
		generateDivision(&user.Division),
		generateChat(&user.Chat),
		user.Department,
		user.Email,
		generateAddresses(&user.Addresses),
		user.Title,
		user.Username,
		generateImages(&user.Images),
		user.Version,
		certificationsString,
		generateBiography(&user.Biography),
		generateEmployerInfo(&user.EmployerInfo),
		user.AcdAutoAnswer,
		generateLastTokenIssued(&user.LastTokenIssued),
	)

	return userString
}

func generateUsers(propertyName string, users *[]User) string {
	if *users == nil || len(*users) <= 0 {
		return ""
	}

	usersString := ""
	for _, user := range *users {
		usersString += generateUser(propertyName, &user)
	}

	return usersString
}

func generateChat(chat *Chat) string {
	if reflect.DeepEqual(*chat, Chat{}) {
		return ""
	}

	chatString := fmt.Sprintf(`
        chat {
            jabber_id = "%s"
        }
        `, chat.JabberId,
	)

	return chatString
}

func generateAddresses(addresses *[]Contact) string {
	if *addresses == nil || len(*addresses) <= 0 {
		return ""
	}

	addressesString := ""
	for _, address := range *addresses {
		addressString := fmt.Sprintf(`
        addresses {
            address = "%s"
            media_type = "%s"
            type = "%s"
            extension = "%s"
            country_code = "%s"
            integration = "%s"
        }
        `, address.Address,
			address.MediaType,
			address.VarType,
			address.Extension,
			address.CountryCode,
			address.Integration,
		)

		addressesString += addressString
	}

	return addressesString
}

func generateImages(images *[]Userimage) string {
	if *images == nil || len(*images) <= 0 {
		return ""
	}

	imagesString := ""
	for _, image := range *images {
		imageString := fmt.Sprintf(`
        	images {
        	    resolution = "%s"
        	    image_uri = "%s"
        	}
        	`, image.Resolution,
			image.ImageUri,
		)

		imagesString += imageString
	}

	return imagesString
}

func generateBiography(biography *Biography) string {
	if reflect.DeepEqual(*biography, Biography{}) {
		return ""
	}

	biographyString := fmt.Sprintf(`
        biography {
            biography = "%s"
            interests = "%s"
            hobbies = "%s"
            spouse = "%s"
            %s
        }
        `, biography.Biography,
		biography.Interests,
		biography.Hobbies,
		biography.Spouse,
		generateEducation(&biography.Education),
	)

	return biographyString
}

func generateEducation(education *[]Education) string {
	if *education == nil || len(*education) <= 0 {
		return ""
	}

	educationString := ""
	for _, ed := range *education {
		edString := fmt.Sprintf(`
			education {
				school = "%s"
				field_of_study = "%s"
				notes = "%s"
				date_start = %v
				date_end = %v
			}
			`, ed.School,
			ed.FieldOfStudy,
			ed.Notes,
			ed.DateStart,
			ed.DateEnd,
		)
		educationString += edString
	}

	return educationString
}

func generateEmployerInfo(info *Employerinfo) string {
	if reflect.DeepEqual(*info, Employerinfo{}) {
		return ""
	}

	infoString := fmt.Sprintf(`
        employer_info {
            official_name = "%s"
            employee_id = "%s"
            employee_type = "%s"
            date_hire = "%s"
        }
        `, info.OfficialName,
		info.EmployeeId,
		info.EmployeeType,
		info.DateHire,
	)

	return infoString
}

func generateLastTokenIssued(token *Oauthlasttokenissued) string {
	if reflect.DeepEqual(*token, Oauthlasttokenissued{}) {
		return ""
	}

	tokenString := fmt.Sprintf(`
        last_token_issued {
            date_issued = %v
        }
        `, token.DateIssued,
	)

	return tokenString
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

	conditionsString := fmt.Sprintf(`
        conditions {
			%s
			date_ranges = [%s]
            %s
			%s
			%s
			%s
			directions = [%s]
			%s
        }
        `, generateUsers("for_users", &conditions.ForUsers),
		dateRangesString,
		generateQueues(&conditions.ForQueues),
		generateWrapupCodes(&conditions.WrapupCodes),
		generateLanguages(&conditions.Languages),
		generateTimeAllowed(&conditions.TimeAllowed),
		directionsString,
		generateDuration(&conditions.Duration),
	)

	return conditionsString
}

func generateLanguages(languages *[]Language) string {
	if *languages == nil || len(*languages) <= 0 {
		return ""
	}

	languagesString := ""
	for _, language := range *languages {
		languageString := fmt.Sprintf(`
        	languages {
        	    id = "%s"
        	    name = "%s"
        	    state = "%s"
        	    version = "%s"
        	}
        	`, language.Id,
			language.Name,
			language.State,
			language.Version,
		)

		languagesString += languageString
	}

	return languagesString
}

func generateQueues(queues *[]Queue) string {
	if *queues == nil || len(*queues) <= 0 {
		return ""
	}

	queuesString := ""

	for _, queue := range *queues {
		queueString := fmt.Sprintf(`
        	for_queues {
        	    id = "%s"
        		name = "%s"
        		%s
				description = "%s"
				modified_by = "%s"
				created_by = "%s"
				%s
				%s
				%s
				%s
				skill_evaluation_method = "%s"
				%s
				%s
				auto_answer_only = %v
				enable_transcription = %v
				enable_manual_assignment = %v
				calling_party_name = "%s"
				calling_party_number = "%s"
				%s
				%s
        		}
        		`, queue.Id,
			queue.Name,
			generateDivision(&queue.Division),
			queue.Description,
			queue.ModifiedBy,
			queue.CreatedBy,
			generatePolicyMediaSettings(&queue.MediaSettings),
			generatePolicyRoutingRules(&queue.RoutingRules),
			generateBullseye(&queue.Bullseye),
			generateAcwSettings(&queue.AcwSettings),
			queue.SkillEvaluationMethod,
			generateDomainEntityRef("queue_flow", &queue.QueueFlow),
			generateDomainEntityRef("whisper_prompt", &queue.WhisperPrompt),
			queue.AutoAnswerOnly,
			queue.EnableTranscription,
			queue.EnableManualAssignment,
			queue.CallingPartyName,
			queue.CallingPartyNumber,
			generateOutboundMessagingAddresses(&queue.OutboundMessagingAddresses),
			generateOutboundEmailAddress(&queue.OutboundEmailAddress),
		)

		queuesString += queueString
	}

	return queuesString
}

func generateOutboundMessagingAddresses(addresses *Queuemessagingaddresses) string {
	if reflect.DeepEqual(*addresses, Queuemessagingaddresses{}) {
		return ""
	}

	addressString := fmt.Sprintf(`
    	outbound_messaging_addresses {
    	    %s
    	}
    	`, generateDomainEntityRef("sms_address", &addresses.SmsAddress),
	)
	return addressString
}

func generateOutboundEmailAddress(address *Queueemailaddress) string {
	if reflect.DeepEqual(*address, Queueemailaddress{}) {
		return ""
	}

	addressString := fmt.Sprintf(`
    	outbound_messaging_address {
    	    %s
			%s
    	}
    	`, generateDomainEntityRef("domain", &address.Domain),
		generateInboundRoute(&address.Route),
	)
	return addressString
}

func generateInboundRoute(route *Inboundroute) string {
	if reflect.DeepEqual(*route, Inboundroute{}) {
		return ""
	}

	routeString := fmt.Sprintf(`
    	route {
    		id = "%s"
    		name = "%s"
    		pattern = "%s"
			%s
			priority = %v
			%s
			%s
			from_name = "%s"
			from_email = "%s"
			%s
			%s
			%s
    	}
    	`, route.Id,
		route.Name,
		route.Pattern,
		generateDomainEntityRef("queue", &route.Queue),
		route.Priority,
		generateDomainEntityRefs("skills", &route.Skills),
		generateDomainEntityRef("language", &route.Language),
		route.FromName,
		route.FromEmail,
		generateDomainEntityRef("flow", &route.Flow),
		generateEmailAddresses(&route.AutoBcc),
		generateDomainEntityRef("spam_flow", &route.SpamFlow),
	)
	return routeString
}

func generateEmailAddresses(addresses *[]Emailaddress) string {
	if *addresses == nil || len(*addresses) <= 0 {
		return ""
	}

	addressesString := ""
	for _, address := range *addresses {
		addressString := fmt.Sprintf(`
        	auto_bcc {
        	    email = "%s"
        	    name = "%s"
        	}
        	`, address.Email,
			address.Name,
		)

		addressesString += addressString
	}

	return addressesString
}

func generateDomainEntityRefs(propertyName string, refs *[]Domainentityref) string {
	if *refs == nil || len(*refs) <= 0 {
		return ""
	}

	refsString := ""
	for _, ref := range *refs {
		refsString += generateDomainEntityRef(propertyName, &ref)
	}

	return refsString
}

func generateAcwSettings(acwSettings *Acwsettings) string {
	if reflect.DeepEqual(*acwSettings, Acwsettings{}) {
		return ""
	}

	acwSettingsString := fmt.Sprintf(`
        acwSettings {
            wrapup_prompt = "%s"
			timeout_ms = %v
        }
        `, acwSettings.WrapupPrompt,
		acwSettings.TimeoutMs,
	)

	return acwSettingsString
}

func generateBullseye(bullseye *Bullseye) string {
	if reflect.DeepEqual(*bullseye, Bullseye{}) {
		return ""
	}

	bullseyeString := fmt.Sprintf(`
        bullseye {
            %s
        }
        `, generateRings(&bullseye.Rings),
	)

	return bullseyeString
}

func generateRings(rings *[]Ring) string {
	if *rings == nil || len(*rings) <= 0 {
		return ""
	}

	ringsString := ""
	for _, ring := range *rings {
		ringString := fmt.Sprintf(`
        rings {
            %s
			%s
        }
        `, generateExpansionCriteria(&ring.ExpansionCriteria),
			generateActions(&ring.Actions),
		)

		ringsString += ringString
	}

	return ringsString
}

func generateActions(actions *Actions) string {
	if reflect.DeepEqual(*actions, Actions{}) {
		return ""
	}

	actionsString := fmt.Sprintf(`
        actions {
			%s
        }
        `, generateSkillsToRemove(&actions.SkillsToRemove),
	)

	return actionsString
}

func generateSkillsToRemove(skillsToRemove *[]Skillstoremove) string {
	if *skillsToRemove == nil || len(*skillsToRemove) <= 0 {
		return ""
	}

	skillsToRemoveString := ""
	for _, skill := range *skillsToRemove {
		skillString := fmt.Sprintf(`
        	skills_to_remove {
        	    name = "%s"
        	    id = "%s"
        	    self_uri = "%s"
        	}
        	`, skill.Name,
			skill.Id,
			skill.SelfUri,
		)

		skillsToRemoveString += skillString
	}

	return skillsToRemoveString
}

func generateExpansionCriteria(expansionCriteria *[]Expansioncriterium) string {
	if *expansionCriteria == nil || len(*expansionCriteria) <= 0 {
		return ""
	}

	expansionCriteriaString := ""
	for _, criterium := range *expansionCriteria {
		criteriumString := fmt.Sprintf(`
        expansion_criteria {
            type = "%s"
            threshold = %v
        }
        `, criterium.VarType,
			criterium.Threshold,
		)

		expansionCriteriaString += criteriumString
	}

	return expansionCriteriaString
}

func generatePolicyRoutingRules(routingRules *[]Routingrule) string {
	if *routingRules == nil || len(*routingRules) <= 0 {
		return ""
	}

	routingRulesString := ""

	for _, routingRule := range *routingRules {
		routingRuleString := fmt.Sprintf(`
        	routing_rules {
        	    operator = "%s"
				threshold = %v
				wait_seconds = %v
        	}
        	`, routingRule.Operator,
			routingRule.Threshold,
			routingRule.WaitSeconds,
		)

		routingRulesString += routingRuleString
	}

	return routingRulesString
}

func generatePolicyMediaSettings(mediaSettings *map[string]Mediasetting) string {
	if *mediaSettings == nil || len(*mediaSettings) <= 0 {
		return ""
	}

	mediaSettingsString := ""
	for key, mediaSetting := range *mediaSettings {
		mediaSettingString := fmt.Sprintf(`
        	%s {
        	    alerting_timeout_seconds = %v
        	    %s
        	}
        	`, key,
			mediaSetting.AlertingTimeoutSeconds,
			generateServiceLevel(&mediaSetting.ServiceLevel),
		)

		mediaSettingsString += mediaSettingString
	}

	return mediaSettingsString
}

func generateServiceLevel(serviceLevel *Servicelevel) string {
	if reflect.DeepEqual(*serviceLevel, Servicelevel{}) {
		return ""
	}

	serviceLevelString := fmt.Sprintf(`
        service_level {
            percentage = %v
            duration_ms = %v
        }
        `, serviceLevel.Percentage,
		serviceLevel.DurationMs,
	)

	return serviceLevelString
}

func generateDivision(division *Division) string {
	if reflect.DeepEqual(*division, Division{}) {
		return ""
	}

	divisionString := fmt.Sprintf(`
        division {
			id = "%s"
            name = "%s"
            self_uri = "%s"
        }
        `, division.Id,
		division.Name,
		division.SelfUri,
	)

	return divisionString
}

func generateAssignSurveys(assignSurveys *[]Surveyassignment) string {
	if *assignSurveys == nil || len(*assignSurveys) <= 0 {
		return ""
	}

	assignSurveysString := ""

	for _, assignSurvey := range *assignSurveys {
		assignSurveyString := fmt.Sprintf(`
        assign_surveys {
            sending_domain = "%s"
            %s
            %s
        }
        `, assignSurvey.SendingDomain,
			generateSurveyForm(&assignSurvey.SurveyForm),
			generateFlow(&assignSurvey.Flow),
		)

		assignSurveysString += assignSurveyString
	}

	return assignSurveysString
}

func generateSurveyForm(surveyForm *Publishedsurveyformreference) string {
	if reflect.DeepEqual(*surveyForm, Publishedsurveyformreference{}) {
		return ""
	}

	surveyFormString := fmt.Sprintf(`
        survey_form {
            name = "%s"
            context_id = "%s"
        }
        `, surveyForm.Name,
		surveyForm.ContextId,
	)

	return surveyFormString
}

func generateFlow(flow *Domainentityref) string {
	if reflect.DeepEqual(*flow, Domainentityref{}) {
		return ""
	}

	flowString := fmt.Sprintf(`
        flow{
            id = "%s"
            name = "%s"
            self_uri = "%s"
        }
        `, flow.Id,
		flow.Name,
		flow.SelfUri,
	)

	return flowString
}
