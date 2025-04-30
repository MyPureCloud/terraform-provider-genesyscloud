locals {
  dependencies = [
    "../../common/random_uuid.tf",
    "../../data-sources/genesyscloud_auth_division_home/data-source.tf",
    "../genesyscloud_routing_skill/resource.tf",
    "../genesyscloud_routing_language/resource.tf",
    "../genesyscloud_location/resource.tf",
    "../genesyscloud_routing_utilization_label/resource.tf"
  ]
}
