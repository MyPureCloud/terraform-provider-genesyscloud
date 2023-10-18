package genesyscloud

import (
	"context"
	"log"
	"sync"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

// SDKClientPool holds a pool of client configs for the Genesys Cloud SDK. One should be
// acquired at the beginning of any resource operation and released on completion.
// This has the benefit of ensuring we don't issue too many concurrent requests and also
// increases throughput as each token will have its own rate limit.
type SDKClientPool struct {
	pool chan *platformclientv2.Configuration
}

var sdkClientPool *SDKClientPool
var sdkClientPoolErr diag.Diagnostics
var once sync.Once

// InitSDKClientPool creates a new pool of Clients with the given provider config
// This must be called during provider initialization before the pool is used
func InitSDKClientPool(max int, version string, providerConfig *schema.ResourceData) diag.Diagnostics {
	once.Do(func() {
		log.Print("Initializing default SDK client.")
		// Initialize the default config for tests and anything else that doesn't use the pool
		err := initClientConfig(providerConfig, version, platformclientv2.GetDefaultConfiguration())
		if err != nil {
			sdkClientPoolErr = err
			return
		}

		log.Printf("Initializing %d SDK clients in the pool.", max)
		sdkClientPool = &SDKClientPool{
			pool: make(chan *platformclientv2.Configuration, max),
		}
		sdkClientPoolErr = sdkClientPool.preFill(providerConfig, version)
	})
	return sdkClientPoolErr
}

func (p *SDKClientPool) preFill(providerConfig *schema.ResourceData, version string) diag.Diagnostics {
	errorChan := make(chan diag.Diagnostics)
	wgDone := make(chan bool)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < cap(p.pool); i++ {
		sdkConfig := platformclientv2.NewConfiguration()
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := initClientConfig(providerConfig, version, sdkConfig)
			if err != nil {
				select {
				case <-ctx.Done():
				case errorChan <- err:
				}
				cancel()
				return
			}
		}()
		p.pool <- sdkConfig
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

func (p *SDKClientPool) acquire() *platformclientv2.Configuration {
	return <-p.pool
}

func (p *SDKClientPool) release(c *platformclientv2.Configuration) {
	select {
	case p.pool <- c:
	default:
		// Pool is full. Don't put it back in the pool
	}
}

type resContextFunc func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
type getAllConfigFunc func(context.Context, *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics)

func CreateWithPooledClient(method resContextFunc) schema.CreateContextFunc {
	return schema.CreateContextFunc(runWithPooledClient(method))
}

func ReadWithPooledClient(method resContextFunc) schema.ReadContextFunc {
	return schema.ReadContextFunc(runWithPooledClient(method))
}

func UpdateWithPooledClient(method resContextFunc) schema.UpdateContextFunc {
	return schema.UpdateContextFunc(runWithPooledClient(method))
}

func DeleteWithPooledClient(method resContextFunc) schema.DeleteContextFunc {
	return schema.DeleteContextFunc(runWithPooledClient(method))
}

// Inject a pooled SDK client connection into a resource method's meta argument
// and automatically return it to the pool on completion
func runWithPooledClient(method resContextFunc) resContextFunc {
	return func(ctx context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
		clientConfig := sdkClientPool.acquire()
		defer sdkClientPool.release(clientConfig)

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

// Inject a pooled SDK client connection into an exporter's getAll* method
func GetAllWithPooledClient(method getAllConfigFunc) resourceExporter.GetAllResourcesFunc {
	return func(ctx context.Context) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
		clientConfig := sdkClientPool.acquire()
		defer sdkClientPool.release(clientConfig)

		// Check if the request has been cancelled
		select {
		case <-ctx.Done():
			return nil, diag.FromErr(ctx.Err()) // Error somewhere, terminate
		default:
		}

		return method(ctx, clientConfig)
	}
}
