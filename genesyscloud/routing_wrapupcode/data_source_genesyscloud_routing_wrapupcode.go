package routing_wrapupcode

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

// Ensure routingWrapupcodeFrameworkDataSource satisfies various data source interfaces.
var (
	_ datasource.DataSource              = &routingWrapupcodeFrameworkDataSource{}
	_ datasource.DataSourceWithConfigure = &routingWrapupcodeFrameworkDataSource{}
)

// routingWrapupcodeFrameworkDataSource defines the data source implementation for Plugin Framework.
type routingWrapupcodeFrameworkDataSource struct {
	clientConfig *platformclientv2.Configuration
}

// routingWrapupcodeFrameworkDataSourceModel describes the data source data model.
type routingWrapupcodeFrameworkDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewRoutingWrapupcodeFrameworkDataSource is a helper function to simplify the provider implementation.
func NewRoutingWrapupcodeFrameworkDataSource() datasource.DataSource {
	return &routingWrapupcodeFrameworkDataSource{}
}

func (d *routingWrapupcodeFrameworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_wrapupcode"
}

func (d *routingWrapupcodeFrameworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for Genesys Cloud Wrap-up Code. Select a wrap-up code by name",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The globally unique identifier for the wrapup code.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Wrap-up code name.",
				Required:    true,
			},
		},
	}
}

func (d *routingWrapupcodeFrameworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerMeta, ok := req.ProviderData.(*provider.ProviderMeta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *provider.ProviderMeta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.clientConfig = providerMeta.ClientConfig
}

func (d *routingWrapupcodeFrameworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config routingWrapupcodeFrameworkDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := getRoutingWrapupcodeProxy(d.clientConfig)
	name := config.Name.ValueString()

	log.Printf("Reading routing wrapupcode data source for name: %s", name)

	// Use retry logic for eventual consistency (15-second timeout as per current implementation)
	var wrapupcodeId string
	retryErr := retry.RetryContext(ctx, 15*time.Second, func() *retry.RetryError {
		id, retryable, _, err := proxy.getRoutingWrapupcodeIdByName(ctx, name)
		if err != nil {
			if !retryable {
				return retry.NonRetryableError(fmt.Errorf("failed to find routing wrapupcode with name '%s': %s", name, err))
			}
			log.Printf("Retrying lookup for routing wrapupcode with name '%s': %s", name, err)
			return retry.RetryableError(fmt.Errorf("failed to find routing wrapupcode with name '%s': %s", name, err))
		}
		wrapupcodeId = id
		return nil
	})

	if retryErr != nil {
		resp.Diagnostics.AddError(
			"Error Reading Routing Wrapupcode Data Source",
			fmt.Sprintf("Could not find routing wrapupcode with name '%s': %s", name, retryErr),
		)
		return
	}

	// Set the ID in the model
	config.Id = types.StringValue(wrapupcodeId)

	log.Printf("Found routing wrapupcode %s with ID %s", name, wrapupcodeId)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
