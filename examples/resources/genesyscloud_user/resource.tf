resource "genesyscloud_user" "example_user" {
  email           = "johnny${random_uuid.uuid.result}@example.com"
  name            = "Johnny Doe"
  password        = "initialP@ssW0rd"
  division_id     = data.genesyscloud_auth_division_home.home.id
  state           = "active"
  department      = "Development"
  title           = "Senior Director"
  manager         = genesyscloud_user.example_user2.id
  acd_auto_answer = true
  profile_skills  = ["Java", "Go"]
  certifications  = ["Certified Developer"]
  addresses {
    other_emails {
      address = "john@gmail.com"
      type    = "HOME"
    }
    phone_numbers {
      number     = "+13174181234"
      media_type = "PHONE"
      type       = "MOBILE"
    }
  }
  routing_skills {
    skill_id    = genesyscloud_routing_skill.example_skill.id
    proficiency = 4.5
  }
  routing_languages {
    language_id = genesyscloud_routing_language.english.id
    proficiency = 4
  }
  locations {
    location_id = genesyscloud_location.hq.id
    notes       = "Office 201"
  }
  employer_info {
    official_name = "Jonathon Doe"
    employee_id   = "12345"
    employee_type = "Full-time"
    date_hire     = "2021-03-18"
  }
  routing_utilization {
    call {
      maximum_capacity = 1
      include_non_acd  = true
    }
    callback {
      maximum_capacity          = 2
      include_non_acd           = false
      interruptible_media_types = ["call", "email"]
    }
    chat {
      maximum_capacity          = 3
      include_non_acd           = false
      interruptible_media_types = ["call"]
    }
    email {
      maximum_capacity          = 2
      include_non_acd           = false
      interruptible_media_types = ["call", "chat"]
    }
    message {
      maximum_capacity          = 4
      include_non_acd           = false
      interruptible_media_types = ["call", "chat"]
    }
    label_utilizations {
      label_id         = genesyscloud_routing_utilization_label.red_label.id
      maximum_capacity = 4
    }
    label_utilizations {
      label_id               = genesyscloud_routing_utilization_label.blue_label.id
      maximum_capacity       = 4
      interrupting_label_ids = [genesyscloud_routing_utilization_label.red_label.id]
    }
  }
}

resource "genesyscloud_user" "example_user2" {
  email = "bobby${random_uuid.uuid.result}@example.com"
  name  = "Bobby Drop Tables"
  title = "CEO"
  state = "active"
}
