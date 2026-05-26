# Implement concurrent pagination for CX as Code resource export

### Description

Based on findings from MRMO-296, the CX as Code exporter currently fetches paginated resources sequentially using a single SDK client. For resources with many pages (e.g., 101 pages for 10,041 skills, 500+ pages for 50,000 users), this creates a significant bottleneck that limits export throughput to 1-2 requests per second during the pagination phase.

Testing showed that skills export took 46 seconds sequentially. With concurrent pagination using 50 pooled SDK clients, the same export completed in 26 seconds (1.77x speedup). For resources with more pages (e.g., users with 500+ pages), the speedup will be even more significant.

This ticket implements concurrent pagination (Phase 1 optimization) where pages 2-N are fetched in parallel using pooled SDK clients, controlled by a configurable max_concurrent_pages parameter.

Key Requirement: The solution should be implemented as a shared, reusable helper function that can be applied across all ~150 resource types in the provider, avoiding code duplication.

### User Story
As a reconciliation service operator, I want resource pagination to execute concurrently so that large organization exports complete in minutes instead of hours.

### Acceptance Criteria

* Implement shared FetchPagesConcurrently() helper function in sdk_client_pool.go or similar central location
* Helper function should be generic and reusable across all resource types
* Add max_concurrent_pages configuration parameter to provider config (default: 50, max: 75 due to OAuth rate limits)
* Update high-priority resource proxies to use the shared helper (user, routing_queue, routing_skill, group, architect_flow)
* Pagination fetches page 1 sequentially to determine total pages, then fetches remaining pages concurrently
* Each page fetch acquires its own pooled SDK client from the token pool
* Semaphore limits concurrent page fetches to configured maximum
* Skills export (101 pages) completes in under 30 seconds (vs baseline 46 seconds)
* Instrumentation logs show 40-50 active clients during pagination phase
* Documentation for how to apply the pattern to remaining resource types

## PoC Implementation Reference

Note: The following pseudo code is from the MRMO-296 PoC that validated the concurrent pagination approach. The final implementation should abstract this pattern into a reusable helper function that works across all resource types.

Example from PoC (routing_skill proxy):

```go
func getAllResourcesFn(ctx context.Context, p *resourceProxy, filters string) (*[]Resource, *platformclientv2.APIResponse, error) {
    ctx = provider.EnsureResourceContext(ctx, ResourceType)
    const pageSize = 100
    var allResources []Resource
    // 1. Fetch first page to get total page count
    firstPage, resp, err := p.api.GetResources(pageSize, 1, filters)
    if err != nil {
        return nil, resp, err
    }
    if firstPage.Entities == nil || len(*firstPage.Entities) == 0 {
        return &allResources, resp, nil
    }
    allResources = append(allResources, *firstPage.Entities...)
    // If only one page, return early
    if firstPage.PageCount == nil || *firstPage.PageCount <= 1 {
        return &allResources, resp, nil
    }
    // 2. Fetch remaining pages concurrently
    totalPages := *firstPage.PageCount
    log.Printf("[PAGINATION] Fetching pages 2-%d concurrently for %s", totalPages, ResourceType)
    type pageResult struct {
        resources []Resource
        err       error
        pageNum   int
    }
    resultsChan := make(chan pageResult, totalPages-1)
    // Get max concurrent pages from config (default 50)
    maxConcurrentPages := 50  // TODO: Make configurable via provider
    semaphore := make(chan struct{}, maxConcurrentPages)
    // Spawn goroutine for each remaining page
    for pageNum := 2; pageNum <= totalPages; pageNum++ {
        go func(page int) {
            semaphore <- struct{}{}        // Acquire semaphore slot
            defer func() { <-semaphore }() // Release semaphore
            // Acquire pooled SDK client for this page
            pooledClient, err := provider.SdkClientPool.Acquire(ctx)
            if err != nil {
                resultsChan <- pageResult{err: err, pageNum: page}
                return
            }
            defer provider.SdkClientPool.Release(pooledClient)
            // Create new proxy with pooled client
            pageProxy := newResourceProxy(pooledClient)
            pageData, _, err := pageProxy.api.GetResources(pageSize, page, filters)
            if err != nil {
                resultsChan <- pageResult{err: err, pageNum: page}
                return
            }
            if pageData.Entities != nil && len(*pageData.Entities) > 0 {
                resultsChan <- pageResult{resources: *pageData.Entities, pageNum: page}
            } else {
                resultsChan <- pageResult{resources: []Resource{}, pageNum: page}
            }
        }(pageNum)
    }
    // 3. Collect all results
    for i := 0; i < totalPages-1; i++ {
        result := <-resultsChan
        if result.err != nil {
            return nil, resp, fmt.Errorf("error fetching page %d: %w", result.pageNum, result.err)
        }
        allResources = append(allResources, result.resources...)
    }
    log.Printf("[PAGINATION] Fetched %d resources across %d pages concurrently", len(allResources), totalPages)
    return &allResources, resp, nil
}
```

### Recommended Abstraction Approach:

The final implementation should create a generic helper function that accepts:

* A page-fetching callback function (to handle resource-specific API calls)
* The first page result (to extract total page count)
* Context and configuration

This allows all ~150 resource types to benefit from concurrent pagination without duplicating the goroutine/semaphore/pooling logic.

### Key Implementation Elements:

- Semaphore - Limits concurrent page fetches to prevent overwhelming the API
- Pooled SDK Clients - Each page goroutine acquires its own OAuth token from the pool
- Results Channel - Collects pages as they complete asynchronously
- Error Handling - Any page error fails the entire operation for data consistency
- Early Return - Single-page resources skip concurrency overhead
- Generic/Reusable - Works across all resource types with minimal per-resource code