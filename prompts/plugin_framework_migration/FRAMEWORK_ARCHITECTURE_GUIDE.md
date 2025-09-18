# Plugin Framework Architecture Guide

This document explains how the Plugin Framework works differently from SDKv2, the complete architecture for Framework-only resources, and the comprehensive fixes applied during the `genesyscloud_routing_language` migration including test infrastructure and export system integration.

## SDKv2 vs Plugin Framework Architecture

### SDKv2 Approach (Old Way)
In SDKv2, resources are registered through direct function calls in test files:

```go
// SDKv2 registration in test files
providerResources[routinglanguage.ResourceType] = routinglanguage.ResourceRoutingLanguage()
```

This requires:
1. A concrete `ResourceRoutingLanguage()` function that returns `*schema.Resource`
2. Manual registration in every test file that needs the resource
3. Direct coupling between test files and resource implementations

### Plugin Framework Approach (New Way)
In Plugin Framework, resources are registered through a factory pattern and dependency injection:

```go
// Framework registration through SetRegistrar
regInstance.RegisterFrameworkResource(ResourceType, NewFrameworkRoutingLanguageResource)
```

## How Framework Resources Work Without SDKv2 Functions

### 1. **Factory Pattern**
Instead of calling a concrete function, Framework uses factory functions:

```go
// Framework factory function
func NewFrameworkRoutingLanguageResource() resource.Resource {
    return &routingLanguageFrameworkResource{}
}
```

### 2. **Automatic Registration**
Framework resources are automatically registered through the provider system:

```go
// In routing_language/resource_genesyscloud_routing_language_schema.go
func SetRegistrar(regInstance registrar.Registrar) {
    regInstance.RegisterFrameworkResource(ResourceType, NewFrameworkRoutingLanguageResource)
    regInstance.RegisterFrameworkDataSource(ResourceType, NewFrameworkRoutingLanguageDataSource)
    regInstance.RegisterExporter(ResourceType, RoutingLanguageExporter())
}
```

### 3. **Muxed Provider Integration**
The muxed provider automatically discovers and includes Framework resources:

```go
// In provider/mux.go
frameworkProviderFactory := NewFrameworkProvider(version, frameworkResources, frameworkDataSources)

muxServer, err := tf6muxserver.NewMuxServer(ctx,
    func() tfprotov6.ProviderServer { return upgradedV6 },           // SDKv2 provider
    func() tfprotov6.ProviderServer {                                // Framework provider
        return providerserver.NewProtocol6(frameworkProviderFactory())()
    },
)
```

## Framework Testing Architecture

### Framework-Specific Test Initialization
Framework resources have their own test setup:

```go
// In routing_language/genesyscloud_routing_language_init_test.go
func (r *registerTestInstance) registerFrameworkTestResources() {
    r.frameworkResourceMapMutex.Lock()
    defer r.frameworkResourceMapMutex.Unlock()
    
    frameworkResources[ResourceType] = NewFrameworkRoutingLanguageResource
}
```

### Test Provider Creation
Framework tests create their own provider instances:

```go
// Framework test creates its own muxed provider
func TestAccFrameworkResourceRoutingLanguageBasic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
            "genesyscloud": func() (tfprotov6.ProviderServer, error) {
                // Create Framework provider with routing_language resource
                frameworkResources := map[string]func() frameworkresource.Resource{
                    ResourceType: NewFrameworkRoutingLanguageResource,
                }
                // ... create muxed provider
            },
        },
        // ... test steps
    })
}
```

## Why Removing SDKv2 Function Doesn't Break Anything

### 1. **Separate Registration Systems**
- **SDKv2**: Uses direct function calls in test files
- **Framework**: Uses factory registration through `SetRegistrar`

### 2. **Independent Test Infrastructure**
- **SDKv2 tests**: Use `resource_genesyscloud_init_test.go`
- **Framework tests**: Use their own test initialization files

### 3. **Muxed Provider Handles Both**
The muxed provider automatically:
- Routes SDKv2 resources to the SDKv2 provider
- Routes Framework resources to the Framework provider
- Presents a unified interface to Terraform

## Complete Flow Diagram

