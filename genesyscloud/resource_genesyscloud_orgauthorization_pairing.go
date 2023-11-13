package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func resourceOrgauthorizationPairing() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud orgauthorization pairing`,

		CreateContext: CreateWithPooledClient(createOrgauthorizationPairing),
		ReadContext:   ReadWithPooledClient(readOrgauthorizationPairing),
		DeleteContext: DeleteWithPooledClient(deleteOrgauthorizationPairing),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`user_ids`: {
				Description: `The list of trustee users that are requesting access. If no users are specified, at least one group is required.  Changing the user_ids attribute will cause the orgauthorization_pairing resource to be dropped and recreated with a new ID.`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`group_ids`: {
				Description: `The list of trustee groups that are requesting access. If no groups are specified, at least one user is required. Changing the group_ids attribute will cause the orgauthorization_pairing resource to be dropped and recreated with a new ID.`,
				Optional:    true,
				ForceNew:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func deleteOrgauthorizationPairing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func createOrgauthorizationPairing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	userIds := lists.InterfaceListToStrings(d.Get("user_ids").([]interface{}))
	groupIds := lists.InterfaceListToStrings(d.Get("group_ids").([]interface{}))

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	organizationAuthorizationApi := platformclientv2.NewOrganizationAuthorizationApiWithConfig(sdkConfig)

	sdktrustrequestcreate := platformclientv2.Trustrequestcreate{
		UserIds:  &userIds,
		GroupIds: &groupIds,
	}

	log.Printf("Creating Orgauthorization Pairing")
	orgauthorizationPairing, _, err := organizationAuthorizationApi.PostOrgauthorizationPairings(sdktrustrequestcreate)
	if err != nil {
		return diag.Errorf("Failed to create Orgauthorization Pairing: %s", err)
	}

	d.SetId(*orgauthorizationPairing.Id)

	log.Printf("Created Orgauthorization Pairing %s", *orgauthorizationPairing.Id)
	return readOrgauthorizationPairing(ctx, d, meta)
}

func readOrgauthorizationPairing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	organizationAuthorizationApi := platformclientv2.NewOrganizationAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Reading Orgauthorization Pairing %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdktrustrequest, resp, getErr := organizationAuthorizationApi.GetOrgauthorizationPairing(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Orgauthorization Pairing %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Orgauthorization Pairing %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceOrgauthorizationPairing())

		schemaUserIds := lists.InterfaceListToStrings(d.Get("user_ids").([]interface{}))
		if sdktrustrequest.Users != nil {
			ids := make([]string, 0)
			for _, item := range *sdktrustrequest.Users {
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
		if sdktrustrequest.Groups != nil {
			ids := make([]string, 0)
			for _, item := range *sdktrustrequest.Groups {
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
		return cc.CheckState()
	})
}
