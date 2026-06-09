package provider

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
)

// FetchPageFunc fetches a single page of entities using the supplied SDK client configuration.
type FetchPageFunc[T any] func(ctx context.Context, clientConfig *platformclientv2.Configuration, pageNum int) ([]T, *platformclientv2.APIResponse, error)

type pageResult[T any] struct {
	pageNum  int
	entities []T
	resp     *platformclientv2.APIResponse
	err      error
}

// FetchPagesConcurrently appends pages 2-N to a first page fetched by the caller.
func FetchPagesConcurrently[T any](
	ctx context.Context,
	resourceType string,
	firstPage []T,
	firstResp *platformclientv2.APIResponse,
	totalPages int,
	primaryClientConfig *platformclientv2.Configuration,
	fetchPage FetchPageFunc[T],
) ([]T, *platformclientv2.APIResponse, error) {
	if totalPages <= 1 {
		return firstPage, firstResp, nil
	}

	if SdkClientPool == nil {
		return fetchRemainingPagesSequentially(ctx, firstPage, firstResp, totalPages, primaryClientConfig, fetchPage)
	}

	maxConcurrentPages := SdkClientPool.GetMaxConcurrentPages()
	if maxConcurrentPages <= 1 {
		return fetchRemainingPagesSequentially(ctx, firstPage, firstResp, totalPages, primaryClientConfig, fetchPage)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	remainingPages := totalPages - 1
	results := make(chan pageResult[T], remainingPages)
	sem := make(chan struct{}, min(maxConcurrentPages, remainingPages))

	poolSize := SdkClientPool.GetMaxClients()
	log.Printf("[PAGINATION] Fetching pages 2-%d concurrently for %s (max_concurrent_pages=%d, token_pool_size=%d)", totalPages, resourceType, maxConcurrentPages, poolSize)

	var peakActiveClients int64
	var wg sync.WaitGroup
	for pageNum := 2; pageNum <= totalPages; pageNum++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results <- pageResult[T]{pageNum: page, resp: firstResp, err: ctx.Err()}
				return
			}

			clientConfig, err := SdkClientPool.acquire(ctx)
			if err != nil {
				results <- pageResult[T]{pageNum: page, resp: firstResp, err: err}
				return
			}
			recordPeakActiveClients(&peakActiveClients)
			defer func() {
				if err := SdkClientPool.release(clientConfig); err != nil {
					log.Printf("[WARN] Error releasing client to pool after pagination page %d: %v", page, err)
				}
			}()

			entities, resp, err := fetchPage(ctx, clientConfig, page)
			results <- pageResult[T]{pageNum: page, entities: entities, resp: resp, err: err}
		}(pageNum)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	pages := make([][]T, totalPages+1)
	var lastResp *platformclientv2.APIResponse
	var firstErr error
	var firstErrPage int
	var firstErrResp *platformclientv2.APIResponse
	for result := range results {
		if result.err != nil {
			cancel()
			if firstErr == nil {
				firstErr = result.err
				firstErrPage = result.pageNum
				firstErrResp = result.resp
			}
			continue
		}
		pages[result.pageNum] = result.entities
		lastResp = result.resp
	}
	if firstErr != nil {
		return nil, responseOrDefault(firstErrResp, firstResp), fmt.Errorf("error fetching page %d: %w", firstErrPage, firstErr)
	}

	allEntities := append([]T{}, firstPage...)
	for pageNum := 2; pageNum <= totalPages; pageNum++ {
		if pages[pageNum] == nil {
			continue
		}
		allEntities = append(allEntities, pages[pageNum]...)
	}

	log.Printf("[PAGINATION] Fetched %d resources across %d pages concurrently for %s (peak_active_clients=%d)", len(allEntities), totalPages, resourceType, atomic.LoadInt64(&peakActiveClients))
	return allEntities, responseOrDefault(lastResp, firstResp), nil
}

// recordPeakActiveClients updates peak with the pool's current active client count after an acquire.
func recordPeakActiveClients(peak *int64) {
	if SdkClientPool == nil || SdkClientPool.metrics == nil {
		return
	}
	active := atomic.LoadInt64(&SdkClientPool.metrics.activeClients)
	for {
		current := atomic.LoadInt64(peak)
		if active <= current {
			return
		}
		if atomic.CompareAndSwapInt64(peak, current, active) {
			return
		}
	}
}

func fetchRemainingPagesSequentially[T any](
	ctx context.Context,
	firstPage []T,
	firstResp *platformclientv2.APIResponse,
	totalPages int,
	clientConfig *platformclientv2.Configuration,
	fetchPage FetchPageFunc[T],
) ([]T, *platformclientv2.APIResponse, error) {
	allEntities := append([]T{}, firstPage...)
	lastResp := firstResp

	for pageNum := 2; pageNum <= totalPages; pageNum++ {
		entities, resp, err := fetchPage(ctx, clientConfig, pageNum)
		if err != nil {
			return nil, responseOrDefault(resp, firstResp), fmt.Errorf("error fetching page %d: %w", pageNum, err)
		}
		allEntities = append(allEntities, entities...)
		lastResp = resp
	}

	return allEntities, responseOrDefault(lastResp, firstResp), nil
}

func responseOrDefault(resp, defaultResp *platformclientv2.APIResponse) *platformclientv2.APIResponse {
	if resp != nil {
		return resp
	}
	return defaultResp
}
