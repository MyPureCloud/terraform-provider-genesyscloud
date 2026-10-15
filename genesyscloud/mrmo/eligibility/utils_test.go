package eligibility

import (
	"context"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

// ─── Fake registry for tests ─────────────────────────────────────────

type fakeRegistry struct {
	exporters map[string]*resourceExporter.ResourceExporter
	allTypes  []string
}

func (f *fakeRegistry) GetExporter(resourceType string) *resourceExporter.ResourceExporter {
	if f.exporters == nil {
		return nil
	}
	return f.exporters[resourceType]
}

func (f *fakeRegistry) GetAllResourceTypes() []string {
	if f.allTypes != nil {
		return f.allTypes
	}
	types := make([]string, 0, len(f.exporters))
	for t := range f.exporters {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

// ─── Helper builders ─────────────────────────────────────────────────

func basicExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: func(ctx context.Context) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
			return nil, nil
		},
	}
}

func exporterWithRefAttrs(refs map[string]string) *resourceExporter.ResourceExporter {
	exp := basicExporter()
	exp.RefAttrs = make(map[string]*resourceExporter.RefAttrSettings)
	for attr, refType := range refs {
		exp.RefAttrs[attr] = &resourceExporter.RefAttrSettings{RefType: refType}
	}
	return exp
}

func exporterWithExportAsDataFunc() *resourceExporter.ResourceExporter {
	exp := basicExporter()
	exp.ExportAsDataFunc = func(ctx context.Context, cfg *platformclientv2.Configuration, m map[string]string) (bool, error) {
		return true, nil
	}
	return exp
}

func exporterWithUnResolvableAttrs(attrs ...string) *resourceExporter.ResourceExporter {
	exp := basicExporter()
	exp.UnResolvableAttributes = make(map[string]*schema.Schema)
	for _, a := range attrs {
		exp.UnResolvableAttributes[a] = &schema.Schema{Type: schema.TypeString}
	}
	return exp
}

func singletonExporter() *resourceExporter.ResourceExporter {
	exp := basicExporter()
	exp.IsSingleton = true
	exp.ExportId = "singleton_type"
	return exp
}

func exporterWithEncodedRefAttrs(refType string) *resourceExporter.ResourceExporter {
	exp := basicExporter()
	exp.EncodedRefAttrs = map[*resourceExporter.JsonEncodeRefAttr]*resourceExporter.RefAttrSettings{
		{Attr: "json_field", NestedAttr: "nested_ref"}: {RefType: refType},
	}
	return exp
}

// ─── Tests ───────────────────────────────────────────────────────────

func TestUnit_TypeNotSupported(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("genesyscloud_nonexistent")

	assertResult(t, result, StatusNotApplicable, ReasonTypeNotSupported)
	if result.ResourceType != "genesyscloud_nonexistent" {
		t.Errorf("expected ResourceType 'genesyscloud_nonexistent', got %q", result.ResourceType)
	}
}

func TestUnit_ExportAsDataFuncIsReadyWithInfoFlag(t *testing.T) {
	// A type with ExportAsDataFunc is still replicable: only SOME instances get
	// emitted as data sources, decided per-instance during export. So the type is
	// Ready, with HasInstancesExportedByReference set for explainability.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_location": exporterWithExportAsDataFunc(),
	}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("genesyscloud_location")

	assertResult(t, result, StatusReady, "")
	if !result.HasInstancesExportedByReference {
		t.Error("expected HasInstancesExportedByReference = true")
	}
}

func TestUnit_ExcludedDenyList(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_routing_queue": basicExporter(),
	}}
	ev := newEvaluator(Config{
		DenyResourceTypes: []string{"genesyscloud_routing_queue"},
	}, reg)

	result := ev.evaluateType("genesyscloud_routing_queue")

	assertResult(t, result, StatusExcluded, ReasonExcludedFromRepl)
}

func TestUnit_ExcludedAllowListMiss(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_routing_queue": basicExporter(),
		"genesyscloud_user":          basicExporter(),
	}}
	ev := newEvaluator(Config{
		AllowResourceTypes: []string{"genesyscloud_user"}, // queue not in allow list
	}, reg)

	result := ev.evaluateType("genesyscloud_routing_queue")

	assertResult(t, result, StatusExcluded, ReasonExcludedFromRepl)
}

