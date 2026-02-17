package phoneplan

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nyaruka/phonenumbers"
)

type E164 struct {
	// DefaultRegion is used when numbers are not in international format.
	// Match the SDK's behavior ("US"), or make it configurable if you need.
	DefaultRegion string
}

func (m E164) Description(_ context.Context) string { return "Canonicalize phone number to E.164" }
func (m E164) MarkdownDescription(_ context.Context) string {
	return m.Description(context.Background())
}

func (m E164) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing for null/unknown
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	raw := strings.TrimSpace(req.ConfigValue.ValueString())
	if raw == "" {
		return
	}

	parsed, err := phonenumbers.Parse(raw, m.DefaultRegion)
	if err != nil {
		// Keep as-is if parsing fails (matches SDK "attempt" semantics)
		log.Printf("E164 plan modifier: parse failed; using raw number - input: %s, err: %s", raw, err.Error())
		return
	}
	formatted := phonenumbers.Format(parsed, phonenumbers.E164)
	resp.PlanValue = types.StringValue(formatted)
	log.Printf("E164 plan modifier: canonicalized - input: %s, e164: %s", raw, formatted)
}
