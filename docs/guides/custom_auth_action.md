---
subcategory: ""
page_title: "Configuring Integration Custom Auth Action"
description: |-
    A guide to configuring the automatically generated custom auth action for Web Services Data Actions integrations.
---

# Configuring Genesys Cloud integration custom auth action

Web Services Data Actions is a type of integration which allows a Genesys Cloud org to interact with third-party web services. Most of the time, invoking these third-party API endpoints requires an authentication mechanism. Genesys Cloud provides three options for configuring credentials:

1. Basic Auth
2. User Defined
3. User Defined (OAuth)

To know more about these credential types go this [article](https://help.mypurecloud.com/articles/credential-types-web-services-data-actions-integration/)

When `User Defined (OAuth)` is selected, Genesys Cloud automatically creates a new Data Action with the `Custom Auth` type. The existence of this Data Action is completely managed by Genesys Cloud and it cannot be manually created nor deleted. The Custom Auth Data Action is also automatically published upon creation so its input and output contracts cannot be modified. This leaves only the Action Configuration (Request and Response) open for modification to meet the needs of the third-party authentication mechanism.

Because of its distinct differences from regular Data Actions, a special resource has been defined to manage the Custom Auth Data Action and its configuration: `genesyscloud_integration_custom_auth_action`.

In this guide we will see examples on how to manage the resource and its behavior on different cases.

## Requirements

The existence of the Custom Auth Data Action is completely managed by Genesys Cloud. Before configuring it with CX as Code, you would need:

1. A Genesys Cloud Web Services Data Actions integration.
2. The integration should be configured with `User Defined (OAuth)` credentials type.

It's best practice to manage the integration, credentials, and custom auth action all on CX as Code but it's also possible to manage only the `genesyscloud_integration_custom_auth_action` on an existing integration. The requirements still have to be met for it to work though.

## Managing the integration and its custom auth action in CX as Code

The following snippet is an example of how you can manage your integration infrastructure along with the credentials type and custom auth data action:

```hcl
resource "genesyscloud_integration_credential" "credential" {
  name                 = "example-credential"
  credential_type_name = "userDefinedOAuth"
  fields = {
    clientId = var.client_id
    clientSecret = var.client_secret
  }
}

resource "genesyscloud_integration" "example_integration" {
  intended_state   = "ENABLED"
  integration_type = "custom-rest-actions"
  config {
    name = "Example Integration Name"
    credentials = {
      basicAuth = genesyscloud_integration_credential.credential.id
    }
  }
}

resource "genesyscloud_integration_custom_auth_action" "auth_action" {
  integration_id = genesyscloud_integration.example_integration.id
  config_request {
  request_url_template = "https://example-domain.com/loginurl"
    request_type         = "POST"
    request_template     = "grant_type=client_credentials"
    headers = {
      Authorization = "Basic $encoding.base64(\"$${credentials.clientId}:$${credentials.clientSecret}\")"
    }
  }
  config_response {
    success_template = "$${rawResult}"
  }
}
```

Take note of the configurations of both `genesyscloud_integration_credential` and the `genesyscloud_integration`: the credential should be of type `userDefinedOAuth`, and the integration should be of type `custom-rest-actions`. Having values other than these, on any of the two resources, would result in failure as Genesys Cloud will not create the Custom Auth Action and CX as Code wouldn't have any `genesyscloud_integration_custom_auth_action` resource to manage.

**BEST PRACTICE**: When configuring credentials and other sensitive information in CX as Code, it is recommended to use [input variables](https://developer.hashicorp.com/terraform/language/values/variables) rather than hardcoding them into the configuration itself. This allows you to secure the variable definitions separately from the infrastructure files.

## 'Deleting' the resource from CX as Code

As mentioned, it is not possible to delete the Custom Auth Data Action from CX as Code as its existence is managed by Genesys Cloud. If you want to delete the Custom Auth Action, then you must replace the credential type of the integration or delete the integration itself.

## Example cases

Here are some other cases that you may run into while configuring the resource in CX as Code and their effects:

* Having an integration and credentials of correct type but not a  `genesyscloud_integration_custom_auth_action` resource.
  * If requirements are met on the integration and its credential type, Genesys Cloud will still create the Custom Auth Data Action. It would just retain default values and won't be managed by CX as Code. This is a completely normal use case. You can choose to manage it anytime by adding the custom auth resource.
* Initially managing the integration and custom auth resources, then deleting only the custom auth resource from CX as Code.
  * The custom auth resource will still exist in Genesys Cloud but will not be managed by CX as Code. It would keep the latest configuration it was set on. The action will only be deleted if the integration is deleted or its credential type is changed.

* Initially managing the integration and custom auth resources, and only deleting the integration resource.
  * Deleting the integration or changing its configuration to not meet the requirements will make Genesys Cloud delete the custom auth action. This will cause the custom auth resource (if it is still defined in CX as Code) to fail.

* Managing the custom auth resource and changing the `integration_id` property in an update.
  * The previous integration it was associated with would keep the latest auth action configuration. The custom auth action itself won't be deleted or transferred to the new integration. Changing the `integration_id` of the resource just means managing and applying the configuration to a different custom auth action (assuming requirements are met).

## Exporting custom auth actions

Genesys Cloud creates the same initial configuration for Custom Auth Data Actions. This means that if the custom auth resource is never managed by CX as Code or it was not modified from its default configuration, then it would be excluded from export. If the custom auth action is modified however (manually or through CX as Code), it would be included in the export as long as the `genesyscloud_integration_custom_auth_action` is defined in the export configuration.

For more details on exporting Genesys Cloud Resources, refer to this [guide](https://registry.terraform.io/providers/MyPureCloud/genesyscloud/latest/docs/guides/export).
