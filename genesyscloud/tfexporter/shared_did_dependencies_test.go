package tfexporter

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func testIvrConfig(dnis ...string) util.JsonMap {
	values := make([]interface{}, len(dnis))
	for i, d := range dnis {
		values[i] = d
	}
	return util.JsonMap{"name": "ivr", "dnis": values}
}

func testUserConfig(numbers ...interface{}) util.JsonMap {
	phones := make([]interface{}, len(numbers))
	for i, n := range numbers {
		phones[i] = map[string]interface{}{"number": n, "media_type": "PHONE", "type": "WORK"}
	}
	return util.JsonMap{
		"email":     "user@example.com",
		"addresses": []interface{}{map[string]interface{}{"phone_numbers": phones}},
	}
}

func ivrDep(label string) string {
	return "$dep$" + sharedDidIvrResourceType + "." + label + "$dep$"
}

func TestUnitSharedDidAddsDependencyForMatchingUserOnly(t *testing.T) {
	maps := map[string]ResourceJSONMaps{
		sharedDidIvrResourceType: {"direct_agent_routing": testIvrConfig("+19164634135")},
		sharedDidUserResourceType: {
			"supervisor": testUserConfig("+19164634135"),
			"agent":      testUserConfig("+19164634999"),
			"no_phone":   util.JsonMap{"email": "x@example.com"},
		},
	}

	added, skipped := addSharedDidDependenciesToConfigMaps(maps)

	if added != 1 || len(skipped) != 0 {
		t.Fatalf("expected 1 added and 0 skipped, got %d added and %d skipped", added, len(skipped))
	}
	users := maps[sharedDidUserResourceType]
	if got := users["supervisor"]["depends_on"]; !reflect.DeepEqual(got, []string{ivrDep("direct_agent_routing")}) {
		t.Errorf("unexpected depends_on for matching user: %v", got)
	}
	if _, ok := users["agent"]["depends_on"]; ok {
		t.Errorf("user without a shared DID must not get depends_on")
	}
	if _, ok := users["no_phone"]["depends_on"]; ok {
		t.Errorf("user without phone numbers must not get depends_on")
	}
	if _, ok := maps[sharedDidIvrResourceType]["direct_agent_routing"]["depends_on"]; ok {
		t.Errorf("IVR must never get a depends_on on the user")
	}
}