func TestUnit_AllowListHit(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_user": basicExporter(),
	}}
	ev := newEvaluator(Config{
		AllowResourceTypes: []string{"genesyscloud_user"},
	}, reg)

	result := ev.evaluateType("genesyscloud_user")

	assertResult(t, result, StatusReady, "")
}

func TestUnit_OutOfScope(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_telephony_providers_edges_site": basicExporter(),
	}}
	ev := newEvaluator(Config{
		OutOfScopeResourceTypes: []string{"genesyscloud_telephony_providers_edges_site"},
	}, reg)

	result := ev.evaluateType("genesyscloud_telephony_providers_edges_site")

	assertResult(t, result, StatusNotApplicable, ReasonExcludedFromRepl)
}

func TestUnit_DependencyNotReady(t *testing.T) {
	// Type A depends on type B, but B has no exporter.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": exporterWithRefAttrs(map[string]string{"b_id": "typeB"}),
		// typeB intentionally not registered
	}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("typeA")

	assertResult(t, result, StatusExcluded, ReasonDependencyNotReady)
	if len(result.Dependencies) == 0 {
		t.Error("expected Dependencies to be populated")
	}
}

func TestUnit_DependencyChainNotReady(t *testing.T) {
	// A → B → C; C has no exporter.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": exporterWithRefAttrs(map[string]string{"b_id": "typeB"}),
		"typeB": exporterWithRefAttrs(map[string]string{"c_id": "typeC"}),
		// typeC not registered
	}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("typeA")

	assertResult(t, result, StatusExcluded, ReasonDependencyNotReady)
}

func TestUnit_DependencyCycle(t *testing.T) {
	// A → B → A (mutual cycle). Both should resolve Ready.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": exporterWithRefAttrs(map[string]string{"b_id": "typeB"}),
		"typeB": exporterWithRefAttrs(map[string]string{"a_id": "typeA"}),
	}}
	ev := newEvaluator(Config{}, reg)

	resultA := ev.evaluateType("typeA")
	assertResult(t, resultA, StatusReady, "")

	resultB := ev.evaluateType("typeB")
	assertResult(t, resultB, StatusReady, "")
}

func TestUnit_SelfReference(t *testing.T) {
	// A references itself (e.g. routing_queue backup_queue_id → routing_queue).
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_routing_queue": exporterWithRefAttrs(map[string]string{
			"backup_queue_id": "genesyscloud_routing_queue",
		}),
	}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("genesyscloud_routing_queue")

	assertResult(t, result, StatusReady, "")
}

func TestUnit_OrgLockedAttrs_DefaultNotBlocking(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_routing_email_domain": exporterWithUnResolvableAttrs("custom_smtp_server_id"),
	}}
	ev := newEvaluator(Config{OrgLockedAttributesBlock: false}, reg)

	result := ev.evaluateType("genesyscloud_routing_email_domain")

	assertResult(t, result, StatusReady, "")
	if !result.HasOrgLockedAttributes {
		t.Error("expected HasOrgLockedAttributes = true")
	}
	if len(result.OrgLockedAttributeNames) != 1 || result.OrgLockedAttributeNames[0] != "custom_smtp_server_id" {
		t.Errorf("expected OrgLockedAttributeNames = [custom_smtp_server_id], got %v", result.OrgLockedAttributeNames)
	}
	if result.Message == "" {
		t.Error("expected an informational message about org-locked attributes")
	}
}

func TestUnit_OrgLockedAttrs_StrictMode(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_telephony_providers_edges_trunk": exporterWithUnResolvableAttrs("edge_id"),
	}}
	ev := newEvaluator(Config{OrgLockedAttributesBlock: true}, reg)

	result := ev.evaluateType("genesyscloud_telephony_providers_edges_trunk")

	assertResult(t, result, StatusNotReady, ReasonOrgLockedAttributes)
	if !result.HasOrgLockedAttributes {
		t.Error("expected HasOrgLockedAttributes = true")
	}
}

