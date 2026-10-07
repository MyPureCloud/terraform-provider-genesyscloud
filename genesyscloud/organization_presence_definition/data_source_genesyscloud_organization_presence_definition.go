package organization_presence_definition

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The data_source_genesyscloud_organization_presence_definition.go contains the read logic for the
organization_presence_definition data source. Presence definitions have no unique name on the API, so
lookup is performed on the computed label (en_US/en, via GenerateComputedName) and can be disambiguated
with the optional system_presence and division_id filters.
*/

func dataSourceOrganizationPresenceDefinitionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getOrganizationPresenceDefinitionProxy(sdkConfig)

	name := d.Get("name").(string)
	systemPresence := d.Get("system_presence").(string)
	divisionId := d.Get("division_id").(string)

	// Retry in case search has not yet indexed the presence definition.
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		definitions, resp, err := proxy.getAllOrganizationPresenceDefinition(ctx)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting organization presence definitions: %s", err), resp))
		}

		matches := filterOrganizationPresenceDefinitions(definitions, name, systemPresence, divisionId)

		if len(matches) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no organization presence definition found with name %s", name), resp))
		}
		if len(matches) > 1 {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("ambiguous match: %d organization presence definitions found with name %s; narrow the search with system_presence and/or division_id", len(matches), name), resp))
		}

		d.SetId(*matches[0].Id)
		return nil
	})
}

// filterOrganizationPresenceDefinitions returns the user-type presence definitions matching the given
// name (computed label) and optional system_presence / division_id filters. It is separated from the
// read function so the matching logic can be unit tested without an API client.
func filterOrganizationPresenceDefinitions(definitions *[]platformclientv2.Organizationpresencedefinition, name, systemPresence, divisionId string) []platformclientv2.Organizationpresencedefinition {
	var matches []platformclientv2.Organizationpresencedefinition
	if definitions == nil {
		return matches
	}
	for _, def := range *definitions {
		// We only provide service for user type presence definitions, matching the resource and exporter behavior.
		if def.VarType == nil || *def.VarType != "User" {
			continue
		}
		if def.LanguageLabels == nil || GenerateComputedName(def) != name {
			continue
		}
		if systemPresence != "" && (def.SystemPresence == nil || *def.SystemPresence != systemPresence) {
			continue
		}
		if divisionId != "" && (def.DivisionId == nil || *def.DivisionId != divisionId) {
			continue
		}
		matches = append(matches, def)
	}
	return matches
}
