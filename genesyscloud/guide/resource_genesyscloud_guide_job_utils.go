package guide

func buildGuideJobRequest(prompt, url string) *GenerateGuideContentRequest {
	jobReq := &GenerateGuideContentRequest{}

	if prompt != "" {
		jobReq.Description = &prompt
	}

	if url != "" {
		jobReq.Url = &url
	}

	return jobReq
}

type GenerateGuideContentRequest struct {
	Id          *string `json:"$id,omitempty"`
	Url         *string `json:"url,omitempty"`
	Description *string `json:"description,omitempty"`
}

type GeneratedGuideContent struct {
	Instruction string                `json:"instruction,omitempty"`
	Variables   []Variable            `json:"variables,omitempty"`
	Resources   GuideVersionResources `json:"resources,omitempty"`
}

type JobResponse struct {
	Id           *string                `json:"id,omitempty"`
	GuideId      *string                `json:"guideId,omitempty"`
	Status       *string                `json:"status,omitempty"`
	GuideContent *GeneratedGuideContent `json:"guideContent,omitempty"`
	Errors       []ErrorBody            `json:"errors,omitempty"`
}

type GuideContent struct {
	Instruction string                `json:"instruction,omitempty"`
	Variables   []Variable            `json:"variables,omitempty"`
	Resources   GuideVersionResources `json:"resources,omitempty"`
}

type Variable struct {
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Description string `json:"description,omitempty"`
}

type DataAction struct {
	ID          string `json:"id,omitempty"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
}

type GuideVersionResources struct {
	DataActions []DataAction `json:"dataActions,omitempty"`
}
