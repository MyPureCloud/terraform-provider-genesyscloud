---
page_title: "genesyscloud_user Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud User.
  Export block label: "{email}"
---
# genesyscloud_user (Resource)

Genesys Cloud User.

Export block label: "{email}"

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

- [POST /api/v2/users](https://developer.mypurecloud.com/api/rest/v2/users/#post-api-v2-users)
- [GET /api/v2/users/{userId}](https://developer.mypurecloud.com/api/rest/v2/users/#get-api-v2-users--userId-)
- [PATCH /api/v2/users/{userId}](https://developer.mypurecloud.com/api/rest/v2/users/#patch-api-v2-users--userId-)
- [DELETE /api/v2/users/{userId}](https://developer.mypurecloud.com/api/rest/v2/users/#delete-api-v2-users--userId-)
- [POST /api/v2/users/search](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-users-search)
- [PUT /api/v2/users/{userId}/routingskills/bulk](https://developer.mypurecloud.com/api/rest/v2/users/#put-api-v2-users--userId--routingskills-bulk)
- [DELETE /api/v2/users/{userId}/routinglanguages/{languageId}](https://developer.mypurecloud.com/api/rest/v2/users/#delete-api-v2-users--userId--routinglanguages--languageId-)
- [PATCH /api/v2/users/{userId}/routinglanguages/bulk](https://developer.mypurecloud.com/api/rest/v2/users/#patch-api-v2-users--userId--routinglanguages-bulk)
- [GET /api/v2/users/{userId}/routinglanguages](https://developer.mypurecloud.com/api/rest/v2/users/#get-api-v2-users--userId--routinglanguages)
- [PUT /api/v2/users/{userId}/profileskills](https://developer.mypurecloud.com/api/rest/v2/users/#put-api-v2-users--userId--profileskills)
- [POST /api/v2/users/{userId}/password](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-users--userId--password)
- [GET /api/v2/routing/users/{userId}/utilization](https://developer.mypurecloud.com/api/rest/v2/users/#get-api-v2-routing-users--userId--utilization)
- [PUT /api/v2/routing/users/{userId}/utilization](https://developer.mypurecloud.com/api/rest/v2/users/#put-api-v2-routing-users--userId--utilization)
- [DELETE /api/v2/routing/users/{userId}/utilization](https://developer.mypurecloud.com/api/rest/v2/users/#delete-api-v2-routing-users--userId--utilization)
- [GET /api/v2/voicemail/userpolicies/{userId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-voicemail-userpolicies--userId-)
- [PATCH /api/v2/voicemail/userpolicies/{userId}](https://developer.genesys.cloud/devapps/api-explorer#patch-api-v2-voicemail-userpolicies--userId-)
- [GET /api/v2/telephony/providers/edges/extensionpools](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-telephony-providers-edges-extensionpools)

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `email` (String) User's primary email and username.
- `name` (String) User's full name.

### Optional

- `acd_auto_answer` (Boolean) Enable ACD auto-answer. Defaults to `false`.
- `addresses` (List of Object) The address settings for this user. If not set, this resource will not manage addresses. (see [below for nested schema](#nestedatt--addresses))
- `certifications` (Set of String) Certifications for this user. If not set, this resource will not manage certifications.
- `department` (String) User's department.
- `division_id` (String) The division to which this user will belong. If not set, the home division will be used.
- `employer_info` (List of Object) The employer info for this user. If not set, this resource will not manage employer info. (see [below for nested schema](#nestedatt--employer_info))
- `locations` (Set of Object) The user placement at each site location. If not set, this resource will not manage user locations. (see [below for nested schema](#nestedatt--locations))
- `manager` (String) User ID of this user's manager.
- `password` (String, Sensitive) User's password. If specified, this is only set on user create.
- `profile_skills` (Set of String) Profile skills for this user. If not set, this resource will not manage profile skills.
- `routing_languages` (Set of Object) Languages and proficiencies for this user. If not set, this resource will not manage user languages. (see [below for nested schema](#nestedatt--routing_languages))
- `routing_skills` (Set of Object) Skills and proficiencies for this user. If not set, this resource will not manage user skills. (see [below for nested schema](#nestedatt--routing_skills))
- `routing_utilization` (List of Object) The routing utilization settings for this user. If empty list, the org default settings are used. If not set, this resource will not manage the users's utilization settings. (see [below for nested schema](#nestedatt--routing_utilization))
- `state` (String) User's state (active | inactive). Default is 'active'. Defaults to `active`.
- `title` (String) User's title.
- `voicemail_userpolicies` (Block List, Max: 1) User's voicemail policies. If not set, default user policies will be applied. (see [below for nested schema](#nestedblock--voicemail_userpolicies))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedatt--addresses"></a>
### Nested Schema for `addresses`

Optional:

- `other_emails` (Set of Object) (see [below for nested schema](#nestedobjatt--addresses--other_emails))
- `phone_numbers` (Set of Object) (see [below for nested schema](#nestedobjatt--addresses--phone_numbers))

<a id="nestedobjatt--addresses--other_emails"></a>
### Nested Schema for `addresses.other_emails`

Optional:

- `address` (String)
- `type` (String)


<a id="nestedobjatt--addresses--phone_numbers"></a>
### Nested Schema for `addresses.phone_numbers`

Optional:

- `extension` (String)
- `extension_pool_id` (String)
- `media_type` (String)
- `number` (String)
- `type` (String)



<a id="nestedatt--employer_info"></a>
### Nested Schema for `employer_info`

Optional:

- `date_hire` (String)
- `employee_id` (String)
- `employee_type` (String)
- `official_name` (String)


<a id="nestedatt--locations"></a>
### Nested Schema for `locations`

Optional:

- `location_id` (String)
- `notes` (String)


<a id="nestedatt--routing_languages"></a>
### Nested Schema for `routing_languages`

Optional:

- `language_id` (String)
- `proficiency` (Number)


<a id="nestedatt--routing_skills"></a>
### Nested Schema for `routing_skills`

Optional:

- `proficiency` (Number)
- `skill_id` (String)


<a id="nestedatt--routing_utilization"></a>
### Nested Schema for `routing_utilization`

Optional:

- `call` (List of Object) (see [below for nested schema](#nestedobjatt--routing_utilization--call))
- `callback` (List of Object) (see [below for nested schema](#nestedobjatt--routing_utilization--callback))
- `chat` (List of Object) (see [below for nested schema](#nestedobjatt--routing_utilization--chat))
- `email` (List of Object) (see [below for nested schema](#nestedobjatt--routing_utilization--email))
- `label_utilizations` (List of Object) (see [below for nested schema](#nestedobjatt--routing_utilization--label_utilizations))
- `message` (List of Object) (see [below for nested schema](#nestedobjatt--routing_utilization--message))

<a id="nestedobjatt--routing_utilization--call"></a>
### Nested Schema for `routing_utilization.call`

Optional:

- `include_non_acd` (Boolean)
- `interruptible_media_types` (Set of String)
- `maximum_capacity` (Number)


<a id="nestedobjatt--routing_utilization--callback"></a>
### Nested Schema for `routing_utilization.callback`

Optional:

- `include_non_acd` (Boolean)
- `interruptible_media_types` (Set of String)
- `maximum_capacity` (Number)


<a id="nestedobjatt--routing_utilization--chat"></a>
### Nested Schema for `routing_utilization.chat`

Optional:

- `include_non_acd` (Boolean)
- `interruptible_media_types` (Set of String)
- `maximum_capacity` (Number)


<a id="nestedobjatt--routing_utilization--email"></a>
### Nested Schema for `routing_utilization.email`

Optional:

- `include_non_acd` (Boolean)
- `interruptible_media_types` (Set of String)
- `maximum_capacity` (Number)


<a id="nestedobjatt--routing_utilization--label_utilizations"></a>
### Nested Schema for `routing_utilization.label_utilizations`

Optional:

- `interrupting_label_ids` (Set of String)
- `label_id` (String)
- `maximum_capacity` (Number)


<a id="nestedobjatt--routing_utilization--message"></a>
### Nested Schema for `routing_utilization.message`

Optional:

- `include_non_acd` (Boolean)
- `interruptible_media_types` (Set of String)
- `maximum_capacity` (Number)



<a id="nestedblock--voicemail_userpolicies"></a>
### Nested Schema for `voicemail_userpolicies`

Optional:

- `alert_timeout_seconds` (Number) The number of seconds to ring the user's phone before a call is transferred to voicemail.
- `send_email_notifications` (Boolean) Whether email notifications are sent to the user when a new voicemail is received.

