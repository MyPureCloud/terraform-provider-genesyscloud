package workforcemanagement_business_unit_activitycode

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The resource_genesyscloud_workforcemanagement_business_unit_activitycode_test.go contains all of the test cases for running the resource
tests for workforcemanagement_business_unit_activitycode.
*/

func TestUnitSplitBusinessUnitActivityCodeTerraformId(t *testing.T) {
	businessUnitId, activityCodeId := SplitBusinessUnitActivityCodeTerraformId("bu-1/code-2")
	if businessUnitId != "bu-1" || activityCodeId != "code-2" {
		t.Errorf("expected bu-1 and code-2, got %s and %s", businessUnitId, activityCodeId)
	}
}

func TestAccResourceBusinessUnitActivityCode(t *testing.T) {
	var (
		resourceLabel = "activity_code_1"
		name1         = "Test Activity Code " + uuid.NewString()
		name2         = "Test Activity Code Updated " + uuid.NewString()
		category      = "Meeting"
	)

	businessUnitId := createTestBusinessUnit(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateBusinessUnitActivityCodeResource(
					resourceLabel,
					strconvQuote(businessUnitId),
					name1,
					category,
					"length_in_minutes = 30",
					"counts_as_paid_time = true",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "business_unit_id", businessUnitId),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "category", category),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "length_in_minutes", "30"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "counts_as_paid_time", util.TrueValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "default_code", util.FalseValue),
				),
			},
			{
				Config: GenerateBusinessUnitActivityCodeResource(
					resourceLabel,
					strconvQuote(businessUnitId),
					name2,
					category,
					"length_in_minutes = 45",
					"counts_as_paid_time = false",
					"counts_as_work_time = true",
					"interruptible = true",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "length_in_minutes", "45"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "counts_as_paid_time", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "counts_as_work_time", util.TrueValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "interruptible", util.TrueValue),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyBusinessUnitActivityCodeDestroyed,
	})
}

func strconvQuote(s string) string {
	return fmt.Sprintf("%q", s)
}

// createTestBusinessUnit creates a business unit for the test to attach activity codes to and removes it when the test finishes.
// There is no business unit resource in this provider so the unit is managed directly through the api.
func createTestBusinessUnit(t *testing.T) string {
	sdkConfig, err := provider.AuthorizeSdk()
	if err != nil {
		t.Skipf("unable to authorize the sdk: %s", err)
	}

	api := platformclientv2.NewWorkforceManagementApiWithConfig(sdkConfig)
	businessUnit, _, err := api.PostWorkforcemanagementBusinessunits(platformclientv2.Createbusinessunitrequest{
		Name: platformclientv2.String("tf_test_bu_" + uuid.NewString()[:8]),
		Settings: &platformclientv2.Createbusinessunitsettingsrequest{
			StartDayOfWeek: platformclientv2.String("Sunday"),
			TimeZone:       platformclientv2.String("Etc/UTC"),
		},
	}, false)
	if err != nil {
		t.Fatalf("failed to create test business unit: %s", err)
	}

	t.Cleanup(func() {
		if _, err := api.DeleteWorkforcemanagementBusinessunit(*businessUnit.Id); err != nil {
			t.Logf("failed to delete test business unit %s: %s", *businessUnit.Id, err)
		}
	})

	return *businessUnit.Id
}

func testVerifyBusinessUnitActivityCodeDestroyed(state *terraform.State) error {
	workforceManagementApi := platformclientv2.NewWorkforceManagementApi()
	for _, res := range state.RootModule().Resources {
		if res.Type != ResourceType {
			continue
		}

		businessUnitId, activityCodeId := SplitBusinessUnitActivityCodeTerraformId(res.Primary.ID)
		activityCode, resp, err := workforceManagementApi.GetWorkforcemanagementBusinessunitActivitycode(businessUnitId, activityCodeId)
		if activityCode != nil && isActive(activityCode) {
			return fmt.Errorf("business unit activity code (%s) still exists", res.Primary.ID)
		}
		if activityCode != nil || util.IsStatus404(resp) {
			// Deleted activity codes are either gone or marked inactive
			continue
		}
		return fmt.Errorf("unexpected error: %s", err)
	}

	return nil
}
