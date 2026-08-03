# Concurrent Pagination

This guide explains how to migrate resource proxy `getAll*` functions from sequential page loops to the shared concurrent pagination helper. The goal is faster CX as Code exports and data source lookups when an org has many pages of a given resource type.

## Background

Large exports spend significant time in proxy `getAll*` functions that list every entity in an org. Historically these functions fetched page 1, then looped `for pageNum := 2; pageNum <= PageCount; pageNum++` on a single SDK client.

The provider now supports parallel page fetches via:

- `genesyscloud/provider/concurrent_pagination.go` — `FetchPagesConcurrently`
- `genesyscloud/provider/sdk_client_pool.go` — OAuth token pool; each concurrent page acquires its own SDK client
- Provider settings `max_concurrent_pages` and `token_pool_size` (see [docs/index.md](./docs/index.md))

When `max_concurrent_pages` is `1` (the default), behavior is identical to the old sequential loop. Higher values fan out pages 2–N in parallel, bounded by a semaphore.

## When to Apply This Pattern

Apply the pattern to any proxy function that:

1. Lists all entities for export or cache hydration
2. Uses offset/page-number pagination (`pageSize`, `pageNumber` or SDK equivalents)
3. Reads `PageCount` from the first response and loops from page 2

Typical function names: `getAll*Fn`, `getAll*Attr` implementations, or nested helpers called by exporters and data sources.

**Do not apply** when:

- The API uses cursor/token pagination with no stable page count (requires a different approach)
- Listing is already filtered server-side to a single page (e.g. name search returning one result)
- If ~5 pages or less are always to be expected, the difference in performace is minimal and concurrent pagination is not worth the complexity. The real benefits become clear when 10+ pages are expected.

## Reference Implementations

These resources are already migrated and are good templates:

| Resource | File |
|----------|------|
| `genesyscloud_group` | `genesyscloud/group/genesyscloud_group_proxy.go` |
| `genesyscloud_routing_skill` | `genesyscloud/routing_skill/genesyscloud_routing_skill_proxy.go` |
| `genesyscloud_routing_queue` | `genesyscloud/routing_queue/genesyscloud_routing_queue_proxy.go` |
| `genesyscloud_script` | `genesyscloud/scripts/genesyscloud_scripts_proxy.go` |
| `genesyscloud_user` | `genesyscloud/user/genesyscloud_user_proxy.go` |
| `genesyscloud_architect_flow` | `genesyscloud/architect_flow/resource_genesyscloud_architect_flow_proxy.go` |

## Migration Steps

### 1. Find candidates

Search for sequential pagination loops in proxy files:

```sh
rg 'for pageNum := 2' genesyscloud --glob '*_proxy.go'
```

Each match is a function that should be evaluated for migration.

### 2. Fetch page 1 on the primary proxy (unchanged)

Keep the existing first-page call on `p` (the proxy passed into the `getAll*Fn`). Handle errors and empty results the same way.

```go
const pageSize = 100

entities, resp, err := p.someApi.GetThings(pageSize, 1, /* other args */)
if err != nil {
    return nil, resp, err
}
if entities.Entities == nil || len(*entities.Entities) == 0 {
    return &allThings, resp, nil
}

allThings = append(allThings, *entities.Entities...)
```

### 3. Resolve `totalPages` safely

Do not dereference `PageCount` without a nil check. Default to `1` when missing.

```go
totalPages := 1
if entities.PageCount != nil {
    totalPages = *entities.PageCount
}
```

