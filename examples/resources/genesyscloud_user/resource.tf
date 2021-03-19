resource "genesyscloud_user" "test_user" {
  email           = "john@example.com"
  name            = "John Doe"
  password        = "initial-password"
  division_id     = genesyscloud_auth_division.home.id
  state           = "active"
  department      = "Development"
  title           = "Senior Director"
  manager         = genesyscloud_user.test-user-manager.id
  acd_auto_answer = true
  profile_skills  = ["Java", "Go"]
  certifications  = ["Certified Developer"]
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
  routing_languages {
    skill_id    = genesyscloud_routing_language.english.id
    proficiency = 4
  }
  locations {
    location_id = genesyscloud_location.main-site.id
    notes       = "Office 201"
  }
  employer_info {
    official_name = "Jonathon Doe"
    employee_id   = "12345"
    employee_type = "Full-time"
    date_hire     = "2021-03-18"
  }
}
