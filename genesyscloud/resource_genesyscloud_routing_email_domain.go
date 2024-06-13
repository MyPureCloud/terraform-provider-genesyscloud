package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
)

func getAllRoutingEmailDomains(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100

		domains, resp, getErr := routingAPI.GetRoutingEmailDomains(pageSize, pageNum, false, "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_routing_email_domain", fmt.Sprintf("Failed to get routing email domains error: %s", getErr), resp)
		}

		if domains.Entities == nil || len(*domains.Entities) == 0 {
			return resources, nil
		}

		for _, domain := range *domains.Entities {

			resources[*domain.Id] = &resourceExporter.ResourceMeta{Name: *domain.Id}
		}
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

func createRoutingEmailDomain(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainID := d.Get("domain_id").(string)
	subdomain := d.Get("subdomain").(bool)
	mxRecordStatus := "VALID"
	if !subdomain {
		mxRecordStatus = "NOT_AVAILABLE"
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	sdkDomain := platformclientv2.Inbounddomain{
		Id:             &domainID,
		SubDomain:      &subdomain,
		MxRecordStatus: &mxRecordStatus,
	}

	log.Printf("Creating routing email domain %s", domainID)
	domain, resp, err := routingAPI.PostRoutingEmailDomains(sdkDomain)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_email_domain", fmt.Sprintf("Failed to create routing email domain %s error: %s", domainID, err), resp)
	}

	d.SetId(*domain.Id)
	log.Printf("Created routing email domain %s", *domain.Id)

	// Other settings must be updated in a PATCH update
	if d.HasChanges("mail_from_domain", "custom_smtp_server_id") {
		return updateRoutingEmailDomain(ctx, d, meta)
	} else {
		return readRoutingEmailDomain(ctx, d, meta)
	}
}

func readRoutingEmailDomain(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingEmailDomain(), constants.DefaultConsistencyChecks, "genesyscloud_routing_email_domain")

	log.Printf("Reading routing email domain %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		domain, resp, getErr := routingAPI.GetRoutingEmailDomain(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_email_domain", fmt.Sprintf("Failed to read routing email domain %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_email_domain", fmt.Sprintf("Failed to read routing email domain %s | error: %s", d.Id(), getErr), resp))
		}

		if domain.SubDomain != nil && *domain.SubDomain {
			// Strip off the regional domain suffix added by the server
			d.Set("domain_id", strings.SplitN(*domain.Id, ".", 2)[0])
		} else {
			d.Set("domain_id", *domain.Id)
		}

		if domain.SubDomain != nil {
			d.Set("subdomain", *domain.SubDomain)
		} else {
			d.Set("subdomain", nil)
		}

		if domain.CustomSMTPServer != nil && domain.CustomSMTPServer.Id != nil {
			d.Set("custom_smtp_server_id", *domain.CustomSMTPServer.Id)
		} else {
			d.Set("custom_smtp_server_id", nil)
		}

		if domain.MailFromSettings != nil && domain.MailFromSettings.MailFromDomain != nil {
			d.Set("mail_from_domain", *domain.MailFromSettings.MailFromDomain)
		} else {
			d.Set("mail_from_domain", nil)
		}

		log.Printf("Read routing email domain %s", d.Id())
		return cc.CheckState(d)
	})
}

func updateRoutingEmailDomain(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	customSMTPServer := d.Get("custom_smtp_server_id").(string)
	mailFromDomain := d.Get("mail_from_domain").(string)
	domainID := d.Get("domain_id").(string)

	if !strings.Contains(mailFromDomain, domainID) || mailFromDomain == domainID {
		return util.BuildDiagnosticError("genesyscloud_routing_email_domain", fmt.Sprintf("domain_id must be a subdomain of mail_from_domain"), fmt.Errorf("domain_id must be a subdomain of mail_from_domain"))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Updating routing email domain %s", d.Id())

	_, resp, err := routingAPI.PatchRoutingEmailDomain(d.Id(), platformclientv2.Inbounddomainpatchrequest{
		MailFromSettings: &platformclientv2.Mailfromresult{
			MailFromDomain: &mailFromDomain,
		},
		CustomSMTPServer: &platformclientv2.Domainentityref{
			Id: &customSMTPServer,
		},
	})
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_email_domain", fmt.Sprintf("Failed to update routing email domain %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated routing email domain %s", d.Id())
	return readRoutingEmailDomain(ctx, d, meta)
}

func deleteRoutingEmailDomain(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting routing email domain %s", d.Id())
	resp, err := routingAPI.DeleteRoutingEmailDomain(d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_routing_email_domain", fmt.Sprintf("Failed to delete routing email domain %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 90*time.Second, func() *retry.RetryError {
		_, resp, err := routingAPI.GetRoutingEmailDomain(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Routing email domain deleted
				log.Printf("Deleted Routing email domain %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_email_domain", fmt.Sprintf("Error deleting Routing email domain %s | error: %s", d.Id(), err), resp))
		}

		routingAPI.DeleteRoutingEmailDomain(d.Id())
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_email_domain", fmt.Sprintf("Routing email domain %s still exists", d.Id()), resp))
	})
}

func GenerateRoutingEmailDomainResource(
	resourceID string,
	domainID string,
	subdomain string,
	fromDomain string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_email_domain" "%s" {
		domain_id = "%s"
		subdomain = %s
        mail_from_domain = %s
	}
	`, resourceID, domainID, subdomain, fromDomain)
}
