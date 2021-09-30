resource "genesyscloud_architect_datatable" "customers" {
  name        = "Customers"
  division_id = genesyscloud_auth_division.home.id
  description = "Table of Customers"
  properties {
    name  = "key"
    type  = "string"
    title = "Email"
  }
  properties {
    name  = "identifier"
    type  = "integer"
    title = "Customer Identifier"
  }
  properties {
    name    = "deleted"
    type    = "boolean"
    title   = "Is Deleted"
    default = "false"
  }
}
