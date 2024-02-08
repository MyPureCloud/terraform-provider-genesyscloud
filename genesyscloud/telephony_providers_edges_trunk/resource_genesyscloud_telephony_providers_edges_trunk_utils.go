package telephony_providers_edges_trunk

import "fmt"

func generateTrunkDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceName, resourceID, name, dependsOnResource)
}
func generateTrunk(
	trunkRes,
	trunkBaseSettingsId,
	edgeGroupId string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		trunk_base_settings_id = %s
		edge_group_id = %s
	}
	`, resourceName, trunkRes, trunkBaseSettingsId, edgeGroupId)
}
