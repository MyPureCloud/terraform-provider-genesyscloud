package knowledge_label

import (
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func buildKnowledgeLabel(labelIn map[string]interface{}) platformclientv2.Labelcreaterequest {
	name := labelIn["name"].(string)
	color := labelIn["color"].(string)

	labelOut := platformclientv2.Labelcreaterequest{
		Name:  &name,
		Color: &color,
	}

	return labelOut
}

func buildKnowledgeLabelUpdate(labelIn map[string]interface{}) platformclientv2.Labelupdaterequest {
	name := labelIn["name"].(string)
	color := labelIn["color"].(string)

	labelOut := platformclientv2.Labelupdaterequest{
		Name:  &name,
		Color: &color,
	}

	return labelOut
}

func flattenKnowledgeLabel(labelIn *platformclientv2.Labelresponse) []interface{} {
	labelOut := make(map[string]interface{})

	if labelIn.Name != nil {
		labelOut["name"] = *labelIn.Name
	}
	if labelIn.Color != nil {
		labelOut["color"] = *labelIn.Color
	}

	return []interface{}{labelOut}
}
