package organization_presence_definition

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The data_source_genesyscloud_organization_presence_definition_test.go contains the acceptance tests for
the organization_presence_definition data source. Each test creates a presence definition resource and
then looks it up through the data source, asserting the resolved id matches the created resource.
*/

// TestAccDataSourceOrganizationPresenceDefinition verifies a presence definition can be looked up by its
// localized label (name).
func TestAccDataSourceOrganizationPresenceDefinition(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel     = "presence-def-ds-1"
		dataSourceLabel   = "presence-def-ds-1-data"
		languageLabelEnus = "DS Lookup " + uuid.NewString()[:8]
		languageLabels    = map[string]string{"en_US": strconv.Quote(languageLabelEnus)}
		languageLabelsStr = util.GenerateMapAttrWithMapProperties("language_labels", languageLabels)
		systemPresence    = "Away"

		resourcePath   = ResourceType + "." + resourceLabel
		dataSourcePath = "data." + ResourceType + "." + dataSourceLabel
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Look up the created presence definition by name only.
				Config: GenerateOrganizationPresenceDefinitionResource(
					resourceLabel,
					languageLabelsStr,
					systemPresence,
					util.NullValue,
				) + generateOrganizationPresenceDefinitionDataSource(
					dataSourceLabel,
					strconv.Quote(languageLabelEnus),
					util.NullValue,
					util.NullValue,
					resourcePath,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourcePath, "id", resourcePath, "id"),
				),
			},
		},
		CheckDestroy: testVerifyOrganizationPresenceDefinitionDestroyed,
	})
}

// TestAccDataSourceOrganizationPresenceDefinitionWithSystemPresence verifies the optional system_presence
// filter resolves the correct definition when supplied.
func TestAccDataSourceOrganizationPresenceDefinitionWithSystemPresence(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel     = "presence-def-ds-2"
		dataSourceLabel   = "presence-def-ds-2-data"
		languageLabelEnus = "DS SysPres " + uuid.NewString()[:8]
		languageLabels    = map[string]string{"en_US": strconv.Quote(languageLabelEnus)}
		languageLabelsStr = util.GenerateMapAttrWithMapProperties("language_labels", languageLabels)
		systemPresence    = "Break"

		resourcePath   = ResourceType + "." + resourceLabel
		dataSourcePath = "data." + ResourceType + "." + dataSourceLabel
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Look up by name and system_presence.
				Config: GenerateOrganizationPresenceDefinitionResource(
					resourceLabel,
					languageLabelsStr,
					systemPresence,
					util.NullValue,
				) + generateOrganizationPresenceDefinitionDataSource(
					dataSourceLabel,
					strconv.Quote(languageLabelEnus),
					strconv.Quote(systemPresence),
					util.NullValue,
					resourcePath,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourcePath, "id", resourcePath, "id"),
					resource.TestCheckResourceAttr(dataSourcePath, "system_presence", systemPresence),
				),
			},
		},
		CheckDestroy: testVerifyOrganizationPresenceDefinitionDestroyed,
	})
}

// generateOrganizationPresenceDefinitionDataSource builds a data source config block. systemPresence and
// divisionId may be util.NullValue to omit the optional filters. dependsOnResource ensures the backing
// resource is created before the data source reads (required for data-source-on-resource references).
func generateOrganizationPresenceDefinitionDataSource(
	dataSourceLabel string,
	name string,
	systemPresence interface{},
	divisionId interface{},
	dependsOnResource string) string {
	var systemPresenceStr string
	if systemPresence != util.NullValue {
		systemPresenceStr = fmt.Sprintf("system_presence = %v", systemPresence)
	}
	var divisionIdStr string
	if divisionId != util.NullValue {
		divisionIdStr = fmt.Sprintf("division_id = %v", divisionId)
	}

	return fmt.Sprintf(`data "%s" "%s" {
		name = %s
		%s
		%s
		depends_on = [%s]
	}
	`, ResourceType, dataSourceLabel, name, systemPresenceStr, divisionIdStr, dependsOnResource)
}