```
Terraform Request for genesyscloud_routing_language
                    ↓
            Muxed Provider Router
                    ↓
         (Detects Framework resource)
                    ↓
            Framework Provider
                    ↓
    NewFrameworkRoutingLanguageResource()
                    ↓
    routingLanguageFrameworkResource{}
                    ↓
        CRUD Operations (Create/Read/Update/Delete)
```

## Key Benefits of Framework Approach

1. **Cleaner Architecture**: No need for manual test registrations
2. **Type Safety**: Framework provides better type checking
3. **Modern APIs**: Uses newer Terraform plugin APIs
4. **Automatic Discovery**: Resources are automatically available
5. **Better Testing**: Framework-specific test utilities

## Resource Implementation Structure

### Framework Resource Structure
```go
type routingLanguageFrameworkResource struct {
    client *platformclientv2.RoutingApi
}

// Required Framework methods
func (r *routingLanguageFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse)
func (r *routingLanguageFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse)
func (r *routingLanguageFrameworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse)
func (r *routingLanguageFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse)
func (r *routingLanguageFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse)
func (r *routingLanguageFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse)
func (r *routingLanguageFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse)
func (r *routingLanguageFrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse)
```

### Data Model Structure
```go
type routingLanguageFrameworkResourceModel struct {
    Id   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
}
```

## Migration Benefits for routing_language

### Before (SDKv2)
- Manual test registration required
- Complex schema definitions
- Limited type safety
- Older plugin APIs

### After (Framework)
- Automatic registration through `SetRegistrar`
- Type-safe schema definitions
- Modern plugin APIs
- Better error handling
- Cleaner test architecture

## Complete Migration Journey: From SDKv2 to Framework-Only

### Phase 1: Initial Migration Issues
When migrating `genesyscloud_routing_language` to Framework-only, multiple issues were encountered across different system components:

#### **1. Test Initialization Compilation Errors**
**Problem**: Multiple test files across packages were failing to compile due to calls to non-existent SDKv2 functions:

```go
providerResources[routinglanguage.ResourceType] = routinglanguage.ResourceRoutingLanguage()
```

**Files Affected**: 6 test initialization files across different packages
- `genesyscloud/user/genesyscloud_user_init_test.go`
- `genesyscloud/recording_media_retention_policy/genesyscloud_recording_media_retention_policy_init_test.go`
- `genesyscloud/routing_email_route/genesyscloud_routing_email_route_init_test.go`
- `genesyscloud/task_management_workitem/genesyscloud_task_management_workitem_init_test.go`
- `genesyscloud/task_management_worktype/genesyscloud_task_management_worktype_init_test.go`
- `genesyscloud/tfexporter/tf_exporter_resource_test.go`

**Solution**: Removed all SDKv2 registrations and added explanatory comments indicating Framework-only migration.

#### **2. Export System Integration Issues**
**Problem**: Export functionality was broken due to missing `getAllRoutingLanguages` function that was deleted with SDKv2 implementation.

**Error**: `undefined: getAllRoutingLanguages`

**Solution**: Created new `GetAllRoutingLanguages` function using the existing proxy layer:
```go
func GetAllRoutingLanguages(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
    proxy := getRoutingLanguageProxy(clientConfig)
    languages, _, err := proxy.getAllRoutingLanguages(ctx, "")
    // ... implementation
}
```

#### **3. Cross-Package Test Dependencies**
**Problem**: Tests in other packages (like `routing_email_route`) that depended on `routing_language` were failing because they used SDKv2-only provider factories.

**Error**: `The provider hashicorp/genesyscloud does not support resource type "genesyscloud_routing_language"`

**Solution**: Updated tests to use muxed provider factories that include both SDKv2 and Framework resources:
```go
ProtoV6ProviderFactories: getMuxedProviderFactories()
```

### Phase 2: Test Infrastructure Architecture Issues

#### **4. TFExporter Test Infrastructure Problems**
**Problem**: The tfexporter test infrastructure had multiple critical issues:
- Duplicate imports causing compilation errors
- Empty placeholder functions that didn't actually register Framework resources
- Circular import dependencies
- Framework resources not accessible to the export system

**Root Cause**: The test infrastructure was not properly implementing the Registrar interface for Framework resources.

