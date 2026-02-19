package users_rules

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

/*
The genesyscloud_users_rules_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *usersRulesProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createUsersRulesFunc func(ctx context.Context, p *usersRulesProxy, usersRules *platformclientv2.Usersrulescreaterulerequest) (*platformclientv2.Usersrulesrule, *platformclientv2.APIResponse, error)
type getAllUsersRulesFunc func(ctx context.Context, p *usersRulesProxy, searchTerm string) (*[]platformclientv2.Usersrulesrule, *platformclientv2.APIResponse, error)
type getUsersRulesIdByNameFunc func(ctx context.Context, p *usersRulesProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error)
type getUsersRulesByIdFunc func(ctx context.Context, p *usersRulesProxy, id string) (usersRules *platformclientv2.Usersrulesrule, response *platformclientv2.APIResponse, err error)
type updateUsersRulesFunc func(ctx context.Context, p *usersRulesProxy, id string, usersRules *platformclientv2.Usersrulesupdaterulerequest) (*platformclientv2.Usersrulesrule, *platformclientv2.APIResponse, error)
type deleteUsersRulesFunc func(ctx context.Context, p *usersRulesProxy, id string) (response *platformclientv2.APIResponse, err error)

// usersRulesProxy contains all of the methods that call genesys cloud APIs.
type usersRulesProxy struct {
	clientConfig              *platformclientv2.Configuration
	usersRulesApi             *platformclientv2.UsersRulesApi
	createUsersRulesAttr      createUsersRulesFunc
	getAllUsersRulesAttr      getAllUsersRulesFunc
	getUsersRulesIdByNameAttr getUsersRulesIdByNameFunc
	getUsersRulesByIdAttr     getUsersRulesByIdFunc
	updateUsersRulesAttr      updateUsersRulesFunc
	deleteUsersRulesAttr      deleteUsersRulesFunc
}

// newUsersRulesProxy initializes the users rules proxy with all of the data needed to communicate with Genesys Cloud
func newUsersRulesProxy(clientConfig *platformclientv2.Configuration) *usersRulesProxy {
	api := platformclientv2.NewUsersRulesApiWithConfig(clientConfig)
	return &usersRulesProxy{
		clientConfig:              clientConfig,
		usersRulesApi:             api,
		createUsersRulesAttr:      createUsersRulesFn,
		getAllUsersRulesAttr:      getAllUsersRulesFn,
		getUsersRulesIdByNameAttr: getUsersRulesIdByNameFn,
		getUsersRulesByIdAttr:     getUsersRulesByIdFn,
		updateUsersRulesAttr:      updateUsersRulesFn,
		deleteUsersRulesAttr:      deleteUsersRulesFn,
	}
}

// getUsersRulesProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getUsersRulesProxy(clientConfig *platformclientv2.Configuration) *usersRulesProxy {
	if internalProxy == nil {
		internalProxy = newUsersRulesProxy(clientConfig)
	}
	return internalProxy
}

// createUsersRules creates a Genesys Cloud users rule
func (p *usersRulesProxy) createUsersRules(ctx context.Context, usersRules *platformclientv2.Usersrulescreaterulerequest) (*platformclientv2.Usersrulesrule, *platformclientv2.APIResponse, error) {
	return p.createUsersRulesAttr(ctx, p, usersRules)
}

// getAllUsersRules retrieves all Genesys Cloud users rules
func (p *usersRulesProxy) getAllUsersRules(ctx context.Context, searchTerm string) (*[]platformclientv2.Usersrulesrule, *platformclientv2.APIResponse, error) {
	return p.getAllUsersRulesAttr(ctx, p, searchTerm)
}

// getUsersRulesIdByName returns a single Genesys Cloud users rule by name
func (p *usersRulesProxy) getUsersRulesIdByName(ctx context.Context, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getUsersRulesIdByNameAttr(ctx, p, name)
}

// getUsersRulesById returns a single Genesys Cloud users rule by Id
func (p *usersRulesProxy) getUsersRulesById(ctx context.Context, id string) (usersRules *platformclientv2.Usersrulesrule, response *platformclientv2.APIResponse, err error) {
	return p.getUsersRulesByIdAttr(ctx, p, id)
}

// updateUsersRules updates a Genesys Cloud users rule
func (p *usersRulesProxy) updateUsersRules(ctx context.Context, id string, usersRules *platformclientv2.Usersrulesupdaterulerequest) (*platformclientv2.Usersrulesrule, *platformclientv2.APIResponse, error) {
	return p.updateUsersRulesAttr(ctx, p, id, usersRules)
}

// deleteUsersRules deletes a Genesys Cloud users rule by Id
func (p *usersRulesProxy) deleteUsersRules(ctx context.Context, id string) (response *platformclientv2.APIResponse, err error) {
	return p.deleteUsersRulesAttr(ctx, p, id)
}

// createUsersRulesFn is an implementation function for creating a Genesys Cloud users rule
func createUsersRulesFn(ctx context.Context, p *usersRulesProxy, usersRules *platformclientv2.Usersrulescreaterulerequest) (*platformclientv2.Usersrulesrule, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	usersRule, resp, err := p.usersRulesApi.PostUsersRules(*usersRules)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to create users rule: %s", err)
	}
	return usersRule, resp, nil
}

// getAllUsersRulesFn is the implementation for retrieving all users rules in Genesys Cloud
func getAllUsersRulesFn(ctx context.Context, p *usersRulesProxy, searchTerm string) (*[]platformclientv2.Usersrulesrule, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	var allUsersRules []platformclientv2.Usersrulesrule
	types := []string{"Learning"}
	const pageSize = 100
	const pageNumber = 1
	const sortOrder = ""
	expand := []string{}
	const enabled = true

	usersRules, resp, err := p.usersRulesApi.GetUsersRules(
		types,
		pageNumber,
		pageSize,
		expand,
		enabled,
		searchTerm,
		sortOrder,
	)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to get users rules: %v", err)
	}
	if usersRules.Entities == nil || len(*usersRules.Entities) == 0 {
		return &allUsersRules, resp, nil
	}
	for _, usersRule := range *usersRules.Entities {
		allUsersRules = append(allUsersRules, usersRule)
	}

	for pageNum := 2; pageNum <= *usersRules.PageCount; pageNum++ {
		usersRules, resp, err := p.usersRulesApi.GetUsersRules(
			types,
			pageNum,
			pageSize,
			expand,
			enabled,
			searchTerm,
			sortOrder,
		)
		if err != nil {
			return nil, resp, fmt.Errorf("Failed to get users rules: %v", err)
		}

		if usersRules.Entities == nil || len(*usersRules.Entities) == 0 {
			break
		}

		for _, usersRule := range *usersRules.Entities {
			allUsersRules = append(allUsersRules, usersRule)
		}
	}

	return &allUsersRules, resp, nil
}

// getUsersRulesIdByNameFn is an implementation of the function to get a Genesys Cloud users rule by name
func getUsersRulesIdByNameFn(ctx context.Context, p *usersRulesProxy, name string) (id string, retryable bool, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	usersRules, resp, err := getAllUsersRulesFn(ctx, p, name)
	if err != nil {
		return "", false, resp, err
	}

	if usersRules == nil || len(*usersRules) == 0 {
		return "", true, resp, fmt.Errorf("No users rule found with name %s", name)
	}

	for _, usersRule := range *usersRules {
		if *usersRule.Name == name {
			log.Printf("Retrieved the users rule id %s by name %s", *usersRule.Id, name)
			return *usersRule.Id, false, resp, nil
		}
	}
	return "", true, resp, fmt.Errorf("Unable to find users rule with name %s", name)
}

// getUsersRulesByIdFn is an implementation of the function to get a Genesys Cloud users rule by Id
func getUsersRulesByIdFn(ctx context.Context, p *usersRulesProxy, id string) (usersRules *platformclientv2.Usersrulesrule, response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	usersRule, resp, err := p.usersRulesApi.GetUsersRule(id)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to retrieve users rule by id %s: %s", id, err)
	}
	return usersRule, resp, nil
}

// updateUsersRulesFn is an implementation of the function to update a Genesys Cloud users rule
func updateUsersRulesFn(ctx context.Context, p *usersRulesProxy, id string, usersRules *platformclientv2.Usersrulesupdaterulerequest) (*platformclientv2.Usersrulesrule, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	usersRule, resp, err := p.usersRulesApi.PatchUsersRule(id, *usersRules)
	if err != nil {
		return nil, resp, fmt.Errorf("Failed to update users rule: %s", err)
	}
	return usersRule, resp, nil
}

// deleteUsersRulesFn is an implementation function for deleting a Genesys Cloud users rule
func deleteUsersRulesFn(ctx context.Context, p *usersRulesProxy, id string) (response *platformclientv2.APIResponse, err error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	resp, err := p.usersRulesApi.DeleteUsersRule(id)
	if err != nil {
		return resp, fmt.Errorf("Failed to delete users rule: %s", err)
	}
	return resp, nil
}
