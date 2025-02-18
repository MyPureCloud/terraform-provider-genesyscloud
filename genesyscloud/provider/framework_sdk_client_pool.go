package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var SdkClientPoolFrameWorkDiagErr diag.Diagnostics

func (f GenesysCloudProvider) InitSDKClientPool(data GenesysCloudProviderModel) diag.Diagnostics {
	Once.Do(func() {
		log.Print("Initializing default SDK client.")
		// Initialize the default config for tests and anything else that doesn't use the Pool
		err := f.InitClientConfig(data, platformclientv2.GetDefaultConfiguration())
		if err != nil {
			SdkClientPoolFrameWorkDiagErr = err
			return
		}

		log.Printf("Initializing %d SDK clients in the Pool.", data.TokenPoolSize.ValueInt32())
		f.SdkClientPool = SDKClientPool{
			Pool: make(chan *platformclientv2.Configuration, data.TokenPoolSize.ValueInt32()),
		}
		SdkClientPoolFrameWorkDiagErr = f.frameworkPreFill(data)
	})
	return SdkClientPoolFrameWorkDiagErr
}

func (f GenesysCloudProvider) frameworkPreFill(data GenesysCloudProviderModel) diag.Diagnostics {
	errorChan := make(chan diag.Diagnostics)
	wgDone := make(chan bool)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < cap(f.SdkClientPool.Pool); i++ {
		sdkConfig := platformclientv2.NewConfiguration()
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := f.InitClientConfig(data, sdkConfig)
			if err != nil {
				select {
				case <-ctx.Done():
				case errorChan <- err:
				}
				cancel()
				return
			}
		}()
		f.SdkClientPool.Pool <- sdkConfig
	}
	go func() {
		wg.Wait()
		close(wgDone)
	}()

	// Wait until either WaitGroup is done or an error is received
	select {
	case <-wgDone:
		return nil
	case err := <-errorChan:
		return err
	}
}

func (f GenesysCloudProvider) InitClientConfig(data GenesysCloudProviderModel, config *platformclientv2.Configuration) diag.Diagnostics {
	accessToken := f.AuthDetails.AccessToken
	oauthclientID := f.AuthDetails.ClientId
	oauthclientSecret := f.AuthDetails.ClientSecret
	basePath := GetRegionBasePath(f.AuthDetails.Region)
	config.BasePath = basePath

	err := frameworkSetUpSDKLogging(data, config)
	if err != nil {
		return err
	}

	frameworkSetupProxy(data, config)
	frameworkSetupGateway(data, config)

	config.AddDefaultHeader("User-Agent", "GC Terraform Provider/"+f.Version)
	config.RetryConfiguration = &platformclientv2.RetryConfiguration{
		RetryWaitMin: time.Second * 1,
		RetryWaitMax: time.Second * 30,
		RetryMax:     20,
		RequestLogHook: func(request *http.Request, count int) {
			sdkDebugRequest := newSDKDebugRequest(request, count)
			request.Header.Set("TF-Correlation-Id", sdkDebugRequest.TransactionId)
			err, jsonStr := sdkDebugRequest.ToJSON()

			if err != nil {
				log.Printf("WARNING: Unable to log RequestLogHook: %s", err)
			}
			log.Println(jsonStr)
		},
		ResponseLogHook: func(response *http.Response) {
			sdkDebugResponse := newSDKDebugResponse(response)
			err, jsonStr := sdkDebugResponse.ToJSON()

			if err != nil {
				log.Printf("WARNING: Unable to log ResponseLogHook: %s", err)
			}
			log.Println(jsonStr)
		},
	}

	if accessToken != "" {
		log.Print("Setting access token set on configuration instance.")
		config.AccessToken = accessToken
	} else {
		config.AutomaticTokenRefresh = true // Enable automatic token refreshing

		return frameworkWithRetries(context.Background(), time.Minute, func() *retry.RetryError {
			authErr := config.AuthorizeClientCredentials(oauthclientID, oauthclientSecret)
			if authErr != nil {
				if !strings.Contains(authErr.Error(), "Auth Error: 400 - invalid_request (rate limit exceeded;") {
					return retry.NonRetryableError(fmt.Errorf("failed to authorize Genesys Cloud client credentials: %v", authErr))
				}
				return retry.RetryableError(fmt.Errorf("exhausted retries on Genesys Cloud client credentials. %v", authErr))
			}

			return nil
		})
	}

	log.Printf("Initialized Go SDK Client. Debug=%t", data.SdkDebug.ValueBool())
	return nil
}

