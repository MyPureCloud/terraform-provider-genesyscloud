package provider

import (
	"bytes"
	"context"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestSDKClientPool_InitAndAcquire(t *testing.T) {
	// Reset the singleton for testing
	SdkClientPool = nil
	SdkClientPoolErr = nil
	Once = sync.Once{}

	// Create ResourceData with the schema
	providerConfig := testProviderConfig(t)
	ctx := context.Background()

	diagErr := InitSDKClientPool(ctx, "test", providerConfig)
	assert.Nil(t, diagErr, "Expected no error initializing pool")
	assert.NotNil(t, SdkClientPool, "Expected pool to be initialized")

	// Test acquiring clients
	client1, err := SdkClientPool.acquire(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, client1)
	metrics := SdkClientPool.GetMetrics()
	assert.Equal(t, int64(1), metrics.activeClients)

	client2, err := SdkClientPool.acquire(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, client2)
	metrics = SdkClientPool.GetMetrics()
	assert.Equal(t, int64(2), metrics.activeClients)

	// Release a client
	SdkClientPool.release(client1)
	metrics = SdkClientPool.GetMetrics()
	assert.Equal(t, int64(1), metrics.activeClients)

	// Acquire the released client
	client3, err := SdkClientPool.acquire(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, client3)
	metrics = SdkClientPool.GetMetrics()
	assert.Equal(t, int64(2), metrics.activeClients)
}

func TestSDKClientPool_AcquireTimeout(t *testing.T) {
	// Reset the singleton for testing
	SdkClientPool = nil
	SdkClientPoolErr = nil
	Once = sync.Once{}

	// Create ResourceData with the schema
	providerConfig := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize:       1,
		AttrTokenAcquireTimeout: "100ms",
		AttrTokenInitTimeout:    "5s",
	})
	ctx := context.Background()

	diagErr := InitSDKClientPool(ctx, "test", providerConfig)
	assert.Nil(t, diagErr, "Expected no error initializing pool")
	assert.NotNil(t, SdkClientPool, "Expected pool to be initialized")

	// Acquire the only client
	client1, err := SdkClientPool.acquire(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, client1)

	// Try to acquire another client (should timeout)
	client2, err := SdkClientPool.acquire(ctx)
	assert.Error(t, err, "Expected timeout error")
	assert.Nil(t, client2)
	assert.Contains(t, err.Error(), "timeout")
}

