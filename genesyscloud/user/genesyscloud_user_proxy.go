package user

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
	"log"
	"strings"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"time"
)

var internalProxy *userProxy

type getAllUsersFunc func(ctx context.Context, p *userProxy) (*[]platformclientv2.User, error)
type createUserFunc func(ctx context.Context, p *userProxy, targetUser platformclientv2.Createuser, email string) (*platformclientv2.User, error, bool)
type getDeletedUserIdFunc func(ctx context.Context, p *userProxy, email string) (*string, error)

//type publishScriptFunc func(ctx context.Context, p *scriptsProxy, scriptId string) error
//type getScriptByNameFunc func(ctx context.Context, p *scriptsProxy, scriptName string) ([]platformclientv2.Script, error)
//type verifyScriptUploadSuccessFunc func(ctx context.Context, p *scriptsProxy, body []byte) (bool, error)
//type scriptWasUploadedSuccessfullyFunc func(ctx context.Context, p *scriptsProxy, uploadId string) (bool, error)
//type getScriptExportUrlFunc func(ctx context.Context, p *scriptsProxy, scriptId string) (string, error)
//type deleteScriptFunc func(ctx context.Context, p *scriptsProxy, scriptId string) error
//type getScriptByIdFunc func(ctx context.Context, p *scriptsProxy, scriptId string) (script *platformclientv2.Script, statusCode int, err error)
//type getPublishedScriptsByNameFunc func(ctx context.Context, p *scriptsProxy, name string) (*[]platformclientv2.Script, error)

// userProxy contains all of the method used to interact with the Genesys Scripts SDK
type userProxy struct {
	mu                   sync.Mutex
	clientConfig         *platformclientv2.Configuration
	usersApi             *platformclientv2.UsersApi
	basePath             string
	accessToken          string
	getAllUsersAttr      getAllUsersFunc
	createUserAttr       createUserFunc
	getDeletedUserIdAttr getDeletedUserIdFunc
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
	mutex := &sync.Mutex{}
	return &userProxy{
		mu:                   *mutex,
		clientConfig:         clientConfig,
		usersApi:             usersAPI,
		basePath:             strings.Replace(usersAPI.Configuration.BasePath, "api", "apps", -1),
		accessToken:          usersAPI.Configuration.AccessToken,
		getAllUsersAttr:      getAllUsersFN,
		createUserAttr:       createUserFN,
		getDeletedUserIdAttr: getDeletedUserIdFN,
	}
}
func (p *userProxy) getAllUserScripts(ctx context.Context) (*[]platformclientv2.User, error) {
	return p.getAllUsersAttr(ctx, p)
}

func (p *userProxy) createUser(ctx context.Context, targetUser platformclientv2.Createuser, email string) (*platformclientv2.User, error, bool) {
	return p.createUserAttr(ctx, p, targetUser, email)
}

func (p *userProxy) getDeletedUserById(ctx context.Context, email string) (*string, error) {
	return p.getDeletedUserIdAttr(ctx, p, email)
}

func getAllUsersFN(ctx context.Context, p *userProxy) (*[]platformclientv2.User, error) {
	var totalUsers []platformclientv2.User

	// Newly created resources often aren't returned unless there's a delay
	time.Sleep(5 * time.Second)

	errorChan := make(chan error)
	wgDone := make(chan bool)
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
				p.mu.Lock()
				totalUsers = append(totalUsers, user)
				p.mu.Unlock()
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
	case <-wgDone:
		return &totalUsers, nil
	case err := <-errorChan:
		return nil, fmt.Errorf("Failed to get page of users: %v", err)
	}
}

func createUserFN(ctx context.Context, p *userProxy, targetUser platformclientv2.Createuser, email string) (*platformclientv2.User, error, bool) {
	user, resp, err := p.usersApi.PostUsers(targetUser)
	if err != nil {
		if resp != nil && resp.Error != nil && (*resp.Error).Code == "general.conflict" {
			// Check for a deleted user
			id, err := p.getDeletedUserById(ctx, email)
			if err != nil {
				return nil, err, true
			}
			if id != nil {
				//d.SetId(*id)
				return nil, restoreDeletedUser(ctx, d, meta, usersAPI), true
			}
		}
		return nil, fmt.Errorf("Failed to create user %s: %s", email, err), true
	}
	return user, nil, false
}

func getDeletedUserIdFN(ctx context.Context, p *userProxy, email string) (*string, error) {
	exactType := "EXACT"
	results, _, err := p.usersApi.PostUsersSearch(platformclientv2.Usersearchrequest{
		Query: &[]platformclientv2.Usersearchcriteria{
			{
				Fields:  &[]string{"email"},
				Value:   &email,
				VarType: &exactType,
			},
			{
				Fields:  &[]string{"state"},
				Values:  &[]string{"deleted"},
				VarType: &exactType,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to search for user %s: %s", email, err)
	}
	if results.Results != nil && len(*results.Results) > 0 {
		// User found
		return (*results.Results)[0].Id, nil
	}
	return nil, nil
}

//STOPPED HERE
func restoreDeletedUserFN(ctx context.Context, p *userProxy, email string, state string) err {
	email := d.Get("email").(string)
	state := d.Get("state").(string)

	log.Printf("Restoring deleted user %s", email)
	patchErr := patchUserWithState(d.Id(), "deleted", platformclientv2.Updateuser{
		State: &state,
	}, usersAPI)
	if patchErr != nil {
		return patchErr
	}
	return updateUser(ctx, d, meta)
}

func patchUser(id string, update platformclientv2.Updateuser, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	return patchUserWithState(id, "", update, usersAPI)
}

func patchUserWithState(id string, state string, update platformclientv2.Updateuser, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	return gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		currentUser, _, getErr := usersAPI.GetUser(id, nil, "", state)
		if getErr != nil {
			return nil, diag.Errorf("Failed to read user %s: %s", id, getErr)
		}

		update.Version = currentUser.Version
		_, resp, patchErr := usersAPI.PatchUser(id, update)
		if patchErr != nil {
			return resp, diag.Errorf("Failed to update user %s: %v", id, patchErr)
		}
		return nil, nil
	})
}
