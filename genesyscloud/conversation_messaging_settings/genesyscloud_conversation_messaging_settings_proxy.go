package conversations_messaging_settings

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *conversationMessagingSettingsProxy

type getAllConversationMessagingSettingsFunc func(ctx context.Context, p *conversationMessagingSettingsProxy) (*[]platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error)
type createConversationMessagingSettingsFunc func(ctx context.Context, p *conversationMessagingSettingsProxy, messagingSettingRequest *platformclientv2.Messagingsettingrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error)
type getConversationMessagingSettingsByIdFunc func(ctx context.Context, p *conversationMessagingSettingsProxy, id string) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error)
type getConversationMessagingSettingsIdByNameFunc func(ctx context.Context, p *conversationMessagingSettingsProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type updateConversationMessagingSettingsFunc func(ctx context.Context, p *conversationMessagingSettingsProxy, id string, messagingSettingRequest *platformclientv2.Messagingsettingpatchrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error)
type deleteConversationMessagingSettingsFunc func(ctx context.Context, p *conversationMessagingSettingsProxy, id string) (*platformclientv2.APIResponse, error)

// conversationMessagingSettingsProxy contains all of the methods that call genesys cloud APIs.
type conversationMessagingSettingsProxy struct {
	clientConfig                                 *platformclientv2.Configuration
	conversationsApi                             *platformclientv2.ConversationsApi
	createConversationMessagingSettingsAttr      createConversationMessagingSettingsFunc
	getAllConversationMessagingSettingsAttr      getAllConversationMessagingSettingsFunc
	getConversationMessagingSettingsIdByNameAttr getConversationMessagingSettingsIdByNameFunc
	getConversationMessagingSettingsByIdAttr     getConversationMessagingSettingsByIdFunc
	updateConversationMessagingSettingsAttr      updateConversationMessagingSettingsFunc
	deleteConversationMessagingSettingsAttr      deleteConversationMessagingSettingsFunc
	messagingSettingsCache                       rc.CacheInterface[platformclientv2.Messagingsetting]
}

// newConversationMessagingSettingsProxy initializes the conversation messaging settings proxy with all of the data needed to communicate with Genesys Cloud
func newConversationMessagingSettingsProxy(clientConfig *platformclientv2.Configuration) *conversationMessagingSettingsProxy {
	api := platformclientv2.NewConversationsApiWithConfig(clientConfig)
	messagingSettingsCache := rc.NewResourceCache[platformclientv2.Messagingsetting]()
	return &conversationMessagingSettingsProxy{
		clientConfig:                                 clientConfig,
		conversationsApi:                             api,
		createConversationMessagingSettingsAttr:      createConversationMessagingSettingsFn,
		getAllConversationMessagingSettingsAttr:      getAllConversationMessagingSettingsFn,
		getConversationMessagingSettingsIdByNameAttr: getConversationMessagingSettingsIdByNameFn,
		getConversationMessagingSettingsByIdAttr:     getConversationMessagingSettingsByIdFn,
		updateConversationMessagingSettingsAttr:      updateConversationMessagingSettingsFn,
		deleteConversationMessagingSettingsAttr:      deleteConversationMessagingSettingsFn,
		messagingSettingsCache:                       messagingSettingsCache,
	}
}

// getConversationMessagingSettingsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getConversationMessagingSettingsProxy(clientConfig *platformclientv2.Configuration) *conversationMessagingSettingsProxy {
	if internalProxy == nil {
		internalProxy = newConversationMessagingSettingsProxy(clientConfig)
	}
	return internalProxy
}

// getConversationMessagingSettings retrieves all Genesys Cloud conversation messaging settings
func (p *conversationMessagingSettingsProxy) getAllConversationMessagingSettings(ctx context.Context) (*[]platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.getAllConversationMessagingSettingsAttr(ctx, p)
}

// createConversationMessagingSettings creates a Genesys Cloud conversation messaging settings
func (p *conversationMessagingSettingsProxy) createConversationMessagingSettings(ctx context.Context, conversationMessagingSettings *platformclientv2.Messagingsettingrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.createConversationMessagingSettingsAttr(ctx, p, conversationMessagingSettings)
}

