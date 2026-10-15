// Package eligibility is the MRMO-facing API for determining whether a CX as Code
// resource type can be replicated from a source org to a target org.
//
// It mirrors the shape of the sibling genesyscloud/mrmo/exporter package: MRMO
// tooling (the replicator) imports this package and calls Evaluate / EvaluateAll
// in-process, the same way it calls exporter.Export / exporter.ExportByType.
//
// Unlike exporter, eligibility evaluation is type-level and purely metadata-driven:
// it reads the provider's registered ResourceExporter metadata and therefore needs
// no org, no credentials, and makes no API calls. The entry functions still return
// diag.Diagnostics to match the exporter package's calling convention so the
// replicator handles both the same way.
package eligibility

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// Evaluate classifies the replication eligibility of a single resource type.
//
// It returns an EvaluateOutput carrying the Result, and a diagnostics object which
// can contain errors (e.g. invalid input) or warnings. A resource type that is not
// supported is NOT an error: it is reported as Result.Status == NotApplicable with
// ReasonCode == TYPE_NOT_SUPPORTED.
func Evaluate(input EvaluateInput, cfg Config) (resp *EvaluateOutput, diags diag.Diagnostics) {
	if err := validateEvaluateInput(input); err != nil {
		return nil, diag.FromErr(err)
	}

	ev := newEvaluator(cfg, defaultLookup())
	result := ev.evaluateType(input.ResourceType)

	log.Printf("Evaluated replication eligibility for %q: %s (%s)", result.ResourceType, result.Status, result.ReasonCode)

	return &EvaluateOutput{Result: result}, diags
}

// EvaluateAll classifies the replication eligibility of every resource type (or the
// provided subset). It returns an EvaluateAllOutput carrying the Report, and a
// diagnostics object.
//
// If input.ResourceTypes is empty, all registered resource types are evaluated.
func EvaluateAll(input EvaluateAllInput, cfg Config) (resp *EvaluateAllOutput, diags diag.Diagnostics) {
	ev := newEvaluator(cfg, defaultLookup())
	report := ev.evaluateAll(input.ResourceTypes...)

	log.Printf("Evaluated replication eligibility for %d resource types (%d flagged)", report.TotalEvaluated, len(report.Flagged))

	return &EvaluateAllOutput{Report: report}, diags
}
