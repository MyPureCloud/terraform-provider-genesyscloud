package routing_language

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

// Ensure routingLanguageFrameworkDataSource satisfies various datasource interfaces.
var (
	_ datasource.DataSource              = &routingLanguageFrameworkDataSource{}
	_ datasource.DataSourceWithConfigure = &routingLanguageFrameworkDataSource{}
)

// routingLanguageFrameworkDataSource defines the data source implementation for Plugin Framework.
type routingLanguageFrameworkDataSource struct {
	clientConfig *platformclientv2.Configuration
}

// routingLanguageFrameworkDataSourceModel describes the data source data model.
type routingLanguageFrameworkDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewFrameworkRoutingLanguageDataSource is a helper function to simplify the provider implementation.
func NewFrameworkRoutingLanguageDataSource() datasource.DataSource {
	return &routingLanguageFrameworkDataSource{}
}

func (d *routingLanguageFrameworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_language"
}

func (d *routingLanguageFrameworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = RoutingLanguageDataSourceSchema()
}

func (d *routingLanguageFrameworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *routingLanguageFrameworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config routingLanguageFrameworkDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := getRoutingLanguageProxy(d.clientConfig)
	name := config.Name.ValueString()

	log.Printf("Reading routing language data source for name: %s", name)

	// Find language by name with retry logic for eventual consistency
	retryErr := retry.RetryContext(ctx, 15*time.Second, func() *retry.RetryError {
		languageId, _, retryable, err := proxy.getRoutingLanguageIdByName(ctx, name)
		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("error requesting language %s: %s", name, err))
		}
		if retryable {
			return retry.RetryableError(fmt.Errorf("error requesting language %s: %s", name, err))
		}

		// Set the ID and name in the state
		config.Id = types.StringValue(languageId)
		config.Name = types.StringValue(name)

		log.Printf("Found routing language %s with ID %s", name, languageId)
		return nil
	})

	if retryErr != nil {
		resp.Diagnostics.AddError(
			"Error Reading Routing Language Data Source",
			fmt.Sprintf("Could not find routing language with name %s: %s", name, retryErr),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
