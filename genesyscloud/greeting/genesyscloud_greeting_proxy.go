package greeting

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

var internalProxy *greetingProxy

type getAllGreetingsFunc func(ctx context.Context, p *greetingProxy) (*[]platformclientv2.Domainentity, *platformclientv2.APIResponse, error)
type getGreetingByIdFunc func(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type updateGreetingFunc func(ctx context.Context, p *greetingProxy, greetingId string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type createGreetingFunc func(ctx context.Context, p *greetingProxy, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type deleteGreetingFunc func(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.APIResponse, error)

type greetingProxy struct {
	clientConfig        *platformclientv2.Configuration
	greetingsApi        *platformclientv2.GreetingsApi
	getAllGreetingsAttr getAllGreetingsFunc
	createGreetingAttr  createGreetingFunc
	getGreetingByIdAttr getGreetingByIdFunc
	updateGreetingAttr  updateGreetingFunc
	deleteGreetingAttr  deleteGreetingFunc
}

func newGreetingProxy(clientConfig *platformclientv2.Configuration) *greetingProxy {
	api := platformclientv2.NewGreetingsApiWithConfig(clientConfig)
	return &greetingProxy{
		clientConfig:        clientConfig,
		greetingsApi:        api,
		getAllGreetingsAttr: getAllGreetingsFn,
		createGreetingAttr:  createGreetingFn,
		getGreetingByIdAttr: getGreetingByIdFn,
		updateGreetingAttr:  updateGreetingFn,
		deleteGreetingAttr:  deleteGreetingFn,
	}
}

func getGreetingProxy(clientConfig *platformclientv2.Configuration) *greetingProxy {
	if internalProxy == nil {
		internalProxy = newGreetingProxy(clientConfig)
	}

	return internalProxy
}

func (p *greetingProxy) getAllGreetings(ctx context.Context) (*[]platformclientv2.Domainentity, *platformclientv2.APIResponse, error) {
	return p.getAllGreetingsAttr(ctx, p)
}
func (p *greetingProxy) createGreeting(ctx context.Context, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.createGreetingAttr(ctx, p, body)
}
func (p *greetingProxy) getGreetingById(ctx context.Context, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.getGreetingByIdAttr(ctx, p, id)
}
func (p *greetingProxy) updateGreeting(ctx context.Context, greetingID string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.updateGreetingAttr(ctx, p, greetingID, body)
}
func (p *greetingProxy) deleteGreeting(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteGreetingAttr(ctx, p, id)
}
func getAllGreetingsFn(ctx context.Context, p *greetingProxy) (*[]platformclientv2.Domainentity, *platformclientv2.APIResponse, error) {
	var allGreetings []platformclientv2.Domainentity
	const pageSize = 100
	orgGreetings, resp, err := p.greetingsApi.GetGreetings(pageSize, 1)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get greetings %s", err)
	}

	if orgGreetings.Entities != nil {
		allGreetings = append(allGreetings, *orgGreetings.Entities...)
	}
	for pageNum := 2; pageNum <= *orgGreetings.PageCount; pageNum++ {
		orgGreetings, resp, err := p.greetingsApi.GetGreetings(pageSize, pageNum)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get greetings %s", err)
		}
		if orgGreetings.Entities != nil {
			allGreetings = append(allGreetings, *orgGreetings.Entities...)
		}
	}
	return &allGreetings, resp, nil
}

func createGreetingFn(ctx context.Context, p *greetingProxy, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.greetingsApi.PostGreetings(*body)
}
func getGreetingByIdFn(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	found, resp, err := getGreetingFromOrganization(ctx, p, id)
	if err != nil {
		return nil, resp, err
	}
	if !found {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("greeting %s not found", id)
	}
	return p.greetingsApi.GetGreeting(id)
}
func updateGreetingFn(ctx context.Context, p *greetingProxy, greetingId string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.greetingsApi.PutGreeting(greetingId, *body)
}
func deleteGreetingFn(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.greetingsApi.DeleteGreeting(id)
}

func getGreetingFromOrganization(ctx context.Context, p *greetingProxy, id string) (bool, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	orgGreetings, resp, err := p.greetingsApi.GetGreetings(pageSize, 1)
	if err != nil {
		return false, resp, err
	}
	if containsGreetingInEntities(orgGreetings.Entities, id) {
		return true, resp, nil
	}

	for pageNum := 2; pageNum <= *orgGreetings.PageCount; pageNum++ {
		orgGreetings, resp, err = p.greetingsApi.GetGreetings(pageSize, pageNum)
		if err != nil {
			return false, resp, err
		}
		if containsGreetingInEntities(orgGreetings.Entities, id) {
			return true, resp, nil
		}
	}

	return false, resp, nil
}

func containsGreetingInEntities(entities *[]platformclientv2.Domainentity, greetingId string) bool {
	if entities == nil {
		return false
	}
	for _, entity := range *entities {
		if entity.Id != nil && *entity.Id == greetingId {
			return true
		}
	}
	return false
}