// getConversationMessagingSettingsById returns a single Genesys Cloud conversation messaging settings by Id
func (p *conversationMessagingSettingsProxy) getConversationMessagingSettingsById(ctx context.Context, id string) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.getConversationMessagingSettingsByIdAttr(ctx, p, id)
}

// getConversationMessagingSettingsIdByName returns a single Genesys Cloud conversation messaging settings by a name
func (p *conversationMessagingSettingsProxy) getConversationMessagingSettingsIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getConversationMessagingSettingsIdByNameAttr(ctx, p, name)
}

// updateConversationMessagingSettings updates a Genesys Cloud conversation messaging settings
func (p *conversationMessagingSettingsProxy) updateConversationMessagingSettings(ctx context.Context, id string, conversationMessagingSettings *platformclientv2.Messagingsettingpatchrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.updateConversationMessagingSettingsAttr(ctx, p, id, conversationMessagingSettings)
}

// deleteConversationMessagingSettings deletes a Genesys Cloud conversation messaging settings by Id
func (p *conversationMessagingSettingsProxy) deleteConversationMessagingSettings(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteConversationMessagingSettingsAttr(ctx, p, id)
}

// getAllConversationMessagingSettingsFn is the implementation for retrieving all conversation messaging settings in Genesys Cloud
func getAllConversationMessagingSettingsFn(ctx context.Context, p *conversationMessagingSettingsProxy) (*[]platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
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

// createConversationMessagingSettingsFn is an implementation function for creating a Genesys Cloud conversation messaging settings
func createConversationMessagingSettingsFn(ctx context.Context, p *conversationMessagingSettingsProxy, conversationMessagingSettings *platformclientv2.Messagingsettingrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PostConversationsMessagingSettings(*conversationMessagingSettings)
}

// getConversationMessagingSettingsByIdFn is an implementation of the function to get a Genesys Cloud conversation messaging settings by Id
func getConversationMessagingSettingsByIdFn(ctx context.Context, p *conversationMessagingSettingsProxy, id string) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	if setting := rc.GetCacheItem(p.messagingSettingsCache, id); setting != nil {
		return setting, nil, nil
	}
	return p.conversationsApi.GetConversationsMessagingSetting(id)
}

// getConversationMessagingSettingsIdByNameFn is an implementation of the function to get a Genesys Cloud conversation messaging settings by name
func getConversationMessagingSettingsIdByNameFn(ctx context.Context, p *conversationMessagingSettingsProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	messagingSettings, resp, err := getAllConversationMessagingSettingsFn(ctx, p)
	if err != nil {
		return "", resp, false, err
	}

	if messagingSettings == nil || len(*messagingSettings) == 0 {
		return "", resp, true, fmt.Errorf("no conversation messaging settings found with name %s", name)
	}

	for _, messagingSetting := range *messagingSettings {
		if *messagingSetting.Name == name {
			log.Printf("Retrieved the conversation messaging settings id %s by name %s", *messagingSetting.Id, name)
			return *messagingSetting.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("unable to find conversation messaging settings with name %s", name)
}

// updateConversationMessagingSettingsFn is an implementation of the function to update a Genesys Cloud conversation messaging settings
func updateConversationMessagingSettingsFn(ctx context.Context, p *conversationMessagingSettingsProxy, id string, messagingSettingRequest *platformclientv2.Messagingsettingpatchrequest) (*platformclientv2.Messagingsetting, *platformclientv2.APIResponse, error) {
	return p.conversationsApi.PatchConversationsMessagingSetting(id, *messagingSettingRequest)
}

// deleteConversationMessagingSettingsFn is an implementation function for deleting a Genesys Cloud conversation messaging settings
func deleteConversationMessagingSettingsFn(ctx context.Context, p *conversationMessagingSettingsProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.conversationsApi.DeleteConversationsMessagingSetting(id)
}