func frameworkSetUpSDKLogging(data GenesysCloudProviderModel, config *platformclientv2.Configuration) diag.Diagnostics {
	var diagErrors diag.Diagnostics = make([]diag.Diagnostic, 0)

	sdkDebugFilePath := data.SdkDebugFilePath.ValueString()
	if data.SdkDebug.ValueBool() {
		config.LoggingConfiguration = &platformclientv2.LoggingConfiguration{
			LogLevel:        platformclientv2.LTrace,
			LogRequestBody:  true,
			LogResponseBody: true,
		}
		config.LoggingConfiguration.SetLogToConsole(false)
		config.LoggingConfiguration.SetLogFilePath(sdkDebugFilePath)

		dir, _ := filepath.Split(sdkDebugFilePath)
		if err := os.MkdirAll(dir, os.ModePerm); os.IsExist(err) {
			diagErrors.AddError("error while creating filepath for "+sdkDebugFilePath, err.Error())
			return diagErrors
		}

		if format := data.SdkDebugFormat.ValueString(); format == "Json" {
			config.LoggingConfiguration.SetLogFormat(platformclientv2.JSON)
		} else {
			config.LoggingConfiguration.SetLogFormat(platformclientv2.Text)
		}
	}
	return nil
}

func frameworkSetupProxy(data GenesysCloudProviderModel, config *platformclientv2.Configuration) {
	if data.Proxy == nil {
		return
	}

	config.ProxyConfiguration = &platformclientv2.ProxyConfiguration{}
	config.ProxyConfiguration.Host = data.Proxy.Host.ValueString()
	config.ProxyConfiguration.Port = data.Proxy.Port.ValueString()
	config.ProxyConfiguration.Protocol = data.Proxy.Protocol.ValueString()

	if data.Proxy.Auth == nil {
		return
	}

	config.ProxyConfiguration.Auth = &platformclientv2.Auth{}
	config.ProxyConfiguration.Auth.UserName = data.Proxy.Auth.Username.ValueString()
	config.ProxyConfiguration.Auth.Password = data.Proxy.Auth.Password.ValueString()
}

func frameworkSetupGateway(data GenesysCloudProviderModel, config *platformclientv2.Configuration) {
	if data.Gateway == nil {
		return
	}

	config.GateWayConfiguration = &platformclientv2.GateWayConfiguration{}
	config.GateWayConfiguration.Host = data.Gateway.Host.ValueString()
	config.GateWayConfiguration.Port = data.Gateway.Port.ValueString()
	config.GateWayConfiguration.Protocol = data.Gateway.Protocol.ValueString()

	for _, param := range data.Gateway.PathParams {
		config.GateWayConfiguration.PathParams = append(config.GateWayConfiguration.PathParams, &platformclientv2.PathParams{
			PathName:  param.PathName.ValueString(),
			PathValue: param.PathValue.ValueString(),
		})
	}

	if data.Gateway.Auth == nil {
		return
	}
	config.GateWayConfiguration.Auth = &platformclientv2.Auth{}

	config.GateWayConfiguration.Auth.UserName = data.Gateway.Auth.Username.ValueString()
	config.GateWayConfiguration.Auth.Password = data.Gateway.Auth.Password.ValueString()
}

func frameworkWithRetries(ctx context.Context, timeout time.Duration, method func() *retry.RetryError) diag.Diagnostics {
	var diagErrors diag.Diagnostics = make([]diag.Diagnostic, 0)

	err := retry.RetryContext(ctx, timeout, method)
	if err != nil && strings.Contains(err.Error(), "timeout while waiting for state to become") {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return frameworkWithRetries(ctx, timeout, method)
	}

	if err != nil {
		diagErrors.AddError(fmt.Sprintf("Operation exceeded timeout %s", timeout.String()), err.Error())
	}

	return diagErrors
}
