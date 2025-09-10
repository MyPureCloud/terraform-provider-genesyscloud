# Phase 2 Migration Implementation Plan
## Plugin Framework Migration: routing_language Resource

### Overview
This document outlines the complete implementation plan for migrating the `routing_language` resource from SDKv2 to Plugin Framework as part of Phase 2 of the provider modernization.

---

## Current State Analysis

### ✅ What's Working (Phase 1 Complete)
- Basic muxing infrastructure exists
- Framework provider skeleton exists  
- SDKv2 provider works perfectly
- Registration system works for SDKv2

### 🚨 Critical Gaps Identified
- Muxer doesn't actually mux (only returns SDKv2)
- Framework provider has no configuration
- Registration system SDKv2-only
- No Framework resource support

### Target Resource: `routing_language`
- **Location**: `genesyscloud/routing_language/`
- **Complexity**: ⭐⭐☆☆☆ (Low-Medium)
- **Schema**: Simple single field (`name`)
- **Operations**: Create, Read, Delete (no Update)
- **Dependencies**: Minimal external dependencies
- **API**: Stable Genesys Cloud RoutingApi

---

## Implementation Stages

### Stage 1: Foundation (Risk: LOW 🟢)
**Goal**: Prepare infrastructure without breaking existing functionality

#### Stage 1.1: Registration System Enhancement (Week 1)
**Risk Level**: 🟢 **LOW** - Pure additive changes

**Files to Modify**:
- `genesyscloud/resource_register/resource_register.go`
- `genesyscloud/provider_registrar/provider_registrar.go`

**Tasks**:
1. **Extend Registrar Interface** (Additive only)
   ```go
   type Registrar interface {
       // Existing SDKv2 methods
       RegisterResource(resourceType string, resource *schema.Resource)
       RegisterDataSource(dataSourceType string, datasource *schema.Resource)
       RegisterExporter(exporterResourceType string, resourceExporter *resourceExporter.ResourceExporter)
       
       // NEW: Framework methods
       RegisterFrameworkResource(resourceType string, resourceFactory func() resource.Resource)
       RegisterFrameworkDataSource(dataSourceType string, dataSourceFactory func() datasource.DataSource)
       
       // NEW: Provider type tracking
       GetResourceProviderType(resourceType string) ProviderType
       GetDataSourceProviderType(dataSourceType string) ProviderType
   }
   ```

2. **Add Framework storage maps** (New variables)
   ```go
   var (
       // Existing SDKv2 maps
       providerResources     = make(map[string]*schema.Resource)
       providerDataSources   = make(map[string]*schema.Resource)
       resourceExporters     = make(map[string]*resourceExporter.ResourceExporter)
       
       // NEW: Framework maps
       frameworkResources    = make(map[string]func() resource.Resource)
       frameworkDataSources  = make(map[string]func() datasource.DataSource)
       
       // NEW: Provider type tracking
       resourceProviderTypes    = make(map[string]ProviderType)
       dataSourceProviderTypes  = make(map[string]ProviderType)
   )
   ```

3. **Add Framework registration methods** (New methods)
4. **Add provider type tracking** (New functionality)

**Safety Measures**:
- ✅ Zero breaking changes - only additions
- ✅ Backward compatibility - existing code unchanged
- ✅ Isolated changes - new code paths only
- ✅ Easy rollback - can remove additions

**Validation Commands**:
```bash
go test ./genesyscloud/provider_registrar/...
go test ./genesyscloud/resource_register/...
```

**Success Criteria**:
- All existing tests pass
- New Framework registration methods available
- Provider type tracking works
- No breaking changes

#### Stage 1.2: Framework Provider Configuration (Week 1-2)
**Risk Level**: 🟡 **MEDIUM** - Modifies existing Framework provider

**Files to Modify**:
- `genesyscloud/provider/framework_provider.go`
- Create: `genesyscloud/provider/framework_provider_meta.go`
- Create: `genesyscloud/provider/framework_utils.go`

**Tasks**:
1. **Add provider schema** to Framework provider
   ```go
   func (p *GenesysCloudFrameworkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
       resp.Schema = schema.Schema{
           Attributes: map[string]schema.Attribute{
               "access_token": schema.StringAttribute{
                   Optional:  true,
                   Sensitive: true,
               },
               "oauthclient_id": schema.StringAttribute{
                   Optional: true,
               },
               // ... all other provider attributes
           },
       }
   }
   ```

2. **Add configuration logic** for API access
3. **Add provider meta sharing** between SDKv2 and Framework
4. **Add Framework utilities** (client pool, etc.)

**Safety Measures**:
- ✅ Framework provider currently unused - safe to modify
- ✅ No impact on SDKv2 - separate code paths
- ✅ Gradual implementation - one component at a time

