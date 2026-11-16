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

func testActivityCode(id string, name string, category string, defaultCode bool, active bool) platformclientv2.Businessunitactivitycode {
	return platformclientv2.Businessunitactivitycode{
		Id:          platformclientv2.String(id),
		Name:        platformclientv2.String(name),
		Category:    platformclientv2.String(category),
		DefaultCode: platformclientv2.Bool(defaultCode),
		Active:      platformclientv2.Bool(active),
	}
}

func TestUnitFindActivityCodeIdByNameSkipsInactiveCodes(t *testing.T) {
	codes := []platformclientv2.Businessunitactivitycode{
		testActivityCode("old", "Training", "Training", false, false),
		testActivityCode("current", "Training", "Training", false, true),
	}
	id, err := findActivityCodeIdByName(codes, "Training")
	if err != nil || id != "current" {
		t.Errorf("expected the active code, got %q (%v)", id, err)
	}
	if _, err := findActivityCodeIdByName(codes, "Missing"); err == nil {
		t.Error("expected an error for a name that does not exist")
	}
}

func TestUnitFindActivityCodeIdByNameRejectsAnEmptyName(t *testing.T) {
	codes := []platformclientv2.Businessunitactivitycode{
		testActivityCode("break", "", "Break", true, true),
		testActivityCode("meal", "", "Meal", true, true),
	}
	if _, err := findActivityCodeIdByName(codes, ""); err == nil {
		t.Error("expected an error, because default activity codes cannot be told apart by an empty name")
	}
}

func TestUnitFindDefaultActivityCodeIdByCategory(t *testing.T) {
	codes := []platformclientv2.Businessunitactivitycode{
		testActivityCode("custom-break", "Short Break", "Break", false, true),
		testActivityCode("break", "", "Break", true, true),
		testActivityCode("meal", "", "Meal", true, true),
	}
	id, err := findDefaultActivityCodeIdByCategory(codes, "Break")
	if err != nil || id != "break" {
		t.Errorf("expected the default Break code, got %q (%v)", id, err)
	}
	if _, err := findDefaultActivityCodeIdByCategory(codes, "Training"); err == nil {
		t.Error("expected an error for a category without a default code")
	}
}
