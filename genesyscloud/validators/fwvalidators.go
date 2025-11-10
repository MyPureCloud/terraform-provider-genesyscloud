package validators

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
)

type e164PhoneNumberValidator struct{}

func FWValidatePhoneNumber() validator.String {
	return &e164PhoneNumberValidator{}
}

func (v *e164PhoneNumberValidator) Description(context.Context) string {
	return "Validates that a string is a phone number in E.164 format."
}

func (v *e164PhoneNumberValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *e164PhoneNumberValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Treat null/unknown as valid (presence is governed by Optional/Required).
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Preserve feature toggle behavior from SDKv2.
	if feature_toggles.BcpModeEnabledExists() {
		return
	}

	number := req.ConfigValue.ValueString()

	utilE164 := util.NewUtilE164Service()
	valid, err := utilE164.IsValidE164Number(number)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Phone Number Validation Error",
			fmt.Sprintf("Error validating E.164 phone number: %v", err),
		)
		return
	}
	if !valid {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid E.164 Phone Number",
			fmt.Sprintf("Value %q is not a valid E.164 phone number (e.g., +15551234567).", number),
		)
	}
}

// FWValidateDate validates an ISO-8601 date string in yyyy-MM-dd format
func FWValidateDate() validator.String {
	return fwValidateDate{}
}

type fwValidateDate struct{}

func (fwValidateDate) Description(_ context.Context) string {
	return "Value must be a valid ISO-8601 date: yyyy-MM-dd."
}

func (fwValidateDate) MarkdownDescription(ctx context.Context) string {
	return fwValidateDate{}.Description(ctx)
}

func (fwValidateDate) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Use the standard date format used throughout the codebase
	if _, err := time.Parse("2006-01-02", req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid date format",
			"Expected the date in ISO-8601 format (yyyy-MM-dd).",
		)
	}
}
