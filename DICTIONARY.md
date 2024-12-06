# Dictionary

This document defines the domain-specific language and naming conventions used in the CX as Code Terraform provider codebase. It serves as a cornerstone for our shared understanding of the provider's architecture and implementation details.

## Purpose and Benefits

By establishing a common vocabulary and consistent naming patterns, we aim to:

1. Enhance code clarity and readability
2. Improve communication among team members
3. Reduce cognitive load when navigating the codebase
4. Facilitate easier onboarding for new contributors
5. Minimize errors stemming from misunderstandings or inconsistent terminology
6. Foster a cohesive mental model of the provider's structure and functionality

This shared language extends beyond just variable names to encompass function parameters, code comments, and overall architectural concepts. By adhering to these conventions, we create a more intuitive and maintainable codebase.

## Striving for Excellence

Our goal is to create software that matters – software that is not only functional but also elegantly crafted and easily understood. This dictionary is a step towards that goal, providing a foundation for clear communication and consistent implementation across our provider.

We encourage all contributors to familiarize themselves with these terms and use them consistently in code, comments, pull requests, and discussions. This shared understanding will help us build a more robust, efficient, and user-friendly CX as Code Terraform provider.

## Example Configurations

Our dictionary will be built using the examples of the following common Terraform code blocks:

### Resource Example

```hcl
resource "genesyscloud_location" "hq" {
name = "Headquarters"
address = "123 Main St, Anytown, USA"
}
```

### Data Source Example

```hcl
data "genesyscloud_user" "example" {
email = "john.doe@example.com"
name = "John Doe"
}
```

## Main Dictionary

### Shared High-Level Terms

These are terms that could generically apply either to a resource, data, or other Terraform block type.

- `blockDeclaration`: The entire block (resource or data)
- `blockType`: The type of block (e.g., "genesyscloud_location" or "genesyscloud_user")
- `blockLabel`: The label given to a specific block instance (e.g., "hq" in the resource example)
- `blockPath`: Combination of block type and label (e.g., "genesyscloud_location.hq")
- `blockDefinition`: The content within the curly braces of a block

### Resource-Specific Terms

These are terms that are specifically referring to `resource` block types (using the Resource Example from above).

- `resourceDeclaration`: The entire resource block (same as blockDeclaration for resources)
- `resourceType`: The type of resource (e.g., "genesyscloud_location")
- `resourceLabel`: The label given to a specific resource instance (same as blockLabel for resources)
- `resourcePath`: Combination of resource type and label (e.g., "genesyscloud_location.hq")
- `resourceDefinition`: The content within the curly braces of a resource (same as blockDefinition for resources)
- `resourceId`: The unique identifier (often a GUID) assigned to the resource (e.g., "1234abcd-56ef-78gh-90ij-klmno123pqrs")

### Data Source-Specific Terms

These are terms that are specifically referring to `data` block types (using the Data Source Example from above).

- `dataResourceDeclaration`: The entire data resource block
- `dataResourceType`: The type of data source (e.g., "genesyscloud_user")
- `dataResourceLabel`: The label given to a specific data resource instance (e.g., "example" in the data source example)
- `dataResourcePath`: Combination of data resource type and label (e.g., "data.genesyscloud_user.example")
- `dataResourceDefinition`: The content within the curly braces of a data resource

### Shared Lower-Level Terms

These are terms that are generic and not specific to `resource` or `data` or other Terraform block types.

#### Attribute Terms

- `attr`: The entire key-value pair of an attribute (e.g., `name = "Headquarters"`)
- `attrKey`: The identifier for an attribute (e.g., "name")
- `attrValue`: The value assigned to an attribute (e.g., "Headquarters")
- `attrRef`: A reference to a specific attribute of a resource or data source (e.g., "genesyscloud_location.hq.name" or "data.genesyscloud_user.example.email")

## Supplemental Information

When more specific references are needed, especially in cases involving multiple resources or data sources of the same type, or when distinguishing between resources and data sources, the following patterns can be used:

1. For multiple resources of the same type:

   - Pattern: `<resourceLabel><specificName>AttrRef`
   - Example: `hqAddressAttrRef` for `genesyscloud_location.hq.address`

2. For multiple data sources of the same type:

   - Pattern: `<dataResourceLabel><specificName>AttrRef`
   - Example: `johnEmailAttrRef` for `data.genesyscloud_user.john.email`

3. For resources of different types:

   - Pattern: `<resourceType><resourceLabel><specificName>AttrRef`
   - Examples:
     - `locationHqNameAttrRef` for `genesyscloud_location.hq.name`
     - `userJohnNameAttrRef` for `data.genesyscloud_location.john.name`

4. To explicitly distinguish between resources and data sources:
   - For resources: `resource<ResourceLabel><specificName>AttrRef`
   - For data sources: `data<DataResourceLabel><specificName>AttrRef`
   - Examples:
     - `resourceHqAddressAttrRef` for `genesyscloud_location.hq.address`
     - `dataExistingHqAddressAttrRef` for `data.genesyscloud_location.existing_hq.address`

These patterns provide flexibility for more specific references when needed, without cluttering the main dictionary.
