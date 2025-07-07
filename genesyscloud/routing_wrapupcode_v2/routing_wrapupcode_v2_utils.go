package routing_wrapupcode_v2

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
	"log"
)

func buildWrapupCodeRequestBody(data WrapupCodeResourceApiModel) platformclientv2.Wrapupcoderequest {
	wrapupCodeReqBody := platformclientv2.Wrapupcoderequest{
		Name:        data.Name.ValueStringPointer(),
		Description: data.Description.ValueStringPointer(),
	}
	if !data.DivisionId.IsNull() && !data.DivisionId.IsUnknown() && data.DivisionId.ValueString() != "" {
		wrapupCodeReqBody.Division = &platformclientv2.Writablestarrabledivision{Id: data.DivisionId.ValueStringPointer()}
	}
	return wrapupCodeReqBody
}

func flattenWrapupCodeResponse(data *WrapupCodeResourceApiModel, respBody platformclientv2.Wrapupcode) {
	data.Id = types.StringValue(*respBody.Id)
	data.Name = types.StringValue(*respBody.Name)
	if respBody.Description != nil {
		data.Description = types.StringValue(*respBody.Description)
	}
	if respBody.Division != nil && respBody.Division.Id != nil {
		data.DivisionId = types.StringValue(*respBody.Division.Id)
	}
}

func frameworkLog(s string) {
	log.Println("(Framework) ", s)
}
