package agentic_virtual_agent_version_publish

/*
   resource_genesyscloud_agentic_virtual_agent_version_publish_utils.go contains Go struct
   definitions for the publish job API request/response models.
*/

// PublishJobRequest is the request body for creating a publish job.
type PublishJobRequest struct {
	VirtualAgentVersion *PublishJobVersionStatus `json:"virtualAgentVersion"`
}

// PublishJobVersionStatus holds the target status for publishing.
type PublishJobVersionStatus struct {
	Status string `json:"status"`
}

// PublishJobResponse is the response from creating or polling a publish job.
type PublishJobResponse struct {
	Id                  string          `json:"id,omitempty"`
	Status              string          `json:"status,omitempty"`
	VirtualAgentVersion *VersionSummary `json:"virtualAgentVersion,omitempty"`
	Errors              []JobError      `json:"errors,omitempty"`
	TokenCount          *int            `json:"tokenCount,omitempty"`
	SelfUri             string          `json:"selfUri,omitempty"`
}

// VersionSummary is the version info returned in a successful job response.
type VersionSummary struct {
	Version *string `json:"version,omitempty"`
	Status  *string `json:"status,omitempty"`
}

// JobError represents an error entry in a failed job response.
type JobError struct {
	Message string       `json:"message,omitempty"`
	Code    string       `json:"code,omitempty"`
	Status  int          `json:"status,omitempty"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail provides field-level error information.
type ErrorDetail struct {
	FieldName string `json:"fieldName,omitempty"`
}

// VersionStatusResponse is a minimal version response used for read (checking version status).
type VersionStatusResponse struct {
	Version *string `json:"version,omitempty"`
	Status  *string `json:"status,omitempty"`
}