**Validation Commands**:
```bash
go test ./genesyscloud/provider/framework_provider_test.go
```

**Success Criteria**:
- Framework provider can be instantiated
- Framework provider can be configured
- API access works from Framework provider
- Provider meta sharing works

---

### Stage 2: Integration (Risk: MEDIUM 🟡)
**Goal**: Connect components without affecting existing resources

#### Stage 2.1: Muxer Enhancement (Week 2)
**Risk Level**: 🟡 **MEDIUM** - Modifies critical muxing logic

**Files to Modify**:
- `genesyscloud/provider/mux.go`

**Tasks**:
1. **Implement actual muxing** (currently returns SDKv2 only)
   ```go
   func NewMuxedProvider(...) func() (func() tfprotov6.ProviderServer, error) {
       return func() (func() tfprotov6.ProviderServer, error) {
           ctx := context.Background()
           
           // SDKv2 provider (upgraded to v6)
           sdkv2Provider := NewSDKv2Provider(version, providerResources, providerDataSources)()
           upgradedV6, err := tf5to6server.UpgradeServer(ctx, sdkv2Provider.GRPCProvider)
           if err != nil {
               return nil, err
           }
           
           // Framework provider (native v6)
           frameworkProvider := NewFrameworkProvider(version)()
           
           // Mux both providers
           muxServer, err := tf6muxserver.NewMuxServer(ctx, upgradedV6, frameworkProvider)
           if err != nil {
               return nil, err
           }
           
           return muxServer.ProviderServer, nil
       }
   }
   ```

2. **Add resource routing logic** based on provider type
3. **Add Framework provider to mux** (with empty resource list)
4. **Add routing validation**

**Safety Measures**:
- ✅ Framework provider has no resources - no routing conflicts
- ✅ All existing resources route to SDKv2 - no behavior change
- ✅ Gradual activation - can enable/disable Framework provider

**Validation Commands**:
```bash
terraform plan  # Should work unchanged
go test ./genesyscloud/provider/...
```

**Success Criteria**:
- Muxer combines both providers
- All existing resources still work through SDKv2
- Framework provider is integrated but unused
- No performance degradation

#### Stage 2.2: Testing Infrastructure (Week 2-3)
**Risk Level**: 🟢 **LOW** - Test-only changes

**Files to Create**:
- `genesyscloud/provider/framework_provider_test.go`
- `genesyscloud/provider/mux_test.go`
- `genesyscloud/provider/framework_utils_test.go`

**Tasks**:
1. **Add Framework testing utilities**
2. **Add muxer testing scenarios**
3. **Add provider routing tests**
4. **Add migration testing framework**

**Safety Measures**:
- ✅ Test-only changes - no production impact
- ✅ Additive testing - existing tests unchanged

**Success Criteria**:
- Framework provider testing utilities work
- Muxer routing can be tested
- Migration testing framework ready

---

### Stage 3: Validation (Risk: LOW 🟢)
**Goal**: Comprehensive validation before resource migration

#### Stage 3.1: End-to-End Testing (Week 3)
**Risk Level**: 🟢 **LOW** - Validation only

**Tasks**:
1. **Full regression testing** - all existing functionality
2. **Performance testing** - ensure no degradation
3. **Muxer routing validation** - correct provider selection
4. **Framework provider validation** - can handle requests

**Validation Commands**:
```bash
# Full test suite
go test ./...

# Performance benchmarks
go test -bench=. ./genesyscloud/provider/...

# Integration tests
terraform plan
terraform apply
terraform destroy
```

**Success Criteria**:
- ✅ All existing tests pass
- ✅ No performance degradation (< 5% overhead)
- ✅ Framework provider can be configured
- ✅ Muxer routes correctly (even with empty Framework resources)

---

### Stage 4: First Resource Migration (Risk: MEDIUM 🟡)
**Goal**: Migrate `routing_language` as proof of concept

#### Stage 4.1: Framework Resource Implementation (Week 4)
**Risk Level**: 🟡 **MEDIUM** - First actual migration

**Files to Create**:
- `genesyscloud/routing_language/framework_resource_genesyscloud_routing_language.go`
- `genesyscloud/routing_language/framework_data_source_genesyscloud_routing_language.go`
- `genesyscloud/routing_language/framework_routing_language_proxy.go`

**Tasks**:
1. **Create Framework routing_language resource**
   ```go
   type routingLanguageResource struct {
       client *platformclientv2.Configuration
   }
   
   func (r *routingLanguageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
       resp.TypeName = req.ProviderTypeName + "_routing_language"
   }
   
   func (r *routingLanguageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
       resp.Schema = schema.Schema{
           Attributes: map[string]schema.Attribute{
               "id": schema.StringAttribute{
                   Computed: true,
                   PlanModifiers: []planmodifier.String{
                       stringplanmodifier.UseStateForUnknown(),
                   },
               },
               "name": schema.StringAttribute{
                   Required: true,
                   PlanModifiers: []planmodifier.String{
                       stringplanmodifier.RequiresReplace(),
                   },
               },
           },
       }
   }
   ```

