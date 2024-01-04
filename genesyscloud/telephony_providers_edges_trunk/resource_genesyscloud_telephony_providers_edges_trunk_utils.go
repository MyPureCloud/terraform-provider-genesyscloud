package telephony_providers_edges_trunk

import "fmt"

func generateTrunkDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_trunk" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
func generateTrunk(
	trunkRes,
	trunkBaseSettingsId,
	edgeGroupId string) string {
	return fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_trunk" "%s" {
		trunk_base_settings_id = %s
		edge_group_id = %s
	}
	`, trunkRes, trunkBaseSettingsId, edgeGroupId)
}
