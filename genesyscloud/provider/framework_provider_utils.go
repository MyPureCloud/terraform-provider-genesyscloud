package provider

import (
	"log"
	"strconv"
)

// configureAuthDetails sets up authentication configuration for the Genesys Cloud provider.
// It populates the AuthDetails struct using values from either the provider configuration
// or environment variables, with provider configuration taking precedence.
//
// Parameters:
//   - data: GenesysCloudProviderModel containing provider configuration values
//   - providerEnvValues: Environment variables for provider configuration
func (f GenesysCloudProvider) configureAuthInfo(data GenesysCloudProviderModel) {
	if f.AttributeEnvValues == nil {
		f.AttributeEnvValues = readProviderEnvVars()
	}

	var authDetails AuthInfo

	if data.AccessToken.ValueString() != "" {
		authDetails.AccessToken = data.AccessToken.ValueString()
	} else if f.AttributeEnvValues.accessToken != "" {
		authDetails.AccessToken = f.AttributeEnvValues.accessToken
	}

	if data.OAuthClientId.ValueString() != "" {
		authDetails.AccessToken = data.OAuthClientId.ValueString()
	} else if f.AttributeEnvValues.clientId != "" {
		authDetails.AccessToken = f.AttributeEnvValues.clientId
	}

	if data.OAuthClientSecret.ValueString() != "" {
		authDetails.ClientSecret = data.OAuthClientSecret.ValueString()
	} else if f.AttributeEnvValues.clientSecret != "" {
		authDetails.ClientSecret = f.AttributeEnvValues.clientSecret
	}

	if data.AwsRegion.ValueString() != "" {
		authDetails.Region = data.AwsRegion.ValueString()
	} else if f.AttributeEnvValues.region != "" {
		authDetails.Region = f.AttributeEnvValues.region
	} else {
		authDetails.Region = awsRegionDefaultValue
	}

	f.AuthDetails = &authDetails
}

// configureSdkDebugInfo configures debugging settings for the Genesys Cloud SDK.
// It sets up debug options including enabling/disabling debug mode, debug format,
// and debug file path using values from either the provider configuration or
// environment variables, with provider configuration taking precedence.
//
// Parameters:
//   - data: GenesysCloudProviderModel containing provider configuration values
//   - providerEnvValues: Environment variables for provider configuration
func (f GenesysCloudProvider) configureSdkDebugInfo(data GenesysCloudProviderModel) {
	if f.AttributeEnvValues == nil {
		f.AttributeEnvValues = readProviderEnvVars()
	}

	var sdkDebugInfo SdkDebugInfo

	if !data.SdkDebug.IsNull() {
		sdkDebugInfo.DebugEnabled = data.SdkDebug.ValueBool()
	} else {
		sdkDebugInfo.DebugEnabled = f.AttributeEnvValues.sdkDebug != ""
	}

	sdkDebugInfo.Format = sdkDebugFormatDefaultValue
	if data.SdkDebugFormat.ValueString() != "" {
		sdkDebugInfo.Format = data.SdkDebugFormat.ValueString()
	} else if f.AttributeEnvValues.sdkDebugFormat != "" {
		sdkDebugInfo.Format = f.AttributeEnvValues.sdkDebugFormat
	}

	sdkDebugInfo.FilePath = sdkDebugFilePathDefaultValue
	if data.SdkDebugFilePath.ValueString() != "" {
		sdkDebugInfo.FilePath = data.SdkDebugFilePath.ValueString()
	} else if f.AttributeEnvValues.sdkDebugFilePath != "" {
		sdkDebugInfo.FilePath = f.AttributeEnvValues.sdkDebugFilePath
	}

	f.SdkDebugInfo = &sdkDebugInfo
}

func (f GenesysCloudProvider) configureRootAttributes(data GenesysCloudProviderModel) {
	if f.AttributeEnvValues == nil {
		f.AttributeEnvValues = readProviderEnvVars()
	}

	f.TokenPoolSize = tokenPoolSizeDefault
	if !data.TokenPoolSize.IsNull() {
		f.TokenPoolSize = data.TokenPoolSize.ValueInt32()
	} else if f.AttributeEnvValues.tokenPoolSize != "" {
		tokenPoolSize, err := strconv.Atoi(f.AttributeEnvValues.tokenPoolSize)
		if err != nil {
			log.Printf("Failed to parse %s env var to int. Defaulting to %d. Error: %s", tokenPoolSizeEnvVar, tokenPoolSizeDefault, err.Error())
		} else {
			f.TokenPoolSize = int32(tokenPoolSize)
		}
	}

	if !data.LogStackTraces.IsNull() {
		f.LogStackTraces = data.LogStackTraces.ValueBool()
	} else {
		f.LogStackTraces = f.AttributeEnvValues.logStackTraces != ""
	}

	f.LogStackTracesFilePath = logStackTracesFilePathDefaultValue
	if data.LogStackTracesFilePath.ValueString() != "" {
		f.LogStackTracesFilePath = data.LogStackTracesFilePath.ValueString()
	} else if f.AttributeEnvValues.logStackTracesFilePath != "" {
		f.LogStackTracesFilePath = f.AttributeEnvValues.logStackTracesFilePath
	}
}