func TestSDKClientPool_Metrics(t *testing.T) {
	// Reset the singleton for testing
	SdkClientPool = nil
	SdkClientPoolErr = nil
	Once = sync.Once{}

	// Create ResourceData with the schema
	providerConfig := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize:       2,
		AttrTokenAcquireTimeout: "1s",
		AttrTokenInitTimeout:    "5s",
	})
	ctx := context.Background()

	diagErr := InitSDKClientPool(ctx, "test", providerConfig)
	assert.Nil(t, diagErr, "Expected no error initializing pool")
	assert.NotNil(t, SdkClientPool, "Expected pool to be initialized")

	// Test initial metrics
	metrics := SdkClientPool.GetMetrics()
	assert.Equal(t, int64(0), metrics.activeClients)
	assert.Equal(t, int64(0), metrics.totalAcquires)
	assert.Equal(t, int64(0), metrics.totalReleases)
	assert.Equal(t, int64(0), metrics.acquireTimeouts)
	assert.True(t, metrics.lastAcquireTime.IsZero())

	// Test metrics formatting with no activity
	formatted := SdkClientPool.formatMetrics()
	assert.Contains(t, formatted, "Active: 0/2")
	assert.Contains(t, formatted, "Acquires: 0")
	assert.Contains(t, formatted, "Releases: 0")
	assert.Contains(t, formatted, "Timeouts: 0")
	assert.Contains(t, formatted, "Last Acquire: never")

	// Test metrics after a successful acquire
	client1, err := SdkClientPool.acquire(ctx)
	assert.Nil(t, err)

	metrics = SdkClientPool.GetMetrics()
	assert.Equal(t, int64(1), metrics.activeClients)
	assert.Equal(t, int64(1), metrics.totalAcquires)
	assert.Equal(t, int64(0), metrics.totalReleases)
	assert.False(t, metrics.lastAcquireTime.IsZero())

	// Test metrics after release
	SdkClientPool.release(client1)
	metrics = SdkClientPool.GetMetrics()
	assert.Equal(t, int64(0), metrics.activeClients)
	assert.Equal(t, int64(1), metrics.totalAcquires)
	assert.Equal(t, int64(1), metrics.totalReleases)

	// Test metrics with timeout
	_, _ = SdkClientPool.acquire(ctx)
	_, _ = SdkClientPool.acquire(ctx)
	_, err = SdkClientPool.acquire(ctx) // Should timeout
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")

	// Check metrics
	metrics = SdkClientPool.GetMetrics()
	assert.Equal(t, int64(2), metrics.activeClients)
	assert.Equal(t, int64(3), metrics.totalAcquires)
	assert.Equal(t, int64(1), metrics.totalReleases)
	assert.Equal(t, int64(1), metrics.acquireTimeouts)
	assert.NotZero(t, metrics.lastAcquireTime)

	// Test final metrics formatting
	formatted = SdkClientPool.formatMetrics()
	assert.Contains(t, formatted, "Active: 2/2")
	assert.Contains(t, formatted, "Acquires: 3")
	assert.Contains(t, formatted, "Releases: 1")
	assert.Contains(t, formatted, "Timeouts: 1")
	assert.NotContains(t, formatted, "Last Acquire: never")
}

func TestSDKClientPool_ConcurrentOperations(t *testing.T) {
	// Reset the singleton for testing
	SdkClientPool = nil
	SdkClientPoolErr = nil
	Once = sync.Once{}

	// Create ResourceData with the schema
	providerConfig := testProviderConfig(t)
	ctx := context.Background()

	err := InitSDKClientPool(ctx, "test", providerConfig)
	assert.Nil(t, err)

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := SdkClientPool.acquire(ctx)
			if err == nil {
				time.Sleep(100 * time.Millisecond) // Simulate work
				SdkClientPool.release(client)
			}
		}()
	}

	wg.Wait()

	metrics := SdkClientPool.GetMetrics()
	assert.Equal(t, int64(0), metrics.activeClients)
	assert.True(t, metrics.totalReleases <= metrics.totalAcquires)
}

func TestSDKClientPool_PreFill(t *testing.T) {
	tests := []struct {
		name        string
		maxClients  int
		initTimeout string
		wantErr     bool
	}{
		{
			name:        "successful_prefill",
			maxClients:  5,
			initTimeout: "10s",
			wantErr:     false,
		},
		{
			name:        "timeout_prefill",
			maxClients:  20,    // Maximum allowed clients
			initTimeout: "1ns", // Set to essentially immediate timeout
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetClientPool()

			// Create a buffer to capture logs
			var logBuffer bytes.Buffer
			log.SetOutput(&logBuffer)
			defer log.SetOutput(os.Stdout)

			// Create ResourceData with the schema
			providerConfig := testProviderConfigCustom(t, map[string]interface{}{
				AttrTokenPoolSize:       tt.maxClients,
				AttrTokenAcquireTimeout: "1s",
				AttrSdkClientPoolDebug:  true,
				AttrTokenInitTimeout:    tt.initTimeout,
			})

			// Initialize the pool
			ctx := context.Background()

			diagErr := InitSDKClientPool(ctx, "test", providerConfig)
			t.Logf("Pool initialization completed with error: %v", diagErr)
			if (diagErr != nil) != tt.wantErr {
				t.Errorf("InitSDKClientPool() error = %v, wantErr %v", diagErr, tt.wantErr)
				return
			}

			// Verify pool was initialized
			assert.NotNil(t, SdkClientPool, "Expected pool to be initialized")

			if !tt.wantErr {
				// Test concurrent access
				ctx := context.Background()
				var wg sync.WaitGroup
				for i := 0; i < 5; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						client, err := SdkClientPool.acquire(ctx)
						if err == nil {
							defer SdkClientPool.release(client)
							// Simulate some work
							time.Sleep(10 * time.Millisecond)
						}
					}()
				}
				wg.Wait()

				// Verify metrics
				metrics := SdkClientPool.GetMetrics()
				assert.Equal(t, int64(0), metrics.activeClients, "Expected 0 active clients")
				assert.Equal(t, metrics.totalAcquires, metrics.totalReleases,
					"Mismatch between acquires and releases")

				// Verify logging
				logs := logBuffer.String()
				if !strings.Contains(logs, "Successfully pre-filled client pool") {
					t.Error("Expected success message in logs")
					t.Logf("Actual logs: %s", logs)
				}
			}

			// Cleanup
			if SdkClientPool != nil {
				err := SdkClientPool.Close(context.Background())
				assert.NoError(t, err, "Failed to close pool")
			}
		})
	}
}

