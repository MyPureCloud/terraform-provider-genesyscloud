package knowledge_knowledgebase

import (
	"fmt"
)

func GenerateKnowledgeKnowledgebaseResource(
	resourceLabel string,
	name string,
	description string,
	coreLanguage string) string {
	return fmt.Sprintf(`resource "genesyscloud_knowledge_knowledgebase" "%s" {
		name = "%s"
        description = "%s"
        core_language = "%s"
	}
	`, resourceLabel, name, description, coreLanguage)
}
