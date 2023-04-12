![tests](https://github.com/MyPureCloud/terraform-provider-genesyscloud/workflows/Tests/badge.svg?branch=main)
# Genesys Cloud Terraform Provider
<img src="https://upload.wikimedia.org/wikipedia/commons/0/04/Terraform_Logo.svg" width="600px" alt="Terraform Logo">

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 1.0.x
-	[Go](https://golang.org/doc/install) >= 1.18

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider by running `make build`

## Using the provider

When using the Terraform CLI, you can run [`terraform init`](https://www.terraform.io/docs/commands/init.html) in the directory containing your provider configuration and Terraform will automatically install the required provider. The Genesys Cloud provider must be configured with an authorized OAuth client ID and secret to call the SDK.

```hcl
terraform {
  required_version = ">= 1.0.0"
  required_providers {
    genesyscloud = {
      source  = "mypurecloud/genesyscloud",
      version = ">= 1.6.0"
    }
  }
}

provider "genesyscloud" {
  oauthclient_id = "<client-id>"
  oauthclient_secret = "<client-secret>"
  aws_region = "<aws-region>"
}

```

The following environment variables may be set to avoid hardcoding OAuth Client information into your Terraform files:

```
GENESYSCLOUD_OAUTHCLIENT_ID
GENESYSCLOUD_OAUTHCLIENT_SECRET
GENESYSCLOUD_ACCESS_TOKEN
GENESYSCLOUD_REGION
```

*Note:* If `GENESYSCLOUD_ACCESS_TOKEN` is set, the Oauth client will use the access token instead of client credentials to make requests.

*Note:* The provider makes Public API calls to perform all of the CRUD operations necessary to manage Genesys Cloud resources. All of these API calls require specific permissions and OAuth scopes. Therefore it is important that you verify your OAuth Client is authorized for all necessary scopes and is assigned an admin role capable of creating, reading, updating, and deleting all resources that your Terraform configuration will manage.

For any issues, questions, or suggestions for the provider, visit the [Genesys Cloud Developer Forum](https://developer.mypurecloud.com/forum/)

### Proxy Configuration

Use of a proxy is accomplished by setting the proxy settings for the provider

The `Proxy` has 3 properties that determine the URL for proxying.

port - Port of the Proxy server
host - Host Ip or DNS of the proxy server
protocol - Protocol required to connect to the Proxy (http or https)

The 'proxy' has another section which is an optional section. 
If the proxy requires authentication to connect to
'auth' needs to be mentioned under the 'Proxy'.

An example of the provider configuration with the proxy: 

```hcl
provider "genesyscloud" {
  oauthclient_id = "<client-id>"
  oauthclient_secret = "<client-secret>"
  aws_region = "<aws-region>"

  proxy {
    host     = "example.com"
    port     = "8443"
    protocol = "https"

    auth {
      username = "john"
      password = "doe"
    }
  }
}
```

The following environment variables may be set to avoid hardcoding Proxy and Auth Client information into your Terraform files:

```
GENESYSCLOUD_PROXY_PORT
GENESYSCLOUD_PROXY_HOST
GENESYSCLOUD_PROXY_PROTOCOL
GENESYSCLOUD_PROXY_AUTH_USERNAME
GENESYSCLOUD_PROXY_AUTH_PASSWORD

```

### Data Sources

There may be cases where you want to reference existing resources in a Terraform configuration file but do not want those resources to be managed by Terraform. This provider supports several data source types that can act as a read-only resource for existing objects in your org. To include one in your configuration, add a `data` block to your configuration file with one of the supported data source types:
```hcl
data "genesyscloud_auth_role" "employee" {
  name = "employee"
}
```
The example above will attempt to find a role named "employee" which can be referenced elsewhere in the config. By default, all data sources will allow you to access the `id` attribute which is useful for setting reference attributes that require IDs. Additional attributes may be added to data sources as needs arise.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

In order to run the full suite of Acceptance tests, run `make testacc`. You can also specify individual tests using the TESTARGS variable:

```sh
$ make testacc TESTARGS="-run TestAccResourceUserBasic"
```

All new resources must have passing acceptance tests and docs in order to be merged. Most of the docs are generated automatically from the schema and examples folder by running `go generate`.

### Adding a new resource type

1. Create new resource and test go files for the resource type, e.g. `resource_genesyscloud_{resource_name}.go` and `resource_genesyscloud_{resource_name}_test.go`. Resource names should typically be the same as (or very similar to) the Public API resource. 
2. Define your resource schema in a method returning a `*schema.Resource`. See existing schemas and [this page](https://www.terraform.io/docs/extend/schemas/index.html) for examples. The schema should closely match Public API schemas, but there are some Terraform schema limitations that may require some deviation from the API.
3. Add the resource name along with the schema method to the `ResourcesMap` found in `provider.go`. This will make the resource available to the plugin.
4. Define methods for the resource's `CreateContext`, `ReadContext`, `UpdateContext`, and `DeleteContext` attributes as necessary. As the names imply, each one should handle one of the CRUD operations for the resource. Some best practices can be found [here](https://www.terraform.io/docs/extend/best-practices/index.html), and existing resources contain many common patterns and examples.
5. If the resource should be exportable, add a method that returns a `*ResourceExporter` for the resource. See `resource_exporter.go` for details on each field in the `ResourceExporter` struct. This method should be added the `getResourceExporters` method in `resource_exporter.go` to make it an exportable resource.
6. Write acceptance test cases that cover all of the attributes and CRUD operations for the resource. The tests should be written in the `resource_genesyscloud_{resource_name}_test.go` file. Acceptance tests modify real resources in a test org and require an OAuth Client authorized to create, update, and delete the resource type in the org. See existing tests for examples and [Terraform Acceptance Test documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html) for more details.
7. Add a new folder for the resource under the `/examples` folder. An example `resource.tf` file for the resource should be added to the folder along with an `apis.md` file listing all of the APIs the resource uses. To generate or update documentation, run `go generate`.

### Using the Provider locally

In order to use a locally compiled version of the provider, the correct binary for your system must be copied to the local `~/.terraform.d/plugins` folder. Run `make sideload` to build the provider and copy it to the correct folder. In your Terraform config file, specify version `0.1.0` and set the provider source to `genesys.com/mypurecloud/genesyscloud`. Run `terraform init` and verify that it finds the local version.

### Branches

Branch names should begin with `feat/` for new features or `bug/` for bug fixes. This ensures that the PR for this branch is correctly labeled and added to the changelog in the next release.

### Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to the provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

### Releases

A GitHub release action will be triggered when pushing version tags starting with 'v'. The release number must follow the [Semantic Versioning](https://semver.org/spec/v2.0.0.html) spec.

```
$ git tag -a v1.1.1 -m "Release v1.1.1"
$ git push origin v1.1.1
```

This action will build binaries, generate a changelog from labeled PRs, and create a draft GitHub release. The GitHub release draft will be reviewed and released manually.
