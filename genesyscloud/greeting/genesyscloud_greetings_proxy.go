package greeting

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

// Part
var internalProxy *greetingProxy

type getAllGreetingsFunc func(ctx context.Context, p *greetingProxy) (*[]platformclientv2.Domainentity, *platformclientv2.APIResponse, error)
type getGreetingByIdFunc func(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type updateGreetingFunc func(ctx context.Context, p *greetingProxy, greetingId string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type createGreetingFunc func(ctx context.Context, p *greetingProxy, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type deleteGreetingFunc func(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.APIResponse, error)

// Part
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

func getGreeetingProxy(clientConfig *platformclientv2.Configuration) *greetingProxy {
	if internalProxy == nil {
		internalProxy = newGreetingProxy(clientConfig)
	}

	return internalProxy
}

// Part
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

// Part
func getAllGreetingsFn(ctx context.Context, p *greetingProxy) (*[]platformclientv2.Domainentity, *platformclientv2.APIResponse, error) {
	var allGreetings []platformclientv2.Domainentity
	const pageSize = 100

	greetings, resp, err := p.greetingsApi.GetGreetings(pageSize, 1)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get greetings %s", err)
	}

	if greetings.Entities == nil || len(*greetings.Entities) == 0 {
		return &allGreetings, resp, nil
	}
	allGreetings = append(allGreetings, *greetings.Entities...)

	for pageNum := 2; pageNum <= *greetings.PageCount; pageNum++ {
		greetings, resp, err := p.greetingsApi.GetGreetings(pageSize, pageNum)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get greetings %s", err)
		}

		if greetings.Entities == nil || len(*greetings.Entities) == 0 {
			return &allGreetings, resp, nil
		}

		allGreetings = append(allGreetings, *greetings.Entities...)
	}
	return &allGreetings, resp, nil
}

func createGreetingFn(ctx context.Context, p *greetingProxy, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.greetingsApi.PostGreetings(*body)
}
func getGreetingByIdFn(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.greetingsApi.GetGreeting(id)
}
func updateGreetingFn(ctx context.Context, p *greetingProxy, greetingId string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.greetingsApi.PutGreeting(greetingId, *body)
}
func deleteGreetingFn(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.greetingsApi.DeleteGreeting(id)
}
