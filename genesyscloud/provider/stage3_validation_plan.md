# Stage 3: End-to-End Testing & Validation Plan

## 🎯 Objective
Validate that all Stage 1 and Stage 2 implementations work correctly together without breaking existing functionality.

## 📋 Validation Checklist

### 3.1 Compilation & Basic Tests
**Purpose**: Ensure all code compiles and basic functionality works

**Commands to run:**
```bash
# 1. Clean build test
go clean -cache
go mod tidy

# 2. Compilation test
go build ./genesyscloud/provider

# 3. Basic unit tests
go test ./genesyscloud/provider_registrar/...
go test ./genesyscloud/resource_register/...
go test ./genesyscloud/provider/...
```

**Expected Results:**
- ✅ All packages compile without errors
- ✅ All unit tests pass
- ✅ No import cycle errors
- ✅ No type redeclaration errors

---

### 3.2 Framework Provider Validation
**Purpose**: Validate Framework provider works independently

**Commands to run:**
```bash
# Framework provider specific tests
go test -v ./genesyscloud/provider/ -run TestFrameworkProvider
go test -v ./genesyscloud/provider/ -run TestFrameworkProviderServer
go test -v ./genesyscloud/provider/ -run TestFrameworkProviderConfigure
```

**Expected Results:**
- ✅ Framework provider creates successfully
- ✅ Provider schema is complete and valid
- ✅ Provider server can be instantiated
- ✅ Configuration validation works

---

### 3.3 Muxer Validation
**Purpose**: Validate muxing logic works correctly

**Commands to run:**
```bash
# Muxer specific tests
go test -v ./genesyscloud/provider/ -run TestNewMuxedProvider
go test -v ./genesyscloud/provider/ -run TestMuxedProviderWithDataSources
go test -v ./genesyscloud/provider/ -run TestMuxedProviderResourceRouting
```

**Expected Results:**
- ✅ SDKv2-only provider works (no Framework resources)
- ✅ Muxed provider works (with Framework resources)
- ✅ Resource routing works correctly
- ✅ Both provider types accessible in muxed mode

---

### 3.4 Registration System Validation
**Purpose**: Validate registration system enhancements

**Commands to run:**
```bash
# Registration system tests
go test -v ./genesyscloud/provider_registrar/ -run TestRegisterFramework
go test -v ./genesyscloud/resource_register/ -run TestFramework
```

**Expected Results:**
- ✅ Framework resources can be registered
- ✅ Framework data sources can be registered
- ✅ Provider type tracking works
- ✅ Resource retrieval methods work

---

### 3.5 Performance Validation
**Purpose**: Ensure no significant performance degradation

**Commands to run:**
```bash
# Performance benchmarks
go test -bench=. ./genesyscloud/provider/ -run TestMuxedProviderPerformance
go test -bench=BenchmarkProvider ./genesyscloud/provider/...

# Memory usage check
go test -benchmem -bench=. ./genesyscloud/provider/
```

**Expected Results:**
- ✅ Muxing overhead < 5%
- ✅ Memory usage reasonable
- ✅ No memory leaks in provider creation

---

### 3.6 Integration Validation
**Purpose**: Validate end-to-end integration

**Commands to run:**
```bash
# Full test suite
go test ./...

# Verbose output for debugging if needed
go test -v ./... | grep -E "(FAIL|PASS|ERROR)"

# Race condition detection
go test -race ./genesyscloud/provider/...
```

**Expected Results:**
- ✅ All existing tests pass
- ✅ No race conditions
- ✅ No breaking changes to existing functionality

---

## 🚨 Critical Success Criteria

### Must Pass Before Stage 4:
1. **Zero Test Failures**: All existing tests must pass
2. **No Breaking Changes**: Existing SDKv2 resources work unchanged
3. **Performance**: < 5% overhead from muxing
4. **Memory**: No memory leaks or excessive usage
5. **Compilation**: Clean build with no warnings

### Red Flags (Stop and Fix):
- ❌ Any existing test failures
- ❌ Import cycle errors
- ❌ Memory leaks
- ❌ Performance degradation > 5%
- ❌ Race conditions

---

## 📊 Validation Report Template

After running all commands, document results:

```
## Stage 3 Validation Results

### Compilation & Basic Tests
- [ ] Clean build: PASS/FAIL
- [ ] Unit tests: PASS/FAIL (X/Y passed)
- [ ] Import cycles: NONE/FOUND

### Framework Provider
- [ ] Provider creation: PASS/FAIL
- [ ] Schema validation: PASS/FAIL
- [ ] Server instantiation: PASS/FAIL

### Muxer
- [ ] SDKv2-only mode: PASS/FAIL
- [ ] Muxed mode: PASS/FAIL
- [ ] Resource routing: PASS/FAIL

### Performance
- [ ] Muxing overhead: X% (target: <5%)
- [ ] Memory usage: ACCEPTABLE/EXCESSIVE
- [ ] Race conditions: NONE/FOUND

### Integration
- [ ] Full test suite: PASS/FAIL (X/Y passed)
- [ ] Breaking changes: NONE/FOUND

### Overall Status: READY FOR STAGE 4 / NEEDS FIXES
```

---

## 🔧 Troubleshooting Guide

### Common Issues & Solutions:

**Import Cycle Errors:**
```bash
go mod graph | grep cycle
# Fix by removing circular dependencies
```

**Test Failures:**
```bash
go test -v ./path/to/failing/package
# Review specific failure and fix
```

**Performance Issues:**
```bash
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
# Analyze performance bottlenecks
```

**Memory Leaks:**
```bash
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
# Analyze memory usage
```

---

## ✅ Stage 3 Completion Criteria

Stage 3 is complete when:
1. All validation commands pass
2. Performance is acceptable
3. No breaking changes detected
4. Validation report shows all green
5. Ready to proceed to Stage 4 (first resource migration)

---

*This validation plan ensures we have a solid foundation before migrating actual resources in Stage 4.*