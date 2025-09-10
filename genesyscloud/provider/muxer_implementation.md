# Terraform Provider Muxing Implementation Guide

## Table of Contents
1. [What is Terraform Plugin Framework?](#what-is-terraform-plugin-framework)
2. [Why Migrate from SDKv2 to Plugin Framework?](#why-migrate-from-sdkv2-to-plugin-framework)
3. [What is Muxing?](#what-is-muxing)
4. [Why Implement Muxing Logic?](#why-implement-muxing-logic)
5. [Incremental Implementation Strategy](#incremental-implementation-strategy)
6. [Phase 1 Implementation Design](#phase-1-implementation-design)
7. [Migration Benefits and Considerations](#migration-benefits-and-considerations)
8. [Future Phases and Roadmap](#future-phases-and-roadmap)
9. [Testing Strategy](#testing-strategy)
10. [Troubleshooting and Common Issues](#troubleshooting-and-common-issues)

---

## What is Terraform Plugin Framework?

The **Terraform Plugin Framework** is the modern, next-generation approach to building Terraform providers, introduced as an evolution from the older Plugin SDK v2. It represents a significant architectural improvement in how providers are developed and maintained.

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

---

## Why Implement Muxing Logic?

### 1. **Zero-Downtime Migration**
- **No breaking changes** to existing Terraform configurations
- **Gradual migration** of resources from SDK v2 to Framework
- **Backward compatibility** maintained throughout the process

### 2. **Risk Mitigation**
- **Incremental testing** of migrated resources
- **Rollback capability** if issues arise
- **Isolated changes** per resource

### 3. **Operational Continuity**
- **Existing infrastructure** continues to work unchanged
- **No immediate action required** from users
- **Smooth transition** for development teams

### 4. **Development Efficiency**
- **Parallel development** of new features
- **Resource-by-resource migration** approach
- **Reduced testing complexity** per migration

---

## Incremental Implementation Strategy

### Phase-Based Approach:

#### **Phase 1: Infrastructure Setup** ✅ (Current)
- Implement muxing infrastructure
- Upgrade SDK v2 provider to Protocol v6
- Create empty Framework provider
- Ensure all existing resources work unchanged

#### **Phase 2: Resource Migration** (Future)
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

## Phase 1 Implementation Design

### Architecture Overview:

```
main.go
├── provider.New() → NewMuxedProvider()
├── NewMuxedProvider()
│   ├── NewSDKv2Provider() → SDK v2 Provider
│   ├── tf5to6server.UpgradeServer() → Protocol v6
│   └── NewFrameworkProvider() → Framework Provider (empty)
└── tf6server.Serve() → Protocol v6 Server
```

### Key Files and Their Roles:

#### **1. `genesyscloud/provider/mux.go`**
```go
func NewMuxedProvider(version string, resources, dataSources map[string]*schema.Resource) 
    func() (func() tfprotov6.ProviderServer, error)
```
- **Purpose**: Creates the muxed provider factory
- **Functionality**: Combines SDK v2 and Framework providers
- **Protocol**: Upgrades SDK v2 to Protocol v6

#### **2. `genesyscloud/provider/framework_provider.go`**
```go
type GenesysCloudFrameworkProvider struct {
    version string
}
```
- **Purpose**: Framework provider implementation
- **Current State**: Empty schema, no resources
- **Future**: Will contain migrated resources

#### **3. `genesyscloud/provider/provider.go`**
```go
func NewSDKv2Provider(version string, resources, dataSources map[string]*schema.Resource) 
    func() *schema.Provider
```
- **Purpose**: Isolated SDK v2 provider creation
- **Functionality**: Maintains existing resource behavior
- **Compatibility**: Full backward compatibility

#### **4. `main.go`**
```go
muxFactoryFuncFunc := provider.New(version, providerResources, providerDataSources)
muxFactoryFunc, err := muxFactoryFuncFunc()
```
- **Purpose**: Entry point with muxing support
- **Server**: Uses `tf6server.Serve()` for Protocol v6
- **Debugging**: Supports managed debug mode

### Dependencies Added:

```go
// go.mod additions
github.com/hashicorp/terraform-plugin-framework v1.15.1
github.com/hashicorp/terraform-plugin-mux v0.20.0
github.com/hashicorp/terraform-plugin-go v0.28.0
```

### Configuration Flow:

1. **Terraform Core** sends request to provider
2. **Protocol v6 Server** receives request
3. **Muxer** determines target provider based on resource
4. **SDK v2 Provider** handles existing resources
5. **Framework Provider** handles migrated resources (future)
6. **Response** returned through same path

---

## Migration Benefits and Considerations

### Benefits:

#### **Immediate Benefits (Phase 1)**
- ✅ **Protocol v6 Support**: Modern Terraform compatibility
- ✅ **Zero Breaking Changes**: Existing configurations work unchanged
- ✅ **Future-Ready**: Infrastructure for gradual migration
- ✅ **Performance**: Eliminates SDK v2 to v6 conversion overhead

#### **Long-term Benefits (Future Phases)**
- 🚀 **Better Performance**: Native Protocol v6 implementation
- 🛠️ **Improved Developer Experience**: Modern Go patterns and type safety
- 🔧 **Enhanced Maintainability**: Cleaner, more testable code
- 📈 **Future-Proofing**: Support for upcoming Terraform features

### Considerations:

#### **Technical Considerations**
- **Memory Usage**: Slight increase due to muxing overhead
- **Complexity**: Additional layer of abstraction
- **Testing**: Need to test both provider paths
- **Debugging**: More complex debugging scenarios

#### **Operational Considerations**
- **Monitoring**: Track resource usage across providers
- **Documentation**: Update user guides and examples
- **Training**: Team education on new patterns
- **Support**: Handle issues across both implementations

---

## Future Phases and Roadmap

### Phase 2: Resource Migration (Q2 2024)
- **Target**: Migrate 20% of high-priority resources
- **Focus**: Simple, well-tested resources
- **Validation**: Comprehensive testing and performance validation
- **Documentation**: Update migration guides

### Phase 3: Scale Migration (Q3 2024)
- **Target**: Migrate 60% of remaining resources
- **Focus**: Complex resources with dependencies
- **Optimization**: Performance tuning and optimization
- **Monitoring**: Production monitoring and feedback

### Phase 4: Completion (Q4 2024)
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

## Testing Strategy

### Phase 1 Testing:
- **Unit Tests**: All existing tests continue to pass
- **Integration Tests**: End-to-end provider functionality
- **Regression Tests**: Ensure no breaking changes
- **Performance Tests**: Baseline performance measurement

### Future Phase Testing:
- **Migration Tests**: Validate resource migration
- **Compatibility Tests**: Ensure backward compatibility
- **Performance Tests**: Measure improvement gains
- **User Acceptance Tests**: Validate user experience

### Test Categories:

#### **1. Functional Testing**
- Resource CRUD operations
- Data source functionality
- Provider configuration
- Error handling

#### **2. Integration Testing**
- Multi-resource scenarios
- Provider switching
- State management
- Configuration validation

#### **3. Performance Testing**
- Resource operation timing
- Memory usage
- Concurrent operations
- Large-scale deployments

#### **4. Compatibility Testing**
- Terraform version compatibility
- Configuration format validation
- State file compatibility
- Provider version compatibility

---

## Troubleshooting and Common Issues

### Common Issues and Solutions:

#### **1. Provider Initialization Errors**
```
Error: Failed to create muxed provider factory
```
**Solution**: Check Protocol v6 server configuration and dependencies

#### **2. Resource Routing Issues**
```
Error: Resource not found in provider
```
**Solution**: Verify resource registration in appropriate provider

#### **3. Schema Conflicts**
```
Error: Schema validation failed
```
**Solution**: Ensure schema compatibility between providers

#### **4. Performance Degradation**
```
Warning: Slower resource operations
```
**Solution**: Monitor muxing overhead and optimize routing

### Debugging Techniques:

#### **1. Enable Debug Logging**
```hcl
provider "genesyscloud" {
  sdk_debug = true
  sdk_debug_file_path = "/tmp/genesyscloud-debug.log"
}
```

#### **2. Protocol v6 Debugging**
```bash
# Enable managed debug mode
terraform-provider-genesyscloud -debug
```

#### **3. Resource Tracing**
```go
// Add logging to muxer routing logic
log.Printf("Routing resource %s to provider %s", resourceName, providerType)
```

### Monitoring and Observability:

#### **1. Metrics to Track**
- Resource operation timing
- Provider routing distribution
- Error rates by provider
- Memory usage patterns

#### **2. Logging Strategy**
- Structured logging for all operations
- Correlation IDs for request tracing
- Provider-specific log levels
- Performance metrics collection

#### **3. Alerting**
- High error rates
- Performance degradation
- Resource routing failures
- Memory usage spikes

---

## Conclusion

The muxing implementation provides a robust, future-proof foundation for migrating the Genesys Cloud Terraform provider from SDK v2 to the Plugin Framework. This incremental approach ensures zero downtime, maintains backward compatibility, and provides a clear path for gradual migration while delivering immediate benefits through Protocol v6 support.

The Phase 1 implementation successfully establishes the muxing infrastructure, setting the stage for future resource migrations and long-term provider modernization.
