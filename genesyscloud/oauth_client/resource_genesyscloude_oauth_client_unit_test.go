package oauth_client

import (
	"context"
	"net/http"
	"sort"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"github.com/stretchr/testify/assert"
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

	ocproxy.getParentOAuthClientTokenAttr = func(context.Context, *oauthClientProxy) (*platformclientv2.Tokeninfo, *platformclientv2.APIResponse, error) {
		orgId := uuid.NewString()
		oauthId := uuid.NewString()
		oauthClientName := "CX as Code"
		name := "mytestorg"

		token := platformclientv2.Tokeninfo{
			Organization: &platformclientv2.Namedentity{
				Id:   &orgId,
				Name: &name,
			},
			HomeOrganization: &platformclientv2.Namedentity{
				Id: &orgId,
			},
			OAuthClient: &platformclientv2.Orgoauthclient{
				Id:   &oauthId,
				Name: &oauthClientName,
				Organization: &platformclientv2.Namedentity{
					Id: &orgId,
				},
			},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &token, apiResponse, nil
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

func TestUnitUpdateTerraformUserWithRole(t *testing.T) {
	getParentOAuthClientTokenCount := 0
	getTerraformUserCount := 0
	getTerraformUserRolesCount := 0
	updateTerraformUserRoleCount := 0

	userId := uuid.NewString()
	existingRoles := &[]platformclientv2.Domainrole{
		platformclientv2.Domainrole{Id: platformclientv2.String(uuid.NewString()), Name: platformclientv2.String("Master Admin")},
		platformclientv2.Domainrole{Id: platformclientv2.String(uuid.NewString()), Name: platformclientv2.String("Admin")},
		platformclientv2.Domainrole{Id: platformclientv2.String(uuid.NewString()), Name: platformclientv2.String("Employee")},
	}

	addedRoles := &[]platformclientv2.Roledivision{
		platformclientv2.Roledivision{RoleId: platformclientv2.String(uuid.NewString())},
		platformclientv2.Roledivision{RoleId: platformclientv2.String(uuid.NewString())},
	}

	contains := func(s []string, searchterm string) bool {
		sort.Strings(s)
		i := sort.SearchStrings(s, searchterm)
		result := (s[i] == searchterm)
		return result
	}
	ocproxy := &oauthClientProxy{}

	ocproxy.getParentOAuthClientTokenAttr = func(context.Context, *oauthClientProxy) (*platformclientv2.Tokeninfo, *platformclientv2.APIResponse, error) {
		orgId := uuid.NewString()
		oauthId := uuid.NewString()

		token := platformclientv2.Tokeninfo{
			Organization: &platformclientv2.Namedentity{
				Id:   &orgId,
				Name: platformclientv2.String("mytestorg"),
			},
			HomeOrganization: &platformclientv2.Namedentity{
				Id: &orgId,
			},
			OAuthClient: &platformclientv2.Orgoauthclient{
				Id:   &oauthId,
				Name: platformclientv2.String("Developer Tools"),
				Organization: &platformclientv2.Namedentity{
					Id: platformclientv2.String("purecloud-builtin"),
				},
			},
		}
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		getParentOAuthClientTokenCount++
		return &token, apiResponse, nil
	}

	ocproxy.getTerraformUserAttr = func(context.Context, *oauthClientProxy) (*platformclientv2.Userme, *platformclientv2.APIResponse, error) {
		userName := "Bill Smith"
		userEmail := "bill.smith@genesys.com"

		user := &platformclientv2.Userme{
			Id:    &userId,
			Name:  &userName,
			Email: &userEmail,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		getTerraformUserCount++
		return user, apiResponse, nil
	}

	ocproxy.getTerraformUserRolesAttr = func(ctx context.Context, proxy *oauthClientProxy, userId string) (*platformclientv2.Userauthorization, *platformclientv2.APIResponse, error) {
		userAuth := &platformclientv2.Userauthorization{
			Roles: existingRoles,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		getTerraformUserRolesCount++
		return userAuth, apiResponse, nil
	}

	ocproxy.updateTerraformUserRolesAttr = func(ctx context.Context, op *oauthClientProxy, userId string, roles []string) (*platformclientv2.Userauthorization, *platformclientv2.APIResponse, error) {
		assert.True(t, contains(roles, *(*existingRoles)[0].Id))
		assert.True(t, contains(roles, *(*existingRoles)[1].Id))
		assert.True(t, contains(roles, *(*existingRoles)[2].Id))
		assert.True(t, contains(roles, *(*addedRoles)[0].RoleId))
		assert.True(t, contains(roles, *(*addedRoles)[1].RoleId))
		updateTerraformUserRoleCount++
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = ocproxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	updateTerraformUserWithRole(ctx, &platformclientv2.Configuration{}, addedRoles)
	assert.Equal(t, getParentOAuthClientTokenCount, 1)
	assert.Equal(t, getTerraformUserCount, 1)
	assert.Equal(t, getTerraformUserRolesCount, 1)
	assert.Equal(t, updateTerraformUserRoleCount, 1)
}
