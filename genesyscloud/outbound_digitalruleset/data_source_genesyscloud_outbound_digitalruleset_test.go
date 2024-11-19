package outbound_digitalruleset

import (
	"fmt"
	"strconv"
	"testing"

	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the outbound digitalruleset Data Source
*/

func TestAccDataSourceOutboundDigitalruleset(t *testing.T) {
	t.Parallel()
	var (
		name1             = "Terraform Digital RuleSet1"
		resourceId        = "digital-rule-set"
		ruleName          = "RuleWork"
		dataSourceId      = "data-digital-rule-set"
		ruleOrder         = "0"
		ruleCategory      = "PreContact"
		contactColumnName = "Work"
		columnOperator    = "Equals"
		columnValue       = "XYZ"
		columnValueType   = "String"

		updatePropertiesWork = "Work"
		updateOption         = "Set"

		contactListResourceId1    = "contact-list-1"
		contactListName1          = "Test Contact List " + uuid.NewString()
		previewModeColumnName     = ""
		previewModeAcceptedValues = []string{}
		columnNames               = []string{strconv.Quote("Cell"), strconv.Quote("Work")}
		automaticTimeZoneMapping  = util.FalseValue
	)

	contactListResourceGenerate := obContactList.GenerateOutboundContactList(
		contactListResourceId1,
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
		obContactList.GenerateEmailColumnsBlock(
			"Work",
			"Work",
			strconv.Quote("Work"),
		),
		obContactList.GeneratePhoneColumnsDataTypeSpecBlock(
			strconv.Quote("Cell"),
			strconv.Quote("TEXT"),
			"1",
			"11",
			"10",
		),
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{

				Config: contactListResourceGenerate +
					GenerateOutboundDigitalRuleSetResource(
						resourceId,
						name1,
						"genesyscloud_outbound_contact_list."+contactListResourceId1+".id",
						GenerateDigitalRules(
							ruleName,
							ruleOrder,
							ruleCategory,
							GenerateDigitalRuleSetConditions(
								GenerateInvertedConditionAttr(util.FalseValue),
								GenerateContactColumnConditionSettings(
									contactColumnName,
									columnOperator,
									columnValue,
									columnValueType,
								),
							),
							GenerateDigitalRuleSetActions(
								GenerateUpdateContactColumnActionSettings(
									updateOption,
									GeneratePropertiesForUpdateContactColumnSettings(updatePropertiesWork, updatePropertiesWork),
								),
							),
							GenerateDigitalRuleSetActions(
								GenerateDoNotSendActionSettings(),
							),
						),
					) +
					generateOutboundDigitalRuleSetDataSource(
						dataSourceId,
						name1,
						"genesyscloud_outbound_digitalruleset."+resourceId,
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_digitalruleset."+dataSourceId, "id",
						"genesyscloud_outbound_digitalruleset."+resourceId, "id"),
				),
			},
		},
	})
}

func generateOutboundDigitalRuleSetDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_digitalruleset" "%s" {
	name       = "%s"
	depends_on = [%s]
}
`, id, name, dependsOn)
}
