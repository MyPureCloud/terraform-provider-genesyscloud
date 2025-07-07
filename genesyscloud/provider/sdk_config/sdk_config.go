package sdk_config

import (
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
	"log"
)

/*
	This package exists to address a cyclic issue between the provider package and resource packages defined with the TF framework.
	All framework resources should be registered with the framework provider in the *GenesysCloudProvider.Resources method inside the provider package,
	while the resource Configure (for now just the Configure method inside genesyscloud/routing_wrapupcode_v2) method wants to access the SDK Configuration in the provider package.

	As a workaround, the provider Configure method will set the SDK config here and the resource Configure method will collect it from here.
	This method is too simple because it ignores all SdkClientPool logic, but it will do for now just to get a framework resource working.
*/

var frameworkConfig *platformclientv2.Configuration

func SetConfig(c *platformclientv2.Configuration) {
	frameworkConfig = c
}

func GetConfig() *platformclientv2.Configuration {
	if frameworkConfig == nil {
		log.Println("[WARN] framework SDK configuration instance is nil. Returning a default configuration.")
		return platformclientv2.GetDefaultConfiguration()
	}
	return frameworkConfig
}
