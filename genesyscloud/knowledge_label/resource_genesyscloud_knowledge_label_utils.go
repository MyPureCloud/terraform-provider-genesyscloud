package knowledge_label

import (
	"fmt"
	"strings"

	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
)

func BuildCompositeKnowledgeLabelID(labelId, knowledgeBaseId string) string {
	return fmt.Sprintf("%s,%s", labelId, knowledgeBaseId)
}

func ParseCompositeKnowledgeLabelID(id string) (labelId, knowledgeBaseId string, _ error) {
	parts := strings.Split(id, ",")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid composite knowledge label ID: %s", id)
	}
	return parts[0], parts[1], nil
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
