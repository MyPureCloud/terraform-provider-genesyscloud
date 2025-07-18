resource "genesyscloud_flow" "inbound_call_flow" {
  filepath          = "${local.working_dir.flow}/inboundcall_flow_example_substitutions.yaml" // Also supports S3 paths e.g. s3://my-bucket/flows/example.yaml
  file_content_hash = filesha256("${local.working_dir.flow}/inboundcall_flow_example_substitutions.yaml")
  // Example flow configuration using substitutions:
  /*
  inboundCall:
    name: "{{flow_name}}"
    defaultLanguage: "{{default_language}}"
    startUpRef: ./menus/menu[mainMenu]
    initialGreeting:
      tts: "{{greeting}}"
    menus:
      - menu:
          name: Main Menu
          audio:
            tts: You are at the Main Menu, press 9 to disconnect.
          refId: mainMenu
          choices:
            - menuDisconnect:
                name: "{{menu_disconnect_name}}"
                dtmf: digit_9
  */
  // see https://developer.genesys.cloud/devapps/archy/flowAuthoring/lesson_07_substitutions
  // these replace the key-value pairs from the --optionsFile when using the archy CLI
  substitutions = {
    flow_name            = "An example flow"
    default_language     = "en-us"
    greeting             = "Hello World"
    menu_disconnect_name = "Disconnect"
  }
}
resource "genesyscloud_flow" "outbound_call_flow" {
  filepath          = "${local.working_dir.flow}/outboundcall_flow_example.yaml"
  file_content_hash = filesha256("${local.working_dir.flow}/outboundcall_flow_example.yaml")
  substitutions = {
    flow_name          = "An example outbound flow"
    home_division_name = data.genesyscloud_auth_division_home.home.name
    contact_list_name  = genesyscloud_outbound_contact_list.contact_list.name
    wrapup_code_name   = genesyscloud_routing_wrapupcode.win.name
  }
}