func TestSDKClientPool_RunWithPooledClient(t *testing.T) {
	// Reset the singleton for testing
	SdkClientPool = nil
	SdkClientPoolErr = nil
	Once = sync.Once{}

	// Create ResourceData with the schema
	providerConfig := testProviderConfig(t)
	ctx := context.Background()

	diagErr := InitSDKClientPool(ctx, "test", providerConfig)
	assert.Nil(t, diagErr, "Expected no error initializing pool")
	assert.NotNil(t, SdkClientPool, "Expected pool to be initialized")

	// Create a test method
	testMethod := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		meta := m.(*ProviderMeta)
		assert.NotNil(t, meta.ClientConfig)
		return nil
	}

	// Test the wrapper
	wrappedMethod := runWithPooledClient(testMethod)
	diags := wrappedMethod(context.Background(), &schema.ResourceData{}, &ProviderMeta{})
	assert.Nil(t, diags)
}

func TestSDKClientPool_ContextCancellation(t *testing.T) {
	// Reset the singleton for testing
	SdkClientPool = nil
	SdkClientPoolErr = nil
	Once = sync.Once{}

	// Create ResourceData with the schema
	providerConfig := testProviderConfig(t)
	ctx := context.Background()

	diagErr := InitSDKClientPool(ctx, "test", providerConfig)
	assert.Nil(t, diagErr, "Expected no error initializing pool")
	assert.NotNil(t, SdkClientPool, "Expected pool to be initialized")

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Test the wrapper with cancelled context
	testMethod := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		return nil
	}

	wrappedMethod := runWithPooledClient(testMethod)
	diags := wrappedMethod(ctx, &schema.ResourceData{}, &ProviderMeta{})
	assert.NotNil(t, diags)
	assert.Contains(t, diags[0].Summary, "context canceled")
}

