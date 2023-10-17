package genesyscloud

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceRoutingEmailDomainSub(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var (
		domainRes = "routing-domain1"
		domainId  = "terraform" + strconv.Itoa(rand.Intn(1000))
	)
	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	CleanupRoutingEmailDomains()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create purecloud subdomain
				Config: GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					trueValue, // Subdomain clear
					nullValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "subdomain", trueValue),
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
	rand.Seed(time.Now().UnixNano())
	var (
		domainRes       = "routing-domain1"
		domainId        = "terraform" + strconv.Itoa(rand.Intn(1000)) + ".com"
		mailFromDomain1 = "test." + domainId
	)
	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	CleanupRoutingEmailDomains()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create custom domain
				Config: GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					falseValue, // Subdomain
					nullValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "subdomain", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "mail_from_domain", ""),
				),
			},
			{
				// Update custom domain
				Config: GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					falseValue, // Subdomain
					strconv.Quote(mailFromDomain1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "subdomain", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_domain."+domainRes, "mail_from_domain", mailFromDomain1),
				),
			},
		},
		CheckDestroy: testVerifyRoutingEmailDomainDestroyed,
	})
}

func testVerifyRoutingEmailDomainDestroyed(state *terraform.State) error {
	routingAPI := platformclientv2.NewRoutingApi()

	diagErr := WithRetries(context.Background(), 180*time.Second, func() *retry.RetryError {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "genesyscloud_routing_email_domain" {
				continue
			}
			_, resp, err := routingAPI.GetRoutingEmailDomain(rs.Primary.ID)
			if err != nil {
				if IsStatus404(resp) {
					continue
				}
				return retry.NonRetryableError(fmt.Errorf("Unexpected error: %s", err))
			}

			return retry.RetryableError(fmt.Errorf("Routing email domain %s still exists", rs.Primary.ID))
		}
		return nil
	})

	if diagErr != nil {
		return fmt.Errorf(fmt.Sprintf("%v", diagErr))
	}

	// Success. All Domains destroyed
	return nil
}
