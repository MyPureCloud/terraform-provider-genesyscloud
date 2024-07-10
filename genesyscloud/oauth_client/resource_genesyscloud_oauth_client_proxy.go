package oauth_client

import (
	"context"
	"log"
	"sync"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
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
type getIntegrationCredentialFunc func(context.Context, *oauthClientProxy, string) (*platformclientv2.Credential, *platformclientv2.APIResponse, error)
type getAllOauthClientsFunc func(ctx context.Context, o *oauthClientProxy) (*[]platformclientv2.Oauthclientlisting, *platformclientv2.APIResponse, error)
type deleteOAuthClientFunc func(context.Context, *oauthClientProxy, string) (*platformclientv2.APIResponse, error)
type deleteIntegrationCredentialFunc func(context.Context, *oauthClientProxy, string) (*platformclientv2.APIResponse, error)

type oauthClientProxy struct {
	clientConfig                    *platformclientv2.Configuration
	oAuthApi                        *platformclientv2.OAuthApi
	integrationApi                  *platformclientv2.IntegrationsApi
	tokenApi                        *platformclientv2.TokensApi
	usersApi                        *platformclientv2.UsersApi
	createdClientCache              map[string]platformclientv2.Oauthclient //Being added for DEVTOOLING-448
	createdClientCacheLock          sync.Mutex
	createOAuthClientAttr           createOAuthClientFunc
	createIntegrationCredentialAttr createIntegrationClientFunc
	getOAuthClientAttr              getOAuthClientFunc
	getParentOAuthClientTokenAttr   getParentOAuthClientTokenFunc
	getTerraformUserAttr            getTerraformUserFunc
	getTerraformUserRolesAttr       getTerraformUserRolesFunc
	updateTerraformUserRolesAttr    updateTerraformUserRolesFunc
	getAllOauthClientsAttr          getAllOauthClientsFunc
	getIntegrationCredentialAttr    getIntegrationCredentialFunc
	updateOAuthClientAttr           updateOAuthClientFunc
	deleteOAuthClientAttr           deleteOAuthClientFunc
	deleteIntegrationCredentialAttr deleteIntegrationCredentialFunc
}

// newAuthClientProxy initializes the proxy with all the data needed to communicate with Genesys Cloud
func newOAuthClientProxy(clientConfig *platformclientv2.Configuration) *oauthClientProxy {

	oAuthApi := platformclientv2.NewOAuthApiWithConfig(clientConfig)
	intApi := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)
	usersApi := platformclientv2.NewUsersApiWithConfig(clientConfig)
	createdClientCache := make(map[string]platformclientv2.Oauthclient)
	tokenApi := platformclientv2.NewTokensApiWithConfig(clientConfig)

	return &oauthClientProxy{
		clientConfig:       clientConfig,
		oAuthApi:           oAuthApi,
		integrationApi:     intApi,
		usersApi:           usersApi,
		tokenApi:           tokenApi,
		createdClientCache: createdClientCache,

		createOAuthClientAttr:           createOAuthClientFn,
		createIntegrationCredentialAttr: createIntegrationCredentialFn,
		updateOAuthClientAttr:           updateOAuthClientFn,
		getOAuthClientAttr:              getOAuthClientFn,
		getParentOAuthClientTokenAttr:   getParentOAuthClientTokenFn,
		getTerraformUserRolesAttr:       getTerraformUserRolesFn,
		getTerraformUserAttr:            getTerraformUserFn,
		updateTerraformUserRolesAttr:    updateTerraformUserRolesFn,

		getIntegrationCredentialAttr:    getIntegrationClientFn,
		getAllOauthClientsAttr:          getAllOauthClientsFn,
		deleteOAuthClientAttr:           deleteOAuthClientFn,
		deleteIntegrationCredentialAttr: deleteIntegrationClientFn,
	}
}

/*
Note:  Normally we do not make proxies or their methods public outside the package. However, we are doing this
specifically for DEVTOOLING-448.  In DEVTOOLING-448, we are adding the ability to cache a OAuthClient that was
created in a Terraform run so that when can use that secret to create a Genesys Cloud Integration Credential in the same
run without having to expose the secret.

We need this so that we can support the ability run CX as Code Accelerator where we can create a OAuth Client, a Role with Permissions
and then an OAuth client with out the need for the user to support passing the Genesys Cloud OAuth Client Credentials
into the integration credential object.  Today the integration credential object has no way of looking up the client id/client secret
without because once the oauth client is created, we dont want to expose the secret.
*/
func GetOAuthClientProxy(clientConfig *platformclientv2.Configuration) *oauthClientProxy {
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

func (o *oauthClientProxy) GetCachedOAuthClient(clientId string) platformclientv2.Oauthclient {
	o.createdClientCacheLock.Lock()
	defer o.createdClientCacheLock.Unlock()
	return o.createdClientCache[clientId]
}

func (o *oauthClientProxy) createOAuthClient(ctx context.Context, oauthClient platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
	oauthClientResult, response, err := o.createOAuthClientAttr(ctx, o, oauthClient)
	if err != nil {
		return oauthClientResult, response, err
	}

	//Being added for DEVTOOLING-448.  This is one of the few places where we want to use a cache outside the export
	o.createdClientCacheLock.Lock()
	defer o.createdClientCacheLock.Unlock()
	o.createdClientCache[*oauthClientResult.Id] = *oauthClientResult
	log.Printf("Successfully added oauth client %s to cache", *oauthClientResult.Id)
	return oauthClientResult, response, err
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
	return o.oAuthApi.GetOauthClient(id)
}
func getAllOauthClientsFn(ctx context.Context, o *oauthClientProxy) (*[]platformclientv2.Oauthclientlisting, *platformclientv2.APIResponse, error) {
	var clients []platformclientv2.Oauthclientlisting
	firstPage, resp, err := o.oAuthApi.GetOauthClients()

	if err != nil {
		return nil, resp, err
	}

	clients = append(clients, *firstPage.Entities...)

	for pageNum := 2; pageNum <= *firstPage.PageCount; pageNum++ {
		page, resp, err := o.oAuthApi.GetOauthClients()

		if err != nil {
			return nil, resp, err
		}

		clients = append(clients, *page.Entities...)
	}

	return &clients, resp, nil
}

func getIntegrationClientFn(ctx context.Context, o *oauthClientProxy, id string) (*platformclientv2.Credential, *platformclientv2.APIResponse, error) {
	return o.integrationApi.GetIntegrationsCredential(id)
}

func deleteOAuthClientFn(ctx context.Context, o *oauthClientProxy, id string) (*platformclientv2.APIResponse, error) {
	return o.oAuthApi.DeleteOauthClient(id)

}

func deleteIntegrationClientFn(ctx context.Context, o *oauthClientProxy, id string) (*platformclientv2.APIResponse, error) {
	return o.integrationApi.DeleteIntegrationsCredential(id)
}

func createOAuthClientFn(ctx context.Context, o *oauthClientProxy, request platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
	return o.oAuthApi.PostOauthClients(request)
}

func createIntegrationCredentialFn(ctx context.Context, o *oauthClientProxy, request platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	return o.integrationApi.PostIntegrationsCredentials(request)
}

func updateOAuthClientFn(ctx context.Context, o *oauthClientProxy, id string, request platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
	return o.oAuthApi.PutOauthClient(id, request)
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
