package main

import (
	"flag"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	dt "terraform-provider-genesyscloud/genesyscloud/architect_datatable"
	dtr "terraform-provider-genesyscloud/genesyscloud/architect_datatable_row"
	emergencyGroup "terraform-provider-genesyscloud/genesyscloud/architect_emergencygroup"
	grammar "terraform-provider-genesyscloud/genesyscloud/architect_grammar"
	grammarLanguage "terraform-provider-genesyscloud/genesyscloud/architect_grammar_language"
	archIvr "terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	authorizatioProduct "terraform-provider-genesyscloud/genesyscloud/authorization_product"
	externalContacts "terraform-provider-genesyscloud/genesyscloud/external_contacts"
	flowMilestone "terraform-provider-genesyscloud/genesyscloud/flow_milestone"
	flowOutcome "terraform-provider-genesyscloud/genesyscloud/flow_outcome"
	"terraform-provider-genesyscloud/genesyscloud/integration"
	integrationAction "terraform-provider-genesyscloud/genesyscloud/integration_action"
	integrationCred "terraform-provider-genesyscloud/genesyscloud/integration_credential"
	integrationCustomAuth "terraform-provider-genesyscloud/genesyscloud/integration_custom_auth_action"
	ob "terraform-provider-genesyscloud/genesyscloud/outbound"
	obAttemptLimit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	obCampaign "terraform-provider-genesyscloud/genesyscloud/outbound_campaign"
	obCampaignRule "terraform-provider-genesyscloud/genesyscloud/outbound_campaignrule"
	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	obs "terraform-provider-genesyscloud/genesyscloud/outbound_ruleset"
	obSequence "terraform-provider-genesyscloud/genesyscloud/outbound_sequence"
	obwm "terraform-provider-genesyscloud/genesyscloud/outbound_wrapupcode_mappings"
	pat "terraform-provider-genesyscloud/genesyscloud/process_automation_trigger"
	recMediaRetPolicy "terraform-provider-genesyscloud/genesyscloud/recording_media_retention_policy"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
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
	webDeployConfig "terraform-provider-genesyscloud/genesyscloud/webdeployments_configuration"
	webDeployDeploy "terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
//go:generate git restore docs/index.md
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

	opts := &plugin.ServeOpts{ProviderFunc: gcloud.New(version, providerResources, providerDataSources)}

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
	dt.SetRegistrar(regInstance)                            //Registering architect data table
	dtr.SetRegistrar(regInstance)                           //Registering architect data table row
	emergencyGroup.SetRegistrar(regInstance)                //Registering architect emergency group
	grammar.SetRegistrar(regInstance)                       //Registering architect grammar
	grammarLanguage.SetRegistrar(regInstance)               //Registering architect grammar language
	edgePhone.SetRegistrar(regInstance)                     //Registering telephony providers edges phone
	edgeSite.SetRegistrar(regInstance)                      //Registering telephony providers edges site
	flowMilestone.SetRegistrar(regInstance)                 //Registering flow milestone
	flowOutcome.SetRegistrar(regInstance)                   //Registering flow outcome
	station.SetRegistrar(regInstance)                       //Registering station
	pat.SetRegistrar(regInstance)                           //Registering process automation triggers
	obs.SetRegistrar(regInstance)                           //Resistering outbound ruleset
	ob.SetRegistrar(regInstance)                            //Registering outbound
	obwm.SetRegistrar(regInstance)                          //Registering outbound wrapup code mappings
	gcloud.SetRegistrar(regInstance)                        //Registering genesyscloud
	obAttemptLimit.SetRegistrar(regInstance)                //Registering outbound attempt limit
	obCampaign.SetRegistrar(regInstance)                    //Registering outbound campaign
	obContactList.SetRegistrar(regInstance)                 //Registering outbound contact list
	obSequence.SetRegistrar(regInstance)                    //Registering outbound sequence
	obCampaignRule.SetRegistrar(regInstance)                //Registering outbound campaignrule
	scripts.SetRegistrar(regInstance)                       //Registering Scripts
	smsAddresses.SetRegistrar(regInstance)                  //Registering routing sms addresses
	integration.SetRegistrar(regInstance)                   //Registering integrations
	integrationCustomAuth.SetRegistrar(regInstance)         //Registering integrations custom auth actions
	integrationAction.SetRegistrar(regInstance)             //Registering integrations actions
	integrationCred.SetRegistrar(regInstance)               //Registering integrations credentials
	recMediaRetPolicy.SetRegistrar(regInstance)             //Registering recording media retention policies
	did.SetRegistrar(regInstance)                           //Registering telephony did
	didPool.SetRegistrar(regInstance)                       //Registering telephony did pools
	archIvr.SetRegistrar(regInstance)                       //Registering architect ivr
	workbin.SetRegistrar(regInstance)                       //Registering task management workbin
	workitemSchema.SetRegistrar(regInstance)                //Registering task management workitem schema
	worktype.SetRegistrar(regInstance)                      //Registering task management worktype
	workitem.SetRegistrar(regInstance)                      //Registering task management workitem
	externalContacts.SetRegistrar(regInstance)              //Registering external contacts
	team.SetRegistrar(regInstance)                          //Registering team
	telephony.SetRegistrar(regInstance)                     //Registering telephony package
	edgeGroup.SetRegistrar(regInstance)                     //Registering edges edge group
	webDeployConfig.SetRegistrar(regInstance)               //Registering webdeployments_config
	webDeployDeploy.SetRegistrar(regInstance)               //Registering webdeployments_deploy
	authorizatioProduct.SetRegistrar(regInstance)           //Registering Authorization Product
	extPool.SetRegistrar(regInstance)                       //Registering Extension Pool
	phoneBaseSettings.SetRegistrar(regInstance)             //Registering Phone Base Settings
	lineBaseSettings.SetRegistrar(regInstance)              //Registering Line Base Settings
	edgesTrunk.SetRegistrar(regInstance)                    //Registering Edges Trunk Settings
	resourceExporter.SetRegisterExporter(resourceExporters) //Registering register exporters
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
