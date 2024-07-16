package routing_email_domain

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceRoutingEmailDomainSub(t *testing.T) {
	var (
		domainRes = "routing-domain1"
		domainId  = "terraformdomain" + strings.Replace(uuid.NewString(), "-", "", -1)
	)

	CleanupRoutingEmailDomains()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create purecloud subdomain
				Config: GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.TrueValue, // Subdomain clear
					util.NullValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "subdomain", util.TrueValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_email_domain." + domainRes,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyRoutingEmailDomainDestroyed,
	})
}

func TestAccResourceRoutingEmailDomainCustom(t *testing.T) {
	var (
		domainRes       = "routing-domain1"
		domainId        = fmt.Sprintf("terraformdomain.%s.com", strings.Replace(uuid.NewString(), "-", "", -1))
		mailFromDomain1 = "test." + domainId
	)

	CleanupRoutingEmailDomains()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create custom domain
				Config: GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue, // Subdomain
					util.NullValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "subdomain", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "mail_from_domain", ""),
				),
			},
			{
				// Update custom domain
				Config: GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue, // Subdomain
					strconv.Quote(mailFromDomain1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "subdomain", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "mail_from_domain", mailFromDomain1),
				),
			},
		},
		CheckDestroy: testVerifyRoutingEmailDomainDestroyed,
	})
}

func testVerifyRoutingEmailDomainDestroyed(state *terraform.State) error {
	routingAPI := platformclientv2.NewRoutingApi()

	diagErr := util.WithRetries(context.Background(), 180*time.Second, func() *retry.RetryError {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "genesyscloud_routing_email_domain" {
				continue
			}
			_, resp, err := routingAPI.GetRoutingEmailDomain(rs.Primary.ID)
			if err != nil {
				if util.IsStatus404(resp) {
					continue
				}
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_email_domain", fmt.Sprintf("Unexpected error: %s", err), resp))
			}

			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_routing_email_domain", fmt.Sprintf("Routing email domain %s still exists", rs.Primary.ID), resp))
		}
		return nil
	})

	if diagErr != nil {
		return fmt.Errorf(fmt.Sprintf("%v", diagErr))
	}

	// Success. All Domains destroyed
	return nil
}
