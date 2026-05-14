package tfexporter

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	architectFlow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/errors"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

type MrMoExportResponse struct {
	Config       util.JsonMap
	ResourceData *schema.ResourceData
}

// ExportForMrMoByID runs the MrMo export pipeline for a single entity without
// calling the resource type's GetResourcesFunc (i.e. without listing every
// instance in the org). It relies on the resource's own Read context to fetch
// the entity by ID, which every Genesys Cloud resource already does via the
// standard RefreshWithoutUpgrade flow inside retrieveGenesysCloudObjectInstancesForMrMo.
//
// For singleton resources (exporter.IsSingleton), this function falls back to
// the legacy ExportForMrMo path because their map key is derived from
// exporter.ExportId rather than a free-form resource ID, and the list-cost is
// trivial (one entry).
//
// If the target entity cannot be fetched (e.g. 404), an error diagnostic is
// returned rather than silently producing an empty export, which is the
// desired behavior for MrMo where we expect the entity to exist.
func (g *GenesysCloudResourceExporter) ExportForMrMoByID(resType, resourceId string, generateOutputFiles bool, exporter *resourceExporter.ResourceExporter) (_ *MrMoExportResponse, diags diag.Diagnostics) {
	if exporter != nil && exporter.IsSingleton {
		log.Printf("ExportForMrMoByID: resource type %s is a singleton; falling back to ExportForMrMo", resType)
		return g.ExportForMrMo(resType, resourceId, generateOutputFiles, exporter)
	}

	exporters := map[string]*resourceExporter.ResourceExporter{resType: exporter}
	g.exporters = &exporters

	if resType == architectFlow.ResourceType {
		var flowExporterToUse *resourceExporter.ResourceExporter
		if generateOutputFiles {
			log.Printf("Replaced flow exporter with new exporter")
			flowExporterToUse = resourceExporter.GetNewFlowResourceExporter()
		} else {
			log.Printf("Using legacy flow exporter")
			flowExporterToUse = architectFlow.ArchitectFlowExporter()
		}
		flowExporterToUse.SetSanitizedResourceMap(exporter.GetSanitizedResourceMap())
		(*g.exporters)[resType] = flowExporterToUse
	}

	// Seed a single-entry SanitizedResourceMap so retrieveGenesysCloudObjectInstancesForMrMo
	// has an ID to iterate. BlockLabel is temporarily the ID; we backfill it
	// from the refreshed state below so the exported HCL/JSON block label and
	// state address match what ExportForMrMo would produce.
	activeExporter := (*g.exporters)[resType]
	activeExporter.SanitizedResourceMap = resourceExporter.ResourceIDMetaMap{
		resourceId: &resourceExporter.ResourceMeta{BlockLabel: resourceId},
	}

	// Fetch the one object via the resource's Read context. This calls
	// resource.RefreshWithoutUpgrade -> readContext -> proxy.get<Resource>ById.
	// It is the only Genesys Cloud API call we actually need.
	diags = append(diags, g.retrieveGenesysCloudObjectInstancesForMrMo()...)
	if diags.HasError() {
		return nil, diags
	}

	// If the resource was not found (404/410), getResourceState returns
	// (nil, nil) and the resource is silently skipped. For MrMo we treat this
	// as a hard error: the caller explicitly named an entity that must exist.
	//
	// The error message deliberately matches the legacy ExportForMrMo format
	// ("Resource <id> not found") because downstream consumers (e.g. the
	// MrMo replicator) inspect the error string to detect the delete-event
	// no-op case. Keeping the two paths' errors identical preserves behavior.
	if len(g.resources) == 0 {
		diags = append(diags, diag.Errorf("Resource %s not found", resourceId)...)
		return nil, diags
	}

	// Backfill a human-readable BlockLabel from the refreshed state so the
	// exported block label and state address are identical to what
	// ExportForMrMo would produce.
	for _, r := range g.resources {
		if r.State == nil {
			continue
		}
		if name, ok := r.State.Attributes["name"]; ok && name != "" {
			meta := activeExporter.SanitizedResourceMap[resourceId]
			if meta != nil {
				meta.BlockLabel = name
				meta.OriginalLabel = name
			}
			r.BlockLabel = name
			r.OriginalLabel = name
		}
	}

	diags = append(diags, g.buildAndExportDependsOnResourcesForFlows()...)
	if diags.HasError() {
		return nil, diags
	}

	diags = append(diags, g.buildResourceConfigMap()...)
	if diags.HasError() {
		return nil, diags
	}

	diags = append(diags, g.buildAndExportDependentResources()...)
	if diags.HasError() {
		return nil, diags
	}

	if generateOutputFiles {
		diags = append(diags, g.generateOutputFiles()...)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &MrMoExportResponse{
		Config: util.JsonMap{
			"resource": g.resourceTypesMaps,
		},
		ResourceData: g.resourceExportedForMrMo,
	}, diags
}

func (g *GenesysCloudResourceExporter) ExportForMrMo(resType, resourceId string, generateOutputFiles bool, exporter *resourceExporter.ResourceExporter) (_ *MrMoExportResponse, diags diag.Diagnostics) {
	exporters := make(map[string]*resourceExporter.ResourceExporter)
	exporters[resType] = exporter
	g.exporters = &exporters

	// Step #2 Retrieve all the individual resources we are going to export
	diags = append(diags, g.retrieveSanitizedResourceMapsForMrMo()...)
	if diags.HasError() {
		return nil, diags
	}

	if resType == architectFlow.ResourceType {
		var flowExporterToUse *resourceExporter.ResourceExporter
		if generateOutputFiles {
			// Use archy export service to download the flow config file (create, update)
			log.Printf("Replaced flow exporter with new exporter")
			flowExporterToUse = resourceExporter.GetNewFlowResourceExporter()
		} else {
			// Use legacy exporter to be more efficient (more appropriate for flow deletion events that don't only require the ResourceData be exported)
			log.Printf("Using legacy flow exporter")
			flowExporterToUse = architectFlow.ArchitectFlowExporter()
		}
		flowExporterToUse.SetSanitizedResourceMap(exporter.GetSanitizedResourceMap())
		(*g.exporters)[resType] = flowExporterToUse
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

	if generateOutputFiles {
		// Step #7 Write the terraform state file along with either the HCL or JSON
		diags = append(diags, g.generateOutputFiles()...)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &MrMoExportResponse{
		Config: util.JsonMap{
			"resource": g.resourceTypesMaps,
		},
		ResourceData: g.resourceExportedForMrMo,
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
		log.Printf("Getting exported resources for [%s]", resType)
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

		if !errors.ContainsPermissionsErrorOnly(err) {
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
