package external_contacts_external_source

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"log"
	"time"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

// getAllAuthExternalContactsExternalSources retrieves all of the external sources via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthExternalContactsExternalSources(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	ep := getExternalContactsExternalSourceProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	externalSources, resp, err := ep.getAllExternalContactsExternalSources(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get External Sources error: %s", err), resp)
	}

	for _, externalSource := range *externalSources {
		if externalSource.Id == nil {
			continue
		}
		log.Printf("Dealing with external source id : %s", *externalSource.Id)
		resources[*externalSource.Id] = &resourceExporter.ResourceMeta{BlockLabel: *externalSource.Id}
	}
	return resources, nil
}

// createExternalContactsExternalSource is used by the external_contacts_external_source resource to create Genesys cloud external source
func createExternalContactsExternalSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalContactsExternalSourceProxy(sdkConfig)

	externalContactsExternalSource, err := getExternalContactsExternalSourceFromResourceData(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to build external source request body", err)
	}

	log.Printf("Creating external source %s", *externalContactsExternalSource.Name)
	externalSource, response, err := proxy.createExternalContactsExternalSource(ctx, &externalContactsExternalSource)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get external source error: %s", err), response)
	}

	d.SetId(*externalSource.Id)
	log.Printf("Created external source %s", *externalSource.Id)
	return readExternalContactsExternalSource(ctx, d, meta)
}

// readExternalContactsExternalSource is used by the external_contacts_external_source resource to read an external source from genesys cloud
func readExternalContactsExternalSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalContactsExternalSourceProxy(sdkConfig)

	log.Printf("Reading external source %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		externalSource, response, err := proxy.getExternalContactsExternalSourceById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404ByInt(response.StatusCode) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read external source %s | error: %s", d.Id(), err), response))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read external source %s | error: %s", d.Id(), err), response))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceExternalContactsExternalSource(), 0, "")

		resourcedata.SetNillableValue(d, "name", externalSource.Name)
		resourcedata.SetNillableValue(d, "active", externalSource.Active)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "link_configuration", externalSource.LinkConfiguration, flattenLinkConfiguration)

		log.Printf("Read external source %s %s", d.Id(), *externalSource.Name)
		return cc.CheckState(d)
	})
}

// updateExternalContactsExternalSource is used by the external_contacts_external_source resource to update an external contacts external source in Genesys Cloud
func updateExternalContactsExternalSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalContactsExternalSourceProxy(sdkConfig)

	externalContactsExternalSource, err := getExternalContactsExternalSourceFromResourceData(d)

	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("failed to build external source error: %s", err), nil)
	}

	log.Printf("Updating external source %s", *externalContactsExternalSource.Name)
	externalSource, response, err := proxy.updateExternalContactsExternalSource(ctx, d.Id(), &externalContactsExternalSource)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update external source error: %s", err), response)
	}

	log.Printf("Updated external source %s", *externalSource.Id)
	return readExternalContactsExternalSource(ctx, d, meta)
}

// deleteExternalContactsExternalSource is used by the external_contacts_external_source resource to delete an external contacts external source from Genesys cloud
func deleteExternalContactsExternalSource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalContactsExternalSourceProxy(sdkConfig)

	response, err := proxy.deleteExternalContactsExternalSource(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete external source %s: %s", d.Id(), err), response)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, response, err := proxy.getExternalContactsExternalSourceById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404ByInt(response.StatusCode) {
				log.Printf("Deleted external source %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting external source %s: %s", d.Id(), err), response))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("external source still exists %s:", d.Id()), response))
	})
}
