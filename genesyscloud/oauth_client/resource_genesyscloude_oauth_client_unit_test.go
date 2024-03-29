package oauth_client

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/provider"

	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

/** This is a unit test to ensure that we populate the internal oAuthCredential cache.  This is to test the fix for DEVTOOLING-448 **/
func TestUnitCreateOAuthClientWithCache(t *testing.T) {
	tName := "my-genesys-oauth-unit-test"
	tDescription := "This is a test Oauth Client Created for my unit Test"
	tAuthorizedGrantType := "CLIENT-CREDENTIALS"
	tAccessTokenValiditySeconds := 86400
	tState := "active"

	tOauthCredId := uuid.NewString()
	tOAuthCredSecret := uuid.NewString()

	oAuthClient := &platformclientv2.Oauthclient{
		Id:                         &tOauthCredId,
		Name:                       &tName,
		Description:                &tDescription,
		Secret:                     &tOAuthCredSecret,
		AuthorizedGrantType:        &tAuthorizedGrantType,
		AccessTokenValiditySeconds: &tAccessTokenValiditySeconds, //Need to have this here because we need to make sure match what is being put into the consistency checker
		State:                      &tState,
	}

	ocproxy := &oauthClientProxy{createdClientCache: make(map[string]platformclientv2.Oauthclient)}

	ocproxy.createOAuthClientAttr = func(ctx context.Context, p *oauthClientProxy, request platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tName, *request.Name)
		assert.Equal(t, tDescription, *request.Description)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return oAuthClient, apiResponse, nil
	}

	ocproxy.getOAuthClientAttr = func(ctx context.Context, p *oauthClientProxy, id string) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return oAuthClient, apiResponse, nil
	}

	internalProxy = ocproxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gc := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceOAuthClient()
	resourceDataMap := map[string]interface{}{
		"name":                  tName,
		"description":           tDescription,
		"authorized_grant_type": tAuthorizedGrantType,
		"state":                 tState,
	}

	d := schema.TestResourceDataRaw(t, resourceSchema.Schema, resourceDataMap)
	diag := createOAuthClient(ctx, d, gc)

	//First lets check the cache
	cachedOAuthClient := ocproxy.GetCachedOAuthClient(tOauthCredId)
	assert.NotNil(t, cachedOAuthClient)
	assert.Equal(t, tOauthCredId, *cachedOAuthClient.Id)
	assert.Equal(t, tOAuthCredSecret, *cachedOAuthClient.Secret)

	// Now lets check the results of the API call
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tOauthCredId, d.Id())
	assert.Equal(t, tName, d.Get("name").(string))
	assert.Equal(t, tDescription, d.Get("description").(string))
	assert.Equal(t, tAuthorizedGrantType, d.Get("authorized_grant_type").(string))
	assert.Equal(t, tState, d.Get("state").(string))
	assert.Equal(t, tAccessTokenValiditySeconds, d.Get("access_token_validity_seconds").(int))

}
