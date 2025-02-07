package provider

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestSDKClientPool_InitAndAcquire(t *testing.T) {
	// Reset the singleton for testing
	SdkClientPool = nil
	SdkClientPoolErr = nil
	Once = sync.Once{}

	// Create ResourceData with the schema
	providerConfig := testProviderConfig(t)

	diagErr := InitSDKClientPool("test", providerConfig)
	assert.Nil(t, diagErr, "Expected no error initializing pool")
	assert.NotNil(t, SdkClientPool, "Expected pool to be initialized")

	// Test acquiring clients
	client1, err := SdkClientPool.acquire()
	assert.Nil(t, err)
	assert.NotNil(t, client1)
	assert.Equal(t, int64(1), atomic.LoadInt64(&SdkClientPool.activeClients))

	client2, err := SdkClientPool.acquire()
	assert.Nil(t, err)
	assert.NotNil(t, client2)
	assert.Equal(t, int64(2), atomic.LoadInt64(&SdkClientPool.activeClients))

	// Release a client
	SdkClientPool.release(client1)
	assert.Equal(t, int64(1), atomic.LoadInt64(&SdkClientPool.activeClients))

	// Acquire the released client
	client3, err := SdkClientPool.acquire()
	assert.Nil(t, err)
	assert.NotNil(t, client3)
	assert.Equal(t, int64(2), atomic.LoadInt64(&SdkClientPool.activeClients))
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

	diagErr := InitSDKClientPool("test", providerConfig)
	assert.Nil(t, diagErr, "Expected no error initializing pool")
	assert.NotNil(t, SdkClientPool, "Expected pool to be initialized")

	// Acquire the only client
	client1, err := SdkClientPool.acquire()
	assert.Nil(t, err)
	assert.NotNil(t, client1)

	// Try to acquire another client (should timeout)
	client2, err := SdkClientPool.acquire()
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

	diagErr := InitSDKClientPool("test", providerConfig)
	assert.Nil(t, diagErr, "Expected no error initializing pool")
	assert.NotNil(t, SdkClientPool, "Expected pool to be initialized")

	// Test initial metrics
	metrics := SdkClientPool.GetMetrics()
	assert.Equal(t, int64(0), metrics["active_clients"])
	assert.Equal(t, int64(0), metrics["total_acquires"])
	assert.Equal(t, int64(0), metrics["total_releases"])
	assert.Equal(t, int64(0), metrics["acquire_timeouts"])
	assert.True(t, metrics["last_acquire_time"].(time.Time).IsZero())

	// Test metrics formatting with no activity
	formatted := SdkClientPool.formatMetrics()
	assert.Contains(t, formatted, "Active: 0/2")
	assert.Contains(t, formatted, "Acquires: 0")
	assert.Contains(t, formatted, "Releases: 0")
	assert.Contains(t, formatted, "Timeouts: 0")
	assert.Contains(t, formatted, "Last Acquire: never")

	// Test metrics after a successful acquire
	client1, err := SdkClientPool.acquire()
	assert.Nil(t, err)

	metrics = SdkClientPool.GetMetrics()
	assert.Equal(t, int64(1), metrics["active_clients"])
	assert.Equal(t, int64(1), metrics["total_acquires"])
	assert.Equal(t, int64(0), metrics["total_releases"])
	assert.False(t, metrics["last_acquire_time"].(time.Time).IsZero())

	// Test metrics after release
	SdkClientPool.release(client1)
	metrics = SdkClientPool.GetMetrics()
	assert.Equal(t, int64(0), metrics["active_clients"])
	assert.Equal(t, int64(1), metrics["total_acquires"])
	assert.Equal(t, int64(1), metrics["total_releases"])

	// Test metrics with timeout
	_, _ = SdkClientPool.acquire()
	_, _ = SdkClientPool.acquire()
	_, err = SdkClientPool.acquire() // Should timeout
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")

	// Check metrics
	metrics = SdkClientPool.GetMetrics()
	assert.Equal(t, int64(2), metrics["active_clients"].(int64))
	assert.Equal(t, int64(3), metrics["total_acquires"].(int64))
	assert.Equal(t, int64(1), metrics["total_releases"].(int64))
	assert.Equal(t, int64(1), metrics["acquire_timeouts"].(int64))
	assert.NotZero(t, metrics["last_acquire_time"].(time.Time))

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

	err := InitSDKClientPool("test", providerConfig)
	assert.Nil(t, err)

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := SdkClientPool.acquire()
			if err == nil {
				time.Sleep(100 * time.Millisecond) // Simulate work
				SdkClientPool.release(client)
			}
		}()
	}

	wg.Wait()

	metrics := SdkClientPool.GetMetrics()
	assert.Equal(t, int64(0), metrics["active_clients"].(int64))
	assert.True(t, metrics["total_releases"].(int64) <= metrics["total_acquires"].(int64))
}

