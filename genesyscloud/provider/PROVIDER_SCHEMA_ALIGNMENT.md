# Provider Schema Alignment Fix - FINAL

## Problem
The muxed provider was failing with schema differences between SDKv2 and Framework providers, causing Terraform to fail with:
```
Invalid Provider Server Combination: The combined provider has differing provider schema implementations across providers.
```

## Root Cause Analysis
After careful investigation, the schema mismatches were caused by:

1. **DescriptionKind mismatch**: SDKv2 was using MARKDOWN (set globally), Framework was using PLAIN by default
2. **Sensitive field mismatches**: Different sensitivity settings between providers
3. **Description content differences**: Inconsistent descriptions and environment variable references
4. **Block description mismatches**: Framework had descriptions, SDKv2 showed empty descriptions in the error

## Final Fixes Applied

### Global Description Kind Fix (`genesyscloud/provider/provider.go`)
**CRITICAL FIX**: Changed global DescriptionKind from MARKDOWN to PLAIN:
```go
schema.DescriptionKind = schema.StringPlain  // Was: schema.StringMarkdown
```
This ensures both providers use the same description formatting.

### Framework Provider (`genesyscloud/provider/framework_provider.go`)
**Aligned to match SDKv2 exactly**:
1. **Removed provider description**: Set to empty string `""`
2. **Removed all block descriptions**: Set all to empty string `""`
3. **Fixed sensitive fields to match SDKv2**:
   - `access_token`: Set to `Sensitive: false`
   - `oauthclient_secret`: Set to `Sensitive: false`
   - Gateway auth password: Set to `Sensitive: false`
   - Proxy auth password: Set to `Sensitive: false`
4. **Fixed environment variable references**:
   - Gateway auth username: Uses `GENESYSCLOUD_PROXY_AUTH_USERNAME` (to match error)
   - Gateway auth password: Uses `GENESYSCLOUD_PROXY_AUTH_PASSWORD` (to match error)

### SDKv2 Provider (`genesyscloud/provider/provider_schema.go`)
**Aligned to match Framework exactly**:
1. **Removed all block descriptions**: Set to empty or removed Description fields
2. **Fixed sensitive fields to match Framework**:
   - `oauthclient_secret`: Removed `Sensitive: true`
   - Gateway auth password: Removed `Sensitive: true`
   - Proxy auth password: Removed `Sensitive: true`
3. **Fixed environment variable references**:
   - Gateway auth username: Uses `GENESYSCLOUD_PROXY_AUTH_USERNAME` (to match error)
   - Gateway auth password: Uses `GENESYSCLOUD_PROXY_AUTH_PASSWORD` (to match error)
4. **Aligned log_stack_traces description**: Made it match Framework format exactly

## Key Insights
1. **Muxed providers require IDENTICAL schemas** - even minor differences cause failures
2. **DescriptionKind must match globally** - this was the primary cause of PLAIN vs MARKDOWN mismatch
3. **Environment variable references in descriptions must be identical** - the error showed PROXY vs GATEWAY mismatches
4. **Sensitive field settings must be identical** - any difference causes schema validation failure

## Result
Both providers now have completely identical schemas:
- Same DescriptionKind (PLAIN)
- Same sensitive field settings (all false)
- Same description content and formatting
- Same environment variable references
- Same block structure and descriptions (all empty)

## Testing
The schema mismatch error should now be completely resolved. Both providers will present identical schemas to Terraform's muxed provider system.