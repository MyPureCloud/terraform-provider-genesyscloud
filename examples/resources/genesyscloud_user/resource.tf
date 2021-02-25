resource "genesyscloud_user" "test_user" {
  email       = "john@example.com"
  name        = "John Doe"
  password    = "initial-password"
  division_id = "505e1036-6f04-405c-a630-de94a8ad2eb8"
  state       = "active"
  department  = "Development"
  title       = "Senior Director"
  manager     = genesyscloud_user.test-user-manager.id
  other_emails {
    address = "john@gmail.com"
    type    = "HOME"
  }
  phone_numbers {
    number     = "3174181234"
    media_type = "PHONE"
    type       = "MOBILE"
  }
  routing_skills {
    skill_id    = genesyscloud_routing_skill.test-skill.id
    proficiency = 4.5
  }
  roles {
    role_id      = genesyscloud_auth_role.custom-role.id
    division_ids = ["505e1036-6f04-405c-a630-de94a8ad2eb8"]
  }
}