func TestRunWithPooledClient(t *testing.T) {
	// Reset the singleton for testing
	SdkClientPool = nil
	SdkClientPoolErr = nil
	Once = sync.Once{}

	// Create ResourceData with the schema
	providerConfig := testProviderConfig(t)

	diagErr := InitSDKClientPool("test", providerConfig)
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

	diagErr := InitSDKClientPool("test", providerConfig)
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

	err := InitSDKClientPool("test", config)
	assert.Nil(t, err)

	// Acquire one client to trigger debug msg
	client1, _ := SdkClientPool.acquire()
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Client acquired from pool")
	assert.NotContains(t, logOutput, "pool near capacity")
	assert.NotContains(t, logOutput, "pool at critical capacity")

	// Clear buffer for next test
	buf.Reset()

	// Acquire two clients to trigger near capacity msg
	_, _ = SdkClientPool.acquire()

	// Should trigger near capacity warning (2 of 4 clients used)
	logOutput = buf.String()
	assert.Contains(t, logOutput, "Client acquired from pool")
	assert.Contains(t, logOutput, "pool near capacity")
	assert.NotContains(t, logOutput, "pool at critical capacity")

	// Clear buffer for next test
	buf.Reset()

	// Acquire one more to trigger critical capacity
	_, _ = SdkClientPool.acquire()

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

	err = InitSDKClientPool("test", config)
	assert.Nil(t, err)

	// Same operations should not generate debug logs
	client1, _ = SdkClientPool.acquire()
	_, _ = SdkClientPool.acquire()
	_, _ = SdkClientPool.acquire()
	SdkClientPool.release(client1)

	logOutput = buf.String()
	assert.NotContains(t, logOutput, "Client acquired from pool - ")      // Debug log
	assert.NotContains(t, logOutput, "Client released from full pool - ") // Debug log
	assert.Contains(t, logOutput, "pool near capacity")
	assert.Contains(t, logOutput, "pool at critical capacity")
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

	err := InitSDKClientPool("test", config)
	assert.Nil(t, err)

	logOutput := buf.String()

	// Should see one default client initialization
	defaultClientLogs := strings.Count(logOutput, "Initializing default SDK client")
	assert.Equal(t, 1, defaultClientLogs, "Expected exactly one default client initialization log")

	// Should see one pool client initializations (poolSize = 2)
	poolClientLogs := strings.Count(logOutput, "Setting access token set on configuration instance")
	assert.Equal(t, 1, poolClientLogs, "Expected exactly one default setting access token logs")
}

func TestSDKClientPool_WrapperFunctions(t *testing.T) {
	resetClientPool()

	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize: 2,
	})

	err := InitSDKClientPool("test", config)
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
	assert.Equal(t, int64(0), metrics["active_clients"])
	assert.Equal(t, metrics["total_acquires"], metrics["total_releases"])
}

func TestSDKClientPool_GetAllWithPooledClient(t *testing.T) {
	resetClientPool()

	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize: 2,
	})

	err := InitSDKClientPool("test", config)
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
	assert.Equal(t, int64(0), metrics["active_clients"])
	assert.Equal(t, metrics["total_acquires"], metrics["total_releases"])
}

func TestSDKClientPool_GetAllWithPooledClientCustom(t *testing.T) {
	resetClientPool()

	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize: 2,
	})

	err := InitSDKClientPool("test", config)
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
	assert.Equal(t, int64(0), metrics["active_clients"])
	assert.Equal(t, metrics["total_acquires"], metrics["total_releases"])
}

// Test error cases
func TestSDKClientPool_GetAllWithPooledClientErrors(t *testing.T) {
	resetClientPool()

	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize: 2,
	})

	err := InitSDKClientPool("test", config)
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
	assert.Equal(t, int64(0), metrics["active_clients"])
	assert.Equal(t, metrics["total_acquires"], metrics["total_releases"])
}

// Test context cancellation
func TestSDKClientPool_GetAllWithPooledClientContext(t *testing.T) {
	resetClientPool()

	config := testProviderConfigCustom(t, map[string]interface{}{
		AttrTokenPoolSize: 2,
	})

	err := InitSDKClientPool("test", config)
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
	assert.Equal(t, int64(0), metrics["active_clients"])
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
