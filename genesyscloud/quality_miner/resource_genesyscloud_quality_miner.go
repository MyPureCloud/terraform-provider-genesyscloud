package quality_miner

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

func getAllQualityMiners(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getQualityMinerProxy(clientConfig)

	miners, resp, getErr := proxy.getAllQualityMiners(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all miners: %v", getErr), resp)
	}

	if miners == nil || len(*miners) == 0 {
		return resources, nil
	}

	for _, miner := range *miners {
		if miner.Id == nil {
			continue
		}
		blockLabel := *miner.Id
		if miner.Name != nil {
			blockLabel = *miner.Name
		}
		resources[*miner.Id] = &resourceExporter.ResourceMeta{BlockLabel: blockLabel}
	}
	return resources, nil
}

func createQualityMiner(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getQualityMinerProxy(sdkConfig)
	name := d.Get("name").(string)

	minerRequest := getQualityMinerFromResourceData(d)

	log.Printf("Creating miner %s", name)
	miner, resp, err := proxy.createQualityMiner(ctx, &minerRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create miner %s error: %s", name, err), resp)
	}

	d.SetId(*miner.Id)

	log.Printf("Created miner %s %s", name, *miner.Id)
	return readQualityMiner(ctx, d, meta)
}

func readQualityMiner(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getQualityMinerProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceQualityMiner(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading miner %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		miner, resp, getErr := proxy.getQualityMinerById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read miner %s | error: %s", d.Id(), getErr), resp))
		}

		setQualityMinerToResourceData(d, miner)

		log.Printf("Read miner %s %s", d.Id(), d.Get("name").(string))
		return cc.CheckState(d)
	})
}

func deleteQualityMiner(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getQualityMinerProxy(sdkConfig)
	name := d.Get("name").(string)

	log.Printf("Deleting miner %s", name)
	resp, err := proxy.deleteQualityMiner(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete miner %s error: %s", name, err), resp)
	}

	log.Printf("Deleted miner %s", d.Id())
	return nil
}
