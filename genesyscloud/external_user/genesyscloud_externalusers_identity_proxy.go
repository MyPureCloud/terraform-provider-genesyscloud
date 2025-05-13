package external_user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"net/url"
	"strings"

	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
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
	externaIdObject, apiResponse, err := callExternalUserAPI(p.externalUserApi, userId, externalIdentity)

	if err != nil {
		return nil, apiResponse, err
	}
	return externaIdObject, apiResponse, err
}

func getAllExternalUserIdentityFn(ctx context.Context, p *externalUserIdentityProxy, userId string) (*[]platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error) {
	externalIdList, response, err := p.externalUserApi.GetUserExternalid(userId)
	for _, externalId := range externalIdList {
		if externalId.ExternalKey == nil || externalId.AuthorityName == nil {
			continue
		}
		rc.SetCache(p.externalUserIdentityCache, createCompoundKey(userId, *externalId.AuthorityName, *externalId.ExternalKey), externalId)
	}

	return &externalIdList, response, err
}

func getExternalUserIdentityByIdFn(ctx context.Context, p *externalUserIdentityProxy, userId, authorityName, externalKey string) (*platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error) {
	if externalId := rc.GetCacheItem(p.externalUserIdentityCache, createCompoundKey(userId, authorityName, externalKey)); externalId != nil {
		return externalId, nil, nil
	}
	externalIdList, response, err := getAllExternalUserIdentityFn(ctx, p, userId)
	if err != nil {
		return nil, response, err
	}
	if externalIdList == nil || len(*externalIdList) == 0 {
		response.StatusCode = 404
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
	apiResponse, err := p.externalUserApi.DeleteUserExternalidAuthorityNameExternalKey(userId, authorityName, externalKey)
	rc.DeleteCacheItem(p.externalUserIdentityCache, createCompoundKey(userId, authorityName, externalKey))
	return apiResponse, err
}

func callExternalUserAPI(userApi *platformclientv2.UsersApi, userId string, externalUser platformclientv2.Userexternalidentifier) (*platformclientv2.Userexternalidentifier, *platformclientv2.APIResponse, error) {
	var httpMethod = "POST"
	path := userApi.Configuration.BasePath + "/api/v2/users/{userId}/externalid"
	path = strings.Replace(path, "{userId}", url.PathEscape(fmt.Sprintf("%v", userId)), -1)

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)
	formParams := url.Values{}
	var postBody interface{}
	var postFileName string
	var fileBytes []byte

	if userApi.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + userApi.Configuration.AccessToken
	}
	for key := range userApi.Configuration.DefaultHeader {
		headerParams[key] = userApi.Configuration.DefaultHeader[key]
	}

	correctedQueryParams := make(map[string]string)
	for k, v := range queryParams {
		if k == "varType" {
			correctedQueryParams["type"] = v
			continue
		}
		correctedQueryParams[k] = v
	}
	queryParams = correctedQueryParams

	localVarHttpContentTypes := []string{"application/json"}

	localVarHttpContentType := userApi.Configuration.APIClient.SelectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		headerParams["Content-Type"] = localVarHttpContentType
	}
	localVarHttpHeaderAccepts := []string{
		"application/json",
	}

	localVarHttpHeaderAccept := userApi.Configuration.APIClient.SelectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		headerParams["Accept"] = localVarHttpHeaderAccept
	}
	postBody = &externalUser

	var successPayload *platformclientv2.Userexternalidentifier
	response, err := userApi.Configuration.APIClient.CallAPI(path, httpMethod, postBody, headerParams, queryParams, formParams, postFileName, fileBytes, "other")
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if err == nil && response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else if response.HasBody {
		json.Unmarshal(response.RawBody, &successPayload)

	}
	return successPayload, response, err
}
