package outbound_ruleset

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"

	obContactList "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
)

/*
The resource_genesyscloud_outbound_ruleset_test.go contains all of the test cases for running the resource
tests for outbound_ruleset.
*/

func TestAccResourceOutboundRulesetNoRules(t *testing.T) {
	t.Parallel()
	var (
		contactListResourceLabel1 = "contact-list-1"
		contactListResourceLabel2 = "contact-list-2"
		contactListName1          = "Test Contact List " + uuid.NewString()
		contactListName2          = "Test Contact List " + uuid.NewString()
		previewModeColumnName     = "Cell"
		previewModeAcceptedValues = []string{strconv.Quote(previewModeColumnName)}
		columnNames               = []string{strconv.Quote("Cell"), strconv.Quote("Home")}
		automaticTimeZoneMapping  = util.FalseValue

		queueResourceLabel1 = "test-queue-1"
		queueResourceLabel2 = "test-queue-2"
		queueName1          = "Terraform Test Queue1-" + uuid.NewString()
		queueName2          = "Terraform Test Queue2-" + uuid.NewString()

		ruleSetResourceLabel = "rule-set"
		ruleSetName1         = "Test Rule Set " + uuid.NewString()
		ruleSetName2         = "Test Rule Set " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: obContactList.GenerateOutboundContactList(
					contactListResourceLabel1,
					contactListName1,
					util.NullValue,
					strconv.Quote(previewModeColumnName),
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					util.NullValue,
					util.NullValue,
					obContactList.GeneratePhoneColumnsBlock(
						"Cell",
						"cell",
						strconv.Quote("Cell"),
					),
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel1,
					queueName1) + fmt.Sprintf(`resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
}`, ruleSetResourceLabel, ruleSetName1, contactListResourceLabel1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "name", ruleSetName1),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "contact_list_id", "genesyscloud_outbound_contact_list."+contactListResourceLabel1, "id"),
				),
			},
			// Update name, contact_list_id and queue_id
			{
				Config: obContactList.GenerateOutboundContactList(
					contactListResourceLabel2,
					contactListName2,
					util.NullValue,
					strconv.Quote(previewModeColumnName),
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					util.NullValue,
					util.NullValue,
					obContactList.GeneratePhoneColumnsBlock(
						"Cell",
						"cell",
						strconv.Quote("Cell"),
					),
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel2,
					queueName2) + fmt.Sprintf(`resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
}`, ruleSetResourceLabel, ruleSetName2, contactListResourceLabel2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "name", ruleSetName2),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "contact_list_id", "genesyscloud_outbound_contact_list."+contactListResourceLabel2, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_ruleset." + ruleSetResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyroutingRulesetDestroyed,
	})
}

