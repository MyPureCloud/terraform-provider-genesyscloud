---
subcategory: ""
page_title: "Export Genesys Cloud Configuration"
description: |-
    A guide to exporting existing Genesys Cloud configuration.
---

# Exporting Existing Genesys Cloud Configuration

For existing orgs, it may be desirable to have Terraform begin managing your configuration. If there are only a handful of users, queues, etc. to manage, Terraform has an option to [import individual resources](https://www.terraform.io/docs/cli/import/index.html) into a Terraform module.

However, if there are a lot of existing resources it can be painful to manually import all of them into a Terraform state file. To make this easier, a special resource has been defined to export that configuration into a local Terraform [JSON configuration file](https://www.terraform.io/docs/language/syntax/json.html) or [TF configuration file](https://www.terraform.io/language/syntax/configuration). Ensure you have the Terraform CLI installed and create a `.tf` file that requires the genesyscloud provider (click the Use Provider drop down to learn how to use the latest version of the required provider). Add a `provider` block and a `genesyscloud_tf_export` resource to that same file:

```hcl
provider "genesyscloud" {
}

resource "genesyscloud_tf_export" "export" {
  directory          = "./genesyscloud"
  resource_types     = ["genesyscloud_user"]
  include_state_file = true
}
```

The configuration can be exported as a `.tf` file by setting `export_as_hcl` to `true`.

You may choose specific resource types to export such as `genesyscloud_user`, or you can export all supported resources by not setting the `resource_types` attribute. You may also choose to export a `.tfstate` file along with the `.tf.json` or `.tf` config file by setting `include_state_file` to true. Generating a state file alongside the config will allow Terraform to begin managing your existing resources even though it did not create them. Excluding the state file will generate configuration that can be applied to a different org.

Once your export resource is configured, run `terraform init` to set up Terraform in that directory followed by `terraform apply` to run the export. Once complete, a new Terraform config file will be created in the chosen directory where you can begin modifying the generated config and running Terraform commands.

If state is exported, the config file may not be able to be applied to another org as it likely contains ID references to objects in the current org. If you choose not to export the state file, the standalone `.tf.json` or `.tf` config file will be stripped of all reference attribute values that cannot be mapped to exported resources. For example if you only export users, any attributes that reference other object types (roles, skills, etc.) will be removed from the config. This is necessary as it would not be possible to apply configuration with references to IDs from a different org.

If exported resources contain references to objects that we don't intend to manage with Terraform or if they cannot be resolved using an API call then a variable will be generated to refer to that object. A definition for that variable will be provided in a generated `terraform.tfvars` file. The reference variables must be filled out with the values of the corresponding resources in a different org before being applied to it.

## Filtering Resources with Regular Expressions

In your Terraform setup, regular expressions can be employed to selectively include or exclude certain resources. Here are various ways to achieve this:

### Include Filter

If you want to include resources that begin or end with “dev” or “test”, use the following format:

```hcl
resource "genesyscloud_tf_export" "include-filter" {
  directory = "./genesyscloud/include-filter"
  export_as_hcl = true
  log_permission_errors = true
  include_filter_resources = ["genesyscloud_group::.*(?:dev|test)$"]
}
```

### Exclude Filter

To exclude certain resources, you can use a similar method:

```hcl
resource "genesyscloud_tf_export" "exclude-filter" {
  directory = "./genesyscloud/exclude-filter"
  export_as_hcl = true
  log_permission_errors = true
  exclude_filter_resources = ["genesyscloud_routing_queue"]
}
```

### Advanced Filter

For more complex filtering scenarios, you can use the `advanced_filter_resources` attribute. This attribute combines the functionality of both include and exclude filters, offering more granular control:

```hcl
resource "genesyscloud_tf_export" "advanced-filter" {
  directory = "./genesyscloud/advanced-filter"
  export_as_hcl = true
  log_permission_errors = true
  advanced_filter_resources {
    include_by_type = ["genesyscloud_routing_queue", "genesyscloud_script"]
    include_by_name = ["genesyscloud_user::Foo", "genesyscloud_location::HQ"]
    exclude_by_name = ["genesyscloud_script::Default.*"]
  }
}
```

In this advanced filter:

- All resources for `genesyscloud_routing_queue` and `genesyscloud_script` will be returned.
- Specific resources for `genesyscloud_user` with a name of `Foo` and `genesyscloud_location` with a name of `HQ` will be returned.
- Any genesyscloud_script resources with names beginning with `Default` will be excluded.

You can also use the `exclude_by_type` attribute in the advanced filter:

```hcl
resource "genesyscloud_tf_export" "advanced-filter" {
  directory = "./genesyscloud/advanced-filter"
  export_as_hcl = true
  log_permission_errors = true
  advanced_filter_resources {
    exclude_by_type = ["genesyscloud_user"]
    include_by_name = ["genesyscloud_user::Foo", "genesyscloud_location::HQ"]
    exclude_by_name = ["genesyscloud_script::Default.*"]
  }
}
```

In this example:

- All resources except for `genesyscloud_user` will be returned.
- `genesyscloud_user` resources with the name `Foo` will be included despite the type exclusion.
- `genesyscloud_script` resources with names beginning with `Default` will not be exported.

The `advanced_filter_resources` attribute provides a powerful way to fine-tune your resource filtering, allowing for more complex inclusion and exclusion patterns.

## Replacing an Exported Resource with a Data Source

In the course of managing your Terraform configuration, circumstances may arise where it becomes desirable to substitute an exported resource with a data source. The following are instances where such an action might be warranted:

```hcl
resource "genesyscloud_tf_export" "export" {
  directory = "./genesyscloud/datasource"
  replace_with_datasource = [
    "genesyscloud_group::Test_Group"
  ]
  include_state_file     = true
  export_as_hcl          = true
  log_permission_errors  = true
  enable_dependency_resolution = false
}
```

## Enable Dependency Resolution

In its standard setup, this Terraform configuration exports only the dependencies explicitly defined in your configuration. However, by enabling `enable_dependency_resolution`, Terraform can automatically export additional dependencies, including static ones associated with an architecture flow. This feature enhances the comprehensiveness of your exports, ensuring that not just the primary resource, but also its related entities, are included.

On the other hand, Terraform also provides the `exclude_attributes` option for instances where certain fields need to be omitted from an export. This, along with the ability to automatically export additional dependencies, contributes to Terraform’s flexible framework for managing resource exports. It allows for granular control over the inclusion or exclusion of elements in the export, ensuring that your exported configuration aligns precisely with your requirements.

## Export State File Comparison

In its standard setup, during a full org download, the exporter doesn't verify if the exported state file is in sync with the exported configuration.
This is an experimental feature enabled just for troubleshooting. To enable this,set env value of ENABLE_EXPORTER_STATE_COMPARISON to true.
