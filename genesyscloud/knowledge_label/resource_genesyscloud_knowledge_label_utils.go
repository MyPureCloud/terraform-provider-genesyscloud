package knowledge_label

import (
	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

func BuildCompositeKnowledgeLabelID(labelId, knowledgeBaseId string) string {
	return resourcedata.BuildCompositeID(labelId, knowledgeBaseId)
}

func ParseCompositeKnowledgeLabelID(id string) (labelId, knowledgeBaseId string, _ error) {
	resourceId, relatedIds, err := resourcedata.ParseCompositeID(id)
	if err != nil {
		return "", "", err
	}
	return resourceId, relatedIds[0], nil
}

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
