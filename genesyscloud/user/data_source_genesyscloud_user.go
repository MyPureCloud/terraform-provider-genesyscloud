package user

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Ensure UserFrameworkDataSource satisfies various data source interfaces
var (
	_                   datasource.DataSource              = &UserFrameworkDataSource{}
	_                   datasource.DataSourceWithConfigure = &UserFrameworkDataSource{}
	dataSourceUserCache *rc.DataSourceCache
)

// UserFrameworkDataSource implements the Terraform Plugin Framework data source for Genesys Cloud Users.
//
// Lifecycle:
// 1. NewUserFrameworkDataSource() - Creates the data source instance (called by provider during registration)
// 2. Metadata() - Registers the data source name "genesyscloud_user" (called once at provider startup)
// 3. Schema() - Defines the data source structure and attributes (called once at provider startup)
// 4. Configure() - Receives API client configuration from provider (called once at provider startup)
// 5. Read() - Fetches user data from Genesys Cloud API (called every time Terraform needs the data)
//
// Usage in Terraform:
//
//	data "genesyscloud_user" "example" {
//	  email = "user@example.com"  # or name = "John Doe"
//	}
//
// The data source uses a cache (dataSourceUserCache) to improve performance by reducing API calls.
type UserFrameworkDataSource struct {
	clientConfig *platformclientv2.Configuration
}

// UserFrameworkDataSourceModel describes the data source data model
// The SDKv2 implementation only has the following fields configured in the source code
type UserFrameworkDataSourceModel struct {
	Id    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
	Name  types.String `tfsdk:"name"`
}

// NewUserFrameworkDataSource creates a new instance of the UserFrameworkDataSource.
// Called by: Provider during registration
// When: Once when the provider is initialized
func NewUserFrameworkDataSource() datasource.DataSource {
	return &UserFrameworkDataSource{}
}

