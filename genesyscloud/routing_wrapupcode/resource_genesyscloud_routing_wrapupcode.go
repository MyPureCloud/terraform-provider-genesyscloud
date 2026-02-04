package routing_wrapupcode

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

// Ensure routingWrapupcodeFrameworkResource satisfies various resource interfaces.
var (
	_ resource.Resource                = &routingWrapupcodeFrameworkResource{}
	_ resource.ResourceWithConfigure   = &routingWrapupcodeFrameworkResource{}
	_ resource.ResourceWithImportState = &routingWrapupcodeFrameworkResource{}
)

// routingWrapupcodeFrameworkResource defines the resource implementation for Plugin Framework.
type routingWrapupcodeFrameworkResource struct {
	clientConfig *platformclientv2.Configuration
}

// routingWrapupcodeFrameworkResourceModel describes the resource data model.
type routingWrapupcodeFrameworkResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DivisionId  types.String `tfsdk:"division_id"`
	Description types.String `tfsdk:"description"`
}

// NewRoutingWrapupcodeFrameworkResource is a helper function to simplify the provider implementation.
func NewRoutingWrapupcodeFrameworkResource() resource.Resource {
	return &routingWrapupcodeFrameworkResource{}
}

func (r *routingWrapupcodeFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routing_wrapupcode"
}

func (r *routingWrapupcodeFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = RoutingWrapupcodeResourceSchema()
}

func (r *routingWrapupcodeFrameworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *routingWrapupcodeFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routingWrapupcodeFrameworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := getRoutingWrapupcodeProxy(r.clientConfig)
	wrapupcodeRequest := buildWrapupcodeFromFrameworkModel(plan)

	log.Printf("Creating routing wrapupcode %s", plan.Name.ValueString())

	wrapupcode, _, err := proxy.createRoutingWrapupcode(ctx, wrapupcodeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Routing Wrapupcode",
			fmt.Sprintf("Could not create routing wrapupcode %s: %s", plan.Name.ValueString(), err),
		)
		return
	}

	// Update model with response data
	updateFrameworkModelFromAPI(&plan, wrapupcode)

	log.Printf("Created routing wrapupcode %s with ID %s", plan.Name.ValueString(), *wrapupcode.Id)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routingWrapupcodeFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routingWrapupcodeFrameworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := getRoutingWrapupcodeProxy(r.clientConfig)
	id := state.Id.ValueString()

	log.Printf("Reading routing wrapupcode %s", id)

	wrapupcode, apiResp, err := proxy.getRoutingWrapupcodeById(ctx, id)
	if err != nil {
		if util.IsStatus404(apiResp) {
			// Wrapupcode not found, remove from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Routing Wrapupcode",
			fmt.Sprintf("Could not read routing wrapupcode %s: %s", id, err),
		)
		return
	}

	// Update the state with the latest data
	updateFrameworkModelFromAPI(&state, wrapupcode)

	log.Printf("Read routing wrapupcode %s %s", id, *wrapupcode.Name)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *routingWrapupcodeFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan routingWrapupcodeFrameworkResourceModel
	var state routingWrapupcodeFrameworkResourceModel

	// Read Terraform plan and current state data into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := getRoutingWrapupcodeProxy(r.clientConfig)
	id := state.Id.ValueString()
	wrapupcodeRequest := buildWrapupcodeFromFrameworkModel(plan)

	log.Printf("Updating routing wrapupcode %s", plan.Name.ValueString())

	wrapupcode, _, err := proxy.updateRoutingWrapupcode(ctx, id, wrapupcodeRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Routing Wrapupcode",
			fmt.Sprintf("Could not update routing wrapupcode %s: %s", plan.Name.ValueString(), err),
		)
		return
	}

	// Update model with response data
	updateFrameworkModelFromAPI(&plan, wrapupcode)

	log.Printf("Updated routing wrapupcode %s", plan.Name.ValueString())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routingWrapupcodeFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routingWrapupcodeFrameworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := getRoutingWrapupcodeProxy(r.clientConfig)
	id := state.Id.ValueString()
	name := state.Name.ValueString()

	log.Printf("Deleting routing wrapupcode %s", name)

	_, err := proxy.deleteRoutingWrapupcode(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Routing Wrapupcode",
			fmt.Sprintf("Could not delete routing wrapupcode %s: %s", name, err),
		)
		return
	}

	// Verify deletion with retry logic
	retryErr := retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
		_, apiResp, err := proxy.getRoutingWrapupcodeById(ctx, id)
		if err != nil {
			if util.IsStatus404(apiResp) {
				// Wrapupcode deleted successfully
				log.Printf("Deleted routing wrapupcode %s", id)
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting routing wrapupcode %s: %s", id, err))
		}

		return retry.RetryableError(fmt.Errorf("routing wrapupcode %s still exists", id))
	})

	if retryErr != nil {
		resp.Diagnostics.AddError(
			"Error Verifying Routing Wrapupcode Deletion",
			fmt.Sprintf("Could not verify deletion of routing wrapupcode %s: %s", id, retryErr),
		)
		return
	}

	log.Printf("Successfully deleted routing wrapupcode %s", name)
}