**Solution**: Complete overhaul of test infrastructure:
```go
// Proper Registrar interface implementation
func (r *registerTestInstance) RegisterFrameworkResource(resourceType string, resourceFactory func() frameworkresource.Resource) {
    currentFrameworkResources, currentFrameworkDataSources := registrar.GetFrameworkResources()
    if currentFrameworkResources == nil {
        currentFrameworkResources = make(map[string]func() frameworkresource.Resource)
    }
    currentFrameworkResources[resourceType] = resourceFactory
    registrar.SetFrameworkResources(currentFrameworkResources, currentFrameworkDataSources)
}

// Proper resource registration using SetRegistrar pattern
func (r *registerTestInstance) registerTestExporters() {
    regInstance := &registerTestInstance{}
    routinglanguage.SetRegistrar(regInstance) // This handles everything
}
```

#### **5. Framework Resource Registration Pattern**
**Problem**: Manual exporter registration was inconsistent with Framework-only approach.

**Solution**: Implemented proper SetRegistrar pattern:
```go
// In routing_language/resource_genesyscloud_routing_language_schema.go
func SetRegistrar(regInstance registrar.Registrar) {
    // Register ALL three components together
    regInstance.RegisterFrameworkResource(ResourceType, NewFrameworkRoutingLanguageResource)
    regInstance.RegisterFrameworkDataSource(ResourceType, NewFrameworkRoutingLanguageDataSource)
    regInstance.RegisterExporter(ResourceType, RoutingLanguageExporter())
}
```

### Phase 3: Architectural Improvements

#### **6. Centralized Provider Factory**
**Problem**: Multiple test files had duplicated `getMuxedProviderFactories()` functions.

**Solution**: Centralized the function in `genesyscloud/provider/provider_utils.go`:
```go
func GetMuxedProviderFactories(
    providerResources map[string]*schema.Resource,
    providerDataSources map[string]*schema.Resource,
    frameworkResources map[string]func() frameworkresource.Resource,
    frameworkDataSources map[string]func() datasource.DataSource,
) map[string]func() (tfprotov6.ProviderServer, error)
```

#### **7. Error Handling Compatibility**
**Problem**: Mixed SDKv2 and Framework utility functions causing compilation errors.

**Solution**: Used Framework-compatible error handling:
- Direct error messages in `resp.Diagnostics.AddError()`
- `retry.RetryContext()` for Framework-compatible retry logic
- Simple `fmt.Errorf()` for error creation

### Migration Results - Complete Success
- ✅ **Compilation errors resolved** across all packages
- ✅ **Export functionality restored** with Framework-compatible implementation
- ✅ **Cross-package test dependencies working** through muxed providers
- ✅ **Test infrastructure properly supporting Framework resources**
- ✅ **Clean architectural separation** between SDKv2 and Framework
- ✅ **Centralized provider factory** eliminating code duplication
- ✅ **Framework-only registration** working correctly

## Key Architectural Patterns Discovered

### 1. **Test Infrastructure Registrar Pattern**
```go
// ❌ Wrong: Empty placeholder functions
func (r *registerTestInstance) RegisterFrameworkResource(resourceType string, resourceFactory func() frameworkresource.Resource) {
    // This is a no-op - WRONG!
}

// ✅ Correct: Actual implementation that stores resources
func (r *registerTestInstance) RegisterFrameworkResource(resourceType string, resourceFactory func() frameworkresource.Resource) {
    currentFrameworkResources, currentFrameworkDataSources := registrar.GetFrameworkResources()
    if currentFrameworkResources == nil {
        currentFrameworkResources = make(map[string]func() frameworkresource.Resource)
    }
    currentFrameworkResources[resourceType] = resourceFactory
    registrar.SetFrameworkResources(currentFrameworkResources, currentFrameworkDataSources)
}
```

### 2. **SetRegistrar Pattern vs Manual Registration**
```go
// ❌ Wrong: Manual exporter registration
func (r *registerTestInstance) registerTestExporters() {
    RegisterExporter(routinglanguage.ResourceType, routinglanguage.RoutingLanguageExporter())
}

// ✅ Correct: Use SetRegistrar pattern
func (r *registerTestInstance) registerTestExporters() {
    regInstance := &registerTestInstance{}
    routinglanguage.SetRegistrar(regInstance) // This handles resource, datasource, AND exporter
}
```

