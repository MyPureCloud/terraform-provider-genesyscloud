resource "genesyscloud_journey_action_template" "terraform_test_-TEST-CASE-" {
  name        = "terraform_test_-TEST-CASE-updated"
  description = "Text and image content offer_updated"
  media_type  = "contentOffer"
  state       = "Active"
  content_offer {
    image_url    = "https://api-cdn.inindca.com/uploads/v1/publicassets/images/d460a77c-9870-404f-9711-4be1cc247b66/d7c29719-095b-45d3-9ceb-f1368bcfcf3f.dragon.png"
    display_mode = "Modal"
    layout_mode  = "RightText"
    title        = "Dragon!"
    headline     = "Save 100%"
    body         = "Book now and add discount code 123456 at checkout to save 100%"
    call_to_action {
      text       = "Dragon!"
      url        = "https://www.genesys.com"
      target = "Self"
    }
    style {
      position {
        top    = "20px"
        bottom = "10px"
        left   = "10px"
        right  = "20px"
      }
      offer {
        padding          = "0px"
        background_color = "#33383d"
      }
      close_button {
        color   = "#f01000"
        opacity = 0.48
      }
      cta_button {
        color            = "#fdfdfd"
        font             = "inherit"
        font_size        = "12pt"
        text_align       = "Center"
        background_color = "#5081e1"
      }
      title {
        color      = "#fdfdfd"
        font       = "inherit"
        font_size  = "18pt"
        text_align = "Center"
      }
      headline {
        color      = "#fdfdfd"
        font       = "inherit"
        font_size  = "14pt"
        text_align = "Center"
      }
      body {
        color      = "#fdfdfd"
        font       = "inherit"
        font_size  = "8pt"
        text_align = "Center"
      }
    }
  }
}

#resource "genesyscloud_journey_action_template" "terraform_test_-TEST-CASE-" {
#  name        = "terraform_test_jay_-TEST-CASE-_updated"
#  description = "Text and image content offer_updated"
#  media_type  = "contentOffer"
#  state       = "Active"
#  content_offer {
#    image_url    = "https://i.pinimg.com/originals/9f/2c/9f/9f2c9f8f986d2c28f3c7f7d52758b231.jpg"
#    display_mode = "Modal"
#    layout_mode  = "ImageOnly"
#    title        = "Exclusive offer!vv"
#    headline     = "Save 10%fff"
#    body         = "Book now and add discount code 123456 at checkout to save 10%fff"
#    call_to_action {
#      text       = "Book now!ff"
#      url        = "https://www.genesys.comff"
#      target_url = "Blank"
#    }
#    style {
#      offer {
#        padding          = "0px"
#        background_color = "#33383d"
#      }
#      close_button {
#        color   = "#f01000"
#        opacity = 0.48
#      }
#      cta_button {
#        color            = "#fdfdfd"
#        font             = "inherit"
#        font_size        = "12pt"
#        text_align       = "Center"
#        background_color = "#5081e1"
#      }
#      title {
#        color      = "#fdfdfd"
#        font       = "inherit"
#        font_size  = "18pt"
#        text_align = "Center"
#      }
#      headline {
#        color      = "#fdfdfd"
#        font       = "inherit"
#        font_size  = "14pt"
#        text_align = "Center"
#      }
#      body {
#        color      = "#fdfdfd"
#        font       = "inherit"
#        font_size  = "8pt"
#        text_align = "Center"
#      }
#    }
#  }
#}