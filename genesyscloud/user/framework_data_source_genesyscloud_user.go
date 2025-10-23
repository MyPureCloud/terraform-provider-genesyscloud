package user

import (
	"context"
	"fmt"
	"log"
	"net/mail"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

// Ensure UserFrameworkDataSource satisfies various data source interfaces
var (
	_ datasource.DataSource              = &UserFrameworkDataSource{}
	_ datasource.DataSourceWithConfigure = &UserFrameworkDataSource{}
)

// UserFrameworkDataSource defines the data source implementation for Genesys Cloud User
type UserFrameworkDataSource struct {
	clientConfig *platformclientv2.Configuration
}

// UserFrameworkDataSourceModel describes the data source data model
type UserFrameworkDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	Name       types.String `tfsdk:"name"`
	State      types.String `tfsdk:"state"`
	DivisionId types.String `tfsdk:"division_id"`
}

// NewUserFrameworkDataSource creates a new instance of the UserFrameworkDataSource
func NewUserFrameworkDataSource() datasource.DataSource {
	return &UserFrameworkDataSource{}
}

// Metadata returns the data source type name
func (d *UserFrameworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the data source
func (d *UserFrameworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source for Genesys Cloud Users. Select a user by email or name. If both email & name are specified, the name won't be used for user lookup",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the user.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "User email.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "User name.",
				Optional:    true,
			},
			"state": schema.StringAttribute{
				Description: "The user's state (active, inactive).",
				Computed:    true,
			},
			"division_id": schema.StringAttribute{
				Description: "The division ID of the user.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *UserFrameworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured
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

// Read refreshes the Terraform state with the latest data
func (d *UserFrameworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config UserFrameworkDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one search field is provided
	if config.Email.IsNull() && config.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Search Field",
			"Either 'email' or 'name' must be specified to search for a user",
		)
		return
	}

	// Determine search key - email takes precedence over name
	var searchKey string
	if !config.Email.IsNull() {
		searchKey = config.Email.ValueString()
	} else {
		searchKey = config.Name.ValueString()
	}

	log.Printf("Searching for user with key: %s", searchKey)

	// Get user ID using the existing proxy method with retry logic
	proxy := GetUserProxy(d.clientConfig)
	userId, diags := d.getUserIdByNameWithRetry(ctx, proxy, searchKey)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set the ID in the state
	config.Id = types.StringValue(userId)

	// Fetch full user details to populate name and email
	user, _, err := proxy.getUserById(ctx, userId, nil, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading User",
			fmt.Sprintf("Could not read user %s: %s", userId, err),
		)
		return
	}

	// Set user details in the state
	if user.Name != nil {
		config.Name = types.StringValue(*user.Name)
	} else {
		config.Name = types.StringNull()
	}
	if user.Email != nil {
		config.Email = types.StringValue(*user.Email)
	} else {
		config.Email = types.StringNull()
	}
	if user.State != nil {
		config.State = types.StringValue(*user.State)
	} else {
		config.State = types.StringNull()
	}
	if user.Division != nil && user.Division.Id != nil {
		config.DivisionId = types.StringValue(*user.Division.Id)
	} else {
		config.DivisionId = types.StringNull()
	}

	log.Printf("Found user with ID: %s, Name: %s, Email: %s, State: %s, DivisionId: %s", userId,
		config.Name.ValueString(), config.Email.ValueString(), config.State.ValueString(), config.DivisionId.ValueString())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// getUserIdByNameWithRetry implements retry logic for user lookup with eventual consistency
func (d *UserFrameworkDataSource) getUserIdByNameWithRetry(ctx context.Context, proxy *userProxy, searchField string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var userId string

	exactSearchType := "EXACT"
	sortOrderAsc := "ASC"
	emailField := "email"

	searchCriteria := platformclientv2.Usersearchcriteria{
		VarType: &exactSearchType,
	}

	// Determine if search field is email or name
	searchFieldValue, searchFieldType := d.emailorNameDisambiguation(searchField)
	searchCriteria.Fields = &[]string{searchFieldType}
	searchCriteria.Value = &searchFieldValue

	err := retry.RetryContext(ctx, 15*time.Second, func() *retry.RetryError {
		users, _, getErr := proxy.getUserByName(ctx, platformclientv2.Usersearchrequest{
			SortBy:    &emailField,
			SortOrder: &sortOrderAsc,
			Query:     &[]platformclientv2.Usersearchcriteria{searchCriteria},
		})

		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("error requesting users: %s", getErr))
		}

		if users.Results == nil || len(*users.Results) == 0 {
			return retry.RetryableError(fmt.Errorf("no users found with search criteria %v", searchCriteria))
		}

		// Select first user in the list
		userId = *(*users.Results)[0].Id
		return nil
	})

	if err != nil {
		diags.AddError(
			"Failed to Find User",
			fmt.Sprintf("Failed to find user with search field '%s': %s", searchField, err),
		)
		return "", diags
	}

	return userId, diags
}

// emailorNameDisambiguation determines if the search field is an email or name
func (d *UserFrameworkDataSource) emailorNameDisambiguation(searchField string) (string, string) {
	emailField := "email"
	nameField := "name"
	_, err := mail.ParseAddress(searchField)
	if err == nil {
		return searchField, emailField
	}
	return searchField, nameField
}
