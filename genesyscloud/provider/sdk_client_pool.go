package provider

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

const (
	// Default timeouts
	DefaultAcquireTimeout = 5 * time.Minute
	DefaultInitTimeout    = 10 * time.Minute

	// Default pool settings
	DefaultMaxClients = 10
	MinClients        = 1
	MaxClients        = 20

	// Logging intervals
	MetricsLoggingInterval = 5 * time.Minute

	// Provider attribute keys
	AttrTokenPoolSize       = "token_pool_size"
	AttrTokenAcquireTimeout = "token_acquire_timeout"
	AttrTokenInitTimeout    = "token_init_timeout"
	AttrSdkClientPoolDebug  = "sdk_client_pool_debug"
)

// Pool capacity thresholds for warnings
const (
	PoolNearCapacityThreshold = 2 // Warn when available clients <= 2
	PoolCriticalThreshold     = 1 // Warn when available clients <= 1
)

// SDKClientPool holds a Pool of client configs for the Genesys Cloud SDK. One should be
// acquired at the beginning of any resource operation and released on completion.
// This has the benefit of ensuring we don't issue too many concurrent requests and also
// increases throughput as each token will have its own rate limit.
type SDKClientPool struct {
	Pool          chan *platformclientv2.Configuration
	activeClients int64
	config        *SDKClientPoolConfig
	metrics       *poolMetrics
}

type SDKClientPoolConfig struct {
	MaxClients     int
	AcquireTimeout time.Duration
	InitTimeout    time.Duration
	DebugLogging   bool
}

type poolMetrics struct {
	totalAcquires   int64
	totalReleases   int64
	acquireTimeouts int64
	lastAcquireTime time.Time
	mu              sync.RWMutex
}

var SdkClientPool *SDKClientPool
var SdkClientPoolErr diag.Diagnostics
var Once sync.Once

// InitSDKClientPool creates a new Pool of Clients with the given provider config
// This must be called during provider initialization before the Pool is used
func InitSDKClientPool(version string, providerConfig *schema.ResourceData) diag.Diagnostics {
	Once.Do(func() {
		log.Print("Initializing default SDK client.")
		// Initialize the default config for tests and anything else that doesn't use the Pool
		err := InitClientConfig(providerConfig, version, platformclientv2.GetDefaultConfiguration(), true)
		if err != nil {
			SdkClientPoolErr = err
			return
		}

		max := MaxClients
		if v, ok := providerConfig.GetOk(AttrTokenPoolSize); ok {
			max = v.(int)
		}

		// Get timeouts from provider config
		acquireTimeout := DefaultAcquireTimeout
		if v, ok := providerConfig.GetOk(AttrTokenAcquireTimeout); ok {
			if parsed, err := time.ParseDuration(v.(string)); err == nil {
				acquireTimeout = parsed
			}
		}

		initTimeout := DefaultInitTimeout
		if v, ok := providerConfig.GetOk(AttrTokenInitTimeout); ok {
			if parsed, err := time.ParseDuration(v.(string)); err == nil {
				initTimeout = parsed
			}
		}

		config := &SDKClientPoolConfig{
			MaxClients:     max,
			AcquireTimeout: acquireTimeout,
			InitTimeout:    initTimeout,
			DebugLogging:   providerConfig.Get(AttrSdkClientPoolDebug).(bool),
		}

		SdkClientPool = &SDKClientPool{
			Pool:    make(chan *platformclientv2.Configuration, max),
			config:  config,
			metrics: &poolMetrics{},
		}
		SdkClientPool.logDebug("Initialized %d SDK clients in the Pool with acquire timeout %v and init timeout %v.", max, acquireTimeout, initTimeout)

		// Start periodic metrics logging
		go func() {
			ticker := time.NewTicker(MetricsLoggingInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					SdkClientPool.logDebug("Client pool status - %s", SdkClientPool.formatMetrics())
				}
			}
		}()

		SdkClientPoolErr = SdkClientPool.preFill(providerConfig, version)
	})
	return SdkClientPoolErr
}

