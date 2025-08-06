package routing_utilization

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingUtilizationLabel "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_utilization_label"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceRoutingUtilizationBasic(t *testing.T) {
	t.Parallel()
	var (
		maxCapacity1  = "3"
		maxCapacity2  = "4"
		utilTypeCall  = "call"
		utilTypeEmail = "email"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingUtilizationResource(
					GenerateRoutingUtilMediaType("call", maxCapacity1, util.FalseValue),
					GenerateRoutingUtilMediaType("callback", maxCapacity1, util.FalseValue),
					GenerateRoutingUtilMediaType("chat", maxCapacity1, util.FalseValue),
					GenerateRoutingUtilMediaType("email", maxCapacity1, util.FalseValue),
					GenerateRoutingUtilMediaType("message", maxCapacity1, util.FalseValue),
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
					GenerateRoutingUtilMediaType("call", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeEmail)),
					GenerateRoutingUtilMediaType("callback", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
					GenerateRoutingUtilMediaType("chat", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
					GenerateRoutingUtilMediaType("email", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
					GenerateRoutingUtilMediaType("message", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
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

		redLabelResourceLabel   = "label_red_resource"
		blueLabelResourceLabel  = "label_blue_resource"
		greenLabelResourceLabel = "label_green_resource"
		redLabelName            = "Terraform Red Label" + uuid.NewString()
		blueLabelName           = "Terraform Blue Label" + uuid.NewString()
		greenLabelName          = "Terraform Green Label" + uuid.NewString()
	)

	if err := CleanupRoutingUtilizationLabel(); err != nil {
		t.Skipf("%v", err) // Skip the test and not fail it
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create
				Config: routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(redLabelResourceLabel, redLabelName, "") +
					routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(blueLabelResourceLabel, blueLabelName, redLabelResourceLabel) +
					routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(greenLabelResourceLabel, greenLabelName, blueLabelResourceLabel) +
					generateRoutingUtilizationResource(
						GenerateRoutingUtilMediaType("call", maxCapacity1, util.FalseValue),
						GenerateRoutingUtilMediaType("callback", maxCapacity1, util.FalseValue),
						GenerateRoutingUtilMediaType("chat", maxCapacity1, util.FalseValue),
						GenerateRoutingUtilMediaType("email", maxCapacity1, util.FalseValue),
						GenerateRoutingUtilMediaType("message", maxCapacity1, util.FalseValue),
						routingUtilizationLabel.GenerateLabelUtilization(redLabelResourceLabel, maxCapacity1),
						routingUtilizationLabel.GenerateLabelUtilization(blueLabelResourceLabel, maxCapacity1, redLabelResourceLabel),
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
					resource.TestCheckResourceAttrSet("genesyscloud_routing_utilization.routing-util", "label_utilizations.0.label_id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "label_utilizations.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_utilization.routing-util", "label_utilizations.1.label_id"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "label_utilizations.1.maximum_capacity", maxCapacity1),
				),
			},
			{
				// Update with a new max capacities and interruptible media types
				Config: routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(redLabelResourceLabel, redLabelName, "") +
					routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(blueLabelResourceLabel, blueLabelName, redLabelResourceLabel) +
					routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(greenLabelResourceLabel, greenLabelName, blueLabelResourceLabel) +
					generateRoutingUtilizationResource(
						GenerateRoutingUtilMediaType("call", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeEmail)),
						GenerateRoutingUtilMediaType("callback", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						GenerateRoutingUtilMediaType("chat", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						GenerateRoutingUtilMediaType("email", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						GenerateRoutingUtilMediaType("message", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						routingUtilizationLabel.GenerateLabelUtilization(redLabelResourceLabel, maxCapacity2),
						routingUtilizationLabel.GenerateLabelUtilization(blueLabelResourceLabel, maxCapacity2, redLabelResourceLabel),
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
				),
			},
			{ //Delete one by one to avoid conflict
				Config: routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(redLabelResourceLabel, redLabelName, "") +
					routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(blueLabelResourceLabel, blueLabelName, redLabelResourceLabel) +
					generateRoutingUtilizationResource(
						GenerateRoutingUtilMediaType("call", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeEmail)),
						GenerateRoutingUtilMediaType("callback", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						GenerateRoutingUtilMediaType("chat", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						GenerateRoutingUtilMediaType("email", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						GenerateRoutingUtilMediaType("message", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						routingUtilizationLabel.GenerateLabelUtilization(redLabelResourceLabel, maxCapacity2),
						routingUtilizationLabel.GenerateLabelUtilization(blueLabelResourceLabel, maxCapacity2, redLabelResourceLabel),
					),
			},
			{
				Config: routingUtilizationLabel.GenerateRoutingUtilizationLabelResource(redLabelResourceLabel, redLabelName, "") +
					generateRoutingUtilizationResource(
						GenerateRoutingUtilMediaType("call", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeEmail)),
						GenerateRoutingUtilMediaType("callback", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						GenerateRoutingUtilMediaType("chat", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						GenerateRoutingUtilMediaType("email", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						GenerateRoutingUtilMediaType("message", maxCapacity2, util.TrueValue, strconv.Quote(utilTypeCall)),
						routingUtilizationLabel.GenerateLabelUtilization(redLabelResourceLabel, maxCapacity2),
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
					allErrors := make([]string, 0)

					assertAttributeEquals(routingUtilization, "", "", &allErrors)
					assertAttributeEquals(routingUtilization, "call.0.include_non_acd", "true", &allErrors)
					assertAttributeEquals(routingUtilization, "call.0.interruptible_media_types.0", "email", &allErrors)
					assertAttributeEquals(routingUtilization, "call.0.maximum_capacity", maxCapacity2, &allErrors)
					assertAttributeEquals(routingUtilization, "callback.0.include_non_acd", "true", &allErrors)
					assertAttributeEquals(routingUtilization, "callback.0.interruptible_media_types.0", "call", &allErrors)
					assertAttributeEquals(routingUtilization, "callback.0.maximum_capacity", maxCapacity2, &allErrors)
					assertAttributeEquals(routingUtilization, "chat.0.include_non_acd", "true", &allErrors)
					assertAttributeEquals(routingUtilization, "chat.0.interruptible_media_types.0", "call", &allErrors)
					assertAttributeEquals(routingUtilization, "chat.0.maximum_capacity", maxCapacity2, &allErrors)
					assertAttributeEquals(routingUtilization, "email.0.include_non_acd", "true", &allErrors)
					assertAttributeEquals(routingUtilization, "email.0.interruptible_media_types.0", "call", &allErrors)
					assertAttributeEquals(routingUtilization, "email.0.maximum_capacity", maxCapacity2, &allErrors)
					assertAttributeEquals(routingUtilization, "message.0.include_non_acd", "true", &allErrors)
					assertAttributeEquals(routingUtilization, "message.0.interruptible_media_types.0", "call", &allErrors)
					assertAttributeEquals(routingUtilization, "message.0.maximum_capacity", maxCapacity2, &allErrors)

					numberOfLabelUtilizations, _ := strconv.Atoi(routingUtilization.Attributes["label_utilizations.#"])
					if numberOfLabelUtilizations != 0 {
						allErrors = append(allErrors, fmt.Sprintf("expected no label_utilizations, found %s", routingUtilization.Attributes["label_utilizations.#"]))
					}

					if len(allErrors) > 0 {
						return errors.New(strings.Join(allErrors[:], "\n"))
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

func generateRoutingUtilizationResource(attributes ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_utilization" "routing-util" {
		%s
	}
	`, strings.Join(attributes, "\n"))
}

func CleanupRoutingUtilizationLabel() error {
	config, err := provider.AuthorizeSdk()
	if err != nil {
		return err
	}

	routingAPI := platformclientv2.NewRoutingApiWithConfig(config)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		labels, _, getErr := routingAPI.GetRoutingUtilizationLabels(pageSize, pageNum, "", "")
		if getErr != nil {
			log.Printf("failed to get page %v of utilization labels: %v", pageNum, getErr)
			return getErr
		}

		if labels.Entities == nil || len(*labels.Entities) == 0 {
			return nil
		}

		for _, label := range *labels.Entities {
			if label.Id != nil && strings.HasPrefix(*label.Name, "Terraform") {
				_, err := routingAPI.DeleteRoutingUtilizationLabel(*label.Id, true)
				if err != nil {
					log.Printf("Failed to delete utilization label %s: %s", *label.Id, err)
					continue
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
}
