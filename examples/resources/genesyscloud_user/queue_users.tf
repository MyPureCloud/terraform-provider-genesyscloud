resource "genesyscloud_user" "queue_manager" {
  email      = "manny_gerr${random_uuid.uuid.result}@example.com"
  name       = "Manny Gerr"
  title      = "Senior Manager"
  department = "Support"
  state      = "active"
}

resource "genesyscloud_user" "queue_user1" {
  email           = "queue1${random_uuid.uuid.result}@example.com"
  name            = "Queue One"
  password        = "initialP@ssW0rd"
  division_id     = data.genesyscloud_auth_division_home.home.id
  state           = "active"
  department      = "Support"
  title           = "Agent"
  manager         = genesyscloud_user.queue_manager.id
  acd_auto_answer = true
  profile_skills  = ["Knitting", "Hockey"]
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

resource "genesyscloud_user" "queue_user2" {
  email           = "queue2${random_uuid.uuid.result}@example.com"
  name            = "Queue Two"
  password        = "initialP@ssW0rd"
  division_id     = data.genesyscloud_auth_division_home.home.id
  state           = "active"
  department      = "Support"
  title           = "Agent"
  manager         = genesyscloud_user.queue_manager.id
  acd_auto_answer = true
  profile_skills  = ["Cooking", "Electrician"]
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

resource "genesyscloud_user" "queue_user3" {
  email           = "queue3${random_uuid.uuid.result}@example.com"
  name            = "Queue Three"
  password        = "initialP@ssW0rd"
  division_id     = data.genesyscloud_auth_division_home.home.id
  state           = "active"
  department      = "Support"
  title           = "Agent"
  manager         = genesyscloud_user.queue_manager.id
  acd_auto_answer = true
  profile_skills  = ["Dog Walking", "Barista"]
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