func (p *SDKClientPool) logDebug(msg string, args ...interface{}) {
	if p.config.DebugLogging {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

func (p *SDKClientPool) formatMetrics() string {
	metrics := p.GetMetrics()
	lastAcquireTime := metrics["last_acquire_time"].(time.Time)

	// Check if the last acquire time is zero, if so, set it to a default value
	var lastAcquireTimeStr string
	if lastAcquireTime.IsZero() {
		lastAcquireTimeStr = "never"
	} else {
		lastAcquireTimeStr = lastAcquireTime.Format(time.RFC3339)
	}

	return fmt.Sprintf("Active: %d/%d, Acquires: %d, Releases: %d, Timeouts: %d, Last Acquire: %s",
		metrics["active_clients"],
		p.config.MaxClients,
		metrics["total_acquires"],
		metrics["total_releases"],
		metrics["acquire_timeouts"],
		lastAcquireTimeStr,
	)
}

func (p *SDKClientPool) preFill(providerConfig *schema.ResourceData, version string) diag.Diagnostics {
	p.logDebug("Prefilling SDK client pool with %d clients.", p.config.MaxClients)

	ctx, cancel := context.WithTimeout(context.Background(), p.config.InitTimeout)
	defer cancel()

	errorChan := make(chan diag.Diagnostics)
	wgDone := make(chan bool)
	var wg sync.WaitGroup

	for i := 0; i < p.config.MaxClients; i++ {
		sdkConfig := platformclientv2.NewConfiguration()
		wg.Add(1)
		go func(config *platformclientv2.Configuration) {
			defer wg.Done()
			if err := InitClientConfig(providerConfig, version, config, false); err != nil {
				select {
				case <-ctx.Done():
				case errorChan <- err:
				}
				cancel()
				return
			}
			select {
			case <-ctx.Done():
				return
			case p.Pool <- config:
			}
		}(sdkConfig)
	}

	go func() {
		wg.Wait()
		close(wgDone)
	}()

	// Wait until either WaitGroup is done or an error is received
	select {
	case <-wgDone:
		p.logDebug("Successfully pre-filled client pool - %s", p.formatMetrics())
		return nil
	case err := <-errorChan:
		p.logDebug("Error pre-filling client pool - %s", p.formatMetrics())
		return err
	case <-ctx.Done():
		p.logDebug("Timed out pre-filling client pool - %s", p.formatMetrics())
		return diag.Errorf("Timed out pre-filling client pool: %v", ctx.Err())
	}
}

func (p *SDKClientPool) acquire() (*platformclientv2.Configuration, error) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), p.config.AcquireTimeout)
	defer cancel()

	select {
	case client := <-p.Pool:
		atomic.AddInt64(&p.activeClients, 1)
		atomic.AddInt64(&p.metrics.totalAcquires, 1)
		p.metrics.mu.Lock()
		p.metrics.lastAcquireTime = time.Now()
		p.metrics.mu.Unlock()

		acquiredMsg := "Client acquired from pool"

		if atomic.LoadInt64(&p.activeClients) >= int64(p.config.MaxClients-PoolCriticalThreshold) {
			log.Printf("[WARN] %s but pool at critical capacity - %s", acquiredMsg, p.formatMetrics())
		} else if atomic.LoadInt64(&p.activeClients) >= int64(p.config.MaxClients-PoolNearCapacityThreshold) {
			log.Printf("[WARN] %s with pool near capacity - %s", acquiredMsg, p.formatMetrics())
		} else {
			p.logDebug("%s - %s", acquiredMsg, p.formatMetrics())
		}
		return client, nil
	case <-timeoutCtx.Done():
		atomic.AddInt64(&p.metrics.acquireTimeouts, 1)
		log.Printf("[WARN] Client acquisition timeout - %s", p.formatMetrics())
		return nil, fmt.Errorf("timeout after %v waiting for available client", p.config.AcquireTimeout)
	}
}

func (p *SDKClientPool) release(c *platformclientv2.Configuration) {
	select {
	case p.Pool <- c:
		atomic.AddInt64(&p.activeClients, -1)
		atomic.AddInt64(&p.metrics.totalReleases, 1)

		if atomic.LoadInt64(&p.activeClients) <= int64(p.config.MaxClients-PoolCriticalThreshold) {
			p.logDebug("Client released from full pool - %s", p.formatMetrics())
		}
	default:
		// Pool is full. Don't put it back in the Pool
		p.logDebug("Attempted to release client to full pool - %s", p.formatMetrics())
	}
}

func (p *SDKClientPool) GetMetrics() map[string]interface{} {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	return map[string]interface{}{
		"active_clients":    atomic.LoadInt64(&p.activeClients),
		"total_acquires":    atomic.LoadInt64(&p.metrics.totalAcquires),
		"total_releases":    atomic.LoadInt64(&p.metrics.totalReleases),
		"acquire_timeouts":  atomic.LoadInt64(&p.metrics.acquireTimeouts),
		"last_acquire_time": p.metrics.lastAcquireTime,
	}
}

type resContextFunc func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
type GetAllConfigFunc func(context.Context, *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics)
type GetCustomConfigFunc func(context.Context, *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, diag.Diagnostics)

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
// and automatically return it to the Pool on completion
func runWithPooledClient(method resContextFunc) resContextFunc {
	return func(ctx context.Context, r *schema.ResourceData, meta interface{}) diag.Diagnostics {
		clientConfig, err := SdkClientPool.acquire()
		if err != nil {
			return diag.FromErr(err)
		}
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

// Inject a pooled SDK client connection into an exporter's getAll* method
func GetAllWithPooledClient(method GetAllConfigFunc) resourceExporter.GetAllResourcesFunc {
	return func(ctx context.Context) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
		clientConfig, err := SdkClientPool.acquire()
		if err != nil {
			return nil, diag.FromErr(err)
		}
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
		clientConfig, err := SdkClientPool.acquire()
		if err != nil {
			return nil, nil, diag.FromErr(err)
		}
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