### 3. **Dependency Architecture Pattern**
```go
// ❌ Wrong: Creates circular dependency
import providerRegistrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"

// ✅ Correct: Use resource_register to avoid cycles
import registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
```

### 4. **Framework Import Management**
```go
// ❌ Wrong: Duplicate imports
import (
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    // ... other imports ...
    "github.com/hashicorp/terraform-plugin-framework/datasource" // DUPLICATE!
    "github.com/hashicorp/terraform-plugin-framework/resource"   // DUPLICATE!
)

// ✅ Correct: Clean imports with aliases
import (
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
)
```

## Architecture Decision Records

### ADR-1: Framework-Only Migration Strategy
**Decision**: Complete replacement of SDKv2 with Framework implementation
**Rationale**: Eliminates complexity of maintaining parallel implementations
**Impact**: Simplified architecture, single code path to maintain

### ADR-2: Test Infrastructure Registrar Implementation
**Decision**: Test infrastructure must implement full Registrar interface, not placeholder functions
**Rationale**: Framework resources need to be accessible to tfexporter via global registrar maps
**Impact**: Enables proper Framework resource testing and export functionality

### ADR-3: SetRegistrar Pattern for Framework Resources
**Decision**: Use SetRegistrar pattern instead of manual resource registration
**Rationale**: Ensures consistent registration of resource, datasource, and exporter together
**Impact**: Reduces registration errors and maintains consistency with main provider

### ADR-4: Dependency Architecture for Test Infrastructure
**Decision**: Use resource_register package, avoid provider_registrar imports in tests
**Rationale**: Prevents circular dependencies while maintaining functionality
**Impact**: Clean dependency graph and maintainable test infrastructure

### ADR-5: Centralized Muxed Provider Factory
**Decision**: Centralize duplicated provider factory functions
**Rationale**: Eliminates code duplication and provides single source of truth
**Impact**: Easier maintenance and consistent behavior across tests

## Summary

The Framework-only migration of `genesyscloud_routing_language` demonstrates a complete architectural transformation:

### ✅ **Framework Architecture Benefits**
- **Modern Plugin APIs**: Uses latest Terraform plugin Framework
- **Type Safety**: Better type checking and validation
- **Simplified Registration**: Automatic discovery through SetRegistrar pattern
- **Clean Separation**: Framework and SDKv2 resources coexist without interference
- **Better Testing**: Framework-specific test utilities and patterns

### ✅ **Migration Success Factors**
- **Complete System Integration**: Export system, test infrastructure, cross-package dependencies all working
- **Proper Registrar Implementation**: Test infrastructure properly supports Framework resources
- **Clean Dependency Architecture**: No circular imports, proper separation of concerns
- **Centralized Provider Management**: Single source of truth for muxed provider factories
- **Framework-Compatible Error Handling**: Proper error patterns for Framework resources

### ✅ **Template for Future Migrations**
This migration establishes proven patterns for migrating other resources:

1. **Implement Framework resource/datasource** using existing proxy
2. **Create comprehensive Framework tests** with proper provider factories
3. **Update registration** to use SetRegistrar pattern
4. **Remove SDKv2 files** completely after Framework implementation is working
5. **Update test infrastructure** to properly implement Registrar interface
6. **Fix cross-package dependencies** using muxed provider factories
7. **Validate export functionality** works with Framework resources

### ✅ **Architectural Insights**
- **Framework resources** are registered through `SetRegistrar`, not manual test file registration
- **Muxed provider** automatically includes Framework resources for cross-package compatibility
- **Test infrastructure** must properly implement Registrar interface, not use placeholder functions
- **Global resource storage** in resource_register package enables system-wide Framework resource access
- **Dependency management** requires careful attention to avoid circular imports

## 🔥 **Major Discovery: Cross-Package Code Duplication Anti-Pattern**

### **Critical Architectural Issue Discovered During routing_wrapupcode Migration**

Following the routing_language template, the routing_wrapupcode migration **uncovered a significant code duplication problem** affecting multiple packages:

