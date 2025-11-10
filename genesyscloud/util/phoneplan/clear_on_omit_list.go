package phoneplan

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ClearOnOmitList applies SDK-like behavior by force-clearing a list
// when omitted from the user config but present in state.
type ClearOnOmitList struct{}

func (m ClearOnOmitList) Description(_ context.Context) string {
	return "Clears the list when omitted in config but present in state."
}

func (m ClearOnOmitList) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m ClearOnOmitList) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.ConfigValue.IsNull() && !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		emptyList := types.ListValueMust(req.StateValue.ElementType(ctx), []attr.Value{})
		resp.PlanValue = emptyList
	}
}
