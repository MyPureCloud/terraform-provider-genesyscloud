package journey_outcome_predictor

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_journey_outcome_predictor_schema.go holds four functions within it:

1.  The registration code that registers the Resource and Exporter for the package.
2.  The resource schema definitions for the journey_outcome_predictor resource.
3.  The resource exporter configuration for the journey_outcome_predictor exporter.
*/
const resourceName = "genesyscloud_journey_outcome_predictor"

// SetRegistrar registers all of the resources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceJourneyOutcomePredictor())
	regInstance.RegisterExporter(resourceName, JourneyOutcomePredictorExporter())
}

// ResourceJourneyOutcomePredictor registers the genesyscloud_journey_outcome_predictor resource with Terraform
func ResourceJourneyOutcomePredictor() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud journey outcome predictor`,

		CreateContext: provider.CreateWithPooledClient(createJourneyOutcomePredictor),
		ReadContext:   provider.ReadWithPooledClient(readJourneyOutcomePredictor),
		DeleteContext: provider.DeleteWithPooledClient(deleteJourneyOutcomePredictor),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"outcome_id": {
				Description: "The outcome associated with this predictor",
				Type: 		 schema.TypeString,
				Required: 	 true,
				ForceNew:    true,
			},
		},
	}
}

// JourneyOutcomePredictorExporter returns the resourceExporter object used to hold the genesyscloud_journey_outcome_predictor exporter's config
func JourneyOutcomePredictorExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthJourneyOutcomePredictors),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"outcome_id": {RefType: "genesyscloud_journey_outcome"},
		},
	}
}