func TestSDKClientPool_LoggingOutput(t *testing.T) {
	resetClientPool()

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	// Restore default logger output after test
	defer log.SetOutput(os.Stderr)

	// Test with debug logging enabled and pool near capacity
	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize:       4,
		AttrTokenAcquireTimeout: "1s",
		AttrTokenInitTimeout:    "5s",
		AttrSdkClientPoolDebug:  true,
	})
	ctx := context.Background()

	err := InitSDKClientPool(ctx, "test", config)
	assert.Nil(t, err)

	// Acquire one client to trigger debug msg
	client1, _ := SdkClientPool.acquire(ctx)
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Client acquired from pool")
	assert.NotContains(t, logOutput, "pool near capacity")
	assert.NotContains(t, logOutput, "pool at critical capacity")

	// Clear buffer for next test
	buf.Reset()

	// Acquire two clients to trigger near capacity msg
	_, _ = SdkClientPool.acquire(ctx)

	// Should trigger near capacity warning (2 of 4 clients used)
	logOutput = buf.String()
	assert.Contains(t, logOutput, "Client acquired from pool")
	assert.Contains(t, logOutput, "pool near capacity")
	assert.NotContains(t, logOutput, "pool at critical capacity")

	// Clear buffer for next test
	buf.Reset()

	// Acquire one more to trigger critical capacity
	_, _ = SdkClientPool.acquire(ctx)

	logOutput = buf.String()
	assert.Contains(t, logOutput, "Client acquired from pool")
	assert.Contains(t, logOutput, "pool at critical capacity")
	assert.NotContains(t, logOutput, "pool near capacity")

	buf.Reset()

	// Release one client to trigger release debug msg
	SdkClientPool.release(client1)

	logOutput = buf.String()
	assert.Contains(t, logOutput, "Client released from full pool - ") // Debug log

	// Test with debug logging disabled
	resetClientPool()
	buf.Reset()

	config = testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize:       4,
		AttrTokenAcquireTimeout: "1s",
		AttrTokenInitTimeout:    "5s",
		AttrSdkClientPoolDebug:  false,
	})

	err = InitSDKClientPool(ctx, "test", config)
	assert.Nil(t, err)

	// Same operations should not generate debug logs
	client1, _ = SdkClientPool.acquire(ctx)
	_, _ = SdkClientPool.acquire(ctx)
	_, _ = SdkClientPool.acquire(ctx)
	SdkClientPool.release(client1)

	logOutput = buf.String()
	assert.NotContains(t, logOutput, "Client acquired from pool - ")      // Debug log
	assert.NotContains(t, logOutput, "Client released from full pool - ") // Debug log
	assert.NotContains(t, logOutput, "pool near capacity")
	assert.NotContains(t, logOutput, "pool at critical capacity")
}

func TestSDKClientPool_InitializationLogging(t *testing.T) {
	resetClientPool()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize:       3,
		AttrTokenAcquireTimeout: "1s",
		AttrTokenInitTimeout:    "5s",
	})
	ctx := context.Background()

	err := InitSDKClientPool(ctx, "test", config)
	assert.Nil(t, err)

	logOutput := buf.String()

	// Should see one default client initialization
	defaultClientLogs := strings.Count(logOutput, "Initializing default SDK client")
	assert.Equal(t, 1, defaultClientLogs, "Expected exactly one default client initialization log")

	// Should see one pool client initializations (poolSize = 2)
	poolClientLogs := strings.Count(logOutput, "Setting access token set on configuration instance")
	assert.Equal(t, 1, poolClientLogs, "Expected exactly one default setting access token logs")
}

func TestSDKClientPool_TimeoutLogging(t *testing.T) {
	resetClientPool()

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// Test with debug logging enabled
	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize:       20,
		AttrTokenAcquireTimeout: "1s",
		AttrTokenInitTimeout:    "1ns",
		AttrSdkClientPoolDebug:  true,
	})
	ctx := context.Background()

	// Initialize pool (should timeout)
	diagErr := InitSDKClientPool(ctx, "test", config)

	// Verify error and logs
	assert.NotNil(t, diagErr, "Expected timeout error")

	// Check diagErr for error message
	assert.Contains(t, diagErr[0].Summary, "Timed out pre-filling client pool")

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Timed out pre-filling client pool")

	// Clear buffer and test with debug logging disabled
	buf.Reset()
	resetClientPool()

	config = testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize:       20,
		AttrTokenAcquireTimeout: "1s",
		AttrTokenInitTimeout:    "1ns",
		AttrSdkClientPoolDebug:  false,
	})

	ctx2 := context.Background()
	diagErr = InitSDKClientPool(ctx2, "test", config)
	assert.NotNil(t, diagErr, "Expected timeout error")

	// Check diagErr for error message
	assert.Contains(t, diagErr[0].Summary, "Timed out pre-filling client pool")

	logOutput = buf.String()
	// Should not see error logged with debug disabled
	assert.NotContains(t, logOutput, "Timed out pre-filling client pool")
}