func (f GenesysCloudProvider) configureProxyAttributes(data GenesysCloudProviderModel) {
	if f.AttributeEnvValues == nil {
		f.AttributeEnvValues = readProviderEnvVars()
	}

	if data.Proxy == nil || len(data.Proxy) == 0 {
		return
	}

	dataProxyObject := data.Proxy[0]

	var proxy Proxy
	defer func() {
		f.Proxy = &proxy
	}()

	// Configure Proxy Host
	if dataProxyObject.Host.ValueString() != "" {
		proxy.Host = dataProxyObject.Host.ValueString()
	} else if f.AttributeEnvValues.proxyHost != "" {
		proxy.Host = f.AttributeEnvValues.proxyHost
	}

	if dataProxyObject.Port.ValueString() != "" {
		proxy.Port = dataProxyObject.Port.ValueString()
	} else if f.AttributeEnvValues.proxyPort != "" {
		proxy.Port = f.AttributeEnvValues.proxyPort
	}

	if dataProxyObject.Auth == nil || len(dataProxyObject.Auth) == 0 {
		return
	}
	proxyAuthObject := dataProxyObject.Auth[0]

	if proxyAuthObject.Username.ValueString() != "" {
		proxy.Auth.Username = proxyAuthObject.Username.ValueString()
	} else if f.AttributeEnvValues.proxyAuthUsername != "" {
		proxy.Auth.Username = f.AttributeEnvValues.proxyAuthUsername
	}

	if proxyAuthObject.Password.ValueString() != "" {
		proxy.Auth.Password = proxyAuthObject.Password.ValueString()
	} else if f.AttributeEnvValues.proxyAuthPassword != "" {
		proxy.Auth.Password = f.AttributeEnvValues.proxyAuthPassword
	}
}

func (f GenesysCloudProvider) configureGatewayAttributes(data GenesysCloudProviderModel) {
	if f.AttributeEnvValues == nil {
		f.AttributeEnvValues = readProviderEnvVars()
	}

	if data.Gateway == nil || len(data.Gateway) == 0 {
		return
	}
	dataGatewayObject := data.Gateway[0]

	var gateway Gateway
	defer func() {
		f.Gateway = &gateway
	}()

	if dataGatewayObject.Host.ValueString() != "" {
		gateway.Host = dataGatewayObject.Host.ValueString()
	} else if f.AttributeEnvValues.gatewayHost != "" {
		gateway.Host = f.AttributeEnvValues.gatewayHost
	}

	if dataGatewayObject.Port.ValueString() != "" {
		gateway.Port = dataGatewayObject.Port.ValueString()
	} else if f.AttributeEnvValues.gatewayPort != "" {
		gateway.Port = f.AttributeEnvValues.gatewayPort
	}

	if dataGatewayObject.Protocol.ValueString() != "" {
		gateway.Protocol = dataGatewayObject.Protocol.ValueString()
	} else if f.AttributeEnvValues.gatewayProtocol != "" {
		gateway.Protocol = f.AttributeEnvValues.gatewayProtocol
	}

	for _, param := range dataGatewayObject.PathParams {
		pathName := ""
		pathValue := ""

		if param.PathName.ValueString() != "" {
			pathName = param.PathName.ValueString()
		} else if f.AttributeEnvValues.gatewayPathParamsName != "" {
			pathName = f.AttributeEnvValues.gatewayPathParamsName
		}

		if param.PathValue.ValueString() != "" {
			pathValue = param.PathValue.ValueString()
		} else if f.AttributeEnvValues.gatewayPathParamsValue != "" {
			pathValue = f.AttributeEnvValues.gatewayPathParamsValue
		}

		gateway.PathParams = append(f.Gateway.PathParams, PathParam{
			PathName:  pathName,
			PathValue: pathValue,
		})
	}

	if dataGatewayObject.Auth == nil || len(dataGatewayObject.Auth) == 0 {
		return
	}
	dataGatewayAuthObject := dataGatewayObject.Auth[0]

	if dataGatewayAuthObject.Username.ValueString() != "" {
		gateway.Auth.Username = dataGatewayAuthObject.Username.ValueString()
	} else if f.AttributeEnvValues.gatewayAuthUsername != "" {
		gateway.Auth.Username = f.AttributeEnvValues.gatewayAuthUsername
	}

	if dataGatewayAuthObject.Password.ValueString() != "" {
		gateway.Auth.Password = dataGatewayAuthObject.Password.ValueString()
	} else if f.AttributeEnvValues.gatewayAuthPassword != "" {
		gateway.Auth.Password = f.AttributeEnvValues.gatewayAuthPassword
	}
}
