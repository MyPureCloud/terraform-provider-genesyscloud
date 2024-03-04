package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestUnitProviderTokenRefresh(t *testing.T) {
	t.Parallel()

	tExpiresIn := 30

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
		"token_pool_size":    1,
	}
	d := schema.TestResourceDataRaw(t, providerSchema, providerDataMap)

	clientPool := &SDKClientPool{
		Pool: make(chan *platformclientv2.Configuration, 1),
	}
	diag := clientPool.preFill(d, "0.1.0")
	assert.Equal(t, false, diag.HasError())

	// Test the timer functionality
	timerFinished := clientPool.startTimer(d)
	assert.Equal(t, timerFinished, true)
}

func randString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
