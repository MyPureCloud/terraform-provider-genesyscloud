package agentic_virtual_agent_version_publish

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
   resource_genesyscloud_agentic_virtual_agent_version_publish_schema.go holds the
   schema and registration for the publish resource.
*/

const ResourceType = "genesyscloud_agentic_virtual_agent_version_publish"

// SetRegistrar registers the resource.
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceAgenticVirtualAgentVersionPublish())
}

// ResourceAgenticVirtualAgentVersionPublish registers the Terraform resource.
func ResourceAgenticVirtualAgentVersionPublish() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Agentic Virtual Agent Version Publish. Publishes a version for testing (TestReady) or production (ProductionReady).",
		CreateContext: provider.CreateWithPooledClient(createPublish),
		ReadContext:   provider.ReadWithPooledClient(readPublish),
		DeleteContext: provider.DeleteWithPooledClient(deletePublish),
		Schema: map[string]*schema.Schema{
			"agent_id": {
				Description: "ID of the parent agentic virtual agent.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"version": {
				Description: "Version number to publish (e.g. '1.0').",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"status": {
				Description:  "Target publish status. TestReady publishes for preview testing, ProductionReady publishes for production use.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"TestReady", "ProductionReady"}, false),
			},
		},
	}
}
