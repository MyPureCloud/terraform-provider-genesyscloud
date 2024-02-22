package outbound_filespecificationtemplate

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

func getAllFileSpecificationTemplates(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getOutboundFilespecificationtemplateProxy(clientConfig)

	fileSpecificationTemplates, err := proxy.getAllOutboundFilespecificationtemplate(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get outbound file specification templates: %v", err)
	}

	for _, fst := range *fileSpecificationTemplates {
		resources[*fst.Id] = &resourceExporter.ResourceMeta{Name: *fst.Name}
	}

	return resources, nil
}

func createOutboundFileSpecificationTemplate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkFileSpecificationTemplate := getFilespecificationtemplateFromResourceData(d)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundFilespecificationtemplateProxy(sdkConfig)

	log.Printf("Creating File Specification Template %s", *sdkFileSpecificationTemplate.Name)
	outboundFileSpecificationTemplate, err := proxy.createOutboundFilespecificationtemplate(ctx, &sdkFileSpecificationTemplate)
	if err != nil {
		return diag.Errorf("Failed to create File Specification Template %s: %s", *sdkFileSpecificationTemplate.Name, err)
	}

	d.SetId(*outboundFileSpecificationTemplate.Id)

	log.Printf("Created File Specification Template %s %s", *outboundFileSpecificationTemplate.Name, *outboundFileSpecificationTemplate.Id)
	return readOutboundFileSpecificationTemplate(ctx, d, meta)
}

func updateOutboundFileSpecificationTemplate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkFileSpecificationTemplate := getFilespecificationtemplateFromResourceData(d)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundFilespecificationtemplateProxy(sdkConfig)

	log.Printf("Updating File Specification Template %s", *sdkFileSpecificationTemplate.Name)
	_, err := proxy.updateOutboundFilespecificationtemplate(ctx, d.Id(), &sdkFileSpecificationTemplate)
	if err != nil {
		return diag.Errorf("Failed to update File Specification Template: %s", err)
	}

	log.Printf("Updated Outbound File Specification Template %s", *sdkFileSpecificationTemplate.Name)
	return readOutboundFileSpecificationTemplate(ctx, d, meta)
}

func readOutboundFileSpecificationTemplate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundFilespecificationtemplateProxy(sdkConfig)

	log.Printf("Reading Outbound File Specification Template %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkFileSpecificationTemplate, resp, getErr := proxy.getOutboundFilespecificationtemplateById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read Outbound File Specification Template %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read Outbound File Specification Template %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundFileSpecificationTemplate())

		resourcedata.SetNillableValue(d, "name", sdkFileSpecificationTemplate.Name)
		resourcedata.SetNillableValue(d, "description", sdkFileSpecificationTemplate.Description)
		resourcedata.SetNillableValue(d, "format", sdkFileSpecificationTemplate.Format)
		resourcedata.SetNillableValue(d, "number_of_header_lines_skipped", sdkFileSpecificationTemplate.NumberOfHeadingLinesSkipped)
		resourcedata.SetNillableValue(d, "number_of_trailer_lines_skipped", sdkFileSpecificationTemplate.NumberOfTrailingLinesSkipped)
		resourcedata.SetNillableValue(d, "header", sdkFileSpecificationTemplate.Header)
		resourcedata.SetNillableValue(d, "delimiter", sdkFileSpecificationTemplate.Delimiter)
		resourcedata.SetNillableValue(d, "delimiter_value", sdkFileSpecificationTemplate.DelimiterValue)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "column_information", sdkFileSpecificationTemplate.ColumnInformation, flattenSdkOutboundFileSpecificationTemplateColumnInformationSlice)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "preprocessing_rule", sdkFileSpecificationTemplate.PreprocessingRules, flattenSdkOutboundFileSpecificationTemplatePreprocessingRulesSlice)

		log.Printf("Read Outbound File Specification Template %s %s", d.Id(), *sdkFileSpecificationTemplate.Name)
		return cc.CheckState()
	})
}

func deleteOutboundFileSpecificationTemplate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getOutboundFilespecificationtemplateProxy(sdkConfig)

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound File Specification Template")
		resp, err := proxy.deleteOutboundFilespecificationtemplate(ctx, d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound File Specification Template: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getOutboundFilespecificationtemplateById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404ByInt(resp) {
				// File Specification Template List deleted
				log.Printf("Deleted Outbound File Specification Template %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting Outbound File Specification Template %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Outbound File Specification Template %s still exists", d.Id()))
	})
}
