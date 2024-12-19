package telephony_providers_edges_extension_pool

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

type ExtensionPoolStruct struct {
	ResourceLabel string
	StartNumber   string
	EndNumber     string
	Description   string
}

func GenerateExtensionPoolResource(extensionPool *ExtensionPoolStruct) string {
	return fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_extension_pool" "%s" {
		start_number = "%s"
		end_number = "%s"
		description = %s
	}
	`, extensionPool.ResourceLabel,
		extensionPool.StartNumber,
		extensionPool.EndNumber,
		extensionPool.Description)
}

func DeleteExtensionPoolWithNumber(startNumber string) error {
	sdkConfig, err := provider.AuthorizeSdk()
	if err != nil {
		return err
	}

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		extensionPools, _, getErr := edgesAPI.GetTelephonyProvidersEdgesExtensionpools(100, pageNum, "", "")
		if getErr != nil {
			return getErr
		}

		if extensionPools.Entities == nil || len(*extensionPools.Entities) == 0 {
			break
		}

		for _, extensionPool := range *extensionPools.Entities {
			if extensionPool.StartNumber != nil && *extensionPool.StartNumber == startNumber {
				_, err := edgesAPI.DeleteTelephonyProvidersEdgesExtensionpool(*extensionPool.Id)
				time.Sleep(20 * time.Second)
				return err
			}
		}
	}

	return nil
}
