# Basic Script with Local File
resource "genesyscloud_script" "example_script" {
  script_name       = "Example script name ${random_uuid.uuid.result}"
  filepath          = "${local.working_dir.script}/email.script.json"
  file_content_hash = filesha256("${local.working_dir.script}/email.script.json")
  substitutions = {
    /* Inside the script file, "{{foo}}" will be replaced with "bar" */
    foo = "bar"
  }
}

# Script with S3 File
resource "genesyscloud_script" "s3_script" {
  script_name       = "S3_Script_Example"
  filepath          = "s3://my-scripts-bucket/scripts/email-flow.json"
  file_content_hash = filesha256("s3://my-scripts-bucket/scripts/email-flow.json")
  substitutions = {
    company_name = "Acme Corp"
    support_email = "support@acme.com"
  }
}

# Script with Division Assignment
resource "genesyscloud_script" "division_script" {
  script_name       = "Division_Specific_Script"
  filepath          = "s3://scripts-bucket/division/call-handling.json"
  file_content_hash = filesha256("s3://scripts-bucket/division/call-handling.json")
  division_id       = "division-id-here"
  substitutions = {
    queue_name = "Sales Queue"
    timeout_seconds = "30"
  }
}

# Mixed Local and S3 Scripts
resource "genesyscloud_script" "local_script" {
  script_name       = "Local_Script"
  filepath          = "./scripts/local-call-flow.json"
  file_content_hash = filesha256("./scripts/local-call-flow.json")
  substitutions = {
    local_setting = "local_value"
  }
}
