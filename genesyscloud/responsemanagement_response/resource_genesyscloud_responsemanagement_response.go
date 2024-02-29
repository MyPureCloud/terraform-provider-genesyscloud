package responsemanagement_response

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"
)

/*
The resource_genesyscloud_responsemanagement_response.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthResponsemanagementResponse retrieves all of the responsemanagement response via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthResponsemanagementResponses(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getResponsemanagementResponseProxy(clientConfig)

	responseManagementResponses, err := proxy.getAllResponsemanagementResponse(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get list of response management responses: %v", err)
	}

	for _, response := range *responseManagementResponses {
		resources[*response.Id] = &resourceExporter.ResourceMeta{Name: *response.Name}
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
		responsemanagementResponse, respCode, err := proxy.createResponsemanagementResponse(ctx, &sdkResponse)
		if err != nil {
			if util.IsStatus412ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("failed to create Responsemanagement Response %s: %s", *sdkResponse.Name, err))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to create Responsemanagement Response %s: %s", *sdkResponse.Name, err))
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

	log.Printf("Reading Responsemanagement Response %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkResponse, respCode, getErr := proxy.getResponsemanagementResponseById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read Responsemanagement Response %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Responsemanagement Response %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceResponsemanagementResponse())

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
		return cc.CheckState()
	})
}

// updateResponsemanagementResponse is used by the responsemanagement_response resource to update an responsemanagement response in Genesys Cloud
func updateResponsemanagementResponse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getResponsemanagementResponseProxy(sdkConfig)

	sdkResponse := getResponseFromResourceData(d)

	log.Printf("Updating Responsemanagement Response %s", *sdkResponse.Name)
	managementResponse, err := proxy.updateResponsemanagementResponse(ctx, d.Id(), &sdkResponse)
	if err != nil {
		return diag.Errorf("Failed to update response management response %s: %s", d.Id(), err)
	}

	log.Printf("Updated Responsemanagement Response %s", *managementResponse.Id)
	return readResponsemanagementResponse(ctx, d, meta)
}

// deleteResponsemanagementResponse is used by the responsemanagement_response resource to delete an responsemanagement response from Genesys cloud
func deleteResponsemanagementResponse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getResponsemanagementResponseProxy(sdkConfig)

	log.Printf("Deleting Responsemanagement Response")
	_, err := proxy.deleteResponsemanagementResponse(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete Responsemanagement Response: %s", err)
	}

	time.Sleep(30 * time.Second) //Give time for any libraries or assets to be deleted
	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getResponsemanagementResponseById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404ByInt(resp) {
				// Responsemanagement Response deleted
				log.Printf("Deleted Responsemanagement Response %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Responsemanagement Response %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Responsemanagement Response %s still exists", d.Id()))
	})
}
