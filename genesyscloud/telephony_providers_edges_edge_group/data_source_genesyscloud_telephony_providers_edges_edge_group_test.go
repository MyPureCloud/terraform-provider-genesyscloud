package telephony_providers_edges_edge_group

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	tpetbs "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_trunkbasesettings"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceEdgeGroup(t *testing.T) {
	t.Parallel()
	var (
		edgeGroupResourceLabel      = "edgeGroup1234"
		edgeGroupDataLabel          = "edgeGroupData"
		edgeGroupName1              = "test edge group " + uuid.NewString()
		edgeGroupDescription1       = "test description 1"
		edgeGroupResourceFullPath   = ResourceType + "." + edgeGroupResourceLabel
		edgeGroupDataSourceFullPath = "data." + ResourceType + "." + edgeGroupDataLabel

		phoneTrunkBaseSettingsResourceLabel1   = "phoneTrunkBaseSettingsRes1"
		phoneTrunkBaseSettingsResourceFullPath = tpetbs.ResourceType + "." + phoneTrunkBaseSettingsResourceLabel1
	)

	phoneTrunkBaseSetting1 := tpetbs.GenerateTrunkBaseSettingsResourceWithCustomAttrs(
		phoneTrunkBaseSettingsResourceLabel1,
		"phone trunk base settings "+uuid.NewString(),
		"",
		"phone_connections_webrtc.json",
		"PHONE",
		false)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: phoneTrunkBaseSetting1 + GenerateEdgeGroupResourceWithCustomAttrs(
					edgeGroupResourceLabel,
					edgeGroupName1,
					edgeGroupDescription1,
					false,
					false,
					GeneratePhoneTrunkBaseIds(phoneTrunkBaseSettingsResourceFullPath+".id"),
				),
			},
			{
				PreConfig: func() {
					t.Log("Sleeping for 1 second")
					time.Sleep(1 * time.Second)
				},
				Config: phoneTrunkBaseSetting1 + GenerateEdgeGroupResourceWithCustomAttrs(
					edgeGroupResourceLabel,
					edgeGroupName1,
					edgeGroupDescription1,
					false,
					false,
					GeneratePhoneTrunkBaseIds(phoneTrunkBaseSettingsResourceFullPath+".id"),
				) + generateEdgeGroupDataSource(
					edgeGroupDataLabel,
					edgeGroupName1,
					edgeGroupResourceFullPath,
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(edgeGroupDataSourceFullPath, "id", edgeGroupResourceFullPath, "id"),
				),
			},
		},
		CheckDestroy: testVerifyEdgeGroupsDestroyed,
	})
}

/*
This test expects that the org has a product called "voice" enabled on it. If the test org does not have this product on it, the test can be skipped or ignored.
*/
func TestAccDataSourceEdgeGroupManaged(t *testing.T) {
	t.Parallel()
	var (
		edgeGroupDataLabel    = "edgeGroupData"
		edgeGroupDataFullPath = fmt.Sprintf("data.%s.%s", ResourceType, edgeGroupDataLabel)
		edgeGroupName1        = "PureCloud Voice - AWS"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateEdgeGroupDataSource(
					edgeGroupDataLabel,
					edgeGroupName1,
					"",
					true,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(edgeGroupDataFullPath, "name", edgeGroupName1),
				),
			},
		},
	})
}

func generateEdgeGroupDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string,
	managed bool) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_edge_group" "%s" {
		name = "%s"
		managed = %t
		depends_on=[%s]
	}
	`, resourceLabel, name, managed, dependsOnResource)
}
