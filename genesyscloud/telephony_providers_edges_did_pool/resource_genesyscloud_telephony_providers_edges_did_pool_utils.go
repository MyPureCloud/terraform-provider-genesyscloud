package telephony_providers_edges_did_pool

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

type DidPoolStruct struct {
	ResourceLabel    string
	StartPhoneNumber string
	EndPhoneNumber   string
	Description      string
	Comments         string
	PoolProvider     string
}

// DeleteDidPoolWithStartAndEndNumber deletes a did pool by start and end number. Used as a cleanup function in tests which
// utilise the did pool resource
func DeleteDidPoolWithStartAndEndNumber(ctx context.Context, startNumber, endNumber string, sdkConfig *platformclientv2.Configuration) (*platformclientv2.APIResponse, error) {
	proxy := getTelephonyDidPoolProxy(sdkConfig)

	didPoolId, _, resp, err := proxy.getTelephonyDidPoolIdByStartAndEndNumber(ctx, startNumber, endNumber)
	if err != nil {
		return resp, err
	}

	return proxy.deleteTelephonyDidPool(ctx, didPoolId)
}

// GenerateDidPoolResource generates a string representation of a did pool resource based on a DidPoolStruct object
func GenerateDidPoolResource(didPool *DidPoolStruct) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		start_phone_number = "%s"
		end_phone_number   = "%s"
		description        = %s
		comments           = %s
		pool_provider      = %s
	}
	`, ResourceType,
		didPool.ResourceLabel,
		didPool.StartPhoneNumber,
		didPool.EndPhoneNumber,
		didPool.Description,
		didPool.Comments,
		didPool.PoolProvider)
}
