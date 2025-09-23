# Terraform Provider Migration: Strategy & Planning Guide

## Table of Contents
1. [Migration Overview](#migration-overview)
2. [What is Terraform Plugin Framework?](#what-is-terraform-plugin-framework)
3. [Why Migrate from SDKv2 to Plugin Framework?](#why-migrate-from-sdkv2-to-plugin-framework)
4. [What is Muxing?](#what-is-muxing)
5. [Incremental Implementation Strategy](#incremental-implementation-strategy)
6. [Phase 2 Implementation Plan](#phase-2-implementation-plan)
7. [Migration Checklist](#migration-checklist)
8. [Risk Mitigation Strategy](#risk-mitigation-strategy)
9. [Timeline and Roadmap](#timeline-and-roadmap)

---

## Migration Overview

This document outlines the complete strategy for migrating the Genesys Cloud Terraform provider from SDKv2 to the Plugin Framework using a muxing approach. The migration is designed to be incremental, safe, and maintain zero breaking changes.

### Key Principles:
- **Zero-Downtime Migration**: No breaking changes to existing configurations
- **Incremental Approach**: Resource-by-resource migration
- **Risk Mitigation**: Each stage can be rolled back independently
- **Backward Compatibility**: Existing infrastructure continues to work unchanged

---

## What is Terraform Plugin Framework?

The **Terraform Plugin Framework** is the modern, next-generation approach to building Terraform providers, introduced as an evolution from the older Plugin SDK v2.

### Key Characteristics:
- **Protocol v6 Native**: Built from the ground up for Terraform's Protocol v6
- **Interface-Based Design**: Uses Go interfaces instead of function-based callbacks
- **Type Safety**: Stronger type safety with `types` package
- **Better Error Handling**: More structured error handling with `diag.Diagnostics`
- **Modern Go Patterns**: Leverages modern Go features and best practices
- **Future-Proof**: Designed to support future Terraform versions and features

### Architecture Comparison:

| Aspect | SDK v2 | Plugin Framework |
|--------|--------|------------------|
| Protocol | v5 (upgraded to v6) | v6 (native) |
| Schema Definition | `*schema.Schema` | `schema.Schema` |
| Resource Methods | `CreateContext`, `ReadContext` | `CreateRequest`, `ReadRequest` |
| Error Handling | `diag.Diagnostics` | `diag.Diagnostics` (enhanced) |
| Type System | `interface{}` | Strongly typed with `types` package |

---

## Why Migrate from SDKv2 to Plugin Framework?

### 1. **Future-Proofing**
- **SDK v2 is in maintenance mode** - no new features or major updates
- **Plugin Framework is actively developed** with new features and improvements
- **Long-term support** for future Terraform versions

### 2. **Performance Improvements**
- **Native Protocol v6 support** eliminates conversion overhead
- **Better memory management** and resource utilization
- **Improved concurrency** handling

### 3. **Developer Experience**
- **Type safety** reduces runtime errors
- **Better IDE support** with modern Go patterns
- **Cleaner, more maintainable code** structure
- **Enhanced debugging** capabilities

### 4. **Terraform Ecosystem Alignment**
- **Consistent with HashiCorp's direction** for provider development
- **Better integration** with Terraform Cloud/Enterprise features
- **Improved compatibility** with future Terraform versions

### 5. **Code Quality and Maintainability**
- **Reduced boilerplate code** for common operations
- **Better separation of concerns** between schema and logic
- **More testable** code structure
- **Easier to extend** and modify

---

## What is Muxing?

**Muxing** (short for "multiplexing") in the context of Terraform providers is a technique that allows multiple provider implementations to coexist within a single provider binary. It acts as a "traffic director" that routes requests to the appropriate provider implementation based on the resource being accessed.

### How Muxing Works:

```
Terraform Core
     ↓
Protocol v6 Server
     ↓
Muxer (Traffic Director)
     ↓
┌─────────────────┬─────────────────┐
│   SDK v2        │   Framework     │
│   Provider      │   Provider      │
│   (Legacy)      │   (Modern)      │
└─────────────────┴─────────────────┘
```

### Key Components:
1. **Protocol v6 Server**: Handles communication with Terraform Core
2. **Muxer**: Routes requests to appropriate provider
3. **SDK v2 Provider**: Handles existing resources (upgraded to v6)
4. **Framework Provider**: Handles migrated resources (native v6)

### Why Implement Muxing Logic?

#### 1. **Zero-Downtime Migration**
- **No breaking changes** to existing Terraform configurations
- **Gradual migration** of resources from SDK v2 to Framework
- **Backward compatibility** maintained throughout the process

#### 2. **Risk Mitigation**
- **Incremental testing** of migrated resources
- **Rollback capability** if issues arise
- **Isolated changes** per resource

#### 3. **Operational Continuity**
- **Existing infrastructure** continues to work unchanged
- **No immediate action required** from users
- **Smooth transition** for development teams

#### 4. **Development Efficiency**
- **Parallel development** of new features
- **Resource-by-resource migration** approach
- **Reduced testing complexity** per migration

---

## Incremental Implementation Strategy

### Phase-Based Approach:

#### **Phase 1: Infrastructure Setup** ✅ (Complete)
- Implement muxing infrastructure
- Upgrade SDK v2 provider to Protocol v6
- Create empty Framework provider
- Ensure all existing resources work unchanged

#### **Phase 2: Resource Migration** (Current)
- Migrate high-priority resources to Framework
- Implement comprehensive testing
- Validate performance improvements
- Update documentation

#### **Phase 3: Optimization** (Future)
- Migrate remaining resources
- Remove SDK v2 dependencies
- Optimize Framework provider
- Performance tuning

#### **Phase 4: Cleanup** (Future)
- Remove SDK v2 code
- Update build processes
- Final documentation updates
- Release preparation

### Migration Criteria:
1. **Resource Complexity**: Start with simple resources
2. **Usage Frequency**: Prioritize commonly used resources
3. **Dependencies**: Consider resource interdependencies
4. **Testing Coverage**: Ensure adequate test coverage exists

---

## Phase 2 Implementation Plan

### Overview
Complete implementation plan for migrating the `routing_language` resource from SDKv2 to Plugin Framework as part of Phase 2 of the provider modernization.

### Current State Analysis

#### ✅ What's Working (Phase 1 Complete)
- Basic muxing infrastructure exists
- Framework provider skeleton exists  
- SDKv2 provider works perfectly
- Registration system works for SDKv2

#### 🚨 Critical Gaps Identified
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

### Implementation Stages

#### Stage 1: Foundation (Risk: LOW 🟢)
**Goal**: Prepare infrastructure without breaking existing functionality

##### Stage 1.1: Registration System Enhancement
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
3. **Add Framework registration methods** (New methods)
4. **Add provider type tracking** (New functionality)

##### Stage 1.2: Framework Provider Configuration
**Risk Level**: 🟡 **MEDIUM** - Modifies existing Framework provider

**Files to Modify**:
- `genesyscloud/provider/framework_provider.go`
- Create: `genesyscloud/provider/framework_provider_meta.go`
- Create: `genesyscloud/provider/framework_utils.go`

**Tasks**:
1. **Add provider schema** to Framework provider
2. **Add configuration logic** for API access
3. **Add provider meta sharing** between SDKv2 and Framework
4. **Add Framework utilities** (client pool, etc.)

#### Stage 2: Integration (Risk: MEDIUM 🟡)
**Goal**: Connect components without affecting existing resources

##### Stage 2.1: Muxer Enhancement
**Risk Level**: 🟡 **MEDIUM** - Modifies critical muxing logic

**Files to Modify**:
- `genesyscloud/provider/mux.go`

**Tasks**:
1. **Implement actual muxing** (currently returns SDKv2 only)
2. **Add resource routing logic** based on provider type
3. **Add Framework provider to mux** (with empty resource list)
4. **Add routing validation**

##### Stage 2.2: Testing Infrastructure
**Risk Level**: 🟢 **LOW** - Test-only changes

**Files to Create**:
- `genesyscloud/provider/framework_provider_test.go`
- `genesyscloud/provider/mux_test.go`
- `genesyscloud/provider/framework_utils_test.go`

#### Stage 3: Validation (Risk: LOW 🟢)
**Goal**: Comprehensive validation before resource migration

##### Stage 3.1: End-to-End Testing
**Tasks**:
1. **Full regression testing** - all existing functionality
2. **Performance testing** - ensure no degradation
3. **Muxer routing validation** - correct provider selection
4. **Framework provider validation** - can handle requests

#### Stage 4: First Resource Migration (Risk: MEDIUM 🟡)
**Goal**: Migrate `routing_language` as proof of concept

##### Stage 4.1: Framework Resource Implementation
**Files to Create**:
- `genesyscloud/routing_language/framework_resource_genesyscloud_routing_language.go`
- `genesyscloud/routing_language/framework_data_source_genesyscloud_routing_language.go`
- `genesyscloud/routing_language/framework_routing_language_proxy.go`

##### Stage 4.2: Migration Testing
**Files to Create**:
- `genesyscloud/routing_language/framework_resource_genesyscloud_routing_language_test.go`
- `genesyscloud/routing_language/framework_data_source_genesyscloud_routing_language_test.go`

---

## Migration Checklist

### 📋 Stage 1: Foundation

#### ✅ Stage 1.1: Registration System Enhancement
- [x] Extend `Registrar` interface in `resource_register.go`
- [x] Add Framework storage maps in `provider_registrar.go`
- [x] Add `RegisterFrameworkResource()` method
- [x] Add `RegisterFrameworkDataSource()` method
- [x] Add provider type tracking (`ProviderType` enum)
- [x] Add `GetResourceProviderType()` method
- [x] Add `GetFrameworkResources()` getter
- [x] Add `GetFrameworkDataSources()` getter
- [ ] **Validation**: All existing tests pass
- [ ] **Validation**: No breaking changes

#### ✅ Stage 1.2: Framework Provider Configuration
- [x] Add full provider schema to `framework_provider.go`
- [x] Implement `Configure()` method with API access
- [x] Create `framework_provider_meta.go`
- [x] Create `framework_utils.go`
- [x] Add client pool integration
- [x] Add provider meta sharing
- [ ] **Validation**: Framework provider configurable
- [ ] **Validation**: API access works

### 📋 Stage 2: Integration

#### ✅ Stage 2.1: Muxer Enhancement
- [x] Import `tf6muxserver` in `mux.go`
- [x] Implement actual muxing with `NewMuxServer()`
- [x] Add Framework provider to mux
- [x] Add resource routing logic
- [x] Add routing validation
- [x] Create `mux_test.go` for testing
- [x] Create `stage2_1_validation.go` for comprehensive validation
- [ ] **Validation**: Muxer combines both providers
- [ ] **Validation**: All existing resources work unchanged

#### ✅ Stage 2.2: Testing Infrastructure
- [x] Create `framework_provider_test.go`
- [x] Create `mux_test.go`
- [x] Fix test type conflicts and naming issues
- [x] Add Framework testing utilities
- [x] Add muxer testing scenarios
- [x] Add comprehensive test coverage
- [x] **Validation**: Testing infrastructure ready

### 📋 Stage 3: Validation

#### ✅ Stage 3.1: End-to-End Testing
- [x] Create comprehensive validation plan (`stage3_validation_plan.md`)
- [x] Organize testing commands by category
- [x] Define success criteria and red flags
- [x] Create validation report template
- [ ] **Execute**: Run compilation & basic tests
- [ ] **Execute**: Run Framework provider validation
- [ ] **Execute**: Run muxer validation  
- [ ] **Execute**: Run performance benchmarks
- [ ] **Execute**: Run full integration tests
- [ ] **Validation**: All tests pass
- [ ] **Validation**: Performance acceptable (< 5% overhead)

### 📋 Stage 4: First Resource Migration

#### ✅ Stage 4.1: Framework Resource Implementation
- [ ] Create `framework_resource_genesyscloud_routing_language.go`
- [ ] Implement `Metadata()` method
- [ ] Implement `Schema()` method
- [ ] Implement `Create()` method
- [ ] Implement `Read()` method
- [ ] Implement `Delete()` method
- [ ] Create `framework_data_source_genesyscloud_routing_language.go`
- [ ] Implement data source methods
- [ ] Create `framework_routing_language_proxy.go`
- [ ] Register Framework resource with provider
- [ ] **Validation**: Framework resource works

#### ✅ Stage 4.2: Migration Testing
- [ ] Create Framework resource tests
- [ ] Create Framework data source tests
- [ ] Run compatibility tests vs SDKv2
- [ ] Run performance comparison
- [ ] Run integration tests
- [ ] **Validation**: Identical behavior to SDKv2
- [ ] **Validation**: All tests pass

### 🚨 Critical Checkpoints

#### Checkpoint 1: Foundation Ready
```bash
# Must pass before proceeding to Stage 2
go test ./genesyscloud/provider_registrar/...
go test ./genesyscloud/resource_register/...
go test ./genesyscloud/provider/...
```

#### Checkpoint 2: Integration Ready
```bash
# Must pass before proceeding to Stage 3
terraform plan  # Should work unchanged
go test ./...   # All tests pass
```

#### Checkpoint 3: Migration Ready
```bash
# Must pass before proceeding to Stage 4
# Full regression + performance tests
go test -bench=. ./...
```

#### Checkpoint 4: Migration Complete
```bash
# Framework resource fully functional
go test ./genesyscloud/routing_language/framework_*
terraform plan -var="use_framework_provider=true"
```

### 🔄 Current Status

**Current Stage**: Stage 2 Complete - Moving to Stage 3
**Next Task**: Stage 3.1 - End-to-End Testing & Validation
**Estimated Time**: 2-3 hours

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

## Timeline and Roadmap

### Implementation Timeline

#### Week 1: Foundation
- **Days 1-3**: Registration system enhancement (Stage 1.1)
- **Days 4-5**: Framework provider configuration (Stage 1.2)

#### Week 2: Integration  
- **Days 1-3**: Muxer enhancement (Stage 2.1)
- **Days 4-5**: Testing infrastructure (Stage 2.2)

#### Week 3: Validation
- **Days 1-2**: End-to-end testing (Stage 3.1)
- **Days 3-5**: Performance and regression testing

#### Week 4-5: First Migration
- **Week 4**: Framework resource implementation (Stage 4.1)
- **Week 5**: Migration testing and validation (Stage 4.2)

### Future Phases Roadmap

#### Phase 3: Scale Migration (Q3 2024)
- **Target**: Migrate 60% of remaining resources
- **Focus**: Complex resources with dependencies
- **Optimization**: Performance tuning and optimization
- **Monitoring**: Production monitoring and feedback

#### Phase 4: Completion (Q4 2024)
- **Target**: Migrate all remaining resources
- **Cleanup**: Remove SDK v2 dependencies
- **Finalization**: Complete documentation and testing
- **Release**: Full Framework provider release

### Success Metrics:
- **Performance**: 20% improvement in resource operations
- **Reliability**: Zero breaking changes
- **Adoption**: Smooth user transition
- **Maintainability**: Reduced code complexity

---

## Notes and Considerations

### Key Insights
- This plan prioritizes safety and incremental progress
- Each stage can be rolled back independently
- The `routing_language` resource is chosen for its simplicity
- Success here creates a template for migrating other resources
- Performance monitoring is critical throughout

### Development Guidelines
- Always maintain backward compatibility
- Test thoroughly at each checkpoint
- Document lessons learned for future migrations
- Monitor performance impact continuously
- Keep rollback procedures ready

---

*This strategy document provides the complete roadmap for safely migrating the Genesys Cloud Terraform provider from SDKv2 to Plugin Framework using a muxing approach.*