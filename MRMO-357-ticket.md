Description

Based on findings from https://inindca.atlassian.net/browse/MRMO-296#icft=MRMO-296, the CX as Code exporter makes redundant API calls when fetching related resource data during export. The existing wrapup code caching in the provider has a critical bug: it creates a new empty cache for each proxy instance, so cache hits never occur during concurrent exports.

Testing showed that for 6,709 queues, the provider made ~174,000 wrapup code API calls (avg 26 pages per queue). With proper org-level caching, this was reduced to ~6,759 calls (98% reduction).

This ticket implements an extensible, package-level caching framework that can be applied to multiple resource types (wrapup codes, queue members, and potentially others) to eliminate redundant API calls during export operations.

Key Requirement: The solution should be implemented as a reusable caching pattern that can be easily extended to other resource types beyond wrapup codes and queue members.

User Story

As a reconciliation service operator, I want frequently-accessed resource data to be cached during export so that redundant API calls are eliminated and export performance is improved across multiple resource types.

Acceptance Criteria

Design and implement an extensible org-level caching framework that can be applied to multiple resource types

Implement org-level wrapup code cache in routing_queue proxy using the framework

Implement queue member cache using the same framework pattern

Cache structure uses sync.RWMutex for thread-safe concurrent access

Caches are package-level (shared across all proxy instances, following existing routingQueueCache pattern)

Load org-wide data once at first access with double-checked locking pattern

Per-resource operations fetch lightweight assignment/reference data, resolve full objects from cache

Caches persist for provider process lifetime

Fallback to legacy direct API calls if cache load fails

Export logs show cache hit/miss statistics per resource type

Queue export shows 96%+ reduction in wrapup code API calls (validated: 174,000 → 6,759 for 6,709 queues)

Documentation for how to apply the caching pattern to additional resource types

PoC Implementation Reference

Note: The following pseudo code is from the https://inindca.atlassian.net/browse/MRMO-296#icft=MRMO-296 PoC that validated the org-level caching approach for wrapup codes. The final implementation should abstract this pattern into a reusable caching framework that can be applied to multiple resource types (wrapup codes, queue members, skills, groups, etc.).

Example from PoC (routing_queue wrapup code caching):

