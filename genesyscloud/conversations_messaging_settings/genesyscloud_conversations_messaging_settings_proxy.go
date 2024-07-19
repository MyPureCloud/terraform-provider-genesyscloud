package conversations_messaging_settings

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *conversationsMessagingSettingsProxy

type getAllConversationsMessagingSettingsFunc func(ctx context.Context, p *conversationsMessagingSettingsProxy) (*[]platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error)
type createConversationsMessagingSettingsFunc func(ctx context.Context, p *conversationsMessagingSettingsProxy, messagingSettingRequest *platformclientv2.Messagingsettingrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error)
type getConversationsMessagingSettingsByIdFunc func(ctx context.Context, p *conversationsMessagingSettingsProxy, id string) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error)
type getConversationsMessagingSettingsIdByNameFunc func(ctx context.Context, p *conversationsMessagingSettingsProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type updateConversationsMessagingSettingsFunc func(ctx context.Context, p *conversationsMessagingSettingsProxy, id string, messagingSettingRequest *platformclientv2.Messagingsettingpatchrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error)
type deleteConversationsMessagingSettingsFunc func(ctx context.Context, p *conversationsMessagingSettingsProxy, id string) (*platformclientv2.APIResponse, error)

// conversationsMessagingSettingsProxy contains all of the methods that call genesys cloud APIs.
type conversationsMessagingSettingsProxy struct {
	clientConfig                                  *platformclientv2.Configuration
	conversationsApi                              *platformclientv2.ConversationsApi
	createConversationsMessagingSettingsAttr      createConversationsMessagingSettingsFunc
	getAllConversationsMessagingSettingsAttr      getAllConversationsMessagingSettingsFunc
	getConversationsMessagingSettingsIdByNameAttr getConversationsMessagingSettingsIdByNameFunc
	getConversationsMessagingSettingsByIdAttr     getConversationsMessagingSettingsByIdFunc
	updateConversationsMessagingSettingsAttr      updateConversationsMessagingSettingsFunc
	deleteConversationsMessagingSettingsAttr      deleteConversationsMessagingSettingsFunc
	messagingSettingsCache                        rc.CacheInterface[platformclientv2.Messagingsetting]
}

// newConversationsMessagingSettingsProxy initializes the conversations messaging settings proxy with all of the data needed to communicate with Genesys Cloud
func newConversationsMessagingSettingsProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingSettingsProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	messagingSettingsCache := rc.NewResourceCache[platformclientv2.Messagingsetting]()
	return &conversationsMessagingSettingsProxy{
		clientConfig:                                  clientConfig,
		conversationsApi:                              api,
		createConversationsMessagingSettingsAttr:      createConversationsMessagingSettingsFn,
		getAllConversationsMessagingSettingsAttr:      getAllConversationsMessagingSettingsFn,
		getConversationsMessagingSettingsIdByNameAttr: getConversationsMessagingSettingsIdByNameFn,
		getConversationsMessagingSettingsByIdAttr:     getConversationsMessagingSettingsByIdFn,
		updateConversationsMessagingSettingsAttr:      updateConversationsMessagingSettingsFn,
		deleteConversationsMessagingSettingsAttr:      deleteConversationsMessagingSettingsFn,
		messagingSettingsCache:                        messagingSettingsCache,
	}
}

// getConversationsMessagingSettingsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getConversationsMessagingSettingsProxy(clientConfig *platformclientv2.Configuration) *conversationsMessagingSettingsProxy {
	if internalProxy == nil {
		internalProxy = newConversationsMessagingSettingsProxy(clientConfig)
	}
	return internalProxy
}

// getConversationsMessagingSettings retrieves all Genesys Cloud conversations messaging settings
func (p *conversationsMessagingSettingsProxy) getAllConversationsMessagingSettings(ctx context.Context) (*[]platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.getAllConversationsMessagingSettingsAttr(ctx, p)
}

// createConversationsMessagingSettings creates a Genesys Cloud conversations messaging settings
func (p *conversationsMessagingSettingsProxy) createConversationsMessagingSettings(ctx context.Context, conversationsMessagingSettings *platformclientv2.Messagingsettingrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.createConversationsMessagingSettingsAttr(ctx, p, conversationsMessagingSettings)
}

// getConversationsMessagingSettingsById returns a single Genesys Cloud conversations messaging settings by Id
func (p *conversationsMessagingSettingsProxy) getConversationsMessagingSettingsById(ctx context.Context, id string) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.getConversationsMessagingSettingsByIdAttr(ctx, p, id)
}