func TestAccResourceOutboundRuleset(t *testing.T) {
	t.Parallel()
	var (
		contactListResourceLabel1 = "contact-list-1"
		contactListName1          = "Test Contact List " + uuid.NewString()
		previewModeColumnName     = "Cell"
		previewModeAcceptedValues = []string{strconv.Quote(previewModeColumnName)}
		columnNames               = []string{strconv.Quote("Cell"), strconv.Quote("Home")}
		automaticTimeZoneMapping  = util.FalseValue

		ruleSetResourceLabel = "rule-set"
		ruleSetName1         = "Test Rule Set " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: obContactList.GenerateOutboundContactList(
					contactListResourceLabel1,
					contactListName1,
					util.NullValue,
					strconv.Quote(previewModeColumnName),
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					util.NullValue,
					util.NullValue,
					obContactList.GeneratePhoneColumnsBlock(
						"Cell",
						"cell",
						strconv.Quote("Cell"),
					),
				) + fmt.Sprintf(`resource "genesyscloud_outbound_ruleset" "%s" {
  name = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
  rules {
    name     = "DO_NOT_DIAL rule"
    order    = 0
    category = "DIALER_PRECALL"
    conditions {
      type  = "phoneNumberCondition"
      value = "0123456789"
    }
    actions {
      type             = "Action"
      action_type_name = "DO_NOT_DIAL"
    }
  }
}`, ruleSetResourceLabel, ruleSetName1, contactListResourceLabel1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "name", ruleSetName1),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "contact_list_id", "genesyscloud_outbound_contact_list."+contactListResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.name", "DO_NOT_DIAL rule"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.order", "0"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.category", "DIALER_PRECALL"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.type", "phoneNumberCondition"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.value", "0123456789"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.actions.0.type", "Action"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.actions.0.action_type_name", "DO_NOT_DIAL"),
				),
			},
			{
				Config: obContactList.GenerateOutboundContactList(
					contactListResourceLabel1,
					contactListName1,
					util.NullValue,
					strconv.Quote(previewModeColumnName),
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					util.NullValue,
					util.NullValue,
					obContactList.GeneratePhoneColumnsBlock(
						"Cell",
						"cell",
						strconv.Quote("Cell"),
					),
				) + fmt.Sprintf(`resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
  rules {
    name     = "DO_NOT_DIAL rule"
    order    = 0
    category = "DIALER_PRECALL"
    conditions {
      type  = "phoneNumberCondition"
      value = "0123456789"
    }
    conditions {
      type  = "phoneNumberCondition"
      value = "1234567890"
    }
    actions {
      type             = "Action"
      action_type_name = "DO_NOT_DIAL"
    }
  }
  rules {
    name     = "CONTACT_UNCALLABLE rule"
    order    = 1
    category = "DIALER_PRECALL"
    conditions {
      type  = "phoneNumberCondition"
      value = "0123456789"
    }
    actions {
      type             = "Action"
      action_type_name = "CONTACT_UNCALLABLE"
    }
  }
}`, ruleSetResourceLabel, ruleSetName1, contactListResourceLabel1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "name", ruleSetName1),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "contact_list_id", "genesyscloud_outbound_contact_list."+contactListResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.name", "DO_NOT_DIAL rule"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.order", "0"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.category", "DIALER_PRECALL"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.type", "phoneNumberCondition"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.value", "0123456789"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.1.type", "phoneNumberCondition"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.1.value", "1234567890"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.actions.0.type", "Action"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.actions.0.action_type_name", "DO_NOT_DIAL"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.1.name", "CONTACT_UNCALLABLE rule"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.1.order", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.1.category", "DIALER_PRECALL"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.1.conditions.0.type", "phoneNumberCondition"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.1.conditions.0.value", "0123456789"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.1.actions.0.type", "Action"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.1.actions.0.action_type_name", "CONTACT_UNCALLABLE"),
				),
			},
			{
				Config: obContactList.GenerateOutboundContactList(
					contactListResourceLabel1,
					contactListName1,
					util.NullValue,
					strconv.Quote(previewModeColumnName),
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					util.NullValue,
					util.NullValue,
					obContactList.GeneratePhoneColumnsBlock(
						"Cell",
						"cell",
						strconv.Quote("Cell"),
					),
				) + fmt.Sprintf(`
resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
  rules {
    name     = "APPEND_CUSTOM_ENTRY_TO_DNC_LIST rule"
    order    = 1
    category = "DIALER_WRAPUP"
    conditions {
      type     = "callAnalysisCondition"
      inverted = false
      value    = "disposition.classification.callable.person"
      operator = "EQUALS"
    }
    actions {
      type             = "Action"
      action_type_name = "APPEND_CUSTOM_ENTRY_TO_DNC_LIST"
      properties = {
        dncListId = "test-dnc-list-id"
        customValue = "test-custom-value"
        neverExpire = "true"
      }
    }
  }
}`, ruleSetResourceLabel, ruleSetName1, contactListResourceLabel1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "name", ruleSetName1),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "contact_list_id", "genesyscloud_outbound_contact_list."+contactListResourceLabel1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.name", "APPEND_CUSTOM_ENTRY_TO_DNC_LIST rule"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.order", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.category", "DIALER_WRAPUP"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.type", "callAnalysisCondition"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.inverted", "false"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.value", "disposition.classification.callable.person"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.operator", "EQUALS"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.actions.0.type", "Action"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.actions.0.action_type_name", "APPEND_CUSTOM_ENTRY_TO_DNC_LIST"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.actions.0.properties.dncListId", "test-dnc-list-id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.actions.0.properties.customValue", "test-custom-value"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.actions.0.properties.neverExpire", "true"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_ruleset." + ruleSetResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyroutingRulesetDestroyed,
	})
}

