package user

import (
	"context"
	"fmt"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"log"
	"net/http"
	"strconv"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The file genesyscloud_user_proxy.go manages the interaction between our software and
the Genesys Cloud SDK. Within this file, we define proxy structures and methods.
We employ a technique called composition for each function on the proxy. This means that each function
is built by combining smaller, independent parts. One advantage of this approach is that it allows us
to isolate and test individual functions more easily. For testing purposes, we can replace or
simulate these smaller parts, known as stubs, to ensure that each function behaves correctly in different scenarios.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *userProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createUserFunc func(ctx context.Context, p *userProxy, createUser *platformclientv2.Createuser) (*platformclientv2.User, *platformclientv2.APIResponse, error)
type GetAllUserFunc func(ctx context.Context, p *userProxy) (*[]platformclientv2.User, *platformclientv2.APIResponse, error)
type getUserByIdFunc func(ctx context.Context, p *userProxy, id string, expand []string, state string) (user *platformclientv2.User, response *platformclientv2.APIResponse, err error)
type updateUserFunc func(ctx context.Context, p *userProxy, id string, updateUser *platformclientv2.Updateuser) (*platformclientv2.User, *platformclientv2.APIResponse, error)
type deleteUserFunc func(ctx context.Context, p *userProxy, id string) (*interface{}, *platformclientv2.APIResponse, error)
type patchUserWithStateFunc func(ctx context.Context, p *userProxy, id string, updateUser *platformclientv2.Updateuser) (*platformclientv2.User, *platformclientv2.APIResponse, error)
type hydrateUserCacheFunc func(ctx context.Context, p *userProxy, pageSize int, pageNum int) (*platformclientv2.Userentitylisting, *platformclientv2.APIResponse, error)
type getUserByNameFunc func(ctx context.Context, p *userProxy, searchUser platformclientv2.Usersearchrequest) (*platformclientv2.Userssearchresponse, *platformclientv2.APIResponse, error)
type getVoicemailUserpoliciesByIdFunc func(ctx context.Context, p *userProxy, id string) (*platformclientv2.Voicemailuserpolicy, *platformclientv2.APIResponse, error)
type updatePasswordFunc func(ctx context.Context, p *userProxy, id string, password string) (*platformclientv2.APIResponse, error)
type getTelephonyExtensionPoolByExtensionFunc func(ctx context.Context, p *userProxy, extNum string) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error)

/*
The userProxy struct holds all the methods responsible for making calls to
the Genesys Cloud APIs. This means that within this struct, you'll find all the functions designed
to interact directly with the various features and services offered by Genesys Cloud,
enabling this terraform provider software to perform tasks like retrieving data, updating information,
or triggering actions within the Genesys Cloud environment.
*/
type userProxy struct {
	clientConfig                             *platformclientv2.Configuration
	userApi                                  *platformclientv2.UsersApi
	routingApi                               *platformclientv2.RoutingApi
	voicemailApi                             *platformclientv2.VoicemailApi
	extensionPoolApi                         *platformclientv2.TelephonyProvidersEdgeApi
	createUserAttr                           createUserFunc
	GetAllUserAttr                           GetAllUserFunc
	getUserByIdAttr                          getUserByIdFunc
	updateUserAttr                           updateUserFunc
	deleteUserAttr                           deleteUserFunc
	patchUserWithStateAttr                   patchUserWithStateFunc
	hydrateUserCacheAttr                     hydrateUserCacheFunc
	getUserByNameAttr                        getUserByNameFunc
	getVoicemailUserpolicicesByIdAttr        getVoicemailUserpoliciesByIdFunc
	updatePasswordAttr                       updatePasswordFunc
	getTelephonyExtensionPoolByExtensionAttr getTelephonyExtensionPoolByExtensionFunc
	userCache                                rc.CacheInterface[platformclientv2.User] //Define the cache for user resource
	extensionPoolCache                       rc.CacheInterface[platformclientv2.Extensionpool]
}

var userCache = rc.NewResourceCache[platformclientv2.User]()
var extensionPoolCache = rc.NewResourceCache[platformclientv2.Extensionpool]()

