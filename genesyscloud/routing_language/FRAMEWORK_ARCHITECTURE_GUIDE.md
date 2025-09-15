# Plugin Framework Architecture Guide

This document explains how the Plugin Framework works differently from SDKv2 and why removing SDKv2 functions doesn't break functionality for the `genesyscloud_routing_language` resource.

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

## Migration Fix: Test Initialization Update

### Problem Encountered During Migration
When migrating `genesyscloud_routing_language` to Framework-only, the `genesyscloud/resource_genesyscloud_init_test.go` file was failing to compile due to a call to a non-existent function:

```go
providerResources[routinglanguage.ResourceType] = routinglanguage.ResourceRoutingLanguage()
```

### Root Cause
The migration process involved:
1. Removing the SDKv2 `ResourceRoutingLanguage()` function
2. The global test initialization file was still trying to register the SDKv2 resource
3. This caused a compilation error: `"undefined: routinglanguage.ResourceRoutingLanguage"`

### Solution Applied
Removed the SDKv2 resource registration from the global test file since the resource is now Framework-only:

#### Before (Broken):
```go
providerResources[routinglanguage.ResourceType] = routinglanguage.ResourceRoutingLanguage()
```

#### After (Fixed):
```go
// routinglanguage.ResourceType removed - migrated to Framework-only
```

### Why This Fix Works
- **Framework resources** are registered through the muxed provider system, not through SDKv2 test initialization
- **Test compatibility**: The routing_language resource has its own Framework-specific test initialization in `genesyscloud/routing_language/genesyscloud_routing_language_init_test.go`
- **No functionality loss**: Framework resources are still fully testable through their own test files
- **Clean separation**: Global test file only handles SDKv2 resources, Framework resources handle their own testing

### Files Modified During Migration
- `genesyscloud/resource_genesyscloud_init_test.go` - Removed SDKv2 registration
- `genesyscloud/routing_language/resource_genesyscloud_routing_language_schema.go` - Framework-only registration
- `genesyscloud/routing_language/genesyscloud_routing_language_init_test.go` - Framework test initialization

### Migration Results
- ✅ Compilation errors resolved
- ✅ Test initialization works properly
- ✅ Framework resources remain fully testable
- ✅ No impact on other SDKv2 resources
- ✅ Clean architectural separation maintained

## Summary

Removing the SDKv2 function doesn't break anything because:

- ✅ **Framework resources** are registered through `SetRegistrar`, not test files
- ✅ **Muxed provider** automatically includes Framework resources
- ✅ **Framework tests** have their own initialization system
- ✅ **Runtime behavior** is handled by the Framework provider
- ✅ **No manual registration** needed in global test files

The Framework approach is more modern, cleaner, and doesn't require the manual wiring that SDKv2 needed. The `genesyscloud_routing_language` resource is now fully Framework-native with better architecture, testing, and maintainability.

This migration demonstrates the clean separation between SDKv2 and Framework resources, where each system manages its own registration and testing infrastructure without interfering with the other.