package greeting_user

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

var internalProxy *greetingProxy

type getAllGreetingsFunc func(ctx context.Context, p *greetingProxy) (*[]platformclientv2.Domainentity, *platformclientv2.APIResponse, error)
type getUserGreetingByIdFunc func(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type updateUserGreetingFunc func(ctx context.Context, p *greetingProxy, greetingId string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type createUserGreetingFunc func(ctx context.Context, p *greetingProxy, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type deleteUserGreetingFunc func(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.APIResponse, error)

type greetingProxy struct {
	clientConfig        *platformclientv2.Configuration
	greetingsApi        *platformclientv2.GreetingsApi
	usersApi            *platformclientv2.UsersApi
	getAllGreetingsAttr getAllGreetingsFunc
	createGreetingAttr  createUserGreetingFunc
	getGreetingByIdAttr getUserGreetingByIdFunc
	updateGreetingAttr  updateUserGreetingFunc
	deleteGreetingAttr  deleteUserGreetingFunc
}

func newGreetingProxy(clientConfig *platformclientv2.Configuration) *greetingProxy {
	api := platformclientv2.NewGreetingsApiWithConfig(clientConfig)
	usersApi := platformclientv2.NewUsersApiWithConfig(clientConfig)
	return &greetingProxy{
		clientConfig:        clientConfig,
		greetingsApi:        api,
		usersApi:            usersApi,
		getAllGreetingsAttr: getAllGreetingsFn,
		createGreetingAttr:  createUserGreetingFn,
		getGreetingByIdAttr: getUserGreetingByIdFn,
		updateGreetingAttr:  updateUserGreetingFn,
		deleteGreetingAttr:  deleteUserGreetingFn,
	}
}

func getGreeetingProxy(clientConfig *platformclientv2.Configuration) *greetingProxy {
	if internalProxy == nil {
		internalProxy = newGreetingProxy(clientConfig)
	}

	return internalProxy
}

func (p *greetingProxy) getAllGreetings(ctx context.Context) (*[]platformclientv2.Domainentity, *platformclientv2.APIResponse, error) {
	return p.getAllGreetingsAttr(ctx, p)
}
func (p *greetingProxy) createUserGreeting(ctx context.Context, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.createGreetingAttr(ctx, p, body)
}
func (p *greetingProxy) getUserGreetingById(ctx context.Context, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.getGreetingByIdAttr(ctx, p, id)
}
func (p *greetingProxy) updateUserGreeting(ctx context.Context, greetingID string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.updateGreetingAttr(ctx, p, greetingID, body)
}
func (p *greetingProxy) deleteUserGreeting(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteGreetingAttr(ctx, p, id)
}
func getAllGreetingsFn(ctx context.Context, p *greetingProxy) (*[]platformclientv2.Domainentity, *platformclientv2.APIResponse, error) {
	var allGreetings []platformclientv2.Domainentity
	const pageSize = 100
	allUsers, resp, err := getAllUsersFn(ctx, p)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get users %s", err)
	}

	for _, user := range *allUsers {
		userGreetings, resp, err := p.greetingsApi.GetUserGreetings(*user.Id, pageSize, 1)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get greetings for user %s: %s", *user.Id, err)
		}
		if userGreetings.Entities != nil {
			allGreetings = append(allGreetings, *userGreetings.Entities...)
		}
		for pageNum := 2; pageNum <= *userGreetings.PageCount; pageNum++ {
			userGreetings, resp, err := p.greetingsApi.GetUserGreetings(*user.Id, pageSize, pageNum)
			if err != nil {
				return nil, resp, fmt.Errorf("failed to get greetings for user %s: %s", *user.Id, err)
			}
			if userGreetings.Entities != nil {
				allGreetings = append(allGreetings, *userGreetings.Entities...)
			}
		}
	}
	return &allGreetings, resp, nil
}

func createUserGreetingFn(ctx context.Context, p *greetingProxy, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.greetingsApi.PostUserGreetings(*body.Owner.Id, *body)
}
func getUserGreetingByIdFn(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.greetingsApi.GetGreeting(id)
}
func updateUserGreetingFn(ctx context.Context, p *greetingProxy, greetingId string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.greetingsApi.PutGreeting(greetingId, *body)
}
func deleteUserGreetingFn(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.greetingsApi.DeleteGreeting(id)
}
func getAllUsersFn(ctx context.Context, p *greetingProxy) (*[]platformclientv2.User, *platformclientv2.APIResponse, error) {
	var allUsers []platformclientv2.User
	const pageSize = 100

	users, resp, err := p.usersApi.GetUsers(pageSize, 1, nil, nil, "", nil, "", "")
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get users %s", err)
	}

	if users.Entities == nil || len(*users.Entities) == 0 {
		return &allUsers, resp, nil
	}
	allUsers = append(allUsers, *users.Entities...)

	for pageNum := 2; pageNum <= *users.PageCount; pageNum++ {
		users, resp, err := p.usersApi.GetUsers(pageSize, pageNum, nil, nil, "", nil, "", "")
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get users %s", err)
		}

		if users.Entities == nil || len(*users.Entities) == 0 {
			return &allUsers, resp, nil
		}

		allUsers = append(allUsers, *users.Entities...)
	}
	return &allUsers, resp, nil
}
