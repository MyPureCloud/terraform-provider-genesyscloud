package provider

import "os"

const (
	accessTokenEnvVar            = "GENESYSCLOUD_ACCESS_TOKEN"
	clientIdEnvVar               = "GENESYSCLOUD_OAUTHCLIENT_ID"
	clientSecretEnvVar           = "GENESYSCLOUD_OAUTHCLIENT_SECRET"
	regionEnvVar                 = "GENESYSCLOUD_REGION"
	sdkDebugEnvVar               = "GENESYSCLOUD_SDK_DEBUG"
	sdkDebugFilePathEnvVar       = "GENESYSCLOUD_SDK_DEBUG_FILE_PATH"
	sdkDebugFormatEnvVar         = "GENESYSCLOUD_SDK_DEBUG_FORMAT"
	tokenPoolSizeEnvVar          = "GENESYSCLOUD_TOKEN_POOL_SIZE"
	logStackTracesEnvVar         = "GENESYSCLOUD_LOG_STACK_TRACES"
	logStackTracesFilePathEnvVar = "GENESYSCLOUD_LOG_STACK_TRACES_FILE_PATH"
	gatewayHostEnvVar            = "GENESYSCLOUD_GATEWAY_HOST"
	gatewayPortEnvVar            = "GENESYSCLOUD_GATEWAY_PORT"
	gatewayProtocolEnvVar        = "GENESYSCLOUD_GATEWAY_PROTOCOL"
	gatewayAuthUsernameEnvVar    = "GENESYSCLOUD_GATEWAY_AUTH_USERNAME"
	gatewayAuthPasswordEnvVar    = "GENESYSCLOUD_GATEWAY_AUTH_PASSWORD"
	gatewayPathParamsNameEnvVar  = "GENESYSCLOUD_GATEWAY_PATH_NAME"
	gatewayPathParamsValueEnvVar = "GENESYSCLOUD_GATEWAY_PATH_VALUE"
	proxyHostEnvVar              = "GENESYSCLOUD_PROXY_HOST"
	proxyPortEnvVar              = "GENESYSCLOUD_PROXY_PORT"
	proxyProtocolEnvVar          = "GENESYSCLOUD_PROXY_PROTOCOL"
	proxyAuthUsernameEnvVar      = "GENESYSCLOUD_PROXY_AUTH_USERNAME"
	proxyAuthPasswordEnvVar      = "GENESYSCLOUD_PROXY_AUTH_PASSWORD"
)

type providerEnvVars struct {
	accessToken            string
	region                 string
	clientId               string
	clientSecret           string
	sdkDebug               string
	sdkDebugFilePath       string
	sdkDebugFormat         string
	tokenPoolSize          string
	logStackTraces         string
	logStackTracesFilePath string
	gatewayHost            string
	gatewayPort            string
	gatewayProtocol        string
	gatewayAuthUsername    string
	gatewayAuthPassword    string
	gatewayPathParamsName  string
	gatewayPathParamsValue string
	proxyHost              string
	proxyPort              string
	proxyProtocol          string
	proxyAuthUsername      string
	proxyAuthPassword      string
}

func readProviderEnvVars() *providerEnvVars {
	return &providerEnvVars{
		accessToken:            os.Getenv(accessTokenEnvVar),
		region:                 os.Getenv(regionEnvVar),
		clientId:               os.Getenv(clientIdEnvVar),
		clientSecret:           os.Getenv(clientSecretEnvVar),
		sdkDebug:               os.Getenv(sdkDebugEnvVar),
		sdkDebugFilePath:       os.Getenv(sdkDebugFilePathEnvVar),
		sdkDebugFormat:         os.Getenv(sdkDebugFormatEnvVar),
		tokenPoolSize:          os.Getenv(tokenPoolSizeEnvVar),
		logStackTraces:         os.Getenv(logStackTracesEnvVar),
		logStackTracesFilePath: os.Getenv(logStackTracesFilePathEnvVar),
		gatewayHost:            os.Getenv(gatewayHostEnvVar),
		gatewayPort:            os.Getenv(gatewayPortEnvVar),
		gatewayProtocol:        os.Getenv(gatewayProtocolEnvVar),
		gatewayAuthUsername:    os.Getenv(gatewayAuthUsernameEnvVar),
		gatewayAuthPassword:    os.Getenv(gatewayAuthPasswordEnvVar),
		gatewayPathParamsName:  os.Getenv(gatewayPathParamsNameEnvVar),
		gatewayPathParamsValue: os.Getenv(gatewayPathParamsValueEnvVar),
		proxyHost:              os.Getenv(proxyHostEnvVar),
		proxyPort:              os.Getenv(proxyPortEnvVar),
		proxyProtocol:          os.Getenv(proxyProtocolEnvVar),
		proxyAuthUsername:      os.Getenv(proxyAuthUsernameEnvVar),
		proxyAuthPassword:      os.Getenv(proxyAuthPasswordEnvVar),
	}
}
