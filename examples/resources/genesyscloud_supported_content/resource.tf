resource "genesyscloud_supported_content" "supported_content" {
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