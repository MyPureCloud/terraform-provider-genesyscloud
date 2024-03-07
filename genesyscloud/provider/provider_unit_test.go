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

	// Get original tokens
	if len(clientPool.Pool) == 0 {
		t.Error("no clients available")
	}
	var originalAccessTokens []string
	for i := 0; i < tPoolSize; i++ {
		config := <-clientPool.Pool
		originalAccessTokens = append(originalAccessTokens, config.AccessToken)
		clientPool.Pool <- config
	}
	sort.Strings(originalAccessTokens)

	runs := 5
	for j := 0; j < runs; j++ {
		// sleep until the accessTokens expire and check that they have been refreshed
		time.Sleep(time.Second * time.Duration(tExpiresIn+5)) // Add some time to give the tokens a chance to refresh

		if len(clientPool.Pool) == 0 {
			t.Error("no clients available")
		}
		var refreshedAccessTokens []string
		for i := 0; i < tPoolSize; i++ {
			config := <-clientPool.Pool
			refreshedAccessTokens = append(refreshedAccessTokens, config.AccessToken)
			clientPool.Pool <- config
		}
		sort.Strings(refreshedAccessTokens)

		// Check each token
		for i := 0; i < tPoolSize; i++ {
			assert.Equal(t, true, originalAccessTokens[i] != refreshedAccessTokens[i], "Original token not updated")
		}

		originalAccessTokens = refreshedAccessTokens
	}
}

func randString(n int) string {
	const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, n)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}
