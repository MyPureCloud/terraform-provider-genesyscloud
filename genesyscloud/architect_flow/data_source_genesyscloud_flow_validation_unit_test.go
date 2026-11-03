package architect_flow

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

func TestUnitBuildValidateFlowReference(t *testing.T) {
	tests := []struct {
		name                              string
		flowId, flowName, flowType, flowV string
		expected                          platformclientv2.Architectflowreference
	}{
		{
			name:     "id only",
			flowId:   "flow-1",
			expected: platformclientv2.Architectflowreference{Id: platformclientv2.String("flow-1")},
		},
		{
			name:     "name and type are sent with type uppercased",
			flowName: "My Flow",
			flowType: "inboundcall",
			flowV:    "2.0",
			expected: platformclientv2.Architectflowreference{
				Name:    platformclientv2.String("My Flow"),
				VarType: platformclientv2.String("INBOUNDCALL"),
				Version: platformclientv2.String("2.0"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildValidateFlowReference(tt.flowId, tt.flowName, tt.flowType, tt.flowV)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("expected %+v, got %+v", tt.expected, got)
			}
		})
	}
}

func TestUnitSplitValidateJobMessages(t *testing.T) {
	msgs := &[]platformclientv2.Architectjobmessage{
		{VarType: platformclientv2.String("Error"), Text: platformclientv2.String("bad action")},
		{VarType: platformclientv2.String("Warning"), Text: platformclientv2.String("unused variable")},
		{VarType: platformclientv2.String("Unknown"), Text: platformclientv2.String("mystery")},
		{VarType: platformclientv2.String("Error")},
	}
	errs, warnings := splitValidateJobMessages(msgs)
	if !reflect.DeepEqual(errs, []string{"bad action", "mystery"}) {
		t.Errorf("unexpected errors: %v", errs)
	}
	if !reflect.DeepEqual(warnings, []string{"unused variable"}) {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	errs, warnings = splitValidateJobMessages(nil)
	if errs == nil || warnings == nil || len(errs) != 0 || len(warnings) != 0 {
		t.Errorf("expected empty non-nil slices for nil messages, got %v / %v", errs, warnings)
	}
}

func TestUnitRunFlowValidation(t *testing.T) {
	originalInterval := validateJobPollInterval
	validateJobPollInterval = time.Millisecond
	defer func() { validateJobPollInterval = originalInterval }()

	const jobId = "job-1"
	okCreate := func(ctx context.Context, p *architectFlowProxy, flow platformclientv2.Architectflowreference) (*platformclientv2.Registerarchitectvalidatejobresponse, *platformclientv2.APIResponse, error) {
		return &platformclientv2.Registerarchitectvalidatejobresponse{Id: platformclientv2.String(jobId)}, nil, nil
	}
	statesThen := func(final *platformclientv2.Architectvalidatejobstateresponse) getFlowValidateJobFunc {
		calls := 0
		return func(ctx context.Context, p *architectFlowProxy, id string) (*platformclientv2.Architectvalidatejobstateresponse, *platformclientv2.APIResponse, error) {
			calls++
			if calls < 3 {
				return &platformclientv2.Architectvalidatejobstateresponse{Id: &id, Status: platformclientv2.String("Started")}, nil, nil
			}
			return final, nil, nil
		}
	}

	tests := []struct {
		name          string
		create        createFlowValidateJobFunc
		get           getFlowValidateJobFunc
		timeout       time.Duration
		expectedErr   string
		expectedValid bool
		expectedErrs  []string
		expectedWarns []string
	}{
		{
			name: "create error is returned",
			create: func(ctx context.Context, p *architectFlowProxy, flow platformclientv2.Architectflowreference) (*platformclientv2.Registerarchitectvalidatejobresponse, *platformclientv2.APIResponse, error) {
				return nil, nil, fmt.Errorf("boom")
			},
			expectedErr: "error registering validate job: boom",
		},
		{
			name: "missing job ID is an error",
			create: func(ctx context.Context, p *architectFlowProxy, flow platformclientv2.Architectflowreference) (*platformclientv2.Registerarchitectvalidatejobresponse, *platformclientv2.APIResponse, error) {
				return &platformclientv2.Registerarchitectvalidatejobresponse{}, nil, nil
			},
			expectedErr: "returned no job ID",
		},
		{
			name:   "status error is returned",
			create: okCreate,
			get: func(ctx context.Context, p *architectFlowProxy, id string) (*platformclientv2.Architectvalidatejobstateresponse, *platformclientv2.APIResponse, error) {
				return nil, nil, fmt.Errorf("not found")
			},
			expectedErr: "error retrieving status of validate job job-1: not found",
		},
		{
			name:   "success with warnings is valid",
			create: okCreate,
			get: statesThen(&platformclientv2.Architectvalidatejobstateresponse{
				Status: platformclientv2.String("Success"),
				Messages: &[]platformclientv2.Architectjobmessage{
					{VarType: platformclientv2.String("Warning"), Text: platformclientv2.String("w1")},
				},
			}),
			expectedValid: true,
			expectedErrs:  []string{},
			expectedWarns: []string{"w1"},
		},
		{
			name:   "failure is invalid and returns errors",
			create: okCreate,
			get: statesThen(&platformclientv2.Architectvalidatejobstateresponse{
				Status: platformclientv2.String("Failure"),
				Messages: &[]platformclientv2.Architectjobmessage{
					{VarType: platformclientv2.String("Error"), Text: platformclientv2.String("e1")},
				},
			}),
			expectedValid: false,
			expectedErrs:  []string{"e1"},
			expectedWarns: []string{},
		},
		{
			name:   "times out while job is still running",
			create: okCreate,
			get: func(ctx context.Context, p *architectFlowProxy, id string) (*platformclientv2.Architectvalidatejobstateresponse, *platformclientv2.APIResponse, error) {
				return &platformclientv2.Architectvalidatejobstateresponse{Status: platformclientv2.String("Started")}, nil, nil
			},
			timeout:     5 * time.Millisecond,
			expectedErr: "timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &architectFlowProxy{createFlowValidateJobAttr: tt.create, getFlowValidateJobAttr: tt.get}
			timeout := tt.timeout
			if timeout == 0 {
				timeout = time.Second
			}
			result, _, err := runFlowValidation(context.Background(), p, buildValidateFlowReference("flow-1", "", "", ""), timeout)

			if tt.expectedErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.expectedErr) {
					t.Fatalf("expected error containing %q, got %v", tt.expectedErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.JobId != jobId {
				t.Errorf("expected job ID %s, got %s", jobId, result.JobId)
			}
			if result.valid() != tt.expectedValid {
				t.Errorf("expected valid=%v, got %v", tt.expectedValid, result.valid())
			}
			if !reflect.DeepEqual(result.Errors, tt.expectedErrs) {
				t.Errorf("expected errors %v, got %v", tt.expectedErrs, result.Errors)
			}
			if !reflect.DeepEqual(result.Warnings, tt.expectedWarns) {
				t.Errorf("expected warnings %v, got %v", tt.expectedWarns, result.Warnings)
			}
		})
	}
}

func TestUnitCreateFlowValidateJobForwardsReference(t *testing.T) {
	var received platformclientv2.Architectflowreference
	p := &architectFlowProxy{
		createFlowValidateJobAttr: func(ctx context.Context, a *architectFlowProxy, flow platformclientv2.Architectflowreference) (*platformclientv2.Registerarchitectvalidatejobresponse, *platformclientv2.APIResponse, error) {
			received = flow
			return &platformclientv2.Registerarchitectvalidatejobresponse{Id: platformclientv2.String("j")}, nil, nil
		},
	}
	ref := buildValidateFlowReference("flow-1", "", "", "3.0")
	if _, _, err := p.CreateFlowValidateJob(context.Background(), ref); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(received, ref) {
		t.Errorf("expected %+v, got %+v", ref, received)
	}
}
