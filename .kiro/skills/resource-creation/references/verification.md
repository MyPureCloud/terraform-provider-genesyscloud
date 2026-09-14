# Verification — Resource Creation

Phase 5 is automated and you run it. Phase 6 is live verification against a real org — hand it to the user, or run it if they ask.

## Phase 5 — Automated verification

```bash
gofmt -w genesyscloud/<name>/ && gofmt -l genesyscloud/<name>/   # empty output = clean
go vet ./genesyscloud/<name>/...
make testunit                                                   # go build + all unit tests (TF_UNIT=1 go test ./... -run TestUnit)
make docs                                                       # regenerates docs; must exit 0, no "Missing APIs file"
```

State plainly which checks passed. `make testunit` builds and runs the whole-repo unit suite; `make docs` runs the apidocs generator + `go generate`. Never hand-edit `docs/`.

## Phase 6 — Manual live verification

This repo uses a **filesystem plugin mirror**, NOT `dev_overrides`.

1. **Build + sideload** from the correct branch:
   ```bash
   make sideload   # builds dist/ and copies to ~/.terraform.d/plugins/genesys.com/mypurecloud/genesyscloud/0.1.0/<os>_<arch>/
   ```
2. **Confirm the binary actually contains the resource** (catches a stale sideload — a common wasted-hour mistake):
   ```bash
   strings ~/.terraform.d/plugins/genesys.com/mypurecloud/genesyscloud/0.1.0/$(go env GOOS)_$(go env GOARCH)/terraform-provider-genesyscloud | grep genesyscloud_<name>
   ```
3. **Test config** in a clean dir (source MUST be the mirror path):
   ```hcl
   terraform {
     required_providers {
       genesyscloud = { source = "genesys.com/mypurecloud/genesyscloud", version = "0.1.0" }
     }
   }
   provider "genesyscloud" {}   # creds from GENESYSCLOUD_OAUTHCLIENT_ID / _SECRET / _REGION env vars

   resource "genesyscloud_<name>" "this" { /* writable fields */ }
   data "genesyscloud_<name>" "current" { depends_on = [genesyscloud_<name>.this] }
   ```
4. **Run and check each path** (with the filesystem mirror you DO run `terraform init`):
   - `terraform init && terraform apply` -> **CREATE (PUT) + READ (GET)**. Outputs should show writable values round-tripped and any read-only field populated by the API.
   - Change a couple values, `terraform apply` again -> **UPDATE**. Plan must show `~ update in-place` (not destroy/recreate).
   - `terraform plan` right after apply -> **drift check**. Must say **"No changes."** A perpetual diff means a field isn't round-tripping — fix it.
   - Set a field out of range (e.g. a TTL to 90), `terraform plan` -> **validation** must fail at plan time, before any API call.
   - `terraform import genesyscloud_<name>.this <fixedId>` (singleton) -> should succeed; a follow-up plan shows no diff.
5. **Test export** — add a `genesyscloud_tf_export` block and apply:
   ```hcl
   resource "genesyscloud_tf_export" "export" {
     directory                = "./exported"
     include_state_file       = true
     export_format            = "hcl"
     include_filter_resources = ["genesyscloud_<name>"]
   }
   ```
   Then verify the exported `.tf`:
   - Exactly one resource block for a singleton (not a broken enumerated collection).
   - All **writable** fields present with the org's real values.
   - **Read-only/computed fields are ABSENT** (if a computed field is emitted as settable config, re-apply breaks — that's a real exporter bug to fix).
   - Round-trip it: drop the exported `.tf` in a clean dir and `terraform plan` — no errors about unknown/invalid attributes.