func TestUnitSharedDidMultipleIvrsAndNumbersAreDeduplicatedAndSorted(t *testing.T) {
	maps := map[string]ResourceJSONMaps{
		sharedDidIvrResourceType: {
			"ivr_b": testIvrConfig("+13175550002", "+13175550003"),
			"ivr_a": testIvrConfig("+13175550001"),
		},
		sharedDidUserResourceType: {
			// Two numbers on ivr_b and one on ivr_a, listed out of order and with a duplicate.
			"supervisor": testUserConfig("+13175550003", "+13175550001", "+13175550002", "+13175550002"),
		},
	}

	added, _ := addSharedDidDependenciesToConfigMaps(maps)

	if added != 2 {
		t.Fatalf("expected 2 dependencies, got %d", added)
	}
	want := []string{ivrDep("ivr_a"), ivrDep("ivr_b")}
	if got := maps[sharedDidUserResourceType]["supervisor"]["depends_on"]; !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestUnitSharedDidSameNumberOnTwoIvrsDependsOnBoth(t *testing.T) {
	maps := map[string]ResourceJSONMaps{
		sharedDidIvrResourceType: {
			"ivr_a": testIvrConfig("+13175550001"),
			"ivr_b": testIvrConfig("+13175550001"),
		},
		sharedDidUserResourceType: {"supervisor": testUserConfig("+13175550001")},
	}

	addSharedDidDependenciesToConfigMaps(maps)

	want := []string{ivrDep("ivr_a"), ivrDep("ivr_b")}
	if got := maps[sharedDidUserResourceType]["supervisor"]["depends_on"]; !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestUnitSharedDidNormalizesFormattingButRequiresExactMatch(t *testing.T) {
	maps := map[string]ResourceJSONMaps{
		sharedDidIvrResourceType: {"ivr": testIvrConfig("+1 (916) 463-4135")},
		sharedDidUserResourceType: {
			"formatted":   testUserConfig("+19164634135"),
			"suffix_only": testUserConfig("+4634135"),
			"no_plus":     testUserConfig("19164634135"),
			"longer":      testUserConfig("+191646341350"),
		},
	}

	added, _ := addSharedDidDependenciesToConfigMaps(maps)

	if added != 1 {
		t.Fatalf("expected only the formatted match, got %d", added)
	}
	users := maps[sharedDidUserResourceType]
	if _, ok := users["formatted"]["depends_on"]; !ok {
		t.Errorf("formatted number should match after normalization")
	}
	for _, label := range []string{"suffix_only", "no_plus", "longer"} {
		if _, ok := users[label]["depends_on"]; ok {
			t.Errorf("%s must not match", label)
		}
	}
}

func TestUnitSharedDidIgnoresEmptyInvalidAndExtensionOnlyValues(t *testing.T) {
	maps := map[string]ResourceJSONMaps{
		sharedDidIvrResourceType: {"ivr": testIvrConfig("", "not-a-number", "${var.dnis}")},
		sharedDidUserResourceType: {
			"empty":          testUserConfig(""),
			"nil":            testUserConfig(nil),
			"text":           testUserConfig("not-a-number"),
			"variable":       testUserConfig("${var.dnis}"),
			"extension_only": util.JsonMap{"addresses": []interface{}{map[string]interface{}{"phone_numbers": []interface{}{map[string]interface{}{"extension": "1234"}}}}},
		},
	}

	added, skipped := addSharedDidDependenciesToConfigMaps(maps)

	if added != 0 || len(skipped) != 0 {
		t.Fatalf("expected no dependencies, got %d added and %d skipped", added, len(skipped))
	}
}

func TestUnitSharedDidNoopWhenOneTypeIsMissing(t *testing.T) {
	onlyUsers := map[string]ResourceJSONMaps{
		sharedDidUserResourceType: {"supervisor": testUserConfig("+19164634135")},
	}
	if added, _ := addSharedDidDependenciesToConfigMaps(onlyUsers); added != 0 {
		t.Errorf("expected no change when IVRs are not exported (e.g. excluded), got %d", added)
	}
	if _, ok := onlyUsers[sharedDidUserResourceType]["supervisor"]["depends_on"]; ok {
		t.Errorf("must not reference an IVR that is not in the export")
	}

	onlyIvrs := map[string]ResourceJSONMaps{
		sharedDidIvrResourceType: {"ivr": testIvrConfig("+19164634135")},
	}
	if added, _ := addSharedDidDependenciesToConfigMaps(onlyIvrs); added != 0 {
		t.Errorf("expected no change when users are not exported, got %d", added)
	}

	if added, _ := addSharedDidDependenciesToConfigMaps(map[string]ResourceJSONMaps{}); added != 0 {
		t.Errorf("expected no change for an empty export")
	}
}

func TestUnitSharedDidMergesWithExistingDependsOn(t *testing.T) {
	user := testUserConfig("+19164634135")
	user["depends_on"] = []string{"$dep$genesyscloud_routing_queue.queue$dep$", ivrDep("ivr")}
	maps := map[string]ResourceJSONMaps{
		sharedDidIvrResourceType:  {"ivr": testIvrConfig("+19164634135")},
		sharedDidUserResourceType: {"supervisor": user},
	}

	added, _ := addSharedDidDependenciesToConfigMaps(maps)

	if added != 0 {
		t.Errorf("dependency already present, expected 0 added, got %d", added)
	}
	want := []string{"$dep$genesyscloud_routing_queue.queue$dep$", ivrDep("ivr")}
	if got := user["depends_on"]; !reflect.DeepEqual(got, want) {
		t.Errorf("existing depends_on must be preserved without duplicates, got %v", got)
	}

	other := testUserConfig("+19164634135")
	other["depends_on"] = []string{"$dep$genesyscloud_routing_queue.queue$dep$"}
	maps[sharedDidUserResourceType]["other"] = other
	addSharedDidDependenciesToConfigMaps(maps)
	want = []string{"$dep$genesyscloud_routing_queue.queue$dep$", ivrDep("ivr")}
	if got := other["depends_on"]; !reflect.DeepEqual(got, want) {
		t.Errorf("expected new dependency appended after existing ones, got %v", got)
	}
}

func TestUnitSharedDidIsIdempotent(t *testing.T) {
	maps := map[string]ResourceJSONMaps{
		sharedDidIvrResourceType:  {"ivr": testIvrConfig("+19164634135")},
		sharedDidUserResourceType: {"supervisor": testUserConfig("+19164634135")},
	}

	addSharedDidDependenciesToConfigMaps(maps)
	first := append([]string{}, maps[sharedDidUserResourceType]["supervisor"]["depends_on"].([]string)...)
	added, _ := addSharedDidDependenciesToConfigMaps(maps)

	if added != 0 {
		t.Errorf("second run must not add anything, added %d", added)
	}
	if got := maps[sharedDidUserResourceType]["supervisor"]["depends_on"]; !reflect.DeepEqual(got, first) {
		t.Errorf("second run changed depends_on: %v -> %v", first, got)
	}
}

func TestUnitSharedDidSkipsDependencyThatWouldCreateCycle(t *testing.T) {
	// IVR -> flow (interpolated reference) and flow -> user (depends_on). Adding user -> IVR would close the loop.
	ivr := testIvrConfig("+19164634135")
	ivr["open_hours_flow_id"] = "${genesyscloud_flow.inbound.id}"
	maps := map[string]ResourceJSONMaps{
		sharedDidIvrResourceType: {"ivr": ivr},
		"genesyscloud_flow": {"inbound": util.JsonMap{
			"filepath":   "flow.yaml",
			"depends_on": []string{"$dep$genesyscloud_user.supervisor$dep$"},
		}},
		sharedDidUserResourceType: {
			"supervisor": testUserConfig("+19164634135"),
			"other":      testUserConfig("+19164634135"),
		},
	}

	added, skipped := addSharedDidDependenciesToConfigMaps(maps)

	if added != 1 {
		t.Errorf("expected only the non-cyclic user to get a dependency, got %d", added)
	}
	if len(skipped) != 1 || skipped[0].UserBlockLabel != "supervisor" || skipped[0].IvrBlockLabel != "ivr" || skipped[0].Number != "+19164634135" {
		t.Fatalf("expected the cyclic dependency to be reported, got %+v", skipped)
	}
	if _, ok := maps[sharedDidUserResourceType]["supervisor"]["depends_on"]; ok {
		t.Errorf("cyclic dependency must not be written")
	}
	if _, ok := maps[sharedDidUserResourceType]["other"]["depends_on"]; !ok {
		t.Errorf("non-cyclic user should still get its dependency")
	}
}

func TestUnitSharedDidIgnoresDataSourceReferencesInCycleCheck(t *testing.T) {
	ivr := testIvrConfig("+19164634135")
	ivr["open_hours_flow_id"] = "${data.genesyscloud_flow.inbound.id}"
	maps := map[string]ResourceJSONMaps{
		sharedDidIvrResourceType:  {"ivr": ivr},
		sharedDidUserResourceType: {"supervisor": testUserConfig("+19164634135")},
	}

	added, skipped := addSharedDidDependenciesToConfigMaps(maps)

	if added != 1 || len(skipped) != 0 {
		t.Errorf("expected 1 added and 0 skipped, got %d added and %d skipped", added, len(skipped))
	}
}

func TestUnitSharedDidRendersRawReferenceInHcl(t *testing.T) {
	maps := map[string]ResourceJSONMaps{
		sharedDidIvrResourceType:  {"direct_agent_routing": testIvrConfig("+19164634135")},
		sharedDidUserResourceType: {"supervisor": testUserConfig("+19164634135")},
	}
	addSharedDidDependenciesToConfigMaps(maps)

	block := string(postProcessHclBytes(instanceStateToHCLBlock(sharedDidUserResourceType, "supervisor", maps[sharedDidUserResourceType]["supervisor"], false)))

	if !strings.Contains(block, "genesyscloud_architect_ivr.direct_agent_routing") {
		t.Fatalf("rendered HCL is missing the IVR reference:\n%s", block)
	}
	if strings.Contains(block, "$dep$") || strings.Contains(block, `"genesyscloud_architect_ivr.direct_agent_routing"`) {
		t.Errorf("depends_on reference must be unquoted and without markers:\n%s", block)
	}
}

func TestUnitSharedDidRendersQuotedReferenceInJson(t *testing.T) {
	maps := map[string]ResourceJSONMaps{
		sharedDidIvrResourceType:  {"direct_agent_routing": testIvrConfig("+19164634135")},
		sharedDidUserResourceType: {"supervisor": testUserConfig("+19164634135")},
	}
	addSharedDidDependenciesToConfigMaps(maps)

	out := string(postProcessJsonBytes([]byte(`{"depends_on": ["` + ivrDep("direct_agent_routing") + `"]}`)))

	if out != `{"depends_on": ["genesyscloud_architect_ivr.direct_agent_routing"]}` {
		t.Errorf("unexpected JSON output: %s", out)
	}
}
