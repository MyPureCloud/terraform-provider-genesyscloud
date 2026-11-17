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

func TestUnitActivityCodeBlockHashOnlyForInactiveCodes(t *testing.T) {
	active := testActivityCode("a", "Training", "Training", false, true)
	inactive := testActivityCode("b", "Training", "Training", false, false)

	if hash, err := activityCodeBlockHash("bu-1", active); err != nil || hash != "" {
		t.Errorf("expected no hash for an active code, got %q (%v)", hash, err)
	}
	hash, err := activityCodeBlockHash("bu-1", inactive)
	if err != nil || hash == "" {
		t.Errorf("expected a hash for an inactive code, got %q (%v)", hash, err)
	}
	if again, _ := activityCodeBlockHash("bu-1", inactive); hash != again {
		t.Error("the hash must be the same every time")
	}
}

func TestUnitActivityCodeBlockHashTellsInactiveCodesApart(t *testing.T) {
	first := testActivityCode("id-1", "Training", "Training", false, false)
	second := testActivityCode("id-2", "Training", "Training", false, false)

	a, _ := activityCodeBlockHash("bu-1", first)
	b, _ := activityCodeBlockHash("bu-1", second)
	if a == b {
		t.Error("two inactive codes with the same name and category must still get different hashes")
	}

	// The same code id can appear in different business units, so the business unit has to be part of the hash
	other, _ := activityCodeBlockHash("bu-2", first)
	if a == other {
		t.Error("the same code id in another business unit must get a different hash")
	}
}
