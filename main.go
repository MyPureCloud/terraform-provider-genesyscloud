package main

import (
	"flag"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	dt "terraform-provider-genesyscloud/genesyscloud/architect_datatable"
	dtr "terraform-provider-genesyscloud/genesyscloud/architect_datatable_row"
	emergencyGroup "terraform-provider-genesyscloud/genesyscloud/architect_emergencygroup"
	flow "terraform-provider-genesyscloud/genesyscloud/architect_flow"
	grammar "terraform-provider-genesyscloud/genesyscloud/architect_grammar"
	grammarLanguage "terraform-provider-genesyscloud/genesyscloud/architect_grammar_language"
	archIvr "terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	architectSchedulegroups "terraform-provider-genesyscloud/genesyscloud/architect_schedulegroups"
	architectSchedules "terraform-provider-genesyscloud/genesyscloud/architect_schedules"
	userPrompt "terraform-provider-genesyscloud/genesyscloud/architect_user_prompt"
	authRole "terraform-provider-genesyscloud/genesyscloud/auth_role"
	authorizatioProduct "terraform-provider-genesyscloud/genesyscloud/authorization_product"
	employeeperformanceExternalmetricsDefinition "terraform-provider-genesyscloud/genesyscloud/employeeperformance_externalmetrics_definitions"
	externalContacts "terraform-provider-genesyscloud/genesyscloud/external_contacts"
	flowLogLevel "terraform-provider-genesyscloud/genesyscloud/flow_loglevel"
	flowMilestone "terraform-provider-genesyscloud/genesyscloud/flow_milestone"
	flowOutcome "terraform-provider-genesyscloud/genesyscloud/flow_outcome"
	"terraform-provider-genesyscloud/genesyscloud/group"
	groupRoles "terraform-provider-genesyscloud/genesyscloud/group_roles"
	idpSalesforce "terraform-provider-genesyscloud/genesyscloud/idp_salesforce"
	"terraform-provider-genesyscloud/genesyscloud/integration"
	integrationAction "terraform-provider-genesyscloud/genesyscloud/integration_action"
	integrationCred "terraform-provider-genesyscloud/genesyscloud/integration_credential"
	integrationCustomAuth "terraform-provider-genesyscloud/genesyscloud/integration_custom_auth_action"
	journeyOutcomePredictor "terraform-provider-genesyscloud/genesyscloud/journey_outcome_predictor"
	oauth "terraform-provider-genesyscloud/genesyscloud/oauth_client"
	oAuthSettings "terraform-provider-genesyscloud/genesyscloud/organization_authentication_settings"
	oAuthPairing "terraform-provider-genesyscloud/genesyscloud/orgauthorization_pairing"
	ob "terraform-provider-genesyscloud/genesyscloud/outbound"
	obAttemptLimit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	obCallableTimeset "terraform-provider-genesyscloud/genesyscloud/outbound_callabletimeset"
	obCallResponseSet "terraform-provider-genesyscloud/genesyscloud/outbound_callanalysisresponseset"
	obCampaign "terraform-provider-genesyscloud/genesyscloud/outbound_campaign"
	obCampaignRule "terraform-provider-genesyscloud/genesyscloud/outbound_campaignrule"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obContactListFilter "terraform-provider-genesyscloud/genesyscloud/outbound_contactlistfilter"
	obDncList "terraform-provider-genesyscloud/genesyscloud/outbound_dnclist"
	obfst "terraform-provider-genesyscloud/genesyscloud/outbound_filespecificationtemplate"
	obs "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	obSequence "terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
	obSettings "terraform-provider-genesyscloud/genesyscloud/outbound_settings"
	obwm "terraform-provider-genesyscloud/genesyscloud/outbound_wrapupcode_mappings"
	pat "terraform-provider-genesyscloud/genesyscloud/process_automation_trigger"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	recMediaRetPolicy "terraform-provider-genesyscloud/genesyscloud/recording_media_retention_policy"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	respmanagementLibrary "terraform-provider-genesyscloud/genesyscloud/responsemanagement_library"
	responsemanagementResponse "terraform-provider-genesyscloud/genesyscloud/responsemanagement_response"
	responsemanagementResponseasset "terraform-provider-genesyscloud/genesyscloud/responsemanagement_responseasset"
	routingEmailRoute "terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	routingQueueConditionalGroupRouting "terraform-provider-genesyscloud/genesyscloud/routing_queue_conditional_group_routing"
	routingQueueOutboundEmailAddress "terraform-provider-genesyscloud/genesyscloud/routing_queue_outbound_email_address"
	smsAddresses "terraform-provider-genesyscloud/genesyscloud/routing_sms_addresses"
	"terraform-provider-genesyscloud/genesyscloud/scripts"
	"terraform-provider-genesyscloud/genesyscloud/station"
	workbin "terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workitem "terraform-provider-genesyscloud/genesyscloud/task_management_workitem"
	workitemSchema "terraform-provider-genesyscloud/genesyscloud/task_management_workitem_schema"
	worktype "terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	"terraform-provider-genesyscloud/genesyscloud/team"
	"terraform-provider-genesyscloud/genesyscloud/telephony"
	did "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did"
	didPool "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	edgeGroup "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_edge_group"
	extPool "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
	lineBaseSettings "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_linebasesettings"
	edgePhone "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phone"
	phoneBaseSettings "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_phonebasesettings"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	edgesTrunk "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunk"
	tfexp "terraform-provider-genesyscloud/genesyscloud/tfexporter"
	userRoles "terraform-provider-genesyscloud/genesyscloud/user_roles"
	webDeployConfig "terraform-provider-genesyscloud/genesyscloud/webdeployments_configuration"
	webDeployDeploy "terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
//go:generate go run terraform-provider-genesyscloud/apidocs
var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	// 0.1.0 is the default version for developing locally
	version string = "0.1.0"

	// goreleaser can also pass the specific commit if you want
	// commit  string = ""
)

