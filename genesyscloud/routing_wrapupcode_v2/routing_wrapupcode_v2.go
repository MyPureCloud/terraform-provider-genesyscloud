package routing_wrapupcode_v2

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v152/platformclientv2"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
)

var (
	_ resource.Resource = &WrapupCodeResource{}
)

type WrapupCodeResource struct {
	clientConfig *platformclientv2.Configuration
}

func (r *WrapupCodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Wrapup Code name.",
			},
			"division_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The division to which this routing wrapupcode will belong. If not set, * will be used to indicate all divisions.",
			},
		},
		Description: "Genesys Cloud Routing Wrapup Code",
	}
}

func (r *WrapupCodeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_wrapupcode_v2"
}

func (r *WrapupCodeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	log.Println("Calling create")

}

func (r *WrapupCodeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	log.Println("Calling read")
}

func (r *WrapupCodeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	log.Println("Calling update")
}

func (r *WrapupCodeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	log.Println("Calling delete")
}

func (r *WrapupCodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(provider.GenesysCloudProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *GenesysCloudProvider, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	clientConfig, err := provider.AcquireSdkClient(ctx, client)
	if err != nil {
		resp.Diagnostics.AddError("Failed to acquire sdk client from pool", err.Error())
		return
	}

	r.clientConfig = clientConfig
}

func NewWrapupCodeResource() resource.Resource {
	return &WrapupCodeResource{}
}