func TestSDKClientPool_WrapperFunctions(t *testing.T) {
	resetClientPool()

	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize: 2,
	})
	ctx := context.Background()

	err := InitSDKClientPool(ctx, "test", config)
	assert.Nil(t, err)

	// Mock resource data
	d := schema.TestResourceDataRaw(t, ProviderSchema(), nil)
	meta := &ProviderMeta{}

	// Test CreateWithPooledClient
	createFunc := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		// Verify we got a client config
		meta := m.(*ProviderMeta)
		assert.NotNil(t, meta.ClientConfig)
		return nil
	}
	wrappedCreate := CreateWithPooledClient(createFunc)
	diags := wrappedCreate(context.Background(), d, meta)
	assert.Nil(t, diags)

	// Test ReadWithPooledClient
	readFunc := func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		meta := m.(*ProviderMeta)
		assert.NotNil(t, meta.ClientConfig)
		return nil
	}
	wrappedRead := ReadWithPooledClient(readFunc)
	diags = wrappedRead(context.Background(), d, meta)
	assert.Nil(t, diags)

	// Verify client was properly released after each operation
	metrics := SdkClientPool.GetMetrics()
	assert.Equal(t, int64(0), metrics.activeClients)
	assert.Equal(t, metrics.totalAcquires, metrics.totalReleases)
}

func TestSDKClientPool_GetAllWithPooledClient(t *testing.T) {
	resetClientPool()

	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize: 2,
	})
	ctx := context.Background()

	err := InitSDKClientPool(ctx, "test", config)
	assert.Nil(t, err)

	// Test GetAllWithPooledClient
	getAllFunc := func(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
		assert.NotNil(t, clientConfig)
		return make(resourceExporter.ResourceIDMetaMap), nil
	}

	wrapped := GetAllWithPooledClient(getAllFunc)
	result, diags := wrapped(context.Background())
	assert.Nil(t, diags)
	assert.NotNil(t, result)

	// Test GetAllWithPooledClientCustom
	getAllCustomFunc := func(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, diag.Diagnostics) {
		assert.NotNil(t, clientConfig)
		return make(resourceExporter.ResourceIDMetaMap), nil, nil
	}

	wrappedCustom := GetAllWithPooledClientCustom(getAllCustomFunc)
	resultCustom, dep, diags := wrappedCustom(context.Background())
	assert.Nil(t, diags)
	assert.NotNil(t, resultCustom)
	assert.Nil(t, dep)

	// Verify clients were released
	metrics := SdkClientPool.GetMetrics()
	assert.Equal(t, int64(0), metrics.activeClients)
	assert.Equal(t, metrics.totalAcquires, metrics.totalReleases)
}

func TestSDKClientPool_GetAllWithPooledClientCustom(t *testing.T) {
	resetClientPool()

	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize: 2,
	})
	ctx := context.Background()

	err := InitSDKClientPool(ctx, "test", config)
	assert.Nil(t, err)

	// Test GetAllWithPooledClientCustom
	getAllCustomFunc := func(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, diag.Diagnostics) {
		assert.NotNil(t, clientConfig)
		return make(resourceExporter.ResourceIDMetaMap), &resourceExporter.DependencyResource{}, nil
	}

	wrappedCustom := GetAllWithPooledClientCustom(getAllCustomFunc)
	resultCustom, dep, diags := wrappedCustom(context.Background())
	assert.Nil(t, diags)
	assert.NotNil(t, resultCustom)
	assert.NotNil(t, dep)

	// Verify client was released
	metrics := SdkClientPool.GetMetrics()
	assert.Equal(t, int64(0), metrics.activeClients)
	assert.Equal(t, metrics.totalAcquires, metrics.totalReleases)
}