/*
The function newUserProxy sets up the user proxy by providing it
with all the necessary information to communicate effectively with Genesys Cloud.
This includes configuring the proxy with the required data and settings so that it can interact
seamlessly with the Genesys Cloud platform.
*/
func newUserProxy(clientConfig *platformclientv2.Configuration) *userProxy {
	userApi := platformclientv2.NewUsersApiWithConfig(clientConfig)      // NewUsersApiWithConfig creates a Genesys Cloud API instance using the provided configuration
	routingApi := platformclientv2.NewRoutingApiWithConfig(clientConfig) // NewRoutingApiWithConfig creates a Genesys Cloud API instance using the provided configuration
	voicemailApi := platformclientv2.NewVoicemailApiWithConfig(clientConfig)
	extensionPoolApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)
	return &userProxy{
		clientConfig:                             clientConfig,
		userApi:                                  userApi,
		routingApi:                               routingApi,
		voicemailApi:                             voicemailApi,
		extensionPoolApi:                         extensionPoolApi,
		userCache:                                userCache,
		extensionPoolCache:                       extensionPoolCache,
		createUserAttr:                           createUserFn,
		GetAllUserAttr:                           GetAllUserFn,
		getUserByIdAttr:                          getUserByIdFn,
		updateUserAttr:                           updateUserFn,
		deleteUserAttr:                           deleteUserFn,
		patchUserWithStateAttr:                   patchUserWithStateFn,
		hydrateUserCacheAttr:                     hydrateUserCacheFn,
		getUserByNameAttr:                        getUserByNameFn,
		getVoicemailUserpolicicesByIdAttr:        getVoicemailUserpoliciesByUserIdFn,
		updatePasswordAttr:                       updatePasswordFn,
		getTelephonyExtensionPoolByExtensionAttr: getTelephonyExtensionPoolByExtensionFn,
	}
}

/*
GetUserProxy serves a dual purpose: first, it functions as a singleton for
the internalProxy, meaning it ensures that only one instance of the internalProxy exists. Second,
it enables us to proxy our tests by allowing us to directly set the internalProxy package variable.
This ensures consistency and control in managing the internalProxy across our codebase, while also
facilitating efficient testing by providing a straightforward way to substitute the proxy for testing purposes.

Note: The singleton pattern has been abandoned for this proxy because DEVTOOLING-991 introduced new endpoints that are sensitive
to rate limiting and ultimately increased the export time (we believe the singleton pattern prevents the resource from pulling
from the client pool and instead uses the one api token, leading to more 429s)
*/
func GetUserProxy(clientConfig *platformclientv2.Configuration) *userProxy {
	// use singleton approach if certain tests are running
	if userTestsAreActive() {
		return internalProxy
	}
	return newUserProxy(clientConfig)
}

// createUser creates a Genesys Cloud User
func (p *userProxy) createUser(ctx context.Context, createUser *platformclientv2.Createuser) (*platformclientv2.User, *platformclientv2.APIResponse, error) {
	return p.createUserAttr(ctx, p, createUser)
}

// GetAllUser retrieves all Genesys Cloud User
func (p *userProxy) GetAllUser(ctx context.Context) (*[]platformclientv2.User, *platformclientv2.APIResponse, error) {
	return p.GetAllUserAttr(ctx, p)
}

// getUserById returns a single Genesys Cloud User by Id
func (p *userProxy) getUserById(ctx context.Context, id string, expand []string, state string) (user *platformclientv2.User, response *platformclientv2.APIResponse, err error) {
	if user := rc.GetCacheItem(p.userCache, id); user != nil { // Get the user from the cache, if not there in the cache then call p.getUserByIdAttr()
		return user, nil, nil
	}
	return p.getUserByIdAttr(ctx, p, id, expand, state)
}

// updateUser updates a Genesys Cloud User
func (p *userProxy) updateUser(ctx context.Context, id string, updateUser *platformclientv2.Updateuser) (*platformclientv2.User, *platformclientv2.APIResponse, error) {
	return p.updateUserAttr(ctx, p, id, updateUser)
}

// deleteUser deletes a Genesys Cloud User by Id
func (p *userProxy) deleteUser(ctx context.Context, id string) (*interface{}, *platformclientv2.APIResponse, error) {
	return p.deleteUserAttr(ctx, p, id)
}

// patchUserWithState updates a Genesys Cloud User
func (p *userProxy) patchUserWithState(ctx context.Context, id string, updateUser *platformclientv2.Updateuser) (*platformclientv2.User, *platformclientv2.APIResponse, error) {
	return p.patchUserWithStateAttr(ctx, p, id, updateUser)
}

