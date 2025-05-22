package organization_presence_definition

import (
	"context"
	"fmt"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_organization_presence_definition.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthOrganizationPresenceDefinition retrieves all of the organization presence definition via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthOrganizationPresenceDefinitions(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newOrganizationPresenceDefinitionProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	organizationPresenceDefinitions, resp, err := proxy.getAllOrganizationPresenceDefinition(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get organization presence definition: %v", err), resp)
	}

	for _, organizationPresenceDefinition := range *organizationPresenceDefinitions {
		// We only provide service for user type presence definitions as these are the only creatable items in GUI/API. (just like we only provide service for user prompts)
		if *organizationPresenceDefinition.VarType == "User" {
			resources[*organizationPresenceDefinition.Id] = &resourceExporter.ResourceMeta{BlockLabel: *organizationPresenceDefinition.SystemPresence + "_" + GenerateComputedName(organizationPresenceDefinition)}
		}
	}

	return resources, nil
}

// createOrganizationPresenceDefinition is used by the organization_presence_definition resource to create Genesys cloud organization presence definition
func createOrganizationPresenceDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOrganizationPresenceDefinitionProxy(sdkConfig)

	organizationPresenceDefinition := getOrganizationPresenceDefinitionFromResourceData(d)

	log.Printf("Creating organization presence definition %s", GenerateComputedName(organizationPresenceDefinition))
	if *organizationPresenceDefinition.DivisionId == "" {
		divisionId := "*"
		organizationPresenceDefinition.DivisionId = &divisionId
	}
	presenceDefinition, resp, err := proxy.createOrganizationPresenceDefinition(ctx, &organizationPresenceDefinition)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create organization presence definition: %s", err), resp)
	}

	d.SetId(*presenceDefinition.Id)
	log.Printf("Created organization presence definition %s", *presenceDefinition.Id)
	return readOrganizationPresenceDefinition(ctx, d, meta)
}

// readOrganizationPresenceDefinition is used by the organization_presence_definition resource to read an organization presence definition from genesys cloud
func readOrganizationPresenceDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOrganizationPresenceDefinitionProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOrganizationPresenceDefinition(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading organization presence definition %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		organizationPresenceDefinition, resp, getErr := proxy.getOrganizationPresenceDefinitionById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read organization presence definition %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read organization presence definition %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "language_labels", organizationPresenceDefinition.LanguageLabels)
		resourcedata.SetNillableValue(d, "system_presence", organizationPresenceDefinition.SystemPresence)
		if organizationPresenceDefinition.DivisionId != nil {
			_ = d.Set("division_id", *organizationPresenceDefinition.DivisionId)
		}
		resourcedata.SetNillableValue(d, "deactivated", organizationPresenceDefinition.Deactivated)

		log.Printf("Read organization presence definition %s %s", d.Id(), GenerateComputedName(*organizationPresenceDefinition))
		return cc.CheckState(d)
	})
}

// updateOrganizationPresenceDefinition is used by the organization_presence_definition resource to update an organization presence definition in Genesys Cloud
func updateOrganizationPresenceDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOrganizationPresenceDefinitionProxy(sdkConfig)

	presenceDefinition := getOrganizationPresenceDefinitionFromResourceData(d)

	log.Printf("Updating organization presence definition %s", GenerateComputedName(presenceDefinition))
	organizationPresenceDefinition, resp, err := proxy.updateOrganizationPresenceDefinition(ctx, d.Id(), &presenceDefinition)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update organization presence definition %s: %s", d.Id(), err), resp)
	}

	log.Printf("Updated organization presence definition %s", *organizationPresenceDefinition.Id)
	return readOrganizationPresenceDefinition(ctx, d, meta)
}

// deleteOrganizationPresenceDefinition is used by the organization_presence_definition resource to delete an organization presence definition from Genesys cloud
func deleteOrganizationPresenceDefinition(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOrganizationPresenceDefinitionProxy(sdkConfig)

	resp, err := proxy.deleteOrganizationPresenceDefinition(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete organization presence definition %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		organizationPresenceDefinition, resp, err := proxy.getOrganizationPresenceDefinitionById(ctx, d.Id())

		if err != nil {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("organization presence definition %s still exists", d.Id()), resp))
		}
		// A delete for this api is actually just a deactivation
		if organizationPresenceDefinition != nil && *organizationPresenceDefinition.Deactivated {
			log.Printf("Deleted organization presence definition %s", d.Id())
			return nil
		}

		return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting organization presence definition %s: %s", d.Id(), "unknown error"), resp))
	})
}

// getOrganizationPresenceDefinitionFromResourceData maps data from schema ResourceData object to a platformclientv2.Organizationpresencedefinition
func getOrganizationPresenceDefinitionFromResourceData(d *schema.ResourceData) platformclientv2.Organizationpresencedefinition {
	rawLanguageLabels := d.Get("language_labels").(map[string]interface{})
	languageLabels := make(map[string]string)
	for key, value := range rawLanguageLabels {
		if strValue, ok := value.(string); ok {
			languageLabels[key] = strValue
		}
	}

	return platformclientv2.Organizationpresencedefinition{
		LanguageLabels: &languageLabels,
		SystemPresence: platformclientv2.String(d.Get("system_presence").(string)),
		DivisionId:     platformclientv2.String(d.Get("division_id").(string)),
		Deactivated:    platformclientv2.Bool(d.Get("deactivated").(bool)),
	}
}

func GenerateComputedName(organizationPresenceDefinition platformclientv2.Organizationpresencedefinition) string {
	// to build a name, we want to use the en_US or en labels
	var computedName, firstValue string
	if label, exists := (*organizationPresenceDefinition.LanguageLabels)["en_US"]; exists {
		computedName = label
	} else if label, exists := (*organizationPresenceDefinition.LanguageLabels)["en"]; exists {
		computedName = label
	} else {
		// if en_US or en labels are not found, we just grab the first one we find, understanding that this will vary over time
		for _, value := range *organizationPresenceDefinition.LanguageLabels {
			firstValue = value
			break
		}
		computedName = firstValue
	}

	return computedName
}
