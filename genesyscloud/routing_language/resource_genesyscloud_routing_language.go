package routing_language

import (
	"context"
	"fmt"
	"log"
	"time"

	pfdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Ensure routingLanguageFrameworkResource satisfies various resource interfaces.
var (
	_ resource.Resource                = &routingLanguageFrameworkResource{}
	_ resource.ResourceWithConfigure   = &routingLanguageFrameworkResource{}
	_ resource.ResourceWithImportState = &routingLanguageFrameworkResource{}
)

// routingLanguageFrameworkResource defines the resource implementation for Plugin Framework.
type routingLanguageFrameworkResource struct {
	clientConfig *platformclientv2.Configuration
}

// routingLanguageFrameworkResourceModel describes the resource data model.
type routingLanguageFrameworkResourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewFrameworkRoutingLanguageResource is a helper function to simplify the provider implementation.
func NewFrameworkRoutingLanguageResource() resource.Resource {
	return &routingLanguageFrameworkResource{}
}

func (r *routingLanguageFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_language"
}

func (r *routingLanguageFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = RoutingLanguageResourceSchema()
}

func (r *routingLanguageFrameworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerMeta, ok := req.ProviderData.(*provider.ProviderMeta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider.ProviderMeta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.clientConfig = providerMeta.ClientConfig
}

func (r *routingLanguageFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routingLanguageFrameworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := getRoutingLanguageProxy(r.clientConfig)
	name := plan.Name.ValueString()

	log.Printf("Creating routing language %s", name)

	language, _, err := proxy.createRoutingLanguage(ctx, &platformclientv2.Language{
		Name: &name,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Routing Language",
			fmt.Sprintf("Could not create routing language %s: %s", name, err),
		)
		return
	}

	// Set the ID and name in the state
	plan.Id = types.StringValue(*language.Id)
	plan.Name = types.StringValue(*language.Name)

	log.Printf("Created routing language %s with ID %s", name, *language.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routingLanguageFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routingLanguageFrameworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := getRoutingLanguageProxy(r.clientConfig)
	id := state.Id.ValueString()

	log.Printf("Reading routing language %s", id)

	language, apiResp, err := proxy.getRoutingLanguageById(ctx, id)
	if err != nil {
		if util.IsStatus404(apiResp) {
			// Language not found, remove from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Routing Language",
			fmt.Sprintf("Could not read routing language %s: %s", id, err),
		)
		return
	}

	// Check if language is deleted
	if language.State != nil && *language.State == "deleted" {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state with the latest data
	state.Id = types.StringValue(*language.Id)
	state.Name = types.StringValue(*language.Name)

	log.Printf("Read routing language %s %s", id, *language.Name)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *routingLanguageFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This resource does not support updates - all changes require replacement
	// This method should never be called due to RequiresReplace plan modifier
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Routing language resources do not support updates. All changes require replacement.",
	)
}

func (r *routingLanguageFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routingLanguageFrameworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := getRoutingLanguageProxy(r.clientConfig)
	id := state.Id.ValueString()
	name := state.Name.ValueString()

	log.Printf("Deleting routing language %s", name)

	_, err := proxy.deleteRoutingLanguage(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Routing Language",
			fmt.Sprintf("Could not delete routing language %s: %s", name, err),
		)
		return
	}

	// Verify deletion with retry logic
	retryErr := retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
		routingLanguage, apiResp, err := proxy.getRoutingLanguageById(ctx, id)
		if err != nil {
			if util.IsStatus404(apiResp) {
				// Language deleted successfully
				log.Printf("Deleted routing language %s", id)
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting routing language %s: %s", id, err))
		}

		if routingLanguage.State != nil && *routingLanguage.State == "deleted" {
			// Language marked as deleted
			log.Printf("Deleted routing language %s", id)
			return nil
		}

		return retry.RetryableError(fmt.Errorf("routing language %s still exists", id))
	})

	if retryErr != nil {
		resp.Diagnostics.AddError(
			"Error Verifying Routing Language Deletion",
			fmt.Sprintf("Could not verify deletion of routing language %s: %s", id, retryErr),
		)
		return
	}

	log.Printf("Successfully deleted routing language %s", name)
}

func (r *routingLanguageFrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// GetAllRoutingLanguages retrieves all routing languages for export using Plugin Framework diagnostics.
// This is the future Phase 2 implementation that will be used once the exporter is updated
// to work natively with Framework types.
//
// Returns:
//   - resourceExporter.ResourceIDMetaMap: Map of language IDs to metadata
//   - pfdiag.Diagnostics: Plugin Framework diagnostics
//
// Note: Currently NOT used by exporter. Exporter uses GetAllRoutingLanguagesSDK (SDK version).
func GetAllRoutingLanguages(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, pfdiag.Diagnostics) {
	var diagErr pfdiag.Diagnostics
	proxy := getRoutingLanguageProxy(clientConfig)
	languages, _, err := proxy.getAllRoutingLanguages(ctx, "")
	if err != nil {
		diagErr.AddError("Failed to get routing languages for export", err.Error())
		return nil, diagErr
	}

	if languages == nil {
		return resourceExporter.ResourceIDMetaMap{}, nil
	}

	exportMap := make(resourceExporter.ResourceIDMetaMap)
	for _, language := range *languages {
		hashedUniqueFields, err := util.QuickHashFields(*language.Name)
		if err != nil {
			diagErr.AddError("Failed to hash language fields", err.Error())
			return nil, diagErr
		}
		exportMap[*language.Id] = &resourceExporter.ResourceMeta{
			BlockLabel: *language.Name,
			// Calculate BlockHash for stable export identity
			BlockHash: hashedUniqueFields,
		}
	}
	return exportMap, nil
}

// GetAllRoutingLanguagesSDK retrieves all routing languages for export using SDK diagnostics.
// This is the Phase 1 implementation that converts SDK types to flat attribute maps
// for the legacy exporter's dependency resolution logic.
//
// IMPORTANT: This function is CURRENTLY USED by the exporter (see RoutingLanguageExporter).
// It implements the lazy fetch pattern for performance optimization.
//
// Returns:
//   - resourceExporter.ResourceIDMetaMap: Map of language IDs to metadata with flat attributes
//   - sdkdiag.Diagnostics: SDK diagnostics (required by current exporter)
//
// Lazy Fetch Pattern:
//   - First API call: Fetch all language IDs and names (lightweight)
//   - Filter: Apply exporter filters to determine which languages to export
//   - Second API call: Fetch full details ONLY for filtered languages (performance optimization)
//
// TODO: Remove this function once all resources are migrated to Plugin Framework
// and the exporter is updated to use GetAllRoutingLanguages (Phase 2).
func GetAllRoutingLanguagesSDK(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, sdkdiag.Diagnostics) {
	proxy := getRoutingLanguageProxy(clientConfig)

	// Step 1: Fetch all languages (lightweight - just IDs and names)
	languages, _, err := proxy.getAllRoutingLanguages(ctx, "")
	if err != nil {
		return nil, sdkdiag.Errorf("Failed to get routing languages for export: %v", err)
	}

	if languages == nil {
		return resourceExporter.ResourceIDMetaMap{}, nil
	}

	// Step 2: Build initial export map with IDs and names
	exportMap := make(resourceExporter.ResourceIDMetaMap)
	for _, language := range *languages {
		hashedUniqueFields, err := util.QuickHashFields(*language.Name)
		if err != nil {
			return nil, sdkdiag.Errorf("Failed to hash language fields: %v", err)
		}
		exportMap[*language.Id] = &resourceExporter.ResourceMeta{
			BlockLabel: *language.Name,
			BlockHash:  hashedUniqueFields,
		}
	}

	// Step 3: Lazy fetch - Get full details ONLY for filtered languages
	// Note: For routing language, the initial fetch already includes all attributes (id, name)
	// so we don't need additional API calls. However, we still build the flat attribute map
	// for consistency with the exporter's dependency resolution logic.
	for _, language := range *languages {
		if language.Id == nil {
			continue
		}

		// Build flat attribute map for exporter (Phase 1 temporary)
		attributes := buildLanguageAttributes(&language)

		// Update export map with attributes
		if meta, exists := exportMap[*language.Id]; exists {
			meta.ExportAttributes = attributes
		} else {
			log.Printf("Warning: Language %s not found in export map", *language.Id)
		}
	}

	return exportMap, nil
}
