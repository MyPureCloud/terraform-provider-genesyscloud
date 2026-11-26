package quality_miner

// @team: PureCloud Speech & Text Analytics
// @jira: GM
// @description: Manage Quality Miners (Intent or Topic). A miner analyzes historical conversations to surface the dominant intents or topics. Miners are immutable once created; changing any attribute recreates the miner.

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_quality_miner"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceQualityMiner())
	regInstance.RegisterDataSource(ResourceType, DataSourceQualityMiner())
	regInstance.RegisterExporter(ResourceType, QualityMinerExporter())
}

func ResourceQualityMiner() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Quality Miner (Intent or Topic). A miner has no update API, so changing any argument will drop and recreate the miner with a new ID.",

		CreateContext: provider.CreateWithPooledClient(createQualityMiner),
		ReadContext:   provider.ReadWithPooledClient(readQualityMiner),
		DeleteContext: provider.DeleteWithPooledClient(deleteQualityMiner),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "The name of the miner (maximum 30 characters). Changing this attribute will cause the miner to be dropped and recreated with a new ID.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 30),
			},
			"miner_type": {
				Description:  "Type of the miner. Changing this attribute will cause the miner to be dropped and recreated with a new ID.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Intent", "Topic"}, false),
			},
			"language": {
				Description: "Language localization code of the miner (e.g. 'en-US'). The API normalizes this to lowercase and assigns a default when omitted, so case differences are ignored and the computed value is accepted. Changing this attribute will cause the miner to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				// The API lowercases the language code (e.g. "en-US" -> "en-us"), so treat case
				// differences as equivalent to avoid a consistency mismatch / perpetual diff.
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
		},
	}
}

func DataSourceQualityMiner() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Quality Miners. Select a miner by name, optionally narrowing by type.",
		ReadContext: provider.ReadWithPooledClient(dataSourceQualityMinerRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the miner.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"miner_type": {
				Description:  "Type of the miner to narrow the search by. Optional; useful when miners of different types share a name.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Intent", "Topic"}, false),
			},
		},
	}
}

func QualityMinerExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllQualityMiners),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
	}
}

func GenerateQualityMinerResource(
	resourceLabel string,
	name string,
	minerType string,
	attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_quality_miner" "%s" {
	name       = "%s"
	miner_type = "%s"
	%s
}
`, resourceLabel, name, minerType, strings.Join(attrs, "\n\t"))
}
