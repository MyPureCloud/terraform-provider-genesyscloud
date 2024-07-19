resource "genesyscloud_conversations_messaging_supportedcontent" "supported_content" {
  name = "test supported_content"
  media_types {
    allow {
      inbound {
        type = "image/*"
      }
      outbound {
        type = "video/mpeg"
      }
    }
  }
}