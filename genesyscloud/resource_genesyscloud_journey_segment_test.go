package genesyscloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v72/platformclientv2"
)

type journeySegmentStruct struct {
	resourceID  string
	displayName string
	color       string
	scope       string
	//context     struct {
	//	patterns struct {
	//	}
	//}
}

func TestAccResourceJourneySegmentBasic(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	resourcePrefix := "genesyscloud_journey_segment."
	journeySegmentIdPrefix := "terraform_test_"
	journeySegmentId := journeySegmentIdPrefix + strconv.Itoa(rand.Intn(1000))
	displayName1 := journeySegmentId
	displayName2 := journeySegmentId + "_updated"
	color1 := "#123456"
	scope1 := "Customer"

	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	cleanupJourneySegments(journeySegmentIdPrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateJourneySegmentResource(&journeySegmentStruct{
					journeySegmentId,
					displayName1,
					color1,
					scope1,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix+journeySegmentId, "display_name", displayName1),
					// TODO
				),
			},
			{
				// Update
				Config: generateJourneySegmentResource(&journeySegmentStruct{
					journeySegmentId,
					displayName2,
					color1,
					scope1,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix+journeySegmentId, "display_name", displayName2),
					// TODO
				),
			},
			{
				// Import/Read
				ResourceName:      resourcePrefix + journeySegmentId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyJourneySegmentsDestroyed,
	})
}

func cleanupJourneySegments(journeySegmentIdPrefix string) {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		journeySegments, _, getErr := journeyApi.GetJourneySegments("", pageSize, pageNum, true, nil, nil, "")
		if getErr != nil {
			return
		}

		if journeySegments.Entities == nil || len(*journeySegments.Entities) == 0 {
			break
		}

		for _, journeySegment := range *journeySegments.Entities {
			if journeySegment.DisplayName != nil && strings.HasPrefix(*journeySegment.DisplayName, journeySegmentIdPrefix) {
				_, delErr := journeyApi.DeleteJourneySegment(*journeySegment.Id)
				if delErr != nil {
					diag.Errorf("failed to delete journey segment %s(%s): %s", *journeySegment.Id, *journeySegment.DisplayName, delErr)
					return
				}
				log.Printf("Deleted journey segment %s(%s)", *journeySegment.Id, *journeySegment.DisplayName)
			}
		}
	}
}

func generateJourneySegmentResource(journeySegment *journeySegmentStruct) string {
	return fmt.Sprintf(`resource "genesyscloud_journey_segment" "%s" {
		display_name = "%s"
		color = "%s"
		scope = "%s"
		context {
			patterns {
				criteria {
					key = "geolocation.postalCode"
					values = ["alma"]
					operator = "equal"
					should_ignore_case = true
					entity_type = "visit"
				}
			} 
		}
		journey {
			patterns {
				criteria {
					key = "journeyCriteria"
					values = ["korte"]
					operator = "equal"
					should_ignore_case = true
				}
				count = 1
				stream_type = "Web"
				session_type = "*"
			} 
		}
	}`, journeySegment.resourceID,
		journeySegment.displayName,
		journeySegment.color,
		journeySegment.scope)
}

func testVerifyJourneySegmentsDestroyed(state *terraform.State) error {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_journey_segment" {
			continue
		}

		journeySegment, resp, err := journeyApi.GetJourneySegment(rs.Primary.ID)
		if journeySegment != nil {
			return fmt.Errorf("journey segment (%s) still exists", rs.Primary.ID)
		}

		if isStatus404(resp) {
			// Journey segment not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("unexpected error: %s", err)
	}
	// Success. All Journey segment destroyed
	return nil
}
