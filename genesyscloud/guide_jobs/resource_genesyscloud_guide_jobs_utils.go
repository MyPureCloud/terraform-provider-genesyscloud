package guide_jobs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func buildGuideJobFromResourceData(d *schema.ResourceData) GenerateGuideContentRequest {
	guideJobReq := GenerateGuideContentRequest{}

	// Set Optional Attributes if they are not null

	description := d.Get("description").(string)
	if description != "" {
		guideJobReq.Description = &description
	}

	url := d.Get("url").(string)
	if url != "" {
		guideJobReq.Url = &url
	}

	return guideJobReq
}

func flattenAddressableEntityRefs(addressableEntityRefs *[]platformclientv2.Addressableentityref) *schema.Set {
	addressableEntityRefList := make([]interface{}, len(*addressableEntityRefs))
	for i, v := range *addressableEntityRefs {
		addressableEntityRefList[i] = *v.Id
	}
	return schema.NewSet(schema.HashString, addressableEntityRefList)
}
