package routing_wrapupcode_v2

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
	"os"
	"strconv"
)

func (r *WrapupCodeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	if r.clientConfig == nil {
		response.Diagnostics.AddError(
			"Uninitialised configuration object",
			"Expected r.clientConfig to not be nil",
		)
		return
	}

	api := platformclientv2.NewRoutingApiWithConfig(r.clientConfig)

	var data WrapupCodeResourceApiModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	wrapupCodeReqBody := buildWrapupCodeRequestBody(data)

	frameworkLog("Creating wrap up code " + strconv.Quote(data.Name.ValueString()))
	respObj, apiResponse, err := api.PostRoutingWrapupcodes(wrapupCodeReqBody)
	if err != nil {
		var apiResponseDetails = "No API response data returned"
		if apiResponse != nil {
			apiResponseDetails = apiResponse.String()
		}
		response.Diagnostics.AddError("failed to create wrap up code", fmt.Sprintf("API Response: %s. Error: %s.", apiResponseDetails, err.Error()))
		return
	}
	frameworkLog(fmt.Sprintf("Successfully created wrap up code '%s'. ID: '%s'", *respObj.Name, *respObj.Id))

	flattenWrapupCodeResponse(&data, *respObj)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *WrapupCodeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	if r.clientConfig == nil {
		response.Diagnostics.AddError(
			"Uninitialised configuration object",
			"Expected r.clientConfig to not be nil",
		)
		return
	}

	api := platformclientv2.NewRoutingApiWithConfig(r.clientConfig)

	var data WrapupCodeResourceApiModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	respBody, resp, err := api.GetRoutingWrapupcode(data.Id.ValueString())
	if err != nil {
		var apiResponseDetails = "No API response data returned"
		if resp != nil {
			apiResponseDetails = resp.String()
		}
		response.Diagnostics.AddError(
			"failed to read wrap up code "+strconv.Quote(data.Id.ValueString()),
			fmt.Sprintf("API Response: %s. Error: %s.", apiResponseDetails, err.Error()),
		)
		return
	}

	flattenWrapupCodeResponse(&data, *respBody)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *WrapupCodeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	if r.clientConfig == nil {
		response.Diagnostics.AddError(
			"Uninitialised configuration object",
			"Expected r.clientConfig to not be nil",
		)
		return
	}

	api := platformclientv2.NewRoutingApiWithConfig(r.clientConfig)

	var data WrapupCodeResourceApiModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	reqBody := buildWrapupCodeRequestBody(data)

	frameworkLog("Updating wrap up code " + strconv.Quote(data.Name.ValueString()))
	respObj, resp, err := api.PutRoutingWrapupcode(data.Id.ValueString(), reqBody)
	if err != nil {
		var apiResponseDetails = "No API response data returned"
		if resp != nil {
			apiResponseDetails = resp.String()
		}
		response.Diagnostics.AddError("failed to update wrap up code", fmt.Sprintf("API Response: %s. Error: %s.", apiResponseDetails, err.Error()))
		return
	}

	frameworkLog("Successfully updated wrap up code " + strconv.Quote(data.Name.ValueString()))
	flattenWrapupCodeResponse(&data, *respObj)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *WrapupCodeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	if r.clientConfig == nil {
		response.Diagnostics.AddError(
			"Uninitialised configuration object",
			"Expected r.clientConfig to not be nil",
		)
		return
	}

	api := platformclientv2.NewRoutingApiWithConfig(r.clientConfig)

	var data WrapupCodeResourceApiModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	frameworkLog("Deleting wrap up code " + strconv.Quote(data.Name.ValueString()))
	apiResponse, err := api.DeleteRoutingWrapupcode(data.Id.ValueString())
	if err != nil {
		apiResponseDetails := "No API response details returned."
		if apiResponse != nil {
			apiResponseDetails = apiResponse.String()
		}
		response.Diagnostics.AddError("Failed to delete wrap up code", fmt.Sprintf("API Response: %s. Error: %s", apiResponseDetails, err.Error()))
		return
	}

	frameworkLog("Successfully deleted wrap up code " + strconv.Quote(data.Name.ValueString()))
}

func (r *WrapupCodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	// TODO: Figure out cyclic issue between this package and 'provider' package

	//
	//client, ok := req.ProviderData.(provider.GenesysCloudProvider)
	//if !ok {
	//	resp.Diagnostics.AddError(
	//		"Unexpected Resource Configure Type",
	//		fmt.Sprintf("Expected *GenesysCloudProvider, got: %T. Please report this issue to the provider developers.", req.ProviderData),
	//	)
	//	return
	//}
	//
	//clientConfig, err := provider.AcquireSdkClient(ctx, client)
	//if err != nil {
	//	resp.Diagnostics.AddError("Failed to acquire sdk client from pool", err.Error())
	//	return
	//}

	clientConfig := platformclientv2.GetDefaultConfiguration()

	err := clientConfig.AuthorizeClientCredentials(os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"), os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET"))
	if err != nil {
		return
	}

	r.clientConfig = clientConfig
}

func NewWrapupCodeResource() resource.Resource {
	return &WrapupCodeResource{}
}