// Test error cases
func TestSDKClientPool_GetAllWithPooledClientErrors(t *testing.T) {
	resetClientPool()

	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize: 2,
	})
	ctx := context.Background()

	err := InitSDKClientPool(ctx, "test", config)
	assert.Nil(t, err)

	// Test error from getAllFunc
	getAllFunc := func(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
		return nil, diag.Errorf("test error")
	}

	wrapped := GetAllWithPooledClient(getAllFunc)
	result, diags := wrapped(context.Background())
	assert.NotNil(t, diags)
	assert.Nil(t, result)
	assert.Contains(t, diags[0].Summary, "test error")

	// Test error from getAllCustomFunc
	getAllCustomFunc := func(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, *resourceExporter.DependencyResource, diag.Diagnostics) {
		return nil, nil, diag.Errorf("test custom error")
	}

	wrappedCustom := GetAllWithPooledClientCustom(getAllCustomFunc)
	resultCustom, dep, diags := wrappedCustom(context.Background())
	assert.NotNil(t, diags)
	assert.Nil(t, resultCustom)
	assert.Nil(t, dep)
	assert.Contains(t, diags[0].Summary, "test custom error")

	// Verify clients were released even with errors
	metrics := SdkClientPool.GetMetrics()
	assert.Equal(t, int64(0), metrics.activeClients)
	assert.Equal(t, metrics.totalAcquires, metrics.totalReleases)
}

// Test context cancellation
func TestSDKClientPool_GetAllWithPooledClientContext(t *testing.T) {
	resetClientPool()

	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize: 2,
	})
	ctx := context.Background()

	err := InitSDKClientPool(ctx, "test", config)
	assert.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	getAllFunc := func(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
		return nil, nil
	}

	wrapped := GetAllWithPooledClient(getAllFunc)
	result, diags := wrapped(ctx)
	assert.NotNil(t, diags)
	assert.Nil(t, result)
	assert.Contains(t, diags[0].Summary, "context canceled")

	// Verify client was released
	metrics := SdkClientPool.GetMetrics()
	assert.Equal(t, int64(0), metrics.activeClients)
}

func TestSDKClientPool_SchemaValidation(t *testing.T) {
	schema := ProviderSchema()

	tests := []struct {
		name          string
		attribute     string
		value         interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:      "Valid pool size",
			attribute: AttrTokenPoolSize,
			value:     DefaultMaxClients,
		},
		{
			name:      "Minimum pool size",
			attribute: AttrTokenPoolSize,
			value:     MinClients,
		},
		{
			name:      "Maximum pool size",
			attribute: AttrTokenPoolSize,
			value:     MaxClients,
		},
		{
			name:          "Pool size too small",
			attribute:     AttrTokenPoolSize,
			value:         MinClients - 1,
			expectError:   true,
			errorContains: "expected token_pool_size to be in the range",
		},
		{
			name:          "Pool size too large",
			attribute:     AttrTokenPoolSize,
			value:         MaxClients + 1,
			expectError:   true,
			errorContains: "expected token_pool_size to be in the range",
		},
		{
			name:          "Pool size wrong type",
			attribute:     AttrTokenPoolSize,
			value:         "not a number",
			expectError:   true,
			errorContains: "expected type of token_pool_size to be integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the schema for this attribute
			attrSchema := schema[tt.attribute]
			assert.NotNil(t, attrSchema, "Schema should exist for attribute")

			// Validate the value
			warns, errs := attrSchema.ValidateFunc(tt.value, tt.attribute)

			if tt.expectError {
				assert.NotEmpty(t, errs)
				assert.Contains(t, errs[0].Error(), tt.errorContains)
			} else {
				assert.Empty(t, errs)
			}
			assert.Empty(t, warns) // We don't expect any warnings
		})
	}
}

// resetClientPool resets the singleton for testing
func resetClientPool() {
	SdkClientPool = nil
	SdkClientPoolErr = nil
	Once = sync.Once{}
}
