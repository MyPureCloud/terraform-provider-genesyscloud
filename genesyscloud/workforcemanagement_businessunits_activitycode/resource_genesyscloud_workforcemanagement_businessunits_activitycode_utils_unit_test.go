package workforcemanagement_businessunits_activitycode

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

func TestUnitActivityCodeBlockHashHandlesBusinessUnitNamesWithSpacesAndUnicode(t *testing.T) {
	inactive := testActivityCode("id-1", "Training", "Training", false, false)
	names := []string{"Support Team", "Support_Team", "Support  Team", " Support Team", "サポート チーム", "Équipe d'été", "Café\nNord", "Quote \" and \\ slash"}

	seen := map[string]string{}
	for _, name := range names {
		hash, err := activityCodeBlockHash(name, inactive)
		if err != nil || hash == "" {
			t.Fatalf("expected a hash for business unit %q, got %q (%v)", name, hash, err)
		}
		if again, _ := activityCodeBlockHash(name, inactive); hash != again {
			t.Errorf("the hash for business unit %q must be the same every time", name)
		}
		if previous, found := seen[hash]; found {
			t.Errorf("business units %q and %q must get different hashes", previous, name)
		}
		seen[hash] = name
	}
}

func planActivityCode(t *testing.T, active string) *terraform.InstanceDiff {
	t.Helper()
	state := &terraform.InstanceState{ID: "bu-1/code-1", Attributes: map[string]string{
		"id":                       "bu-1/code-1",
		"business_unit_id":         "bu-1",
		"name":                     "Training",
		"category":                 "Training",
		"active":                   active,
		"default_code":             "false",
		"length_in_minutes":        "30",
		"counts_as_paid_time":      "false",
		"counts_as_work_time":      "false",
		"counts_toward_shrinkage":  "false",
		"planned_shrinkage":        "false",
		"interruptible":            "false",
		"secondary_presence_ids.#": "0",
	}}
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"business_unit_id": "bu-1",
		"name":             "Training",
		"category":         "Training",
	})
	diff, err := ResourceBusinessUnitActivityCode().Diff(context.Background(), state, config, nil)
	if err != nil {
		t.Fatalf("unexpected error planning the activity code: %v", err)
	}
	return diff
}

func TestUnitReplaceWhenInactive(t *testing.T) {
	if diff := planActivityCode(t, "false"); diff == nil || !diff.RequiresNew() {
		t.Errorf("expected an inactive activity code that is still configured to be replaced, got %+v", diff)
	}
	if diff := planActivityCode(t, "true"); diff != nil && diff.RequiresNew() {
		t.Errorf("an active activity code must not be replaced, got %+v", diff)
	}
}
