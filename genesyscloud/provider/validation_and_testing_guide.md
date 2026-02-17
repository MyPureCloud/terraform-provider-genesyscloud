# Validation & Testing Guide

## Table of Contents
1. [Stage 3: End-to-End Testing & Validation Plan](#stage-3-end-to-end-testing--validation-plan)
2. [Validation Commands and Procedures](#validation-commands-and-procedures)
3. [Testing Strategy and Coverage](#testing-strategy-and-coverage)
4. [Performance and Benchmarking](#performance-and-benchmarking)
5. [Troubleshooting and Common Issues](#troubleshooting-and-common-issues)
6. [Test Infrastructure and Utilities](#test-infrastructure-and-utilities)
7. [Validation Report Templates](#validation-report-templates)

---

## Stage 3: End-to-End Testing & Validation Plan

### 🎯 Objective
Validate that all Stage 1 and Stage 2 implementations work correctly together without breaking existing functionality.

### 📋 Validation Checklist

#### 3.1 Compilation & Basic Tests
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

#### 3.2 Framework Provider Validation
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

#### 3.3 Muxer Validation
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

#### 3.4 Registration System Validation
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

#### 3.5 Performance Validation
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

#### 3.6 Integration Validation
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

### 🚨 Critical Success Criteria

#### Must Pass Before Stage 4:
1. **Zero Test Failures**: All existing tests must pass
2. **No Breaking Changes**: Existing SDKv2 resources work unchanged
3. **Performance**: < 5% overhead from muxing
4. **Memory**: No memory leaks or excessive usage
5. **Compilation**: Clean build with no warnings

#### Red Flags (Stop and Fix):
- ❌ Any existing test failures
- ❌ Import cycle errors
- ❌ Memory leaks
- ❌ Performance degradation > 5%
- ❌ Race conditions

---

## Validation Commands and Procedures

### Quick Validation Commands

#### Start Stage 1.1
```bash
# Edit the registrar interface
code genesyscloud/resource_register/resource_register.go
```

#### Validate Progress
```bash
# Run tests after each change
go test ./genesyscloud/provider_registrar/...
go test ./genesyscloud/resource_register/...
```

#### Check Current State
```bash
# Verify existing functionality still works
go test ./...
```

### Comprehensive Test Suite

#### Full Regression Testing
```bash
# Clean environment
go clean -cache
go mod tidy

# Full compilation check
go build ./...

# Complete test suite
go test -v ./... > test_results.log 2>&1

# Check for failures
grep -E "(FAIL|ERROR)" test_results.log
```

#### Performance Benchmarking
```bash
# Baseline performance measurement
go test -bench=. -benchmem ./genesyscloud/provider/ > baseline_perf.log

# Compare performance after changes
go test -bench=. -benchmem ./genesyscloud/provider/ > current_perf.log

# Analyze differences
diff baseline_perf.log current_perf.log
```

#### Memory Leak Detection
```bash
# Memory profiling
go test -memprofile=mem.prof -bench=. ./genesyscloud/provider/

# Analyze memory usage
go tool pprof mem.prof
```

#### Race Condition Detection
```bash
# Race condition testing
go test -race -v ./genesyscloud/provider/...

# Extended race testing
go test -race -count=10 ./genesyscloud/provider/...
```

### Terraform Integration Testing

#### Basic Terraform Operations
```bash
# Initialize Terraform
terraform init

# Validate configuration
terraform validate

# Plan without changes
terraform plan

# Check for provider loading issues
terraform providers
```

#### Provider Schema Validation
```bash
# Export provider schema
terraform providers schema -json > provider_schema.json

# Validate schema structure
jq '.provider_schemas' provider_schema.json
```

---

## Testing Strategy and Coverage

### Test Categories

#### 1. Unit Tests
**Purpose**: Test individual components in isolation

**Coverage Areas**:
- Provider registration functions
- Framework provider configuration
- Muxer logic components
- Utility functions

**Example Tests**:
```bash
# Provider registrar tests
go test -v ./genesyscloud/provider_registrar/ -run TestRegisterFramework

# Resource register tests
go test -v ./genesyscloud/resource_register/ -run TestFramework

# Framework provider tests
go test -v ./genesyscloud/provider/ -run TestFrameworkProvider
```

#### 2. Integration Tests
**Purpose**: Test component interactions

**Coverage Areas**:
- SDKv2 and Framework provider integration
- Muxer routing between providers
- Configuration sharing
- Resource lifecycle operations

**Example Tests**:
```bash
# Muxer integration tests
go test -v ./genesyscloud/provider/ -run TestMuxedProvider

# Provider routing tests
go test -v ./genesyscloud/provider/ -run TestResourceRouting
```

#### 3. Performance Tests
**Purpose**: Validate performance characteristics

**Coverage Areas**:
- Provider creation overhead
- Resource operation latency
- Memory usage patterns
- Concurrent operation handling

**Example Tests**:
```bash
# Performance benchmarks
go test -bench=BenchmarkProvider ./genesyscloud/provider/

# Memory benchmarks
go test -bench=. -benchmem ./genesyscloud/provider/
```

#### 4. Compatibility Tests
**Purpose**: Ensure backward compatibility

**Coverage Areas**:
- Existing resource behavior
- Configuration compatibility
- State file compatibility
- API interaction consistency

**Example Tests**:
```bash
# Compatibility validation
go test -v ./... -run TestBackwardCompatibility

# Schema consistency tests
go test -v ./genesyscloud/provider/ -run TestSchemaConsistency
```

### Test Infrastructure Components

#### Mock Objects
```go
// Mock Framework resource
type mockFrameworkResource struct{}

func (r *mockFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_test_resource"
}

// Mock Framework data source
type mockFrameworkDataSource struct{}

func (d *mockFrameworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_test_data_source"
}
```

#### Test Utilities
```go
// Provider creation helpers
func createTestSDKv2Provider() *schema.Provider
func createTestFrameworkProvider() provider.Provider
func createTestMuxedProvider() tfprotov6.ProviderServer

// Validation helpers
func validateProviderSchema(provider tfprotov6.ProviderServer) error
func validateResourceRouting(provider tfprotov6.ProviderServer) error
```

#### Test Data Management
```go
// Test configuration templates
const testProviderConfig = `
provider "genesyscloud" {
  oauthclient_id     = "test-client-id"
  oauthclient_secret = "test-client-secret"
  aws_region         = "us-east-1"
}
`

// Test resource configurations
const testResourceConfig = `
resource "genesyscloud_routing_language" "test" {
  name = "test-language"
}
`
```

---

## Performance and Benchmarking

### Performance Metrics

#### Key Performance Indicators
1. **Provider Creation Time**: Time to instantiate provider
2. **Resource Operation Latency**: CRUD operation timing
3. **Memory Usage**: Heap allocation and garbage collection
4. **Concurrent Performance**: Multi-threaded operation handling

#### Benchmark Tests

##### Provider Creation Benchmark
```go
func BenchmarkProviderCreation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        provider := createTestMuxedProvider()
        _ = provider
    }
}
```

##### Resource Operation Benchmark
```go
func BenchmarkResourceOperations(b *testing.B) {
    provider := createTestMuxedProvider()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Simulate resource CRUD operations
        performResourceOperations(provider)
    }
}
```

##### Memory Usage Benchmark
```go
func BenchmarkMemoryUsage(b *testing.B) {
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        provider := createTestMuxedProvider()
        performOperations(provider)
    }
}
```

### Performance Targets

#### Acceptable Performance Thresholds
- **Muxing Overhead**: < 5% compared to SDKv2-only
- **Memory Increase**: < 10% additional memory usage
- **Latency Impact**: < 2ms additional latency per operation
- **Concurrent Performance**: No degradation in multi-threaded scenarios

#### Performance Monitoring Commands
```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./genesyscloud/provider/
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./genesyscloud/provider/
go tool pprof mem.prof

# Trace analysis
go test -trace=trace.out -bench=. ./genesyscloud/provider/
go tool trace trace.out
```

---

## Troubleshooting and Common Issues

### Common Issues & Solutions

#### Import Cycle Errors
```bash
# Detect import cycles
go mod graph | grep cycle

# Fix by removing circular dependencies
# Refactor code to break dependency chains
```

**Solution Strategy**:
- Use dependency injection
- Create interface abstractions
- Move shared code to separate packages

#### Test Failures
```bash
# Run specific failing test
go test -v ./path/to/failing/package -run TestSpecificTest

# Get detailed failure information
go test -v -failfast ./...
```

**Common Causes**:
- Type redeclaration conflicts
- Schema mismatches between providers
- Missing mock implementations
- Race conditions in concurrent tests

#### Performance Issues
```bash
# Profile performance bottlenecks
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Analyze memory allocations
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

**Optimization Strategies**:
- Lazy initialization of providers
- Efficient resource routing
- Memory pool usage
- Goroutine management

#### Memory Leaks
```bash
# Detect memory leaks
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Check for goroutine leaks
go test -race -v ./...
```

**Prevention Techniques**:
- Proper resource cleanup
- Context cancellation
- Goroutine lifecycle management
- Connection pool management

### Debugging Techniques

#### 1. Enable Debug Logging
```hcl
provider "genesyscloud" {
  sdk_debug = true
  sdk_debug_file_path = "/tmp/genesyscloud-debug.log"
}
```

#### 2. Protocol v6 Debugging
```bash
# Enable managed debug mode
terraform-provider-genesyscloud -debug
```

#### 3. Resource Tracing
```go
// Add logging to muxer routing logic
log.Printf("Routing resource %s to provider %s", resourceName, providerType)
```

#### 4. Test Debugging
```bash
# Run tests with verbose output
go test -v ./genesyscloud/provider/ -run TestSpecificTest

# Run single test with debugging
go test -v -run TestSpecificTest ./genesyscloud/provider/
```

### Error Patterns and Solutions

#### Schema Validation Errors
```
Error: Invalid Provider Server Combination: The combined provider has differing provider schema implementations
```

**Solution**:
- Ensure identical schemas between SDKv2 and Framework providers
- Check DescriptionKind consistency
- Validate sensitive field settings
- Verify environment variable references

#### Resource Routing Errors
```
Error: Resource not found in provider
```

**Solution**:
- Verify resource registration in appropriate provider
- Check provider type tracking
- Validate muxer routing logic
- Ensure resource factory functions are correct

#### Provider Initialization Errors
```
Error: Failed to create muxed provider factory
```

**Solution**:
- Check Protocol v6 server configuration
- Validate provider dependencies
- Ensure proper factory function wrapping
- Verify import statements

---

## Test Infrastructure and Utilities

### Test Helper Functions

#### Provider Creation Helpers
```go
// Create test SDKv2 provider
func createTestSDKv2Provider() *schema.Provider {
    return &schema.Provider{
        Schema: map[string]*schema.Schema{
            "oauthclient_id": {
                Type:     schema.TypeString,
                Optional: true,
            },
        },
        ResourcesMap: map[string]*schema.Resource{
            "test_resource": testResource(),
        },
    }
}

// Create test Framework provider
func createTestFrameworkProvider() provider.Provider {
    return &testFrameworkProvider{
        version: "test",
    }
}

// Create test muxed provider
func createTestMuxedProvider() (func() tfprotov6.ProviderServer, error) {
    // Implementation details...
}
```

#### Validation Helpers
```go
// Validate provider schema consistency
func validateProviderSchemas(sdkProvider *schema.Provider, frameworkProvider provider.Provider) error {
    // Schema comparison logic...
}

// Validate resource routing
func validateResourceRouting(muxedProvider tfprotov6.ProviderServer) error {
    // Routing validation logic...
}

// Validate performance metrics
func validatePerformanceMetrics(baseline, current PerformanceMetrics) error {
    // Performance comparison logic...
}
```

#### Mock Implementations
```go
// Mock Framework resource with complete interface
type testFrameworkResource struct{}

func (r *testFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_test_resource"
}

func (r *testFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed: true,
            },
            "name": schema.StringAttribute{
                Required: true,
            },
        },
    }
}

// Implement all CRUD methods...
```

### Test Data and Configurations

#### Test Configuration Templates
```go
const (
    testProviderConfigBasic = `
provider "genesyscloud" {
  oauthclient_id     = "test-client-id"
  oauthclient_secret = "test-client-secret"
  aws_region         = "us-east-1"
}
`

    testResourceConfigBasic = `
resource "genesyscloud_routing_language" "test" {
  name = "test-language"
}
`

    testDataSourceConfigBasic = `
data "genesyscloud_routing_language" "test" {
  name = "test-language"
}
`
)
```

#### Test Environment Setup
```go
// Setup test environment
func setupTestEnvironment() {
    // Set environment variables
    os.Setenv("GENESYSCLOUD_OAUTHCLIENT_ID", "test-client-id")
    os.Setenv("GENESYSCLOUD_OAUTHCLIENT_SECRET", "test-client-secret")
    os.Setenv("GENESYSCLOUD_AWS_REGION", "us-east-1")
}

// Cleanup test environment
func cleanupTestEnvironment() {
    // Clean up resources
    // Reset environment variables
    // Clear test data
}
```

### Automated Test Scripts

#### Linux/Mac Test Runner (`test_runner.sh`)
```bash
#!/bin/bash

echo "Running comprehensive test suite..."

# Clean environment
go clean -cache
go mod tidy

# Compilation tests
echo "Testing compilation..."
go build ./genesyscloud/provider || exit 1

# Unit tests
echo "Running unit tests..."
go test ./genesyscloud/provider_registrar/... || exit 1
go test ./genesyscloud/resource_register/... || exit 1
go test ./genesyscloud/provider/... || exit 1

# Integration tests
echo "Running integration tests..."
go test -v ./genesyscloud/provider/ -run TestMuxedProvider || exit 1

# Performance tests
echo "Running performance tests..."
go test -bench=. ./genesyscloud/provider/ || exit 1

echo "All tests passed!"
```

#### Windows Test Runner (`test_runner.bat`)
```batch
@echo off
echo Running comprehensive test suite...

REM Clean environment
go clean -cache
go mod tidy

REM Compilation tests
echo Testing compilation...
go build ./genesyscloud/provider
if %errorlevel% neq 0 exit /b %errorlevel%

REM Unit tests
echo Running unit tests...
go test ./genesyscloud/provider_registrar/...
if %errorlevel% neq 0 exit /b %errorlevel%

go test ./genesyscloud/resource_register/...
if %errorlevel% neq 0 exit /b %errorlevel%

go test ./genesyscloud/provider/...
if %errorlevel% neq 0 exit /b %errorlevel%

REM Integration tests
echo Running integration tests...
go test -v ./genesyscloud/provider/ -run TestMuxedProvider
if %errorlevel% neq 0 exit /b %errorlevel%

REM Performance tests
echo Running performance tests...
go test -bench=. ./genesyscloud/provider/
if %errorlevel% neq 0 exit /b %errorlevel%

echo All tests passed!
```

---

## Validation Report Templates

### Stage 3 Validation Results Template

```markdown
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

### Performance Benchmark Report Template

```markdown
## Performance Benchmark Report

### Test Environment
- Go Version: X.X.X
- OS: Windows/Linux/Mac
- CPU: X cores
- Memory: X GB

### Baseline Metrics (SDKv2 Only)
- Provider Creation: X ms
- Resource Operations: X ms/op
- Memory Usage: X MB
- Allocations: X allocs/op

### Current Metrics (Muxed Provider)
- Provider Creation: X ms (+Y%)
- Resource Operations: X ms/op (+Y%)
- Memory Usage: X MB (+Y%)
- Allocations: X allocs/op (+Y%)

### Performance Impact Analysis
- Overhead within acceptable limits: YES/NO
- Memory usage reasonable: YES/NO
- No performance regressions: YES/NO

### Recommendations
- [List any performance optimization recommendations]
```

### Test Coverage Report Template

```markdown
## Test Coverage Report

### Unit Test Coverage
- Provider Registrar: X% coverage
- Resource Register: X% coverage
- Framework Provider: X% coverage
- Muxer Logic: X% coverage

### Integration Test Coverage
- Provider Integration: X scenarios tested
- Resource Routing: X scenarios tested
- Configuration Sharing: X scenarios tested

### Performance Test Coverage
- Benchmark Tests: X tests
- Memory Tests: X tests
- Concurrency Tests: X tests

### Missing Coverage Areas
- [List areas that need additional test coverage]

### Test Quality Metrics
- Total Tests: X
- Passing Tests: X
- Failing Tests: X
- Test Execution Time: X seconds
```

---

*This comprehensive validation and testing guide ensures thorough validation of the SDKv2 to Plugin Framework migration at every stage.*