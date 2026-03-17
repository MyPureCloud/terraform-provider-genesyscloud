package exporter

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	providerRegistrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter"
)

// Export takes the resource type and resource ID of and exports that resource.
//
// It returns the exported data that would be written to the .tf.json file during export, the schema.ResourceData object representing the resource
// (this can be passed to the C/U/D context functions), and a diagnostics object which can contain errors or warnings.
func Export(ctx context.Context, input ExportInput, clientConfig *platformclientv2.Configuration) (resp *ExportOutput, diags diag.Diagnostics) {
	if err := validateExportInput(input); err != nil {
		return nil, diag.FromErr(err)
	}

	generateDefaults(&input)

	log.Println("Activating MRMO")
	mrmo.Activate(clientConfig)

	providerMeta := &provider.ProviderMeta{ClientConfig: clientConfig}

	log.Println("Initialising provider resource maps")
	_, _ = providerRegistrar.GetProviderResources()

	log.Printf("Creating %s resource config", tfexporter.ResourceType)
	exportResourceConfig := createExportResourceData(tfexporter.ResourceTfExport().Schema, input)

	log.Println("Creating the genesyscloud resource exporter")
	gcResourceExporter, newExporterDiags := tfexporter.NewGenesysCloudResourceExporter(ctx, exportResourceConfig, providerMeta, tfexporter.IncludeResources)
	if newExporterDiags != nil {
		diags = append(diags, newExporterDiags...)
	}
	if diags.HasError() {
		log.Printf("Caught error after defining genesys cloud resource exporter: %v", diags)
		return nil, diags
	}

	log.Println("Getting the resource exporter by resource type")
	exporter := providerRegistrar.GetResourceExporterByResourceType(input.ResourceType)

	log.Printf("Exporting %s resource to '%s'. ID: '%s'", input.ResourceType, input.Directory, input.EntityId)
	exportResponse, exportDiags := gcResourceExporter.ExportForMrMo(input.ResourceType, input.EntityId, input.GenerateOutputFiles, exporter)
	if exportDiags != nil {
		diags = append(diags, exportDiags...)
	}

	if diags.HasError() {
		log.Printf("Error returned from ExportForMrMo: %v", diags)
		return nil, diags
	}

	return &ExportOutput{
		ExportData:           exportResponse.Config,
		ExportDataPath:       input.Directory,
		ExportedResourceData: exportResponse.ResourceData,
		ResourceExporter:     exporter,
	}, diags
}
