package framework_custom_validators

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = durationValidator{}

type durationValidator struct{}

func (v durationValidator) Description(_ context.Context) string {
	return "value must be a valid duration string"
}

func (v durationValidator) MarkdownDescription(_ context.Context) string {
	return "value must be a valid duration string"
}

func (v durationValidator) ValidateString(_ context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()
	_, err := time.ParseDuration(value)
	if err != nil {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid Duration",
			fmt.Sprintf("Expected a valid duration string, got: %s. Error: %v", value, err),
		)
	}
}

func ValidateDuration() validator.String {
	return durationValidator{}
}
