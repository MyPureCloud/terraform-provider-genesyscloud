package task_management_worktype_status

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	workbin "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_workbin"
	workType "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/task_management_worktype"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccResourceTaskManagementWorktypeStatusDescription covers DEVTOOLING-1904 (GitHub issue #2663): removing
// `description` from a status must clear it in Genesys Cloud. Before the fix the PATCH request left the field out,
// so the old value stayed and the apply failed with "mismatch on attribute description".
func TestAccResourceTaskManagementWorktypeStatusDescription(t *testing.T) {
	t.Parallel()
	var (
		wbResourceLabel = "workbin_1"
		wbName          = "wb_" + uuid.NewString()

		wtResourceLabel = "worktype_1"
		wtName          = "wt_" + uuid.NewString()

		statusResourceLabel = "status_1"
		statusName          = "status-" + uuid.NewString()
		statusAddr          = ResourceType + "." + statusResourceLabel

		description1 = "first description"
		description2 = "second description"
	)

	// The status block is written by hand because the shared generator always writes a description attribute,
	// and this test needs to remove the attribute from the config completely.
	baseConfig := workbin.GenerateWorkbinResource(wbResourceLabel, wbName, "workbin for DEVTOOLING-1904", util.NullValue) +
		workType.GenerateWorktypeResourceBasic(
			wtResourceLabel,
			wtName,
			"worktype for DEVTOOLING-1904",
			fmt.Sprintf("genesyscloud_task_management_workbin.%s.id", wbResourceLabel),
			"",
		)
	statusConfig := func(descriptionLine string) string {
		return baseConfig + fmt.Sprintf(`
resource "%s" "%s" {
  worktype_id = genesyscloud_task_management_worktype.%s.id
  name        = "%s"
  category    = "Waiting"
  %s
}
`, ResourceType, statusResourceLabel, wtResourceLabel, statusName, descriptionLine)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create with a description
				Config: statusConfig(fmt.Sprintf(`description = "%s"`, description1)),
				Check:  resource.TestCheckResourceAttr(statusAddr, "description", description1),
			},
			{
				// Remove the description: it must be cleared, with no mismatch error and no follow-up diff
				Config: statusConfig(""),
				Check:  resource.TestCheckResourceAttr(statusAddr, "description", ""),
			},
			{
				// Set a description again after it was cleared
				Config: statusConfig(fmt.Sprintf(`description = "%s"`, description2)),
				Check:  resource.TestCheckResourceAttr(statusAddr, "description", description2),
			},
			{
				// Change the description to a different value
				Config: statusConfig(fmt.Sprintf(`description = "%s"`, description1)),
				Check:  resource.TestCheckResourceAttr(statusAddr, "description", description1),
			},
			{
				// Import/Read
				ResourceName:      statusAddr,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTaskManagementWorktypeStatusDestroyed,
	})
}
