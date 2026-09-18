package architect_flow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
)

func TestUnitGenerateDownloadUrlFn(t *testing.T) {
	const mockFlowId = "mock-id"
	const mockVersion = "1.0"

	tests := []struct {
		name                string
		createJobFunc       func(*architectFlowProxy, string, string) (*platformclientv2.Registerarchitectexportjobresponse, *platformclientv2.APIResponse, error)
		pollDownloadUrlFunc func(*architectFlowProxy, string, float64) (string, error)
		expectedError       string
		expectedUrl         string
	}{
		{
			name: "Should fail when createExportJob returns error",
			createJobFunc: func(proxy *architectFlowProxy, id, version string) (*platformclientv2.Registerarchitectexportjobresponse, *platformclientv2.APIResponse, error) {
				return nil, nil, fmt.Errorf("mock create error")
			},
			expectedError: "mock create error",
		},
		{
			name: "Should fail when export job response is nil",
			createJobFunc: func(proxy *architectFlowProxy, id, version string) (*platformclientv2.Registerarchitectexportjobresponse, *platformclientv2.APIResponse, error) {
				return nil, nil, nil
			},
			expectedError: "no export job flow ID returned for flow " + mockFlowId,
		},
		{
			name: "Should fail when export job ID is nil",
			createJobFunc: func(proxy *architectFlowProxy, id, version string) (*platformclientv2.Registerarchitectexportjobresponse, *platformclientv2.APIResponse, error) {
				return &platformclientv2.Registerarchitectexportjobresponse{
					Id: nil,
				}, nil, nil
			},
			expectedError: "no export job flow ID returned for flow " + mockFlowId,
		},
		{
			name: "Should fail when polling for download URL fails",
			createJobFunc: func(proxy *architectFlowProxy, id, version string) (*platformclientv2.Registerarchitectexportjobresponse, *platformclientv2.APIResponse, error) {
				return &platformclientv2.Registerarchitectexportjobresponse{
					Id: platformclientv2.String("mock-id"),
				}, nil, nil
			},
			pollDownloadUrlFunc: func(a *architectFlowProxy, jobId string, timeout float64) (string, error) {
				return "", fmt.Errorf("mock poll error")
			},
			expectedError: "mock poll error",
		},
		{
			name: "Should succeed with valid download URL",
			createJobFunc: func(proxy *architectFlowProxy, id, version string) (*platformclientv2.Registerarchitectexportjobresponse, *platformclientv2.APIResponse, error) {
				return &platformclientv2.Registerarchitectexportjobresponse{
					Id: platformclientv2.String(mockFlowId),
				}, nil, nil
			},
			pollDownloadUrlFunc: func(a *architectFlowProxy, jobId string, timeout float64) (string, error) {
				return "https://example.com/download", nil
			},
			expectedUrl: "https://example.com/download",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up proxy instance with test case functions
			proxyInstance := &architectFlowProxy{
				createExportJobAttr: tt.createJobFunc,
			}

			// Set poll function if provided in test case
			if tt.pollDownloadUrlFunc != nil {
				proxyInstance.pollExportJobForDownloadUrlAttr = tt.pollDownloadUrlFunc
			}

			// Execute function being tested
			url, err := generateDownloadUrlFn(proxyInstance, mockFlowId, mockVersion)

			// Assert results
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			// Assert success case
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if url != tt.expectedUrl {
				t.Errorf("expected URL %q, got %q", tt.expectedUrl, url)
			}
		})
	}
}

