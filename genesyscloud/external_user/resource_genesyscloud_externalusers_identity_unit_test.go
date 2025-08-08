package external_user

import (
	"context"
	"fmt"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestGetAllHelperExternalUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		userList       *[]platformclientv2.User
		mockSetup      func(*externalUserIdentityProxy)
		expectedError  bool
		expectedLength int
	}{
		{
			name:           "Nil user list",
			userList:       nil,
			mockSetup:      func(e *externalUserIdentityProxy) {},
			expectedError:  false,
			expectedLength: 0,
		},
		{
			name:           "Empty user list",
			userList:       &[]platformclientv2.User{},
			mockSetup:      func(e *externalUserIdentityProxy) {},
			expectedError:  false,
			expectedLength: 0,
		},
		{
			name: "User with nil ID",
			userList: &[]platformclientv2.User{
				{
					Id: nil,
				},
			},
			mockSetup:      func(e *externalUserIdentityProxy) {},
			expectedError:  false,
			expectedLength: 0,
		},
		{
			name: "Success case",
			userList: &[]platformclientv2.User{
				{
					Id: platformclientv2.String("user1"),
				},
				{
					Id: platformclientv2.String("user2"),
				}, {
					Id: platformclientv2.String("user3"),
				},
			},
			mockSetup: func(e *externalUserIdentityProxy) {
				e.getAllExternalUserIdentityAttr = func(ctx context.Context, p *externalUserIdentityProxy, userId string) (*[]platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error) {
					authorityNameOne := "auth1"
					externalKeyOne := "key1"
					authorityNameTwo := "auth2"
					externalKeyTwo := "key2"
					return &[]platformclientv2.Userexternalidentifier{
						{
							AuthorityName: &authorityNameOne,
							ExternalKey:   &externalKeyOne,
						}, {
							AuthorityName: &authorityNameTwo,
							ExternalKey:   &externalKeyTwo,
						},
					}, nil, nil
				}
			},
			expectedError:  false,
			expectedLength: 6,
		},
		{
			name: "External user API error",
			userList: &[]platformclientv2.User{
				{
					Id: platformclientv2.String("user1"),
				},
			},
			mockSetup: func(e *externalUserIdentityProxy) {
				e.getAllExternalUserIdentityAttr = func(ctx context.Context, p *externalUserIdentityProxy, userId string) (*[]platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error) {
					return nil, &platformclientv2.APIResponse{}, fmt.Errorf("API error")
				}
			},
			expectedError:  true,
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock proxy
			proxy := &externalUserIdentityProxy{}
			tt.mockSetup(proxy)

			// Call function
			resources, _, err := getAllHelperExternalUser(ctx, proxy, tt.userList)

			// Check error
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			// Check resources length
			assert.Equal(t, tt.expectedLength, len(resources))

			// For success case, verify the compound key
			if tt.name == "Success case" {
				expectedKey := createCompoundKey("user1", "auth1", "key1")
				_, exists := resources[expectedKey]
				assert.True(t, exists)
				assert.Equal(t, expectedKey, resources[expectedKey].BlockLabel)
			}
		})
	}
}
