---
subcategory: ""
page_title: "Export Genesys Cloud Configuration"
description: |-
    A guide to exporting existing Genesys Cloud configuration.
---

# Exporting existing Genesys Cloud configuration

For existing orgs, it may be desirable to have Terraform begin managing your configuration. If there are only a handful of users, queues, etc. to manage, Terraform has an option to [import individual resources](https://www.terraform.io/docs/cli/import/index.html) into a Terraform module.

However, if there are a lot of existing resources it can be painful to manually import all of them into a Terraform state file. To make this easier, a special resource has been defined to export that configuration into a local [Terraform JSON configuration file](https://www.terraform.io/docs/language/syntax/json.html). Ensure you have the Terraform CLI installed and create a `.tf` file that requires the genesyscloud provider as shown below. Add a `genesyscloud_tf_export` resource to that same file:
```hcl
terraform {
  required_providers {
    genesyscloud = {
      source  = "mypurecloud/genesyscloud"
      version = "~> 1.0.0"
    }
  }
}

resource "genesyscloud_tf_export" "export" {
  directory          = "~/genesyscloud"
  resource_types     = ["genesyscloud_user"]
  include_state_file = true
}
```

You may choose specific resource types to export such as `genesyscloud_user`, or you can export all supported resources by not setting the `resource_types` attribute. You may also choose to export a `.tfstate` file along with the `.tf.json` config file by setting `include_state_file` to true. Generating a state file alongside the config will allow Terraform to begin managing your existing resources even though it did not create them. Excluding the state file will generate configuration that can be applied to a different org.

Once your export resource is configured, run `terraform init` to set up Terraform in that directory followed by `terraform apply` to run the export. Once complete, a new Terraform config file will be created in the chosen directory where you can begin modifying the generated config and running Terraform commands.

If state is exported, the config file may not be able to be applied to another org as it likely contains ID references to objects in the current org. If you choose not to export the state file, the standalone `.tf.json` config file will be stripped of all reference attribute values that cannot be mapped to exported resources. For example if you only export users, any attributes that reference other object types (roles, skills, etc.) will be removed from the config. This is necessary as it would not be possible to apply configuration with references to IDs from a different org.