func TestUnitPollExportJobForDownloadUrlFn(t *testing.T) {
	tests := []struct {
		name             string
		jobId            string
		timeoutSeconds   float64
		getJobStatusFunc func(*architectFlowProxy, string) (*platformclientv2.Architectexportjobstateresponse, *platformclientv2.APIResponse, error)
		expectedError    string
		expectedUrl      string
	}{
		{
			name:           "Should timeout after specified duration",
			jobId:          "timeout-job",
			timeoutSeconds: 0.1,
			getJobStatusFunc: func(a *architectFlowProxy, jobId string) (*platformclientv2.Architectexportjobstateresponse, *platformclientv2.APIResponse, error) {
				return &platformclientv2.Architectexportjobstateresponse{
					Status: platformclientv2.String("Started"),
				}, nil, nil
			},
			expectedError: "timed out after",
		},
		{
			name:           "Should return error when getExportJobStatusById fails",
			jobId:          "error-job",
			timeoutSeconds: 5,
			getJobStatusFunc: func(a *architectFlowProxy, jobId string) (*platformclientv2.Architectexportjobstateresponse, *platformclientv2.APIResponse, error) {
				return nil, nil, fmt.Errorf("API error")
			},
			expectedError: "API error",
		},
		{
			name:           "Should return error for failed job status",
			jobId:          "failed-job",
			timeoutSeconds: 5,
			getJobStatusFunc: func(a *architectFlowProxy, jobId string) (*platformclientv2.Architectexportjobstateresponse, *platformclientv2.APIResponse, error) {
				return &platformclientv2.Architectexportjobstateresponse{
					Status: platformclientv2.String("Failure"),
					Messages: &[]platformclientv2.Architectjobmessage{
						{
							Text: platformclientv2.String("mock message text"),
						},
					},
				}, nil, nil
			},
			expectedError: "mock message text",
		},
		{
			name:           "Should return error for unexpected job status",
			jobId:          "unexpected-status-job",
			timeoutSeconds: 5,
			getJobStatusFunc: func(a *architectFlowProxy, jobId string) (*platformclientv2.Architectexportjobstateresponse, *platformclientv2.APIResponse, error) {
				return &platformclientv2.Architectexportjobstateresponse{
					Status: platformclientv2.String("Unknown"),
				}, nil, nil
			},
			expectedError: "unexpected job status Unknown",
		},
		{
			name:           "Should return error when download URL is nil",
			jobId:          "nil-url-job",
			timeoutSeconds: 5,
			getJobStatusFunc: func(a *architectFlowProxy, jobId string) (*platformclientv2.Architectexportjobstateresponse, *platformclientv2.APIResponse, error) {
				return &platformclientv2.Architectexportjobstateresponse{
					Status:      platformclientv2.String("Success"),
					DownloadUrl: nil,
				}, nil, nil
			},
			expectedError: "was a success but no download ID was returned",
		},
		{
			name:           "Should return download URL on success",
			jobId:          "success-job",
			timeoutSeconds: 5,
			getJobStatusFunc: func(a *architectFlowProxy, jobId string) (*platformclientv2.Architectexportjobstateresponse, *platformclientv2.APIResponse, error) {
				return &platformclientv2.Architectexportjobstateresponse{
					Status:      platformclientv2.String("Success"),
					DownloadUrl: platformclientv2.String("https://example.com/download"),
				}, nil, nil
			},
			expectedUrl: "https://example.com/download",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create proxy instance with mock function
			proxyInstance := &architectFlowProxy{
				getExportJobStatusByIdAttr: tt.getJobStatusFunc,
			}

			// Execute function being tested
			url, err := pollExportJobForDownloadUrlFn(proxyInstance, tt.jobId, tt.timeoutSeconds)

			// Assert results
			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			// Assert success case
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if url != tt.expectedUrl {
				t.Errorf("expected URL %q, got %q", tt.expectedUrl, url)
			}
		})
	}
}

func TestUnitBuildRegisterArchitectJobRequest(t *testing.T) {
	tests := []struct {
		name         string
		createStubs  bool
		expectedBody string
	}{
		{
			name:         "Should omit createStubs when false",
			createStubs:  false,
			expectedBody: `{}`,
		},
		{
			name:         "Should include createStubs when true",
			createStubs:  true,
			expectedBody: `{"createStubs":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := buildRegisterArchitectJobRequest(tt.createStubs)

			if tt.createStubs {
				if body.CreateStubs == nil {
					t.Fatal("expected CreateStubs to be set, got nil")
				}
				if !*body.CreateStubs {
					t.Errorf("expected CreateStubs to be true, got %v", *body.CreateStubs)
				}
			} else if body.CreateStubs != nil {
				t.Errorf("expected CreateStubs to be nil, got %v", *body.CreateStubs)
			}

			marshalled, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("unexpected error marshalling request body: %v", err)
			}
			if string(marshalled) != tt.expectedBody {
				t.Errorf("expected POST body %s, got %s", tt.expectedBody, string(marshalled))
			}
		})
	}
}

func TestUnitCreateFlowsDeployJob(t *testing.T) {
	tests := []struct {
		name        string
		createStubs bool
	}{
		{
			name:        "Should forward create_stubs false to the implementation function",
			createStubs: false,
		},
		{
			name:        "Should forward create_stubs true to the implementation function",
			createStubs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedCreateStubs bool

			proxyInstance := &architectFlowProxy{
				createArchitectFlowJobsAttr: func(ctx context.Context, a *architectFlowProxy, createStubs bool) (*platformclientv2.Registerarchitectjobresponse, *platformclientv2.APIResponse, error) {
					receivedCreateStubs = createStubs
					return &platformclientv2.Registerarchitectjobresponse{
						Id: platformclientv2.String("mock-job-id"),
					}, nil, nil
				},
			}

			job, _, err := proxyInstance.CreateFlowsDeployJob(context.Background(), tt.createStubs)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if job == nil || job.Id == nil {
				t.Fatal("expected a job response with an ID")
			}
			if receivedCreateStubs != tt.createStubs {
				t.Errorf("expected createStubs %v to be passed to the implementation function, got %v", tt.createStubs, receivedCreateStubs)
			}
		})
	}
}
