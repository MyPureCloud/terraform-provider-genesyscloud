resource "genesyscloud_user" "director" {
  email = "dirk_rector${random_uuid.uuid.result}@example.com"
  name  = "Dirk Rector"
  title = "Senior Director"
  state = "active"
}


resource "genesyscloud_user" "evaluator_user" {
  email           = "eve_al_yuator${random_uuid.uuid.result}@example.com"
  name            = "Eve Al-Yuator"
  password        = "initialP@ssW0rd"
  division_id     = data.genesyscloud_auth_division_home.home.id
  state           = "active"
  department      = "Sales"
  title           = "Supervisor"
  manager         = genesyscloud_user.director.id
  acd_auto_answer = true
  addresses {
    other_emails {
      address = "eay@gmail.com"
      type    = "HOME"
    }
    phone_numbers {
      number     = "+13174182345"
      media_type = "PHONE"
      type       = "MOBILE"
    }
  }
  locations {
    location_id = genesyscloud_location.hq.id
    notes       = "Office 207"
  }
}

data "genesyscloud_auth_role" "quality_evaluator" {
  name = "Quality Evaluator"
}

resource "genesyscloud_user_roles" "evaluator_user_role" {
  user_id = genesyscloud_user.evaluator_user.id
  roles {
    role_id      = data.genesyscloud_auth_role.quality_evaluator.id
    division_ids = [data.genesyscloud_auth_division_home.home.id]
  }
}

resource "genesyscloud_user" "quality_admin" {
  email           = "q_litle${random_uuid.uuid.result}@example.com"
  name            = "Q Little"
  password        = "initialP@ssW0rd"
  division_id     = data.genesyscloud_auth_division_home.home.id
  state           = "active"
  department      = "Quality"
  title           = "Manager"
  manager         = genesyscloud_user.director.id
  acd_auto_answer = true
  addresses {
    other_emails {
      address = "qlittle@gmail.com"
      type    = "HOME"
    }
    phone_numbers {
      number     = "+13174182342"
      media_type = "PHONE"
      type       = "MOBILE"
    }
  }
  locations {
    location_id = genesyscloud_location.hq.id
    notes       = "Office 242"
  }
}

data "genesyscloud_auth_role" "quality_admin" {
  name = "Quality Administrator"
}

resource "genesyscloud_user_roles" "quality_admin_role" {
  user_id = genesyscloud_user.quality_admin.id
  roles {
    role_id      = data.genesyscloud_auth_role.quality_admin.id
    division_ids = [data.genesyscloud_auth_division_home.home.id]
  }
}
