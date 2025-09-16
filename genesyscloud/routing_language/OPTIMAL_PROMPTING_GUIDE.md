# SDKv2 to Plugin Framework Migration: Optimal Prompting Guide

This guide provides optimized AI prompting strategies for migrating Terraform resources from SDKv2 to Plugin Framework efficiently. Based on real migration experience with `genesyscloud_routing_language`.

## Migration Complexity Levels

### Simple Resources
- **Characteristics**: Basic CRUD operations, minimal dependencies, no complex integrations
- **Examples**: Simple configuration resources, basic data sources
- **Typical Issues**: Schema conversion, basic registration updates

### Medium Resources  
- **Characteristics**: Multiple dependencies, some integrations, moderate complexity
- **Examples**: Resources with relationships, custom validation, moderate business logic
- **Typical Issues**: Provider integration, test infrastructure, dependency management

### Complex Resources
- **Characteristics**: Heavy integrations, export functionality, complex business logic, multiple system touchpoints
- **Examples**: Core routing resources, resources with export/import, multi-system integrations
- **Typical Issues**: Muxed provider compatibility, export system integration, schema alignment, test infrastructure overhaul

## Our Case Study: Complex Resource Migration
- **Resource**: `genesyscloud_routing_language` 
- **Complexity**: High (export integration, muxed provider, test infrastructure)
- **Actual Journey**: Multiple sessions, 25+ prompts, reactive problem-solving
- **Issues Encountered**: Duplicate imports, circular dependencies, empty function implementations, Framework resource registration failures
- **Optimal Approach**: 1-2 sessions, 5-8 prompts, proactive system design with architectural analysis

## Optimal Prompts by Resource Complexity

### Simple Resource Migration Prompt

**For basic resources with minimal dependencies:**

```markdown
# Simple Framework Resource Migration

I need to migrate `[RESOURCE_NAME]` from SDKv2 to Plugin Framework.

## Resource Context:
- **Type**: [Resource/DataSource/Both]
- **Complexity**: Simple (basic CRUD, minimal dependencies)
- **Current Implementation**: [Brief description]

## Migration Requirements:
1. Convert SDKv2 implementation to Framework
2. Update schema definitions with proper types
3. Implement Framework CRUD methods
4. Update test files and registrations
5. Create basic documentation

## Technical Constraints:
- Maintain existing functionality
- Ensure test compatibility
- Follow Framework best practices

Please provide complete Framework implementation with updated tests and registration.
```

### Medium Resource Migration Prompt

**For resources with moderate complexity and dependencies:**

```markdown
# Medium Complexity Framework Migration

I need to migrate `[RESOURCE_NAME]` from SDKv2 to Plugin Framework with dependency management.

## Resource Context:
- **Type**: [Resource/DataSource/Both]
- **Complexity**: Medium (multiple dependencies, some integrations)
- **Dependencies**: [List key dependencies]
- **Integrations**: [List system integrations]

## Migration Requirements:
1. Framework implementation with dependency handling
2. Schema conversion with relationship management
3. Integration point updates
4. Comprehensive test coverage
5. Dependency documentation

## Technical Considerations:
- Dependency resolution strategy
- Integration compatibility
- Test infrastructure updates
- Performance considerations
- Circular dependency avoidance
- Framework resource registration patterns

Please analyze dependencies and provide complete migration solution with proper registrar implementation.
```

### Complex Resource Migration Prompt

**For resources with heavy integrations and system-wide impact:**

