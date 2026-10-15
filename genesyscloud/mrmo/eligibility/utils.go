package eligibility

import (
	"errors"
	"fmt"
	"sort"

	providerRegistrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider_registrar"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

// exporterLookup abstracts the provider's exporter registry so the classification
// logic can be unit-tested with fake exporters instead of starting the real provider.
type exporterLookup interface {
	// GetExporter returns the ResourceExporter for the given type, or nil if none.
	GetExporter(resourceType string) *resourceExporter.ResourceExporter
	// GetAllResourceTypes returns every registered resource type name.
	GetAllResourceTypes() []string
}

// registrarLookup is the production exporterLookup. It wraps provider_registrar,
// using the same accessors mrmo/exporter relies on.
type registrarLookup struct{}

func (registrarLookup) GetExporter(resourceType string) *resourceExporter.ResourceExporter {
	return providerRegistrar.GetResourceExporterByResourceType(resourceType)
}

func (registrarLookup) GetAllResourceTypes() []string {
	return providerRegistrar.GetResourceTypeNames()
}

// defaultLookup returns the production registry-backed lookup.
func defaultLookup() exporterLookup {
	return registrarLookup{}
}

// validateEvaluateInput validates the input to Evaluate.
func validateEvaluateInput(input EvaluateInput) error {
	if input.ResourceType == "" {
		return errors.New("'ResourceType' is a required field")
	}
	return nil
}

// evaluator holds the config + lookup + pre-built config sets for classification.
type evaluator struct {
	cfg    Config
	lookup exporterLookup

	denySet       map[string]bool
	allowSet      map[string]bool // nil means "no allow list"
	outOfScopeSet map[string]bool
}

// newEvaluator builds an evaluator with pre-computed config sets.
func newEvaluator(cfg Config, lookup exporterLookup) *evaluator {
	e := &evaluator{cfg: cfg, lookup: lookup}
	e.denySet = toSet(cfg.DenyResourceTypes)
	if len(cfg.AllowResourceTypes) > 0 {
		e.allowSet = toSet(cfg.AllowResourceTypes)
	}
	e.outOfScopeSet = toSet(cfg.OutOfScopeResourceTypes)
	return e
}

// evaluateType classifies a single resource type.
func (e *evaluator) evaluateType(resourceType string) Result {
	memo := map[string]Result{}
	visiting := map[string]bool{}
	return e.evaluateInternal(resourceType, memo, visiting)
}

// evaluateAll classifies every resource type (or the given subset) into a Report.
func (e *evaluator) evaluateAll(resourceTypes ...string) Report {
	if len(resourceTypes) == 0 {
		resourceTypes = e.lookup.GetAllResourceTypes()
	}

	// Copy and de-duplicate the input before sorting. Copying avoids mutating the
	// caller's slice (or the registry's backing slice from GetAllResourceTypes),
	// and de-duplication keeps TotalEvaluated/CountByStatus accurate if a caller
	// passes the same type more than once.
	types := dedupeSorted(resourceTypes)

	memo := map[string]Result{}
	visiting := map[string]bool{}

	results := make([]Result, 0, len(types))
	for _, rt := range types {
		results = append(results, e.evaluateInternal(rt, memo, visiting))
	}

	counts := map[Status]int{}
	flagged := make([]Result, 0)
	for _, r := range results {
		counts[r.Status]++
		if r.Status != StatusReady {
			flagged = append(flagged, r)
		}
	}

	return Report{
		Results:        results,
		CountByStatus:  counts,
		Flagged:        flagged,
		TotalEvaluated: len(results),
	}
}

// evaluateInternal is the recursive core. memo caches results; visiting detects cycles.
func (e *evaluator) evaluateInternal(rt string, memo map[string]Result, visiting map[string]bool) Result {
	if result, ok := memo[rt]; ok {
		return result
	}

	// Cycle detection: a mutual reference (e.g. routing_queue backup → routing_queue)
	// does not make a type ineligible.
	//
	// IMPORTANT: do NOT memoize this cycle sentinel. It is a provisional "assume
	// Ready" used only to break the recursion for the node currently being
	// evaluated higher up the stack. That in-progress node computes and memoizes
	// its real verdict when it unwinds. Caching this sentinel would poison the memo
	// with a partial result.
	if visiting[rt] {
		return Result{
			ResourceType: rt,
			Status:       StatusReady,
			Message:      "Resolved as Ready (cycle detected in dependency chain)",
		}
	}

	visiting[rt] = true
	defer delete(visiting, rt)

	exporter := e.lookup.GetExporter(rt)

	// Step 1: no exporter registered → not supported for replication.
	if exporter == nil {
		return e.memoize(memo, rt, Result{
			ResourceType: rt,
			Status:       StatusNotApplicable,
			ReasonCode:   ReasonTypeNotSupported,
			Message:      fmt.Sprintf("No ResourceExporter registered for type %q", rt),
		})
	}

	// Step 2: out-of-scope (config).
	if e.outOfScopeSet[rt] {
		return e.memoize(memo, rt, Result{
			ResourceType: rt,
			Status:       StatusNotApplicable,
			ReasonCode:   ReasonExcludedFromRepl,
			Message:      fmt.Sprintf("Type %q is configured as out of replication scope", rt),
		})
	}

	// Step 3: allow-list miss (config).
	if e.allowSet != nil && !e.allowSet[rt] {
		return e.memoize(memo, rt, Result{
			ResourceType: rt,
			Status:       StatusExcluded,
			ReasonCode:   ReasonExcludedFromRepl,
			Message:      fmt.Sprintf("Type %q is not in the configured allow list", rt),
		})
	}

	// Step 4: deny list (config).
	if e.denySet[rt] {
		return e.memoize(memo, rt, Result{
			ResourceType: rt,
			Status:       StatusExcluded,
			ReasonCode:   ReasonExcludedFromRepl,
			Message:      fmt.Sprintf("Type %q is in the configured deny list", rt),
		})
	}

	// Note on ExportAsDataFunc: it is NOT a type-level blocker. It is a per-instance
	// concern — only some instances (static/built-in ones that already exist in the
	// target org) get emitted as data sources, while others are created normally.
	// The export itself makes that decision per instance, so the type remains
	// replicable. We record it below as an informational flag only.
	exportsSomeByReference := exporter.ExportAsDataFunc != nil

	// Step 5: dependency closure.
	//
	// A dependency is "satisfied" if it will be present in the target org — i.e. it
	// is itself Ready. A dependency blocks only when it genuinely will not be in the
	// target org: TYPE_NOT_SUPPORTED, EXCLUDED (deny/allow/dependency), out-of-scope,
	// or NOT_READY. (Types that export some instances by reference are now classified
	// Ready, so depending on one is naturally satisfied here.)
	deps := collectDependencies(exporter)
	for _, dep := range deps {
		if dep == rt {
			continue // self-reference — not a blocker
		}
		depResult := e.evaluateInternal(dep, memo, visiting)
		if !dependencySatisfied(depResult) {
			return e.memoize(memo, rt, Result{
				ResourceType: rt,
				Status:       StatusExcluded,
				ReasonCode:   ReasonDependencyNotReady,
				Message:      fmt.Sprintf("Type %q depends on %q which is %s (%s)", rt, dep, depResult.Status, depResult.ReasonCode),
				Dependencies: deps,
			})
		}
	}

	// Org-locked attribute info (informational even when not blocking).
	orgLockedNames := sortedKeys(exporter.UnResolvableAttributes)
	hasOrgLocked := len(orgLockedNames) > 0

	// Step 6: org-locked attributes in strict mode.
	if hasOrgLocked && e.cfg.OrgLockedAttributesBlock {
		return e.memoize(memo, rt, Result{
			ResourceType:                    rt,
			Status:                          StatusNotReady,
			ReasonCode:                      ReasonOrgLockedAttributes,
			Message:                         fmt.Sprintf("Type %q has org-locked attributes %v that cannot be resolved across orgs", rt, orgLockedNames),
			Dependencies:                    deps,
			HasOrgLockedAttributes:          true,
			OrgLockedAttributeNames:         orgLockedNames,
			IsSingleton:                     exporter.IsSingleton,
			HasInstancesExportedByReference: exportsSomeByReference,
		})
	}

	// Step 7: Ready.
	result := Result{
		ResourceType:                    rt,
		Status:                          StatusReady,
		Dependencies:                    deps,
		HasOrgLockedAttributes:          hasOrgLocked,
		OrgLockedAttributeNames:         orgLockedNames,
		IsSingleton:                     exporter.IsSingleton,
		HasInstancesExportedByReference: exportsSomeByReference,
	}
	if hasOrgLocked {
		result.Message = fmt.Sprintf("Ready; org-locked attributes %v need manual edit in target org", orgLockedNames)
	}
	return e.memoize(memo, rt, result)
}

// memoize stores and returns a result.
func (e *evaluator) memoize(memo map[string]Result, rt string, result Result) Result {
	memo[rt] = result
	return result
}

// dependencySatisfied reports whether a dependency's verdict means it will be
// present in the target org (so it does not block the dependent resource).
//
// A dependency is satisfied only when it is itself Ready. It is not satisfied for
// TYPE_NOT_SUPPORTED, EXCLUDED, out-of-scope NotApplicable, or NotReady, because in
// those cases the dependency genuinely will not be created/available in the target
// org. (Types that export some instances by reference are classified Ready, so they
// are satisfied here without a special case.)
func dependencySatisfied(dep Result) bool {
	return dep.Status == StatusReady
}

// collectDependencies extracts unique referenced resource types from the exporter's
// RefAttrs and EncodedRefAttrs. Returns a sorted, deduplicated slice.
func collectDependencies(exp *resourceExporter.ResourceExporter) []string {
	seen := map[string]bool{}

	for _, ref := range exp.RefAttrs {
		if ref != nil && ref.RefType != "" {
			seen[ref.RefType] = true
		}
	}
	for _, ref := range exp.EncodedRefAttrs {
		if ref != nil && ref.RefType != "" {
			seen[ref.RefType] = true
		}
	}

	result := make([]string, 0, len(seen))
	for t := range seen {
		result = append(result, t)
	}
	sort.Strings(result)
	return result
}

// dedupeSorted returns a new sorted slice with duplicates removed. The input slice
// is not modified.
func dedupeSorted(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	sort.Strings(out)
	return out
}

func toSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
