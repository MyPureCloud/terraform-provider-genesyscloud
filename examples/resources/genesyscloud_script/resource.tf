resource "genesyscloud_script" "script" {
  script_name       = "Example script name ${random_uuid.uuid.result}"
  filepath          = "the script file path"
  file_content_hash = filesha256("the script file path")
  substitutions = {
    /* Inside the script file, "{{foo}}" will be replaced with "bar" */
    foo = "bar"
  }
}