2. **Implement Framework CRUD operations**
3. **Add Framework data source**
4. **Register with Framework provider**

**Safety Measures**:
- ✅ Feature flag controlled - can disable Framework resource
- ✅ Parallel implementation - SDKv2 version remains
- ✅ Gradual rollout - test environments first

**Success Criteria**:
- Framework resource implements all required interfaces
- CRUD operations work identically to SDKv2
- Data source works identically to SDKv2
- Resource registers with Framework provider

#### Stage 4.2: Migration Testing (Week 4-5)
**Risk Level**: 🟡 **MEDIUM** - Validation of migrated resource

**Files to Create**:
- `genesyscloud/routing_language/framework_resource_genesyscloud_routing_language_test.go`
- `genesyscloud/routing_language/framework_data_source_genesyscloud_routing_language_test.go`

**Tasks**:
1. **Comprehensive resource testing**
2. **Compatibility testing** - same behavior as SDKv2
3. **Performance comparison**
4. **Integration testing**

**Validation Commands**:
```bash
# Framework resource tests
go test ./genesyscloud/routing_language/framework_*

# Compatibility tests
go test ./genesyscloud/routing_language/...

# Integration tests with both providers
terraform plan -var="use_framework_provider=true"
```

**Success Criteria**:
- All Framework resource tests pass
- Behavior identical to SDKv2 version
- Performance comparable to SDKv2
- Integration tests pass

---

## Risk Mitigation Strategy

### Risk Assessment by Stage

| Stage | Risk Level | Mitigation Strategy | Rollback Plan |
|-------|------------|-------------------|---------------|
| 1.1 | 🟢 LOW | Additive only, no breaking changes | Remove new code |
| 1.2 | 🟡 MEDIUM | Framework provider unused | Revert Framework provider |
| 2.1 | 🟡 MEDIUM | Empty Framework resources | Disable Framework in mux |
| 2.2 | 🟢 LOW | Test-only changes | Remove test code |
| 3.1 | 🟢 LOW | Validation only | N/A |
| 4.1 | 🟡 MEDIUM | Feature flag controlled | Disable Framework resource |
| 4.2 | 🟡 MEDIUM | Parallel SDKv2 implementation | Route back to SDKv2 |

### Safety Checkpoints

#### Checkpoint 1: After Stage 1
```bash
✅ All existing tests pass
✅ No breaking changes introduced
✅ Registration system enhanced
✅ Framework provider configurable
```

#### Checkpoint 2: After Stage 2
```bash
✅ Muxer works with both providers
✅ All resources still route to SDKv2
✅ Framework provider integrated but empty
✅ Testing infrastructure ready
```

#### Checkpoint 3: After Stage 3
```bash
✅ Full regression testing passed
✅ Performance impact acceptable
✅ Infrastructure ready for migration
✅ Rollback procedures tested
```

#### Checkpoint 4: After Stage 4
```bash
✅ First resource successfully migrated
✅ Framework resource works identically to SDKv2
✅ Migration process validated
✅ Ready for additional resource migrations
```

---

## Implementation Timeline

### Week 1: Foundation
- **Days 1-3**: Registration system enhancement (Stage 1.1)
- **Days 4-5**: Framework provider configuration (Stage 1.2)

### Week 2: Integration  
- **Days 1-3**: Muxer enhancement (Stage 2.1)
- **Days 4-5**: Testing infrastructure (Stage 2.2)

### Week 3: Validation
- **Days 1-2**: End-to-end testing (Stage 3.1)
- **Days 3-5**: Performance and regression testing

### Week 4-5: First Migration
- **Week 4**: Framework resource implementation (Stage 4.1)
- **Week 5**: Migration testing and validation (Stage 4.2)

---

## Success Metrics

### Overall Success Criteria
- Zero breaking changes to existing functionality
- Framework resource behaves identically to SDKv2
- Performance impact < 5%
- All tests pass
- Clear migration template for future resources

### Stage-Specific Success Metrics
Each stage has specific success criteria listed in the stage details above.

---

## Next Steps

1. **Review this plan** with the team
2. **Start with Stage 1.1** - Registration system enhancement
3. **Validate each checkpoint** before proceeding
4. **Document lessons learned** for future resource migrations

---

## Notes and Considerations

- This plan prioritizes safety and incremental progress
- Each stage can be rolled back independently
- The `routing_language` resource is chosen for its simplicity
- Success here creates a template for migrating other resources
- Performance monitoring is critical throughout
