package sdk_config

import (
	"errors"
	"log"
	"sync"

	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

/*
	This package exists to address a cyclic issue between the provider package and resource packages defined with the TF framework.
	All framework resources should be registered with the framework provider in the *GenesysCloudProvider.Resources method inside the provider package,
	while the resource Configure (for now just the Configure method inside genesyscloud/routing_wrapupcode_v2) method wants to access the SDK Configuration in the provider package.

	As a workaround, the provider Configure method will set the SDK config here and the resource Configure method will collect it from here.
	This method is too simple because it ignores all SdkClientPool logic, but it will do for now just to get a framework resource working.
*/

var (
	frameworkConfig *platformclientv2.Configuration
	mutex           sync.RWMutex
	configSet       bool
)

var (
	ErrConfigNotSet = errors.New("framework SDK configuration has not been set")
	ErrConfigNil    = errors.New("framework SDK configuration is nil")
)

// SetConfig sets the framework SDK configuration
func SetConfig(c *platformclientv2.Configuration) error {
	if c == nil {
		return ErrConfigNil
	}

	mutex.Lock()
	defer mutex.Unlock()

	frameworkConfig = c
	configSet = true
	return nil
}

// GetConfig returns the framework SDK configuration
func GetConfig() (*platformclientv2.Configuration, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	if !configSet {
		return nil, ErrConfigNotSet
	}

	if frameworkConfig == nil {
		return nil, ErrConfigNil
	}

	return frameworkConfig, nil
}

// GetConfigOrDefault returns the framework SDK configuration or a default configuration if not set
func GetConfigOrDefault() *platformclientv2.Configuration {
	config, err := GetConfig()
	if err != nil {
		log.Printf("[WARN] %v. Returning a default configuration.", err)
		return platformclientv2.GetDefaultConfiguration()
	}
	return config
}

// IsConfigSet returns true if the configuration has been set
func IsConfigSet() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return configSet
}

// Reset clears the framework SDK configuration
func Reset() {
	mutex.Lock()
	defer mutex.Unlock()
	frameworkConfig = nil
	configSet = false
}
