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
- [ ] Import `tf6muxserver` in `mux.go`
- [ ] Implement actual muxing with `NewMuxServer()`
- [ ] Add Framework provider to mux
- [ ] Add resource routing logic
- [ ] Add routing validation
- [ ] **Validation**: Muxer combines both providers
- [ ] **Validation**: All existing resources work unchanged

#### ✅ Stage 2.2: Testing Infrastructure
- [ ] Create `framework_provider_test.go`
- [ ] Create `mux_test.go`
- [ ] Create `framework_utils_test.go`
- [ ] Add Framework testing utilities
- [ ] Add muxer testing scenarios
- [ ] Add migration testing framework
- [ ] **Validation**: Testing infrastructure ready

### 📋 Stage 3: Validation

#### ✅ Stage 3.1: End-to-End Testing
- [ ] Run full regression test suite
- [ ] Run performance benchmarks
- [ ] Validate muxer routing
- [ ] Validate Framework provider
- [ ] Test rollback procedures
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

**Current Stage**: Stage 1.2 Complete - Moving to Stage 2.1
**Next Task**: Stage 2.1 - Muxer Enhancement
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

- **Stage 1.1 Complete**: Registration system enhanced with Framework support
- **Stage 1.2 Complete**: Framework provider fully configured with schema and API access
- **Provider Meta Sharing**: SDKv2 and Framework providers can share configuration
- **Next**: Need to implement actual muxing in Stage 2.1 
