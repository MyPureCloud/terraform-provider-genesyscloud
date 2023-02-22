package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

func TestAccResourceOutboundRulesetNoRules(t *testing.T) {
	t.Parallel()
	var (
		contactListResourceId1    = "contact-list-1"
		contactListResourceId2    = "contact-list-2"
		contactListName1          = "Test Contact List " + uuid.NewString()
		contactListName2          = "Test Contact List " + uuid.NewString()
		previewModeColumnName     = "Cell"
		previewModeAcceptedValues = []string{strconv.Quote(previewModeColumnName)}
		columnNames               = []string{strconv.Quote("Cell"), strconv.Quote("Home")}
		automaticTimeZoneMapping  = falseValue

		queueResource1 = "test-queue-1"
		queueResource2 = "test-queue-2"
		queueName1     = "Terraform Test Queue1-" + uuid.NewString()
		queueName2     = "Terraform Test Queue2-" + uuid.NewString()

		ruleSetResourceId = "rule-set"
		ruleSetName1      = "Test Rule Set " + uuid.NewString()
		ruleSetName2      = "Test Rule Set " + uuid.NewString()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: generateOutboundContactList(
					contactListResourceId1,
					contactListName1,
					"",
					previewModeColumnName,
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					"",
					"",
					generatePhoneColumnsBlock(
						"Cell",
						"cell",
						"Cell",
					),
				) + GenerateRoutingQueueResourceBasic(
					queueResource1,
					queueName1) + fmt.Sprintf(`resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
}`, ruleSetResourceId, ruleSetName1, contactListResourceId1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "name", ruleSetName1),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_ruleset."+ruleSetResourceId, "contact_list_id", "genesyscloud_outbound_contact_list."+contactListResourceId1, "id"),
				),
			},
			// Update name, contact_list_id and queue_id
			{
				Config: generateOutboundContactList(
					contactListResourceId2,
					contactListName2,
					"",
					previewModeColumnName,
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					"",
					"",
					generatePhoneColumnsBlock(
						"Cell",
						"cell",
						"Cell",
					),
				) + GenerateRoutingQueueResourceBasic(
					queueResource2,
					queueName2) + fmt.Sprintf(`resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
}`, ruleSetResourceId, ruleSetName2, contactListResourceId2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "name", ruleSetName2),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_ruleset."+ruleSetResourceId, "contact_list_id", "genesyscloud_outbound_contact_list."+contactListResourceId2, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_ruleset." + ruleSetResourceId,
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
		contactListResourceId1    = "contact-list-1"
		contactListName1          = "Test Contact List " + uuid.NewString()
		previewModeColumnName     = "Cell"
		previewModeAcceptedValues = []string{strconv.Quote(previewModeColumnName)}
		columnNames               = []string{strconv.Quote("Cell"), strconv.Quote("Home")}
		automaticTimeZoneMapping  = falseValue

		ruleSetResourceId = "rule-set"
		ruleSetName1      = "Test Rule Set " + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: generateOutboundContactList(
					contactListResourceId1,
					contactListName1,
					"",
					previewModeColumnName,
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					"",
					"",
					generatePhoneColumnsBlock(
						"Cell",
						"cell",
						"Cell",
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
}`, ruleSetResourceId, ruleSetName1, contactListResourceId1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "name", ruleSetName1),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_ruleset."+ruleSetResourceId, "contact_list_id", "genesyscloud_outbound_contact_list."+contactListResourceId1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.name", "DO_NOT_DIAL rule"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.order", "0"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.category", "DIALER_PRECALL"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.conditions.0.type", "phoneNumberCondition"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.conditions.0.value", "0123456789"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.actions.0.type", "Action"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.actions.0.action_type_name", "DO_NOT_DIAL"),
				),
			},
			{
				Config: generateOutboundContactList(
					contactListResourceId1,
					contactListName1,
					"",
					previewModeColumnName,
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					"",
					"",
					generatePhoneColumnsBlock(
						"Cell",
						"cell",
						"Cell",
					),
				) + fmt.Sprintf(`resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
  rules {
    name     = "DO_NOT_DIAL rule"
    order    = 0
    category = "DIALER_PRECALL"
    conditions {
      type                       = "phoneNumberCondition"
      value                      = "0123456789"
    }
    conditions {
      type                       = "phoneNumberCondition"
      value                      = "1234567890"
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
      type                       = "phoneNumberCondition"
      value                      = "0123456789"
    }
    actions {
      type             = "Action"
      action_type_name = "CONTACT_UNCALLABLE"
    }
  }
}`, ruleSetResourceId, ruleSetName1, contactListResourceId1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "name", ruleSetName1),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_ruleset."+ruleSetResourceId, "contact_list_id", "genesyscloud_outbound_contact_list."+contactListResourceId1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.name", "DO_NOT_DIAL rule"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.order", "0"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.category", "DIALER_PRECALL"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.conditions.0.type", "phoneNumberCondition"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.conditions.0.value", "0123456789"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.conditions.1.type", "phoneNumberCondition"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.conditions.1.value", "1234567890"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.actions.0.type", "Action"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.actions.0.action_type_name", "DO_NOT_DIAL"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.1.name", "CONTACT_UNCALLABLE rule"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.1.order", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.1.category", "DIALER_PRECALL"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.1.conditions.0.type", "phoneNumberCondition"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.1.conditions.0.value", "0123456789"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.1.actions.0.type", "Action"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.1.actions.0.action_type_name", "CONTACT_UNCALLABLE"),
				),
			},
			{
				Config: generateOutboundContactList(
					contactListResourceId1,
					contactListName1,
					"",
					previewModeColumnName,
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					"",
					"",
					generatePhoneColumnsBlock(
						"Cell",
						"cell",
						"Cell",
					),
				) + fmt.Sprintf(`resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  contact_list_id = genesyscloud_outbound_contact_list.%s.id
  rules {
    name     = "CONTACT_UNCALLABLE rule"
    order    = 1
    category = "DIALER_PRECALL"
    conditions {
      type                       = "phoneNumberCondition"
      value                      = "0123456789"
    }
    conditions {
      type                       = "phoneNumberCondition"
      value                      = "1234567890"
    }
    actions {
      type             = "Action"
      action_type_name = "CONTACT_UNCALLABLE"
    }
  }
}`, ruleSetResourceId, ruleSetName1, contactListResourceId1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "name", ruleSetName1),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_ruleset."+ruleSetResourceId, "contact_list_id", "genesyscloud_outbound_contact_list."+contactListResourceId1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.name", "CONTACT_UNCALLABLE rule"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.order", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.category", "DIALER_PRECALL"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.conditions.0.type", "phoneNumberCondition"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.conditions.0.value", "0123456789"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.conditions.1.type", "phoneNumberCondition"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.conditions.1.value", "1234567890"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.actions.0.type", "Action"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_ruleset."+ruleSetResourceId, "rules.0.actions.0.action_type_name", "CONTACT_UNCALLABLE"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_ruleset." + ruleSetResourceId,
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
		} else if isStatus404(resp) {
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
