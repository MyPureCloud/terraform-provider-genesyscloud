package eligibility

// Status represents the replication readiness of a resource type.
type Status string

const (
	// StatusReady indicates the resource is CXasCode-enabled and meets all readiness criteria.
	StatusReady Status = "REPLICATION_READY"

	// StatusExcluded indicates the resource is supported by CXasCode but excluded from this
	// replication scope (e.g. via deny list, or because a dependency is not replication-ready).
	StatusExcluded Status = "REPLICATION_EXCLUDED"

	// StatusNotApplicable indicates the resource type is outside the scope of replication
	// (e.g. no exporter registered, or only replicable by data-source reference).
	StatusNotApplicable Status = "REPLICATION_NOT_APPLICABLE"

	// StatusNotReady indicates the resource is applicable for replication but fails a readiness
	// check (e.g. org-locked attributes when strict mode is enabled).
	StatusNotReady Status = "REPLICATION_NOT_READY"
)

// ReasonCode explains why a resource is not Ready.
type ReasonCode string

const (
	// ReasonTypeNotSupported indicates no ResourceExporter is registered for the type.
	ReasonTypeNotSupported ReasonCode = "TYPE_NOT_SUPPORTED"

	// ReasonExcludedFromRepl indicates the type was excluded by configuration
	// (deny list or allow-list miss).
	ReasonExcludedFromRepl ReasonCode = "EXCLUDED_FROM_REPLICATION"

	// ReasonDependencyNotReady indicates one or more dependencies (via RefAttrs) are
	// not replication-ready.
	ReasonDependencyNotReady ReasonCode = "DEPENDENCY_NOT_READY"

	// ReasonOrgLockedAttributes indicates the exporter has UnResolvableAttributes
	// (org/region-locked values) and strict mode (OrgLockedAttributesBlock) is enabled.
	ReasonOrgLockedAttributes ReasonCode = "ORG_LOCKED_ATTRIBUTES"
)

// Result is the per-resource-type eligibility verdict.
type Result struct {
	// ResourceType is the Terraform resource type name (e.g. "genesyscloud_routing_queue").
	ResourceType string `json:"resourceType"`

	// Status is the replication readiness classification.
	Status Status `json:"status"`

	// ReasonCode explains why the resource is not Ready. Empty when Status is Ready.
	ReasonCode ReasonCode `json:"reasonCode,omitempty"`

	// Message is a human-readable explanation of the verdict.
	Message string `json:"message,omitempty"`

	// Dependencies lists the resource types this type references (from RefAttrs/EncodedRefAttrs).
	Dependencies []string `json:"dependencies,omitempty"`

	// HasOrgLockedAttributes is true when the exporter declares UnResolvableAttributes.
	// Populated even when Status is Ready (informational).
	HasOrgLockedAttributes bool `json:"hasOrgLockedAttributes,omitempty"`

	// OrgLockedAttributeNames lists the attribute paths that are org/region-locked.
	// Populated even when Status is Ready (informational).
	OrgLockedAttributeNames []string `json:"orgLockedAttributeNames,omitempty"`

	// IsSingleton is true when the exporter manages a single org-wide resource instance.
	// Informational — replication semantics differ for singletons (update-in-place, not create).
	IsSingleton bool `json:"isSingleton,omitempty"`

	// HasInstancesExportedByReference is true when the exporter declares an
	// ExportAsDataFunc, meaning SOME instances of this type may be static/built-in
	// and get emitted as data sources (referenced in the target org rather than
	// created). This is informational only and does NOT block replication: whether
	// a given instance is created or referenced is decided per-instance during
	// export, so the type as a whole remains replicable.
	HasInstancesExportedByReference bool `json:"hasInstancesExportedByReference,omitempty"`
}

// Report is the batch evaluation output returned by EvaluateAll.
type Report struct {
	// Results contains the verdict for every evaluated resource type.
	Results []Result `json:"results"`

	// CountByStatus summarizes how many types fell into each status.
	CountByStatus map[Status]int `json:"countByStatus"`

	// Flagged contains only the non-Ready results (with reasons).
	Flagged []Result `json:"flagged"`

	// TotalEvaluated is the total number of types evaluated.
	TotalEvaluated int `json:"totalEvaluated"`
}

// EvaluateInput is the input to Evaluate (mirrors exporter.ExportInput).
type EvaluateInput struct {
	// ResourceType is the Terraform resource type to evaluate (e.g. "genesyscloud_routing_queue").
	ResourceType string
}

// EvaluateOutput is the result of Evaluate (mirrors exporter.ExportOutput).
type EvaluateOutput struct {
	// Result is the per-type eligibility verdict.
	Result Result `json:"result"`
}

// EvaluateAllInput is the input to EvaluateAll (mirrors exporter.ExportByTypeInput).
type EvaluateAllInput struct {
	// ResourceTypes optionally restricts evaluation to a subset. If empty, all
	// registered resource types are evaluated.
	ResourceTypes []string
}

// EvaluateAllOutput is the result of EvaluateAll (mirrors exporter.ExportByTypeOutput).
type EvaluateAllOutput struct {
	// Report is the batch eligibility summary.
	Report Report `json:"report"`
}

// Config controls the eligibility evaluator's behavior.
type Config struct {
	// DenyResourceTypes lists types that should be classified as Excluded / EXCLUDED_FROM_REPLICATION.
	DenyResourceTypes []string

	// AllowResourceTypes, when non-empty, means anything NOT in this list is
	// classified as Excluded / EXCLUDED_FROM_REPLICATION.
	AllowResourceTypes []string

	// OutOfScopeResourceTypes lists types outside replication scope → NotApplicable.
	OutOfScopeResourceTypes []string

	// OrgLockedAttributesBlock controls handling of UnResolvableAttributes.
	// false (default): types with org-locked attributes are Ready with an informational note.
	// true: those types are classified as NotReady / ORG_LOCKED_ATTRIBUTES.
	OrgLockedAttributesBlock bool
}
