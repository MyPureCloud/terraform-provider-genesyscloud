package journey_views

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceJourneyViewsBasic(t *testing.T) {
	var (
		name                = "test journey from tf 1"
		duration            = "P1Y"
		elementsId          = "ac6c61b5-1cd4-4c6e-a8a5-edb74d9117eb"
		elementsName        = "Wrap Up"
		attributeType       = "Event"
		attributeId         = "a416328b-167c-0365-d0e1-f072cd5d4ded"
		attributeSource     = "Voice"
		filterType          = "And"
		predicatesDimension = "mediaType"
		predicatesValues    = "VOICE"
		predicatesOperator  = "Matches"
		predicatesNoValue   = false
		testUserResource    = "user_resource1"
		testUserName        = "nameUser1" + uuid.NewString()
		testUserEmail       = uuid.NewString() + "@example.com"
		journeyResource     = "journey_resource1"
		emptyElementBlock   = ""
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				//Create
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + generateJourneyView(journeyResource, name, duration, emptyElementBlock),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "duration", duration),
				),
			},
			{
				//Update
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + generateJourneyView(journeyResource, name, duration, generateElements(
					elementsId,
					elementsName,
					generateAttributes(attributeType, attributeId, attributeSource),
					generateFilter(filterType, generatePredicates(predicatesDimension, predicatesValues, predicatesOperator, predicatesNoValue)),
				)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "duration", duration),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.id", elementsId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.name", elementsName),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.attributes.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.attributes.0.type", attributeType),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.attributes.0.id", attributeId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.attributes.0.source", attributeSource),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.filter.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.filter.0.type", "And"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.filter.0.predicates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.filter.0.predicates.0.dimension", predicatesDimension),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.filter.0.predicates.0.values.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.filter.0.predicates.0.values.0", predicatesValues),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.filter.0.predicates.0.operator", predicatesOperator),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.filter.0.predicates.0.no_value", fmt.Sprintf("%t", predicatesNoValue)),
				),
			},
			{
				//Update without filter
				Config: generateUserWithCustomAttrs(testUserResource, testUserEmail, testUserName) + generateJourneyView(journeyResource, name, duration, generateElements(
					elementsId,
					elementsName,
					generateAttributes(attributeType, attributeId, attributeSource),
					"",
				)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "duration", duration),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.id", elementsId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.name", elementsName),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.attributes.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.attributes.0.type", attributeType),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.attributes.0.id", attributeId),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.attributes.0.source", attributeSource),
					resource.TestCheckResourceAttr("genesyscloud_journey_views."+journeyResource, "elements.0.filter.#", "0"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_journey_views." + journeyResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyJourneyViewsDestroyed,
	})
}

func generateUserWithCustomAttrs(resourceID string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceID, email, name, strings.Join(attrs, "\n"))
}

func generateJourneyView(journeyResource string, name string, duration string, elementsBlock string) string {
	return fmt.Sprintf(`resource "genesyscloud_journey_views" "%s" {
    duration = "%s"
    name = "%s"
    %s
}`, journeyResource, duration, name, func() string {
		if elementsBlock != "" {
			return elementsBlock
		}
		return ""
	}())
}

func generateElements(id string, name string, attributesBlock string, filter string) string {
	return fmt.Sprintf(`
    elements {
        id = "%s"
        name = "%s"
        %s
        %s
    }`, id, name, attributesBlock, filter)
}

func generateFilter(filterType string, nestedBlocks ...string) string {
	return fmt.Sprintf(`
        filter {
            type       = "%s"
            %s
        }`, filterType, strings.Join(nestedBlocks, "\n"))
}

func generateAttributes(attributeType string, attributeId string, attributeSource string) string {
	return fmt.Sprintf(`
        attributes {
            type   = "%s"
            id     = "%s"
            source = "%s"
        }`, attributeType, attributeId, attributeSource)
}

func generatePredicates(dimension string, values string, operator string, noValue bool) string {
	return fmt.Sprintf(`
            predicates  {
                dimension = "%s"
                values    = ["%s"]
                operator  = "%s"
                no_value  = %v
            }`, dimension, values, operator, noValue)
}

func testVerifyJourneyViewsDestroyed(state *terraform.State) error {
	journeyViewApi := platformclientv2.NewJourneyApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_journey_views" {
			continue
		}

		journeyView, resp, err := journeyViewApi.GetJourneyView(rs.Primary.ID)
		if journeyView != nil {
			return fmt.Errorf("journeyView (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// JourneyView not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All journeyView destroyed
	return nil
}
