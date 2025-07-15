# Script Examples

This directory contains examples for the `genesyscloud_script` resource, demonstrating various usage patterns including local filesystem and S3 file support.

## Examples

### Basic Script with Local File
- **File**: `resource.tf` (example_script)
- **Description**: Standard script using local JSON file
- **Use Case**: Scripts stored in your local project directory

### Script with S3 File
- **File**: `resource.tf` (s3_script)
- **Description**: Script using JSON file stored in Amazon S3
- **Use Case**: Centralized script management in S3
- **Requirements**: AWS credentials configured

### Script with Division Assignment
- **File**: `resource.tf` (division_script)
- **Description**: Script with specific division assignment and S3 file
- **Use Case**: Multi-division organizations with centralized script storage

### Local Script Example
- **File**: `resource.tf` (local_script)
- **Description**: Script using local file with relative path
- **Use Case**: Scripts stored in project subdirectories

## File Support

The `genesyscloud_script` resource supports:

### Local Files
- Standard paths: `/path/to/script.json`
- Relative paths: `./scripts/email.script.json`

### S3 Files
- S3 URI format: `s3://bucket-name/path/to/script.json`
- Alternative format: `s3a://bucket-name/path/to/script.json`

### AWS Credentials
S3 files use the standard AWS credential chain:
- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- AWS credentials file (`~/.aws/credentials`)
- IAM roles (EC2 instance profiles, EKS service accounts)
- AWS SSO profiles

## Usage Notes

1. **File Content Hash**: Always use `filesha256()` function for the `file_content_hash` field to detect changes
2. **Script Format**: Scripts must be in Genesys Cloud's JSON format
3. **Substitutions**: Use the `substitutions` map to replace variables in your script files
4. **S3 Permissions**: Ensure your AWS credentials have read access to the S3 buckets and objects
5. **Division ID**: Optional field to assign scripts to specific divisions

## Test Data

The `email.script.json` file contains a sample Genesys Cloud script that can be used for testing purposes. 