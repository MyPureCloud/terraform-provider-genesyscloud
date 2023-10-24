resource "genesyscloud_integration_custom_auth_action" "example-custom-auth-action" {
  integration_id = genesyscloud_integration.example_integ.id
  name           = "Example Custom Auth Action"
  config_request {
    # Use '$${' to indicate a literal '${' in template strings. Otherwise Terraform will attempt to interpolate the string
    # See https://www.terraform.io/docs/language/expressions/strings.html#escape-sequences
    request_url_template = "$${credentials.loginUrl}"
    request_type         = "POST"
    request_template     = "grant_type=client_credentials"
    headers = {
      Authorization = "Basic $encoding.base64(\"$${credentials.clientId}:$${credentials.clientSecret}\")"
      Content-Type  = "application/x-www-form-urlencoded"
    }
  }
  config_response {
    translation_map = {
      tokenValue = "$.token"
    }
    success_template = "{ \"token\": $${tokenValue} }"
  }
}
