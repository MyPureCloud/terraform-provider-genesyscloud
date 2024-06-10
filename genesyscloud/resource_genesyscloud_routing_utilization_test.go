package genesyscloud

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
)

func TestAccResourceBasicRoutingUtilization(t *testing.T) {
	t.Parallel()
	var (
		maxCapacity1  = "3"
		maxCapacity2  = "4"
		utilTypeCall  = "call"
		utilTypeEmail = "email"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingUtilizationResource(
					generateRoutingUtilMediaType("call", maxCapacity1, util.FalseValue),
					generateRoutingUtilMediaType("callback", maxCapacity1, util.FalseValue),
					generateRoutingUtilMediaType("chat", maxCapacity1, util.FalseValue),
					generateRoutingUtilMediaType("email", maxCapacity1, util.FalseValue),
					generateRoutingUtilMediaType("message", maxCapacity1, util.FalseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.interruptible_media_types"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.interruptible_media_types"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.interruptible_media_types"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.interruptible_media_types"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.interruptible_media_types"),
				),
			},
			{
				// Update with a new max capacities and interruptible media types
				Config: generateRoutingUtilizationResource(
					generateRoutingUtilMediaType("call", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeEmail)),
					generateRoutingUtilMediaType("callback", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
					generateRoutingUtilMediaType("chat", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
					generateRoutingUtilMediaType("email", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
					generateRoutingUtilMediaType("message", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "message.0.interruptible_media_types", utilTypeCall),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_utilization.routing-util",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceRoutingUtilizationWithLabels(t *testing.T) {
	var (
		maxCapacity1  = "3"
		maxCapacity2  = "4"
		utilTypeCall  = "call"
		utilTypeEmail = "email"

		redLabelResource   = "label_red_resource"
		blueLabelResource  = "label_blue_resource"
		greenLabelResource = "label_green_resource"
		redLabelName       = "Terraform Red Label" + uuid.NewString()
		blueLabelName      = "Terraform Blue Label" + uuid.NewString()
		greenLabelName     = "Terraform Green Label" + uuid.NewString()
	)

	CleanupRoutingUtilizationLabel()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			util.TestAccPreCheck(t)
			if err := checkIfLabelsAreEnabled(); err != nil {
				t.Skipf("%v", err) // be sure to skip the test and not fail it
			}
		},
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateRoutingUtilizationLabelResource(redLabelResource, redLabelName, "") +
					GenerateRoutingUtilizationLabelResource(blueLabelResource, blueLabelName, redLabelResource) +
					GenerateRoutingUtilizationLabelResource(greenLabelResource, greenLabelName, blueLabelResource) +
					generateRoutingUtilizationResource(
						generateRoutingUtilMediaType("call", maxCapacity1, util.FalseValue),
						generateRoutingUtilMediaType("callback", maxCapacity1, util.FalseValue),
						generateRoutingUtilMediaType("chat", maxCapacity1, util.FalseValue),
						generateRoutingUtilMediaType("email", maxCapacity1, util.FalseValue),
						generateRoutingUtilMediaType("message", maxCapacity1, util.FalseValue),
						generateLabelUtilization(redLabelResource, maxCapacity1),
						generateLabelUtilization(blueLabelResource, maxCapacity1, redLabelResource),
					),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for resources to be updated
						return nil
					},
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.interruptible_media_types"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.interruptible_media_types"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.interruptible_media_types"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.interruptible_media_types"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.include_non_acd", util.FalseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.interruptible_media_types"),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_utilization.routing-util", "label_utilizations.0.label_id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "label_utilizations.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_utilization.routing-util", "label_utilizations.1.label_id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "label_utilizations.1.maximum_capacity", maxCapacity1),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
				),
			},
			{
				// Update with a new max capacities and interruptible media types
				Config: GenerateRoutingUtilizationLabelResource(redLabelResource, redLabelName, "") +
					GenerateRoutingUtilizationLabelResource(blueLabelResource, blueLabelName, redLabelResource) +
					GenerateRoutingUtilizationLabelResource(greenLabelResource, greenLabelName, blueLabelResource) +
					generateRoutingUtilizationResource(
						generateRoutingUtilMediaType("call", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeEmail)),
						generateRoutingUtilMediaType("callback", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("chat", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("email", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("message", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						generateLabelUtilization(redLabelResource, maxCapacity2),
						generateLabelUtilization(blueLabelResource, maxCapacity2, redLabelResource),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.include_non_acd", util.TrueValue),
					util.ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "message.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_utilization.routing-util", "label_utilizations.0.label_id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "label_utilizations.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_utilization.routing-util", "label_utilizations.1.label_id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "label_utilizations.1.maximum_capacity", maxCapacity2),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
				),
			},
			{ //Delete one by one to avoid conflict
				Config: GenerateRoutingUtilizationLabelResource(redLabelResource, redLabelName, "") +
					GenerateRoutingUtilizationLabelResource(blueLabelResource, blueLabelName, redLabelResource) +
					generateRoutingUtilizationResource(
						generateRoutingUtilMediaType("call", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeEmail)),
						generateRoutingUtilMediaType("callback", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("chat", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("email", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("message", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						generateLabelUtilization(redLabelResource, maxCapacity2),
						generateLabelUtilization(blueLabelResource, maxCapacity2, redLabelResource),
					),
			},
			{
				Config: GenerateRoutingUtilizationLabelResource(redLabelResource, redLabelName, "") +
					generateRoutingUtilizationResource(
						generateRoutingUtilMediaType("call", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeEmail)),
						generateRoutingUtilMediaType("callback", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("chat", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("email", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						generateRoutingUtilMediaType("message", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						generateLabelUtilization(redLabelResource, maxCapacity2),
					),
			},
			{
				// Import/Read
				ResourceName: "genesyscloud_routing_utilization.routing-util",
				ImportState:  true,
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					// When importing, there's no previous state, so no label utilizations are added to the state
					if len(s) != 1 {
						return fmt.Errorf("expected 1 state: %#v", s)
					}

					routingUtilization := s[0]
					errors := make([]string, 0)

					assertAttributeEquals(routingUtilization, "", "", &errors)
					assertAttributeEquals(routingUtilization, "call.0.include_non_acd", "true", &errors)
					assertAttributeEquals(routingUtilization, "call.0.interruptible_media_types.0", "email", &errors)
					assertAttributeEquals(routingUtilization, "call.0.maximum_capacity", maxCapacity2, &errors)
					assertAttributeEquals(routingUtilization, "callback.0.include_non_acd", "true", &errors)
					assertAttributeEquals(routingUtilization, "callback.0.interruptible_media_types.0", "call", &errors)
					assertAttributeEquals(routingUtilization, "callback.0.maximum_capacity", maxCapacity2, &errors)
					assertAttributeEquals(routingUtilization, "chat.0.include_non_acd", "true", &errors)
					assertAttributeEquals(routingUtilization, "chat.0.interruptible_media_types.0", "call", &errors)
					assertAttributeEquals(routingUtilization, "chat.0.maximum_capacity", maxCapacity2, &errors)
					assertAttributeEquals(routingUtilization, "email.0.include_non_acd", "true", &errors)
					assertAttributeEquals(routingUtilization, "email.0.interruptible_media_types.0", "call", &errors)
					assertAttributeEquals(routingUtilization, "email.0.maximum_capacity", maxCapacity2, &errors)
					assertAttributeEquals(routingUtilization, "message.0.include_non_acd", "true", &errors)
					assertAttributeEquals(routingUtilization, "message.0.interruptible_media_types.0", "call", &errors)
					assertAttributeEquals(routingUtilization, "message.0.maximum_capacity", maxCapacity2, &errors)

					numberOfLabelUtilizations, _ := strconv.Atoi(routingUtilization.Attributes["label_utilizations.#"])
					if numberOfLabelUtilizations != 0 {
						errors = append(errors, fmt.Sprintf("expected no label_utilizations, found %s", routingUtilization.Attributes["label_utilizations.#"]))
					}

					if len(errors) > 0 {
						return fmt.Errorf(strings.Join(errors[:], "\n"))
					}

					return nil
				},
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second)
						return nil
					},
				),
			},
		},
	})
}

func assertAttributeEquals(state *terraform.InstanceState, attributeName, expectedValue string, errors *[]string) {
	if state.Attributes[attributeName] != expectedValue {
		*errors = append(*errors, fmt.Sprintf("expected %s to be %s, actual: %s", attributeName, expectedValue, state.Attributes[attributeName]))
	}
}

func generateRoutingUtilMediaType(
	mediaType string,
	maxCapacity string,
	includeNonAcd string,
	interruptTypes ...string) string {
	return fmt.Sprintf(`%s {
		maximum_capacity = %s
		include_non_acd = %s
		interruptible_media_types = [%s]
	}
	`, mediaType, maxCapacity, includeNonAcd, strings.Join(interruptTypes, ","))
}

func generateLabelUtilization(
	labelResource string,
	maxCapacity string,
	interruptingLabelResourceNames ...string) string {

	interruptingLabelResources := make([]string, 0)
	for _, resourceName := range interruptingLabelResourceNames {
		interruptingLabelResources = append(interruptingLabelResources, "genesyscloud_routing_utilization_label."+resourceName+".id")
	}

	return fmt.Sprintf(`label_utilizations {
		label_id = genesyscloud_routing_utilization_label.%s.id
		maximum_capacity = %s
		interrupting_label_ids = [%s]
	}
	`, labelResource, maxCapacity, strings.Join(interruptingLabelResources, ","))
}

func generateRoutingUtilizationResource(attributes ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_utilization" "routing-util" {
		%s
	}
	`, strings.Join(attributes, "\n"))
}

func CleanupRoutingUtilizationLabel() {
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		labels, _, getErr := routingAPI.GetRoutingUtilizationLabels(pageSize, pageNum, "", "")
		if getErr != nil {
			log.Printf("failed to get page %v of routing email domains: %v", pageNum, getErr)
			return
		}

		if labels.Entities == nil || len(*labels.Entities) == 0 {
			return
		}

		for _, label := range *labels.Entities {
			if label.Id != nil && strings.HasPrefix(*label.Name, "Terraform") {
				_, err := routingAPI.DeleteRoutingUtilizationLabel(*label.Id, true)
				if err != nil {
					log.Printf("Failed to delete routing email domain %s: %s", *label.Id, err)
					continue
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
}
