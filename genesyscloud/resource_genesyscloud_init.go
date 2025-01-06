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

	l.RegisterDataSource("genesyscloud_auth_division_home", DataSourceAuthDivisionHome())
	l.RegisterDataSource("genesyscloud_journey_outcome", dataSourceJourneyOutcome())
	l.RegisterDataSource("genesyscloud_journey_segment", dataSourceJourneySegment())
	l.RegisterDataSource("genesyscloud_knowledge_knowledgebase", dataSourceKnowledgeKnowledgebase())
	l.RegisterDataSource("genesyscloud_organizations_me", DataSourceOrganizationsMe())
	l.RegisterDataSource("genesyscloud_quality_forms_evaluation", DataSourceQualityFormsEvaluations())
	l.RegisterDataSource("genesyscloud_quality_forms_survey", dataSourceQualityFormsSurvey())
	l.RegisterDataSource("genesyscloud_widget_deployment", dataSourceWidgetDeployments())
}

func registerResources(l registrar.Registrar) {

	l.RegisterResource("genesyscloud_journey_outcome", ResourceJourneyOutcome())
	l.RegisterResource("genesyscloud_journey_segment", ResourceJourneySegment())
	l.RegisterResource("genesyscloud_knowledge_knowledgebase", ResourceKnowledgeKnowledgebase())
	l.RegisterResource("genesyscloud_quality_forms_evaluation", ResourceEvaluationForm())
	l.RegisterResource("genesyscloud_quality_forms_survey", ResourceSurveyForm())
	l.RegisterResource("genesyscloud_widget_deployment", ResourceWidgetDeployment())

}

func registerExporters(l registrar.Registrar) {
	l.RegisterExporter("genesyscloud_journey_outcome", JourneyOutcomeExporter())
	l.RegisterExporter("genesyscloud_journey_segment", JourneySegmentExporter())
	l.RegisterExporter("genesyscloud_knowledge_knowledgebase", KnowledgeKnowledgebaseExporter())
	l.RegisterExporter("genesyscloud_quality_forms_evaluation", EvaluationFormExporter())
	l.RegisterExporter("genesyscloud_quality_forms_survey", SurveyFormExporter())
	l.RegisterExporter("genesyscloud_widget_deployment", WidgetDeploymentExporter())
}
