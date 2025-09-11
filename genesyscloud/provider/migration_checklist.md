# Phase 2 Migration Checklist
## Quick Reference for Implementation Progress

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

---

## 🚨 Critical Checkpoints

### Checkpoint 1: Foundation Ready
```bash
# Must pass before proceeding to Stage 2
go test ./genesyscloud/provider_registrar/...
go test ./genesyscloud/resource_register/...
go test ./genesyscloud/provider/...
```

### Checkpoint 2: Integration Ready
```bash
# Must pass before proceeding to Stage 3
terraform plan  # Should work unchanged
go test ./...   # All tests pass
```

### Checkpoint 3: Migration Ready
```bash
# Must pass before proceeding to Stage 4
# Full regression + performance tests
go test -bench=. ./...
```

### Checkpoint 4: Migration Complete
```bash
# Framework resource fully functional
go test ./genesyscloud/routing_language/framework_*
terraform plan -var="use_framework_provider=true"
```

---

## 🔄 Current Status

**Current Stage**: Stage 2 Complete - Moving to Stage 3
**Next Task**: Stage 3.1 - End-to-End Testing & Validation
**Estimated Time**: 2-3 hours

---

## 📞 Quick Commands

### Start Stage 1.1
```bash
# Edit the registrar interface
code genesyscloud/resource_register/resource_register.go
```

### Validate Progress
```bash
# Run tests after each change
go test ./genesyscloud/provider_registrar/...
go test ./genesyscloud/resource_register/...
```

### Check Current State
```bash
# Verify existing functionality still works
go test ./...
```

---

## 📝 Notes Section
*Use this space to track issues, decisions, and lessons learned*

- **Stage 1 Complete**: Registration system and Framework provider fully implemented
  - **Stage 1.1**: Registration system enhanced with Framework support
  - **Stage 1.2**: Framework provider fully configured with schema and API access
- **Stage 2 Complete**: Integration and testing infrastructure fully implemented
  - **Stage 2.1**: Muxer enhancement implemented with actual tf6muxserver integration
  - **Stage 2.2**: Comprehensive testing infrastructure with resolved type conflicts
- **Provider Meta Sharing**: SDKv2 and Framework providers can share configuration
- **Muxing Logic**: Automatically detects Framework resources and creates muxed provider when needed
- **Test Infrastructure**: Complete test coverage for both Framework provider and muxing functionality
- **Next**: Stage 3 - End-to-End Testing & Validation before first resource migration 
