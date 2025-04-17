package mrmo

import "github.com/mypurecloud/platform-client-sdk-go/v154/platformclientv2"

var clientConfig *platformclientv2.Configuration
var isActive bool

func setClientConfig(config *platformclientv2.Configuration) {
	clientConfig = config
}

func GetClientConfig() *platformclientv2.Configuration {
	return clientConfig
}

func IsActive() bool {
	return isActive
}

func Activate(clientConfig *platformclientv2.Configuration) {
	setClientConfig(clientConfig)
	isActive = true
}

func Deactivate() {
	isActive = false
}
