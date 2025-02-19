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

	if data.Proxy == nil {
		return
	}

	var proxy Proxy
	defer func() {
		f.Proxy = &proxy
	}()

	// Configure Proxy Host
	if data.Proxy.Host.ValueString() != "" {
		proxy.Host = data.Proxy.Host.ValueString()
	} else if f.AttributeEnvValues.proxyHost != "" {
		proxy.Host = f.AttributeEnvValues.proxyHost
	}

	if data.Proxy.Port.ValueString() != "" {
		proxy.Port = data.Proxy.Port.ValueString()
	} else if f.AttributeEnvValues.proxyPort != "" {
		proxy.Port = f.AttributeEnvValues.proxyPort
	}

	if data.Proxy.Auth == nil {
		return
	}

	if data.Proxy.Auth.Username.ValueString() != "" {
		proxy.Auth.Username = data.Proxy.Auth.Username.ValueString()
	} else if f.AttributeEnvValues.proxyAuthUsername != "" {
		proxy.Auth.Username = f.AttributeEnvValues.proxyAuthUsername
	}

	if data.Proxy.Auth.Password.ValueString() != "" {
		proxy.Auth.Password = data.Proxy.Auth.Password.ValueString()
	} else if f.AttributeEnvValues.proxyAuthPassword != "" {
		proxy.Auth.Password = f.AttributeEnvValues.proxyAuthPassword
	}
}

func (f GenesysCloudProvider) configureGatewayAttributes(data GenesysCloudProviderModel) {
	if f.AttributeEnvValues == nil {
		f.AttributeEnvValues = readProviderEnvVars()
	}

	if data.Gateway == nil {
		return
	}

	var gateway Gateway
	defer func() {
		f.Gateway = &gateway
	}()

	if data.Gateway.Host.ValueString() != "" {
		gateway.Host = data.Gateway.Host.ValueString()
	} else if f.AttributeEnvValues.gatewayHost != "" {
		gateway.Host = f.AttributeEnvValues.gatewayHost
	}

	if data.Gateway.Port.ValueString() != "" {
		gateway.Port = data.Gateway.Port.ValueString()
	} else if f.AttributeEnvValues.gatewayPort != "" {
		gateway.Port = f.AttributeEnvValues.gatewayPort
	}

	if data.Gateway.Protocol.ValueString() != "" {
		gateway.Protocol = data.Gateway.Protocol.ValueString()
	} else if f.AttributeEnvValues.gatewayProtocol != "" {
		gateway.Protocol = f.AttributeEnvValues.gatewayProtocol
	}

	for _, param := range data.Gateway.PathParams {
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

	if data.Gateway.Auth == nil {
		return
	}
	if data.Gateway.Auth.Username.ValueString() != "" {
		gateway.Auth.Username = data.Gateway.Auth.Username.ValueString()
	} else if f.AttributeEnvValues.gatewayAuthUsername != "" {
		gateway.Auth.Username = f.AttributeEnvValues.gatewayAuthUsername
	}

	if data.Gateway.Auth.Password.ValueString() != "" {
		gateway.Auth.Password = data.Gateway.Auth.Password.ValueString()
	} else if f.AttributeEnvValues.gatewayAuthPassword != "" {
		gateway.Auth.Password = f.AttributeEnvValues.gatewayAuthPassword
	}
}
