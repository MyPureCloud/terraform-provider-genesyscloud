package routing_queue

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceRoutingQueueCreate(t *testing.T) {
	tId := uuid.NewString()
	tName := "unit test routing queue"
	testRoutingQueue := generateRoutingQueueData(tId, tName)

	queueProxy := &RoutingQueueProxy{}
	queueProxy.getRoutingQueueByIdAttr = func(ctx context.Context, p *RoutingQueueProxy, queueId string, checkCache bool) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, queueId)
		routingQueue := &testRoutingQueue

		queue := convertCreateQueuetoQueue(*routingQueue)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return queue, apiResponse, nil
	}

	queueProxy.createRoutingQueueAttr = func(ctx context.Context, p *RoutingQueueProxy, routingQueue *platformclientv2.Createqueuerequest) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		routingQueue.Id = &tId
		queue := convertCreateQueuetoQueue(*routingQueue)

		assert.Equal(t, testRoutingQueue.Id, routingQueue.Id, "ID Not Equal")
		assert.Equal(t, testRoutingQueue.Name, routingQueue.Name, "Name Not Equal")
		assert.Equal(t, testRoutingQueue.Description, routingQueue.Description, "Description Not Equal")
		assert.Equal(t, testRoutingQueue.ScoringMethod, routingQueue.ScoringMethod, "Scoring Method Not Equal")
		assert.Equal(t, testRoutingQueue.SkillEvaluationMethod, routingQueue.SkillEvaluationMethod, "Skill Evaluation Method Not Equal")
		assert.Equal(t, testRoutingQueue.AutoAnswerOnly, routingQueue.AutoAnswerOnly, "Auto Answer Only Not Equal")
		assert.Equal(t, testRoutingQueue.EnableTranscription, routingQueue.EnableTranscription, "Enable Transcription Not Equal")
		assert.Equal(t, testRoutingQueue.EnableAudioMonitoring, routingQueue.EnableAudioMonitoring, "Enable Audio Monitoring Not Equal")
		assert.Equal(t, testRoutingQueue.EnableManualAssignment, routingQueue.EnableManualAssignment, "Enable Manual Assignment Not Equal")
		assert.Equal(t, testRoutingQueue.CallingPartyName, routingQueue.CallingPartyName, "Calling Party Name Not Equal")
		assert.Equal(t, testRoutingQueue.CallingPartyNumber, routingQueue.CallingPartyNumber, "Calling Party Number Not Equal")
		assert.Equal(t, testRoutingQueue.PeerId, routingQueue.PeerId, "Peer ID Not Equal")
		assert.Equal(t, testRoutingQueue.AcwSettings, routingQueue.AcwSettings, "ACW Settings Not Equal")
		assert.Equal(t, testRoutingQueue.OutboundMessagingAddresses, routingQueue.OutboundMessagingAddresses, "Outbound Messaging Addresses Not Equal")
		assert.Equal(t, testRoutingQueue.SuppressInQueueCallRecording, routingQueue.SuppressInQueueCallRecording, "Suppress In-Queue Call Recording Not Equal")
		assert.Equal(t, testRoutingQueue.DirectRouting, routingQueue.DirectRouting, "Direct Routing Not Equal")
		assert.Equal(t, testRoutingQueue.AgentOwnedRouting, routingQueue.AgentOwnedRouting, "Agent Owned Routing Not Equal")
		assert.Equal(t, testRoutingQueue.RoutingRules, routingQueue.RoutingRules, "Routing Rules Not Equal")
		assert.Equal(t, testRoutingQueue.MediaSettings, routingQueue.MediaSettings, "Media Settings Not Equal")
		assert.Equal(t, testRoutingQueue.QueueFlow, routingQueue.QueueFlow, "Queue Flow Not Equal")
		assert.Equal(t, testRoutingQueue.EmailInQueueFlow, routingQueue.EmailInQueueFlow, "Email In Queue Flow Not Equal")
		assert.Equal(t, testRoutingQueue.MessageInQueueFlow, routingQueue.MessageInQueueFlow, "Message In Queue Flow Not Equal")
		assert.Equal(t, testRoutingQueue.WhisperPrompt, routingQueue.WhisperPrompt, "Whisper Prompt Not Equal")
		assert.Equal(t, testRoutingQueue.OnHoldPrompt, routingQueue.OnHoldPrompt, "On Hold Prompt Not Equal")
		assert.Equal(t, testRoutingQueue.DefaultScripts, routingQueue.DefaultScripts, "Default Scripts Not Equal")
		assert.Equal(t, testRoutingQueue.MediaSettings.Message.SubTypeSettings, routingQueue.MediaSettings.Message.SubTypeSettings, "SubTypeSettings Not Equal")
		assert.Equal(t, testRoutingQueue.CannedResponseLibraries, routingQueue.CannedResponseLibraries, "Canned Response Libraries not equal")
		assert.Equal(t, *testRoutingQueue.LastAgentRoutingMode, *routingQueue.LastAgentRoutingMode, "LastAgentRoutingMode Not Equal")
		return queue, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	queueProxy.getAllRoutingQueueWrapupCodesAttr = func(ctx context.Context, proxy *RoutingQueueProxy, queueId string) (*[]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
		wrapupCodes := []platformclientv2.Wrapupcode{
			{
				Id:   platformclientv2.String("wrapupCode1"),
				Name: platformclientv2.String("Wrapup Code 1"),
			},
			{
				Id:   platformclientv2.String("wrapupCode2"),
				Name: platformclientv2.String("Wrapup Code 2"),
			},
		}

		return &wrapupCodes, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	err := setRoutingQueueUnitTestsEnvVar()
	if err != nil {
		t.Skipf("failed to set env variable %s: %s", unitTestsAreActiveEnv, err.Error())
	}

	internalProxy = queueProxy
	defer func() {
		internalProxy = nil
		err = unsetRoutingQueueUnitTestsEnvVar()
		if err != nil {
			t.Logf("Failed to unset env variable %s: %s", unitTestsAreActiveEnv, err.Error())
		}
	}()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceRoutingQueue().Schema
	testRoutingQueue.CannedResponseLibraries = nil // setting this to nil because TestResourceDataRaw seems to have problems with TypeSet
	resourceDataMap := buildRoutingQueueResourceMap(tId, *testRoutingQueue.Name, testRoutingQueue)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := createRoutingQueue(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceRoutingQueueRead(t *testing.T) {
	tId := uuid.NewString()
	tName := "unit test routing queue"
	testRoutingQueue := generateRoutingQueueData(tId, tName)
	testRoutingQueue.CannedResponseLibraries = nil // setting this to nil because TestResourceDataRaw seems to have problems with TypeSet

	queueProxy := &RoutingQueueProxy{}
	queueProxy.getRoutingQueueByIdAttr = func(ctx context.Context, proxy *RoutingQueueProxy, id string, checkCache bool) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		routingQueue := &testRoutingQueue

		queue := convertCreateQueuetoQueue(*routingQueue)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return queue, apiResponse, nil
	}

	queueProxy.getAllRoutingQueueWrapupCodesAttr = func(ctx context.Context, proxy *RoutingQueueProxy, queueId string) (*[]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
		wrapupCodes := []platformclientv2.Wrapupcode{
			{
				Id:   platformclientv2.String("wrapupCode1"),
				Name: platformclientv2.String("Wrapup Code 1"),
			},
			{
				Id:   platformclientv2.String("wrapupCode2"),
				Name: platformclientv2.String("Wrapup Code 2"),
			},
		}

		return &wrapupCodes, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	err := setRoutingQueueUnitTestsEnvVar()
	if err != nil {
		t.Skipf("failed to set env variable %s: %s", unitTestsAreActiveEnv, err.Error())
	}

	internalProxy = queueProxy
	defer func() {
		internalProxy = nil
		err := unsetRoutingQueueUnitTestsEnvVar()
		if err != nil {
			t.Logf("Failed to unset env variable %s: %s", unitTestsAreActiveEnv, err.Error())
		}
	}()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceRoutingQueue().Schema

	resourceDataMap := buildRoutingQueueResourceMap(tId, *testRoutingQueue.Name, testRoutingQueue)
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readRoutingQueue(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())

	routingQueue := getRoutingQueueFromResourceData(d)
	routingQueue.Id = platformclientv2.String(d.Id())

	assert.Equal(t, *testRoutingQueue.Id, d.Id(), "ID Not Equal")
	assert.Equal(t, *testRoutingQueue.Name, d.Get("name").(string), "Name Not Equal")
	assert.Equal(t, *testRoutingQueue.Description, d.Get("description").(string), "Description Not Equal")
	assert.Equal(t, testRoutingQueue.ScoringMethod, routingQueue.ScoringMethod, "Scoring Method Not Equal")
	assert.Equal(t, testRoutingQueue.SkillEvaluationMethod, routingQueue.SkillEvaluationMethod, "Skill Evaluation Method Not Equal")
	assert.Equal(t, testRoutingQueue.AutoAnswerOnly, routingQueue.AutoAnswerOnly, "Auto Answer Only Not Equal")
	assert.Equal(t, testRoutingQueue.EnableTranscription, routingQueue.EnableTranscription, "Enable Transcription Not Equal")
	assert.Equal(t, testRoutingQueue.EnableAudioMonitoring, routingQueue.EnableAudioMonitoring, "Enable Audio Monitoring Not Equal")
	assert.Equal(t, testRoutingQueue.EnableManualAssignment, routingQueue.EnableManualAssignment, "Enable Manual Assignment Not Equal")
	assert.Equal(t, testRoutingQueue.CallingPartyName, routingQueue.CallingPartyName, "Calling Party Name Not Equal")
	assert.Equal(t, testRoutingQueue.CallingPartyNumber, routingQueue.CallingPartyNumber, "Calling Party Number Not Equal")
	assert.Equal(t, testRoutingQueue.PeerId, routingQueue.PeerId, "Peer ID Not Equal")
	assert.Equal(t, testRoutingQueue.AcwSettings, routingQueue.AcwSettings, "ACW Settings Not Equal")
	assert.Equal(t, testRoutingQueue.OutboundMessagingAddresses, routingQueue.OutboundMessagingAddresses, "Outbound Messaging Addresses Not Equal")
	assert.Equal(t, testRoutingQueue.SuppressInQueueCallRecording, routingQueue.SuppressInQueueCallRecording, "Suppress In-Queue Call Recording Not Equal")
	assert.Equal(t, testRoutingQueue.DirectRouting, routingQueue.DirectRouting, "Direct Routing Not Equal")
	assert.Equal(t, testRoutingQueue.AgentOwnedRouting, routingQueue.AgentOwnedRouting, "Agent Owned Routing Not Equal")
	assert.Equal(t, testRoutingQueue.RoutingRules, routingQueue.RoutingRules, "Routing Rules Not Equal")
	assert.Equal(t, testRoutingQueue.MediaSettings, routingQueue.MediaSettings, "Media Settings Not Equal")
	assert.Equal(t, testRoutingQueue.QueueFlow, routingQueue.QueueFlow, "Queue Flow Not Equal")
	assert.Equal(t, testRoutingQueue.EmailInQueueFlow, routingQueue.EmailInQueueFlow, "Email In Queue Flow Not Equal")
	assert.Equal(t, testRoutingQueue.MessageInQueueFlow, routingQueue.MessageInQueueFlow, "Message In Queue Flow Not Equal")
	assert.Equal(t, testRoutingQueue.WhisperPrompt, routingQueue.WhisperPrompt, "Whisper Prompt Not Equal")
	assert.Equal(t, testRoutingQueue.OnHoldPrompt, routingQueue.OnHoldPrompt, "On Hold Prompt Not Equal")
	assert.Equal(t, testRoutingQueue.DefaultScripts, routingQueue.DefaultScripts, "Default Scripts Not Equal")
	assert.Equal(t, testRoutingQueue.MediaSettings.Message.SubTypeSettings, routingQueue.MediaSettings.Message.SubTypeSettings, "SubTypeSettings Not Equal")
	assert.Equal(t, *testRoutingQueue.LastAgentRoutingMode, d.Get("last_agent_routing_mode").(string), "LastAgentRoutingMode Not Equal")
}

func TestUnitResourceRoutingQueueUpdate(t *testing.T) {
	tId := uuid.NewString()
	tName := "updated queue name"
	testRoutingQueue := generateRoutingQueueData(tId, tName)
	testRoutingQueue.CannedResponseLibraries = nil // setting this to nil because TestResourceDataRaw seems to have problems with TypeSet

	queueProxy := &RoutingQueueProxy{}
	queueProxy.getRoutingQueueByIdAttr = func(ctx context.Context, proxy *RoutingQueueProxy, id string, checkCache bool) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		routingQueue := &testRoutingQueue

		queue := convertCreateQueuetoQueue(*routingQueue)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return queue, apiResponse, nil
	}

	queueProxy.updateRoutingQueueAttr = func(ctx context.Context, proxy *RoutingQueueProxy, id string, routingQueue *platformclientv2.Queuerequest) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		routingQueue.Id = &tId

		assert.Equal(t, testRoutingQueue.Id, routingQueue.Id, "ID Not Equal")
		assert.Equal(t, testRoutingQueue.Name, routingQueue.Name, "Name Not Equal")
		assert.Equal(t, testRoutingQueue.Description, routingQueue.Description, "Description Not Equal")
		assert.Equal(t, testRoutingQueue.ScoringMethod, routingQueue.ScoringMethod, "Scoring Method Not Equal")
		assert.Equal(t, testRoutingQueue.SkillEvaluationMethod, routingQueue.SkillEvaluationMethod, "Skill Evaluation Method Not Equal")
		assert.Equal(t, testRoutingQueue.AutoAnswerOnly, routingQueue.AutoAnswerOnly, "Auto Answer Only Not Equal")
		assert.Equal(t, testRoutingQueue.EnableTranscription, routingQueue.EnableTranscription, "Enable Transcription Not Equal")
		assert.Equal(t, testRoutingQueue.EnableAudioMonitoring, routingQueue.EnableAudioMonitoring, "Enable Audio Monitoring Not Equal")
		assert.Equal(t, testRoutingQueue.EnableManualAssignment, routingQueue.EnableManualAssignment, "Enable Manual Assignment Not Equal")
		assert.Equal(t, testRoutingQueue.CallingPartyName, routingQueue.CallingPartyName, "Calling Party Name Not Equal")
		assert.Equal(t, testRoutingQueue.CallingPartyNumber, routingQueue.CallingPartyNumber, "Calling Party Number Not Equal")
		assert.Equal(t, testRoutingQueue.PeerId, routingQueue.PeerId, "Peer ID Not Equal")
		assert.Equal(t, testRoutingQueue.AcwSettings, routingQueue.AcwSettings, "ACW Settings Not Equal")
		assert.Equal(t, testRoutingQueue.OutboundMessagingAddresses, routingQueue.OutboundMessagingAddresses, "Outbound Messaging Addresses Not Equal")
		assert.Equal(t, testRoutingQueue.SuppressInQueueCallRecording, routingQueue.SuppressInQueueCallRecording, "Suppress In-Queue Call Recording Not Equal")
		assert.Equal(t, testRoutingQueue.DirectRouting, routingQueue.DirectRouting, "Direct Routing Not Equal")
		assert.Equal(t, testRoutingQueue.AgentOwnedRouting, routingQueue.AgentOwnedRouting, "Agent Owned Routing Not Equal")
		assert.Equal(t, testRoutingQueue.RoutingRules, routingQueue.RoutingRules, "Routing Rules Not Equal")
		assert.Equal(t, testRoutingQueue.MediaSettings, routingQueue.MediaSettings, "Media Settings Not Equal")
		assert.Equal(t, testRoutingQueue.QueueFlow, routingQueue.QueueFlow, "Queue Flow Not Equal")
		assert.Equal(t, testRoutingQueue.EmailInQueueFlow, routingQueue.EmailInQueueFlow, "Email In Queue Flow Not Equal")
		assert.Equal(t, testRoutingQueue.MessageInQueueFlow, routingQueue.MessageInQueueFlow, "Message In Queue Flow Not Equal")
		assert.Equal(t, testRoutingQueue.WhisperPrompt, routingQueue.WhisperPrompt, "Whisper Prompt Not Equal")
		assert.Equal(t, testRoutingQueue.OnHoldPrompt, routingQueue.OnHoldPrompt, "On Hold Prompt Not Equal")
		assert.Equal(t, testRoutingQueue.DefaultScripts, routingQueue.DefaultScripts, "Default Scripts Not Equal")
		assert.Equal(t, testRoutingQueue.MediaSettings.Message.SubTypeSettings, routingQueue.MediaSettings.Message.SubTypeSettings, "SubTypeSettings Not Equal")
		assert.Equal(t, *testRoutingQueue.LastAgentRoutingMode, *routingQueue.LastAgentRoutingMode, "LastAgentRoutingMode Not Equal")

		return nil, nil, nil
	}

	queueProxy.getAllRoutingQueueWrapupCodesAttr = func(ctx context.Context, proxy *RoutingQueueProxy, queueId string) (*[]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
		wrapupCodes := []platformclientv2.Wrapupcode{
			{
				Id:   platformclientv2.String("wrapupCode1"),
				Name: platformclientv2.String("Wrapup Code 1"),
			},
			{
				Id:   platformclientv2.String("wrapupCode2"),
				Name: platformclientv2.String("Wrapup Code 2"),
			},
		}

		return &wrapupCodes, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	err := setRoutingQueueUnitTestsEnvVar()
	if err != nil {
		t.Skipf("failed to set env variable %s: %s", unitTestsAreActiveEnv, err.Error())
	}

	internalProxy = queueProxy
	defer func() {
		internalProxy = nil
		err := unsetRoutingQueueUnitTestsEnvVar()
		if err != nil {
			t.Logf("Failed to unset env variable %s: %s", unitTestsAreActiveEnv, err.Error())
		}
	}()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceRoutingQueue().Schema
	//Setup a map of values
	resourceDataMap := buildRoutingQueueResourceMap(tId, *testRoutingQueue.Name, testRoutingQueue)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateRoutingQueue(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, *testRoutingQueue.Name, d.Get("name").(string))
}

func TestUnitResourceRoutingQueueDelete(t *testing.T) {
	tId := uuid.NewString()
	tName := "unit test routing queue"
	testRoutingQueue := generateRoutingQueueData(tId, tName)
	testRoutingQueue.CannedResponseLibraries = nil // setting this to nil because TestResourceDataRaw seems to have problems with TypeSet

	queueProxy := &RoutingQueueProxy{}
	queueProxy.deleteRoutingQueueAttr = func(ctx context.Context, p *RoutingQueueProxy, queueId string, forceDelete bool) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, queueId)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	queueProxy.getRoutingQueueByIdAttr = func(ctx context.Context, proxy *RoutingQueueProxy, id string, checkCache bool) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
		err := fmt.Errorf("Unable to find targeted queue: %s", id)
		return nil, apiResponse, err
	}

	err := setRoutingQueueUnitTestsEnvVar()
	if err != nil {
		t.Skipf("failed to set env variable %s: %s", unitTestsAreActiveEnv, err.Error())
	}

	internalProxy = queueProxy
	defer func() {
		internalProxy = nil
		err := unsetRoutingQueueUnitTestsEnvVar()
		if err != nil {
			t.Logf("Failed to unset env variable %s: %s", unitTestsAreActiveEnv, err.Error())
		}
	}()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceRoutingQueue().Schema
	//Setup a map of values
	resourceDataMap := buildRoutingQueueResourceMap(tId, *testRoutingQueue.Name, testRoutingQueue)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteRoutingQueue(ctx, d, gcloud)
	assert.Nil(t, diag)
	assert.Equal(t, tId, d.Id())
}

