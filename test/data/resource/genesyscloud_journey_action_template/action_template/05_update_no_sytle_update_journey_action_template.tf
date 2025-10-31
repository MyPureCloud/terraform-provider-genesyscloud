resource "genesyscloud_journey_action_template" "terraform_test_-TEST-CASE-" {
  name        = "terraform_test_-TEST-CASE-"
  description = "Text and image content offer"
  media_type  = "contentOffer"
  state       = "Active"
  content_offer {
    image_url      = "https://api-cdn.inindca.com/uploads/v1/publicassets/images/d460a77c-9870-404f-9711-4be1cc247b66/d7c29719-095b-45d3-9ceb-f1368bcfcf3f.dragon.png"
    display_mode   = "Modal"
    layout_mode    = "RightText"
    title          = "Dragon!"
    headline       = "Save 100%"
    body           = "Book now and add discount code 123456 at checkout to save 100%"
    image_alt_text = "Image updated"
    call_to_action {
      text   = "Dragon!"
      url    = "https://www.genesys.com"
      target = "Self"
    }
  }
}
