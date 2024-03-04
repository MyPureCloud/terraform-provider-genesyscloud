package provider

import (
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
)

var internalProxy *providerProxy

type authorizeConfigFunc func(p *providerProxy, config *platformclientv2.Configuration, oauthClientId string, oauthClientSecret string) error

type providerProxy struct {
	authorizeConfigAttr authorizeConfigFunc
}

func newProviderProxy() *providerProxy {
	return &providerProxy{
		authorizeConfigAttr: authorizeConfigFn,
	}
}

func getProviderProxy() *providerProxy {
	if internalProxy == nil {
		internalProxy = newProviderProxy()
	}

	return internalProxy
}

func (p *providerProxy) authorizeConfig(config *platformclientv2.Configuration, oauthClientId string, oauthClientSecret string) error {
	return p.authorizeConfigAttr(p, config, oauthClientId, oauthClientSecret)
}

func authorizeConfigFn(p *providerProxy, config *platformclientv2.Configuration, oauthClientId string, oauthClientSecret string) error {
	err := config.AuthorizeClientCredentials(oauthClientId, oauthClientSecret)
	if err != nil {
		return fmt.Errorf("failed to authorize Genesys Cloud client credentials: %v", err)
	}

	return nil
}
