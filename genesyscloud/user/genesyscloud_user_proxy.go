package user

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
	"strings"
	"sync"
	"time"
)

var internalProxy *userProxy

type getAllUsersFunc func(ctx context.Context, p *userProxy) (*[]platformclientv2.User, error)
type updateUserRoutingUtilizationFunc func(ctx context.Context, p *userProxy, userId string, usersAPI *platformclientv2.Utilization) error
type deleteRoutingUserUtilizationFunc func(ctx context.Context, p *userProxy, userId string) error
type updateUserProfileSkillsFunc func(ctx context.Context, p *userProxy, userId string, skills []string) ([]string, *platformclientv2.APIResponse, error)
type getUserRoutingLanguagesFunc func(ctx context.Context, p *userProxy, userId string) (*[]platformclientv2.Userroutinglanguage, error)
type deleteUserRoutinglanguageFunc func(ctx context.Context, p *userProxy, userId string, langId string) (*platformclientv2.APIResponse, error)
type updateUserRoutinglanguagesBulkFunc func(ctx context.Context, p *userProxy, userId string, userRoutingLanguages []platformclientv2.Userroutinglanguagepost) (*platformclientv2.APIResponse, error)
type updateUserRoutingskillsBulkFunc func(ctx context.Context, p *userProxy, userId string, skdSkills []platformclientv2.Userroutingskillpost) (*platformclientv2.APIResponse, error)

// userProxy contains all of the method used to interact with the Genesys Scripts SDK
type userProxy struct {
	clientConfig                       *platformclientv2.Configuration
	usersApi                           *platformclientv2.UsersApi
	basePath                           string
	accessToken                        string
	getAllUsersAttr                    getAllUsersFunc
	updateUserRoutingUtilizationAttr   updateUserRoutingUtilizationFunc
	updateUserProfileSkillsAttr        updateUserProfileSkillsFunc
	deleteRoutingUserUtilizationAttr   deleteRoutingUserUtilizationFunc
	getUserRoutingLanguagesAttr        getUserRoutingLanguagesFunc
	deleteUserRoutinglanguageAttr      deleteUserRoutinglanguageFunc
	updateUserRoutinglanguagesBulkAttr updateUserRoutinglanguagesBulkFunc
	updateUserRoutingskillsBulkAttr    updateUserRoutingskillsBulkFunc
}

// getUserProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getUserProxy(clientConfig *platformclientv2.Configuration) *userProxy {
	if internalProxy == nil {
		internalProxy = newUserProxy(clientConfig)
	}

	return internalProxy
}

// newUserProxy initializes the Scripts proxy with all of the data needed to communicate with Genesys Cloud
func newUserProxy(clientConfig *platformclientv2.Configuration) *userProxy {
	usersAPI := platformclientv2.NewUsersApiWithConfig(clientConfig)

	return &userProxy{
		clientConfig:                       clientConfig,
		usersApi:                           usersAPI,
		basePath:                           strings.Replace(usersAPI.Configuration.BasePath, "api", "apps", -1),
		accessToken:                        usersAPI.Configuration.AccessToken,
		getAllUsersAttr:                    getAllUsersFN,
		updateUserRoutingUtilizationAttr:   updateUserRoutingUtilizationFN,
		updateUserProfileSkillsAttr:        updateUserProfileSkillsFN,
		deleteRoutingUserUtilizationAttr:   deleteRoutingUserUtilizationFN,
		getUserRoutingLanguagesAttr:        getUserRoutingLanguagesFN,
		deleteUserRoutinglanguageAttr:      deleteUserRoutingLanguageFN,
		updateUserRoutinglanguagesBulkAttr: updateUserRoutinglanguagesBulkFN,
		updateUserRoutingskillsBulkAttr:    updateUserRoutingskillsBulkFN,
	}
}
func (p *userProxy) getAllUsers(ctx context.Context) (*[]platformclientv2.User, error) {
	return p.getAllUsersAttr(ctx, p)
}

func (p *userProxy) updateUserRoutingUtilization(ctx context.Context, userId string, utilization *platformclientv2.Utilization) error {
	return p.updateUserRoutingUtilizationAttr(ctx, p, userId, utilization)
}

func (p *userProxy) deleteRoutingUserUtilization(ctx context.Context, userId string) error {
	return p.deleteRoutingUserUtilizationAttr(ctx, p, userId)
}

func (p *userProxy) updateUserProfileSkills(ctx context.Context, userId string, skills []string) ([]string, *platformclientv2.APIResponse, error) {
	return p.updateUserProfileSkillsAttr(ctx, p, userId, skills)
}

func (p *userProxy) getUserRoutingLanguages(ctx context.Context, userId string) (*[]platformclientv2.Userroutinglanguage, error) {
	return p.getUserRoutingLanguagesAttr(ctx, p, userId)
}

func (p *userProxy) deleteUserRoutinglanguage(ctx context.Context, userId string, langId string) (*platformclientv2.APIResponse, error) {
	return p.deleteUserRoutinglanguageAttr(ctx, p, userId, langId)
}