func TestUnit_Singleton(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_routing_settings": singletonExporter(),
	}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("genesyscloud_routing_settings")

	assertResult(t, result, StatusReady, "")
	if !result.IsSingleton {
		t.Error("expected IsSingleton = true")
	}
}

func TestUnit_ReadyClean(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"genesyscloud_auth_division": basicExporter(),
	}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("genesyscloud_auth_division")

	assertResult(t, result, StatusReady, "")
	if result.HasOrgLockedAttributes {
		t.Error("expected HasOrgLockedAttributes = false")
	}
	if result.IsSingleton {
		t.Error("expected IsSingleton = false")
	}
	if len(result.Dependencies) != 0 {
		t.Errorf("expected empty Dependencies, got %v", result.Dependencies)
	}
}

func TestUnit_EvaluateAll_MixedStatuses(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"readyType":      basicExporter(),
		"dataSourceType": exporterWithExportAsDataFunc(),
		"orgLockedType":  exporterWithUnResolvableAttrs("secret"),
		// "missingType" intentionally absent -> TYPE_NOT_SUPPORTED
	}}
	ev := newEvaluator(Config{}, reg)

	report := ev.evaluateAll("readyType", "dataSourceType", "orgLockedType", "missingType")

	if report.TotalEvaluated != 4 {
		t.Errorf("expected TotalEvaluated=4, got %d", report.TotalEvaluated)
	}
	// readyType + dataSourceType (ExportAsDataFunc is now Ready + info flag) +
	// orgLockedType (default non-blocking) = 3 Ready.
	if report.CountByStatus[StatusReady] != 3 {
		t.Errorf("expected 3 Ready, got %d", report.CountByStatus[StatusReady])
	}
	// Only missingType (no exporter) is NotApplicable.
	if report.CountByStatus[StatusNotApplicable] != 1 {
		t.Errorf("expected 1 NotApplicable, got %d", report.CountByStatus[StatusNotApplicable])
	}
	if len(report.Flagged) != 1 {
		t.Errorf("expected 1 Flagged, got %d", len(report.Flagged))
	}
}

func TestUnit_EvaluateAll_Subset(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": basicExporter(),
		"typeB": basicExporter(),
		"typeC": basicExporter(),
	}}
	ev := newEvaluator(Config{}, reg)

	report := ev.evaluateAll("typeA", "typeC") // only evaluate 2 of 3

	if report.TotalEvaluated != 2 {
		t.Errorf("expected TotalEvaluated=2, got %d", report.TotalEvaluated)
	}
}

func TestUnit_EvaluateAll_EmptyRegistry(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{}}
	ev := newEvaluator(Config{}, reg)

	report := ev.evaluateAll()

	if report.TotalEvaluated != 0 {
		t.Errorf("expected TotalEvaluated=0, got %d", report.TotalEvaluated)
	}
	if len(report.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(report.Results))
	}
}

func TestUnit_EvaluateAll_NoArgs_AllRegistered(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": basicExporter(),
		"typeB": basicExporter(),
	}}
	ev := newEvaluator(Config{}, reg)

	report := ev.evaluateAll() // no args => all registered

	if report.TotalEvaluated != 2 {
		t.Errorf("expected TotalEvaluated=2, got %d", report.TotalEvaluated)
	}
}

func TestUnit_EvaluateAll_DoesNotMutateCallerSlice(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": basicExporter(),
		"typeB": basicExporter(),
		"typeC": basicExporter(),
	}}
	ev := newEvaluator(Config{}, reg)

	// Deliberately unsorted so an in-place sort would be observable.
	input := []string{"typeC", "typeA", "typeB"}
	before := append([]string(nil), input...)

	_ = ev.evaluateAll(input...)

	for i := range before {
		if input[i] != before[i] {
			t.Fatalf("evaluateAll mutated the caller's slice: before=%v after=%v", before, input)
		}
	}
}