var providerResources map[string]*schema.Resource
var providerDataSources map[string]*schema.Resource
var resourceExporters map[string]*resourceExporter.ResourceExporter

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	providerResources = make(map[string]*schema.Resource)
	providerDataSources = make(map[string]*schema.Resource)
	resourceExporters = make(map[string]*resourceExporter.ResourceExporter)

	registerResources()

	opts := &plugin.ServeOpts{ProviderFunc: provider.New(version, providerResources, providerDataSources)}

	if debugMode {
		opts.Debug = true
		opts.ProviderAddr = "registry.terraform.io/mypurecloud/genesyscloud"
	}
	plugin.Serve(opts)
}

type RegisterInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
	exporterMapMutex   sync.RWMutex
}

func registerResources() {
	regInstance := &RegisterInstance{}
	authRole.SetRegistrar(regInstance)                                     //Registering auth_role
	oauth.SetRegistrar(regInstance)                                        //Registering oauth_client
	dt.SetRegistrar(regInstance)                                           //Registering architect data table
	dtr.SetRegistrar(regInstance)                                          //Registering architect data table row
	emergencyGroup.SetRegistrar(regInstance)                               //Registering architect emergency group
	architectSchedulegroups.SetRegistrar(regInstance)                      //Registering architect schedule groups
	architectSchedules.SetRegistrar(regInstance)                           //Registering architect schedules
	employeeperformanceExternalmetricsDefinition.SetRegistrar(regInstance) //Registering employee performance external metrics definitions
	grammar.SetRegistrar(regInstance)                                      //Registering architect grammar
	grammarLanguage.SetRegistrar(regInstance)                              //Registering architect grammar language
	groupRoles.SetRegistrar(regInstance)                                   //Registering group roles
	edgePhone.SetRegistrar(regInstance)                                    //Registering telephony providers edges phone
	edgeSite.SetRegistrar(regInstance)                                     //Registering telephony providers edges site
	flow.SetRegistrar(regInstance)                                         //Registering architect flow
	flowLogLevel.SetRegistrar(regInstance)                                 //Registering flow log Level
	flowMilestone.SetRegistrar(regInstance)                                //Registering flow milestone
	flowOutcome.SetRegistrar(regInstance)                                  //Registering flow outcome
	station.SetRegistrar(regInstance)                                      //Registering station
	pat.SetRegistrar(regInstance)                                          //Registering process automation triggers
	obs.SetRegistrar(regInstance)                                          //Resistering outbound ruleset
	ob.SetRegistrar(regInstance)                                           //Registering outbound
	obwm.SetRegistrar(regInstance)                                         //Registering outbound wrapup code mappings
	oAuthSettings.SetRegistrar(regInstance)                                //Registering organization authentication settings
	gcloud.SetRegistrar(regInstance)                                       //Registering genesyscloud
	obAttemptLimit.SetRegistrar(regInstance)                               //Registering outbound attempt limit
	obCallableTimeset.SetRegistrar(regInstance)                            //Registering outbound callable timeset
	obCampaign.SetRegistrar(regInstance)                                   //Registering outbound campaign
	obContactList.SetRegistrar(regInstance)                                //Registering outbound contact list
	obContactListFilter.SetRegistrar(regInstance)                          //Registering outbound contact list filter
	obSequence.SetRegistrar(regInstance)                                   //Registering outbound sequence
	obCampaignRule.SetRegistrar(regInstance)                               //Registering outbound campaignrule
	obSettings.SetRegistrar(regInstance)                                   //Registering outbound settings
	obCallResponseSet.SetRegistrar(regInstance)                            //Registering outbound call analysis response set
	obCampaign.SetRegistrar(regInstance)                                   //Registering outbound campaign
	obfst.SetRegistrar(regInstance)                                        //Registering outbound file specification template
	obDncList.SetRegistrar(regInstance)                                    //Registering outbound dnclist
	oAuthPairing.SetRegistrar(regInstance)                                 //Registering orgauthorization pairing
	scripts.SetRegistrar(regInstance)                                      //Registering Scripts
	smsAddresses.SetRegistrar(regInstance)                                 //Registering routing sms addresses
	idpSalesforce.SetRegistrar(regInstance)                                //Registering idp salesforce
	integration.SetRegistrar(regInstance)                                  //Registering integrations
	integrationCustomAuth.SetRegistrar(regInstance)                        //Registering integrations custom auth actions
	integrationAction.SetRegistrar(regInstance)                            //Registering integrations actions
	integrationCred.SetRegistrar(regInstance)                              //Registering integrations credentials
	recMediaRetPolicy.SetRegistrar(regInstance)                            //Registering recording media retention policies
	responsemanagementResponse.SetRegistrar(regInstance)                   //Registering responsemanagement responses
	responsemanagementResponseasset.SetRegistrar(regInstance)              //Registering responsemanagement response asset
	respmanagementLibrary.SetRegistrar(regInstance)                        //Registering responsemanagement library
	routingEmailRoute.SetRegistrar(regInstance)                            //Registering routing email route
	did.SetRegistrar(regInstance)                                          //Registering telephony did
	didPool.SetRegistrar(regInstance)                                      //Registering telephony did pools
	archIvr.SetRegistrar(regInstance)                                      //Registering architect ivr
	workbin.SetRegistrar(regInstance)                                      //Registering task management workbin
	workitemSchema.SetRegistrar(regInstance)                               //Registering task management workitem schema
	worktype.SetRegistrar(regInstance)                                     //Registering task management worktype
	workitem.SetRegistrar(regInstance)                                     //Registering task management workitem
	externalContacts.SetRegistrar(regInstance)                             //Registering external contacts
	team.SetRegistrar(regInstance)                                         //Registering team
	telephony.SetRegistrar(regInstance)                                    //Registering telephony package
	edgeGroup.SetRegistrar(regInstance)                                    //Registering edges edge group
	webDeployConfig.SetRegistrar(regInstance)                              //Registering webdeployments_config
	webDeployDeploy.SetRegistrar(regInstance)                              //Registering webdeployments_deploy
	authorizatioProduct.SetRegistrar(regInstance)                          //Registering Authorization Product
	extPool.SetRegistrar(regInstance)                                      //Registering Extension Pool
	phoneBaseSettings.SetRegistrar(regInstance)                            //Registering Phone Base Settings
	lineBaseSettings.SetRegistrar(regInstance)                             //Registering Line Base Settings
	edgesTrunk.SetRegistrar(regInstance)                                   //Registering Edges Trunk Settings
	resourceExporter.SetRegisterExporter(resourceExporters)                //Registering register exporters
	userRoles.SetRegistrar(regInstance)                                    //Registering user roles
	journeyOutcomePredictor.SetRegistrar(regInstance)                      //Registering journey outcome predictor
	group.SetRegistrar(regInstance)                                        //Registering group
	userPrompt.SetRegistrar(regInstance)                                   //Registering user prompt
	routingQueue.SetRegistrar(regInstance)                                 //Registering routing queue
	routingQueueConditionalGroupRouting.SetRegistrar(regInstance)          //Registering routing queue conditional group routing
	routingQueueOutboundEmailAddress.SetRegistrar(regInstance)             //Registering routing queue outbound email address

	// setting resources for Use cases  like TF export where provider is used in resource classes.
	tfexp.SetRegistrar(regInstance) //Registering tf exporter
	registrar.SetResources(providerResources, providerDataSources)

}

func (r *RegisterInstance) RegisterResource(resourceName string, resource *schema.Resource) {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()
	providerResources[resourceName] = resource
}

func (r *RegisterInstance) RegisterDataSource(dataSourceName string, datasource *schema.Resource) {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()
	providerDataSources[dataSourceName] = datasource
}

func (r *RegisterInstance) RegisterExporter(exporterName string, resourceExporter *resourceExporter.ResourceExporter) {
	r.exporterMapMutex.Lock()
	defer r.exporterMapMutex.Unlock()
	resourceExporters[exporterName] = resourceExporter
}