func TestAccResourceOutboundRulesetTimeAndDateCondition(t *testing.T) {
	t.Parallel()
	var (
		contactListResourceLabel  = "contact-list"
		contactListName           = "Test Contact List " + uuid.NewString()
		previewModeColumnName     = "Cell"
		previewModeAcceptedValues = []string{strconv.Quote(previewModeColumnName)}
		columnNames               = []string{strconv.Quote("Cell"), strconv.Quote("Home")}
		automaticTimeZoneMapping  = util.FalseValue

		ruleSetResourceLabel = "rule-set"
		ruleSetName          = "Test Rule Set " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: obContactList.GenerateOutboundContactList(
					contactListResourceLabel,
					contactListName,
					util.NullValue,
					strconv.Quote(previewModeColumnName),
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					util.NullValue,
					util.NullValue,
					obContactList.GeneratePhoneColumnsBlock(
						"Cell",
						"cell",
						strconv.Quote("Cell"),
					),
				) + fmt.Sprintf(`resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
  rules {
    name     = "Time and Date rule"
    order    = 0
    category = "DIALER_PRECALL"
    conditions {
      type                 = "timeAndDateCondition"
      time_zone_id         = "Europe/Dublin"
      match_any_conditions = false
      sub_conditions {
        type     = "timeOfDay"
        operator = "BETWEEN"
        range {
          min = "09:00"
          max = "17:00"
        }
      }
      sub_conditions {
        type     = "dayOfWeek"
        operator = "IN"
        range {
          in_set = ["1", "2", "3", "4", "5"]
        }
      }
      sub_conditions {
        type     = "specificDate"
        operator = "BEFORE"
        threshold_value = "2025-10-10"
      }
    }
    actions {
      type             = "Action"
      action_type_name = "DO_NOT_DIAL"
    }
  }
}`, ruleSetResourceLabel, ruleSetName, contactListResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "name", ruleSetName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.name", "Time and Date rule"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.type", "timeAndDateCondition"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.time_zone_id", "Europe/Dublin"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.match_any_conditions", "false"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.0.type", "timeOfDay"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.0.operator", "BETWEEN"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.0.range.0.min", "09:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.0.range.0.max", "17:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.1.type", "dayOfWeek"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.1.operator", "IN"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.1.range.0.in_set.0", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.1.range.0.in_set.4", "5"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.2.type", "specificDate"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.2.operator", "BEFORE"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.2.threshold_value", "2025-10-10"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.2.include_year", "true"),
				),
			},
			{
				Config: obContactList.GenerateOutboundContactList(
					contactListResourceLabel,
					contactListName,
					util.NullValue,
					strconv.Quote(previewModeColumnName),
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					util.NullValue,
					util.NullValue,
					obContactList.GeneratePhoneColumnsBlock(
						"Cell",
						"cell",
						strconv.Quote("Cell"),
					),
				) + fmt.Sprintf(`resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
  rules {
    name     = "Time and Date rule"
    order    = 0
    category = "DIALER_PRECALL"
    conditions {
      type                 = "timeAndDateCondition"
      time_zone_id         = "America/Chicago"
      match_any_conditions = true
      sub_conditions {
        type     = "timeOfDay"
        operator = "BETWEEN"
        range {
          min = "08:00"
          max = "18:00"
        }
      }
      sub_conditions {
        type     = "dayOfMonth"
        operator = "IN"
        range {
          in_set = ["1", "15", "LAST_DAY"]
        }
      }
    }
    actions {
      type             = "Action"
      action_type_name = "CONTACT_UNCALLABLE"
    }
  }
}`, ruleSetResourceLabel, ruleSetName, contactListResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "name", ruleSetName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.time_zone_id", "America/Chicago"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.match_any_conditions", "true"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.0.range.0.min", "08:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.0.range.0.max", "18:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.1.type", "dayOfMonth"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.1.range.0.in_set.0", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.conditions.0.sub_conditions.1.range.0.in_set.2", "LAST_DAY"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceLabel, "rules.0.actions.0.action_type_name", "CONTACT_UNCALLABLE"),
				),
			},
			{
				ResourceName:      "genesyscloud_outbound_ruleset." + ruleSetResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyroutingRulesetDestroyed,
	})
}

func testVerifyroutingRulesetDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_outbound_ruleset" {
			continue
		}
		ruleset, resp, err := outboundAPI.GetOutboundRuleset(rs.Primary.ID)
		if ruleset != nil {
			return fmt.Errorf("ruleset (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// ruleset not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All rulesets destroyed
	return nil
}