func (p *userProxy) updateUserRoutinglanguages(ctx context.Context, userId string, userRoutingLanguages []platformclientv2.Userroutinglanguagepost) (*platformclientv2.APIResponse, error) {
	return p.updateUserRoutinglanguagesBulkAttr(ctx, p, userId, userRoutingLanguages)
}

func (p *userProxy) updateUserRoutingSkills(ctx context.Context, userId string, sdkSkills []platformclientv2.Userroutingskillpost) (*platformclientv2.APIResponse, error) {
	return p.updateUserRoutingskillsBulkAttr(ctx, p, userId, sdkSkills)
}

func getAllUsersFN(ctx context.Context, p *userProxy) (*[]platformclientv2.User, error) {
	var totalUsers []platformclientv2.User

	// Newly created resources often aren't returned unless there's a delay
	time.Sleep(5 * time.Second)

	errorChan := make(chan error)
	wgDone := make(chan bool)
	userChan := make(chan platformclientv2.User)

	defer close(errorChan)
	defer close(userChan)

	// Cancel remaining goroutines if an error occurs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	//Anonymous function to look up all of the users for a particular status.  Since this is an anonymous function it can use "closures" to access variables in the out loop
	getUsersByStatus := func(userStatus string) {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			users, _, getErr := p.usersApi.GetUsers(pageSize, pageNum, nil, nil, "", nil, "", userStatus)
			if getErr != nil {
				select {
				case <-ctx.Done():
				case errorChan <- getErr:
				}
				cancel()
				return
			}

			if users.Entities == nil || len(*users.Entities) == 0 {
				break
			}

			for _, user := range *users.Entities {
				userChan <- user
			}
		}
	}

	//Spin off a go routine to retrieve all inactive users
	wg.Add(1)
	go func() {
		defer wg.Done()
		getUsersByStatus("inactive")
	}()

	// Spin off a go routine to retrieve all active users
	wg.Add(1)
	go func() {
		defer wg.Done()
		// get all active users
		getUsersByStatus("active")
	}()

	//Spin off a go routine to block until the wait group is done and then close the wait group is done
	go func() {
		wg.Wait()
		close(wgDone)
	}()

	// Wait until either WaitGroup is done or an error is received
	select {
	case user := <-userChan:
		totalUsers = append(totalUsers, user)
	case <-wgDone:
		return &totalUsers, nil
	case err := <-errorChan:
		return nil, fmt.Errorf("Failed to get page of users: %v", err)
	}
	return &totalUsers, nil
}

func updateUserRoutingUtilizationFN(ctx context.Context, p *userProxy, userId string, utilization *platformclientv2.Utilization) error {
	_, _, err := p.usersApi.PutRoutingUserUtilization(userId, *utilization)
	if err != nil {
		return err
	}
	return nil
}

func deleteRoutingUserUtilizationFN(ctx context.Context, p *userProxy, userId string) error {
	_, err := p.usersApi.DeleteRoutingUserUtilization(userId)
	if err != nil {
		return err
	}
	return nil
}

func updateUserProfileSkillsFN(ctx context.Context, p *userProxy, userId string, skills []string) ([]string, *platformclientv2.APIResponse, error) {
	return p.usersApi.PutUserProfileskills(userId, skills)
}

func getUserRoutingLanguagesFN(ctx context.Context, p *userProxy, userId string) (*[]platformclientv2.Userroutinglanguage, error) {
	var sdkLanguages []platformclientv2.Userroutinglanguage
	maxPageSize := 50
	for pageNum := 1; ; pageNum++ {
		langs, _, err := p.usersApi.GetUserRoutinglanguages(userId, maxPageSize, pageNum, "")
		if err != nil {
			return nil, fmt.Errorf("Failed to query languages for user %s: %s", userId, err)
		}
		if langs == nil || langs.Entities == nil || len(*langs.Entities) == 0 {
			return &sdkLanguages, nil
		}
		for _, language := range *langs.Entities {
			sdkLanguages = append(sdkLanguages, language)
		}
	}
}

func deleteUserRoutingLanguageFN(ctx context.Context, p *userProxy, userId string, langId string) (*platformclientv2.APIResponse, error) {
	return p.usersApi.DeleteUserRoutinglanguage(userId, langId)
}

func updateUserRoutinglanguagesBulkFN(ctx context.Context, p *userProxy, userId string, userRoutingLanguages []platformclientv2.Userroutinglanguagepost) (*platformclientv2.APIResponse, error) {
	_, resp, err := p.usersApi.PatchUserRoutinglanguagesBulk(userId, userRoutingLanguages)
	return resp, err
}

func updateUserRoutingskillsBulkFN(ctx context.Context, p *userProxy, userId string, sdkSkills []platformclientv2.Userroutingskillpost) (*platformclientv2.APIResponse, error) {
	_, resp, err := p.usersApi.PutUserRoutingskillsBulk(userId, sdkSkills)
	return resp, err
}
