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
- [GET /api/v2/telephony/providers/edges/extensionpools/{extensionPoolId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-telephony-providers-edges-extensionpools--extensionPoolId-)

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

- `acd_auto_answer` (Boolean) Enable ACD auto-answer.
- `addresses` (Block List) The address settings for this user. If not set, this resource will not manage addresses. (see [below for nested schema](#nestedblock--addresses))
- `certifications` (Set of String) Certifications for this user. If not set, this resource will not manage certifications.
- `department` (String) User's department.
- `division_id` (String) The division to which this user will belong. If not set, the home division will be used.
- `employer_info` (Block List) The employer info for this user. If not set, this resource will not manage employer info. (see [below for nested schema](#nestedblock--employer_info))
- `locations` (Block Set) The user placement at each site location. If not set, this resource will not manage user locations. (see [below for nested schema](#nestedblock--locations))
- `manager` (String) User ID of this user's manager.
- `password` (String, Sensitive) User's password. If specified, this is only set on user create.
- `profile_skills` (Set of String) Profile skills for this user. If not set, this resource will not manage profile skills.
- `routing_languages` (Block Set) Languages and proficiencies for this user. If not set, this resource will not manage user languages. (see [below for nested schema](#nestedblock--routing_languages))
- `routing_skills` (Block Set) Skills and proficiencies for this user. If not set, this resource will not manage user skills. (see [below for nested schema](#nestedblock--routing_skills))
- `routing_utilization` (Block List) The routing utilization settings for this user. If empty list, the org default settings are used. If not set, this resource will not manage the users's utilization settings. (see [below for nested schema](#nestedblock--routing_utilization))
- `state` (String) User's state (active | inactive). Default is 'active'.
- `title` (String) User's title.
- `voicemail_userpolicies` (Block List) User's voicemail policies. If not set, default user policies will be applied. (see [below for nested schema](#nestedblock--voicemail_userpolicies))

### Read-Only

- `id` (String) The ID of the user.

<a id="nestedblock--addresses"></a>
### Nested Schema for `addresses`

Optional:

- `other_emails` (Block Set) Other Email addresses for this user. (see [below for nested schema](#nestedblock--addresses--other_emails))
- `phone_numbers` (Block Set) Phone number addresses for this user. (see [below for nested schema](#nestedblock--addresses--phone_numbers))

<a id="nestedblock--addresses--other_emails"></a>
### Nested Schema for `addresses.other_emails`

Required:

- `address` (String) Email address.

Optional:

- `type` (String) Type of email address (WORK | HOME).


<a id="nestedblock--addresses--phone_numbers"></a>
### Nested Schema for `addresses.phone_numbers`

Optional:

- `extension` (String) Phone number extension
- `extension_pool_id` (String) Id of the extension pool which contains this extension.
- `media_type` (String) Media type of phone number (SMS | PHONE).
- `number` (String) Phone number. Phone number must be in an E.164 number format.
- `type` (String) Type of number (WORK | WORK2 | WORK3 | WORK4 | HOME | MOBILE | OTHER).



<a id="nestedblock--employer_info"></a>
### Nested Schema for `employer_info`

Optional:

- `date_hire` (String) Hiring date. Dates must be an ISO-8601 string. For example: yyyy-MM-dd.
- `employee_id` (String) Employee ID.
- `employee_type` (String) Employee type (Full-time | Part-time | Contractor).
- `official_name` (String) User's official name.


<a id="nestedblock--locations"></a>
### Nested Schema for `locations`

Required:

- `location_id` (String) ID of location.

Optional:

- `notes` (String) Optional description on the user's location.


<a id="nestedblock--routing_languages"></a>
### Nested Schema for `routing_languages`

Required:

- `language_id` (String) ID of routing language.
- `proficiency` (Number) Proficiency is a rating from 0 to 5 on how competent an agent is for a particular language. It is used when a queue is set to 'Best available language' mode to allow acd interactions to target agents with higher proficiency ratings.


<a id="nestedblock--routing_skills"></a>
### Nested Schema for `routing_skills`

Required:

- `proficiency` (Number) Rating from 0.0 to 5.0 on how competent an agent is for a particular skill. It is used when a queue is set to 'Best available skills' mode to allow acd interactions to target agents with higher proficiency ratings.
- `skill_id` (String) ID of routing skill.


<a id="nestedblock--routing_utilization"></a>
### Nested Schema for `routing_utilization`

Optional:

- `call` (Block List) Call media settings. If not set, this reverts to the default media type settings. (see [below for nested schema](#nestedblock--routing_utilization--call))
- `callback` (Block List) Callback media settings. If not set, this reverts to the default media type settings. (see [below for nested schema](#nestedblock--routing_utilization--callback))
- `chat` (Block List) Chat media settings. If not set, this reverts to the default media type settings. (see [below for nested schema](#nestedblock--routing_utilization--chat))
- `email` (Block List) Email media settings. If not set, this reverts to the default media type settings. (see [below for nested schema](#nestedblock--routing_utilization--email))
- `label_utilizations` (Block List) Label utilization settings. If not set, default label settings will be applied. This is in PREVIEW and should not be used unless the feature is available to your organization. (see [below for nested schema](#nestedblock--routing_utilization--label_utilizations))
- `message` (Block List) Message media settings. If not set, this reverts to the default media type settings. (see [below for nested schema](#nestedblock--routing_utilization--message))

<a id="nestedblock--routing_utilization--call"></a>
### Nested Schema for `routing_utilization.call`

Required:

- `maximum_capacity` (Number) Maximum capacity of conversations of this media type. Value must be between 0 and 25.

Optional:

- `include_non_acd` (Boolean) Block this media type when on a non-ACD conversation.
- `interruptible_media_types` (Set of String) Set of other media types that can interrupt this media type (call | callback | chat | email | message).


<a id="nestedblock--routing_utilization--callback"></a>
### Nested Schema for `routing_utilization.callback`

Required:

- `maximum_capacity` (Number) Maximum capacity of conversations of this media type. Value must be between 0 and 25.

Optional:

- `include_non_acd` (Boolean) Block this media type when on a non-ACD conversation.
- `interruptible_media_types` (Set of String) Set of other media types that can interrupt this media type.


<a id="nestedblock--routing_utilization--chat"></a>
### Nested Schema for `routing_utilization.chat`

Required:

- `maximum_capacity` (Number) Maximum capacity of conversations of this media type. Value must be between 0 and 25.

Optional:

- `include_non_acd` (Boolean) Block this media type when on a non-ACD conversation.
- `interruptible_media_types` (Set of String) Set of other media types that can interrupt this media type.


<a id="nestedblock--routing_utilization--email"></a>
### Nested Schema for `routing_utilization.email`

Required:

- `maximum_capacity` (Number) Maximum capacity of conversations of this media type. Value must be between 0 and 25.

Optional:

- `include_non_acd` (Boolean) Block this media type when on a non-ACD conversation.
- `interruptible_media_types` (Set of String) Set of other media types that can interrupt this media type.


<a id="nestedblock--routing_utilization--label_utilizations"></a>
### Nested Schema for `routing_utilization.label_utilizations`

Required:

- `label_id` (String) Id of the label being configured.
- `maximum_capacity` (Number) Maximum capacity of conversations with this label. Value must be between 0 and 25.

Optional:

- `interrupting_label_ids` (Set of String) Set of other labels that can interrupt this label.


<a id="nestedblock--routing_utilization--message"></a>
### Nested Schema for `routing_utilization.message`

Required:

- `maximum_capacity` (Number) Maximum capacity of conversations of this media type. Value must be between 0 and 25.

Optional:

- `include_non_acd` (Boolean) Block this media type when on a non-ACD conversation.
- `interruptible_media_types` (Set of String) Set of other media types that can interrupt this media type.



<a id="nestedblock--voicemail_userpolicies"></a>
### Nested Schema for `voicemail_userpolicies`

Optional:

- `alert_timeout_seconds` (Number) The number of seconds to ring the user's phone before a call is transferred to voicemail.
- `send_email_notifications` (Boolean) Whether email notifications are sent to the user when a new voicemail is received.

