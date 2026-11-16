package workforcemanagement_businessunits_activitycode

import (
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

func TestUnitActivityCodeLabelUsesTheCategoryForABlankDefaultName(t *testing.T) {
	blank := platformclientv2.Businessunitactivitycode{Name: platformclientv2.String(""), Category: platformclientv2.String("Break")}
	if got := activityCodeLabel(blank); got != "default_Break" {
		t.Errorf("expected default_Break, got %s", got)
	}

	named := platformclientv2.Businessunitactivitycode{Name: platformclientv2.String("Team Meeting"), Category: platformclientv2.String("Meeting")}
	if got := activityCodeLabel(named); got != "Team Meeting" {
		t.Errorf("expected the name to be used, got %s", got)
	}
}
