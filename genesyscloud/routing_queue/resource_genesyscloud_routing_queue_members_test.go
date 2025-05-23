package routing_queue

import (
	"context"
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitUpdateQueueUserRingNum(t *testing.T) {

	// Minimize the number of retries
	previousRetries := util.SetMaxRetriesForTests(3)
	defer util.SetMaxRetriesForTests(previousRetries)

	// Test cases
	tests := []struct {
		name          string
		queueID       string
		userID        string
		ringNum       int
		mockResponses []mockResponse
		expectedError bool
		expectedCalls int
	}{
		{
			name:    "successful_update",
			queueID: "queue-123",
			userID:  "user-456",
			ringNum: 2,
			mockResponses: []mockResponse{
				{
					resp: &platformclientv2.APIResponse{StatusCode: 200},
					err:  nil,
				},
			},
			expectedError: false,
			expectedCalls: 1,
		},
		{
			name:    "retry_success_after_404",
			queueID: "queue-123",
			userID:  "user-456",
			ringNum: 3,
			mockResponses: []mockResponse{
				{
					resp: &platformclientv2.APIResponse{StatusCode: 404},
					err:  fmt.Errorf("not found"),
				},
				{
					resp: &platformclientv2.APIResponse{StatusCode: 200},
					err:  nil,
				},
				// This response should not be called but it's added to ensure that the response returns immediately after a 200
				{
					resp: &platformclientv2.APIResponse{StatusCode: 200},
					err:  nil,
				},
			},
			expectedError: false,
			expectedCalls: 2,
		},
		{
			name:    "permanent_failure",
			queueID: "queue-123",
			userID:  "user-456",
			ringNum: 4,
			mockResponses: []mockResponse{
				{
					resp: &platformclientv2.APIResponse{StatusCode: 500},
					err:  fmt.Errorf("internal server error"),
				},
				// This response should not be called but it's added to ensure that the response returns immediately after a 500
				{
					resp: &platformclientv2.APIResponse{StatusCode: 404},
					err:  fmt.Errorf("not found"),
				},
			},
			expectedError: true,
			expectedCalls: 1,
		},
		{
			name:    "max_retries_exceeded",
			queueID: "queue-123",
			userID:  "user-456",
			ringNum: 5,
			mockResponses: []mockResponse{
				{
					resp: &platformclientv2.APIResponse{StatusCode: 404},
					err:  fmt.Errorf("not found"),
				},
				{
					resp: &platformclientv2.APIResponse{StatusCode: 404},
					err:  fmt.Errorf("not found"),
				},
				{
					resp: &platformclientv2.APIResponse{StatusCode: 404},
					err:  fmt.Errorf("not found"),
				},
				// This response should not be called but it's added to ensure that the response returns immediately after the max retries 404
				{
					resp: &platformclientv2.APIResponse{StatusCode: 404},
					err:  fmt.Errorf("not found"),
				},
			},
			expectedError: true,
			expectedCalls: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track number of calls
			callCount := 0
			currentResponse := 0

			// Create mock proxy
			queueProxy := RoutingQueueProxy{}
			queueProxy.updateRoutingQueueMemberAttr = func(ctx context.Context, p *RoutingQueueProxy, queueID string, userID string, body platformclientv2.Queuemember) (*platformclientv2.APIResponse, error) {
				// Verify parameters
				assert.Equal(t, tt.queueID, queueID)
				assert.Equal(t, tt.userID, userID)
				assert.Equal(t, tt.userID, *body.Id)
				assert.Equal(t, tt.ringNum, *body.RingNumber)

				callCount++

				// Return the appropriate response
				if currentResponse < len(tt.mockResponses) {
					resp := tt.mockResponses[currentResponse]
					currentResponse++
					return resp.resp, resp.err
				}

				return &platformclientv2.APIResponse{StatusCode: 500}, fmt.Errorf("unexpected call")
			}

			err := setRoutingQueueUnitTestsEnvVar()
			if err != nil {
				t.Skipf("failed to set env variable %s: %s", unitTestsAreActiveEnv, err.Error())
			}

			internalProxy = &queueProxy
			defer func() {
				internalProxy = nil
				err = unsetRoutingQueueUnitTestsEnvVar()
				if err != nil {
					t.Logf("Failed to unset env variable %s: %s", unitTestsAreActiveEnv, err.Error())
				}
			}()

			// Call the function
			diags := updateQueueUserRingNum(tt.queueID, tt.userID, tt.ringNum, &platformclientv2.Configuration{})

			// Assert results
			if tt.expectedError {
				assert.NotNil(t, diags, "Expected error diagnostics")
			} else {
				assert.Nil(t, diags, "Expected no error diagnostics")
			}

			// Verify number of calls
			assert.Equal(t, tt.expectedCalls, callCount, "Unexpected number of calls to updateRoutingQueueMember")
		})
	}
}

type mockResponse struct {
	resp *platformclientv2.APIResponse
	err  error
}