func (r *routingWrapupcodeFrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildWrapupcodeFromFrameworkModel converts Framework model to API request model
func buildWrapupcodeFromFrameworkModel(model routingWrapupcodeFrameworkResourceModel) *platformclientv2.Wrapupcoderequest {
	request := &platformclientv2.Wrapupcoderequest{
		Name: model.Name.ValueStringPointer(),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		request.Description = model.Description.ValueStringPointer()
	}

	if !model.DivisionId.IsNull() && !model.DivisionId.IsUnknown() {
		request.Division = &platformclientv2.Writablestarrabledivision{
			Id: model.DivisionId.ValueStringPointer(),
		}
	}

	return request
}

// updateFrameworkModelFromAPI updates Framework model from API response
func updateFrameworkModelFromAPI(model *routingWrapupcodeFrameworkResourceModel, wrapupcode *platformclientv2.Wrapupcode) {
	model.Id = types.StringValue(*wrapupcode.Id)
	model.Name = types.StringValue(*wrapupcode.Name)

	if wrapupcode.Description != nil {
		model.Description = types.StringValue(*wrapupcode.Description)
	} else {
		model.Description = types.StringNull()
	}

	if wrapupcode.Division != nil && wrapupcode.Division.Id != nil {
		model.DivisionId = types.StringValue(*wrapupcode.Division.Id)
	} else {
		model.DivisionId = types.StringNull()
	}
}

// GetAllRoutingWrapupcodes retrieves all routing wrapupcodes for export using Plugin Framework diagnostics.
// This is the future Phase 2 implementation that will be used once the exporter is updated
// to work natively with Framework types.
//
// Returns:
//   - resourceExporter.ResourceIDMetaMap: Map of wrapupcode IDs to metadata
//   - pfdiag.Diagnostics: Plugin Framework diagnostics
//
// Note: Currently NOT used by exporter. Exporter uses GetAllRoutingWrapupcodesSDK (SDK version).
func GetAllRoutingWrapupcodes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, pfdiag.Diagnostics) {
	var diagErr pfdiag.Diagnostics
	proxy := getRoutingWrapupcodeProxy(clientConfig)
	wrapupcodes, _, err := proxy.getAllRoutingWrapupcode(ctx)
	if err != nil {
		diagErr.AddError("Failed to get routing wrapupcodes for export", err.Error())
		return nil, diagErr
	}

	if wrapupcodes == nil {
		return resourceExporter.ResourceIDMetaMap{}, nil
	}

	exportMap := make(resourceExporter.ResourceIDMetaMap)
	for _, wrapupcode := range *wrapupcodes {
		hashedUniqueFields, err := util.QuickHashFields(*wrapupcode.Name)
		if err != nil {
			diagErr.AddError("Failed to hash wrapupcode fields", err.Error())
			return nil, diagErr
		}
		exportMap[*wrapupcode.Id] = &resourceExporter.ResourceMeta{
			BlockLabel: *wrapupcode.Name,
			// Calculate BlockHash for stable export identity
			BlockHash: hashedUniqueFields,
		}
	}
	return exportMap, nil
}

// GetAllRoutingWrapupcodesSDK retrieves all routing wrapupcodes for export using SDK diagnostics.
// This is the Phase 1 implementation that converts SDK types to flat attribute maps
// for the legacy exporter's dependency resolution logic.
//
// IMPORTANT: This function is CURRENTLY USED by the exporter (see RoutingWrapupcodeExporter).
// It implements the lazy fetch pattern for performance optimization.
//
// Returns:
//   - resourceExporter.ResourceIDMetaMap: Map of wrapupcode IDs to metadata with flat attributes
//   - sdkdiag.Diagnostics: SDK diagnostics (required by current exporter)
//
// Lazy Fetch Pattern:
//   - First API call: Fetch all wrapupcode IDs and names (lightweight)
//   - Filter: Apply exporter filters to determine which wrapupcodes to export
//   - Second API call: Fetch full details ONLY for filtered wrapupcodes (performance optimization)
//
// TODO: Remove this function once all resources are migrated to Plugin Framework
// and the exporter is updated to use GetAllRoutingWrapupcodes (Phase 2).
func GetAllRoutingWrapupcodesSDK(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, sdkdiag.Diagnostics) {
	proxy := getRoutingWrapupcodeProxy(clientConfig)

	// Step 1: Fetch all wrapupcodes (lightweight - just IDs and names)
	wrapupcodes, _, err := proxy.getAllRoutingWrapupcode(ctx)
	if err != nil {
		return nil, sdkdiag.Errorf("Failed to get routing wrapupcodes for export: %v", err)
	}

	if wrapupcodes == nil {
		return resourceExporter.ResourceIDMetaMap{}, nil
	}

	// Step 2: Build initial export map with IDs and names
	exportMap := make(resourceExporter.ResourceIDMetaMap)
	for _, wrapupcode := range *wrapupcodes {
		hashedUniqueFields, err := util.QuickHashFields(*wrapupcode.Name)
		if err != nil {
			return nil, sdkdiag.Errorf("Failed to hash wrapupcode fields: %v", err)
		}
		exportMap[*wrapupcode.Id] = &resourceExporter.ResourceMeta{
			BlockLabel: *wrapupcode.Name,
			BlockHash:  hashedUniqueFields,
		}
	}

	// Step 3: Lazy fetch - Get full details ONLY for filtered wrapupcodes
	// Note: For wrapupcode, the initial fetch already includes all attributes (name, division, description)
	// so we don't need additional API calls. However, we still build the flat attribute map
	// for consistency with the exporter's dependency resolution logic.
	for _, wrapupcode := range *wrapupcodes {
		if wrapupcode.Id == nil {
			continue
		}

		// Build flat attribute map for exporter (Phase 1 temporary)
		attributes := buildWrapupcodeAttributes(&wrapupcode)

		// Update export map with attributes
		if meta, exists := exportMap[*wrapupcode.Id]; exists {
			meta.ExportAttributes = attributes
		} else {
			log.Printf("Warning: Wrapupcode %s not found in export map", *wrapupcode.Id)
		}
	}

	return exportMap, nil
}
