package responsemanagement_response

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_responsemanagement_response.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthResponsemanagementResponse retrieves all of the responsemanagement response via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthResponsemanagementResponses(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getResponsemanagementResponseProxy(clientConfig)

	responseManagementResponses, resp, err := proxy.getAllResponsemanagementResponse(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get list of response management responses error: %s", err), resp)
	}

	for _, response := range *responseManagementResponses {
		resources[*response.Id] = &resourceExporter.ResourceMeta{BlockLabel: *response.Name}
	}
	return resources, nil
}

// createResponsemanagementResponse is used by the responsemanagement_response resource to create Genesys cloud responsemanagement response
func createResponsemanagementResponse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getResponsemanagementResponseProxy(sdkConfig)

	sdkResponse := getResponseFromResourceData(d)

	diagErr := util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		log.Printf("Creating Responsemanagement Response %s", *sdkResponse.Name)
		responsemanagementResponse, resp, err := proxy.createResponsemanagementResponse(ctx, &sdkResponse)
		if err != nil {
			if util.IsStatus412(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to create Responsemanagement Response %s | error: %s", *sdkResponse.Name, err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to create Responsemanagement Response %s | error: %s", *sdkResponse.Name, err), resp))
		}
		d.SetId(*responsemanagementResponse.Id)
		log.Printf("Created Responsemanagement Response %s %s", *sdkResponse.Name, *responsemanagementResponse.Id)
		return nil
	})

	if diagErr != nil {
		return diagErr
	}
	return readResponsemanagementResponse(ctx, d, meta)
}

// readResponsemanagementResponse is used by the responsemanagement_response resource to read a responsemanagement response from genesys cloud
func readResponsemanagementResponse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getResponsemanagementResponseProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceResponsemanagementResponse(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Responsemanagement Response %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkResponse, resp, getErr := proxy.getResponsemanagementResponseById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Responsemanagement Response %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Responsemanagement Response %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", sdkResponse.Name)
		if sdkResponse.Libraries != nil {
			d.Set("library_ids", util.SdkDomainEntityRefArrToList(*sdkResponse.Libraries))
		}
		resourcedata.SetNillableValueWithSchemaSetWithFunc(d, "texts", sdkResponse.Texts, flattenResponseTexts)
		resourcedata.SetNillableValueWithSchemaSetWithFunc(d, "substitutions", sdkResponse.Substitutions, flattenResponseSubstitutions)
		resourcedata.SetNillableValue(d, "interaction_type", sdkResponse.InteractionType)
		if sdkResponse.SubstitutionsSchema != nil && sdkResponse.SubstitutionsSchema.Id != nil {
			d.Set("substitutions_schema_id", *sdkResponse.SubstitutionsSchema.Id)
		}
		if sdkResponse.ResponseType != nil {
			d.Set("response_type", *sdkResponse.ResponseType)
		}
		if sdkResponse.MessagingTemplate != nil {
			d.Set("messaging_template", flattenMessagingTemplate(sdkResponse.MessagingTemplate))
		}
		resourcedata.SetNillableValueWithSchemaSetWithFunc(d, "asset_ids", sdkResponse.Assets, flattenAddressableEntityRefs)
		resourcedata.SetNillableValueWithSchemaSetWithFunc(d, "footer", sdkResponse.Footer, flattenFooterTemplate)

		log.Printf("Read Responsemanagement Response %s %s", d.Id(), *sdkResponse.Name)
		return cc.CheckState(d)
	})
}

// updateResponsemanagementResponse is used by the responsemanagement_response resource to update an responsemanagement response in Genesys Cloud
func updateResponsemanagementResponse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getResponsemanagementResponseProxy(sdkConfig)

	sdkResponse := getResponseFromResourceData(d)

	log.Printf("Updating Responsemanagement Response %s", *sdkResponse.Name)
	managementResponse, resp, err := proxy.updateResponsemanagementResponse(ctx, d.Id(), &sdkResponse)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update response management response %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated Responsemanagement Response %s", *managementResponse.Id)
	return readResponsemanagementResponse(ctx, d, meta)
}

// deleteResponsemanagementResponse is used by the responsemanagement_response resource to delete an responsemanagement response from Genesys cloud
func deleteResponsemanagementResponse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getResponsemanagementResponseProxy(sdkConfig)

	log.Printf("Deleting Responsemanagement Response")
	resp, err := proxy.deleteResponsemanagementResponse(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete Responsemanagement Response %s error: %s", d.Id(), err), resp)
	}

	time.Sleep(30 * time.Second) //Give time for any libraries or assets to be deleted
	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getResponsemanagementResponseById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Responsemanagement Response deleted
				log.Printf("Deleted Responsemanagement Response %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Responsemanagement Response %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Responsemanagement Response %s still exists", d.Id()), resp))
	})
}
