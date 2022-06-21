package genesyscloud

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v72/platformclientv2"
)

type journeySegmentStruct struct {
	resourceID           string
	displayName          string
	color                string
	scope                string
	shouldDisplayToAgent bool
	context              string
	journey              string
}

type contextStruct struct {
	key              string
	values           string
	operator         string
	shouldIgnoreCase bool
	entityType       string
}

type journeyStruct struct {
	count            int
	streamType       string
	sessionType      string
	eventName        string
	key              string
	values           string
	operator         string
	shouldIgnoreCase bool
}

func TestAccResourceJourneySegmentBasic(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	const resourcePrefix = "genesyscloud_journey_segment."
	const journeySegmentIdPrefix = "terraform_test_"
	journeySegmentId := journeySegmentIdPrefix + strconv.Itoa(rand.Intn(1000))
	displayName1 := journeySegmentId
	displayName2 := journeySegmentId + "_updated"
	const color1 = "#008000"
	const color2 = "#308000"
	const scope1 = "Session"
	const scope2 = "Session"
	const shouldDisplayToAgent1 = false
	const shouldDisplayToAgent2 = true

	const contextPatternCriteriaKey1 = "geolocation.postalCode"
	const contextPatternCriteriaValues1 = "something"
	const contextPatternCriteriaOperator1 = "equal"
	const contextPatternCriteriaShouldIgnoreCase1 = true
	const contextPatternCriteriaEntityType1 = "visit"

	const contextPatternCriteriaKey2 = "geolocation.region"
	const contextPatternCriteriaValues2 = "something1"
	const contextPatternCriteriaOperator2 = "containsAll"
	const contextPatternCriteriaShouldIgnoreCase2 = false
	const contextPatternCriteriaEntityType2 = "visit"

	const journeyCount1 = 1
	const journeyStreamType = "Web"
	const journeySessionType = "web"
	const journeyEventName1 = "EventName"
	const journeyPatternCriteriaKey1 = "page.hostname"
	const journeyPatternCriteriaValues1 = "something_else"
	const journeyPatternCriteriaOperator1 = "equal"
	const journeyPatternCriteriaShouldIgnoreCase1 = false

	const journeyCount2 = 1
	const journeyEventName2 = "OtherEventName"
	const journeyPatternCriteriaKey2 = "attributes.bleki.value"
	const journeyPatternCriteriaValues2 = "Blabla"
	const journeyPatternCriteriaOperator2 = "notEqual"
	const journeyPatternCriteriaShouldIgnoreCase2 = true

	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	cleanupJourneySegments(journeySegmentIdPrefix)

	journeySegmentResource1 := generateJourneySegmentResource(&journeySegmentStruct{
		journeySegmentId,
		displayName1,
		color1,
		scope1,
		shouldDisplayToAgent1,
		generateContext(&contextStruct{
			contextPatternCriteriaKey1,
			contextPatternCriteriaValues1,
			contextPatternCriteriaOperator1,
			contextPatternCriteriaShouldIgnoreCase1,
			contextPatternCriteriaEntityType1,
		}),
		generateJourney(&journeyStruct{
			journeyCount1,
			journeyStreamType,
			journeySessionType,
			journeyEventName1,
			journeyPatternCriteriaKey1,
			journeyPatternCriteriaValues1,
			journeyPatternCriteriaOperator1,
			journeyPatternCriteriaShouldIgnoreCase1,
		}),
	})
	journeySegmentResource2 := generateJourneySegmentResource(&journeySegmentStruct{
		journeySegmentId,
		displayName2,
		color2,
		scope2,
		shouldDisplayToAgent2,
		generateContext(&contextStruct{
			contextPatternCriteriaKey2,
			contextPatternCriteriaValues2,
			contextPatternCriteriaOperator2,
			contextPatternCriteriaShouldIgnoreCase2,
			contextPatternCriteriaEntityType2,
		}),
		generateJourney(&journeyStruct{
			journeyCount2,
			journeyStreamType,
			journeySessionType,
			journeyEventName2,
			journeyPatternCriteriaKey2,
			journeyPatternCriteriaValues2,
			journeyPatternCriteriaOperator2,
			journeyPatternCriteriaShouldIgnoreCase2,
		}),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: journeySegmentResource1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix+journeySegmentId, "display_name", displayName1),
					resource.TestCheckResourceAttr(resourcePrefix+journeySegmentId, "color", color1),
					resource.TestCheckResourceAttr(resourcePrefix+journeySegmentId, "scope", scope1),
				),
			},
			{
				// Update
				Config: journeySegmentResource2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix+journeySegmentId, "display_name", displayName2),
					resource.TestCheckResourceAttr(resourcePrefix+journeySegmentId, "color", color2),
					resource.TestCheckResourceAttr(resourcePrefix+journeySegmentId, "scope", scope1),
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

func generateJourneySegmentResource(journeySegment *journeySegmentStruct) string {
	return fmt.Sprintf(`resource "genesyscloud_journey_segment" "%s" {
		display_name = "%s"
		color = "%s"
		scope = "%s"
		should_display_to_agent = %t
		%s
		%s
	}`, journeySegment.resourceID,
		journeySegment.displayName,
		journeySegment.color,
		journeySegment.scope,
		journeySegment.shouldDisplayToAgent,
		journeySegment.context,
		journeySegment.journey)
}

func generateContext(context *contextStruct) string {
	return fmt.Sprintf(`context {
			patterns {
				criteria {
					key = "%s"
					values = ["%s"]
					operator = "%s"
					should_ignore_case = %t
					entity_type = "%s"
				}
			}
		}`, context.key,
		context.values,
		context.operator,
		context.shouldIgnoreCase,
		context.entityType,
	)
}

func generateJourney(journey *journeyStruct) string {
	return fmt.Sprintf(`journey {
			patterns {
				criteria {
					key = "%s"
					values = ["%s"]
					operator = "%s"
					should_ignore_case = %t
				}
				count = %d
				stream_type = "%s"
				session_type = "%s"
				event_name = "%s"
			} 
		}`, journey.key,
		journey.values,
		journey.operator,
		journey.shouldIgnoreCase,
		journey.count,
		journey.streamType,
		journey.sessionType,
		journey.eventName,
	)
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

func cleanupJourneySegments(journeySegmentIdPrefix string) {
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)

	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
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
					diag.Errorf("failed to delete journey segment %s (%s): %s", *journeySegment.Id, *journeySegment.DisplayName, delErr)
					return
				}
				log.Printf("Deleted journey segment %s (%s)", *journeySegment.Id, *journeySegment.DisplayName)
			}
		}

		pageCount = *journeySegments.PageCount
	}
}
