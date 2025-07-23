package tfexporter

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

func (g *GenesysCloudResourceExporter) ExportForMrMo(resType string, exporter *resourceExporter.ResourceExporter, resourceId string) (_ util.JsonMap, diags diag.Diagnostics) {
	exporters := make(map[string]*resourceExporter.ResourceExporter)
	exporters[resType] = exporter
	g.exporters = &exporters

	// Step #2 Retrieve all the individual resources we are going to export
	diags = append(diags, g.retrieveSanitizedResourceMapsForMrMo()...)
	if diags.HasError() {
		return nil, diags
	}

	// Step #3 Filter out all resources from the resType exporters SanitizedResourceMap besides the one at index "{resourceId}"
	diags = append(diags, g.filterResourceMetaMapBasedOnID(resType, resourceId)...)
	if diags.HasError() {
		return nil, diags
	}

	// Step #3 Retrieve the individual genesys cloud object instance
	diags = append(diags, g.retrieveGenesysCloudObjectInstancesForMrMo()...)
	if diags.HasError() {
		return nil, diags
	}

	// Step #4 export dependent resources for the flows
	diags = append(diags, g.buildAndExportDependsOnResourcesForFlows()...)
	if diags.HasError() {
		return nil, diags
	}

	// Step #5 Convert the Genesys Cloud resources to neutral format (e.g. map of maps)
	diags = append(diags, g.buildResourceConfigMap()...)
	if diags.HasError() {
		return nil, diags
	}

	// Step #6 export dependents for other resources
	diags = append(diags, g.buildAndExportDependentResources()...)
	if diags.HasError() {
		return nil, diags
	}

	return util.JsonMap{
		"resource": g.resourceTypesMaps,
	}, diags
}

// filterResourceMetaMapBasedOnID works on behalf of Mr Mo to filter SanitizedResourceMap to only include the ResourceIDMetaMap at key entityID.
// In other words, it filters the resource map so that it only includes the entity Mr Mo is handling.
func (g *GenesysCloudResourceExporter) filterResourceMetaMapBasedOnID(resType, entityId string) (diags diag.Diagnostics) {
	defer func() {
		if r := recover(); r != nil {
			diags = diag.Errorf("Recovered in filterResourceMetaMapBasedOnID: %v", r)
		}
	}()

	exporter := (*g.exporters)[resType]

	if rm, ok := exporter.SanitizedResourceMap[entityId]; ok {
		exporter.SanitizedResourceMap = resourceExporter.ResourceIDMetaMap{
			entityId: rm,
		}
	} else {
		diags = append(diags, diag.Errorf("Resource %s not found", entityId)...)
		return
	}

	(*g.exporters)[resType] = exporter

	return diags
}

// retrieveSanitizedResourceMapsForMrMo will retrieve a list of all resources to be exported.
func (g *GenesysCloudResourceExporter) retrieveSanitizedResourceMapsForMrMo() (diagErr diag.Diagnostics) {
	log.Printf("Retrieving map of Genesys Cloud resources to export")
	var filter []string
	if exportableResourceTypes, ok := g.d.GetOk("resource_types"); ok {
		filter = lists.InterfaceListToStrings(exportableResourceTypes.([]any))
	}

	if exportableResourceTypes, ok := g.d.GetOk("include_filter_resources"); ok {
		filter = lists.InterfaceListToStrings(exportableResourceTypes.([]any))
	}

	if exportableResourceTypes, ok := g.d.GetOk("exclude_filter_resources"); ok {
		filter = lists.InterfaceListToStrings(exportableResourceTypes.([]any))
	}

	newFilter := make([]string, 0)
	for _, f := range filter {
		if strings.Contains(f, "::") {
			newFilter = append(newFilter, f)
		}
	}

	//Retrieve a map of all objects we are going to build.  Apply the filter that will remove specific classes of an object
	log.Println("Building sanitized resource maps")
	diagErr = g.buildSanitizedResourceMapsForMrMo(*g.exporters, newFilter, g.logPermissionErrors)
	if diagErr.HasError() {
		return diagErr
	}

	//Check to see if we found any exporters.  If we did find the exporter
	if len(*g.exporters) == 0 {
		diagErr = append(diagErr, diag.Errorf("No valid resource types to export.")...)
		return diagErr
	}

	return diagErr
}

// retrieveGenesysCloudObjectInstancesForMrMo will take a list the exporter and then return the actual terraform Genesys Cloud data
func (g *GenesysCloudResourceExporter) retrieveGenesysCloudObjectInstancesForMrMo() diag.Diagnostics {
	log.Printf("Retrieving Genesys Cloud objects from Genesys Cloud")
	for resType, exporter := range *g.exporters {
		log.Printf("Getting exported resources for [%s] o0o", resType)
		typeResources, err := g.getResourcesForType(resType, g.provider, exporter, g.meta)
		if err != nil {
			return err
		}
		g.resources = append(g.resources, typeResources...)
	}
	return nil
}

func (g *GenesysCloudResourceExporter) buildSanitizedResourceMapsForMrMo(exporters map[string]*resourceExporter.ResourceExporter, filter []string, logErrors bool) diag.Diagnostics {
	for resourceType, exporter := range exporters {
		log.Printf("Getting all resources for type %s", resourceType)
		exporter.FilterResource = g.resourceFilter

		err := exporter.LoadSanitizedResourceMap(g.ctx, resourceType, filter)
		if err == nil {
			log.Printf("Found %d resources for type %s", len(exporter.SanitizedResourceMap), resourceType)
			continue
		}

		if !containsPermissionsErrorOnly(err) {
			return err
		}

		if logErrors {
			log.Println(err)
			log.Printf("Logging permission error for %s. Resuming export...", resourceType)
			continue
		}
		return err
	}
	return nil
}

func WriteConfigForMrMo(jsonMap map[string]interface{}, path string) (diags diag.Diagnostics) {
	sortedJsonMap := sortJSONMap(jsonMap)
	dataJSONBytes, err := json.MarshalIndent(sortedJsonMap, "", "  ")
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
		return
	}

	log.Printf("Writing export config file to %s", path)
	diags = append(diags, files.WriteToFile(postProcessJsonBytes(dataJSONBytes), path)...)
	return
}