func TestUnit_EvaluateAll_DedupesInput(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": basicExporter(),
		"typeB": basicExporter(),
	}}
	ev := newEvaluator(Config{}, reg)

	// typeA passed three times, typeB once.
	report := ev.evaluateAll("typeA", "typeA", "typeB", "typeA")

	if report.TotalEvaluated != 2 {
		t.Errorf("expected TotalEvaluated=2 after de-dup, got %d", report.TotalEvaluated)
	}
	if report.CountByStatus[StatusReady] != 2 {
		t.Errorf("expected 2 Ready after de-dup, got %d", report.CountByStatus[StatusReady])
	}
	if len(report.Results) != 2 {
		t.Errorf("expected 2 results after de-dup, got %d", len(report.Results))
	}
}

func TestUnit_ConfigPrecedence_DenyBeatsMetadata(t *testing.T) {
	// Type would be Ready based on metadata alone, but is on the deny list.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"cleanType": basicExporter(),
	}}
	ev := newEvaluator(Config{
		DenyResourceTypes: []string{"cleanType"},
	}, reg)

	result := ev.evaluateType("cleanType")

	assertResult(t, result, StatusExcluded, ReasonExcludedFromRepl)
}

func TestUnit_ConfigPrecedence_OutOfScopeBeatsExportAsData(t *testing.T) {
	// Type has ExportAsDataFunc (which alone would be Ready + info flag) but is also
	// in OutOfScopeResourceTypes. Out-of-scope is checked first → NotApplicable wins.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"scopedType": exporterWithExportAsDataFunc(),
	}}
	ev := newEvaluator(Config{
		OutOfScopeResourceTypes: []string{"scopedType"},
	}, reg)

	result := ev.evaluateType("scopedType")

	assertResult(t, result, StatusNotApplicable, ReasonExcludedFromRepl)
}

func TestUnit_EncodedRefAttrs(t *testing.T) {
	// Type has EncodedRefAttrs pointing to a type with no exporter.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"parentType": exporterWithEncodedRefAttrs("missingChild"),
	}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("parentType")

	assertResult(t, result, StatusExcluded, ReasonDependencyNotReady)
}

func TestUnit_MultipleOrgLockedAttrs(t *testing.T) {
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"multiLocked": exporterWithUnResolvableAttrs("edge_id", "secret_field", "filepath"),
	}}
	ev := newEvaluator(Config{OrgLockedAttributesBlock: false}, reg)

	result := ev.evaluateType("multiLocked")

	assertResult(t, result, StatusReady, "")
	if len(result.OrgLockedAttributeNames) != 3 {
		t.Errorf("expected 3 org-locked attrs, got %v", result.OrgLockedAttributeNames)
	}
	// Should be sorted
	expected := []string{"edge_id", "filepath", "secret_field"}
	for i, name := range result.OrgLockedAttributeNames {
		if name != expected[i] {
			t.Errorf("expected attr[%d]=%q, got %q", i, expected[i], name)
		}
	}
}

func TestUnit_DependencyReadyChain(t *testing.T) {
	// A → B → C; all have exporters. Should all be Ready.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": exporterWithRefAttrs(map[string]string{"b_id": "typeB"}),
		"typeB": exporterWithRefAttrs(map[string]string{"c_id": "typeC"}),
		"typeC": basicExporter(),
	}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("typeA")
	assertResult(t, result, StatusReady, "")

	if len(result.Dependencies) != 1 || result.Dependencies[0] != "typeB" {
		t.Errorf("expected Dependencies=[typeB], got %v", result.Dependencies)
	}
}

func TestUnit_ThreeWayCycle(t *testing.T) {
	// A → B → C → A
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": exporterWithRefAttrs(map[string]string{"b_id": "typeB"}),
		"typeB": exporterWithRefAttrs(map[string]string{"c_id": "typeC"}),
		"typeC": exporterWithRefAttrs(map[string]string{"a_id": "typeA"}),
	}}
	ev := newEvaluator(Config{}, reg)

	// All three should resolve Ready (cycle doesn't block).
	for _, rt := range []string{"typeA", "typeB", "typeC"} {
		result := ev.evaluateType(rt)
		assertResult(t, result, StatusReady, "")
	}
}