func TestUnitBuildSdkMediaSettingCallback(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected *platformclientv2.Callbackmediasettings
	}{
		{
			name:     "Empty input",
			input:    []any{},
			expected: nil,
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "All fields populated",
			input: []any{
				map[string]interface{}{
					"alerting_timeout_sec":             30,
					"service_level_percentage":         0.8,
					"service_level_duration_ms":        20000,
					"enable_auto_answer":               true,
					"auto_end_delay_seconds":           20,
					"auto_dial_delay_seconds":          4,
					"enable_auto_dial_and_end":         false,
					"mode":                             "AgentFirst",
					"auto_answer_alert_tone_seconds":   float64(5),
					"manual_answer_alert_tone_seconds": float64(6),
					"pacing_modifier":                  float64(3),
					"live_voice_reaction_type":         "HangUp",
					"live_voice_flow_id":               "123",
					"answering_machine_reaction_type":  "Transfer",
					"answering_machine_flow_id":        "321",
				},
			},
			expected: &platformclientv2.Callbackmediasettings{
				AlertingTimeoutSeconds:       platformclientv2.Int(30),
				ServiceLevel:                 &platformclientv2.Servicelevel{Percentage: platformclientv2.Float64(0.8), DurationMs: platformclientv2.Int(20000)},
				EnableAutoAnswer:             platformclientv2.Bool(true),
				AutoEndDelaySeconds:          platformclientv2.Int(20),
				AutoDialDelaySeconds:         platformclientv2.Int(4),
				EnableAutoDialAndEnd:         platformclientv2.Bool(false),
				Mode:                         platformclientv2.String("AgentFirst"),
				AutoAnswerAlertToneSeconds:   platformclientv2.Float64(5),
				ManualAnswerAlertToneSeconds: platformclientv2.Float64(6),
				PacingModifier:               platformclientv2.Float64(3),
				LiveVoiceReactionType:        platformclientv2.String("HangUp"),
				LiveVoiceFlow:                &platformclientv2.Domainentityref{Id: platformclientv2.String("123")},
				AnsweringMachineReactionType: platformclientv2.String("Transfer"),
				AnsweringMachineFlow:         &platformclientv2.Domainentityref{Id: platformclientv2.String("321")},
			},
		},
		{
			name: "Zero values are not built when they shouldn't be & vice versa",
			input: []any{
				map[string]interface{}{
					"alerting_timeout_sec":             0,
					"service_level_percentage":         float64(0),
					"service_level_duration_ms":        0,
					"enable_auto_answer":               false,
					"auto_end_delay_seconds":           0,
					"auto_dial_delay_seconds":          0,
					"enable_auto_dial_and_end":         false,
					"mode":                             "",
					"auto_answer_alert_tone_seconds":   float64(0),
					"manual_answer_alert_tone_seconds": float64(0),
					"pacing_modifier":                  float64(0),
					"live_voice_reaction_type":         "",
					"live_voice_flow_id":               "",
					"answering_machine_reaction_type":  "",
					"answering_machine_flow_id":        "",
				},
			},
			expected: &platformclientv2.Callbackmediasettings{
				AlertingTimeoutSeconds:       platformclientv2.Int(0),
				ServiceLevel:                 &platformclientv2.Servicelevel{Percentage: platformclientv2.Float64(0), DurationMs: platformclientv2.Int(0)},
				EnableAutoAnswer:             platformclientv2.Bool(false),
				AutoEndDelaySeconds:          platformclientv2.Int(0),
				AutoDialDelaySeconds:         platformclientv2.Int(0),
				EnableAutoDialAndEnd:         platformclientv2.Bool(false),
				Mode:                         nil,
				AutoAnswerAlertToneSeconds:   platformclientv2.Float64(0),
				ManualAnswerAlertToneSeconds: platformclientv2.Float64(0),
				PacingModifier:               nil,
				LiveVoiceReactionType:        nil,
				LiveVoiceFlow:                nil,
				AnsweringMachineReactionType: nil,
				AnsweringMachineFlow:         nil,
			},
		},
		{
			name: "Empty map produces all nil fields",
			input: []any{
				map[string]interface{}{},
			},
			expected: &platformclientv2.Callbackmediasettings{
				AlertingTimeoutSeconds:       nil,
				ServiceLevel:                 nil,
				EnableAutoAnswer:             nil,
				AutoEndDelaySeconds:          nil,
				AutoDialDelaySeconds:         nil,
				EnableAutoDialAndEnd:         nil,
				Mode:                         nil,
				AutoAnswerAlertToneSeconds:   nil,
				ManualAnswerAlertToneSeconds: nil,
				PacingModifier:               nil,
				LiveVoiceReactionType:        nil,
				LiveVoiceFlow:                nil,
				AnsweringMachineReactionType: nil,
				AnsweringMachineFlow:         nil,
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := buildSdkMediaSettingCallback(tt.input)

			// Compare expected and actual results
			if tt.expected == nil && actual != nil {
				t.Errorf("Expected nil but got %v", actual)
				return
			}
			if tt.expected != nil && actual == nil {
				t.Errorf("Expected %v but got nil", tt.expected)
				return
			}

			if tt.expected == nil || actual == nil {
				return
			}

			if !util.EquivalentJsons(tt.expected.String(), actual.String()) {
				t.Errorf("JSON objects are not equal.\nExpected: %s\nActual: %s\n", tt.expected.String(), actual.String())
				return
			}
		})
	}
}