// hydrateUserCache
func (p *userProxy) hydrateUserCache(ctx context.Context, pageSize int, pageNum int) (*platformclientv2.Userentitylisting, *platformclientv2.APIResponse, error) {
	return p.hydrateUserCacheAttr(ctx, p, pageSize, pageNum)
}

// getUserByName
func (p *userProxy) getUserByName(ctx context.Context, searchUser platformclientv2.Usersearchrequest) (*platformclientv2.Userssearchresponse, *platformclientv2.APIResponse, error) {
	return p.getUserByNameAttr(ctx, p, searchUser)
}

// getVoicemailUserpoliciesById
func (p *userProxy) getVoicemailUserpoliciesById(ctx context.Context, id string) (*platformclientv2.Voicemailuserpolicy, *platformclientv2.APIResponse, error) {
	return p.getVoicemailUserpolicicesByIdAttr(ctx, p, id)
}

// updatePassword
func (p *userProxy) updatePassword(ctx context.Context, userId string, newPassword string) (*platformclientv2.APIResponse, error) {
	return p.updatePasswordAttr(ctx, p, userId, newPassword)
}

func (p *userProxy) getTelephonyExtensionPoolByExtension(ctx context.Context, extNum string) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	return p.getTelephonyExtensionPoolByExtensionAttr(ctx, p, extNum)
}

// createUserFn is an implementation function for creating a Genesys Cloud user
func createUserFn(_ context.Context, p *userProxy, createUser *platformclientv2.Createuser) (*platformclientv2.User, *platformclientv2.APIResponse, error) {
	return p.userApi.PostUsers(*createUser)
}

// getUserByIdFn is an implementation of the function to get a Genesys Cloud user by Id
func getUserByIdFn(_ context.Context, p *userProxy, id string, expand []string, state string) (user *platformclientv2.User, response *platformclientv2.APIResponse, err error) {
	return p.userApi.GetUser(id, expand, "", state)
}

// hydrateUserCacheFn
func hydrateUserCacheFn(_ context.Context, p *userProxy, pageSize int, pageNum int) (*platformclientv2.Userentitylisting, *platformclientv2.APIResponse, error) {
	return p.userApi.GetUsers(pageSize, 1, nil, nil, "", nil, "", "")
}

// getUserByNameFn
func getUserByNameFn(_ context.Context, p *userProxy, searchUser platformclientv2.Usersearchrequest) (*platformclientv2.Userssearchresponse, *platformclientv2.APIResponse, error) {
	return p.userApi.PostUsersSearch(searchUser)
}

// deleteUserFn is an implementation function for deleting a Genesys Cloud user
func deleteUserFn(_ context.Context, p *userProxy, id string) (*interface{}, *platformclientv2.APIResponse, error) {
	data, resp, err := p.userApi.DeleteUser(id)
	if err != nil {
		return nil, resp, err
	}
	rc.DeleteCacheItem(p.userCache, id)
	return data, nil, nil
}

func patchUserWithStateFn(_ context.Context, p *userProxy, id string, updateUser *platformclientv2.Updateuser) (*platformclientv2.User, *platformclientv2.APIResponse, error) {
	return p.userApi.PatchUser(id, *updateUser)
}

func updateUserFn(_ context.Context, p *userProxy, id string, updateUser *platformclientv2.Updateuser) (*platformclientv2.User, *platformclientv2.APIResponse, error) {
	return p.userApi.PatchUser(id, *updateUser)
}

