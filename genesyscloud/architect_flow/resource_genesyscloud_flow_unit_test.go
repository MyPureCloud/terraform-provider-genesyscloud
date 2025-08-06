package architect_flow

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestUnitGenerateDownloadUrlFn(t *testing.T) {
	const mockFlowId = "mock-id"

	tests := []struct {
		name                string
		createJobFunc       func(*architectFlowProxy, string) (*platformclientv2.Registerarchitectexportjobresponse, *platformclientv2.APIResponse, error)
		pollDownloadUrlFunc func(*architectFlowProxy, string, float64) (string, error)
		expectedError       string
		expectedUrl         string
	}{
		{
			name: "Should fail when createExportJob returns error",
			createJobFunc: func(proxy *architectFlowProxy, id string) (*platformclientv2.Registerarchitectexportjobresponse, *platformclientv2.APIResponse, error) {
				return nil, nil, fmt.Errorf("mock create error")
			},
			expectedError: "mock create error",
		},
		{
			name: "Should fail when export job response is nil",
			createJobFunc: func(proxy *architectFlowProxy, id string) (*platformclientv2.Registerarchitectexportjobresponse, *platformclientv2.APIResponse, error) {
				return nil, nil, nil
			},
			expectedError: "no export job flow ID returned for flow " + mockFlowId,
		},
		{
			name: "Should fail when export job ID is nil",
			createJobFunc: func(proxy *architectFlowProxy, id string) (*platformclientv2.Registerarchitectexportjobresponse, *platformclientv2.APIResponse, error) {
				return &platformclientv2.Registerarchitectexportjobresponse{
					Id: nil,
				}, nil, nil
			},
			expectedError: "no export job flow ID returned for flow " + mockFlowId,
		},
		{
			name: "Should fail when polling for download URL fails",
			createJobFunc: func(proxy *architectFlowProxy, id string) (*platformclientv2.Registerarchitectexportjobresponse, *platformclientv2.APIResponse, error) {
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
			createJobFunc: func(proxy *architectFlowProxy, id string) (*platformclientv2.Registerarchitectexportjobresponse, *platformclientv2.APIResponse, error) {
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
			url, err := generateDownloadUrlFn(proxyInstance, mockFlowId)

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
