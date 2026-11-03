package architect_flow

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const (
	ValidationDataSourceType = "genesyscloud_flow_validation"

	validateJobStatusSuccess = "Success"
	validateJobStatusFailure = "Failure"

	validateJobMessageTypeWarning = "Warning"

	defaultValidateJobTimeout = 5 * time.Minute
)

// validateJobPollInterval is a variable so unit tests can shorten it.
var validateJobPollInterval = 2 * time.Second

// flowValidationResult is the outcome of a completed validate job.
type flowValidationResult struct {
	JobId    string
	Status   string
	Errors   []string
	Warnings []string
}

func (r flowValidationResult) valid() bool {
	return r.Status == validateJobStatusSuccess
}

func DataSourceFlowValidation() *schema.Resource {
	return &schema.Resource{
		Description: `Validates an existing Genesys Cloud Architect flow by running a validate job (POST /api/v2/flows/validate/jobs) and waiting for it to finish.

Select the flow by ` + "`flow_id`" + `, or by ` + "`name`" + ` and ` + "`type`" + `. Validation runs every time the data source is read. Set ` + "`fail_on_error`" + ` to make the plan or apply fail when the flow has validation errors.

In addition to the permissions listed below, the OAuth client needs ` + "`architect:jobValidate:create`" + ` to register the validate job.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceFlowValidationRead),
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(defaultValidateJobTimeout),
		},
		Schema: map[string]*schema.Schema{
			"flow_id": {
				Description:  "ID of the flow to validate. Conflicts with `name` and `type`.",
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"flow_id", "name"},
			},
			"name": {
				Description:  "Name of the flow to validate. Requires `type`.",
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"type"},
			},
			"type": {
				Description:   "Type of the flow to validate. Requires `name`. Valid options: " + strings.Join(validFlowTypes, ", "),
				Type:          schema.TypeString,
				Optional:      true,
				RequiredWith:  []string{"name"},
				ConflictsWith: []string{"flow_id"},
				ValidateFunc:  validation.StringInSlice(validFlowTypes, true),
			},
			"version": {
				Description: "Version of the flow to validate. If omitted, the API chooses the version.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"fail_on_error": {
				Description: "If true, return an error diagnostic when validation fails. If false, failures are only reported through `valid` and `errors`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"job_id": {
				Description: "ID of the validate job.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "Final status of the validate job: `Success` or `Failure`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"valid": {
				Description: "True when the flow passed validation. Warnings do not make a flow invalid.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"errors": {
				Description: "Error messages returned by the validate job.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"warnings": {
				Description: "Warning messages returned by the validate job.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceFlowValidationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	p := getArchitectFlowProxy(sdkConfig)

	flowRef := buildValidateFlowReference(
		d.Get("flow_id").(string),
		d.Get("name").(string),
		d.Get("type").(string),
		d.Get("version").(string),
	)
	flowLabel := describeFlowReference(flowRef)

	result, resp, err := runFlowValidation(ctx, p, flowRef, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return util.BuildAPIDiagnosticError(ValidationDataSourceType, fmt.Sprintf("failed to validate flow %s: %s", flowLabel, err), resp)
	}

	d.SetId(result.JobId)
	_ = d.Set("job_id", result.JobId)
	_ = d.Set("status", result.Status)
	_ = d.Set("valid", result.valid())
	_ = d.Set("errors", result.Errors)
	_ = d.Set("warnings", result.Warnings)

	var diags diag.Diagnostics
	for _, w := range result.Warnings {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Flow %s validation warning", flowLabel),
			Detail:   w,
		})
	}

	if !result.valid() && d.Get("fail_on_error").(bool) {
		detail := strings.Join(result.Errors, "\n")
		if detail == "" {
			detail = fmt.Sprintf("Validate job %s finished with status %s and returned no error messages.", result.JobId, result.Status)
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Flow %s failed validation (job %s)", flowLabel, result.JobId),
			Detail:   detail,
		})
	}

	return diags
}

// buildValidateFlowReference builds the flow reference sent to the validate job.
// The API's flow type enum is uppercase (e.g. INBOUNDCALL), so type is uppercased.
func buildValidateFlowReference(flowId, name, varType, version string) platformclientv2.Architectflowreference {
	ref := platformclientv2.Architectflowreference{}
	if flowId != "" {
		ref.Id = platformclientv2.String(flowId)
	}
	if name != "" {
		ref.Name = platformclientv2.String(name)
	}
	if varType != "" {
		ref.VarType = platformclientv2.String(strings.ToUpper(varType))
	}
	if version != "" {
		ref.Version = platformclientv2.String(version)
	}
	return ref
}

// describeFlowReference returns a label for a flow reference for use in diagnostics.
func describeFlowReference(ref platformclientv2.Architectflowreference) string {
	if ref.Id != nil {
		return *ref.Id
	}
	return fmt.Sprintf("'%s' (%s)", derefString(ref.Name), derefString(ref.VarType))
}

// runFlowValidation starts a validate job for a single flow and polls it until it reaches
// Success or Failure, or until timeout elapses.
func runFlowValidation(ctx context.Context, p *architectFlowProxy, flowRef platformclientv2.Architectflowreference, timeout time.Duration) (*flowValidationResult, *platformclientv2.APIResponse, error) {
	job, resp, err := p.CreateFlowValidateJob(ctx, flowRef)
	if err != nil {
		return nil, resp, fmt.Errorf("error registering validate job: %w", err)
	}
	if job == nil || job.Id == nil {
		return nil, resp, fmt.Errorf("validate job registration returned no job ID")
	}
	jobId := *job.Id
	log.Printf("Registered flow validate job %s for flow %s", jobId, describeFlowReference(flowRef))

	deadline := time.Now().Add(timeout)
	for {
		state, resp, err := p.GetFlowValidateJob(ctx, jobId)
		if err != nil {
			return nil, resp, fmt.Errorf("error retrieving status of validate job %s: %w", jobId, err)
		}

		status := ""
		if state != nil && state.Status != nil {
			status = *state.Status
		}
		log.Printf("Flow validate job %s status: %s", jobId, status)

		if status == validateJobStatusSuccess || status == validateJobStatusFailure {
			errs, warnings := splitValidateJobMessages(state.Messages)
			return &flowValidationResult{
				JobId:    jobId,
				Status:   status,
				Errors:   errs,
				Warnings: warnings,
			}, resp, nil
		}

		if time.Now().After(deadline) {
			return nil, resp, fmt.Errorf("timed out after %s waiting for validate job %s to finish (last status: %q)", timeout, jobId, status)
		}

		select {
		case <-ctx.Done():
			return nil, resp, fmt.Errorf("validate job %s: %w", jobId, ctx.Err())
		case <-time.After(validateJobPollInterval):
		}
	}
}

// splitValidateJobMessages separates job messages into errors and warnings.
// Messages with an unknown type are treated as errors so they are never silently dropped.
func splitValidateJobMessages(messages *[]platformclientv2.Architectjobmessage) (errs []string, warnings []string) {
	errs, warnings = []string{}, []string{}
	if messages == nil {
		return errs, warnings
	}
	for _, msg := range *messages {
		text := derefString(msg.Text)
		if text == "" {
			continue
		}
		if msg.VarType != nil && *msg.VarType == validateJobMessageTypeWarning {
			warnings = append(warnings, text)
		} else {
			errs = append(errs, text)
		}
	}
	return errs, warnings
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