func buildRoutingQueueResourceMap(tId string, tName string, testRoutingQueue platformclientv2.Createqueuerequest) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":                                tId,
		"name":                              tName,
		"description":                       *testRoutingQueue.Description,
		"scoring_method":                    *testRoutingQueue.ScoringMethod,
		"skill_evaluation_method":           *testRoutingQueue.SkillEvaluationMethod,
		"auto_answer_only":                  *testRoutingQueue.AutoAnswerOnly,
		"enable_transcription":              *testRoutingQueue.EnableTranscription,
		"enable_audio_monitoring":           *testRoutingQueue.EnableAudioMonitoring,
		"enable_manual_assignment":          *testRoutingQueue.EnableManualAssignment,
		"calling_party_name":                *testRoutingQueue.CallingPartyName,
		"calling_party_number":              *testRoutingQueue.CallingPartyNumber,
		"peer_id":                           *testRoutingQueue.PeerId,
		"source_queue_id":                   *testRoutingQueue.SourceQueueId,
		"acw_timeout_ms":                    *testRoutingQueue.AcwSettings.TimeoutMs,
		"acw_wrapup_prompt":                 *testRoutingQueue.AcwSettings.WrapupPrompt,
		"outbound_messaging_sms_address_id": *testRoutingQueue.OutboundMessagingAddresses.SmsAddress.Id,
		"outbound_messaging_open_messaging_recipient_id": *testRoutingQueue.OutboundMessagingAddresses.OpenMessagingRecipient.Id,
		"outbound_messaging_whatsapp_recipient_id":       *testRoutingQueue.OutboundMessagingAddresses.WhatsAppRecipient.Id,
		"suppress_in_queue_call_recording":               *testRoutingQueue.SuppressInQueueCallRecording,
		"direct_routing":                                 flattenDirectRouting(testRoutingQueue.DirectRouting),
		"agent_owned_routing":                            flattenAgentOwnedRouting(testRoutingQueue.AgentOwnedRouting),
		"routing_rules":                                  flattenRoutingRules(testRoutingQueue.RoutingRules),
		"media_settings_call":                            flattenMediaSetting(testRoutingQueue.MediaSettings.Call),
		"media_settings_email":                           flattenMediaEmailSetting(testRoutingQueue.MediaSettings.Email),
		"media_settings_chat":                            flattenMediaSetting(testRoutingQueue.MediaSettings.Chat),
		"media_settings_callback":                        flattenMediaSettingCallback(testRoutingQueue.MediaSettings.Callback),
		"media_settings_message":                         flattenMediaSettingsMessage(testRoutingQueue.MediaSettings.Message),
		"queue_flow_id":                                  *testRoutingQueue.QueueFlow.Id,
		"email_in_queue_flow_id":                         *testRoutingQueue.EmailInQueueFlow.Id,
		"message_in_queue_flow_id":                       *testRoutingQueue.MessageInQueueFlow.Id,
		"whisper_prompt_id":                              *testRoutingQueue.WhisperPrompt.Id,
		"on_hold_prompt_id":                              *testRoutingQueue.OnHoldPrompt.Id,
		"default_script_ids":                             flattenDefaultScripts(*testRoutingQueue.DefaultScripts),
		"canned_response_libraries":                      flattenCannedResponse(testRoutingQueue.CannedResponseLibraries),
		"last_agent_routing_mode":                        *testRoutingQueue.LastAgentRoutingMode,
	}
	return resourceDataMap
}

