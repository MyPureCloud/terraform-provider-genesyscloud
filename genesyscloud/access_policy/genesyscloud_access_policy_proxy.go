package access_policy

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

/*
The genesyscloud_access_policy_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *accessPolicyProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createAccessPolicyFunc func(ctx context.Context, p *accessPolicyProxy, targetResource string, policy *platformclientv2.Authorizationpolicy) (*platformclientv2.Authorizationpolicy, *platformclientv2.APIResponse, error)
type getAllAccessPolicyFunc func(ctx context.Context, p *accessPolicyProxy) (*[]platformclientv2.Authorizationpolicy, *platformclientv2.APIResponse, error)
type getAccessPolicyIdByNameFunc func(ctx context.Context, p *accessPolicyProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getAccessPolicyByIdFunc func(ctx context.Context, p *accessPolicyProxy, id string) (policy *platformclientv2.Authorizationpolicy, response *platformclientv2.APIResponse, err error)
type updateAccessPolicyFunc func(ctx context.Context, p *accessPolicyProxy, id string, policy *platformclientv2.Authorizationpolicy) (*platformclientv2.Authorizationpolicy, *platformclientv2.APIResponse, error)
type deleteAccessPolicyFunc func(ctx context.Context, p *accessPolicyProxy, id string, targetResource string, subjectId string) (response *platformclientv2.APIResponse, err error)

// accessPolicyProxy contains all of the methods that call genesys cloud APIs.
type accessPolicyProxy struct {
	clientConfig                *platformclientv2.Configuration
	authorizationApi            *platformclientv2.AuthorizationApi
	createAccessPolicyAttr      createAccessPolicyFunc
	getAllAccessPolicyAttr      getAllAccessPolicyFunc
	getAccessPolicyIdByNameAttr getAccessPolicyIdByNameFunc
	getAccessPolicyByIdAttr     getAccessPolicyByIdFunc
	updateAccessPolicyAttr      updateAccessPolicyFunc
	deleteAccessPolicyAttr      deleteAccessPolicyFunc
}

// newAccessPolicyProxy initializes the access policy proxy with all of the data needed to communicate with Genesys Cloud
func newAccessPolicyProxy(clientConfig *platformclientv2.Configuration) *accessPolicyProxy {
	api := platformclientv2.NewAuthorizationApiWithConfig(clientConfig)
	return &accessPolicyProxy{
		clientConfig:                clientConfig,
		authorizationApi:            api,
		createAccessPolicyAttr:      createAccessPolicyFn,
		getAllAccessPolicyAttr:      getAllAccessPolicyFn,
		getAccessPolicyIdByNameAttr: getAccessPolicyIdByNameFn,
		getAccessPolicyByIdAttr:     getAccessPolicyByIdFn,
		updateAccessPolicyAttr:      updateAccessPolicyFn,
		deleteAccessPolicyAttr:      deleteAccessPolicyFn,
	}
}

// getAccessPolicyProxy acts as a singleton for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getAccessPolicyProxy(clientConfig *platformclientv2.Configuration) *accessPolicyProxy {
	if internalProxy == nil {
		internalProxy = newAccessPolicyProxy(clientConfig)
	}
	return internalProxy
}

// createAccessPolicy creates a Genesys Cloud access policy
func (p *accessPolicyProxy) createAccessPolicy(ctx context.Context, targetResource string, policy *platformclientv2.Authorizationpolicy) (*platformclientv2.Authorizationpolicy, *platformclientv2.APIResponse, error) {
	return p.createAccessPolicyAttr(ctx, p, targetResource, policy)
}

// getAllAccessPolicy retrieves all Genesys Cloud access policies
func (p *accessPolicyProxy) getAllAccessPolicy(ctx context.Context) (*[]platformclientv2.Authorizationpolicy, *platformclientv2.APIResponse, error) {
	return p.getAllAccessPolicyAttr(ctx, p)
}

// getAccessPolicyIdByName returns a single Genesys Cloud access policy by name
func (p *accessPolicyProxy) getAccessPolicyIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getAccessPolicyIdByNameAttr(ctx, p, name)
}

// getAccessPolicyById returns a single Genesys Cloud access policy by Id
func (p *accessPolicyProxy) getAccessPolicyById(ctx context.Context, id string) (policy *platformclientv2.Authorizationpolicy, response *platformclientv2.APIResponse, err error) {
	return p.getAccessPolicyByIdAttr(ctx, p, id)
}

// updateAccessPolicy updates a Genesys Cloud access policy
func (p *accessPolicyProxy) updateAccessPolicy(ctx context.Context, id string, policy *platformclientv2.Authorizationpolicy) (*platformclientv2.Authorizationpolicy, *platformclientv2.APIResponse, error) {
	return p.updateAccessPolicyAttr(ctx, p, id, policy)
}

// deleteAccessPolicy deletes a Genesys Cloud access policy
func (p *accessPolicyProxy) deleteAccessPolicy(ctx context.Context, id string, targetResource string, subjectId string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteAccessPolicyAttr(ctx, p, id, targetResource, subjectId)
}

// createAccessPolicyFn is an implementation function for creating a Genesys Cloud access policy
func createAccessPolicyFn(ctx context.Context, p *accessPolicyProxy, targetResource string, policy *platformclientv2.Authorizationpolicy) (*platformclientv2.Authorizationpolicy, *platformclientv2.APIResponse, error) {
	_ = provider.EnsureResourceContext(ctx, ResourceType)

	createdPolicy, resp, err := p.authorizationApi.PostAuthorizationPoliciesTarget(targetResource, *policy, false)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create access policy: %s", err)
	}
	return createdPolicy, resp, nil
}

// getAllAccessPolicyFn is the implementation for retrieving all access policies in Genesys Cloud
func getAllAccessPolicyFn(ctx context.Context, p *accessPolicyProxy) (*[]platformclientv2.Authorizationpolicy, *platformclientv2.APIResponse, error) {
	_ = provider.EnsureResourceContext(ctx, ResourceType)

	var allPolicies []platformclientv2.Authorizationpolicy
	const pageSize = 100
	var after string

	for {
		policies, resp, err := p.authorizationApi.GetAuthorizationPolicies(after, pageSize)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get access policies: %v", err)
		}
		if policies.Entities == nil || len(*policies.Entities) == 0 {
			break
		}

		allPolicies = append(allPolicies, *policies.Entities...)

		// Cursor-based pagination: use NextUri to determine if there are more pages
		if policies.NextUri == nil || *policies.NextUri == "" {
			break
		}

		// Extract the "after" cursor from the last entity's ID
		entities := *policies.Entities
		lastEntity := entities[len(entities)-1]
		if lastEntity.Id != nil {
			after = *lastEntity.Id
		} else {
			break
		}
	}

	return &allPolicies, nil, nil
}

// getAccessPolicyIdByNameFn is an implementation of the function to get a Genesys Cloud access policy by name
func getAccessPolicyIdByNameFn(ctx context.Context, p *accessPolicyProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	_ = provider.EnsureResourceContext(ctx, ResourceType)

	// The API doesn't support name-based search, so we must iterate through all policies
	policies, resp, getErr := p.getAllAccessPolicy(ctx)
	if getErr != nil {
		return "", false, resp, getErr
	}

	if policies == nil {
		return "", true, resp, fmt.Errorf("no access policy found with name %s", name)
	}

	for _, policy := range *policies {
		if policy.Name != nil && *policy.Name == name {
			log.Printf("Retrieved the access policy id %s by name %s", *policy.Id, name)
			return *policy.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("unable to find access policy with name %s", name)
}

// getAccessPolicyByIdFn is an implementation of the function to get a Genesys Cloud access policy by Id
func getAccessPolicyByIdFn(ctx context.Context, p *accessPolicyProxy, id string) (policy *platformclientv2.Authorizationpolicy, response *platformclientv2.APIResponse, err error) {
	_ = provider.EnsureResourceContext(ctx, ResourceType)

	policy, response, err = p.authorizationApi.GetAuthorizationPolicy(id)
	if err != nil {
		return nil, response, fmt.Errorf("failed to retrieve access policy by id %s: %s", id, err)
	}
	return policy, response, nil
}

// updateAccessPolicyFn is an implementation of the function to update a Genesys Cloud access policy
func updateAccessPolicyFn(ctx context.Context, p *accessPolicyProxy, id string, policy *platformclientv2.Authorizationpolicy) (*platformclientv2.Authorizationpolicy, *platformclientv2.APIResponse, error) {
	_ = provider.EnsureResourceContext(ctx, ResourceType)

	updatedPolicy, resp, err := p.authorizationApi.PutAuthorizationPolicy(id, *policy, false)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update access policy %s: %s", id, err)
	}
	return updatedPolicy, resp, nil
}

// deleteAccessPolicyFn is an implementation function for deleting a Genesys Cloud access policy
func deleteAccessPolicyFn(ctx context.Context, p *accessPolicyProxy, id string, targetResource string, subjectId string) (response *platformclientv2.APIResponse, err error) {
	_ = provider.EnsureResourceContext(ctx, ResourceType)

	response, err = p.authorizationApi.DeleteAuthorizationPoliciesTargetSubjectSubjectId(targetResource, subjectId)
	if err != nil {
		return response, fmt.Errorf("failed to delete access policy %s: %s", id, err)
	}
	return response, nil
}
