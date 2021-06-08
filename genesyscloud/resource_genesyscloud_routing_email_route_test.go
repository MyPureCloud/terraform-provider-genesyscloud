package genesyscloud

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v45/platformclientv2"
)

func TestAccResourceRoutingEmailRoute(t *testing.T) {
	var (
		domainRes     = "routing-domain1"
		domainId      = "tfroutetest" + strconv.Itoa(rand.Intn(1000)) + ".com"
		queueResource = "email-queue"
		queueName     = "Terraform Email Queue-" + uuid.NewString()
		langResource  = "email-lang"
		langName      = "tflang" + uuid.NewString()
		skillResource = "test-skill1"
		skillName     = "Terraform Skill" + uuid.NewString()
		routeRes      = "email-route1"
		routeRes2     = "email-route2"
		routePattern1 = "terraform1"
		routePattern2 = "terraform2"
		routePattern3 = "terraform3"
		fromEmail1    = "terraform1@test.com"
		fromEmail2    = "terraform2@test.com"
		fromName1     = "John Terraform"
		fromName2     = "Jane Terraform"
		priority1     = "1"
		bccEmail1     = "test1@" + domainId
		bccEmail2     = "test2@" + domainId
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create email domain and basic route
				Config: generateRoutingEmailDomainResource(
					domainRes,
					domainId,
					falseValue,
					nullValue,
				) + generateRoutingEmailRouteResource(
					routeRes,
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern1,
					fromName1,
					fromEmail1,
					generateRoutingAutoBcc(fromName1, bccEmail1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "pattern", routePattern1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "from_name", fromName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "from_email", fromEmail1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "auto_bcc.0.name", fromName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "auto_bcc.0.email", bccEmail1),
				),
			},
			{
				// Update email route and add a queue, language, and skill
				Config: generateRoutingEmailDomainResource(
					domainRes,
					domainId,
					falseValue,
					nullValue,
				) + generateRoutingQueueResourceBasic(
					queueResource,
					queueName,
				) + generateRoutingLanguageResource(
					langResource,
					langName,
				) + generateRoutingSkillResource(
					skillResource,
					skillName,
				) + generateRoutingEmailRouteResource(
					routeRes,
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern2,
					fromName2,
					fromEmail2,
					generateRoutingAutoBcc(fromName2, bccEmail2),
					generateRoutingReplyEmail(
						"genesyscloud_routing_email_domain."+domainRes+".id",
						"genesyscloud_routing_email_route."+routeRes2+".id",
					),
					generateRoutingEmailQueueSettings(
						"genesyscloud_routing_queue."+queueResource+".id",
						priority1,
						"genesyscloud_routing_language."+langResource+".id",
						"genesyscloud_routing_skill."+skillResource+".id",
					),
				) + generateRoutingEmailRouteResource( // Second route to use as the reply_email_address
					routeRes2,
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern3,
					fromName1,
					fromEmail1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "pattern", routePattern2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "from_name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "from_email", fromEmail2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "auto_bcc.0.name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "auto_bcc.0.email", bccEmail2),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "queue_id", "genesyscloud_routing_queue."+queueResource, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "language_id", "genesyscloud_routing_language."+langResource, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "skill_ids.0", "genesyscloud_routing_skill."+skillResource, "id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "priority", priority1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "reply_email_address.0.domain_id", domainId),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "reply_email_address.0.route_id", "genesyscloud_routing_email_route."+routeRes2, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:        "genesyscloud_routing_email_route." + routeRes,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: domainId + "/",
			},
		},
		CheckDestroy: testVerifyRoutingEmailRouteDestroyed,
	})
}

func generateRoutingEmailRouteResource(
	resourceID string,
	domainID string,
	pattern string,
	fromName string,
	fromEmail string,
	otherAttrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_email_route" "%s" {
            domain_id = %s
            pattern = "%s"
            from_name = "%s"
            from_email = "%s"
            %s
        }
        `, resourceID, domainID, pattern, fromName, fromEmail, strings.Join(otherAttrs, "\n"))
}

func generateRoutingEmailQueueSettings(
	queueId string,
	priority string,
	languageId string,
	skillIds ...string) string {
	return fmt.Sprintf(`
        queue_id = %s
        priority = %s
        language_id = %s
        skill_ids = [%s]
	`, queueId, priority, languageId, strings.Join(skillIds, ","))
}

func generateRoutingAutoBcc(
	name string,
	email string) string {
	return fmt.Sprintf(`
        auto_bcc {
            name = "%s"
            email = "%s"
        }
	`, name, email)
}

func generateRoutingReplyEmail(
	domainID string,
	routeID string) string {
	return fmt.Sprintf(`
        reply_email_address {
            domain_id = %s
            route_id = %s
        }
	`, domainID, routeID)
}

func testVerifyRoutingEmailRouteDestroyed(state *terraform.State) error {
	routingAPI := platformclientv2.NewRoutingApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_routing_email_route" {
			continue
		}

		var route *platformclientv2.Inboundroute
		for pageNum := 1; ; pageNum++ {
			routes, resp, getErr := routingAPI.GetRoutingEmailDomainRoutes(rs.Primary.Attributes["domain_id"], 100, pageNum, "")
			if getErr != nil {
				if resp != nil && resp.StatusCode == 404 {
					// Domain not found
					continue
				}
				return fmt.Errorf("Failed to get page of email routes for domain %s: %v", rs.Primary.Attributes["domain_id"], getErr)
			}

			if routes.Entities == nil || len(*routes.Entities) == 0 {
				break
			}

			for _, queryRoute := range *routes.Entities {
				if queryRoute.Id != nil && *queryRoute.Id == rs.Primary.ID {
					route = &queryRoute
					break
				}
			}
		}

		if route != nil {
			return fmt.Errorf("Route (%s) still exists", rs.Primary.ID)
		} else {
			// Route not found as expected
			continue
		}
	}
	// Success. All routes destroyed
	return nil
}
