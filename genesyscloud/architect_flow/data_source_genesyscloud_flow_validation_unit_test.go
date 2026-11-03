package architect_flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
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

func TestUnitRunFlowValidationContextDeadlineIsTimeout(t *testing.T) {
	originalInterval := validateJobPollInterval
	validateJobPollInterval = 50 * time.Millisecond
	defer func() { validateJobPollInterval = originalInterval }()

	p := &architectFlowProxy{
		createFlowValidateJobAttr: func(ctx context.Context, a *architectFlowProxy, flow platformclientv2.Architectflowreference) (*platformclientv2.Registerarchitectvalidatejobresponse, *platformclientv2.APIResponse, error) {
			return &platformclientv2.Registerarchitectvalidatejobresponse{Id: platformclientv2.String("job-ctx")}, nil, nil
		},
		getFlowValidateJobAttr: func(ctx context.Context, a *architectFlowProxy, id string) (*platformclientv2.Architectvalidatejobstateresponse, *platformclientv2.APIResponse, error) {
			return &platformclientv2.Architectvalidatejobstateresponse{Status: platformclientv2.String("Started")}, nil, nil
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, _, err := runFlowValidation(ctx, p, buildValidateFlowReference("flow-1", "", "", ""), time.Hour)
	var timeoutErr *validateJobTimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("expected validateJobTimeoutError, got %v", err)
	}
	if timeoutErr.JobId != "job-ctx" {
		t.Errorf("expected job ID job-ctx, got %s", timeoutErr.JobId)
	}
}

// diagnosticCodeAndJobID extracts the code and jobId fields from a diagnostic's JSON detail.
func diagnosticCodeAndJobID(t *testing.T, d diag.Diagnostic) (string, string) {
	t.Helper()
	var detail struct {
		Code  string `json:"code"`
		JobID string `json:"jobId"`
	}
	if err := json.Unmarshal([]byte(d.Detail), &detail); err != nil {
		t.Fatalf("diagnostic detail is not JSON: %v (%s)", err, d.Detail)
	}
	return detail.Code, detail.JobID
}

func TestUnitBuildFlowValidationErrorDiagnostics(t *testing.T) {
	t.Run("timeout is coded with job ID", func(t *testing.T) {
		err := fmt.Errorf("wrapped: %w", &validateJobTimeoutError{JobId: "job-t", Timeout: time.Minute, LastStatus: "Started"})
		diags := buildFlowValidationErrorDiagnostics(err, "flow-1", nil)
		if len(diags) != 1 || diags[0].Severity != diag.Error {
			t.Fatalf("expected one error diagnostic, got %+v", diags)
		}
		code, jobId := diagnosticCodeAndJobID(t, diags[0])
		if code != util.DiagnosticCodeFlowValidateTimeout || jobId != "job-t" {
			t.Errorf("expected code %s and job ID job-t, got %q / %q", util.DiagnosticCodeFlowValidateTimeout, code, jobId)
		}
	})

	t.Run("other errors are not coded", func(t *testing.T) {
		diags := buildFlowValidationErrorDiagnostics(fmt.Errorf("boom"), "flow-1", nil)
		if len(diags) != 1 {
			t.Fatalf("expected one diagnostic, got %+v", diags)
		}
		if code, _ := diagnosticCodeAndJobID(t, diags[0]); code != "" {
			t.Errorf("expected no code, got %q", code)
		}
	})
}

func TestUnitBuildFlowValidationResultDiagnostics(t *testing.T) {
	failed := &flowValidationResult{JobId: "job-f", Status: "Failure", Errors: []string{"e1", "e2"}, Warnings: []string{"w1"}}
	passed := &flowValidationResult{JobId: "job-s", Status: "Success", Errors: []string{}, Warnings: []string{"w1"}}

	t.Run("failure with fail_on_error is a coded error plus warnings", func(t *testing.T) {
		diags := buildFlowValidationResultDiagnostics(failed, "flow-1", true, nil)
		if len(diags) != 2 || diags[0].Severity != diag.Warning || diags[1].Severity != diag.Error {
			t.Fatalf("expected a warning then an error, got %+v", diags)
		}
		code, jobId := diagnosticCodeAndJobID(t, diags[1])
		if code != util.DiagnosticCodeFlowValidateFailed || jobId != "job-f" {
			t.Errorf("expected code %s and job ID job-f, got %q / %q", util.DiagnosticCodeFlowValidateFailed, code, jobId)
		}
		if !strings.Contains(diags[1].Summary, "e1") || !strings.Contains(diags[1].Summary, "e2") {
			t.Errorf("expected error messages in summary, got %q", diags[1].Summary)
		}
	})

	t.Run("failure without fail_on_error only returns warnings", func(t *testing.T) {
		diags := buildFlowValidationResultDiagnostics(failed, "flow-1", false, nil)
		if diags.HasError() || len(diags) != 1 {
			t.Errorf("expected only one warning, got %+v", diags)
		}
	})

	t.Run("success with fail_on_error only returns warnings", func(t *testing.T) {
		diags := buildFlowValidationResultDiagnostics(passed, "flow-1", true, nil)
		if diags.HasError() || len(diags) != 1 {
			t.Errorf("expected only one warning, got %+v", diags)
		}
	})
}
