package oauth_client

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

var internalProxy *oauthClientProxy

type createOAuthClientFunc func(context.Context, *oauthClientProxy, platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error)
type createIntegrationClientFunc func(context.Context, *oauthClientProxy, platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error)
type updateOAuthClientFunc func(context.Context, *oauthClientProxy, string, platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error)
type getOAuthClientFunc func(context.Context, *oauthClientProxy, string) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error)
type getParentOAuthClientTokenFunc func(context.Context, *oauthClientProxy) (*platformclientv2.Tokeninfo, *platformclientv2.APIResponse, error)
type getTerraformUserFunc func(context.Context, *oauthClientProxy) (*platformclientv2.Userme, *platformclientv2.APIResponse, error)
type getTerraformUserRolesFunc func(context.Context, *oauthClientProxy, string) (*platformclientv2.Userauthorization, *platformclientv2.APIResponse, error)
type updateTerraformUserRolesFunc func(context.Context, *oauthClientProxy, string, []string) (*platformclientv2.Userauthorization, *platformclientv2.APIResponse, error)
type getHomeDivisionInfoFunc func(context.Context, *oauthClientProxy) (*platformclientv2.Authzdivision, *platformclientv2.APIResponse, error)
type getIntegrationCredentialFunc func(context.Context, *oauthClientProxy, string) (*platformclientv2.Credential, *platformclientv2.APIResponse, error)
type getAllOauthClientsFunc func(ctx context.Context, o *oauthClientProxy) (*[]platformclientv2.Oauthclientlisting, *platformclientv2.APIResponse, error)
type deleteOAuthClientFunc func(context.Context, *oauthClientProxy, string) (*platformclientv2.APIResponse, error)
type deleteIntegrationCredentialFunc func(context.Context, *oauthClientProxy, string) (*platformclientv2.APIResponse, error)

type oauthClientProxy struct {
	clientConfig   *platformclientv2.Configuration
	oauthApi       *platformclientv2.OAuthApi
	usersApi       *platformclientv2.UsersApi
	tokenApi       *platformclientv2.TokensApi
	authApi        *platformclientv2.AuthorizationApi
	integrationApi *platformclientv2.IntegrationsApi

	createOAuthClientAttr           createOAuthClientFunc
	createIntegrationCredentialAttr createIntegrationClientFunc
	getOAuthClientAttr              getOAuthClientFunc
	getParentOAuthClientTokenAttr   getParentOAuthClientTokenFunc
	getTerraformUserAttr            getTerraformUserFunc
	getTerraformUserRolesAttr       getTerraformUserRolesFunc
	updateTerraformUserRolesAttr    updateTerraformUserRolesFunc
	getHomeDivisionInfoAttr         getHomeDivisionInfoFunc
	getAllOauthClientsAttr          getAllOauthClientsFunc
	getIntegrationCredentialAttr    getIntegrationCredentialFunc
	updateOAuthClientAttr           updateOAuthClientFunc
	deleteOAuthClientAttr           deleteOAuthClientFunc
	deleteIntegrationCredentialAttr deleteIntegrationCredentialFunc
}

// newArchitectIvrProxy initializes the proxy with all the data needed to communicate with Genesys Cloud
func newOAuthClientProxy(clientConfig *platformclientv2.Configuration) *oauthClientProxy {
	oauthApi := platformclientv2.NewOAuthApiWithConfig(clientConfig)
	usersApi := platformclientv2.NewUsersApiWithConfig(clientConfig)
	intApi := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)
	tokenApi := platformclientv2.NewTokensApiWithConfig(clientConfig)

	return &oauthClientProxy{
		clientConfig:   clientConfig,
		oauthApi:       oauthApi,
		usersApi:       usersApi,
		tokenApi:       tokenApi,
		integrationApi: intApi,

		createOAuthClientAttr:           createOAuthClientFn,
		createIntegrationCredentialAttr: createIntegrationCredentialFn,
		updateOAuthClientAttr:           updateOAuthClientFn,
		getOAuthClientAttr:              getOAuthClientFn,
		getParentOAuthClientTokenAttr:   getParentOAuthClientTokenFn,
		getTerraformUserRolesAttr:       getTerraformUserRolesFn,
		getTerraformUserAttr:            getTerraformUserFn,
		updateTerraformUserRolesAttr:    updateTerraformUserRolesFn,
		getHomeDivisionInfoAttr:         getHomeDivisionInfoFn,
		getIntegrationCredentialAttr:    getIntegrationClientFn,
		getAllOauthClientsAttr:          getAllOauthClientsFn,
		deleteOAuthClientAttr:           deleteOAuthClientFn,
		deleteIntegrationCredentialAttr: deleteIntegrationClientFn,
	}
}

func getOAuthClientProxy(clientConfig *platformclientv2.Configuration) *oauthClientProxy {
	if internalProxy == nil {
		internalProxy = newOAuthClientProxy(clientConfig)
	}
	return internalProxy
}

func (o *oauthClientProxy) deleteOAuthClient(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return o.deleteOAuthClientAttr(ctx, o, id)
}

func (o *oauthClientProxy) deleteIntegrationCredential(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return o.deleteIntegrationCredentialAttr(ctx, o, id)
}

func (o *oauthClientProxy) getOAuthClient(ctx context.Context, id string) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
	return o.getOAuthClientAttr(ctx, o, id)
}

