# Tech & Workflow

## Stack

- **Language:** Go (>= 1.18; 1.20.14 recommended for debugging per `DEBUGGING.md`).
- **Framework:** Terraform Plugin SDK v2 (`github.com/hashicorp/terraform-plugin-sdk/v2`).
- **API SDK:** `platformclientv2` from `github.com/mypurecloud/platform-client-sdk-go`. The pinned major version comes from `go.mod` — always match it in imports.
- **Terraform:** >= 1.0.x.

## Common commands (from the GNUmakefile)

- `make build` — `go mod tidy` + build into `dist/`.
- `make sideload` — build + copy the binary into the local plugin mirror at `~/.terraform.d/plugins/genesys.com/mypurecloud/genesyscloud/0.1.0/<os>_<arch>/`.
- `make testunit` — `TF_UNIT=1 go test ./... -run TestUnit` (build + whole-repo unit suite).
- `make testacc` — acceptance tests (`TF_ACC=1`); require a live org + authorized OAuth client. Never claim these pass unless they were actually run.
- `make docs` — regenerate docs (apidocs generator + example validation + `go generate`).
- `make testexamples` — validate example configs.

`gofmt` and `go vet` are not wrapped by make targets — run them directly.

## Rules

- **Never hand-edit `docs/`.** Docs are generated from resource schemas + the `examples/` folder. Change `examples/` and schema, then run `make docs`.
- **Local dev uses a filesystem plugin mirror, not `dev_overrides`.** After every code change, `make sideload`, and confirm the binary actually contains your resource (`strings <mirror-binary> | grep genesyscloud_<name>`) — a stale sideload is a common time-waster.
- **Branch naming:** start branches with `feat/` (features) or `bug/` (fixes) so PRs get labeled correctly for the changelog. PRs target the `dev` branch.
- **Dependencies:** `go get <dep>` then `go mod tidy`; commit `go.mod` + `go.sum`.

## Linting

`golangci.yaml` enables `gosec`, `godot`, `misspell`, `stylecheck`.
