package phoneplan

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NullIfEmpty struct{}

func (m NullIfEmpty) Description(_ context.Context) string { return "Set string to null when empty" }
func (m NullIfEmpty) MarkdownDescription(_ context.Context) string {
	return m.Description(context.Background())
}

func (m NullIfEmpty) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if strings.TrimSpace(req.ConfigValue.ValueString()) == "" {
		resp.PlanValue = types.StringNull()
	}
}
