resource "genesyscloud_intent_category" "example_intent_category" {
  name = "Example Intent Category"
}

resource "genesyscloud_customer_intent" "example_customer_intent" {
  name        = "Example Customer Intent"
  description = "Example customer intent description"
  expiry_time = 24
  category_id = genesyscloud_intent_category.example_intent_category.id
}

# Example with source intents
resource "genesyscloud_customer_intent" "example_with_source_intents" {
  name        = "Example Customer Intent with Source Intents"
  description = "Customer intent with mapped source intents"
  expiry_time = 48
  category_id = genesyscloud_intent_category.example_intent_category.id

  source_intents {
    source_intent_id   = "bot-intent-id-1"
    source_intent_name = "Bot Intent 1"
    source_type        = "Bot"
    source_id          = "bot-id-1"
    source_name        = "My Bot"
  }

  source_intents {
    source_intent_id   = "copilot-intent-id-2"
    source_intent_name = "Copilot Intent 2"
    source_type        = "Copilot"
    source_id          = "copilot-id-1"
    source_name        = "My Copilot"
  }
}
