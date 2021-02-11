![tests](https://github.com/MyPureCloud/terraform-provider-genesyscloud/workflows/Tests/badge.svg?branch=main)
# Genesys Cloud Terraform Provider
<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.15

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command: 
```sh
$ go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to the provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

When using the Terraform CLI, you can run [`terraform init`](https://www.terraform.io/docs/commands/init.html) in the directory containing your provider configuration and Terraform will automatically install the defined provider. The Genesys Cloud provider msut be configured with an authorized OAuth client ID and secret to call the SDK.

```hcl
terraform {
  required_version = "~> 0.13.0"
  required_providers {
    genesyscloud = {
      source  = "genesys/genesyscloud"
      version = "1.0.0"
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
GENESYSCLOUD_REGION
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

```sh
$ make testacc
```

*Note:* Acceptance tests create real resources and require an OAuth Client authorized to create, update, and delete all resources in your org. The OAuth Client information must be set in the `GENESYSCLOUD_*` environment variables prior to running the tests.