#### **Problem Identified:**
- **6 packages** had nearly identical custom `getMuxedProviderFactoriesFor[Package]()` functions
- **11 test files** contained duplicated Framework resource inclusion logic
- **Code duplication** across packages creating maintenance burden and inconsistency risk

#### **Anti-Pattern Example:**
```go
// ❌ Code Duplication: Each package had its own nearly identical function
func getMuxedProviderFactoriesForOutboundCampaign() map[string]func() (tfprotov6.ProviderServer, error) {
    return provider.GetMuxedProviderFactories(
        providerResources, providerDataSources,
        map[string]func() frameworkresource.Resource{
            routingWrapupcode.ResourceType: routingWrapupcode.NewRoutingWrapupcodeFrameworkResource,
        },
        map[string]func() datasource.DataSource{
            routingWrapupcode.ResourceType: routingWrapupcode.NewRoutingWrapupcodeFrameworkDataSource,
        },
    )
}

func getMuxedProviderFactoriesForRoutingQueue() map[string]func() (tfprotov6.ProviderServer, error) {
    // Nearly identical implementation - CODE DUPLICATION!
    return provider.GetMuxedProviderFactories(/* same pattern */)
}
// ... 4+ more nearly identical functions
```

#### **Solution Implemented:**
```go
// ✅ DRY Principle: Direct usage in test cases (no custom functions)
resource.Test(t, resource.TestCase{
    PreCheck: func() { util.TestAccPreCheck(t) },
    ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
        providerResources,
        providerDataSources,
        map[string]func() frameworkresource.Resource{
            routingWrapupcode.ResourceType: routingWrapupcode.NewRoutingWrapupcodeFrameworkResource,
        },
        map[string]func() datasource.DataSource{
            routingWrapupcode.ResourceType: routingWrapupcode.NewRoutingWrapupcodeFrameworkDataSource,
        },
    ),
    // ... test steps
})
```

#### **Required Imports for Framework Integration:**
```go
import (
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
    routingWrapupcode "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
)
```

#### **Packages Fixed:**
- `genesyscloud/routing_queue` - 2 files updated
- `genesyscloud/outbound_campaign` - 2 files updated
- `genesyscloud/outbound_sequence` - 2 files updated
- `genesyscloud/outbound_campaignrule` - 2 files updated
- `genesyscloud/outbound_callanalysisresponseset` - 2 files updated
- `genesyscloud/outbound_wrapupcode_mappings` - 1 file updated
- `genesyscloud/tfexporter` - Special case (includes both routing_language and routing_wrapupcode)

#### **Architectural Benefits:**
- ✅ **Eliminated Code Duplication** - Removed 6 nearly identical custom functions
- ✅ **Established DRY Principle** - Single pattern used across all packages
- ✅ **Improved Maintainability** - Changes only need to be made in one place
- ✅ **Consistent Architecture** - Standardized approach across all test files
- ✅ **Reduced Technical Debt** - Cleaner codebase with better structure
- ✅ **Template for Future** - Established pattern for future Framework migrations

### **New Standard: Framework Integration Pattern**

All future Framework migrations should follow this **consistent, DRY pattern**:

1. **No Custom Provider Factory Functions** - Use direct `provider.GetMuxedProviderFactories()` calls
2. **Consistent Imports** - Standard set of Framework imports in all test files
3. **Direct Usage** - Include Framework resource maps directly in test cases
4. **Reusable Pattern** - Same approach across all packages and resources

This discovery transformed the Framework migration from a single-resource effort into a **codebase-wide architectural improvement** that benefits all future development.

## 🏆 **Dual Migration Success: Template Validation**

The routing_language migration established the template, and the routing_wrapupcode migration **validated and improved** it:

- ✅ **Template Proven** - Works across different resource types and complexities
- ✅ **Architecture Improved** - Eliminated code duplication anti-pattern
- ✅ **Patterns Established** - Consistent approach across all packages
- ✅ **Foundation Created** - Ready for team-wide Framework migration

The `genesyscloud_routing_language` resource is now fully Framework-native with modern architecture, comprehensive testing, proper export integration, and serves as a **proven, battle-tested template** for future Framework migrations that also **improves overall codebase architecture**.