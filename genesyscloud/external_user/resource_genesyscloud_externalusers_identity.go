package external_user

import (
	"context"
	"fmt"
	"log"
	"time"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	userResource "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func getAllExternalUserIdentity(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {

	externalUserproxy := getExternalUserIdentityProxy(clientConfig)
	userProxy := userResource.GetUserProxy(clientConfig)

	userList, userResponse, err := userProxy.GetAllUser(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get genesys users error: %s", err), userResponse)
	}
	resources, response, err := getAllHelperExternalUser(ctx, externalUserproxy, userList)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get external users error: %s", err), response)
	}
	return resources, nil
}

func createExternalUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalUserIdentityProxy(sdkConfig)
	var externalUserObject platformclientv2.Userexternalidentifier

	userId := d.Get("user_id").(string)
	authorityName := d.Get("authority_name").(string)
	externalKey := d.Get("external_key").(string)

	externalUserObject.AuthorityName = &authorityName
	externalUserObject.ExternalKey = &externalKey

	log.Printf("Creating external user for genesys user %s", userId)
	externalUser, response, err := proxy.createExternalUserIdentity(ctx, userId, externalUserObject)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get external user %s", err), response)
	}
	id := createCompoundKey(userId, *externalUser.AuthorityName, *externalUser.ExternalKey)

	d.SetId(id)
	log.Printf("Created external user %s for genesys user %s", id, userId)
	return readExternalUser(ctx, d, meta)
}

func readExternalUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalUserIdentityProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceExternalUserIdentity(), constants.ConsistencyChecks(), ResourceType)
	userId, authorityName, externalKey, err := splitCompoundKey(d.Id())
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to split compound key", err)
	}
	log.Printf("Reading external user %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		externalUser, response, err := proxy.getExternalUserIdentityById(ctx, userId, authorityName, externalKey)
		if err != nil {
			if util.IsStatus404ByInt(response.StatusCode) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read organization contact %s | error: %s", d.Id(), err), response))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read organization contact %s | error: %s", d.Id(), err), response))
		}

		resourcedata.SetNillableValue(d, "user_id", platformclientv2.String(userId))
		resourcedata.SetNillableValue(d, "authority_name", externalUser.AuthorityName)
		resourcedata.SetNillableValue(d, "external_key", externalUser.ExternalKey)

		log.Printf("Read external user %s", d.Id())
		return cc.CheckState(d)
	})
}

func updateExternalUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalUserIdentityProxy(sdkConfig)
	var externalUserObject platformclientv2.Userexternalidentifier

	oldUserId, oldAuthorityName, oldExternalKey, err := splitCompoundKey(d.Id())
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to split compound key", err)
	}
	log.Printf("Update context - deleting external user %s for genesys user %s", d.Id(), oldUserId)
	deleteResponse, deleteErr := proxy.deleteExternalUserIdentity(ctx, oldUserId, oldAuthorityName, oldExternalKey)
	if deleteErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete external user %s: %s", d.Id(), deleteErr), deleteResponse)
	}
	userId := d.Get("user_id").(string)
	authorityName := d.Get("authority_name").(string)
	externalKey := d.Get("external_key").(string)

	externalUserObject.AuthorityName = &authorityName
	externalUserObject.ExternalKey = &externalKey
	log.Printf("updating external user %s for genesys user %s", d.Id(), userId)
	externalUser, response, err := proxy.createExternalUserIdentity(ctx, userId, externalUserObject)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update external user error: %s", err), response)
	}
	id := createCompoundKey(userId, *externalUser.AuthorityName, *externalUser.ExternalKey)

	d.SetId(id)
	log.Printf("updated external user %s for genesys user %s", id, userId)
	return readExternalUser(ctx, d, meta)
}

func deleteExternalUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalUserIdentityProxy(sdkConfig)
	userId, authorityName, externalKey, err := splitCompoundKey(d.Id())
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to split compound key", err)
	}

	response, deleteErr := proxy.deleteExternalUserIdentity(ctx, userId, authorityName, externalKey)
	if deleteErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete external user %s: %s", d.Id(), deleteErr), response)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, response, err := proxy.getExternalUserIdentityById(ctx, userId, authorityName, externalKey)
		if err != nil {
			if util.IsStatus404ByInt(response.StatusCode) {
				log.Printf("Deleted external user %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting external user %s: %s", d.Id(), err), response))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("external user still exists %s:", d.Id()), response))
	})

}

func getAllHelperExternalUser(ctx context.Context, externalUserproxy *externalUserIdentityProxy, userList *[]platformclientv2.User) (resourceExporter.ResourceIDMetaMap, *platformclientv2.APIResponse, error) {
	resources := make(resourceExporter.ResourceIDMetaMap)

	if userList == nil || len(*userList) == 0 {
		return resources, nil, nil
	}

	for _, eachUser := range *userList {
		if eachUser.Id == nil {
			continue
		}
		userId := *eachUser.Id
		externalUseList, response, err := externalUserproxy.getAllExternalUserIdentity(ctx, userId)
		if err != nil || externalUseList == nil {
			return nil, response, err
		}
		for _, externalUser := range *externalUseList {

			id := createCompoundKey(userId, *externalUser.AuthorityName, *externalUser.ExternalKey)
			resources[id] = &resourceExporter.ResourceMeta{BlockLabel: id}
		}

	}

	return resources, nil, nil
}
