package telephony

import (
	"fmt"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTrunkBaseSettings(t *testing.T) {
	t.Parallel()
	var (
		trunkBaseSettingsRes     = "trunkBaseSettings"
		trunkBaseSettingsDataRes = "trunkBaseSettingsData"
		name                     = "test trunk base settings " + uuid.NewString()
		description              = "test description"
		trunkMetaBaseId          = "external_sip_pcv_byoc_carrier.json"
		trunkType                = "EXTERNAL"
		managed                  = false
		locationResourceId       = "location"
		siteId                   = "site"
	)

	referencedResources := gcloud.GenerateLocationResource(
		locationResourceId,
		"tf location "+uuid.NewString(),
		"HQ1",
		[]string{},
		gcloud.GenerateLocationEmergencyNum(
			"+13178791201",
			gcloud.NullValue,
		),
		gcloud.GenerateLocationAddress(
			"7601 Interactive Way",
			"Indianapolis",
			"IN",
			"US",
			"46278",
		),
	) + edgeSite.GenerateSiteResourceWithCustomAttrs(
		siteId,
		"tf site "+uuid.NewString(),
		"test description",
		"genesyscloud_location."+locationResourceId+".id",
		"Cloud",
		false,
		"[\"us-east-1\"]",
		gcloud.NullValue,
		gcloud.NullValue,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: referencedResources + GenerateTrunkBaseSettingsResourceWithCustomAttrs(
					trunkBaseSettingsRes,
					name,
					description,
					trunkMetaBaseId,
					"genesyscloud_telephony_providers_edges_site."+siteId+".id",
					trunkType,
					managed,
				) + generateTrunkBaseSettingsDataSource(
					trunkBaseSettingsDataRes,
					name,
					"genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsDataRes, "id", "genesyscloud_telephony_providers_edges_trunkbasesettings."+trunkBaseSettingsRes, "id"),
				),
			},
		},
	})
}

func generateTrunkBaseSettingsDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_trunkbasesettings" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
