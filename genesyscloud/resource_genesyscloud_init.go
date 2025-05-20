package genesyscloud

import (
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(l registrar.Registrar) {

	registerDataSources(l)
	registerResources(l)
	registerExporters(l)
}

func registerDataSources(l registrar.Registrar) {
	l.RegisterDataSource("genesyscloud_auth_division_home", DataSourceAuthDivisionHome())
	l.RegisterDataSource("genesyscloud_organizations_me", DataSourceOrganizationsMe())
	l.RegisterDataSource("genesyscloud_quality_forms_evaluation", DataSourceQualityFormsEvaluations())
}

func registerResources(l registrar.Registrar) {
	l.RegisterResource("genesyscloud_quality_forms_evaluation", ResourceEvaluationForm())
}

func registerExporters(l registrar.Registrar) {
	l.RegisterExporter("genesyscloud_quality_forms_evaluation", EvaluationFormExporter())
}
