package provider

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchPagesConcurrentlyPreservesPageOrder(t *testing.T) {
	resetClientPool()
	defer resetClientPool()

	SdkClientPool = newTestPaginationPool(2)

	results, resp, err := FetchPagesConcurrently(
		context.Background(),
		"test_resource",
		[]int{1},
		&platformclientv2.APIResponse{},
		3,
		&platformclientv2.Configuration{},
		func(ctx context.Context, clientConfig *platformclientv2.Configuration, pageNum int) ([]int, *platformclientv2.APIResponse, error) {
			require.NotNil(t, clientConfig)
			if pageNum == 2 {
				time.Sleep(10 * time.Millisecond)
			}
			return []int{pageNum}, &platformclientv2.APIResponse{}, nil
		},
	)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, []int{1, 2, 3}, results)

	metrics := SdkClientPool.GetMetrics()
	assert.Equal(t, int64(0), metrics.activeClients)
	assert.Equal(t, metrics.totalAcquires, metrics.totalReleases)
}

func TestFetchPagesConcurrentlyFallsBackWithoutPool(t *testing.T) {
	resetClientPool()
	defer resetClientPool()

	var mu sync.Mutex
	var fetchedPages []int
	primaryClientConfig := &platformclientv2.Configuration{}

	results, _, err := FetchPagesConcurrently(
		context.Background(),
		"test_resource",
		[]int{1},
		&platformclientv2.APIResponse{},
		3,
		primaryClientConfig,
		func(ctx context.Context, clientConfig *platformclientv2.Configuration, pageNum int) ([]int, *platformclientv2.APIResponse, error) {
			assert.Same(t, primaryClientConfig, clientConfig)
			mu.Lock()
			fetchedPages = append(fetchedPages, pageNum)
			mu.Unlock()
			return []int{pageNum}, &platformclientv2.APIResponse{}, nil
		},
	)

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, results)
	assert.Equal(t, []int{2, 3}, fetchedPages)
}

func TestFetchPagesConcurrentlyReturnsPageErrors(t *testing.T) {
	resetClientPool()
	defer resetClientPool()

	SdkClientPool = newTestPaginationPool(2)

	_, _, err := FetchPagesConcurrently(
		context.Background(),
		"test_resource",
		[]int{1},
		&platformclientv2.APIResponse{},
		3,
		&platformclientv2.Configuration{},
		func(ctx context.Context, clientConfig *platformclientv2.Configuration, pageNum int) ([]int, *platformclientv2.APIResponse, error) {
			if pageNum == 2 {
				return nil, &platformclientv2.APIResponse{}, fmt.Errorf("boom")
			}
			return []int{pageNum}, &platformclientv2.APIResponse{}, nil
		},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error fetching page 2")
}

func newTestPaginationPool(maxClients int) *SDKClientPool {
	pool := &SDKClientPool{
		Pool: make(chan *platformclientv2.Configuration, maxClients),
		config: &SDKClientPoolConfig{
			MaxClients:         maxClients,
			MaxConcurrentPages: maxClients,
			AcquireTimeout:     time.Second,
		},
		metrics: &poolMetrics{},
		done:    make(chan struct{}),
		ctx:     context.Background(),
	}

	for i := 0; i < maxClients; i++ {
		pool.Pool <- &platformclientv2.Configuration{}
	}

	return pool
}
