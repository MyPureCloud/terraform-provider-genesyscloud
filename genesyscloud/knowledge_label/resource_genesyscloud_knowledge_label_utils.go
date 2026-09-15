package knowledge_label

import (
	"fmt"
	"strings"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

const knowledgeLabelIdSeparator = ","

// BuildKnowledgeLabelId builds the knowledge_label composite resource ID, which is always in
// the format <knowledge-label-id>,<knowledge-base-id>. Exported so that external consumers of
// this provider's resources don't need to duplicate this logic themselves.
func BuildKnowledgeLabelId(knowledgeLabelId, knowledgeBaseId string) (id string) {
	return fmt.Sprintf("%s%s%s", knowledgeLabelId, knowledgeLabelIdSeparator, knowledgeBaseId)
}

// SplitKnowledgeLabelId splits the knowledge_label composite resource ID, which is always in
// the format <knowledge-label-id>,<knowledge-base-id>, back into its parts. Exported so that
// external consumers of this provider's resources don't need to duplicate this logic themselves.
func SplitKnowledgeLabelId(id string) (knowledgeLabelId string, knowledgeBaseId string) {
	parts := strings.Split(id, knowledgeLabelIdSeparator)
	return parts[0], parts[1]
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