func generateRoutingQueueData(id, name string) platformclientv2.Createqueuerequest {
	var (
		description           = "Unit Test Description"
		scoringMethod         = "TimestampAndPriority"
		skillEvaluationMethod = "ALL"
		callingPartyName      = "Unit Test Inc."
		callingPartyNumber    = "123"
		peerId                = "5696a54c-4009-4e63-826c-311679deeb97"
		sourceQueueId         = "5696a54c-4009-4e63-826c-311679deeb97"
		backupQueueId         = "5696a54c-4009-4e63-826c-311679deeb97"
		lastAgentRoutingMode  = "QueueMembersOnly"

		acwWrapupPrompt = "MANDATORY_TIMEOUT"
		acwTimeoutMs    = 300000

		queueFlow     = generateRandomDomainEntityRef()
		emailFlow     = generateRandomDomainEntityRef()
		messageFlow   = generateRandomDomainEntityRef()
		whisperPrompt = generateRandomDomainEntityRef()
		onHoldPrompt  = generateRandomDomainEntityRef()

		// ACW Settings
		acwSettings = platformclientv2.Acwsettings{
			WrapupPrompt: &acwWrapupPrompt,
			TimeoutMs:    &acwTimeoutMs,
		}

		// Outbound Messaging Addresses
		smsAddress             = generateRandomDomainEntityRef()
		openMessagingRecipient = generateRandomDomainEntityRef()
		whatsAppRecipient      = generateRandomDomainEntityRef()

		messagingAddress = platformclientv2.Queuemessagingaddresses{
			SmsAddress:             &smsAddress,
			OpenMessagingRecipient: &openMessagingRecipient,
			WhatsAppRecipient:      &whatsAppRecipient,
		}

		// Direct Routing
		callMediaSettings    = generateDirectRoutingMediaSettings()
		emailMediaSettings   = generateDirectRoutingMediaSettings()
		messageMediaSettings = generateDirectRoutingMediaSettings()

		directRouting = platformclientv2.Directrouting{
			CallMediaSettings:    &callMediaSettings,
			EmailMediaSettings:   &emailMediaSettings,
			MessageMediaSettings: &messageMediaSettings,
			WaitForAgent:         platformclientv2.Bool(false),
			AgentWaitSeconds:     platformclientv2.Int(20),
			BackupQueueId:        &backupQueueId,
		}

		// Agent Owned Routing
		agentOwnedRouting = platformclientv2.Agentownedrouting{
			EnableAgentOwnedCallbacks:  platformclientv2.Bool(true),
			MaxOwnedCallbackHours:      platformclientv2.Int(1),
			MaxOwnedCallbackDelayHours: platformclientv2.Int(2),
		}

		// Routing Rules
		rules = []platformclientv2.Routingrule{
			{
				Operator:    platformclientv2.String("MEETS_THRESHOLD"),
				Threshold:   platformclientv2.Int(9),
				WaitSeconds: platformclientv2.Float64(300),
			},
		}

		// Media Settings
		call     = generateMediaSettings()
		callback = generateCallbackMediaSettings()
		chat     = generateMediaSettings()
		email    = generateMediaEmailSettings()
		message  = GenerateMediaSettingsMessageWithSubType()

		mediaSettings = platformclientv2.Queuemediasettings{
			Call:     &call,
			Callback: &callback,
			Chat:     &chat,
			Email:    &email,
			Message:  &message,
		}

		// Default Scripts
		sId = uuid.NewString()

		script = platformclientv2.Script{
			Id: &sId,
		}

		defaultScripts = map[string]platformclientv2.Script{
			"script1": script,
		}
		libraryIds              = []string{"ABC", "XYZ"}
		cannedResponseLibraries = platformclientv2.Cannedresponselibraries{
			Mode:       platformclientv2.String("SelectedOnly"),
			LibraryIds: &libraryIds,
		}
	)

	return platformclientv2.Createqueuerequest{
		Id:                           &id,
		Name:                         &name,
		Description:                  &description,
		ScoringMethod:                &scoringMethod,
		SkillEvaluationMethod:        &skillEvaluationMethod,
		AutoAnswerOnly:               platformclientv2.Bool(true),
		EnableTranscription:          platformclientv2.Bool(true),
		EnableAudioMonitoring:        platformclientv2.Bool(true),
		EnableManualAssignment:       platformclientv2.Bool(true),
		CallingPartyName:             &callingPartyName,
		CallingPartyNumber:           &callingPartyNumber,
		PeerId:                       &peerId,
		SourceQueueId:                &sourceQueueId,
		AcwSettings:                  &acwSettings,
		SuppressInQueueCallRecording: platformclientv2.Bool(true),
		DirectRouting:                &directRouting,
		AgentOwnedRouting:            &agentOwnedRouting,
		RoutingRules:                 &rules,
		MediaSettings:                &mediaSettings,
		QueueFlow:                    &queueFlow,
		EmailInQueueFlow:             &emailFlow,
		MessageInQueueFlow:           &messageFlow,
		WhisperPrompt:                &whisperPrompt,
		OnHoldPrompt:                 &onHoldPrompt,
		DefaultScripts:               &defaultScripts,
		OutboundMessagingAddresses:   &messagingAddress,
		CannedResponseLibraries:      &cannedResponseLibraries,
		LastAgentRoutingMode:         &lastAgentRoutingMode,
	}
}