// GetAllUserFn is the implementation for retrieving all user in Genesys Cloud
func GetAllUserFn(_ context.Context, p *userProxy) (*[]platformclientv2.User, *platformclientv2.APIResponse, error) {

	//Inner function to get user based on status
	getUsersByStatus := func(userStatus string) (*[]platformclientv2.User, *platformclientv2.APIResponse, error) {
		users := []platformclientv2.User{}
		const pageSize = 100
		expandedAttributes := []string{
			// Expands
			"skills",
			"languages",
			"locations",
			"profileSkills",
			"certifications",
			"employerInfo",
		}
		usersList, apiResponse, err := p.userApi.GetUsers(pageSize, 1, nil, nil, "", expandedAttributes, "", userStatus)
		if err != nil {
			return nil, apiResponse, err
		}
		users = append(users, *usersList.Entities...)

		for pageNum := 2; pageNum <= *usersList.PageCount; pageNum++ {
			usersList, apiResponse, err := p.userApi.GetUsers(pageSize, pageNum, nil, nil, "", expandedAttributes, "", userStatus)

			//DEVTOOLING-862 - This is a blocker for the BCP team as before this if check was put in the code would fail when it hit 10K of inactive users.
			//The BCP team (Cesar Branco has asked to write a warning to the log) and just return what we currently have.
			//Long-term solution is working with Joe Fruland to change the backend API.
			if userStatus == "inactive" && apiResponse != nil && apiResponse.StatusCode == http.StatusBadRequest {
				log.Printf("WARNING!!: The maximum number of inactive users (10,000) have been retrieved from the API.  No further exports of inactive users will occur.")
				return &users, apiResponse, nil
			}

			if err != nil {
				return nil, apiResponse, err
			}
			users = append(users, *usersList.Entities...)
		}

		return &users, apiResponse, nil
	}

	// Get all "active" and "inactive" users
	allUsers := []platformclientv2.User{}

	activeUsers, apiResponse, err := getUsersByStatus("active")
	if err != nil {
		return nil, apiResponse, fmt.Errorf("failed to get 'active' users %v", err)
	}
	allUsers = append(allUsers, *activeUsers...)

	inactiveUsers, apiResponse, err := getUsersByStatus("inactive")
	if err != nil {
		return nil, apiResponse, fmt.Errorf("failed to get 'inactive' users %v", err)
	}
	allUsers = append(allUsers, *inactiveUsers...)

	// Cache the architect schedules resource into the p.userCache for later use
	for _, user := range allUsers {
		rc.SetCache(p.userCache, *user.Id, user)
	}

	return &allUsers, apiResponse, nil
}

func getVoicemailUserpoliciesByUserIdFn(_ context.Context, p *userProxy, id string) (*platformclientv2.Voicemailuserpolicy, *platformclientv2.APIResponse, error) {
	return p.voicemailApi.GetVoicemailUserpolicy(id)
}

func updatePasswordFn(_ context.Context, p *userProxy, userId string, newPassword string) (*platformclientv2.APIResponse, error) {
	// Get the user's current password
	return p.userApi.PostUserPassword(userId, platformclientv2.Changepasswordrequest{
		NewPassword: &newPassword,
	})
}

func getTelephonyExtensionPoolByExtensionFn(_ context.Context, p *userProxy, extNum string) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var allPools []platformclientv2.Extensionpool

	extensionPoolList, apiResponse, err := p.extensionPoolApi.GetTelephonyProvidersEdgesExtensionpools(pageSize, 1, "", "", []string{})
	if err != nil {
		return nil, apiResponse, err
	}

	// Get the cached pools if they are available
	if rc.GetCacheSize(p.extensionPoolCache) == *extensionPoolList.Total && rc.GetCacheSize(p.extensionPoolCache) != 0 {
		allPools = *rc.GetCache(p.extensionPoolCache)
	} else if rc.GetCacheSize(p.extensionPoolCache) != *extensionPoolList.Total || rc.GetCacheSize(p.extensionPoolCache) != 0 {
		// The cache is populated but not with the right data, clear the cache so it can be re populated
		p.extensionPoolCache = rc.NewResourceCache[platformclientv2.Extensionpool]()

		allPools = append(allPools, *extensionPoolList.Entities...)

		for pageNumber := 2; pageNumber <= *extensionPoolList.PageCount; pageNumber++ {
			extensionPoolList, apiResponse, err := p.extensionPoolApi.GetTelephonyProvidersEdgesExtensionpools(pageSize, pageNumber, "", "", []string{})
			if err != nil {
				return nil, apiResponse, err
			}
			allPools = append(allPools, *extensionPoolList.Entities...)
		}
	}

	for _, pool := range allPools {
		rc.SetCache(p.extensionPoolCache, *pool.Id, pool)
	}

	extNumInt, err := strconv.Atoi(extNum)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid extension number: %s", err)
	}
	for _, pool := range allPools {
		startNum, err := strconv.Atoi(*pool.StartNumber)
		if err != nil {
			log.Printf("invalid start number: %v", err)
			continue
		}
		endNum, err := strconv.Atoi(*pool.EndNumber)
		if err != nil {
			log.Printf("invalid end number: %v", err)
			continue
		}

		if extNumInt > startNum && extNumInt < endNum {
			return &pool, apiResponse, nil
		}
	}

	return nil, nil, fmt.Errorf("unable to find corresponding extension pool")
}
