package agentic_virtual_agent

/*
   resource_genesyscloud_agentic_virtual_agent_utils.go contains Go struct definitions
   for the Agentic Virtual Agent API request/response models, and helper functions
   for marshaling/unmarshaling data between Terraform state and the API.
*/

// AgenticVirtualAgent represents the full agent resource as returned by the API.
type AgenticVirtualAgent struct {
	Id                           *string                     `json:"id,omitempty"`
	Name                         *string                     `json:"name,omitempty"`
	ImageUri                     *string                     `json:"imageUri,omitempty"`
	Status                       *string                     `json:"status,omitempty"`
	LatestSavedVersion           *AgenticVirtualAgentVersion `json:"latestSavedVersion,omitempty"`
	LatestProductionReadyVersion *AgenticVirtualAgentVersion `json:"latestProductionReadyVersion,omitempty"`
	DateCreated                  *string                     `json:"dateCreated,omitempty"`
	DateModified                 *string                     `json:"dateModified,omitempty"`
	SelfUri                      *string                     `json:"selfUri,omitempty"`
}

// AgenticVirtualAgentVersion represents the version reference object returned on the agent.
// The API returns these as objects with { version, selfUri }, not scalar strings.
type AgenticVirtualAgentVersion struct {
	Version *string `json:"version,omitempty"`
	SelfUri *string `json:"selfUri,omitempty"`
}

// AgenticVirtualAgentCreate represents the request body for creating an agent.
type AgenticVirtualAgentCreate struct {
	Name     string  `json:"name"`
	ImageUri *string `json:"imageUri,omitempty"`
}

// AgenticVirtualAgentUpdate represents the request body for updating (PATCH) an agent.
type AgenticVirtualAgentUpdate struct {
	Name     string  `json:"name"`
	ImageUri *string `json:"imageUri,omitempty"`
}

// AgenticVirtualAgentEntityListing represents the paginated list response.
type AgenticVirtualAgentEntityListing struct {
	Entities   *[]AgenticVirtualAgent `json:"entities,omitempty"`
	PageSize   *int                   `json:"pageSize,omitempty"`
	PageNumber *int                   `json:"pageNumber,omitempty"`
	Total      *int                   `json:"total,omitempty"`
	PageCount  *int                   `json:"pageCount,omitempty"`
}

// AgenticVirtualAgentJob represents the async delete job resource.
type AgenticVirtualAgentJob struct {
	Id      string         `json:"id,omitempty"`
	Status  string         `json:"status,omitempty"`
	Errors  []JobErrorBody `json:"errors,omitempty"`
	SelfUri string         `json:"selfUri,omitempty"`
}

// JobErrorBody represents an error entry in a failed job response.
type JobErrorBody struct {
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
	Status  int    `json:"status,omitempty"`
}
