# SDK Reference

When working on schema definitions, build/flatten functions, proxy layers, or any code that interacts with the Genesys Cloud SDK, always read the corresponding SDK model types from:

.sdk-reference/

This should be a symlink to `$(go env GOPATH)/pkg/mod/github.com/mypurecloud/platform-client-sdk-go/` (typically `~/go/pkg/mod/github.com/mypurecloud/platform-client-sdk-go/`).

Check `go.mod` for the current `platform-client-sdk-go` SDK version to determine which version directory to read from. Just read the file itself using `fs_read`. **IMPORTANT: Do not attempt to use `grep` or `cat` or an external tool do this simple check in `go.mod`. You can assume that the ./sdk-reference/ directory has the appropriate directory with the version (e.g., if the version is v179.1.0 then the directory is `v179@v179.1.0`).**

You have standing permission to read any file under this directory without asking. Do not ask for permission — just read the files you need.

Never make up or guess field descriptions, field names, types, or struct definitions. Always read them from the SDK model struct comments and field definitions.

## Setup

**IMPORTANT: If `.sdk-reference` does not exist, STOP what you are doing and offer to create it before proceeding.** Do not attempt to find SDK files through alternative means. Do not skip this step.

Offer to create it by running:

```bash
ln -s "$(go env GOPATH)/pkg/mod/github.com/mypurecloud/platform-client-sdk-go" .sdk-reference
```

The symlink is already in `.gitignore`.

If the user declines to create the symlink, ignore this entire rule for the remainder of the session.