// getConversationsMessagingSettingsIdByName returns a single Genesys Cloud conversations messaging settings by a name
func (p *conversationsMessagingSettingsProxy) getConversationsMessagingSettingsIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getConversationsMessagingSettingsIdByNameAttr(ctx, p, name)
}

// updateConversationsMessagingSettings updates a Genesys Cloud conversations messaging settings
func (p *conversationsMessagingSettingsProxy) updateConversationsMessagingSettings(ctx context.Context, id string, conversationsMessagingSettings *platformclientv2.Messagingsettingpatchrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.updateConversationsMessagingSettingsAttr(ctx, p, id, conversationsMessagingSettings)
}

// deleteConversationsMessagingSettings deletes a Genesys Cloud conversations messaging settings by Id
func (p *conversationsMessagingSettingsProxy) deleteConversationsMessagingSettings(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteConversationsMessagingSettingsAttr(ctx, p, id)
}

// getAllConversationsMessagingSettingsFn is the implementation for retrieving all conversations messaging settings in Genesys Cloud
func getAllConversationsMessagingSettingsFn(ctx context.Context, p *conversationsMessagingSettingsProxy) (*[]platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	var (
		allMessagingSettings []platformclientv2.Messagingsetting
		pageSize             = 100
		response             *platformclientv2.APIResponse
	)

	messagingSettings, resp, err := p.conversationsApi.GetConversationsMessagingSettings(pageSize, 1)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get messaging setting request: %v", err)
	}

	if messagingSettings.Entities == nil || len(*messagingSettings.Entities) == 0 {
		return &allMessagingSettings, resp, nil
	}
	allMessagingSettings = append(allMessagingSettings, *messagingSettings.Entities...)

	for pageNum := 2; pageNum <= *messagingSettings.PageCount; pageNum++ {
		messagingSettings, resp, err := p.conversationsApi.GetConversationsMessagingSettings(pageSize, pageNum)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get messaging setting request: %v", err)
		}
		response = resp

		if messagingSettings.Entities == nil || len(*messagingSettings.Entities) == 0 {
			break
		}
		allMessagingSettings = append(allMessagingSettings, *messagingSettings.Entities...)
	}

	for _, setting := range allMessagingSettings {
		rc.SetCache(p.messagingSettingsCache, *setting.Id, setting)
	}

	return &allMessagingSettings, response, nil
}

// createConversationsMessagingSettingsFn is an implementation function for creating a Genesys Cloud conversations messaging settings
func createConversationsMessagingSettingsFn(ctx context.Context, p *conversationsMessagingSettingsProxy, conversationsMessagingSettings *platformclientv2.Messagingsettingrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PostConversationsMessagingSettings(*conversationsMessagingSettings)
}

// getConversationsMessagingSettingsByIdFn is an implementation of the function to get a Genesys Cloud conversations messaging settings by Id
func getConversationsMessagingSettingsByIdFn(ctx context.Context, p *conversationsMessagingSettingsProxy, id string) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	if setting := rc.GetCacheItem(p.messagingSettingsCache, id); setting != nil {
		return setting, nil, nil
	}
	return p.conversationsApi.GetConversationsMessagingSetting(id)
}

// getConversationsMessagingSettingsIdByNameFn is an implementation of the function to get a Genesys Cloud conversations messaging settings by name
func getConversationsMessagingSettingsIdByNameFn(ctx context.Context, p *conversationsMessagingSettingsProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	messagingSettings, resp, err := getAllConversationsMessagingSettingsFn(ctx, p)
	if err != nil {
		return "", resp, false, err
	}

	if messagingSettings == nil || len(*messagingSettings) == 0 {
		return "", resp, true, fmt.Errorf("no conversations messaging settings found with name %s", name)
	}

	for _, messagingSetting := range *messagingSettings {
		if *messagingSetting.Name == name {
			log.Printf("Retrieved the conversations messaging settings id %s by name %s", *messagingSetting.Id, name)
			return *messagingSetting.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("unable to find conversations messaging settings with name %s", name)
}

// updateConversationsMessagingSettingsFn is an implementation of the function to update a Genesys Cloud conversations messaging settings
func updateConversationsMessagingSettingsFn(ctx context.Context, p *conversationsMessagingSettingsProxy, id string, messagingSettingRequest *platformclientv2.Messagingsettingpatchrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PatchConversationsMessagingSetting(id, *messagingSettingRequest)
}

// deleteConversationsMessagingSettingsFn is an implementation function for deleting a Genesys Cloud conversations messaging settings
func deleteConversationsMessagingSettingsFn(ctx context.Context, p *conversationsMessagingSettingsProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.conversationsApi.DeleteConversationsMessagingSetting(id)
}