```markdown
# Complex Framework Resource Migration with System Integration

I need to migrate `[RESOURCE_NAME]` from SDKv2 to Plugin Framework with complete system integration.

## Resource Context:
- **Type**: [Resource/DataSource/Both]
- **Complexity**: High (heavy integrations, export functionality, system-wide impact)
- **System Integrations**: [List all: export, import, muxed provider, etc.]
- **Dependencies**: [List all dependent resources/systems]

## Comprehensive Requirements:

## Migration Requirements:
1. **Complete Framework Migration**: Convert SDKv2 resource to Framework-only implementation
2. **Remove SDKv2 Dependencies**: Clean up all SDKv2 code and registrations
3. **Muxed Provider Compatibility**: Ensure schema alignment between SDKv2 and Framework providers
4. **Export System Integration**: Make sure Framework resource works with export functionality
5. **Test Infrastructure**: Update all test files and initialization code
6. **Documentation**: Create comprehensive migration documentation

## Technical Constraints:
- Must work in muxed provider environment
- Export system must support Framework resources
- No breaking changes to existing SDKv2 resources
- Maintain backward compatibility
- Avoid circular import dependencies
- Proper Framework resource registration in test infrastructure

## Expected Deliverables:
- Framework resource implementation
- Framework data source implementation
- Updated provider schema alignment
- Export system modifications (if needed)
- Test infrastructure updates with proper Registrar interface implementation
- Complete documentation with migration patterns

## Critical Integration Points to Address:
1. **Test Infrastructure**: Implement proper Registrar interface, avoid empty placeholder functions
2. **Framework Registration**: Use SetRegistrar pattern, not manual exporter registration
3. **Circular Dependencies**: Use resource_register package, avoid provider_registrar imports in tests
4. **Duplicate Imports**: Clean import management for Framework packages
5. **Global Resource Storage**: Ensure Framework resources are stored in global registrar maps

Please analyze the current codebase, identify all integration points, anticipate system-wide impacts, and provide a complete solution with all necessary fixes.
```

## Migration Strategy Templates

### Pre-Migration Analysis Prompt

**Use this before starting any migration:**

```markdown
# Pre-Migration Analysis Request

Please analyze `[RESOURCE_NAME]` for Framework migration planning.

## Analysis Requirements:
1. **Complexity Assessment**: Classify as Simple/Medium/Complex
2. **Dependency Mapping**: Identify all dependencies and integrations
3. **Risk Assessment**: Potential issues and challenges
4. **Migration Strategy**: Recommended approach and timeline
5. **Testing Strategy**: Required test updates and coverage

## Current Implementation Review:
- Schema complexity and custom types
- Business logic complexity
- Integration points (export, import, etc.)
- Test coverage and infrastructure
- Documentation requirements

Please provide comprehensive analysis with migration recommendations.
```

### 2. Systematic Issue Resolution Prompt

**For handling multiple related issues:**

```markdown
# Systematic Technical Issue Resolution

I'm encountering multiple related issues with Framework resource integration. Please analyze and fix ALL related problems systematically:

## Current Issues:
1. Export system failing with "Resource type not defined"
2. Provider schema mismatches in muxed provider
3. Compilation errors in test files (duplicate imports, undefined variables)
4. Runtime panics during export operations
5. Circular import dependencies in test infrastructure
6. Empty placeholder functions not actually registering Framework resources
7. Framework resources not accessible to tfexporter due to improper registration

## Analysis Request:
1. **Root Cause Analysis**: Identify the underlying architectural issues
2. **Dependency Mapping**: Map all interconnected systems that need updates
3. **Comprehensive Solution**: Provide fixes for all related issues in proper order
4. **Testing Strategy**: Ensure all fixes work together

## Technical Context:
- Using muxed provider (SDKv2 + Framework)
- Export system needs Framework resource support
- Framework resources need to integrate with existing systems
- Must maintain backward compatibility
- Test infrastructure must implement proper Registrar interface
- Avoid circular dependencies between tfexporter and provider_registrar packages

## Specific Technical Requirements:
1. **Registrar Interface Implementation**: Test infrastructure must properly implement all Registrar methods
2. **Framework Resource Storage**: Use global resource_register maps, not local storage
3. **SetRegistrar Pattern**: Call resourcePackage.SetRegistrar() instead of manual registrations
4. **Import Management**: Clean Framework package imports, avoid duplicates
5. **Dependency Architecture**: Use resource_register package to avoid circular imports

Please provide a complete solution that addresses all issues holistically rather than piecemeal fixes.
```

### 3. Architecture-First Documentation Prompt

**For comprehensive documentation:**

```markdown
# Complete Technical Architecture Documentation

Please create comprehensive documentation that covers:

## Architecture Documentation:
1. **System Architecture**: How Framework and SDKv2 providers work together
2. **Migration Process**: Step-by-step migration methodology
3. **Integration Points**: All systems that interact with resources
4. **Troubleshooting Guide**: Common issues and solutions
5. **Best Practices**: Optimal approaches for future migrations

## Technical Deep-Dive:
- Explain the muxed provider architecture
- Document the export system integration
- Cover test infrastructure differences
- Include code examples and flow diagrams

## Consolidation Request:
- Create single comprehensive guide instead of multiple files
- Include all fixes and solutions applied
- Provide context for architectural decisions
- Make it a reference for future migrations

Please analyze our complete journey and create definitive documentation.
```

## Key Principles for Optimal Prompting

