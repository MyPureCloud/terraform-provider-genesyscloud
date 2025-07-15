# Genesys Cloud Export HTTP Server Specification

## Overview
Create a new HTTP server package that provides REST API endpoints for initiating and managing Genesys Cloud Terraform exports. This server will wrap the existing `tfexporter` functionality and provide asynchronous job management capabilities.

## Package Structure
```
genesyscloud/
├── export_server/
│   ├── server.go              # Main HTTP server implementation
│   ├── handlers.go            # HTTP request handlers
│   ├── job_manager.go         # Job state management
│   ├── auth.go                # Authentication logic
│   ├── models.go              # Request/response models
│   └── config.go              # Server configuration
```

## Core Components

### 1. Job Management System
- **Job Store**: In-memory or persistent storage for job state
- **Job States**: `pending`, `running`, `completed`, `failed`, `cancelled`
- **Job Metadata**: ID, status, created_at, updated_at, progress, error_message

### 2. Authentication
- **Bearer Token**: OAuth access token in Authorization header
- **Environment Variables**: Fallback to `GENESYSCLOUD_OAUTHCLIENT_ID`, `GENESYSCLOUD_OAUTHCLIENT_SECRET`, `GENESYSCLOUD_REGION`
- **401 Response**: When authentication fails
- **Note**: Do not attempt to generate enumss or case statements around the different API regionns.  If not auth token ispresent only use values in the environment variables

### 3. Export Parameters
All parameters from `genesyscloud_tf_export` resource:
- `directory` (string)
- `include_filter_resources` ([]string)
- `exclude_filter_resources` ([]string)
- `include_state_file` (bool)
- `export_format` (string: "hcl", "json", "json_hcl", "hcl_json")
- `split_files_by_resource` (bool)
- `log_permission_errors` (bool)
- `exclude_attributes` ([]string)
- `enable_dependency_resolution` (bool)
- `ignore_cyclic_deps` (bool)
- `compress` (bool)
- `export_computed` (bool)
- `use_legacy_architect_flow_exporter` (bool)

## API Endpoints

### 1. Start Export Job
```
POST /api/v1/exports
Content-Type: application/json
Authorization: Bearer <access_token>

Request Body:
{
  "directory": "./exports/job-123",
  "include_filter_resources": ["genesyscloud_user", "genesyscloud_group"],
  "export_format": "hcl",
  "include_state_file": true,
  "split_files_by_resource": false
}

Response (202 Accepted):
{
  "job_id": "job-123",
  "status": "pending",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### 2. Get Job Status
```
GET /api/v1/exports/{job_id}

Response (200 OK - Job Running):
{
  "job_id": "job-123",
  "status": "running",
  "progress": 45,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}

Response (204 No Content - Job Complete):
{
  "job_id": "job-123",
  "status": "completed",
  "progress": 100,
  "created_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:40:00Z"
}
```

### 3. Cancel Job
```
DELETE /api/v1/exports/{job_id}

Response (200 OK - Job Cancelled):
{
  "job_id": "job-123",
  "status": "cancelled",
  "cancelled_at": "2024-01-15T10:35:00Z"
}

Response (204 No Content - Job Already Complete):
{
  "job_id": "job-123",
  "status": "completed",
  "message": "Job already completed"
}
```

### 4. Get Export Results
```
GET /api/v1/exports/{job_id}/results

Response (200 OK - Results Available):
{
  "job_id": "job-123",
  "status": "completed",
  "files": [
    {
      "name": "genesyscloud.tf",
      "size": 1024,
      "type": "hcl"
    },
    {
      "name": "terraform.tfstate",
      "size": 2048,
      "type": "state"
    }
  ],
  "export_directory": "./exports/job-123"
}

Response (204 No Content - Job Still Running):
{
  "job_id": "job-123",
  "status": "running",
  "message": "Export still in progress"
}
```

### 5. Download Export Files
```
GET /api/v1/exports/{job_id}/download

Response (200 OK):
Content-Type: application/zip
Content-Disposition: attachment; filename="export-job-123.zip"
[ZIP file content]

Response (404 Not Found):
{
  "error": "Export files not found or job not completed"
}
```

## Implementation Details

### Job Manager Interface
```go
type JobManager interface {
    CreateJob(params ExportParams) (string, error)
    GetJobStatus(jobID string) (*JobStatus, error)
    CancelJob(jobID string) error
    GetJobResults(jobID string) (*JobResults, error)
    GetJobFiles(jobID string) ([]byte, error)
}
```

### Export Worker
```go
type ExportWorker struct {
    jobManager JobManager
    tfExporter *tfexporter.GenesysCloudResourceExporter
}

func (w *ExportWorker) ProcessJob(jobID string, params ExportParams) {
    // 1. Update job status to "running"
    // 2. Execute tfexporter.Export()
    // 3. Update job status to "completed" or "failed"
    // 4. Store results in job-specific directory
}
```

### File Organization
```
exports/
├── job-123/
│   ├── genesyscloud.tf
│   ├── terraform.tfstate
│   └── export.zip
├── job-456/
│   ├── genesyscloud.tf.json
│   └── terraform.tfstate
└── job-789/
    └── [cancelled job - no files]
```

## Error Handling
- **400 Bad Request**: Invalid export parameters
- **401 Unauthorized**: Missing or invalid authentication
- **404 Not Found**: Job ID not found
- **409 Conflict**: Job already in progress
- **500 Internal Server Error**: Export process failure

## Configuration
```go
type ServerConfig struct {
    Port            int    `default:"8080"`
    ExportBaseDir   string `default:"./exports"`
    MaxConcurrentJobs int `default:"5"`
    JobTimeout      time.Duration `default:"30m"`
}
```

## Security Considerations
- Validate all input parameters
- Sanitize file paths to prevent directory traversal
- Implement rate limiting
- Log all operations for audit trail
- Clean up old job files periodically

## Testing
-  Please write unit tests for all code generated
-  Please write integration tests for all code generated

## Additional Notes

