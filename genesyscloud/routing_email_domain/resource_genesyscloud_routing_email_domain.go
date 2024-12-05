package routing_email_domain

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllRoutingEmailDomains(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getRoutingEmailDomainProxy(clientConfig)

	domains, resp, getErr := proxy.getAllRoutingEmailDomains(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get routing email domains error: %s", getErr), resp)
	}

	if domains == nil || len(*domains) == 0 {
		return resources, nil
	}

	for _, domain := range *domains {
		resources[*domain.Id] = &resourceExporter.ResourceMeta{BlockLabel: *domain.Id}
	}
	return resources, nil
}

func createRoutingEmailDomain(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailDomainProxy(sdkConfig)

	domainID := d.Get("domain_id").(string)
	subdomain := d.Get("subdomain").(bool)
	mxRecordStatus := "VALID"
	if !subdomain {
		mxRecordStatus = "NOT_AVAILABLE"
	}

	sdkDomain := platformclientv2.Inbounddomain{
		Id:             &domainID,
		SubDomain:      &subdomain,
		MxRecordStatus: &mxRecordStatus,
	}

	log.Printf("Creating routing email domain %s", domainID)
	domain, resp, err := proxy.createRoutingEmailDomain(ctx, &sdkDomain)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create routing email domain %s error: %s", domainID, err), resp)
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
	proxy := getRoutingEmailDomainProxy(sdkConfig)

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingEmailDomain(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading routing email domain %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		domain, resp, getErr := proxy.getRoutingEmailDomainById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read routing email domain %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read routing email domain %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "subdomain", domain.SubDomain)
		resourcedata.SetNillableReference(d, "custom_smtp_server_id", domain.CustomSMTPServer)

		if domain.SubDomain != nil && *domain.SubDomain {
			// Strip off the regional domain suffix added by the server
			_ = d.Set("domain_id", strings.SplitN(*domain.Id, ".", 2)[0])
		} else {
			_ = d.Set("domain_id", *domain.Id)
		}

		if domain.MailFromSettings != nil && domain.MailFromSettings.MailFromDomain != nil {
			_ = d.Set("mail_from_domain", *domain.MailFromSettings.MailFromDomain)
		} else {
			_ = d.Set("mail_from_domain", nil)
		}

		log.Printf("Read routing email domain %s", d.Id())
		return cc.CheckState(d)
	})
}

func updateRoutingEmailDomain(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailDomainProxy(sdkConfig)

	customSMTPServer := d.Get("custom_smtp_server_id").(string)
	mailFromDomain := d.Get("mail_from_domain").(string)
	domainID := d.Get("domain_id").(string)

	if !strings.Contains(mailFromDomain, domainID) || mailFromDomain == domainID {
		return util.BuildDiagnosticError(ResourceType, "domain_id must be a subdomain of mail_from_domain", fmt.Errorf("domain_id must be a subdomain of mail_from_domain"))
	}

	log.Printf("Updating routing email domain %s", d.Id())

	_, resp, err := proxy.updateRoutingEmailDomain(ctx, d.Id(), &platformclientv2.Inbounddomainpatchrequest{
		MailFromSettings: &platformclientv2.Mailfromresult{
			MailFromDomain: &mailFromDomain,
		},
		CustomSMTPServer: &platformclientv2.Domainentityref{
			Id: &customSMTPServer,
		},
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update routing email domain %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated routing email domain %s", d.Id())
	return readRoutingEmailDomain(ctx, d, meta)
}

func deleteRoutingEmailDomain(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailDomainProxy(sdkConfig)

	log.Printf("Deleting routing email domain %s", d.Id())
	resp, err := proxy.deleteRoutingEmailDomain(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete routing email domain %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 90*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getRoutingEmailDomainById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Routing email domain %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Routing email domain %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Routing email domain %s still exists", d.Id()), resp))
	})
}

func GenerateRoutingEmailDomainResource(
	resourceLabel string,
	domainID string,
	subdomain string,
	fromDomain string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_email_domain" "%s" {
		domain_id = "%s"
		subdomain = %s
        mail_from_domain = %s
	}
	`, resourceLabel, domainID, subdomain, fromDomain)
}
