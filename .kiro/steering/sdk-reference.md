---
inclusion: fileMatch
fileMatchPattern: ['**/*_proxy.go', '**/*_schema.go', '**/*_utils.go']
---

# Reading the Genesys Cloud SDK

When writing schema, build/flatten, or proxy code, read the SDK model types and API methods directly — never guess field names, types, or method signatures.

## Where the SDK lives

The `platformclientv2` SDK is not vendored; it's in the Go module cache:

```bash
SDK="$(go env GOPATH)/pkg/mod/github.com/mypurecloud/platform-client-sdk-go/vNNN@vNNN.0.0/platformclientv2"
```

Determine `vNNN` from `go.mod` (`grep platform-client-sdk-go go.mod`). Some setups also symlink it to `.sdk-reference/` at the repo root (see `.amazonq/rules/sdk-reference.md`); if that symlink exists, read from there.

## What to read

```bash
cat "$SDK/<model>.go"                                # field names, pointer types, read-only comments
grep -n "func (a XxxApi) .*<Verb>" "$SDK/xxxapi.go"  # exact method names, signatures, query params
```

## Rules

- **Match the pinned SDK major** in every import. A field in a newer SDK may not exist in the pinned one.
- **Never invent** SDK method names, model field names, field types, or descriptions — copy them from the SDK source.
- Note **pointer types** (`*int`, `*bool`, ...) — they map to `resourcedata.SetNillableValue` / `GetNillableValue`.
- Note query params like `createDefault` that change GET behavior.
