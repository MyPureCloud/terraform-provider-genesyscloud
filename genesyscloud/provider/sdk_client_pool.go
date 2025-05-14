package provider

import (
	"context"
	"fmt"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	prl "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/panic_recovery_logger"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
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
	Pool    chan *platformclientv2.Configuration
	config  *SDKClientPoolConfig
	metrics *poolMetrics
	done    chan struct{} // For cleanup
}

type SDKClientPoolConfig struct {
	AcquireTimeout time.Duration
	InitTimeout    time.Duration
	MaxClients     int
	DebugLogging   bool
}

type poolMetrics struct {
	activeClients   int64
	acquireTimeouts int64
	totalAcquires   int64
	totalReleases   int64
	lastAcquireTime time.Time
	mu              sync.RWMutex
}

func (m *poolMetrics) recordAcquire() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalAcquires++
	m.activeClients++
	m.lastAcquireTime = time.Now()
}

func (m *poolMetrics) recordRelease() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalReleases++
	m.activeClients--
}

var SdkClientPool *SDKClientPool
var SdkClientPoolErr diag.Diagnostics
var Once sync.Once

// InitSDKClientPool creates a new Pool of Clients with the given provider config
// This must be called during provider initialization before the Pool is used
func InitSDKClientPool(ctx context.Context, version string, providerConfig *schema.ResourceData) diag.Diagnostics {
	Once.Do(func() {
		log.Print("Initializing default SDK client.")
		// Initialize the default config for tests and anything else that doesn't use the Pool
		err := InitClientConfig(ctx, providerConfig, version, platformclientv2.GetDefaultConfiguration(), true)
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
			parsed, err := time.ParseDuration(v.(string))
			if err != nil {
				SdkClientPoolErr = diag.Errorf("Failed to parse token acquire timeout: %v", err)
				return
			}
			acquireTimeout = parsed
		}

		initTimeout := DefaultInitTimeout
		if v, ok := providerConfig.GetOk(AttrTokenInitTimeout); ok {
			parsed, err := time.ParseDuration(v.(string))
			if err != nil {
				SdkClientPoolErr = diag.Errorf("Failed to parse token init timeout: %v", err)
				return
			}
			initTimeout = parsed
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
			done:    make(chan struct{}),
		}
		SdkClientPool.logDebug("Initialized %d SDK clients in the Pool with acquire timeout %v and init timeout %v.", max, acquireTimeout, initTimeout)

		SdkClientPool.startMetricsLogging()
		SdkClientPoolErr = SdkClientPool.preFill(ctx, providerConfig, version)
	})
	return SdkClientPoolErr
}

func (p *SDKClientPool) startMetricsLogging() {
	// Start periodic metrics logging
	go func() {
		ticker := time.NewTicker(MetricsLoggingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.logDebug("Client pool status - %s", p.formatMetrics())
			case <-p.done:
				return
			}
		}
	}()
}

