resource "genesyscloud_script" "email" {
  script_name       = "Example script name ${random_uuid.uuid.result}"
  filepath          = "${local.working_dir.script}/email.script.json"
  file_content_hash = filesha256("${local.working_dir.script}/email.script.json")
  substitutions = {
    /* Inside the script file, "{{foo}}" will be replaced with "bar" */
    foo = "bar"
  }
}
resource "random_uuid" "uuid" {
}
