package routing_email_route

import (
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailDomain "terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingLanguage "terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingSkill "terraform-provider-genesyscloud/genesyscloud/routing_skill"

	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func TestAccResourceRoutingEmailRoute(t *testing.T) {
	var (
		domainResourceLabel = "routing-domain1"
		domainId            = fmt.Sprintf("terraformroutes.%s.com", strings.Replace(uuid.NewString(), "-", "", -1))
		queueResourceLabel  = "email-queue"
		queueName           = "Terraform Email Queue-" + uuid.NewString()
		langResourceLabel   = "email-lang"
		langName            = "tflang" + uuid.NewString()
		skillResourceLabel  = "test-skill1"
		skillName           = "Terraform Skill" + uuid.NewString()
		routeResourceLabel1 = "email-route1"
		routeResourceLabel2 = "email-route2"
		routePattern1       = "terraform1"
		routePattern2       = "terraform2"
		routePattern3       = "terraform3"
		fromEmail1          = "terraform1@test.com"
		fromEmail2          = "terraform2@test.com"
		fromName1           = "John Terraform"
		fromName2           = "Jane Terraform"
		priority1           = "1"
		bccEmail1           = "test1@" + domainId
		bccEmail2           = "test2@" + domainId
	)

	CleanupRoutingEmailDomains()

	domainId = fmt.Sprintf("terraformroutes.%s.com", strings.Replace(uuid.NewString(), "-", "", -1))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create email domain and basic route
				Config: routingEmailDomain.GenerateRoutingEmailDomainResource(
					domainResourceLabel,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + GenerateRoutingEmailRouteResource(
					routeResourceLabel1,
					"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
					routePattern1,
					fromName1,
					fmt.Sprintf("from_email = \"%s\"", fromEmail1),
					generateRoutingAutoBcc(fromName1, bccEmail1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "pattern", routePattern1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "from_name", fromName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "from_email", fromEmail1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "auto_bcc.0.name", fromName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "auto_bcc.0.email", bccEmail1),
				),
			},
			{
				// Update email route and add a queue, language, and skill
				Config: routingEmailDomain.GenerateRoutingEmailDomainResource(
					domainResourceLabel,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName,
				) + routingLanguage.GenerateRoutingLanguageResource(
					langResourceLabel,
					langName,
				) + routingSkill.GenerateRoutingSkillResource(
					skillResourceLabel,
					skillName,
				) + GenerateRoutingEmailRouteResource(
					routeResourceLabel1,
					"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
					routePattern2,
					fromName2,
					generateRoutingReplyEmail(
						false,
						"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
						"genesyscloud_routing_email_route."+routeResourceLabel2+".id",
					),
					generateRoutingEmailQueueSettings(
						"genesyscloud_routing_queue."+queueResourceLabel+".id",
						priority1,
						"genesyscloud_routing_language."+langResourceLabel+".id",
						"genesyscloud_routing_skill."+skillResourceLabel+".id",
					),
				) + GenerateRoutingEmailRouteResource( // Second route to use as the reply_email_address
					routeResourceLabel2,
					"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
					routePattern3,
					fromName1,
					fmt.Sprintf("from_email = \"%s\"", fromEmail1),
					generateRoutingAutoBcc(fromName2, bccEmail2),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "pattern", routePattern2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "from_name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "from_email", ""),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeResourceLabel1, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeResourceLabel1, "language_id", "genesyscloud_routing_language."+langResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeResourceLabel1, "skill_ids.0", "genesyscloud_routing_skill."+skillResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "priority", priority1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "reply_email_address.0.domain_id", domainId),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeResourceLabel1, "reply_email_address.0.route_id", "genesyscloud_routing_email_route."+routeResourceLabel2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "pattern", routePattern3),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "from_name", fromName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "from_email", fromEmail1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "auto_bcc.0.name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "auto_bcc.0.email", bccEmail2),
				),
			},
			{
				// Update email reply to true
				Config: routingEmailDomain.GenerateRoutingEmailDomainResource(
					domainResourceLabel,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName,
				) + routingLanguage.GenerateRoutingLanguageResource(
					langResourceLabel,
					langName,
				) + routingSkill.GenerateRoutingSkillResource(
					skillResourceLabel,
					skillName,
				) + GenerateRoutingEmailRouteResource(
					routeResourceLabel1,
					"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
					routePattern2,
					fromName2,
					generateRoutingReplyEmail(
						true,
						"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
						"",
					),
					generateRoutingEmailQueueSettings(
						"genesyscloud_routing_queue."+queueResourceLabel+".id",
						priority1,
						"genesyscloud_routing_language."+langResourceLabel+".id",
						"genesyscloud_routing_skill."+skillResourceLabel+".id",
					),
				) + GenerateRoutingEmailRouteResource( // Second route to use as the reply_email_address
					routeResourceLabel2,
					"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
					routePattern3,
					fromName1,
					fmt.Sprintf("from_email = \"%s\"", fromEmail2),
					generateRoutingAutoBcc(fromName2, bccEmail2),
				),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for resources to be updated
						return nil
					},
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "pattern", routePattern2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "from_name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "from_email", ""),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeResourceLabel1, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeResourceLabel1, "language_id", "genesyscloud_routing_language."+langResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeResourceLabel1, "skill_ids.0", "genesyscloud_routing_skill."+skillResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "priority", priority1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "reply_email_address.0.domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "reply_email_address.0.route_id", ""),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "reply_email_address.0.self_reference_route", "true"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "auto_bcc.0.name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "auto_bcc.0.email", bccEmail2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "pattern", routePattern3),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "from_name", fromName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "from_email", fromEmail2),
				),
			},
			{
				// Update email reply to false and set a route id
				Config: routingEmailDomain.GenerateRoutingEmailDomainResource(
					domainResourceLabel,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + routingQueue.GenerateRoutingQueueResourceBasic(
					queueResourceLabel,
					queueName,
				) + routingLanguage.GenerateRoutingLanguageResource(
					langResourceLabel,
					langName,
				) + routingSkill.GenerateRoutingSkillResource(
					skillResourceLabel,
					skillName,
				) + GenerateRoutingEmailRouteResource(
					routeResourceLabel1,
					"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
					routePattern2,
					fromName2,
					generateRoutingAutoBcc(fromName2, bccEmail2),
					generateRoutingReplyEmail(
						false,
						"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
						"genesyscloud_routing_email_route."+routeResourceLabel2+".id",
					),
					generateRoutingEmailQueueSettings(
						"genesyscloud_routing_queue."+queueResourceLabel+".id",
						priority1,
						"genesyscloud_routing_language."+langResourceLabel+".id",
						"genesyscloud_routing_skill."+skillResourceLabel+".id",
					),
				) + GenerateRoutingEmailRouteResource( // Second route to use as the reply_email_address
					routeResourceLabel2,
					"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
					routePattern3,
					fromName1,
					fmt.Sprintf("from_email = \"%s\"", fromEmail1),
					generateRoutingAutoBcc(fromName2, bccEmail2),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "pattern", routePattern2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "from_name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "from_email", ""),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeResourceLabel1, "queue_id", "genesyscloud_routing_queue."+queueResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeResourceLabel1, "language_id", "genesyscloud_routing_language."+langResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeResourceLabel1, "skill_ids.0", "genesyscloud_routing_skill."+skillResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "priority", priority1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "reply_email_address.0.domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "reply_email_address.0.domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "auto_bcc.0.name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel1, "auto_bcc.0.email", bccEmail2),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeResourceLabel1, "reply_email_address.0.route_id", "genesyscloud_routing_email_route."+routeResourceLabel2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "auto_bcc.0.name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "auto_bcc.0.email", bccEmail2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "pattern", routePattern3),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "from_name", fromName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeResourceLabel2, "from_email", fromEmail1),
				),
			},
			{
				// Import/Read
				ResourceName:        "genesyscloud_routing_email_route." + routeResourceLabel1,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: domainId + "/",
			},
		},
		CheckDestroy: testVerifyRoutingEmailRouteDestroyed,
	})
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
	selfReferenceRoute bool,
	domainID string,
	routeID string) string {

	if selfReferenceRoute {
		return fmt.Sprintf(`
        reply_email_address {
            domain_id = %s
            self_reference_route = true
        }
	`, domainID)
	} else {
		return fmt.Sprintf(`
        reply_email_address {
            domain_id = %s
            route_id = %s
			self_reference_route = false
        }
	`, domainID, routeID)
	}
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
				if util.IsStatus404(resp) {
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

func CleanupRoutingEmailDomains() {
	config, err := provider.AuthorizeSdk()
	if err != nil {
		return
	}
	routingAPI := platformclientv2.NewRoutingApiWithConfig(config)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		routingEmailDomains, _, getErr := routingAPI.GetRoutingEmailDomains(pageSize, pageNum, false, "")
		if getErr != nil {
			log.Printf("failed to get page %v of routing email domains: %v", pageNum, getErr)
			return
		}

		if routingEmailDomains.Entities == nil || len(*routingEmailDomains.Entities) == 0 {
			break
		}

		for _, routingEmailDomain := range *routingEmailDomains.Entities {
			if routingEmailDomain.Name != nil && strings.HasPrefix(*routingEmailDomain.Name, "terraformroutes") {
				_, err := routingAPI.DeleteRoutingEmailDomain(*routingEmailDomain.Id)
				if err != nil {
					log.Printf("Failed to delete routing email domain %s: %s", *routingEmailDomain.Id, err)
					return
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
}
