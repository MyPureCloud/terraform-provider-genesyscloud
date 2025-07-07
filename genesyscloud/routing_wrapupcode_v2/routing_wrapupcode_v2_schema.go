package routing_wrapupcode_v2

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

var (
	_ resource.Resource = &WrapupCodeResource{}
)

type WrapupCodeResource struct {
	clientConfig *platformclientv2.Configuration
}

type WrapupCodeResourceApiModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DivisionId  types.String `tfsdk:"division_id"`
	Description types.String `tfsdk:"description"`
}

func (r *WrapupCodeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "API generated identifier for the wrap-up code.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Wrapup Code name.",
			},
			"division_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The division to which this routing wrapupcode will belong. If not set, * will be used to indicate all divisions.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The wrap-up code description.",
			},
		},
		Description: "Genesys Cloud Routing Wrapup Code",
	}
}

func (r *WrapupCodeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_wrapupcode_v2"
}
