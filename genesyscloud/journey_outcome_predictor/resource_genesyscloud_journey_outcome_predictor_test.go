package journey_outcome_predictor

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
)

func TestAccResourceJourneyOutcomePredictor(t *testing.T) {
	t.Parallel()
	var (
		fullResourceName = "genesyscloud_journey_outcome_predictor.test_predictor"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: predictorResource(),
			},
			{
				ResourceName:            fullResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
			},
		},
		CheckDestroy: testVerifyPredictorDestroyed,
	})
}

func predictorResource() string {
	return fmt.Sprint(`
	resource "genesyscloud_journey_outcome" "test_outcome" {
		is_active    = true
		display_name = "example journey outcome - delete"
		description  = "description of journey outcome"
		is_positive  = true
		journey {
			patterns {
			criteria {
				key                = "page.url"
				values             = ["forms/car-loan/"]
				operator           = "containsAny"
				should_ignore_case = true
			}
			count        = 1
			stream_type  = "Web"
			session_type = "web"
			}
		}
	}

	resource "genesyscloud_journey_outcome_predictor" "test_predictor" {
		outcome {
			id = "${genesyscloud_journey_outcome.test_outcome.id}"
		}
	}
	`)
}

func testVerifyPredictorDestroyed(state *terraform.State) error {
	journeyAPI := platformclientv2.NewJourneyApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_journey_outcome_predictor" {
			continue
		}

		predictor, resp, err := journeyAPI.GetJourneyOutcomesPredictor(rs.Primary.ID)
		if predictor != nil {
			return fmt.Errorf("Predictor (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Predictor not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All predictors destroyed
	return nil

}