func (o *oauthClientProxy) getParentOAuthClientToken(ctx context.Context) (*platformclientv2.Tokeninfo, *platformclientv2.APIResponse, error) {
	return o.getParentOAuthClientTokenAttr(ctx, o)
}

func (o *oauthClientProxy) GetTerraformUser(ctx context.Context) (*platformclientv2.Userme, *platformclientv2.APIResponse, error) {
	return o.getTerraformUserAttr(ctx, o)
}

func (o *oauthClientProxy) GetTerraformUserRoles(ctx context.Context, userId string) (*platformclientv2.Userauthorization, *platformclientv2.APIResponse, error) {
	return o.getTerraformUserRolesAttr(ctx, o, userId)
}

func (o *oauthClientProxy) UpdateTerraformUserRoles(ctx context.Context, userId string, roles []string) (*platformclientv2.Userauthorization, *platformclientv2.APIResponse, error) {
	return o.updateTerraformUserRolesAttr(ctx, o, userId, roles)
}

func (o *oauthClientProxy) getIntegrationCredential(ctx context.Context, id string) (*platformclientv2.Credential, *platformclientv2.APIResponse, error) {
	return o.getIntegrationCredentialAttr(ctx, o, id)
}

func (o *oauthClientProxy) createOAuthClient(ctx context.Context, oauthClient platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
	return o.createOAuthClientAttr(ctx, o, oauthClient)
}

func (o *oauthClientProxy) createIntegrationClient(ctx context.Context, credential platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	return o.createIntegrationCredentialAttr(ctx, o, credential)
}

func (o *oauthClientProxy) updateOAuthClient(ctx context.Context, id string, client platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
	return o.updateOAuthClientAttr(ctx, o, id, client)
}

func (o *oauthClientProxy) getAllOAuthClients(ctx context.Context) (*[]platformclientv2.Oauthclientlisting, *platformclientv2.APIResponse, error) {
	return o.getAllOauthClientsAttr(ctx, o)
}

func (o *oauthClientProxy) getHomeDivisionInfo(ctx context.Context) (*platformclientv2.Authzdivision, *platformclientv2.APIResponse, error) {
	return o.getHomeDivisionInfo(ctx)
}

func getOAuthClientFn(ctx context.Context, o *oauthClientProxy, id string) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
	return o.oauthApi.GetOauthClient(id)
}

func getParentOAuthClientTokenFn(ctx context.Context, o *oauthClientProxy) (*platformclientv2.Tokeninfo, *platformclientv2.APIResponse, error) {
	return o.tokenApi.GetTokensMe(false)
}

func getTerraformUserFn(ctx context.Context, o *oauthClientProxy) (*platformclientv2.Userme, *platformclientv2.APIResponse, error) {
	return o.usersApi.GetUsersMe(nil, "")
}

func getTerraformUserRolesFn(ctx context.Context, o *oauthClientProxy, userId string) (*platformclientv2.Userauthorization, *platformclientv2.APIResponse, error) {
	return o.usersApi.GetUserRoles(userId)
}

func updateTerraformUserRolesFn(ctx context.Context, o *oauthClientProxy, userId string, roles []string) (*platformclientv2.Userauthorization, *platformclientv2.APIResponse, error) {
	return o.usersApi.PutUserRoles(userId, roles)
}

func getHomeDivisionInfoFn(ctx context.Context, o *oauthClientProxy) (*platformclientv2.Authzdivision, *platformclientv2.APIResponse, error) {
	return o.authApi.GetAuthorizationDivisionsHome()
}

func getAllOauthClientsFn(ctx context.Context, o *oauthClientProxy) (*[]platformclientv2.Oauthclientlisting, *platformclientv2.APIResponse, error) {
	var clients []platformclientv2.Oauthclientlisting
	firstPage, resp, err := o.oauthApi.GetOauthClients()
	if err != nil {
		return nil, resp, err
	}

	for _, entity := range *firstPage.Entities {
		clients = append(clients, entity)
	}

	for pageNum := 2; pageNum <= *firstPage.PageCount; pageNum++ {
		page, resp, err := o.oauthApi.GetOauthClients()

		if err != nil {
			return nil, resp, err
		}

		for _, entity := range *page.Entities {
			clients = append(clients, entity)
		}
	}

	return &clients, resp, nil
}

func getIntegrationClientFn(ctx context.Context, o *oauthClientProxy, id string) (*platformclientv2.Credential, *platformclientv2.APIResponse, error) {
	return o.integrationApi.GetIntegrationsCredential(id)
}

func deleteOAuthClientFn(ctx context.Context, o *oauthClientProxy, id string) (*platformclientv2.APIResponse, error) {
	return o.oauthApi.DeleteOauthClient(id)
}

func deleteIntegrationClientFn(ctx context.Context, o *oauthClientProxy, id string) (*platformclientv2.APIResponse, error) {
	return o.integrationApi.DeleteIntegrationsCredential(id)
}

func createOAuthClientFn(ctx context.Context, o *oauthClientProxy, request platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
	return o.oauthApi.PostOauthClients(request)
}

func createIntegrationCredentialFn(ctx context.Context, o *oauthClientProxy, request platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	return o.integrationApi.PostIntegrationsCredentials(request)
}

func updateOAuthClientFn(ctx context.Context, o *oauthClientProxy, id string, request platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
	return o.oauthApi.PutOauthClient(id, request)
}