### 1. **Comprehensive Scope Definition**
```markdown
❌ Poor: "Fix the export issue"
✅ Better: "Fix export system to support Framework resources with complete integration"
```

### 2. **Context-Rich Requests**
```markdown
❌ Poor: "The resource isn't working"
✅ Better: "Framework resource fails in muxed provider environment with export system integration"
```

### 3. **Systematic Approach**
```markdown
❌ Poor: "Fix this error" (reactive)
✅ Better: "Analyze all integration points and provide comprehensive solution" (proactive)
```

### 4. **Deliverable Specification**
```markdown
❌ Poor: "Help me migrate this resource"
✅ Better: "Migrate resource with: implementation + tests + documentation + integration fixes"
```

### 5. **Architectural Awareness**
```markdown
❌ Poor: "Make this work"
✅ Better: "Ensure compatibility with muxed provider, export system, and test infrastructure"
```

## Optimal Single-Session Prompt

**If starting fresh, this single prompt could have achieved most of our work:**

```markdown
# Complete Framework Resource Migration with System Integration

I need to migrate `genesyscloud_routing_language` from SDKv2 to Framework with full system integration.

## Comprehensive Requirements:

### 1. Migration Scope
- Convert SDKv2 resource/datasource to Framework implementation
- Remove all SDKv2 code and dependencies
- Update provider registration system
- Ensure muxed provider compatibility

### 2. System Integration
- Fix export system to support Framework resources
- Resolve provider schema mismatches
- Update test infrastructure and initialization
- Handle any compilation or runtime errors

### 3. Architecture Considerations
- Muxed provider environment (SDKv2 + Framework)
- Export system integration requirements
- Test infrastructure compatibility
- Backward compatibility maintenance

### 4. Deliverables
- Complete Framework implementation
- All integration fixes
- Updated test infrastructure
- Comprehensive documentation
- Migration methodology guide

Please analyze the current codebase, identify all integration points, anticipate potential issues, and provide a complete end-to-end solution with proper documentation.

Focus on systematic, architectural solutions rather than incremental fixes.
```

## Benefits of Optimal Prompting

### Efficiency Gains:
- **Reduced Sessions**: 3 sessions → 1-2 sessions
- **Fewer Iterations**: ~20+ prompts → 5-8 prompts
- **Faster Resolution**: Proactive vs reactive problem-solving
- **Better Quality**: Comprehensive solutions vs piecemeal fixes

### Technical Benefits:
- **Holistic Solutions**: Address root causes, not just symptoms
- **Architectural Thinking**: Consider all integration points upfront
- **Systematic Approach**: Logical progression of fixes
- **Complete Documentation**: Single comprehensive guide

### Communication Benefits:
- **Clear Expectations**: Specific deliverables defined
- **Rich Context**: Full technical background provided
- **Scope Clarity**: Comprehensive requirements specified
- **Quality Standards**: Architecture and integration focus

## Lessons Learned

1. **Start with Architecture**: Understand the complete system before making changes
2. **Think Systematically**: Anticipate related issues and dependencies
3. **Be Comprehensive**: Request complete solutions, not partial fixes
4. **Provide Context**: Rich technical background enables better solutions
5. **Specify Deliverables**: Clear expectations lead to better outcomes
6. **Avoid Placeholder Functions**: Empty functions that don't actually implement required functionality cause runtime failures
7. **Understand Dependency Architecture**: Know which packages can import which to avoid circular dependencies
8. **Test Infrastructure Matters**: Framework resources need proper test infrastructure support, not just runtime support
9. **Global State Management**: Framework resources must be stored in global registrar maps for system-wide access
10. **SetRegistrar Pattern**: Use the established SetRegistrar pattern instead of manual resource registration

## Team Migration Workflow

### Phase 1: Planning
1. **Use Pre-Migration Analysis Prompt** to assess complexity
2. **Choose appropriate migration prompt** based on complexity
3. **Plan integration testing** strategy
4. **Coordinate with team** on dependencies

### Phase 2: Implementation  
1. **Execute migration prompt** with full context
2. **Review generated solution** comprehensively
3. **Test integration points** systematically
4. **Document changes** and decisions

### Phase 3: Validation
1. **Run comprehensive tests** (unit, integration, acceptance)
2. **Validate export/import** functionality if applicable
3. **Check muxed provider** compatibility
4. **Update team documentation**

## Common Pitfalls to Avoid

