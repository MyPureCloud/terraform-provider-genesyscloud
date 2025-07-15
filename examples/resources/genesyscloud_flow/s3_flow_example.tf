# Example: Using S3 files with Genesys Cloud Architect Flow
# 
# This example demonstrates how to use S3-stored flow files with the
# genesyscloud_flow resource. The S3 bucket and key are automatically
# detected from the filepath.

# Example 1: Using S3 file with s3:// protocol
resource "genesyscloud_flow" "s3_flow_example" {
  filepath = "s3://my-bucket/flows/inboundcall_flow.yaml"
  file_content_hash = filesha256("s3://my-bucket/flows/inboundcall_flow.yaml")
  
  substitutions = {
    name = "My S3 Flow"
    type = "inboundcall"
  }
}

# Example 2: Using S3 file with s3a:// protocol (alternative)
resource "genesyscloud_flow" "s3a_flow_example" {
  filepath = "s3a://my-bucket/flows/inboundemail_flow.yaml"
  file_content_hash = filesha256("s3a://my-bucket/flows/inboundemail_flow.yaml")
  
  substitutions = {
    name = "My S3a Flow"
    type = "inboundemail"
  }
}

# Example 3: Mixed local and S3 files
resource "genesyscloud_flow" "local_flow_example" {
  filepath = "./local_flows/inboundcall_flow.yaml"
  file_content_hash = filesha256("./local_flows/inboundcall_flow.yaml")
  
  substitutions = {
    name = "My Local Flow"
    type = "inboundcall"
  }
}

resource "genesyscloud_flow" "s3_flow_example_2" {
  filepath = "s3://my-bucket/flows/inboundcall_flow_2.yaml"
  file_content_hash = filesha256("s3://my-bucket/flows/inboundcall_flow_2.yaml")
  
  substitutions = {
    name = "My S3 Flow 2"
    type = "inboundcall"
  }
}

# Example 4: Using force_unlock with S3 files
resource "genesyscloud_flow" "s3_flow_with_force_unlock" {
  filepath = "s3://my-bucket/flows/locked_flow.yaml"
  file_content_hash = filesha256("s3://my-bucket/flows/locked_flow.yaml")
  force_unlock = true
  
  substitutions = {
    name = "My Force Unlocked S3 Flow"
    type = "inboundcall"
  }
} 