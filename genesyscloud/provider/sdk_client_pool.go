package provider

import (
	"context"
	"log"
	"sync"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	prl "terraform-provider-genesyscloud/genesyscloud/util/panic_recovery_logger"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v152/platformclientv2"
)

// SDKClientPool holds a Pool of client configs for the Genesys Cloud SDK. One should be
// acquired at the beginning of any resource operation and released on completion.
// This has the benefit of ensuring we don't issue too many concurrent requests and also
// increases throughput as each token will have its own rate limit.
type SDKClientPool struct {
	Pool chan *platformclientv2.Configuration
}

var SdkClientPool *SDKClientPool
var SdkClientPoolErr diag.Diagnostics
var Once sync.Once

func (p *SDKClientPool) acquire() *platformclientv2.Configuration {
	return <-p.Pool
}

func (p *SDKClientPool) release(c *platformclientv2.Configuration) {
	select {
	case p.Pool <- c:
	default:
		// Pool is full. Don't put it back in the Pool
	}
}

// InitSDKClientPool creates a new Pool of Clients with the given provider config
// This must be called during provider initialization before the Pool is used
func InitSDKClientPool(max int, version string, providerConfig *schema.ResourceData) diag.Diagnostics {
	Once.Do(func() {
		log.Print("Initializing default SDK client.")
		// Initialize the default config for tests and anything else that doesn't use the Pool
		err := InitClientConfig(providerConfig, version, platformclientv2.GetDefaultConfiguration())
		if err != nil {
			log.Println("Caught error from InitClientConfig: ", err)
			SdkClientPoolErr = err
			return
		}

		log.Printf("Initializing %d SDK clients in the Pool.", max)
		SdkClientPool = &SDKClientPool{
			Pool: make(chan *platformclientv2.Configuration, max),
		}
		SdkClientPoolErr = SdkClientPool.preFill(providerConfig, version)
	})
	return SdkClientPoolErr
}

func (p *SDKClientPool) preFill(providerConfig *schema.ResourceData, version string) diag.Diagnostics {
	errorChan := make(chan diag.Diagnostics)
	wgDone := make(chan bool)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < cap(p.Pool); i++ {
		sdkConfig := platformclientv2.NewConfiguration()
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := InitClientConfig(providerConfig, version, sdkConfig)
			if err != nil {
				select {
				case <-ctx.Done():
				case errorChan <- err:
				}
				cancel()
				return
			}
		}()
		p.Pool <- sdkConfig
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

type resContextFunc func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
type GetAllConfigFunc func(context.Context, *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics)
type GetCustomConfigFunc func(context.Context, *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, diag.Diagnostics)

func CreateWithPooledClient(method resContextFunc) schema.CreateContextFunc {
	methodWrappedWithRecover := wrapWithRecover(method, constants.Create)
	return schema.CreateContextFunc(runWithPooledClient(methodWrappedWithRecover))
}

func ReadWithPooledClient(method resContextFunc) schema.ReadContextFunc {
	methodWrappedWithRecover := wrapWithRecover(method, constants.Read)
	return schema.ReadContextFunc(runWithPooledClient(methodWrappedWithRecover))
}

func UpdateWithPooledClient(method resContextFunc) schema.UpdateContextFunc {
	methodWrappedWithRecover := wrapWithRecover(method, constants.Update)
	return schema.UpdateContextFunc(runWithPooledClient(methodWrappedWithRecover))
}

func DeleteWithPooledClient(method resContextFunc) schema.DeleteContextFunc {
	methodWrappedWithRecover := wrapWithRecover(method, constants.Delete)
	return schema.DeleteContextFunc(runWithPooledClient(methodWrappedWithRecover))
}

func wrapWithRecover(method resContextFunc, operation constants.CRUDOperation) resContextFunc {
	return func(ctx context.Context, r *schema.ResourceData, meta any) (diagErr diag.Diagnostics) {
		panicRecoverLogger := prl.GetPanicRecoveryLoggerInstance()
		if !panicRecoverLogger.LoggerEnabled {
			return method(ctx, r, meta)
		}

		defer func() {
			if r := recover(); r != nil {
				err := panicRecoverLogger.HandleRecovery(r, operation)
				if err != nil {
					diagErr = diag.FromErr(err)
				}
			}
		}()

		return method(ctx, r, meta)
	}
}

// Inject a pooled SDK client connection into a resource method's meta argument
// and automatically return it to the Pool on completion
func runWithPooledClient(method resContextFunc) resContextFunc {
	return func(ctx context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
		clientConfig := SdkClientPool.acquire()
		defer SdkClientPool.release(clientConfig)

		// Check if the request has been cancelled
		select {
		case <-ctx.Done():
			return diag.FromErr(ctx.Err()) // Error somewhere, terminate
		default:
		}

		// Copy to a new providerMeta object and set the sdk config
		newMeta := *meta.(*ProviderMeta)
		newMeta.ClientConfig = clientConfig
		return method(ctx, r, &newMeta)
	}
}

// GetAllWithPooledClient injects a pooled SDK client connection into an exporter's getAll* method
func GetAllWithPooledClient(method GetAllConfigFunc) resourceExporter.GetAllResourcesFunc {
	return func(ctx context.Context) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
		clientConfig := SdkClientPool.acquire()
		defer SdkClientPool.release(clientConfig)

		// Check if the request has been cancelled
		select {
		case <-ctx.Done():
			return nil, diag.FromErr(ctx.Err()) // Error somewhere, terminate
		default:
		}

		return method(ctx, clientConfig)
	}
}

func GetAllWithPooledClientCustom(method GetCustomConfigFunc) resourceExporter.GetAllCustomResourcesFunc {
	return func(ctx context.Context) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, diag.Diagnostics) {
		clientConfig := SdkClientPool.acquire()
		defer SdkClientPool.release(clientConfig)

		// Check if the request has been cancelled
		select {
		case <-ctx.Done():
			return nil, nil, diag.FromErr(ctx.Err()) // Error somewhere, terminate
		default:
		}

		return method(ctx, clientConfig)
	}
}
