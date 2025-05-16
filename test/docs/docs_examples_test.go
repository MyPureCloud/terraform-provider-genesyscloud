package testing

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/test/docs/examples"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

// Set to TRUE to display the full output of the content being passed to Terraform with line numbers
// This is useful for debugging the output of the Terraform configuration
var SHOW_EXAMPLE_TERRAFORM_CONFIG_OUTPUT_WITH_LINES = true

func TestExampleResources(t *testing.T) {

	var domain string
	var authorizationProducts []string

	planOnly, err := strconv.ParseBool(os.Getenv("TF_PLAN_ONLY"))
	if err != nil {
		planOnly = false
	}
	if planOnly {
		fmt.Fprintln(os.Stdout, "Sanity testing the resources defined in the examples...")
	} else {
		fmt.Fprintln(os.Stdout, "Acceptance testing the resources defined in the examples...")
		provider.AuthorizeSdk()
		authAPI := platformclientv2.NewAuthorizationApi()
		productEntities, api, err := authAPI.GetAuthorizationProducts()
		if err != nil {
			err = fmt.Errorf("Failed to get authorization products from the API: %s", err)
			t.Fatal(err)
		}
		authorizationProducts = make([]string, *productEntities.Total)
		for _, product := range *productEntities.Entities {
			authorizationProducts = append(authorizationProducts, *product.Id)
		}
		domain = strings.Join(strings.Split(api.Response.Request.URL.Host, ".")[1:], ".")
	}

	resources := []string{
		// "genesyscloud_architect_datatable",
		// "genesyscloud_architect_datatable_row",
		// "genesyscloud_architect_emergencygroup",
		// "genesyscloud_architect_grammar",
		// "genesyscloud_architect_grammar_language",
		// "genesyscloud_architect_ivr",
		// "genesyscloud_architect_schedulegroups",
		// "genesyscloud_architect_schedules",
		// "genesyscloud_architect_user_prompt",
		// "genesyscloud_auth_division",
		// "genesyscloud_auth_role",
		// "genesyscloud_conversations_messaging_integrations_instagram",
		// "genesyscloud_conversations_messaging_integrations_open",
		// "genesyscloud_conversations_messaging_integrations_whatsapp",
		// "genesyscloud_conversations_messaging_settings",
		// "genesyscloud_conversations_messaging_settings_default",
		// "genesyscloud_conversations_messaging_supportedcontent",
		// "genesyscloud_conversations_messaging_supportedcontent_default",
		// "genesyscloud_employeeperformance_externalmetrics_definitions",
		// "genesyscloud_externalcontacts_contact",
		// "genesyscloud_externalcontacts_external_source",
		// "genesyscloud_externalcontacts_organization",
		// "genesyscloud_flow",
		// "genesyscloud_flow_loglevel",
		// "genesyscloud_flow_milestone",
		// "genesyscloud_flow_outcome",
		// "genesyscloud_group",
		// "genesyscloud_group_roles",
		// "genesyscloud_idp_adfs",
		// "genesyscloud_idp_generic",
		// "genesyscloud_idp_gsuite",
		// "genesyscloud_idp_okta",
		// "genesyscloud_idp_onelogin",
		// "genesyscloud_idp_ping",
		// "genesyscloud_idp_salesforce",
		// "genesyscloud_integration_credential",
		// "genesyscloud_integration",
		// "genesyscloud_integration_action",
		// "genesyscloud_integration_custom_auth_action",
		// "genesyscloud_integration_custom_auth_action",
		// "genesyscloud_integration_facebook",
		// "genesyscloud_journey_action_map",
		// "genesyscloud_journey_action_template",
		// "genesyscloud_journey_outcome",
		// "genesyscloud_journey_outcome_predictor",
		// "genesyscloud_journey_segment",
		// "genesyscloud_journey_view_schedule",
		// "genesyscloud_journey_views",
		// "genesyscloud_knowledge_category",
		// "genesyscloud_knowledge_document",
		// "genesyscloud_knowledge_document_variation",
		// "genesyscloud_knowledge_knowledgebase",
		// "genesyscloud_knowledge_label",
		// "genesyscloud_location",
		// "genesyscloud_oauth_client",
		// "genesyscloud_organization_authentication_settings",
		// "genesyscloud_orgauthorization_pairing",
		// "genesyscloud_outbound_attempt_limit",
		// "genesyscloud_outbound_callabletimeset",
		// "genesyscloud_outbound_callanalysisresponseset",
		// "genesyscloud_outbound_campaign",
		// "genesyscloud_outbound_campaignrule",
		// "genesyscloud_outbound_contact_list",
		// "genesyscloud_outbound_contact_list_contact",
		// "genesyscloud_outbound_contact_list_template",
		// "genesyscloud_outbound_contactlistfilter",
		// "genesyscloud_outbound_digitalruleset",
		// "genesyscloud_outbound_dnclist",
		// "genesyscloud_outbound_filespecificationtemplate",
		// "genesyscloud_outbound_messagingcampaign",
		// "genesyscloud_outbound_ruleset",
		// "genesyscloud_outbound_sequence",
		// "genesyscloud_outbound_settings",
		// "genesyscloud_outbound_wrapupcodemappings",
		// "genesyscloud_processautomation_trigger",
		// "genesyscloud_quality_forms_evaluation",
		// "genesyscloud_quality_forms_survey",
		// "genesyscloud_recording_media_retention_policy",
		// "genesyscloud_responsemanagement_library",
		// "genesyscloud_responsemanagement_response",
		// "genesyscloud_responsemanagement_responseasset",
		// "genesyscloud_routing_email_domain",
		// "genesyscloud_routing_email_route",
		// "genesyscloud_routing_language",
		// "genesyscloud_routing_queue",
		// "genesyscloud_routing_queue_conditional_group_routing",
		// "genesyscloud_routing_queue_outbound_email_address",
		// "genesyscloud_routing_settings",
		// "genesyscloud_routing_skill",
		// "genesyscloud_routing_skill_group",
		// "genesyscloud_routing_sms_address",
		// "genesyscloud_routing_utilization",
		// "genesyscloud_routing_utilization_label",
		// "genesyscloud_routing_wrapupcode",
		// "genesyscloud_script",
		// "genesyscloud_task_management_workbin",
		// "genesyscloud_task_management_workitem",
		// "genesyscloud_task_management_workitem_schema",
		// "genesyscloud_task_management_worktype",
		// "genesyscloud_task_management_worktype_flow_datebased_rule",
		// "genesyscloud_task_management_worktype_flow_onattributechange_rule",
		// "genesyscloud_task_management_worktype_flow_oncreate_rule",
		// "genesyscloud_task_management_worktype_status",
		// "genesyscloud_task_management_worktype_status_transition",
		// "genesyscloud_team",
		// "genesyscloud_telephony_providers_edges_did_pool",
		// "genesyscloud_telephony_providers_edges_edge_group",
		// "genesyscloud_telephony_providers_edges_extension_pool",
		// "genesyscloud_telephony_providers_edges_phone",
		// "genesyscloud_telephony_providers_edges_phonebasesettings",
		// "genesyscloud_telephony_providers_edges_site",
		// "genesyscloud_telephony_providers_edges_site_outbound_route",
		// "genesyscloud_telephony_providers_edges_trunk", // DEPRECATE
		// "genesyscloud_telephony_providers_edges_trunkbasesettings",
		// "genesyscloud_tf_export",
		// "genesyscloud_user",
		// "genesyscloud_user_roles",
		// "genesyscloud_webdeployments_configuration",
		// "genesyscloud_webdeployments_deployment",
	}

	providerResources, providerDataSources := provider_registrar.GetProviderResources()
	providerFactories := provider.GetProviderFactories(providerResources, providerDataSources)

	// Add some extra built in providers to be able to be used
	providerFactories = provider.CombineProviderFactories(providerFactories, UtilsProviderFactory())

	// Get absolute path of current working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	successfulResourceTypes := make(map[string]string, len(resources))

	for _, resourceType := range resources {
		exampleDir := filepath.Join(wd, "..", "..", "examples", "resources", resourceType)

		t.Run(resourceType, func(t *testing.T) {

			example, _, err := examples.LoadExampleWithDependencies(filepath.Join(exampleDir, "resource.tf"), nil)
			if err != nil {
				t.Fatal(err)
			}
			resourceExampleContent, err := example.GenerateOutput()
			if err != nil {
				t.Fatal(err)
			}
			checks := example.GenerateChecks()

			if !planOnly {
				// Add arbitrary sleep to allow API to catch up before attempting to delete
				// Also provides a great place to place a breakpoint if needing to pause after Terraform Create and before Delete
				checks = append(checks, func(s *terraform.State) error {
					time.Sleep(2 * time.Second)
					return nil
				})
			}

			successfulResourceTypes[resourceType] = "success"

			// Run test
			resource.Test(t, resource.TestCase{
				PreCheck: func() {
					if !planOnly {
						util.TestAccPreCheck(t)
						if SHOW_EXAMPLE_TERRAFORM_CONFIG_OUTPUT_WITH_LINES {
							// 12 is the number of lines the provider block (not shown) takes up before outputting the rest of the config
							// Retained for debugging purposes, allows the line numbers in error messages to line up.
							util.PrintStringWithLineNumbers(resourceExampleContent, 12)
						}
					}
				},
				ErrorCheck: func(err error) error {
					successfulResourceTypes[resourceType] = "errored"
					return err
				},
				ProviderFactories: providerFactories,
				ExternalProviders: map[string]resource.ExternalProvider{
					"random": {
						Source:            "hashicorp/random",
						VersionConstraint: "3.7.2",
					},
					"time": {
						Source:            "hashicorp/time",
						VersionConstraint: "0.13.1",
					},
				},
				Steps: []resource.TestStep{
					{
						SkipFunc: func() (bool, error) {
							shouldSkip := example.GenerateSkipFunc(domain, authorizationProducts)
							if shouldSkip {
								successfulResourceTypes[resourceType] = "skipped"
							}
							return shouldSkip, nil
						},
						Config: string(resourceExampleContent),
						Check: resource.ComposeTestCheckFunc(
							// arbitrary check with sleep
							checks...,
						),
						PlanOnly:           planOnly,
						ExpectNonEmptyPlan: planOnly,
					},
				},
			})
			if t.Failed() {
				successfulResourceTypes[resourceType] = "failed"
			}

			if !planOnly {
				// Pause for five seconds to allow GC API to finish processing delete
				time.Sleep(time.Second * 5)
			}

		})
	}

	io.WriteString(os.Stdout, "The following resources were successfull:\n")
	for srt, status := range successfulResourceTypes {
		if status == "success" {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srt))
		}
	}
	io.WriteString(os.Stdout, "The following resources were errored:\n")
	for srt, status := range successfulResourceTypes {
		if status == "errored" {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srt))
		}
	}
	io.WriteString(os.Stdout, "The following resources were failed:\n")
	for srt, status := range successfulResourceTypes {
		if status == "failed" {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srt))
		}
	}
	io.WriteString(os.Stdout, "The following resources were skipped:\n")
	for srt, status := range successfulResourceTypes {
		if status == "skipped" {
			io.WriteString(os.Stdout, fmt.Sprintf("  - %s\n", srt))
		}
	}
}
