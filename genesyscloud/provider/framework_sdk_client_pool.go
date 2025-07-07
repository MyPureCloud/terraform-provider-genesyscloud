package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider/sdk_config"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var SdkClientPoolFrameWorkDiagErr diag.Diagnostics
var frameWorkOnce sync.Once

func (f *GenesysCloudProvider) InitSDKClientPool() diag.Diagnostics {
	frameWorkOnce.Do(func() {
		log.Print("(Framework) Initializing default SDK client.")
		// Initialize the default config for tests and anything else that doesn't use the Pool
		err := f.InitClientConfig(platformclientv2.GetDefaultConfiguration())
		if err != nil {
			log.Println("(Framework) Caught error from InitClientConfig: ", err)
			SdkClientPoolFrameWorkDiagErr = err
			return
		}

		maxClients := MaxClients
		if v := f.TokenPoolSize; v != 0 {
			maxClients = int(v)
		}

		log.Printf("(Framework) Initializing %d SDK clients in the Pool.", f.TokenPoolSize)
		f.SdkClientPool = SDKClientPool{
			Pool: make(chan *platformclientv2.Configuration, maxClients),
		}
		SdkClientPoolFrameWorkDiagErr = f.frameworkPreFill()
	})
	return SdkClientPoolFrameWorkDiagErr
}

func (f *GenesysCloudProvider) frameworkPreFill() diag.Diagnostics {
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
			err := f.InitClientConfig(sdkConfig)
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

func AcquireSdkClient(ctx context.Context, client GenesysCloudProvider) (*platformclientv2.Configuration, error) {
	clientConfig, err := client.SdkClientPool.acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer func(SdkClientPool *SDKClientPool, c *platformclientv2.Configuration) {
		err = SdkClientPool.release(c)
		if err != nil {
			log.Printf("Error releasing client: %s", err.Error())
		}
	}(&client.SdkClientPool, clientConfig)

	// Check if the request has been cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err() // Error somewhere, terminate
	default:
	}

	return clientConfig, nil
}

func (f *GenesysCloudProvider) InitClientConfig(config *platformclientv2.Configuration) diag.Diagnostics {
	if f.AuthDetails == nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("AuthDetails object is still nil!!", "AuthDetails object is still nil!!"),
		}
	}
	accessToken := f.AuthDetails.AccessToken
	oauthclientID := f.AuthDetails.ClientId
	oauthclientSecret := f.AuthDetails.ClientSecret
	basePath := GetRegionBasePath(f.AuthDetails.Region)
	config.BasePath = basePath

	if err := f.setUpSDKLogging(config); err != nil {
		return err
	}

	f.setupProxy(config)
	f.setupGateway(config)

	config.AddDefaultHeader("User-Agent", "GC Terraform Provider/"+f.Version)
	config.RetryConfiguration = &platformclientv2.RetryConfiguration{
		RetryWaitMin: time.Second * 1,
		RetryWaitMax: time.Second * 30,
		RetryMax:     20,
		RequestLogHook: func(request *http.Request, count int) {
			sdkDebugReq := newSDKDebugRequest(request, count)
			request.Header.Set("TF-Correlation-Id", sdkDebugReq.TransactionId)
			err, jsonStr := sdkDebugReq.ToJSON()
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

	sdk_config.SetConfig(config)

	log.Printf("Initialized Go SDK Client. Debug=%t", f.SdkDebugInfo.DebugEnabled)
	return nil
}

func (f *GenesysCloudProvider) setUpSDKLogging(config *platformclientv2.Configuration) diag.Diagnostics {
	var diagErrors diag.Diagnostics = make([]diag.Diagnostic, 0)

	if f.SdkDebugInfo == nil {
		return nil
	}

	sdkDebugFilePath := f.SdkDebugInfo.FilePath
	if f.SdkDebugInfo.DebugEnabled {
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

		if format := f.SdkDebugInfo.Format; format == "Json" {
			config.LoggingConfiguration.SetLogFormat(platformclientv2.JSON)
		} else {
			config.LoggingConfiguration.SetLogFormat(platformclientv2.Text)
		}
	}
	return nil
}

func (f *GenesysCloudProvider) setupProxy(config *platformclientv2.Configuration) {
	if f.Proxy == nil {
		return
	}

	config.ProxyConfiguration = &platformclientv2.ProxyConfiguration{}

	config.ProxyConfiguration.Host = f.Proxy.Host
	config.ProxyConfiguration.Port = f.Proxy.Port
	config.ProxyConfiguration.Protocol = f.Proxy.Protocol

	if f.Proxy.Auth == nil {
		return
	}

	config.ProxyConfiguration.Auth = &platformclientv2.Auth{}
	config.ProxyConfiguration.Auth.UserName = f.Proxy.Auth.Username
	config.ProxyConfiguration.Auth.Password = f.Proxy.Auth.Password
}

func (f *GenesysCloudProvider) setupGateway(config *platformclientv2.Configuration) {
	if f.Gateway == nil {
		return
	}

	config.GateWayConfiguration = &platformclientv2.GateWayConfiguration{}
	config.GateWayConfiguration.Host = f.Gateway.Host
	config.GateWayConfiguration.Port = f.Gateway.Port
	config.GateWayConfiguration.Protocol = f.Gateway.Protocol

	for _, param := range f.Gateway.PathParams {
		config.GateWayConfiguration.PathParams = append(config.GateWayConfiguration.PathParams, &platformclientv2.PathParams{
			PathName:  param.PathName,
			PathValue: param.PathValue,
		})
	}

	if f.Gateway.Auth == nil {
		return
	}
	config.GateWayConfiguration.Auth = &platformclientv2.Auth{}

	config.GateWayConfiguration.Auth.UserName = f.Gateway.Auth.Username
	config.GateWayConfiguration.Auth.Password = f.Gateway.Auth.Password
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
