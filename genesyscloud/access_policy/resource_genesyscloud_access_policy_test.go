package access_policy

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

func TestAccResourceAccessPolicy(t *testing.T) {
	var (
		resourceLabel  = "test-access-policy"
		name1          = "TF Test Policy - " + uuid.NewString()
		name2          = "TF Test Policy Updated - " + uuid.NewString()
		description1   = "Test access policy created by Terraform acceptance tests"
		description2   = "Updated description for the test access policy"
		targetResource = "authorization:role:view"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create an access policy with basic fields
				Config: generateAccessPolicyResource(
					resourceLabel,
					name1,
					description1,
					targetResource,
					"DENY",
					"ALL",
					"",     // no subject_id for ALL
					"true", // enabled
					"",     // no condition
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_access_policy."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_access_policy."+resourceLabel, "description", description1),
					resource.TestCheckResourceAttr("genesyscloud_access_policy."+resourceLabel, "target_resource", targetResource),
					resource.TestCheckResourceAttr("genesyscloud_access_policy."+resourceLabel, "effect", "DENY"),
					resource.TestCheckResourceAttr("genesyscloud_access_policy."+resourceLabel, "subject_type", "ALL"),
					resource.TestCheckResourceAttr("genesyscloud_access_policy."+resourceLabel, "enabled", "true"),
				),
			},
			{
				// Update name and description, disable the policy
				Config: generateAccessPolicyResource(
					resourceLabel,
					name2,
					description2,
					targetResource,
					"DENY",
					"ALL",
					"",
					"false", // disable it
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_access_policy."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_access_policy."+resourceLabel, "description", description2),
					resource.TestCheckResourceAttr("genesyscloud_access_policy."+resourceLabel, "enabled", "false"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_access_policy." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyAccessPolicyDestroyed,
	})
}

func TestAccResourceAccessPolicyWithCondition(t *testing.T) {
	t.Skip("Skipping until valid condition attributes are confirmed for the test target resource. " +
		"The condition JSON structure depends on the specific target resource being used.")

	var (
		resourceLabel  = "test-access-policy-condition"
		name           = "TF Test Policy With Condition - " + uuid.NewString()
		description    = "Test access policy with condition JSON"
		targetResource = "authorization:role:view"
		conditionJSON  = `{"and":[{"attribute":"subject.role.name","operator":"eq","value":"employee"}]}`
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with condition
				Config: generateAccessPolicyResourceWithCondition(
					resourceLabel,
					name,
					description,
					targetResource,
					"DENY",
					"ALL",
					"true",
					conditionJSON,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_access_policy."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_access_policy."+resourceLabel, "effect", "DENY"),
					resource.TestCheckResourceAttrSet("genesyscloud_access_policy."+resourceLabel, "condition_json"),
				),
			},
		},
		CheckDestroy: testVerifyAccessPolicyDestroyed,
	})
}

// generateAccessPolicyResource generates a basic access policy resource configuration for testing
func generateAccessPolicyResource(
	resourceLabel string,
	name string,
	description string,
	targetResource string,
	effect string,
	subjectType string,
	subjectId string,
	enabled string,
	conditionJSON string,
) string {
	subjectIdAttr := ""
	if subjectId != "" {
		subjectIdAttr = fmt.Sprintf(`subject_id = "%s"`, subjectId)
	}

	conditionAttr := ""
	if conditionJSON != "" {
		conditionAttr = fmt.Sprintf(`condition_json = jsonencode(%s)`, conditionJSON)
	}

	return fmt.Sprintf(`resource "genesyscloud_access_policy" "%s" {
		name            = "%s"
		description     = "%s"
		target_resource = "%s"
		effect          = "%s"
		subject_type    = "%s"
		%s
		enabled         = %s
		%s
	}
	`, resourceLabel, name, description, targetResource, effect, subjectType, subjectIdAttr, enabled, conditionAttr)
}

// generateAccessPolicyResourceWithCondition generates an access policy resource with a condition JSON string
func generateAccessPolicyResourceWithCondition(
	resourceLabel string,
	name string,
	description string,
	targetResource string,
	effect string,
	subjectType string,
	enabled string,
	conditionJSON string,
) string {
	return fmt.Sprintf(`resource "genesyscloud_access_policy" "%s" {
		name            = "%s"
		description     = "%s"
		target_resource = "%s"
		effect          = "%s"
		subject_type    = "%s"
		enabled         = %s
		condition_json  = jsonencode(%s)
	}
	`, resourceLabel, name, description, targetResource, effect, subjectType, enabled, conditionJSON)
}

// testVerifyAccessPolicyDestroyed verifies that all access policies created during the test have been destroyed
func testVerifyAccessPolicyDestroyed(state *terraform.State) error {
	authApi := platformclientv2.NewAuthorizationApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		policy, resp, err := authApi.GetAuthorizationPolicy(rs.Primary.ID)
		if policy != nil {
			return fmt.Errorf("access policy (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Policy not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}

	// All access policies confirmed destroyed
	return nil
}
