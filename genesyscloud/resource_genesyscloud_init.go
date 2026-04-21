package genesyscloud

import (
	caseManagementCaseplan "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/case_management_caseplan"
	caseManagementStageplan "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/case_management_stageplan"
	caseManagementStepplan "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/case_management_stepplan"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(l registrar.Registrar) {
	caseManagementCaseplan.SetRegistrar(l)
	caseManagementStageplan.SetRegistrar(l)
	caseManagementStepplan.SetRegistrar(l)

	registerDataSources(l)
	registerResources(l)
	registerExporters(l)
}

func registerDataSources(l registrar.Registrar) {
	l.RegisterDataSource("genesyscloud_auth_division_home", DataSourceAuthDivisionHome())
	l.RegisterDataSource(DataSourceOrganizationsMeResourceType, DataSourceOrganizationsMe())
}

func registerResources(l registrar.Registrar) {
}

func registerExporters(l registrar.Registrar) {
}
