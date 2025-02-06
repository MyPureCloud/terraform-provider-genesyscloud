package external_user

import (
	"context"
	"fmt"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v152/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *externalUserIdentityProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createExternalUserIdentityFunc func(ctx context.Context, p *externalUserIdentityProxy, userId string, externalIdentity platformclientv2.Userexternalidentifier) (*platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error)
type getAllExternalUserIdentityFunc func(ctx context.Context, p *externalUserIdentityProxy, userId string) (*[]platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error)
type getExternalUserIdentityByIdFunc func(ctx context.Context, p *externalUserIdentityProxy, userId, authorityName, externalKey string) (*platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error)
type deleteExternalUserIdentityFunc func(ctx context.Context, p *externalUserIdentityProxy, userId, authorityName, externalKey string) (*platformclientv2.APIResponse, error)

// ExternalUserIdentityProxy contains all of the methods that call genesys cloud APIs.
type externalUserIdentityProxy struct {
	clientConfig                    *platformclientv2.Configuration
	externalUserApi                 *platformclientv2.UsersApi
	createExternalUserIdentityAttr  createExternalUserIdentityFunc
	getAllExternalUserIdentityAttr  getAllExternalUserIdentityFunc
	getExternalUserIdentityByIdAttr getExternalUserIdentityByIdFunc
	deleteExternalUserIdentityAttr  deleteExternalUserIdentityFunc
	externalUserIdentityCache       rc.CacheInterface[platformclientv2.Userexternalidentifier]
}

func newExternalUserIdentityProxy(clientConfig *platformclientv2.Configuration) *externalUserIdentityProxy {
	api := platformclientv2.NewUsersApiWithConfig(clientConfig)
	externalUserIdentityCache := rc.NewResourceCache[platformclientv2.Userexternalidentifier]()
	return &externalUserIdentityProxy{
		clientConfig:                    clientConfig,
		externalUserApi:                 api,
		createExternalUserIdentityAttr:  createExternalUserIdentityFn,
		getAllExternalUserIdentityAttr:  getAllExternalUserIdentityFn,
		getExternalUserIdentityByIdAttr: getExternalUserIdentityByIdFn,
		deleteExternalUserIdentityAttr:  deleteExternalUserIdentityFn,
		externalUserIdentityCache:       externalUserIdentityCache,
	}
}

func getExternalUserIdentityProxy(clientConfig *platformclientv2.Configuration) *externalUserIdentityProxy {
	if internalProxy == nil {
		internalProxy = newExternalUserIdentityProxy(clientConfig)
	}

	return internalProxy
}

func (p *externalUserIdentityProxy) createExternalUserIdentity(ctx context.Context, userId string, externalIdentity platformclientv2.Userexternalidentifier) (*platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error) {
	return p.createExternalUserIdentityAttr(ctx, p, userId, externalIdentity)
}

func (p *externalUserIdentityProxy) getAllExternalUserIdentity(ctx context.Context, userId string) (*[]platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error) {
	return p.getAllExternalUserIdentityAttr(ctx, p, userId)
}

func (p *externalUserIdentityProxy) getExternalUserIdentityById(ctx context.Context, userId, authorityName, externalKey string) (*platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error) {
	return p.getExternalUserIdentityByIdAttr(ctx, p, userId, authorityName, externalKey)
}

func (p *externalUserIdentityProxy) deleteExternalUserIdentity(ctx context.Context, userId, authorityName, externalKey string) (*platformclientv2.APIResponse, error) {
	return p.deleteExternalUserIdentityAttr(ctx, p, userId, authorityName, externalKey)
}

func createExternalUserIdentityFn(ctx context.Context, p *externalUserIdentityProxy, userId string, externalIdentity platformclientv2.Userexternalidentifier) (*platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error) {
	externalIdList, apiResponse, err := p.externalUserApi.PostUserExternalid(userId, externalIdentity)
	if err != nil {
		return nil, apiResponse, err
	}
	if len(externalIdList) == 0 {
		return nil, apiResponse, fmt.Errorf("could not create a external User Identity for userId:%s ", userId)
	}
	externaIdObject := externalIdList[0]
	return &externaIdObject, apiResponse, err
}

func getAllExternalUserIdentityFn(ctx context.Context, p *externalUserIdentityProxy, userId string) (*[]platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error) {
	externalIdList, response, err := p.externalUserApi.GetUserExternalid(userId)
	return &externalIdList, response, err
}

func getExternalUserIdentityByIdFn(ctx context.Context, p *externalUserIdentityProxy, userId, authorityName, externalKey string) (*platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error) {

	externalIdList, response, err := getAllExternalUserIdentityFn(ctx, p, userId)
	if err != nil {
		return nil, response, err
	}
	if externalIdList == nil || len(*externalIdList) == 0 {
		return nil, response, fmt.Errorf("could not find a external User Identity for userId :%s authorityName:%s externalKey:%s ", userId, authorityName, externalKey)
	}
	for _, externalId := range *externalIdList {
		if *externalId.ExternalKey == externalKey && *externalId.AuthorityName == authorityName {
			externalIdCopy := externalId
			return &externalIdCopy, response, nil
		}
	}
	return nil, response, fmt.Errorf("could not find a external User Identity for userId :%s authorityName:%s externalKey:%s ", userId, authorityName, externalKey)
}

func deleteExternalUserIdentityFn(ctx context.Context, p *externalUserIdentityProxy, userId, authorityName, externalKey string) (*platformclientv2.APIResponse, error) {
	return p.externalUserApi.DeleteUserExternalidAuthorityNameExternalKey(userId, authorityName, externalKey)
}
