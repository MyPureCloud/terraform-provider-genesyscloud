resource "genesyscloud_conversations_messaging_settings" "example_settings" {
  name = "Sample Messaging Settings"
  content {
    story {
      mention {
        inbound = "Enabled"
      }
      reply {
        inbound = "Enabled"
      }
    }
  }
  event {
    typing {
      on {
        inbound  = "Enabled"
        outbound = "Enabled"
      }
    }
  }
}