Apply any API-specific caps **before** calling `FetchPagesConcurrently` (see [Special cases](#special-cases) below).

### 4. Replace the sequential loop with `FetchPagesConcurrently`

Remove:

```go
for pageNum := 2; pageNum <= *entities.PageCount; pageNum++ {
    // ...
}
```

Replace with:

```go
allThings, resp, err = provider.FetchPagesConcurrently(ctx, ResourceType, allThings, resp, totalPages, p.clientConfig,
    func(ctx context.Context, clientConfig *platformclientv2.Configuration, pageNum int) ([]platformclientv2.Thing, *platformclientv2.APIResponse, error) {
        ctx = provider.EnsureResourceContext(ctx, ResourceType)
        pageProxy := newThingProxy(clientConfig)
        pageList, pageResp, pageErr := pageProxy.someApi.GetThings(pageSize, pageNum, /* same args as page 1 */)
        if pageErr != nil {
            return nil, pageResp, fmt.Errorf("failed to get page of things: %w", pageErr)
        }
        if pageList.Entities == nil || len(*pageList.Entities) == 0 {
            return []platformclientv2.Thing{}, pageResp, nil
        }
        return *pageList.Entities, pageResp, nil
    },
)
if err != nil {
    return nil, resp, err
}
```

### 5. Keep cache population after pagination

If the function populates a resource cache, leave that loop **after** `FetchPagesConcurrently` returns the full slice. Do not write to the cache inside the per-page callback.

```go
for _, thing := range allThings {
    rc.SetCache(p.thingCache, *thing.Id, thing)
}
```

### 6. Preserve context and logging conventions

At the start of `getAll*Fn`:

```go
ctx = provider.EnsureResourceContext(ctx, ResourceType)
```

Repeat `EnsureResourceContext` inside the page callback so SDK debug logs attribute requests to the correct resource type.

## Requirements for the Page Callback

| Rule | Why |
|------|-----|
| Create a **new proxy** per page (`newXxxProxy(clientConfig)`) | Each goroutine must use the pooled `clientConfig`, not `p.clientConfig` |
| Pass the **same API arguments** as page 1 (filters, expands, sort) | Page 2+ must return the same logical result set |
| Return an **empty slice**, not `nil`, when a page has no entities | Avoids nil-append edge cases |
| Wrap errors with context (`failed to get page of ...`) | Easier export failure diagnosis |
| Do **not** mutate shared proxy state in the callback | Callbacks run concurrently |

`FetchPagesConcurrently` fetches page 1 sequentially on the caller's client, then pages 2–N concurrently. It merges results in page order regardless of completion order.

## Token pool size and export time

Raising `token_pool_size` does **not** always shorten total export time. Each token is minted at **provider startup** via `POST /oauth/token` (client credentials). Terraform initializes the provider on both plan and apply, so the pool is prefilled twice per `terraform apply`.

A higher pool size means more OAuth grants up front, before any resource API calls run. The Genesys OAuth endpoint has its own rate limit, separate from per-token API limits. Bursting many parallel token requests triggers `400 rate limit exceeded` responses; failed mints retry for up to one minute each. That prefill delay can dominate end-to-end export time even when pagination itself is faster.

Benchmarking a 456-page user export (`pageSize=25`, ~11,378 users, `max_concurrent_pages=75`):

| `token_pool_size` | Total apply time | Pagination phase |
|---|---|---|
| 10 | ~23s | ~19s |
| 50 | ~82s | ~18s |

With 50 tokens, pagination was ~2s faster, but provider init added ~60s before the first user API call. With 10 tokens, init took ~1s. Pagination parallelism is capped by whichever is lower: `max_concurrent_pages` or `token_pool_size`.

**Practical guidance:**

- Set `token_pool_size` to match `max_concurrent_pages`, not higher. Extra pool slots you never acquire during pagination only add OAuth prefill cost.
- Tune for **total apply time**, not pagination speed alone.

## Checklist per Resource

- [ ] Identify `getAll*Fn` (or equivalent) with `for pageNum := 2`
- [ ] Page 1 fetch unchanged; nil-safe `totalPages`
- [ ] Sequential loop replaced; page callback uses `newXxxProxy(clientConfig)`
- [ ] Same filters/expands/args on every page
- [ ] Cache population remains after full fetch
- [ ] `EnsureResourceContext` at function entry and in callback
- [ ] `go build ./...` passes
- [ ] Spot-check export or acceptance test for the resource type

## Related Files

| File | Purpose |
|------|---------|
| `genesyscloud/provider/concurrent_pagination.go` | `FetchPagesConcurrently` implementation |
| `genesyscloud/provider/concurrent_pagination_test.go` | Helper unit tests |
| `genesyscloud/provider/sdk_client_pool.go` | Token pool acquire/release for parallel pages |
| `genesyscloud/provider/provider_schema.go` | `max_concurrent_pages`, `token_pool_size` schema |
