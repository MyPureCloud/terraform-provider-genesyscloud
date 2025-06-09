package guides

type Guide struct {
	Id                           *string `json:"id,omitempty"`
	Name                         *string `json:"name,omitempty"`
	Source                       *string `json:"source,omitempty"`
	Status                       *string `json:"status,omitempty"`
	LatestSavedVersion           *string `json:"latestSavedVersion,omitempty"`
	LatestProductionReadyVersion *string `json:"latestProductionReadyVersion,omitempty"`
}
