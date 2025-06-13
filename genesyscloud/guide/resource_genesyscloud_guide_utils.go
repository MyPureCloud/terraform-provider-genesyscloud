package guide

type Guide struct {
	Id                           *string `json:"id,omitempty"`
	Name                         *string `json:"name,omitempty"`
	Source                       *string `json:"source,omitempty"`
	Status                       *string `json:"status,omitempty"`
	LatestSavedVersion           *string `json:"latestSavedVersion,omitempty"`
	LatestProductionReadyVersion *string `json:"latestProductionReadyVersion,omitempty"`
}

//type GuideJob struct {
//	SetFieldNames map[string]bool `json:"setFieldNames,omitempty"`
//	Id:            nil,
//	Status:        nil,
//	Errors:        nil,
//	Guide:         nil,
//	SelfUri:       nil,
//}

//type Guidecontentgenerationjob struct {
//	SetFieldNames map[string]bool `json:"setFieldNames,omitempty"`
//	Id:            nil,
//	Status:        nil,
//	Errors:        nil,
//	Guide:         nil,
//	GuideContent:  nil,
//	SelfUri:       nil,
//}

type Createguide struct {
	SetFieldNames map[string]bool `json:"setFieldNames,omitempty"`
	Name          *string         `json:"name,omitempty"`
	Source        *string         `json:"source,omitempty"`
}

//type Updateguide struct {
//	SetFieldNames map[string]bool `json:"setFieldNames,omitempty"`
//	Instruction:   nil,
//	Variables:     nil,
//	Resources:     nil,
//}

type Createguideversion struct {
	SetFieldNames map[string]bool `json:"setFieldNames,omitempty"`
	Name          *string         `json:"name,omitempty"`
	Source        *string         `json:"source,omitempty"`
}

//type Updateguideversion struct {
//	SetFieldNames map[string]bool `json:"setFieldNames,omitempty"`
//	Instruction:   nil,
//	Variables:     nil,
//	Resources:     nil,
//}
//
//type Generatedguidecontent struct{
//	SetFieldNames map[string]bool `json:"setFieldNames,omitempty"`
//	Instruction:   nil,
//	Variables:     nil,
//	Resources:     nil,
//}
//
//type Generateguidecontentrequest struct {
//	SetFieldNames map[string]bool `json:"setFieldNames,omitempty"`
//	Description:   nil,
//	Url:           nil,
//}