func (p *SDKClientPool) logDebug(msg string, args ...interface{}) {
	if p.config.DebugLogging {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

func (p *SDKClientPool) formatMetrics() string {
	metrics := p.GetMetrics()
	lastAcquireTime := metrics.lastAcquireTime

	// Check if the last acquire time is zero, if so, set it to a default value
	var lastAcquireTimeStr string
	if lastAcquireTime.IsZero() {
		lastAcquireTimeStr = "never"
	} else {
		lastAcquireTimeStr = lastAcquireTime.Format(time.RFC3339)
	}

	return fmt.Sprintf("Active: %d/%d, Acquires: %d, Releases: %d, Timeouts: %d, Last Acquire: %s",
		metrics.activeClients,
		p.config.MaxClients,
		metrics.totalAcquires,
		metrics.totalReleases,
		metrics.acquireTimeouts,
		lastAcquireTimeStr,
	)
}

func (p *SDKClientPool) preFill(ctx context.Context, providerConfig *schema.ResourceData, version string) diag.Diagnostics {
	p.logDebug("Prefilling SDK client pool with %d clients.", p.config.MaxClients)

	ctx, cancel := context.WithTimeout(ctx, p.config.InitTimeout)
	defer cancel()

	errorChan := make(chan diag.Diagnostics, p.config.MaxClients)
	var wg sync.WaitGroup

	// Create a semaphore to limit concurrent initializations
	concurrentInits := int(math.Min(
		float64(p.config.MaxClients),
		math.Max(5, float64(p.config.MaxClients/4))))
	sem := make(chan struct{}, concurrentInits)

	// Create a done channel for signaling goroutine cleanup
	initDone := make(chan struct{})
	defer close(initDone)

	for i := 0; i < p.config.MaxClients; i++ {
		sdkConfig := platformclientv2.NewConfiguration()
		wg.Add(1)
		go func(config *platformclientv2.Configuration) {
			defer wg.Done()
			defer func() {
				<-sem // Release semaphore
			}()

			// Try to acquire semaphore with context awareness
			select {
			case sem <- struct{}{}: // Acquire semaphore
			case <-ctx.Done():
				return
			case <-initDone:
				return
			}

			if err := InitClientConfig(ctx, providerConfig, version, config, false); err != nil {
				select {
				case errorChan <- err:
				case <-ctx.Done():
					log.Printf("[WARN] Context cancelled while trying to send error: %v", err)
				default:
				}
				return
			}

			// Try to add to pool with context awareness
			cleanup := false
			select {
			case p.Pool <- config:
				// Successfully added to the pool
			case <-ctx.Done():
				cleanup = true
			case <-initDone:
				cleanup = true
			}

			// Cleanup the config if we can't add it to the pool
			if cleanup {
				if err := cleanupConfiguration(config); err != nil {
					p.logDebug("Error cleaning up configuration during cancellation: %v", err)
				}
			}
		}(sdkConfig)
	}

	// Use a separate goroutine to collect errors
	var resultErr diag.Diagnostics
	errDone := make(chan struct{})
	go func() {
		defer close(errDone)
		for {
			select {
			case err, ok := <-errorChan:
				if !ok {
					return
				}
				resultErr = append(resultErr, err...)
			case <-ctx.Done():
				return
			case <-initDone:
				return
			}
		}
	}()

	// Wait for completion or context cancellation
	go func() {
		wg.Wait()
		close(errorChan)
	}()

	// Wait until either WaitGroup is done or an error is received
	select {
	case <-errDone:
		if resultErr != nil {
			p.logDebug("Error pre-filling client pool - %s", p.formatMetrics())
			return resultErr
		}
		p.logDebug("Successfully pre-filled client pool - %s", p.formatMetrics())
		return nil
	case <-ctx.Done():
		p.logDebug("Timed out pre-filling client pool - %s", p.formatMetrics())
		return diag.Errorf("Timed out pre-filling client pool: %v", ctx.Err())
	}
}

func (p *SDKClientPool) acquire(ctx context.Context) (*platformclientv2.Configuration, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, p.config.AcquireTimeout)
	defer cancel()

	select {
	case client := <-p.Pool:
		if client == nil {
			return nil, fmt.Errorf("received nil client from the pool")
		}
		p.metrics.recordAcquire()

		acquiredMsg := "Client acquired from pool"

		remaining := int64(p.config.MaxClients) - atomic.LoadInt64(&p.metrics.activeClients)
		if remaining <= int64(PoolCriticalThreshold) {
			p.logDebug("[WARN] %s but pool at critical capacity - %s", acquiredMsg, p.formatMetrics())
		} else if remaining <= int64(PoolNearCapacityThreshold) {
			p.logDebug("[WARN] %s with pool near capacity - %s", acquiredMsg, p.formatMetrics())
		} else {
			p.logDebug("%s - %s", acquiredMsg, p.formatMetrics())
		}
		return client, nil
	case <-timeoutCtx.Done():
		p.metrics.mu.Lock()
		p.metrics.acquireTimeouts++
		p.metrics.mu.Unlock()
		p.logDebug("[WARN] Client acquisition timeout - %s", p.formatMetrics())
		return nil, fmt.Errorf("timeout after %v waiting for available client: %v", p.config.AcquireTimeout, timeoutCtx.Err())
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *SDKClientPool) release(c *platformclientv2.Configuration) error {
	if c == nil {
		return fmt.Errorf("attempted to release a nil configuration ?!?")
	}
	p.metrics.recordRelease()

	// Add timeout to prevent indefinite blocking
	timeout := time.After(30 * time.Second)
	select {
	case p.Pool <- c:

		if atomic.LoadInt64(&p.metrics.activeClients) <= int64(p.config.MaxClients-PoolCriticalThreshold) {
			p.logDebug("Client released from full pool - %s", p.formatMetrics())
		}
		return nil

	case <-timeout:
		p.logDebug("Timed out attempting to release client - %s", p.formatMetrics())
		return fmt.Errorf("timeout releasing client to pool (size: %d)", p.config.MaxClients)

	default:
		// Pool is full. Don't put it back in the Pool
		p.logDebug("Attempted to release client to full pool - %s", p.formatMetrics())
		return nil
	}
}

func (p *SDKClientPool) GetMaxClients() int {
	return p.config.MaxClients
}

func cleanupConfiguration(config *platformclientv2.Configuration) error {
	// Perform any necessary cleanup of the configuration
	if config == nil {
		return nil
	}

	// Clean up logging configuration
	if config.LoggingConfiguration != nil {
		// Close any open log files
		config.LoggingConfiguration = nil
	}

	// Clean up proxy configuration
	if config.ProxyConfiguration != nil {
		// Clear sensitive data
		if config.ProxyConfiguration.Auth != nil {
			config.ProxyConfiguration.Auth.Password = ""
			config.ProxyConfiguration.Auth.UserName = ""
			config.ProxyConfiguration.Auth = nil
		}
		config.ProxyConfiguration = nil
	}

	// Clean up gateway configuration
	if config.GateWayConfiguration != nil {
		// Clear sensitive data
		if config.GateWayConfiguration.Auth != nil {
			config.GateWayConfiguration.Auth.Password = ""
			config.GateWayConfiguration.Auth.UserName = ""
			config.GateWayConfiguration.Auth = nil
		}
		// Clear path params
		if config.GateWayConfiguration.PathParams != nil {
			config.GateWayConfiguration.PathParams = nil
		}
		config.GateWayConfiguration = nil
	}

	// Clear any access tokens or sensitive data
	config.AccessToken = ""
	config.BasePath = ""

	// Clear any default headers that might contain sensitive info
	config.DefaultHeader = make(map[string]string)

	// Disable automatic token refresh
	config.AutomaticTokenRefresh = false

	// Clear retry configuration
	if config.RetryConfiguration != nil {
		config.RetryConfiguration = nil
	}
	return nil
}

func (p *SDKClientPool) Close(ctx context.Context) error {
	metrics := p.GetMetrics()
	if metrics.activeClients > 0 {
		p.logDebug("[WARN] Closing pool with %d active clients", metrics.activeClients)
	}

	if p.done == nil {
		return nil
	}
	close(p.done) // Signal all goroutines to stop
	p.done = nil

	drainCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		select {
		case c := <-p.Pool:
			if err := cleanupConfiguration(c); err != nil {
				log.Printf("[WARN] Error cleaning up configuration: %v", err)
			}
		case <-drainCtx.Done():
			return fmt.Errorf("timeout while draining client pool: %v", drainCtx.Err())
		default:
			p.logDebug("Closed SDK client pool - %s", p.formatMetrics())
			return nil
		}
	}
}

func (p *SDKClientPool) GetMetrics() poolMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	return poolMetrics{
		activeClients:   p.metrics.activeClients,
		totalAcquires:   p.metrics.totalAcquires,
		totalReleases:   p.metrics.totalReleases,
		acquireTimeouts: p.metrics.acquireTimeouts,
		lastAcquireTime: p.metrics.lastAcquireTime,
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
		clientConfig, err := SdkClientPool.acquire(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
		defer func() {
			if err := SdkClientPool.release(clientConfig); err != nil {
				log.Printf("[WARN] Error releasing client to pool: %v", err)
			}
		}()

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
		clientConfig, err := SdkClientPool.acquire(ctx)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		defer func() {
			if err := SdkClientPool.release(clientConfig); err != nil {
				log.Printf("[WARN] Error releasing client to pool: %v", err)
			}
		}()

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
		clientConfig, err := SdkClientPool.acquire(ctx)
		if err != nil {
			return nil, nil, diag.FromErr(err)
		}
		defer func() {
			if err := SdkClientPool.release(clientConfig); err != nil {
				log.Printf("[WARN] Error releasing client to pool: %v", err)
			}
		}()

		// Check if the request has been cancelled
		select {
		case <-ctx.Done():
			return nil, nil, diag.FromErr(ctx.Err()) // Error somewhere, terminate
		default:
		}

		return method(ctx, clientConfig)
	}
}
