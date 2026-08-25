package outbound_attempt_limit

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func TestUnitValidateResetPeriod(t *testing.T) {
	t.Parallel()

	validate := validation.StringInSlice(validResetPeriods, true)

	valid := []string{
		"NEVER",
		"TODAY",
		"never",
		"today",
		"DAYS_2",
		"DAYS_15",
		"DAYS_30",
		"days_30",
	}
	for _, value := range valid {
		if _, errs := validate(value, "reset_period"); len(errs) != 0 {
			t.Errorf("expected %q to be a valid reset_period, got %v", value, errs)
		}
	}

	invalid := []string{
		"DAYS_1",
		"DAYS_31",
		"DAYS_0",
		"DAY_30",
		"WEEKLY",
		"30",
	}
	for _, value := range invalid {
		if _, errs := validate(value, "reset_period"); len(errs) == 0 {
			t.Errorf("expected %q to be an invalid reset_period", value)
		}
	}
}
