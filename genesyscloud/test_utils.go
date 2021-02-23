package genesyscloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const nullValue = "null"
const trueValue = "true"
const falseValue = "false"

// Verify default division is home division
func testDefaultHomeDivision(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		homeDivID, err := getHomeDivisionID()
		if err != nil {
			return fmt.Errorf("Failed to query home division: %v", err)
		}

		r := state.RootModule().Resources[resource]
		if r == nil {
			return fmt.Errorf("%s not found in state", resource)
		}

		a := r.Primary.Attributes

		if a["division_id"] != homeDivID {
			return fmt.Errorf("expected division to be home division %s", homeDivID)
		}

		return nil
	}
}
