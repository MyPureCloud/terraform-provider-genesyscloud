package routing_email_route

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/architect_flow"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

func TestAccResourceRoutingEmailRoute(t *testing.T) {
	var (
		domainRes          = "routing-domain1"
		domainId           = fmt.Sprintf("terraform.%s.com", strings.Replace(uuid.NewString(), "-", "", -1))
		queueResource      = "email-queue"
		queueName          = "Terraform Email Queue-" + uuid.NewString()
		langResource       = "email-lang"
		langName           = "tflang" + uuid.NewString()
		skillResource      = "test-skill1"
		skillName          = "Terraform Skill" + uuid.NewString()
		routeRes           = "email-route1"
		routeRes2          = "email-route2"
		routePattern1      = "terraform1"
		routePattern2      = "terraform2"
		routePattern3      = "terraform3"
		fromEmail1         = "terraform1@test.com"
		fromEmail2         = "terraform2@test.com"
		fromName1          = "John Terraform"
		fromName2          = "Jane Terraform"
		priority1          = "1"
		bccEmail1          = "test1@" + domainId
		bccEmail2          = "test2@" + domainId
		emailFlowResource1 = "test_flow1"
		emailFlowFilePath1 = "../../examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml"
	)

	CleanupRoutingEmailDomains()

	// Test error configs
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Confirm mutual exclusivity of reply_email_address and from_email
				Config: gcloud.GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + generateRoutingEmailRouteResource(
					routeRes+"expectFail",
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern1,
					fromName1,
					fmt.Sprintf("from_email = \"%s\"", fromEmail1),
					generateRoutingReplyEmail(
						false,
						"genesyscloud_routing_email_domain."+domainRes+".id",
						"genesyscloud_routing_email_route."+routeRes2+".id",
					),
				) + generateRoutingEmailRouteResource( // Second route to use as the reply_email_address
					routeRes2,
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern3,
					fromName1,
					fmt.Sprintf("from_email = \"%s\"", fromEmail1),
					generateRoutingAutoBcc(fromName2, bccEmail2),
				),
				ExpectError: regexp.MustCompile("Conflicting configuration arguments"),
			},
			{
				// Confirm mutual exclusivity of reply_email_address and auto_bcc
				Config: gcloud.GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + generateRoutingEmailRouteResource(
					routeRes+"expectFail",
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern1,
					fromName1,
					generateRoutingAutoBcc(fromName1, bccEmail1),
					generateRoutingReplyEmail(
						false,
						"genesyscloud_routing_email_domain."+domainRes+".id",
						"genesyscloud_routing_email_route."+routeRes2+".id",
					),
				) + generateRoutingEmailRouteResource( // Second route to use as the reply_email_address
					routeRes2,
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern3,
					fromName1,
					fmt.Sprintf("from_email = \"%s\"", fromEmail1),
					generateRoutingAutoBcc(fromName2, bccEmail2),
				),
				ExpectError: regexp.MustCompile("Conflicting configuration arguments"),
			},
			{
				// Confirm mutual exclusivity of flow_id and queue_id
				Config: gcloud.GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + gcloud.GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName,
				) + gcloud.GenerateRoutingLanguageResource(
					langResource,
					langName,
				) + gcloud.GenerateRoutingSkillResource(
					skillResource,
					skillName,
				) + architect_flow.GenerateFlowResource(
					emailFlowResource1,
					emailFlowFilePath1,
					"",
					false,
				) + generateRoutingEmailRouteResource(
					routeRes+"expectFail",
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern1,
					fromName1,
					fmt.Sprintf("from_email = \"%s\"", fromEmail1),
					generateRoutingEmailQueueSettings(
						"genesyscloud_routing_queue."+queueResource+".id",
						priority1,
						"genesyscloud_routing_language."+langResource+".id",
						"genesyscloud_routing_skill."+skillResource+".id",
					),
					fmt.Sprintf("flow_id = genesyscloud_flow.%s.id", emailFlowResource1),
				),
				ExpectError: regexp.MustCompile("Conflicting configuration arguments"),
			},
		},
		CheckDestroy: testVerifyRoutingEmailRouteDestroyed,
	})

	// Standard acceptance tests
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create email domain and basic route
				Config: gcloud.GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + generateRoutingEmailRouteResource(
					routeRes,
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern1,
					fromName1,
					fmt.Sprintf("from_email = \"%s\"", fromEmail1),
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
				Config: gcloud.GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + gcloud.GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName,
				) + gcloud.GenerateRoutingLanguageResource(
					langResource,
					langName,
				) + gcloud.GenerateRoutingSkillResource(
					skillResource,
					skillName,
				) + generateRoutingEmailRouteResource(
					routeRes,
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern2,
					fromName2,
					generateRoutingReplyEmail(
						false,
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
					fmt.Sprintf("from_email = \"%s\"", fromEmail1),
					generateRoutingAutoBcc(fromName2, bccEmail2),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "pattern", routePattern2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "from_name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "from_email", ""),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "queue_id", "genesyscloud_routing_queue."+queueResource, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "language_id", "genesyscloud_routing_language."+langResource, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "skill_ids.0", "genesyscloud_routing_skill."+skillResource, "id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "priority", priority1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "reply_email_address.0.domain_id", domainId),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "reply_email_address.0.route_id", "genesyscloud_routing_email_route."+routeRes2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "pattern", routePattern3),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "from_name", fromName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "from_email", fromEmail1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "auto_bcc.0.name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "auto_bcc.0.email", bccEmail2),
				),
			},
			{
				// Update email reply to true
				Config: gcloud.GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + gcloud.GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName,
				) + gcloud.GenerateRoutingLanguageResource(
					langResource,
					langName,
				) + gcloud.GenerateRoutingSkillResource(
					skillResource,
					skillName,
				) + generateRoutingEmailRouteResource(
					routeRes,
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern2,
					fromName2,
					generateRoutingReplyEmail(
						true,
						"genesyscloud_routing_email_domain."+domainRes+".id",
						"",
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
					fmt.Sprintf("from_email = \"%s\"", fromEmail2),
					generateRoutingAutoBcc(fromName2, bccEmail2),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "pattern", routePattern2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "from_name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "from_email", ""),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "queue_id", "genesyscloud_routing_queue."+queueResource, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "language_id", "genesyscloud_routing_language."+langResource, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "skill_ids.0", "genesyscloud_routing_skill."+skillResource, "id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "priority", priority1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "reply_email_address.0.domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "reply_email_address.0.route_id", ""),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "reply_email_address.0.self_reference_route", "true"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "auto_bcc.0.name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "auto_bcc.0.email", bccEmail2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "pattern", routePattern3),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "from_name", fromName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "from_email", fromEmail2),
				),
			},
			{
				// Update email reply to false and set a route id
				Config: gcloud.GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + gcloud.GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName,
				) + gcloud.GenerateRoutingLanguageResource(
					langResource,
					langName,
				) + gcloud.GenerateRoutingSkillResource(
					skillResource,
					skillName,
				) + generateRoutingEmailRouteResource(
					routeRes,
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern2,
					fromName2,
					generateRoutingReplyEmail(
						false,
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
					fmt.Sprintf("from_email = \"%s\"", fromEmail1),
					generateRoutingAutoBcc(fromName2, bccEmail2),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "pattern", routePattern2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "from_name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "from_email", ""),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "queue_id", "genesyscloud_routing_queue."+queueResource, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "language_id", "genesyscloud_routing_language."+langResource, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "skill_ids.0", "genesyscloud_routing_skill."+skillResource, "id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "priority", priority1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "reply_email_address.0.domain_id", domainId),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes, "reply_email_address.0.domain_id", domainId),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_email_route."+routeRes, "reply_email_address.0.route_id", "genesyscloud_routing_email_route."+routeRes2, "id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "auto_bcc.0.name", fromName2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "auto_bcc.0.email", bccEmail2),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "pattern", routePattern3),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "from_name", fromName1),
					resource.TestCheckResourceAttr("genesyscloud_routing_email_route."+routeRes2, "from_email", fromEmail1),
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
	otherAttrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_email_route" "%s" {
            domain_id = %s
            pattern = "%s"
            from_name = "%s"
            %s
        }
        `, resourceID, domainID, pattern, fromName, strings.Join(otherAttrs, "\n"))
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
	sdkConfig, err := provider.AuthorizeSdk()
	if err != nil {
		log.Fatal(err)
	}
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		routingEmailDomains, _, getErr := routingAPI.GetRoutingEmailDomains(pageNum, pageSize, false, "")
		if getErr != nil {
			return
		}

		if routingEmailDomains.Entities == nil || len(*routingEmailDomains.Entities) == 0 {
			return
		}

		for _, routingEmailDomain := range *routingEmailDomains.Entities {
			if routingEmailDomain.Id != nil && strings.HasPrefix(*routingEmailDomain.Id, "terraform") {
				_, err := routingAPI.DeleteRoutingEmailDomain(*routingEmailDomain.Id)
				if err != nil {
					log.Printf("Failed to delete routing email domain %s: %s", *routingEmailDomain.Id, err)
					continue
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
}
