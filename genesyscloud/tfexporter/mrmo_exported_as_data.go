package tfexporter

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

// dataSourceInstanceState returns a copy of state that carries the attributes of the data source
// block the instance is exported as. state itself is left untouched.
//
// The exporter used to assign the data source attributes to state.Attributes in place. That state
// is the same object the schema.ResourceData handed to MRMO wraps, so MRMO, which reads the
// ResourceData after the export, saw the data source attributes (often just "name") instead of the
// instance's own. Anything the provider decided from the full attributes, such as whether the
// instance is managed by Genesys Cloud, could no longer be read back from them.
func dataSourceInstanceState(exporter *resourceExporter.ResourceExporter, dataSource *schema.Resource, state *terraform.InstanceState) *terraform.InstanceState {
	attributes := make(map[string]string)
	for attr := range dataSource.SchemaMap() {
		key, val := exporter.DataResolver(state, attr)
		attributes[key] = val
	}
	out := state.DeepCopy()
	out.Attributes = attributes
	return out
}

// recordMrMoExportedAsData notes that the instance with this id was exported as a data source
// rather than a resource. id is the instance's state id, which is also schema.ResourceData.Id().
//
// The decision is the exporter's own, taken after ExportAsDataFunc ran against the instance's full
// attributes (and, for types such as outbound routes, the live API). Recording it lets MRMO use
// that decision instead of evaluating the function again.
func (g *GenesysCloudResourceExporter) recordMrMoExportedAsData(resType, id string) {
	g.resourcesExportedForMrMoMutex.Lock()
	defer g.resourcesExportedForMrMoMutex.Unlock()
	if g.mrmoExportedAsData == nil {
		g.mrmoExportedAsData = make(map[string]map[string]struct{})
	}
	if g.mrmoExportedAsData[resType] == nil {
		g.mrmoExportedAsData[resType] = make(map[string]struct{})
	}
	g.mrmoExportedAsData[resType][id] = struct{}{}
}

// mrMoExportedAsData returns the ids of the instances of resType that were exported as data
// sources. The map is a copy and is never nil.
func (g *GenesysCloudResourceExporter) mrMoExportedAsData(resType string) map[string]bool {
	g.resourcesExportedForMrMoMutex.Lock()
	defer g.resourcesExportedForMrMoMutex.Unlock()
	out := make(map[string]bool, len(g.mrmoExportedAsData[resType]))
	for id := range g.mrmoExportedAsData[resType] {
		out[id] = true
	}
	return out
}

// mrMoResourceExportedAsData reports whether data, as returned in an MRMO export response, was
// exported as a data source.
func (g *GenesysCloudResourceExporter) mrMoResourceExportedAsData(resType string, data *schema.ResourceData) bool {
	if data == nil {
		return false
	}
	return g.mrMoExportedAsData(resType)[data.Id()]
}
