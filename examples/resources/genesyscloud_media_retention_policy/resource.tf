resource "genesyscloud_media_retention_policy" "policy1" {
    name = "Example Policy"
    description = "a media retention policy"
    enabled = true  
    media_policies {
        call_policy {
            actions {   
                retain_recording = true
                assign_surveys{
                    survey_form{
                        name = "ronan test"
                        context_id = "714b523e-6fc9-44fa-9eff-cd5f6a13430a"
                    }
                    flow{
                        id = module.survey_flow.flow_id
						name = "SendSurvey"
						self_uri = format("/api/v2/flows/%s", module.survey_flow.flow_id)
                    }
                    sending_domain = "surveys.mypurecloud.com"
                }
            }
            conditions {
                for_queues{
                    id = "c3275a86-3a0c-4a06-beaf-ca1bf096b7b5"
                    name = "Transcription Queue"
                    division {
                        name = "New Home"
                    }
                }
            }
        }
    }
}