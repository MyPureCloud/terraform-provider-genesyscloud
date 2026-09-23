package greeting_user

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
)

var internalProxy *greetingProxy
var greetingCache = rc.NewResourceCache[platformclientv2.Greeting]()

const (
	userGreetingsPageSize          = 100
	defaultUserGreetingConcurrency = 10
	maxUserGreetingConcurrency     = 20
	userGreetingProgressInterval   = 500
)

type getAllGreetingsFunc func(ctx context.Context, p *greetingProxy) (*[]platformclientv2.Domainentity, *platformclientv2.APIResponse, error)
type getUserGreetingByIdFunc func(ctx context.Context, p *greetingProxy, userId string, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type updateUserGreetingFunc func(ctx context.Context, p *greetingProxy, greetingId string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type createUserGreetingFunc func(ctx context.Context, p *greetingProxy, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error)
type deleteUserGreetingFunc func(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.APIResponse, error)

type greetingProxy struct {
	clientConfig        *platformclientv2.Configuration
	greetingsApi        *platformclientv2.GreetingsApi
	getAllGreetingsAttr getAllGreetingsFunc
	createGreetingAttr  createUserGreetingFunc
	getGreetingByIdAttr getUserGreetingByIdFunc
	updateGreetingAttr  updateUserGreetingFunc
	deleteGreetingAttr  deleteUserGreetingFunc
	greetingCache       rc.CacheInterface[platformclientv2.Greeting]
}

func newGreetingProxy(clientConfig *platformclientv2.Configuration) *greetingProxy {
	api := platformclientv2.NewGreetingsApiWithConfig(clientConfig)
	return &greetingProxy{
		clientConfig:        clientConfig,
		greetingsApi:        api,
		getAllGreetingsAttr: getAllGreetingsFn,
		createGreetingAttr:  createUserGreetingFn,
		getGreetingByIdAttr: getUserGreetingByIdFn,
		updateGreetingAttr:  updateUserGreetingFn,
		deleteGreetingAttr:  deleteUserGreetingFn,
		greetingCache:       greetingCache,
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
func (p *greetingProxy) getUserGreetingById(ctx context.Context, userId string, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	if greeting := rc.GetCacheItem(p.greetingCache, id); greeting != nil {
		return greeting, nil, nil
	}
	return p.getGreetingByIdAttr(ctx, p, userId, id)
}
func (p *greetingProxy) updateUserGreeting(ctx context.Context, greetingID string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	return p.updateGreetingAttr(ctx, p, greetingID, body)
}
func (p *greetingProxy) deleteUserGreeting(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	rc.DeleteCacheItem(p.greetingCache, id)
	return p.deleteGreetingAttr(ctx, p, id)
}

type userGreetingCollectResult struct {
	entities []platformclientv2.Domainentity
	resp     *platformclientv2.APIResponse
	err      error
}

func getAllGreetingsFn(ctx context.Context, p *greetingProxy) (*[]platformclientv2.Domainentity, *platformclientv2.APIResponse, error) {
	allUsers, resp, err := getAllUsersForGreetingExport(ctx, p.clientConfig)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get users: %w", err)
	}
	if allUsers == nil || len(*allUsers) == 0 {
		empty := []platformclientv2.Domainentity{}
		return &empty, resp, nil
	}

	userIDs := make([]string, 0, len(*allUsers))
	for _, u := range *allUsers {
		if u.Id != nil {
			userIDs = append(userIDs, *u.Id)
		}
	}
	if len(userIDs) == 0 {
		empty := []platformclientv2.Domainentity{}
		return &empty, resp, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	concurrency := userGreetingFetchConcurrency()
	if concurrency > len(userIDs) {
		concurrency = len(userIDs)
	}
	log.Printf("Discovering user greetings for %d users (concurrency=%d)", len(userIDs), concurrency)

	jobs := make(chan string, len(userIDs))
	for _, userID := range userIDs {
		jobs <- userID
	}
	close(jobs)

	results := make(chan userGreetingCollectResult, concurrency)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			greetingsApi, releaseGreetingsApi, apiErr := greetingsApiForWorker(ctx, p.clientConfig)
			if apiErr != nil {
				results <- userGreetingCollectResult{err: fmt.Errorf("failed to acquire greetings API client: %w", apiErr)}
				return
			}
			defer releaseGreetingsApi()

			for userID := range jobs {
				if ctx.Err() != nil {
					return
				}
				entities, pageResp, pageErr := collectUserGreetingsForUser(ctx, p, greetingsApi, userID)
				results <- userGreetingCollectResult{entities: entities, resp: pageResp, err: pageErr}
				if pageErr != nil {
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allGreetings []platformclientv2.Domainentity
	var lastResp *platformclientv2.APIResponse
	var collectErr error
	var processedUsers int64

	for result := range results {
		if result.err != nil {
			if collectErr == nil {
				collectErr = result.err
				cancel()
			}
			continue
		}
		if result.resp != nil {
			lastResp = result.resp
		}
		if len(result.entities) > 0 {
			allGreetings = append(allGreetings, result.entities...)
		}
		completed := atomic.AddInt64(&processedUsers, 1)
		if completed%userGreetingProgressInterval == 0 || completed == int64(len(userIDs)) {
			log.Printf("User greeting discovery progress: %d/%d users processed", completed, len(userIDs))
		}
	}

	if collectErr != nil {
		return nil, lastResp, collectErr
	}

	return &allGreetings, lastResp, nil
}

func getAllUsersForGreetingExport(ctx context.Context, clientConfig *platformclientv2.Configuration) (*[]platformclientv2.User, *platformclientv2.APIResponse, error) {
	userProxy := user.GetUserProxy(clientConfig)
	return userProxy.GetAllUser(ctx)
}

func userGreetingFetchConcurrency() int {
	concurrency := defaultUserGreetingConcurrency
	if provider.SdkClientPool != nil {
		if poolSize := provider.SdkClientPool.GetMaxClients(); poolSize > 0 {
			concurrency = poolSize
		}
	}
	if concurrency > maxUserGreetingConcurrency {
		return maxUserGreetingConcurrency
	}
	return concurrency
}

func greetingsApiForWorker(ctx context.Context, baseConfig *platformclientv2.Configuration) (*platformclientv2.GreetingsApi, func(), error) {
	if provider.SdkClientPool != nil {
		clientConfig, err := provider.SdkClientPool.Acquire(ctx)
		if err != nil {
			return nil, func() {}, err
		}
		return platformclientv2.NewGreetingsApiWithConfig(clientConfig), func() {
			if releaseErr := provider.SdkClientPool.Release(clientConfig); releaseErr != nil {
				log.Printf("failed to release SDK client after user greeting fetch: %v", releaseErr)
			}
		}, nil
	}
	return platformclientv2.NewGreetingsApiWithConfig(baseConfig), func() {}, nil
}

func collectUserGreetingsForUser(ctx context.Context, p *greetingProxy, greetingsApi *platformclientv2.GreetingsApi, userID string) ([]platformclientv2.Domainentity, *platformclientv2.APIResponse, error) {
	var collected []platformclientv2.Domainentity

	userGreetings, resp, err := greetingsApi.GetUserGreetings(userID, userGreetingsPageSize, 1)
	if err != nil {
		if isGreetingsPermissionDenied(resp) {
			log.Printf("Skipping greetings for user %s: permission denied (403)", userID)
			return collected, resp, nil
		}
		return nil, resp, fmt.Errorf("failed to get greetings for user %s: %w", userID, err)
	}

	collected = appendGreetingPage(p, greetingsApi, collected, userGreetings.Entities)

	pageCount := 1
	if userGreetings != nil && userGreetings.PageCount != nil {
		pageCount = *userGreetings.PageCount
	}
	for pageNum := 2; pageNum <= pageCount; pageNum++ {
		if err := ctx.Err(); err != nil {
			return nil, resp, err
		}

		userGreetings, resp, err = greetingsApi.GetUserGreetings(userID, userGreetingsPageSize, pageNum)
		if err != nil {
			if isGreetingsPermissionDenied(resp) {
				log.Printf("Skipping greetings for user %s page %d: permission denied (403)", userID, pageNum)
				return collected, resp, nil
			}
			return nil, resp, fmt.Errorf("failed to get greetings for user %s: %w", userID, err)
		}
		collected = appendGreetingPage(p, greetingsApi, collected, userGreetings.Entities)
	}

	return collected, resp, nil
}

func isGreetingsPermissionDenied(resp *platformclientv2.APIResponse) bool {
	return resp != nil && resp.StatusCode == http.StatusForbidden
}

func appendGreetingPage(p *greetingProxy, greetingsApi *platformclientv2.GreetingsApi, collected []platformclientv2.Domainentity, entities *[]platformclientv2.Domainentity) []platformclientv2.Domainentity {
	if entities == nil {
		return collected
	}
	collected = append(collected, *entities...)
	warmGreetingExportCacheFromEntities(p, greetingsApi, entities)
	return collected
}

func warmGreetingExportCacheFromEntities(p *greetingProxy, greetingsApi *platformclientv2.GreetingsApi, entities *[]platformclientv2.Domainentity) {
	if !tfexporter_state.IsExporterActive() {
		return
	}
	for _, entity := range *entities {
		if entity.Id == nil {
			continue
		}
		g, _, gErr := greetingsApi.GetGreeting(*entity.Id)
		if gErr != nil {
			log.Printf("failed to cache greeting %s: %v", *entity.Id, gErr)
			continue
		}
		if g != nil && g.Id != nil {
			rc.SetCache(p.greetingCache, *g.Id, *g)
		}
	}
}

func createUserGreetingFn(ctx context.Context, p *greetingProxy, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	g, resp, err := p.greetingsApi.PostUserGreetings(*body.Owner.Id, *body)
	if err != nil {
		return g, resp, err
	}
	if g != nil && g.Id != nil {
		rc.SetCache(p.greetingCache, *g.Id, *g)
	}
	return g, resp, nil
}
func getUserGreetingByIdFn(ctx context.Context, p *greetingProxy, userId string, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	if tfexporter_state.IsExporterActive() {
		log.Printf("Could not read greeting '%s' from cache. Reading from the API...", id)
		g, resp, err := p.greetingsApi.GetGreeting(id)
		if err != nil {
			return g, resp, err
		}
		if g != nil && g.Id != nil {
			rc.SetCache(p.greetingCache, *g.Id, *g)
		}
		return g, resp, nil
	}
	if userId == "" {
		return p.greetingsApi.GetGreeting(id)
	}
	return getGreetingFromUser(ctx, p, userId, id)
}

func getGreetingFromUser(ctx context.Context, p *greetingProxy, userId string, id string) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	userGreetings, resp, err := p.greetingsApi.GetUserGreetings(userId, pageSize, 1)
	if err != nil {
		return nil, resp, err
	}
	if greeting := findGreetingInEntities(userGreetings.Entities, userId, id); greeting != nil {
		return greeting, resp, nil
	}

	pageCount := 1
	if userGreetings != nil && userGreetings.PageCount != nil {
		pageCount = *userGreetings.PageCount
	}
	for pageNum := 2; pageNum <= pageCount; pageNum++ {
		userGreetings, resp, err = p.greetingsApi.GetUserGreetings(userId, pageSize, pageNum)
		if err != nil {
			return nil, resp, err
		}
		if greeting := findGreetingInEntities(userGreetings.Entities, userId, id); greeting != nil {
			return greeting, resp, nil
		}
	}

	return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("greeting %s not found for user %s", id, userId)
}

func findGreetingInEntities(entities *[]platformclientv2.Domainentity, userId string, greetingId string) *platformclientv2.Greeting {
	if entities == nil {
		return nil
	}
	for _, entity := range *entities {
		if entity.Id != nil && *entity.Id == greetingId {
			return &platformclientv2.Greeting{
				Id:        entity.Id,
				Name:      entity.Name,
				VarType:   platformclientv2.String("NAME"),
				OwnerType: platformclientv2.String("USER"),
				Owner: &platformclientv2.Domainentity{
					Id: platformclientv2.String(userId),
				},
			}
		}
	}
	return nil
}
func updateUserGreetingFn(ctx context.Context, p *greetingProxy, greetingId string, body *platformclientv2.Greeting) (*platformclientv2.Greeting, *platformclientv2.APIResponse, error) {
	g, resp, err := p.greetingsApi.PutGreeting(greetingId, *body)
	if err != nil {
		return g, resp, err
	}
	if g != nil && g.Id != nil {
		rc.SetCache(p.greetingCache, *g.Id, *g)
	}
	return g, resp, nil
}
func deleteUserGreetingFn(ctx context.Context, p *greetingProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.greetingsApi.DeleteGreeting(id)
}
