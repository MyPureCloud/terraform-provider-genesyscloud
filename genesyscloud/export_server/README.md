# Genesys Cloud Export HTTP Server

This package provides a REST API server for initiating and managing Genesys Cloud Terraform exports asynchronously. The server wraps the existing `tfexporter` functionality and provides job management capabilities.

## Features

- **Asynchronous Export Jobs**: Start export jobs and track their progress
- **Job Management**: Create, monitor, cancel, and retrieve export jobs
- **File Downloads**: Download completed exports as ZIP files
- **Authentication**: Support for OAuth tokens and environment variables
- **Automatic Cleanup**: Clean up old job files and data
- **Graceful Shutdown**: Proper cleanup on server shutdown

## API Endpoints

### 1. Create Export Job
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

## Authentication

The server supports two authentication methods:

1. **Bearer Token**: Include `Authorization: Bearer <access_token>` header
2. **Environment Variables**: Set the following environment variables:
   - `GENESYSCLOUD_OAUTHCLIENT_ID`
   - `GENESYSCLOUD_OAUTHCLIENT_SECRET`
   - `GENESYSCLOUD_REGION`

## Configuration

The server can be configured using environment variables:

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `EXPORT_SERVER_PORT` | 8080 | HTTP server port |
| `EXPORT_SERVER_BASE_DIR` | ./exports | Base directory for export files |
| `EXPORT_SERVER_MAX_JOBS` | 5 | Maximum concurrent jobs |
| `EXPORT_SERVER_JOB_TIMEOUT` | 30m | Job timeout duration |
| `EXPORT_SERVER_CLEANUP_INTERVAL` | 1h | Cleanup interval |
| `EXPORT_SERVER_MAX_JOB_AGE` | 24h | Maximum age of job files |

## Usage

### Starting the Server

```go
package main

import (
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/export_server"
)

func main() {
    // Start server with default configuration
    export_server.StartServer()
}
```

### Using the Server Programmatically

```go
package main

import (
    "log"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/export_server"
)

func main() {
    // Create custom configuration
    config := export_server.DefaultConfig()
    config.Port = 9090
    config.ExportBaseDir = "/tmp/exports"

    // Create and start server
    server := export_server.NewServer(config)
    
    log.Printf("Starting server on port %d", config.Port)
    if err := server.Start(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

## Job States

- **pending**: Job created, waiting to start
- **running**: Job is currently executing
- **completed**: Job finished successfully
- **failed**: Job failed with error
- **cancelled**: Job was cancelled by user

## File Organization

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

## Security Considerations

- All input parameters are validated
- File paths are sanitized to prevent directory traversal
- Authentication is required for all endpoints
- Job files are isolated by job ID
- Old job files are automatically cleaned up

## Testing

The server includes comprehensive unit tests for all components:

- Job management functionality
- HTTP request handling
- Authentication logic
- Export worker integration
- File operations

Run tests with:
```bash
go test ./genesyscloud/export_server/...
```

## Dependencies

- `github.com/google/uuid` - Job ID generation
- `github.com/mypurecloud/platform-client-sdk-go` - Genesys Cloud SDK
- `github.com/hashicorp/terraform-plugin-sdk` - Terraform plugin SDK
- Standard library packages: `net/http`, `archive/zip`, `os`, `time`, etc. 