# S3 File Upload Enhancement for Architect Flow Resource

## Context
You are a senior Genesys Cloud Go developer working on the Terraform CX as Code provider. The main source code is in the `genesyscloud/` directory.

## Current State Analysis
The `architect_flow` resource currently only supports local filesystem file uploads. The file handling is primarily done in:
- `genesyscloud/architect_flow/resource_genesyscloud_flow.go` (lines 120-140)
- `genesyscloud/util/files/util_files.go` (S3Uploader struct and related functions)

## Requirements

### 1. **S3 Integration Enhancement**
- Add S3 bucket support to the existing local filesystem functionality
- Maintain backward compatibility with local file paths
- Support both local and S3 file sources simultaneously

### 2. **S3 Bucket Detection**
- Automatically detect S3 bucket from file path/name
- Support standard S3 URI formats: `s3://bucket-name/path/to/file.yaml`
- Support alternative formats: `s3a://bucket-name/path/to/file.yaml`

### 3. **AWS Credential Resolution**
- Implement standard AWS credential chain resolution:
  - Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
  - AWS credentials file (~/.aws/credentials)
  - IAM roles (EC2 instance profiles, EKS service accounts)
  - AWS SSO profiles
- Use AWS SDK v2 for Go for credential management

### 4. **Reusable Utility Functions**
- Create new utility functions in `genesyscloud/util/files/` package
- Functions should be generic and reusable across other resources
- Include proper error handling and logging
- Add unit tests for all new functions

### 5. **Implementation Details**

#### New Utility Functions to Create:
```go
// In genesyscloud/util/files/s3_utils.go
func IsS3Path(path string) bool
func ParseS3URI(uri string) (bucket, key string, err error)
func DownloadS3File(bucket, key string) (io.Reader, error)
func UploadS3File(bucket, key string, reader io.Reader) error
func GetS3FileReader(path string) (io.Reader, *os.File, error)
```

#### Modified Functions:
- Update `DownloadOrOpenFile()` in `util_files.go` to handle S3 paths
- Modify `NewS3Uploader()` to support S3 file sources
- Update architect flow resource to use new S3 utilities

### 6. **Testing Requirements**
- Unit tests for all new S3 utility functions
- Integration tests for S3 file upload/download scenarios
- Mock AWS credentials for testing
- Test both local and S3 file scenarios

### 7. **Dependencies**
- Add AWS SDK v2 for Go: `github.com/aws/aws-sdk-go-v2`
- Add AWS S3 service: `github.com/aws/aws-sdk-go-v2/service/s3`
- Add AWS config: `github.com/aws/aws-sdk-go-v2/config`

## Implementation Steps

1. **Examine existing code** in `architect_flow` and `util/files` packages
2. **Create S3 utility functions** in new `s3_utils.go` file
3. **Modify existing file handling** to support S3 paths
4. **Update architect flow resource** to use new utilities
5. **Add comprehensive tests** for all new functionality
6. **Update documentation** for new S3 capabilities

## Success Criteria
- Architect flow resource can upload files from both local filesystem and S3
- S3 bucket is automatically detected from file path
- AWS credentials are resolved using standard chain
- All existing functionality remains unchanged
- Comprehensive test coverage for new features
- Clean, reusable utility functions in the utils package

## Files to Focus On
- `genesyscloud/architect_flow/resource_genesyscloud_flow.go`
- `genesyscloud/util/files/util_files.go`
- `genesyscloud/util/files/s3_utils.go` (new file)
- `genesyscloud/util/files/util_files_test.go`
- `genesyscloud/architect_flow/resource_genesyscloud_flow_test.go`