func TestUnit_DependencyOnDeniedType(t *testing.T) {
	// A depends on B. B is on the deny list. A should get DEPENDENCY_NOT_READY.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": exporterWithRefAttrs(map[string]string{"b_id": "typeB"}),
		"typeB": basicExporter(),
	}}
	ev := newEvaluator(Config{
		DenyResourceTypes: []string{"typeB"},
	}, reg)

	result := ev.evaluateType("typeA")

	assertResult(t, result, StatusExcluded, ReasonDependencyNotReady)
}

func TestUnit_DependencyOnDataSourceType(t *testing.T) {
	// A depends on B. B has ExportAsDataFunc, so B is itself Ready (its per-instance
	// data-source handling happens during export). A ready dependency does not block
	// A → A is Ready.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": exporterWithRefAttrs(map[string]string{"b_id": "typeB"}),
		"typeB": exporterWithExportAsDataFunc(),
	}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("typeA")

	assertResult(t, result, StatusReady, "")
}

func TestUnit_DependencyOnExcludedTypeStillBlocks(t *testing.T) {
	// A depends on B. B is a real resource but is on the deny list (EXCLUDED).
	// An excluded dependency genuinely won't be in the target org → A is blocked.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": exporterWithRefAttrs(map[string]string{"b_id": "typeB"}),
		"typeB": basicExporter(),
	}}
	ev := newEvaluator(Config{DenyResourceTypes: []string{"typeB"}}, reg)

	result := ev.evaluateType("typeA")

	assertResult(t, result, StatusExcluded, ReasonDependencyNotReady)
}

func TestUnit_DependencyOnTypeNotSupportedStillBlocks(t *testing.T) {
	// A depends on B. B has no exporter (TYPE_NOT_SUPPORTED) → genuinely absent → blocks A.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": exporterWithRefAttrs(map[string]string{"b_id": "typeB"}),
		// typeB not registered
	}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("typeA")

	assertResult(t, result, StatusExcluded, ReasonDependencyNotReady)
}

func TestUnit_DependencyChainThroughReplicateByRef(t *testing.T) {
	// A → B (normal) → C (replicate-by-ref). C is satisfied, so B is Ready, so A is Ready.
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeA": exporterWithRefAttrs(map[string]string{"b_id": "typeB"}),
		"typeB": exporterWithRefAttrs(map[string]string{"c_id": "typeC"}),
		"typeC": exporterWithExportAsDataFunc(),
	}}
	ev := newEvaluator(Config{}, reg)

	assertResult(t, ev.evaluateType("typeA"), StatusReady, "")
	assertResult(t, ev.evaluateType("typeB"), StatusReady, "")
}

func TestUnit_NilRefAttrSettings(t *testing.T) {
	// RefAttrs map with a nil value entry — should not panic.
	exp := basicExporter()
	exp.RefAttrs = map[string]*resourceExporter.RefAttrSettings{
		"some_attr": nil,
	}
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeWithNilRef": exp,
	}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("typeWithNilRef")

	assertResult(t, result, StatusReady, "")
}

func TestUnit_EmptyRefType(t *testing.T) {
	// RefAttrs entry with empty RefType — should be ignored as a dependency.
	exp := basicExporter()
	exp.RefAttrs = map[string]*resourceExporter.RefAttrSettings{
		"some_attr": {RefType: ""},
	}
	reg := &fakeRegistry{exporters: map[string]*resourceExporter.ResourceExporter{
		"typeWithEmptyRef": exp,
	}}
	ev := newEvaluator(Config{}, reg)

	result := ev.evaluateType("typeWithEmptyRef")

	assertResult(t, result, StatusReady, "")
	if len(result.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %v", result.Dependencies)
	}
}

// ─── Assertion helper ────────────────────────────────────────────────

func assertResult(t *testing.T, result Result, expectedStatus Status, expectedReason ReasonCode) {
	t.Helper()
	if result.Status != expectedStatus {
		t.Errorf("expected Status=%q, got %q (reason=%q, message=%q)",
			expectedStatus, result.Status, result.ReasonCode, result.Message)
	}
	if expectedReason != "" && result.ReasonCode != expectedReason {
		t.Errorf("expected ReasonCode=%q, got %q (message=%q)",
			expectedReason, result.ReasonCode, result.Message)
	}
}
