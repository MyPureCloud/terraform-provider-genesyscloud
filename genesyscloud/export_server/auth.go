package export_server

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	AccessToken       string
	OAuthClientID     string
	OAuthClientSecret string
	Region            string
}

// GetAuthConfig extracts authentication information from request headers and environment variables
func GetAuthConfig(r *http.Request) (*AuthConfig, error) {
	// First, try to get bearer token from Authorization header
	authHeader := r.Header.Get("Authorization")
	var accessToken string

	if authHeader != "" {
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return nil, fmt.Errorf("invalid authorization header format")
		}
		accessToken = strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Get OAuth credentials from environment variables
	oauthClientID := os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID")
	oauthClientSecret := os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET")
	region := os.Getenv("GENESYSCLOUD_REGION")

	// If no access token provided, we need OAuth credentials
	if accessToken == "" {
		if oauthClientID == "" || oauthClientSecret == "" || region == "" {
			return nil, fmt.Errorf("missing required environment variables: GENESYSCLOUD_OAUTHCLIENT_ID, GENESYSCLOUD_OAUTHCLIENT_SECRET, GENESYSCLOUD_REGION")
		}
	}

	return &AuthConfig{
		AccessToken:       accessToken,
		OAuthClientID:     oauthClientID,
		OAuthClientSecret: oauthClientSecret,
		Region:            region,
	}, nil
}

// ValidateAuth validates the authentication configuration and returns an SDK config
func ValidateAuth(authConfig *AuthConfig) (*platformclientv2.Configuration, error) {
	// Create SDK configuration
	sdkConfig := platformclientv2.GetDefaultConfiguration()

	// Set the base path using the existing provider function
	basePath := provider.GetRegionBasePath(authConfig.Region)
	sdkConfig.BasePath = basePath

	if authConfig.AccessToken != "" {
		// Use access token if provided
		sdkConfig.AccessToken = authConfig.AccessToken
	} else {
		// Use OAuth client credentials
		err := sdkConfig.AuthorizeClientCredentials(authConfig.OAuthClientID, authConfig.OAuthClientSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to authorize with client credentials: %w", err)
		}
	}

	return sdkConfig, nil
}

// AuthenticateRequest validates authentication for an HTTP request
func AuthenticateRequest(r *http.Request) (*platformclientv2.Configuration, error) {
	authConfig, err := GetAuthConfig(r)
	if err != nil {
		return nil, err
	}

	return ValidateAuth(authConfig)
}