// Metadata sets the data source type name to "genesyscloud_user".
// Called by: Terraform Plugin Framework
// When: Once during provider initialization (step 1 of 4)
// Purpose: Registers the data source name used in Terraform configs
func (d *UserFrameworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the data source structure (input/output attributes).
// Called by: Terraform Plugin Framework
// When: Once during provider initialization (step 2 of 4), after Metadata()
// Purpose: Defines available fields in: data "genesyscloud_user" "example" { ... }
func (d *UserFrameworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = UserDataSourceSchema()
}

// Configure receives the provider's API client configuration.
// Called by: Terraform Plugin Framework
// When: Once during provider initialization (step 3 of 4), after Schema()
// Purpose: Receives API credentials and client config to enable API calls
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

// Read refreshes the Terraform state with the latest data.
// Called by: Terraform Plugin Framework
// When: Every time Terraform needs to read the data source (step 4 of 4)
//   - During `terraform plan`
//   - During `terraform apply`
//   - During `terraform refresh`
//
// Purpose: Fetches user data from Genesys Cloud API and stores it in Terraform state
func (d *UserFrameworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config UserFrameworkDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one search field is provided
	if config.Email.IsNull() && config.Name.IsNull() {
		resp.Diagnostics.Append(
			util.BuildFrameworkDiagnosticError(
				ResourceType,
				"no user search field specified",
				nil,
			)...,
		)
		return
	}

	// Determine search key
	var searchKey string
	if !config.Email.IsNull() {
		searchKey = config.Email.ValueString()
	}
	if !config.Name.IsNull() {
		searchKey = config.Name.ValueString()
	}

	log.Printf("Searching for user with key: %s", searchKey)

	// Initialize cache if not already initialized
	if dataSourceUserCache == nil {
		dataSourceUserCache = rc.NewDataSourceCache(d.clientConfig, hydrateUserCache, getUserByName)
	}

	// Retrieve user ID from cache or API
	userId, sdkDiags := rc.RetrieveId(dataSourceUserCache, ResourceType, searchKey, ctx)
	if sdkDiags.HasError() {
		frameworkDiags := util.ConvertSDKDiagnosticsToFramework(sdkDiags)
		resp.Diagnostics.Append(frameworkDiags...)
		return
	}

	// Set the ID in the state
	config.Id = types.StringValue(userId)

	// Fetch the full user details to populate name and email
	proxy := GetUserProxy(d.clientConfig)
	user, response, err := proxy.getUserById(ctx, userId, []string{}, "")
	if err != nil {
		resp.Diagnostics.Append(
			util.BuildFrameworkAPIDiagnosticError(
				ResourceType,
				fmt.Sprintf("Failed to retrieve user details for ID %s: %s", userId, err),
				response,
			)...,
		)
		return
	}

	// Populate name and email from the API response
	if user.Name != nil {
		config.Name = types.StringValue(*user.Name)
	}
	if user.Email != nil {
		config.Email = types.StringValue(*user.Email)
	}

	log.Printf("Found user with ID: %s, Name: %s, Email: %s", userId, config.Name.ValueString(), config.Email.ValueString())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// getUserByName retrieves a user ID by searching for a user by name or email.
// Note: Returns SDKv2 diag.Diagnostics for compatibility with resource_cache infrastructure.
// The cache expects: func(*DataSourceCache, string, context.Context) (string, diag.Diagnostics)
// This will be updated when the cache is migrated to Plugin Framework.
func getUserByName(c *rc.DataSourceCache, searchField string, ctx context.Context) (string, sdkdiag.Diagnostics) {
	log.Printf("getUserByName for data source %s", ResourceType)
	proxy := GetUserProxy(c.ClientConfig)
	userId := ""
	exactSearchType := "EXACT"
	sortOrderAsc := "ASC"
	emailField := "email"

	searchCriteria := platformclientv2.Usersearchcriteria{
		VarType: &exactSearchType,
	}
	searchFieldValue, searchFieldType := emailorNameDisambiguation(searchField)
	searchCriteria.Fields = &[]string{searchFieldType}
	searchCriteria.Value = &searchFieldValue

	sdkDiags := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		users, resp, getErr := proxy.getUserByName(ctx, platformclientv2.Usersearchrequest{
			SortBy:    &emailField,
			SortOrder: &sortOrderAsc,
			Query:     &[]platformclientv2.Usersearchcriteria{searchCriteria},
		})
		if getErr != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error requesting users: %s", getErr), resp))
		}

		if users.Results == nil || len(*users.Results) == 0 {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No users found with search criteria %v", searchCriteria), resp))
		}

		// Select first user in the list
		userId = *(*users.Results)[0].Id
		return nil
	})

	log.Printf("getUserByName completed for data source %s", ResourceType)
	return userId, sdkDiags
}

func hydrateUserCache(c *rc.DataSourceCache, ctx context.Context) error {
	log.Printf("hydrating cache for data source %s", ResourceType)
	proxy := GetUserProxy(c.ClientConfig)
	const pageSize = 100
	users, response, err := proxy.hydrateUserCache(ctx, pageSize, 1)
	if err != nil {
		return fmt.Errorf("failed to get first page of users: %v %v", err, response)
	}

	if users.Entities == nil || len(*users.Entities) == 0 {
		return nil
	}

	for _, user := range *users.Entities {
		c.Cache[*user.Name] = *user.Id
		c.Cache[*user.Email] = *user.Id
	}

	for pageNum := 2; pageNum <= *users.PageCount; pageNum++ {
		users, response, err := proxy.hydrateUserCache(ctx, pageSize, pageNum)

		log.Printf("hydrating cache for data source %s with page number: %v", ResourceType, pageNum)
		if err != nil {
			return fmt.Errorf("failed to get page of users: %v %v", err, response)
		}
		if users.Entities == nil || len(*users.Entities) == 0 {
			break
		}
		// Add ids to cache
		for _, user := range *users.Entities {
			c.Cache[*user.Name] = *user.Id
			c.Cache[*user.Email] = *user.Id
		}
	}
	log.Printf("cache hydration completed for data source %s", ResourceType)
	return nil
}
