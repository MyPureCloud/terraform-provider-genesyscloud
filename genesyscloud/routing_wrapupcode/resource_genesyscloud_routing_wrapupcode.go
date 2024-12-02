package routing_wrapupcode

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllRoutingWrapupCodes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getRoutingWrapupcodeProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	wrapupcodes, proxyResponse, getErr := proxy.getAllRoutingWrapupcode(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of routing wrapupcode error: %s", getErr), proxyResponse)
	}

	for _, wrapupcode := range *wrapupcodes {
		resources[*wrapupcode.Id] = &resourceExporter.ResourceMeta{BlockLabel: *wrapupcode.Name}
	}

	return resources, nil
}

func createRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingWrapupcodeProxy(sdkConfig)

	name := d.Get("name").(string)
	wrapupCode := buildWrapupCodeFromResourceData(d)

	log.Printf("Creating wrapupcode %s", name)
	wrapupcodeResponse, proxyResponse, err := proxy.createRoutingWrapupcode(ctx, wrapupCode)

	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create wrapupcode %s error: %s", name, err), proxyResponse)
	}

	d.SetId(*wrapupcodeResponse.Id)
	log.Printf("Created wrapupcode %s %s", name, *wrapupcodeResponse.Id)
	return readRoutingWrapupCode(ctx, d, meta)
}

func readRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingWrapupcodeProxy(sdkConfig)

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingWrapupCode(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading wrapupcode %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		wrapupcode, proxyResponse, err := proxy.getRoutingWrapupcodeById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(proxyResponse) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read wrapupcode %s | error: %s", d.Id(), err), proxyResponse))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read wrapupcode %s | error: %s", d.Id(), err), proxyResponse))
		}

		resourcedata.SetNillableValue(d, "name", wrapupcode.Name)
		if wrapupcode.Division != nil && wrapupcode.Division.Id != nil {
			_ = d.Set("division_id", *wrapupcode.Division.Id)
		}

		log.Printf("Read wrapupcode %s %s", d.Id(), *wrapupcode.Name)
		return cc.CheckState(d)
	})
}

func updateRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingWrapupcodeProxy(sdkConfig)

	name := d.Get("name").(string)
	wrapupCode := buildWrapupCodeFromResourceData(d)

	log.Printf("Updating wrapupcode %s", name)
	_, proxyUpdResponse, err := proxy.updateRoutingWrapupcode(ctx, d.Id(), wrapupCode)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update wrapupcode %s error: %s", name, err), proxyUpdResponse)
	}

	log.Printf("Updated wrapupcode %s", name)

	return readRoutingWrapupCode(ctx, d, meta)
}

func deleteRoutingWrapupCode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingWrapupcodeProxy(sdkConfig)

	name := d.Get("name").(string)

	log.Printf("Deleting wrapupcode %s", name)
	proxyDelResponse, err := proxy.deleteRoutingWrapupcode(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete wrapupcode %s error: %s", name, err), proxyDelResponse)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, proxyGetResponse, err := proxy.getRoutingWrapupcodeById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(proxyGetResponse) {
				// Routing wrapup code deleted
				log.Printf("Deleted Routing wrapup code %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Routing wrapup code %s | error: %s", d.Id(), err), proxyGetResponse))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Routing wrapup code %s still exists", d.Id()), proxyGetResponse))
	})
}

func buildWrapupCodeFromResourceData(d *schema.ResourceData) *platformclientv2.Wrapupcoderequest {
	name := d.Get("name").(string)
	divisionId, _ := d.Get("division_id").(string)
	wrapupCode := &platformclientv2.Wrapupcoderequest{
		Name: &name,
	}
	if divisionId != "" {
		wrapupCode.Division = &platformclientv2.Writablestarrabledivision{Id: &divisionId}
	}
	return wrapupCode
}

func GenerateRoutingWrapupcodeResource(resourceLabel string, name string, divisionId string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		division_id = %s
	}
	`, ResourceType, resourceLabel, name, divisionId)
}
