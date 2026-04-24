resource "genesyscloud_users_rules" "example_users_rules" {
  name                    = "Example name"
  description             = "Example description"
  type                    = "Learning"
  criteria {
    operator = "Or"
    group {
      operator  = "And"
      container = "ManagementUnit"
      values {
        context_id = "b3e8c85e-1ae3-4c41-9914-d08d96a12e78"
        ids        = ["bd80395d-686e-4776-a9d5-a01f36a8cf87", "95981a5f-c4a9-4cc9-9991-51b822399421"]
      }
      values {
        context_id = "2cb0c813-c92e-433a-ae33-2a0d0d392c54"
        ids        = ["41be2fea-c455-49ac-88e7-1a09f102fec6"]
      }
    }
    group {
      operator  = "Not"
      container = "User"
      values {
        ids = ["b1261658-4535-4a9d-9393-eb67b0dfc2a5"]
      }
    }
  }
  criteria {
    operator = "Or"
    group {
      operator  = "And"
      container = "Division"
      values {
        ids = ["f87b8ef3-6e83-41e5-a153-97f79eeffff3"]
      }
    }
  }
}