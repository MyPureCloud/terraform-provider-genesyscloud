package exporter

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	providerRegistrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

type Credentials struct {
	ClientId     string
	ClientSecret string
	Region       string
}

func Export(ctx context.Context, resourceType, resourceId string, creds Credentials) (_ util.JsonMap, diags diag.Diagnostics) {
	providerMeta, err := getProviderConfig(creds)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	log.Println("Activating MRMO")
	mrmo.Activate(providerMeta.ClientConfig)

	log.Println("Initialising provider resource maps")
	_, _ = providerRegistrar.GetProviderResources()

	log.Println("Creating genesyscloud_tf_export resource config")
	exportResourceConfig := createExportResourceData(tfexporter.ResourceTfExport().Schema, tfexporter.ResourceType)

	log.Println("Creating the genesyscloud resource exporter")
	gcResourceExporter, newExporterDiags := tfexporter.NewGenesysCloudResourceExporter(ctx, exportResourceConfig, providerMeta, tfexporter.IncludeResources)
	if newExporterDiags.HasError() {
		return nil, diag.Errorf("error creating genesyscloud resource exporter: %v", newExporterDiags)
	}

	log.Println("Getting the resource exporter by resource type")
	exporter := providerRegistrar.GetResourceExporterByResourceType(resourceType)

	log.Printf("Exporting %s resource. ID: '%s'", resourceType, resourceId)
	config, exportDiags := gcResourceExporter.ExportForMrMo(resourceType, exporter, resourceId)
	if exportDiags.HasError() {
		return nil, diag.Errorf("error exporting resource: %v", exportDiags)
	}

	return config, nil
}

func getProviderConfig(creds Credentials) (_ *provider.ProviderMeta, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("getProviderConfig: %w", err)
		}
	}()

	config := platformclientv2.GetDefaultConfiguration()
	config.BasePath = provider.GetRegionBasePath(creds.Region)

	err = config.AuthorizeClientCredentials(creds.ClientId, creds.ClientSecret)
	if err != nil {
		return nil, err
	}

	return &provider.ProviderMeta{
		ClientConfig: config,
	}, nil
}

// createExportResourceData generates the export resource config that the genesyscloud tf exporter will use
func createExportResourceData(s map[string]*schema.Schema, resType string) *schema.ResourceData {
	config := map[string]any{
		"directory":                os.TempDir(),
		"include_state_file":       true,
		"export_format":            "json",
		"include_filter_resources": []any{resType},
	}

	var t testing.T
	return schema.TestResourceDataRaw(&t, s, config)
}
