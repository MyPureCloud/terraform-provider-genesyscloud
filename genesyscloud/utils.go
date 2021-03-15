package genesyscloud

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

type jsonMap map[string]interface{}

// Attempt to get the home division once during a provider run
var divOnce sync.Once
var homeDivID string
var homeDivErr diag.Diagnostics

func getHomeDivisionID() (string, diag.Diagnostics) {
	divOnce.Do(func() {
		authAPI := platformclientv2.NewAuthorizationApi()
		homeDiv, _, err := authAPI.GetAuthorizationDivisionsHome()
		if err != nil {
			homeDivErr = diag.Errorf("Failed to query home division: %s", err)
			return
		}
		homeDivID = *homeDiv.Id
	})

	if homeDivErr != nil {
		return "", homeDivErr
	}
	return homeDivID, nil
}

func buildSdkDomainEntityRef(d *schema.ResourceData, idAttr string) *platformclientv2.Domainentityref {
	idVal := d.Get(idAttr).(string)
	if idVal == "" {
		return nil
	}
	return &platformclientv2.Domainentityref{Id: &idVal}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// difference returns the elements in a that aren't in b
func sliceDifference(a, b []string) []string {
	var diff []string
	if len(a) == 0 {
		return diff
	}
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func stringListToSet(list []string) *schema.Set {
	interfaceList := make([]interface{}, len(list))
	for i, v := range list {
		interfaceList[i] = v
	}
	return schema.NewSet(schema.HashString, interfaceList)
}

func setToStringList(strSet *schema.Set) *[]string {
	interfaceList := strSet.List()
	strList := make([]string, len(interfaceList))
	for i, s := range interfaceList {
		strList[i] = s.(string)
	}
	return &strList
}

func interfaceListToStrings(interfaceList []interface{}) []string {
	strs := make([]string, len(interfaceList))
	for i, val := range interfaceList {
		strs[i] = val.(string)
	}
	return strs
}

func withRetries(ctx context.Context, timeout time.Duration, method func() *resource.RetryError) diag.Diagnostics {
	return diag.FromErr(resource.RetryContext(ctx, timeout, method))
}

type checkResponseFunc func(resp *platformclientv2.APIResponse) bool
type callSdkFunc func() (*platformclientv2.APIResponse, diag.Diagnostics)

// Retries up to 10 times while the shouldRetry condition returns true
// Useful for adding custom retry logic to normally non-retryable error codes
func retryWhen(shouldRetry checkResponseFunc, callSdk callSdkFunc) diag.Diagnostics {
	var lastErr diag.Diagnostics
	for i := 0; i < 10; i++ {
		resp, lastErr := callSdk()
		if lastErr != nil {
			if resp != nil && shouldRetry(resp) {
				// Wait a second and try again
				time.Sleep(time.Second)
				continue
			} else {
				return lastErr
			}
		}
		// Success
		return nil
	}
	return diag.Errorf("Exhausted retries. Last error: %v", lastErr)
}

func isVersionMismatch(resp *platformclientv2.APIResponse) bool {
	// Version mismatch from directory may be a 409 or 400 with specific error message
	if resp != nil {
		if resp.StatusCode == 409 || (resp.StatusCode == 400 &&
			resp.Error != nil && strings.Contains((*resp.Error).Message, "does not match the current version")) {
			return true
		}
	}
	return false
}
