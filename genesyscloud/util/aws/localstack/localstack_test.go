package localstack

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestGetLocalStackPort(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "returns default when env var not set",
			envValue: "",
			expected: defaultLocalStackPort,
		},
		{
			name:     "returns env var value when set",
			envValue: "8080",
			expected: "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment after test
			defer os.Unsetenv(localStackPortEnvVar)

			if tt.envValue != "" {
				os.Setenv(localStackPortEnvVar, tt.envValue)
			} else {
				os.Unsetenv(localStackPortEnvVar)
			}

			result := GetLocalStackPort()
			if result != tt.expected {
				t.Errorf("GetLocalStackPort() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSetLocalStackPort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected string
	}{
		{
			name:     "sets port correctly",
			port:     "9090",
			expected: "9090",
		},
		{
			name:     "sets empty port",
			port:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment after test
			defer os.Unsetenv(localStackPortEnvVar)

			err := setLocalStackPort(tt.port)
			if err != nil {
				t.Errorf("setLocalStackPort() error = %v", err)
				return
			}

			result := os.Getenv(localStackPortEnvVar)
			if result != tt.expected {
				t.Errorf("Environment variable = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetECRLoginPassword(t *testing.T) {
	// This function requires AWS ECR client, so we'll test the error handling
	// by passing a nil client
	t.Run("handles nil ECR client", func(t *testing.T) {
		ctx := context.Background()
		_, err := getECRLoginPasswork(ctx, nil)
		if err == nil {
			t.Error("Expected error when passing nil ECR client")
		}
	})
}

func TestLocalStackManager_ConfigureLocalStackSettings(t *testing.T) {
	tests := []struct {
		name           string
		containerName  string
		image          string
		port           string
		expectedConfig LocalStackManager
	}{
		{
			name:          "all parameters provided",
			containerName: "test-container",
			image:         "test-image:latest",
			port:          "8080",
			expectedConfig: LocalStackManager{
				containerName: "test-container",
				imageURI:      "test-image:latest",
				port:          "8080",
				endpoint:      "http://localhost:8080",
			},
		},
		{
			name:          "empty container name uses default",
			containerName: "",
			image:         "custom-image:latest",
			port:          "9090",
			expectedConfig: LocalStackManager{
				containerName: defaultLocalStackContainerName,
				imageURI:      "custom-image:latest",
				port:          "9090",
				endpoint:      "http://localhost:9090",
			},
		},
		{
			name:          "empty image uses default",
			containerName: "custom-container",
			image:         "",
			port:          "7070",
			expectedConfig: LocalStackManager{
				containerName: "custom-container",
				imageURI:      defaultLocalStackImage,
				port:          "7070",
				endpoint:      "http://localhost:7070",
			},
		},
		{
			name:          "empty port uses default",
			containerName: "custom-container",
			image:         "custom-image:latest",
			port:          "",
			expectedConfig: LocalStackManager{
				containerName: "custom-container",
				imageURI:      "custom-image:latest",
				port:          defaultLocalStackPort,
				endpoint:      "http://localhost:" + defaultLocalStackPort,
			},
		},
		{
			name:          "all empty parameters use defaults",
			containerName: "",
			image:         "",
			port:          "",
			expectedConfig: LocalStackManager{
				containerName: defaultLocalStackContainerName,
				imageURI:      defaultLocalStackImage,
				port:          defaultLocalStackPort,
				endpoint:      "http://localhost:" + defaultLocalStackPort,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &LocalStackManager{}

			// Call the method
			manager.configureLocalStackSettings(tt.containerName, tt.image, tt.port)

			// Verify the configuration
			if manager.containerName != tt.expectedConfig.containerName {
				t.Errorf("containerName = %v, want %v", manager.containerName, tt.expectedConfig.containerName)
			}

			if manager.imageURI != tt.expectedConfig.imageURI {
				t.Errorf("imageURI = %v, want %v", manager.imageURI, tt.expectedConfig.imageURI)
			}

			if manager.port != tt.expectedConfig.port {
				t.Errorf("port = %v, want %v", manager.port, tt.expectedConfig.port)
			}

			if manager.endpoint != tt.expectedConfig.endpoint {
				t.Errorf("endpoint = %v, want %v", manager.endpoint, tt.expectedConfig.endpoint)
			}
		})
	}
}

func TestLocalStackManager_ConfigureLocalStackSettings_EndpointFormat(t *testing.T) {
	t.Run("endpoint format is correct", func(t *testing.T) {
		manager := &LocalStackManager{}

		// Test with a specific port
		testPort := "4567"
		manager.configureLocalStackSettings("test-container", "test-image", testPort)

		expectedEndpoint := "http://localhost:" + testPort
		if manager.endpoint != expectedEndpoint {
			t.Errorf("endpoint = %v, want %v", manager.endpoint, expectedEndpoint)
		}

		// Verify the endpoint format is always http://localhost:port
		if !strings.HasPrefix(manager.endpoint, "http://localhost:") {
			t.Errorf("endpoint should start with 'http://localhost:', got %v", manager.endpoint)
		}
	})
}
