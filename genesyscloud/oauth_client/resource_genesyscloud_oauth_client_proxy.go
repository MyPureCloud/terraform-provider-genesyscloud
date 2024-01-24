package oauth_client

import (
	"context"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

var internalProxy *oauthClientProxy

type createOAuthClientFunc func(context.Context, *oauthClientProxy, platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error)
type createIntegrationClientFunc func(context.Context, *oauthClientProxy, platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error)
type updateOAuthClientFunc func(context.Context, *oauthClientProxy, string, platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error)
type getOAuthClientFunc func(context.Context, *oauthClientProxy, string) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error)
type getIntegrationCredentialFunc func(context.Context, *oauthClientProxy, string) (*platformclientv2.Credential, *platformclientv2.APIResponse, error)
type getAllOauthClientsFunc func(ctx context.Context, o *oauthClientProxy) (*[]platformclientv2.Oauthclientlisting, error)
type deleteOAuthClientFunc func(context.Context, *oauthClientProxy, string) (*platformclientv2.APIResponse, error)
type deleteIntegrationCredentialFunc func(context.Context, *oauthClientProxy, string) (*platformclientv2.APIResponse, error)

type oauthClientProxy struct {
	clientConfig   *platformclientv2.Configuration
	api            *platformclientv2.OAuthApi
	integrationApi *platformclientv2.IntegrationsApi

	createOAuthClientAttr           createOAuthClientFunc
	createIntegrationCredentialAttr createIntegrationClientFunc
	getOAuthClientAttr              getOAuthClientFunc
	getAllOauthClientsAttr          getAllOauthClientsFunc
	getIntegrationCredentialAttr    getIntegrationCredentialFunc
	updateOAuthClientAttr           updateOAuthClientFunc
	deleteOAuthClientAttr           deleteOAuthClientFunc
	deleteIntegrationCredentialAttr deleteIntegrationCredentialFunc
}

// newArchitectIvrProxy initializes the proxy with all the data needed to communicate with Genesys Cloud
func newOAuthClientProxy(clientConfig *platformclientv2.Configuration) *oauthClientProxy {
	api := platformclientv2.NewOAuthApiWithConfig(clientConfig)
	intApi := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)

	return &oauthClientProxy{
		clientConfig:   clientConfig,
		api:            api,
		integrationApi: intApi,

		createOAuthClientAttr:           createOAuthClientFn,
		createIntegrationCredentialAttr: createIntegrationCredentialFn,
		updateOAuthClientAttr:           updateOAuthClientFn,
		getOAuthClientAttr:              getOAuthClientFn,
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

func (o *oauthClientProxy) getAllOAuthClients(ctx context.Context) (*[]platformclientv2.Oauthclientlisting, error) {
	return o.getAllOauthClientsAttr(ctx, o)
}

func getOAuthClientFn(ctx context.Context, o *oauthClientProxy, id string) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
	return o.api.GetOauthClient(id)
}

func getAllOauthClientsFn(ctx context.Context, o *oauthClientProxy) (*[]platformclientv2.Oauthclientlisting, error) {
	var clients []platformclientv2.Oauthclientlisting
	firstPage, _, err := o.api.GetOauthClients()

	if err != nil {
		return nil, err
	}

	for _, entity := range *firstPage.Entities {
		clients = append(clients, entity)

	}

	for pageNum := 2; pageNum <= *firstPage.PageCount; pageNum++ {
		page, _, err := o.api.GetOauthClients()

		if err != nil {
			return nil, err
		}

		for _, entity := range *page.Entities {
			clients = append(clients, entity)

		}
	}

	return &clients, nil
}

func getIntegrationClientFn(ctx context.Context, o *oauthClientProxy, id string) (*platformclientv2.Credential, *platformclientv2.APIResponse, error) {
	return o.integrationApi.GetIntegrationsCredential(id)
}

func deleteOAuthClientFn(ctx context.Context, o *oauthClientProxy, id string) (*platformclientv2.APIResponse, error) {
	return o.api.DeleteOauthClient(id)
}

func deleteIntegrationClientFn(ctx context.Context, o *oauthClientProxy, id string) (*platformclientv2.APIResponse, error) {
	return o.integrationApi.DeleteIntegrationsCredential(id)
}

func createOAuthClientFn(ctx context.Context, o *oauthClientProxy, request platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
	return o.api.PostOauthClients(request)
}

func createIntegrationCredentialFn(ctx context.Context, o *oauthClientProxy, request platformclientv2.Credential) (*platformclientv2.Credentialinfo, *platformclientv2.APIResponse, error) {
	return o.integrationApi.PostIntegrationsCredentials(request)
}

func updateOAuthClientFn(ctx context.Context, o *oauthClientProxy, id string, request platformclientv2.Oauthclientrequest) (*platformclientv2.Oauthclient, *platformclientv2.APIResponse, error) {
	return o.api.PutOauthClient(id, request)
}
