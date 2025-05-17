resource "genesyscloud_group" "bullseye_rings_group1" {
  name          = "Bullseye Rings Group 1"
  description   = "Group for Bullseye Rings"
  type          = "official"
  visibility    = "public"
  rules_visible = true
  owner_ids     = [genesyscloud_user.queue_manager.id]
  member_ids    = [genesyscloud_user.queue_user1.id]
  roles_enabled = true
  calls_enabled = false
}

resource "genesyscloud_group" "bullseye_rings_group2" {
  name          = "Bullseye Rings Group 2"
  description   = "Group 2 for Bullseye Rings"
  type          = "official"
  visibility    = "public"
  rules_visible = true
  owner_ids     = [genesyscloud_user.queue_manager.id]
  member_ids    = [genesyscloud_user.queue_user2.id]
  roles_enabled = true
  calls_enabled = false
}
