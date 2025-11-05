package exporter

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

type Credentials struct {
	ClientId     string
	ClientSecret string
	Region       string
}

type ExportInput struct {
	// ResourceType - The resource type of the entity to export e.g. genesyscloud_flow
	ResourceType string

	// EntityId - The identifier of the entity we want to export.
	EntityId string

	// GenerateOutputFiles - If set to false, no export folders or files will be created and only the ResourceData will be returned.
	GenerateOutputFiles bool

	// IncludeStateFile - Whether or not to include the state file in the export. Irrelevant if GenerateOutputFiles is false.
	IncludeStateFile bool

	// Directory - The output directory that export data will be written to. Required if GenerateOutputFiles is true.
	Directory string
}

type ExportOutput struct {
	// ExportData is the exported data that would be written to the .tf.json file during export
	ExportData util.JsonMap

	// ExportDataPath is the path to the export directory i.e. the value of "directory" in the genesyscloud_tf_export resource config
	ExportDataPath string

	// ExportedResourceData is the *schema.ResourceData representation of the exported resesource. This is the type we can pass into
	// the create and update context functions.
	ExportedResourceData *schema.ResourceData

	// ResourceExporter is the resource exporter used. This is returned to MRMO so that it can access the RefAttrs during GUID resolution.
	ResourceExporter *resource_exporter.ResourceExporter
}
