package journey_view_schedule

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
resource_genesycloud_journey_view_schedule_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the journey_view_schedule resource.
3.  The datasource schema definitions for the journey_view_schedule datasource.
4.  The resource exporter configuration for the journey_view_schedule exporter.
*/

const ResourceType = "genesyscloud_journey_view_schedule"

// SetRegistrar registers all the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceJourneyViewSchedule())
	regInstance.RegisterExporter(ResourceType, JourneyViewScheduleExporter())
	// No datasource
}

// ResourceJourneyViewSchedule registers the genesyscloud_journey_view_schedule resource with Terraform
func ResourceJourneyViewSchedule() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Journey View Schedule",

		CreateContext: provider.CreateWithPooledClient(createJourneyViewSchedule),
		ReadContext:   provider.ReadWithPooledClient(readJourneyViewSchedule),
		UpdateContext: provider.UpdateWithPooledClient(updateJourneyViewSchedule),
		DeleteContext: provider.DeleteWithPooledClient(deleteJourneyViewSchedule),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"journey_view_id": {
				Description: "Journey view ID of the schedule. Changing this will cause the schedule to be dropped and recreated for the new view ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"frequency": {
				Description:  "Frequency of execution (Daily | Weekly | Monthly).",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Daily", "Weekly", "Monthly"}, false),
			},
			// All other fields like dateModified and last modified user are read only, cannot be set by user
		},
	}
}

// JourneyViewScheduleExporter returns the resourceExporter object used to hold the genesyscloud_journey_view_schedule exporter's config
func JourneyViewScheduleExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllJourneyViewSchedule),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
	}
}