func convertCreateQueuetoQueue(req platformclientv2.Createqueuerequest) *platformclientv2.Queue {
	return &platformclientv2.Queue{
		Id:                           req.Id,
		Name:                         req.Name,
		Description:                  req.Description,
		ScoringMethod:                req.ScoringMethod,
		SkillEvaluationMethod:        req.SkillEvaluationMethod,
		AutoAnswerOnly:               req.AutoAnswerOnly,
		EnableTranscription:          req.EnableTranscription,
		EnableAudioMonitoring:        req.EnableAudioMonitoring,
		EnableManualAssignment:       req.EnableManualAssignment,
		CallingPartyName:             req.CallingPartyName,
		CallingPartyNumber:           req.CallingPartyNumber,
		PeerId:                       req.PeerId,
		AcwSettings:                  req.AcwSettings,
		OutboundMessagingAddresses:   req.OutboundMessagingAddresses,
		SuppressInQueueCallRecording: req.SuppressInQueueCallRecording,
		DirectRouting:                req.DirectRouting,
		AgentOwnedRouting:            req.AgentOwnedRouting,
		RoutingRules:                 req.RoutingRules,
		MediaSettings:                req.MediaSettings,
		QueueFlow:                    req.QueueFlow,
		EmailInQueueFlow:             req.EmailInQueueFlow,
		MessageInQueueFlow:           req.MessageInQueueFlow,
		WhisperPrompt:                req.WhisperPrompt,
		OnHoldPrompt:                 req.OnHoldPrompt,
		DefaultScripts:               req.DefaultScripts,
		CannedResponseLibraries:      req.CannedResponseLibraries,
		LastAgentRoutingMode:         req.LastAgentRoutingMode,
	}
}

