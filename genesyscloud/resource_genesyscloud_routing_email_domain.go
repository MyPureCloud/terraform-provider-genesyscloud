package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v77/platformclientv2"
)

func getAllRoutingEmailDomains(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	domains, _, getErr := routingAPI.GetRoutingEmailDomains(false)
	if getErr != nil {
		return nil, diag.Errorf("Failed to get routing email domains: %v", getErr)
	}

	if domains.Entities == nil || len(*domains.Entities) == 0 {
		return resources, nil
	}

	for _, domain := range *domains.Entities {
		resources[*domain.Id] = &ResourceMeta{Name: *domain.Id}
	}

	return resources, nil
}

func routingEmailDomainExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllRoutingEmailDomains),
		UnResolvableAttributes: map[string]*schema.Schema{
			"custom_smtp_server_id": resourceRoutingEmailDomain().Schema["custom_smtp_server_id"],
		},
	}
}

func resourceRoutingEmailDomain() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Email Domain",

		CreateContext: createWithPooledClient(createRoutingEmailDomain),
		ReadContext:   readWithPooledClient(readRoutingEmailDomain),
		UpdateContext: updateWithPooledClient(updateRoutingEmailDomain),
		DeleteContext: deleteWithPooledClient(deleteRoutingEmailDomain),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Description: "Unique Id of the domain such as: 'example.com'. If subdomain is true, the Genesys Cloud regional domain is appended.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"subdomain": {
				Description: "Indicates if this a Genesys Cloud sub-domain. If true, then the appropriate DNS records are created for sending/receiving email.",
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	sdkDomain := platformclientv2.Inbounddomain{
		Id:             &domainID,
		SubDomain:      &subdomain,
		MxRecordStatus: &mxRecordStatus,
	}

	log.Printf("Creating routing email domain %s", domainID)
	domain, _, err := routingAPI.PostRoutingEmailDomains(sdkDomain)
	if err != nil {
		return diag.Errorf("Failed to create routing email domain %s: %s", domainID, err)
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading routing email domain %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		domain, resp, getErr := routingAPI.GetRoutingEmailDomain(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read routing email domain %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read routing email domain %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceRoutingEmailDomain())
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
		return cc.CheckState()
	})
}

func updateRoutingEmailDomain(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	customSMTPServer := d.Get("custom_smtp_server_id").(string)
	mailFromDomain := d.Get("mail_from_domain").(string)
	domainID := d.Get("domain_id").(string)

	if !strings.Contains(mailFromDomain, domainID) || mailFromDomain == domainID {
		return diag.Errorf("domain_id must be a subdomain of mail_from_domain")
	}

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Updating routing email domain %s", d.Id())

	_, _, err := routingAPI.PatchRoutingEmailDomain(d.Id(), platformclientv2.Inbounddomainpatchrequest{
		MailFromSettings: &platformclientv2.Mailfromresult{
			MailFromDomain: &mailFromDomain,
		},
		CustomSMTPServer: &platformclientv2.Domainentityref{
			Id: &customSMTPServer,
		},
	})
	if err != nil {
		return diag.Errorf("Failed to update routing email domain %s: %s", d.Id(), err)
	}

	log.Printf("Updated routing email domain %s", d.Id())
	return readRoutingEmailDomain(ctx, d, meta)
}

func deleteRoutingEmailDomain(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting routing email domain %s", d.Id())
	_, err := routingAPI.DeleteRoutingEmailDomain(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete routing email domain %s: %s", d.Id(), err)
	}

	return withRetries(ctx, 90*time.Second, func() *resource.RetryError {
		_, resp, err := routingAPI.GetRoutingEmailDomain(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// Routing email domain deleted
				log.Printf("Deleted Routing email domain %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Routing email domain %s: %s", d.Id(), err))
		}

		routingAPI.DeleteRoutingEmailDomain(d.Id())
		return resource.RetryableError(fmt.Errorf("Routing email domain %s still exists", d.Id()))
	})
}
