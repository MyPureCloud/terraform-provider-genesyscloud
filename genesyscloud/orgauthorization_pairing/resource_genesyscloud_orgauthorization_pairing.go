package orgauthorization_pairing

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func createOrgauthorizationPairing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOrgauthorizationPairingProxy(sdkConfig)

	userIds := lists.InterfaceListToStrings(d.Get("user_ids").([]interface{}))
	groupIds := lists.InterfaceListToStrings(d.Get("group_ids").([]interface{}))
	trustRequestCreate := platformclientv2.Trustrequestcreate{
		UserIds:  &userIds,
		GroupIds: &groupIds,
	}

	log.Printf("Creating Orgauthorization Pairing")

	pairing, resp, err := proxy.createOrgauthorizationPairing(ctx, &trustRequestCreate)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create Orgauthorization Pairing | error: %s", err), resp)
	}

	d.SetId(*pairing.Id)
	log.Printf("Created Orgauthorization Pairing %s", *pairing.Id)
	return readOrgauthorizationPairing(ctx, d, meta)
}

func readOrgauthorizationPairing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOrgauthorizationPairingProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOrgauthorizationPairing(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Orgauthorization Pairing %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		trustRequest, resp, getErr := proxy.getOrgauthorizationPairingById(ctx, d.Id())

		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Orgauthorization Pairing %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Orgauthorization Pairing %s | error: %s", d.Id(), getErr), resp))
		}

		schemaUserIds := lists.InterfaceListToStrings(d.Get("user_ids").([]interface{}))
		if trustRequest.Users != nil {
			ids := make([]string, 0)
			for _, item := range *trustRequest.Users {
				ids = append(ids, *item.Id)
			}
			// if lists are the same: Set in original order to avoid plan not empty error
			if lists.AreEquivalent(schemaUserIds, ids) {
				d.Set("user_ids", schemaUserIds)
			} else {
				d.Set("user_ids", ids)
			}
		}

		schemaGroupIds := lists.InterfaceListToStrings(d.Get("group_ids").([]interface{}))
		if trustRequest.Groups != nil {
			ids := make([]string, 0)
			for _, item := range *trustRequest.Groups {
				ids = append(ids, *item.Id)
			}
			// if lists are the same: Set in original order to avoid plan not empty error
			if lists.AreEquivalent(schemaGroupIds, ids) {
				d.Set("group_ids", schemaGroupIds)
			} else {
				d.Set("group_ids", ids)
			}
		}

		log.Printf("Read Orgauthorization Pairing %s", d.Id())
		return cc.CheckState(d)
	})
}

func deleteOrgauthorizationPairing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
