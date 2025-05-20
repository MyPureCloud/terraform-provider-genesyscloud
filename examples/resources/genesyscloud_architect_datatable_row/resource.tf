resource "genesyscloud_architect_datatable_row" "john-smith" {
  datatable_id = genesyscloud_architect_datatable.customers.id
  key_value    = "johnsmith@example.com"
  properties_json = jsonencode({
    "identifier" = 2749
    "address"    = "123 Main Street"
    "vip"        = true
  })
}
