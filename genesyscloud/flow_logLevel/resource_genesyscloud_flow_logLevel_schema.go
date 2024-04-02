package flow_logLevel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_flow_loglevel_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the flow_loglevel resource.
3.  The datasource schema definitions for the flow_loglevel datasource.
4.  The resource exporter configuration for the flow_loglevel exporter.
*/
const resourceName = "genesyscloud_flow_loglevel"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceFlowLoglevel())
	regInstance.RegisterExporter(resourceName, FlowLoglevelExporter())
}

// ResourceFlowLoglevel registers the genesyscloud_flow_loglevel resource with Terraform
func ResourceFlowLoglevel() *schema.Resource {

	flowCharacteristics := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"communications": {
				Description: "Communications are either audio or digital communications sent to or received from a participant.  An example here would be the initial greeting in an inbound call flow where it plays a greeting message to the participant.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"event_error": {
				Description: "Whether to report flow error events.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"event_other": {
				Description: "Whether to report events other than errors or warnings such as a language change, loop event.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"event_warning": {
				Description: "Whether to report flow warning events.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"execution_input_outputs": {
				Description: "Whether to report input setting values and output data values for individual execution items above.  For example, if you have FlowExecutionInputOutputs and a Call Data Action ran in a flow, if FlowExecutionItems was enabled you'd see the fact a Call Data Action ran and the output path it took but nothing about which Data Action it ran, the input data sent to it at flow runtime and the data returned from it.  If you enable this characteristic, execution data will contain this additional detail.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"execution_items": {
				Description: "Whether to report execution data about individual actions, menus, states, tasks, etc. etc. that ran during execution of the flow.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"names": {
				Description: "This characteristic specifies whether or not name information should be emitted in execution data such as action, task, state or even the flow name itself.  Names are very handy from a readability standpoint but they do take up additional space in flow execution data instances.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"variables": {
				Description: "Whether to report assignment of values to variables in flow execution data. It's important to remember there is a difference between variable value assignments and output data from an action.  If you have a Call Digital Bot flow action in an Inbound Message flow and there is no variable bound to the Exit Reason output but FlowExecutionInputOutputs is enabled, you will still be able to see the exit reason from the digital bot flow in execution data even though it is not bound to a variable.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
	return &schema.Resource{
		Description: `Genesys Cloud flow log level`,

		CreateContext: provider.CreateWithPooledClient(createFlowLogLevel),
		ReadContext:   provider.ReadWithPooledClient(readFlowLogLevel),
		UpdateContext: provider.UpdateWithPooledClient(updateFlowLogLevel),
		DeleteContext: provider.DeleteWithPooledClient(deleteFlowLogLevel),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"flow_id": {
				Description: "The flowId for this characteristics set",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"flow_log_level": {
				Description: "The logLevel for this characteristics set",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"flow_characteristics": {
				Description: "Shows what characteristics are enabled for this log level",
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				Elem:        flowCharacteristics,
			},
		},
	}
}

// FlowLoglevelExporter returns the resourceExporter object used to hold the genesyscloud_flow_logLevels exporter's config
func FlowLoglevelExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllFlowLogLevels),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"external_organization": {}, //Need to add this when we external orgs implemented
		},
	}
}
