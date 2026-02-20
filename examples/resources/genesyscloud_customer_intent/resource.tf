resource "genesyscloud_intent_category" "example_intent_category" {
  name = "Example Intent Category"
}

resource "genesyscloud_customer_intent" "example_customer_intent" {
  name        = "Example Customer Intent"
  description = "Example customer intent description"
  expiry_time = 24
  category_id = genesyscloud_intent_category.example_intent_category.id
}
