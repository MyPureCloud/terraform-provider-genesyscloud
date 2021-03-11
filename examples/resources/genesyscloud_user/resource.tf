resource "genesyscloud_user" "test_user" {
  email       = "john@example.com"
  name        = "John Doe"
  password    = "initial-password"
  division_id = genesyscloud_auth_division.home.id
  state       = "active"
  department  = "Development"
  title       = "Senior Director"
  manager     = genesyscloud_user.test-user-manager.id
  addresses {
    other_emails {
      address = "john@gmail.com"
      type    = "HOME"
    }
    phone_numbers {
      number     = "3174181234"
      media_type = "PHONE"
      type       = "MOBILE"
    }
  }
  routing_skills {
    skill_id    = genesyscloud_routing_skill.test-skill.id
    proficiency = 4.5
  }
  roles {
    role_id      = genesyscloud_auth_role.custom-role.id
    division_ids = [genesyscloud_auth_division.marketing.id]
  }
}
