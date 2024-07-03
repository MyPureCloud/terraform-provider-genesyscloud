package routing_email_domain

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const resourceName = "genesyscloud_routing_email_domain"

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceRoutingEmailDomain())
	regInstance.RegisterDataSource(resourceName, DataSourceRoutingEmailDomain())
	regInstance.RegisterExporter(resourceName, RoutingEmailDomainExporter())
}

func ResourceRoutingEmailDomain() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Email Domain",

		CreateContext: provider.CreateWithPooledClient(createRoutingEmailDomain),
		ReadContext:   provider.ReadWithPooledClient(readRoutingEmailDomain),
		UpdateContext: provider.UpdateWithPooledClient(updateRoutingEmailDomain),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingEmailDomain),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Description: "Unique Id of the domain such as: 'example.com'. If subdomain is true, the Genesys Cloud regional domain is appended. Changing the domain_id attribute will cause the routing_email_domain to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"subdomain": {
				Description: "Indicates if this a Genesys Cloud sub-domain. If true, then the appropriate DNS records are created for sending/receiving email. Changing the subdomain attribute will cause the routing_email_domain to be dropped and recreated with a new ID.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
			"mail_from_domain": {
				Description: "The custom MAIL FROM domain. This must be a subdomain of your email domain",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"custom_smtp_server_id": {
				Description: "The ID of the custom SMTP server integration to use when sending outbound emails from this domain.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

// Returns the schema for the routing email domain
func DataSourceRoutingEmailDomain() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Email Domains. Select an email domain by name",
		ReadContext: provider.ReadWithPooledClient(DataSourceRoutingEmailDomainRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Email domain name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func RoutingEmailDomainExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingEmailDomains),
		UnResolvableAttributes: map[string]*schema.Schema{
			"custom_smtp_server_id": ResourceRoutingEmailDomain().Schema["custom_smtp_server_id"],
		},
	}
}
