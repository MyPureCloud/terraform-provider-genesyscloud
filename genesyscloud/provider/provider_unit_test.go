package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestUnitProviderTokenRefresh(t *testing.T) {
	t.Parallel()

	tExpiresIn := 15
	tPoolSize := 5

	testProxy := &providerProxy{}
	testProxy.authorizeConfigAttr = func(p *providerProxy, config *platformclientv2.Configuration, oauthClientId string, oauthClientSecret string) error {
		config.AccessToken = randString(40)
		config.AccessTokenExpiresIn = tExpiresIn

		return nil
	}

	internalProxy = testProxy
	defer func() { internalProxy = nil }()

	providerSchema := New("0.1.0", make(map[string]*schema.Resource), make(map[string]*schema.Resource))().Schema
	providerDataMap := map[string]interface{}{
		"oauthclient_id":     "1234",
		"oauthclient_secret": "1234",
		"token_pool_size":    tPoolSize,
	}
	d := schema.TestResourceDataRaw(t, providerSchema, providerDataMap)

	clientPool := &SDKClientPool{
		Pool: make(chan *platformclientv2.Configuration, tPoolSize),
	}
	diag := clientPool.preFill(d, "0.1.0") // Initialize the config pool
	assert.Equal(t, false, diag.HasError())

	var originalAccessTokens []string
	for config := range clientPool.Pool {
		originalAccessTokens = append(originalAccessTokens, config.AccessToken)
	}

	// After the configs has been created we sleep until the accessTokens expire and check that they have been refreshed
	time.Sleep(time.Duration(tExpiresIn + 15)) // Add some time to give the tokens a chance to refresh

	var refreshedAccessTokens []string
	for config := range clientPool.Pool {
		refreshedAccessTokens = append(refreshedAccessTokens, config.AccessToken)
	}

	assert.Equal(t, false, areEqual(originalAccessTokens, refreshedAccessTokens), "Original tokens not updated")
}

func areEqual(tokens1, tokens2 []string) bool {
	// Check if lengths are equal
	if len(tokens1) != len(tokens2) {
		return false
	}

	// Sort both groups
	sort.Strings(tokens1)
	sort.Strings(tokens2)

	// Compare each element
	for i := range tokens1 {
		if tokens1[i] != tokens2[i] {
			return false
		}
	}

	return true
}

func randString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