func generateMediaSettings() platformclientv2.Mediasettings {
	return platformclientv2.Mediasettings{
		EnableAutoAnswer:       platformclientv2.Bool(true),
		AlertingTimeoutSeconds: platformclientv2.Int(20),
		ServiceLevel: &platformclientv2.Servicelevel{
			Percentage: platformclientv2.Float64(0.7),
			DurationMs: platformclientv2.Int(10000),
		},
	}
}

func GenerateMediaSettingsMessageWithSubType() platformclientv2.Messagemediasettings {
	subTypeMap := make(map[string]platformclientv2.Basemediasettings)
	baseMediaSettings := platformclientv2.Basemediasettings{
		EnableAutoAnswer: platformclientv2.Bool(true),
	}
	subTypeMap["instagram"] = baseMediaSettings
	return platformclientv2.Messagemediasettings{
		EnableAutoAnswer:       platformclientv2.Bool(true),
		AlertingTimeoutSeconds: platformclientv2.Int(20),
		ServiceLevel: &platformclientv2.Servicelevel{
			Percentage: platformclientv2.Float64(0.7),
			DurationMs: platformclientv2.Int(10000),
		},
		SubTypeSettings: &subTypeMap,
	}
}

func generateMediaEmailSettings() platformclientv2.Emailmediasettings {
	return platformclientv2.Emailmediasettings{
		EnableAutoAnswer:       platformclientv2.Bool(true),
		AlertingTimeoutSeconds: platformclientv2.Int(20),
		ServiceLevel: &platformclientv2.Servicelevel{
			Percentage: platformclientv2.Float64(0.7),
			DurationMs: platformclientv2.Int(10000),
		},
	}
}

