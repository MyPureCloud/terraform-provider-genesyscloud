# genesyscloud_flow

A Genesys Cloud Flow resource that allows you to manage architect flows using YAML files stored locally or in S3.

## Example Usage

### Local File

```hcl
resource "genesyscloud_flow" "example" {
  filepath = "./flows/inboundcall_flow.yaml"
  file_content_hash = filesha256("./flows/inboundcall_flow.yaml")
  
  substitutions = {
    name = "My Flow"
    type = "inboundcall"
  }
}
```

### S3 File

```hcl
resource "genesyscloud_flow" "s3_example" {
  filepath = "s3://my-bucket/flows/inboundcall_flow.yaml"
  file_content_hash = filesha256("s3://my-bucket/flows/inboundcall_flow.yaml")
  
  substitutions = {
    name = "My S3 Flow"
    type = "inboundcall"
  }
}
```

### Force Unlock

```hcl
resource "genesyscloud_flow" "force_unlock_example" {
  filepath = "s3://my-bucket/flows/locked_flow.yaml"
  file_content_hash = filesha256("s3://my-bucket/flows/locked_flow.yaml")
  force_unlock = true
  
  substitutions = {
    name = "My Force Unlocked Flow"
    type = "inboundcall"
  }
}
```

## Argument Reference

* `filepath` - (Required) Path to the YAML file containing the flow configuration. Supports:
  * Local file paths (e.g., `./flows/flow.yaml`)
  * S3 URIs (e.g., `s3://bucket-name/path/to/flow.yaml`)
  * S3a URIs (e.g., `s3a://bucket-name/path/to/flow.yaml`)
  * HTTP URLs (e.g., `http://example.com/flow.yaml`)

* `file_content_hash` - (Required) SHA256 hash of the file content. Used to detect changes.

* `substitutions` - (Optional) Key-value pairs for substituting values in the YAML file. Keys should be wrapped in `{{}}` in the YAML file.

* `force_unlock` - (Optional) Force unlock the flow before publishing. Useful when a flow is locked by another user.

## S3 Support

The resource supports reading flow files from Amazon S3. The following S3 URI formats are supported:

* `s3://bucket-name/path/to/file.yaml`
* `s3a://bucket-name/path/to/file.yaml`

### AWS Credentials

The resource uses the standard AWS credential chain to authenticate with S3:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. AWS credentials file (`~/.aws/credentials`)
3. IAM roles (EC2 instance profiles, EKS service accounts)
4. AWS SSO profiles

### S3 File Requirements

* The S3 object must be accessible with the configured AWS credentials
* The file must be a valid YAML flow configuration
* The file content should be consistent (avoid frequent updates to the same key)

## Import

Flows can be imported using their ID:

```bash
terraform import genesyscloud_flow.example flow-id
```

## Notes

* Changing the flow name will result in the creation of a new flow with a new GUID, while the original flow will persist in your org.
* The `force_unlock` option publishes the 'draft' architect flow and then publishes the flow named in this resource, mirroring the behavior found in the archy CLI tool.
* S3 files are downloaded and processed in memory, so very large files may impact performance.
* The resource automatically detects S3 paths and handles authentication using the AWS SDK v2 credential chain. 