package guide_version

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
