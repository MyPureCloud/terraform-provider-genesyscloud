![tests](https://github.com/MyPureCloud/terraform-provider-genesyscloud/workflows/Tests/badge.svg?branch=main)

# Genesys Cloud Terraform Provider

<img src="https://upload.wikimedia.org/wikipedia/commons/0/04/Terraform_Logo.svg" width="600px" alt="Terraform Logo">

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0.x
- [Go](https://golang.org/doc/install) >= 1.18

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

_Note:_ If `GENESYSCLOUD_ACCESS_TOKEN` is set, the Oauth client will use the access token instead of client credentials to make requests.

_Note:_ The provider makes Public API calls to perform all of the CRUD operations necessary to manage Genesys Cloud resources. All of these API calls require specific permissions and OAuth scopes. Therefore it is important that you verify your OAuth Client is authorized for all necessary scopes and is assigned an admin role capable of creating, reading, updating, and deleting all resources that your Terraform configuration will manage.

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

**Please branch off of the latest version of dev and be sure to set dev as the target branch of your pull request.**

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

In order to run the full suite of Acceptance tests, run `make testacc`. You can also specify individual tests using the TESTARGS variable:

```sh
$ make testacc TESTARGS="-run TestAccResourceUserBasic"
```

All new resources must have passing acceptance tests and docs in order to be merged. Most of the docs are generated automatically from the schema and examples folder by running `make docs`.

To run all of the unit tests:

```sh
$make testunit
```

### Adding a new resource type

1. Create new package inside `genesyscloud` with the following files. The package name should match the name of the resource (minus, the genesyscloud\_ prefix).
   - `resource_genesyscloud_{resource_name}_schema.go` - The file containing the schema definition for the resource, data source, and exporter. It also contains a constant variable which defines the resource name, and a public function called `SetRegistrar`, which will be called from outside the package to register the resource with the provider. The schema should closely match Public API schemas, but there are some Terraform schema limitations that may require some deviation from the API.
   - `resource_genesyscloud_{resource_name}.go` - Contains the create, read, update and delete functions for the resource (the methods for the resource's `CreateContext`, `ReadContext`, `UpdateContext`, and `DeleteContext` attributes.) It also contains the getAll function to be used by the exporter. If you have a few helper functions present in this file, that is fine, but if there are more than 1-2 helper functions you should create a `resource_genesyscloud_{resource_name}_utils.go` to contain the business logic.
   - `resource_genesyscloud_{resource_name}_test.go` - Contains the resource tests. Write acceptance test cases that cover all of the attributes and CRUD operations for the resource. Acceptance tests modify real resources in a test org and require an OAuth Client authorized to create, update, and delete the resource type in the org. See existing tests for examples and [Terraform Acceptance Test documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html) for more details. Unit tests are also encouraged where applicable (for helper functions etc.)
   - `data_source_genesyscloud_{resource_name}.go` - This file contains all of the data source logic for a resource. The data source should call any Genesys Cloud APIs through its API proxy class. All functions and variables in this class should be private.
   - `data_source_genesyscloud_{resource_name}_test.go` - Contains the data source tests.
   - `genesyscloud_{resource_name}_proxy.go` - This contains all of the API logic for interacting with the Genesys Cloud APIs. This is meant to be an isolated layer from Terraform, so know Terraform objects should be passed back and forth to this code. All functions and variables in this class should be private.
   - `genesyscloud_{resource_name}_init_test.go` - This file contains all of the logic needed to initialize a test case for your resource. All functions and variables in this class should be private.
2. Add a new folder for the resource and data source under the `/examples` folder. An example `resource.tf` file for the resource should be added to the folder along with an `apis.md` file listing all of the APIs the resource uses. To generate the documentation, run `go generate`. **Note:** Everything inside the `docs` directory is generated based off schema data and the content inside `examples`. Do not manually edit anything inside `docs`.
3. Import your package to `main.go` at the root of the project and, from the `registerResources` function, call the SetRegistrar function passing in the `regInstance` variable.

If you want to go off of an example, we recommend using the [external contacts](https://github.com/MyPureCloud/terraform-provider-genesyscloud/tree/main/genesyscloud/external_contacts) package.

### Cx As Code Resource Generator

[The Cx as Code Resource Generator](https://github.com/MyPureCloud/cxascode-resource-generator) is a tool that can help generate resources for Cx as Code and speed up development. The resource generator will generate resources using the package structure mentioned above. The project can be found [here](https://github.com/MyPureCloud/cxascode-resource-generator) and all usage is documented in the README. Please note that the resource generator is not perfect, it is a tool to help with development and the generated code will require review and the package will still need to be registered manually in `main.go`.

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