### ❌ Poor Prompting Practices:
- Vague requests: "Migrate this resource"
- Incremental fixes: "Fix this one error"
- Missing context: Not mentioning integrations
- Reactive approach: Fixing issues as they appear
- Ignoring test infrastructure: "Just make the resource work"
- Accepting placeholder functions: Empty functions that don't actually work
- Not considering circular dependencies: Importing packages without checking dependency graphs

### ✅ Best Prompting Practices:
- Comprehensive scope: "Migrate with full system integration including test infrastructure"
- Proactive analysis: "Identify and prevent potential issues including circular dependencies"
- Rich context: "Resource integrates with export system, muxed provider, and test infrastructure"
- Systematic approach: "Provide complete solution with testing strategy and proper registrar implementation"
- Architecture awareness: "Ensure proper dependency management and avoid circular imports"
- Implementation verification: "Ensure all functions actually implement required functionality, no placeholders"

## Resource-Specific Considerations

### Export-Enabled Resources
Always mention export system integration in your prompts:
```markdown
"This resource supports export functionality and must integrate with the export system"
```

### Core Infrastructure Resources
Emphasize system-wide impact:
```markdown
"This is a core resource with system-wide dependencies and muxed provider requirements"
```

### Data Sources with Complex Queries
Highlight query complexity:
```markdown
"Data source has complex filtering and query capabilities that must be preserved"
```

## Success Metrics

### Efficiency Indicators:
- **Session Count**: Target 1-2 sessions vs 3+ sessions
- **Prompt Count**: Target 5-8 prompts vs 20+ prompts  
- **Issue Resolution**: Proactive vs reactive problem-solving
- **Documentation Quality**: Comprehensive vs fragmented

### Quality Indicators:
- **Test Coverage**: Comprehensive from start
- **Integration Success**: All systems work together
- **Documentation Completeness**: Single comprehensive guide
- **Team Knowledge Transfer**: Clear migration methodology

## Template Customization Guide

### For Your Resource:
1. **Replace `[RESOURCE_NAME]`** with actual resource name
2. **Fill in complexity level** based on pre-analysis
3. **List specific integrations** (export, import, dependencies)
4. **Add resource-specific requirements**
5. **Include team-specific constraints**

### Example Customization:
```markdown
# Complex Framework Resource Migration with System Integration

I need to migrate `genesyscloud_routing_queue` from SDKv2 to Plugin Framework with complete system integration.

## Resource Context:
- **Type**: Resource and DataSource
- **Complexity**: High (core routing resource, export functionality, multiple dependencies)
- **System Integrations**: Export system, import functionality, muxed provider, dependency resolution
- **Dependencies**: routing_language, routing_skill, location, user
```

## Critical Technical Patterns Discovered

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

### 5. **Resource Package SetRegistrar Implementation**
```go
// In resource package (e.g., routing_language/resource_genesyscloud_routing_language_schema.go)
func SetRegistrar(regInstance registrar.Registrar) {
    // Register ALL three components in one call
    regInstance.RegisterFrameworkResource(ResourceType, NewFrameworkRoutingLanguageResource)
    regInstance.RegisterFrameworkDataSource(ResourceType, NewFrameworkRoutingLanguageDataSource)
    regInstance.RegisterExporter(ResourceType, RoutingLanguageExporter())
}
```

## Architecture Decision Records

### ADR-1: Test Infrastructure Registrar Implementation
**Decision**: Test infrastructure must implement full Registrar interface, not placeholder functions
**Rationale**: Framework resources need to be accessible to tfexporter via global registrar maps
**Impact**: Enables proper Framework resource testing and export functionality

### ADR-2: SetRegistrar Pattern for Framework Resources
**Decision**: Use SetRegistrar pattern instead of manual resource registration
**Rationale**: Ensures consistent registration of resource, datasource, and exporter together
**Impact**: Reduces registration errors and maintains consistency with main provider

### ADR-3: Dependency Architecture for Test Infrastructure
**Decision**: Use resource_register package, avoid provider_registrar imports in tests
**Rationale**: Prevents circular dependencies while maintaining functionality
**Impact**: Clean dependency graph and maintainable test infrastructure

### ADR-4: Global Resource Storage Strategy
**Decision**: Framework resources must be stored in global resource_register maps
**Rationale**: Enables system-wide access by tfexporter and other components
**Impact**: Proper integration between Framework resources and existing systems

This approach transforms reactive problem-solving into proactive system design and implementation, making migrations faster, more reliable, and better documented for the entire team.