func generateCallbackMediaSettings() platformclientv2.Callbackmediasettings {
	return platformclientv2.Callbackmediasettings{
		EnableAutoAnswer:       platformclientv2.Bool(true),
		AlertingTimeoutSeconds: platformclientv2.Int(20),
		ServiceLevel: &platformclientv2.Servicelevel{
			Percentage: platformclientv2.Float64(0.7),
			DurationMs: platformclientv2.Int(10000),
		},
		EnableAutoDialAndEnd:         platformclientv2.Bool(true),
		AutoDialDelaySeconds:         platformclientv2.Int(10),
		AutoEndDelaySeconds:          platformclientv2.Int(10),
		Mode:                         platformclientv2.String("AgentFirst"),
		AutoAnswerAlertToneSeconds:   platformclientv2.Float64(12),
		ManualAnswerAlertToneSeconds: platformclientv2.Float64(10),
	}
}

func generateDirectRoutingMediaSettings() platformclientv2.Directroutingmediasettings {
	return platformclientv2.Directroutingmediasettings{
		UseAgentAddressOutbound: platformclientv2.Bool(false),
	}
}

func generateRandomDomainEntityRef() platformclientv2.Domainentityref {
	id := uuid.NewString()
	return platformclientv2.Domainentityref{
		Id: &id,
	}
}
