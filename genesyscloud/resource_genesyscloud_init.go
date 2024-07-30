package genesyscloud

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(l registrar.Registrar) {

	registerDataSources(l)
	registerResources(l)
	registerExporters(l)
}

func registerDataSources(l registrar.Registrar) {

	l.RegisterDataSource("genesyscloud_routing_wrapupcode", DataSourceRoutingWrapupcode())
	l.RegisterDataSource("genesyscloud_location", DataSourceLocation())
	l.RegisterDataSource("genesyscloud_auth_division_home", DataSourceAuthDivisionHome())
	l.RegisterDataSource("genesyscloud_auth_division", dataSourceAuthDivision())
	l.RegisterDataSource("genesyscloud_auth_division_home", DataSourceAuthDivisionHome())
	l.RegisterDataSource("genesyscloud_journey_action_map", dataSourceJourneyActionMap())
	l.RegisterDataSource("genesyscloud_journey_action_template", dataSourceJourneyActionTemplate())
	l.RegisterDataSource("genesyscloud_journey_outcome", dataSourceJourneyOutcome())
	l.RegisterDataSource("genesyscloud_journey_segment", dataSourceJourneySegment())
	l.RegisterDataSource("genesyscloud_knowledge_knowledgebase", dataSourceKnowledgeKnowledgebase())
	l.RegisterDataSource("genesyscloud_knowledge_category", dataSourceKnowledgeCategory())
	l.RegisterDataSource("genesyscloud_knowledge_label", dataSourceKnowledgeLabel())
	l.RegisterDataSource("genesyscloud_location", DataSourceLocation())
	l.RegisterDataSource("genesyscloud_organizations_me", DataSourceOrganizationsMe())
	l.RegisterDataSource("genesyscloud_quality_forms_evaluation", DataSourceQualityFormsEvaluations())
	l.RegisterDataSource("genesyscloud_quality_forms_survey", dataSourceQualityFormsSurvey())
	l.RegisterDataSource("genesyscloud_routing_wrapupcode", DataSourceRoutingWrapupcode())
	l.RegisterDataSource("genesyscloud_user", DataSourceUser())
	l.RegisterDataSource("genesyscloud_widget_deployment", dataSourceWidgetDeployments())
}

func registerResources(l registrar.Registrar) {

	l.RegisterResource("genesyscloud_location", ResourceLocation())
	l.RegisterResource("genesyscloud_auth_division", ResourceAuthDivision())
	l.RegisterResource("genesyscloud_journey_action_map", ResourceJourneyActionMap())
	l.RegisterResource("genesyscloud_journey_action_template", ResourceJourneyActionTemplate())
	l.RegisterResource("genesyscloud_journey_outcome", ResourceJourneyOutcome())
	l.RegisterResource("genesyscloud_journey_segment", ResourceJourneySegment())
	l.RegisterResource("genesyscloud_knowledge_knowledgebase", ResourceKnowledgeKnowledgebase())
	l.RegisterResource("genesyscloud_knowledge_document", ResourceKnowledgeDocument())
	l.RegisterResource("genesyscloud_knowledge_v1_document", ResourceKnowledgeDocumentV1())
	l.RegisterResource("genesyscloud_knowledge_document_variation", ResourceKnowledgeDocumentVariation())
	l.RegisterResource("genesyscloud_knowledge_category", ResourceKnowledgeCategory())
	l.RegisterResource("genesyscloud_knowledge_v1_category", ResourceKnowledgeCategoryV1())
	l.RegisterResource("genesyscloud_knowledge_label", ResourceKnowledgeLabel())
	l.RegisterResource("genesyscloud_location", ResourceLocation())
	l.RegisterResource("genesyscloud_quality_forms_evaluation", ResourceEvaluationForm())
	l.RegisterResource("genesyscloud_quality_forms_survey", ResourceSurveyForm())
	l.RegisterResource("genesyscloud_routing_wrapupcode", ResourceRoutingWrapupCode())
	l.RegisterResource("genesyscloud_user", ResourceUser())
	l.RegisterResource("genesyscloud_widget_deployment", ResourceWidgetDeployment())

}

func registerExporters(l registrar.Registrar) {
	l.RegisterExporter("genesyscloud_auth_division", AuthDivisionExporter())
	l.RegisterExporter("genesyscloud_journey_action_map", JourneyActionMapExporter())
	l.RegisterExporter("genesyscloud_journey_action_template", JourneyActionTemplateExporter())
	l.RegisterExporter("genesyscloud_journey_outcome", JourneyOutcomeExporter())
	l.RegisterExporter("genesyscloud_journey_segment", JourneySegmentExporter())
	l.RegisterExporter("genesyscloud_knowledge_knowledgebase", KnowledgeKnowledgebaseExporter())
	l.RegisterExporter("genesyscloud_knowledge_document", KnowledgeDocumentExporter())
	l.RegisterExporter("genesyscloud_knowledge_category", KnowledgeCategoryExporter())
	l.RegisterExporter("genesyscloud_location", LocationExporter())
	l.RegisterExporter("genesyscloud_quality_forms_evaluation", EvaluationFormExporter())
	l.RegisterExporter("genesyscloud_quality_forms_survey", SurveyFormExporter())
	l.RegisterExporter("genesyscloud_routing_wrapupcode", RoutingWrapupCodeExporter())
	l.RegisterExporter("genesyscloud_user", UserExporter())
	l.RegisterExporter("genesyscloud_widget_deployment", WidgetDeploymentExporter())
	l.RegisterExporter("genesyscloud_knowledge_v1_document", KnowledgeDocumentExporterV1())
	l.RegisterExporter("genesyscloud_knowledge_document_variation", KnowledgeDocumentVariationExporter())
	l.RegisterExporter("genesyscloud_knowledge_label", KnowledgeLabelExporter())
	l.RegisterExporter("genesyscloud_knowledge_v1_category", KnowledgeCategoryExporterV1())

}