```go
// Package-level caches (shared across all proxy instances)
var routingQueueCache = rc.NewResourceCache[platformclientv2.Queue]()
var orgWrapupCodeCache = rc.NewResourceCache[platformclientv2.Wrapupcode]()
var orgWrapupCodeCacheLoaded bool
var orgWrapupCodeCacheMutex sync.RWMutex

// Load org-level wrapup codes once (thread-safe, double-checked locking)
func (p *RoutingQueueProxy) loadOrgWrapupCodes(ctx context.Context) error {
    // Read lock check (fast path)
    orgWrapupCodeCacheMutex.RLock()
    if orgWrapupCodeCacheLoaded {
        orgWrapupCodeCacheMutex.RUnlock()
        log.Printf("[WRAPUP-CACHE] Org wrapup code cache already loaded")
        return nil
    }
    orgWrapupCodeCacheMutex.RUnlock()
    
    // Write lock for loading
    orgWrapupCodeCacheMutex.Lock()
    defer orgWrapupCodeCacheMutex.Unlock()
    
    // Double-check after acquiring write lock
    if orgWrapupCodeCacheLoaded {
        return nil
    }
    
    log.Printf("[WRAPUP-CACHE] Loading ALL org wrapup codes (one-time operation)")
    
    var allWrapupcodes []platformclientv2.Wrapupcode
    const pageSize = 100
    pageNum := 1
    
    // Fetch ALL org-level wrapup codes (not queue-specific)
    for {
        wrapupcodes, _, err := p.routingApi.GetRoutingWrapupcodes(pageSize, pageNum, "", "", "", []string{}, []string{})
        if err != nil {
            return fmt.Errorf("failed to load org wrapup codes page %d: %w", pageNum, err)
        }
        
        if wrapupcodes.Entities == nil || len(*wrapupcodes.Entities) == 0 {
            break
        }
        
        allWrapupcodes = append(allWrapupcodes, *wrapupcodes.Entities...)
        
        if wrapupcodes.PageCount == nil || pageNum >= *wrapupcodes.PageCount {
            break
        }
        pageNum++
    }
    
    // Cache ALL wrapup codes by ID
    for _, wc := range allWrapupcodes {
        if wc.Id != nil {
            rc.SetCache(orgWrapupCodeCache, *wc.Id, wc)
        }
    }
    
    orgWrapupCodeCacheLoaded = true
    log.Printf("[WRAPUP-CACHE] Loaded %d org wrapup codes in %d API calls", len(allWrapupcodes), pageNum)
    return nil
}

// Get queue wrapup codes using org-level cache
func getAllRoutingQueueWrapupCodesFn(ctx context.Context, p *RoutingQueueProxy, queueId string) (*[]platformclientv2.Wrapupcode, *platformclientv2.APIResponse, error) {
    // Step 1: Load org-level cache (happens once, thread-safe)
    if err := p.loadOrgWrapupCodes(ctx); err != nil {
        log.Printf("[WRAPUP-CACHE] Failed to load org cache, falling back to per-queue fetch: %v", err)
        return getAllRoutingQueueWrapupCodesLegacy(ctx, p, queueId)
    }
    
    // Step 2: Fetch ONLY the queue-wrapupcode assignments (1 lightweight API call)
    const pageSize = 100
    queueWrapupCodes, apiResponse, err := p.routingApi.GetRoutingQueueWrapupcodes(queueId, pageSize, 1, "")
    if err != nil {
        return nil, apiResponse, fmt.Errorf("failed to get queue wrapup code assignments: %w", err)
    }
    
    if queueWrapupCodes.Entities == nil || len(*queueWrapupCodes.Entities) == 0 {
        return &[]platformclientv2.Wrapupcode{}, apiResponse, nil
    }
    
    // Step 3: Look up full wrapup code objects from org-level cache
    var result []platformclientv2.Wrapupcode
    cacheHits := 0
    for _, entity := range *queueWrapupCodes.Entities {
        if entity.Id != nil {
            wc := rc.GetCacheItem(orgWrapupCodeCache, *entity.Id)
            if wc != nil {
                result = append(result, *wc)
                cacheHits++
            } else {
                log.Printf("[WRAPUP-CACHE] WARNING: Wrapup code %s not found in org cache", *entity.Id)
                result = append(result, entity)
            }
        }
    }
    
    log.Printf("[WRAPUP-CACHE] Queue %s: Retrieved %d wrapup codes (%d cache hits, 1 API call vs %d pages)", 
        queueId, len(result), cacheHits, *queueWrapupCodes.PageCount)
    
    return &result, apiResponse, nil
}
```

Recommended Abstraction Approach:

The final implementation should create a generic caching framework that:

Provides a reusable cache initialization pattern with double-checked locking

Supports multiple cache types (wrapup codes, queue members, skills, groups, etc.)

Uses generic types where possible to reduce code duplication

Includes helper functions for common operations (load-once, cache lookup, fallback)

Provides clear extension points for adding new cached resource types

Example extensible design:

// Generic org-level cache manager
type OrgCache[T any] struct {
    cache      *rc.ResourceCache[T]
    loaded     bool
    mutex      sync.RWMutex
    loadFunc   func(context.Context) ([]T, error)
    cacheKey   func(T) string
}

// Reusable load-once pattern
func (oc *OrgCache[T]) EnsureLoaded(ctx context.Context) error {
    // Double-checked locking pattern
    // ... (reusable across all cache types)
}

This allows queue members, wrapup codes, and future resource types to benefit from the same caching infrastructure without duplicating the thread-safety and load-once logic.

Key Implementation Elements:

Package-Level Cache - Shared across all proxy instances (not singleton, follows existing routingQueueCache pattern)

Double-Checked Locking - Thread-safe cache loading with RWMutex for performance

One-Time Load - ALL org resources loaded once at first access

Lightweight Per-Resource Calls - Only fetch assignments/references, resolve from cache

Fallback Safety - Falls back to legacy direct API calls if cache load fails

Instrumentation - Logs cache hits/misses and API call reduction

Extensible Design - Easy to add new cached resource types following the same pattern

Validated Performance Impact (Wrapup Codes):

Before: ~174,000 wrapup API calls for 6,709 queues (avg 26 pages/queue)

After: ~6,759 API calls (50 org-level + 6,709 assignment lookups)

Reduction: 98% fewer API calls

