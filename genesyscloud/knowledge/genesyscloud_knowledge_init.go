package knowledge

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource("genesyscloud_knowledge_document_variation", ResourceKnowledgeDocumentVariation())
	l.RegisterExporter("genesyscloud_knowledge_document_variation", KnowledgeDocumentVariationExporter())

}
