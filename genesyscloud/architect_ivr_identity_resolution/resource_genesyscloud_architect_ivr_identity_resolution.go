package architect_ivr_identity_resolution

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	consistencyChecker "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
)

func getAllArchitectIvrIdentityResolution(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getArchitectIvrIdentityResolutionProxy(clientConfig)

	IVRs, resp, err := proxy.architectIvrProxy.GetAllArchitectIvrs(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, "failed to list arhitect IVRs for identity resolution export", resp)
	}
	if IVRs == nil {
		return resources, nil
	}

	for _, ivr := range *IVRs {
		if ivr.Id == nil || ivr.Name == nil {
			continue
		}

		config, getResp, getErr := proxy.getArchitectIvrIdentityResolution(ctx, *ivr.Id)
		if getErr != nil {
			if util.IsStatus404(getResp) {
				continue
			}
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get identity resolution config for IVR %s", *ivr.Id), getResp)
		}

		if isDefaultIdentityResolutionConfig(config) {
			continue
		}

		resources[*ivr.Id] = &resourceExporter.ResourceMeta{BlockLabel: *ivr.Name + "-identity-resolution"}
	}

	return resources, nil
}

func createArchitectIvrIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ivrId := d.Get("ivr_id").(string)
	log.Printf("creating identity resolution for IVR %s", ivrId)
	d.SetId(ivrId)

	return updateArchitectIvrIdentityResolution(ctx, d, meta)
}

func readArchitectIvrIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectIvrIdentityResolutionProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectIvrIdentityResolution(), constants.ConsistencyChecks(), ResourceType)

	ivrId := d.Id()
	log.Printf("reading identity resolution for IVR %s", ivrId)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		config, resp, getErr := proxy.getArchitectIvrIdentityResolution(ctx, ivrId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				log.Printf("IVR %s not found, removing identity resolution from state", ivrId)
				d.SetId("")
				return nil
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read identity resolution for IVR %s: %s", ivrId, getErr), resp))
		}

		_ = d.Set("ivr_id", ivrId)
		_ = d.Set("resolve_identities", config.ResolveIdentities)
		_ = d.Set("division_id", config.Division.Id)

		log.Printf("read identity resolution for IVR %s", ivrId)
		return cc.CheckState(d)
	})
}

func updateArchitectIvrIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectIvrIdentityResolutionProxy(sdkConfig)
	ivrId := d.Id()

	config, err := buildIdentityResolutionIVRConfig(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to build identity resolution config", err)
	}

	log.Printf("updating identity resolution for IVR %s", ivrId)
	_, resp, putErr := proxy.putArchitectIvrIdentityResolution(ctx, ivrId, config)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update identity resolution for IVR %s", ivrId), resp)
	}

	log.Printf("updated identity resolution for IVR %s", ivrId)
	return readArchitectIvrIdentityResolution(ctx, d, meta)
}

func deleteArchitectIvrIdentityResolution(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectIvrIdentityResolutionProxy(sdkConfig)
	ivrId := d.Id()

	log.Printf("resetting identity resolution for IVR %s to default", ivrId)

	_, resp, getErr := proxy.getArchitectIvrById(ctx, ivrId)
	if getErr != nil {
		if util.IsStatus404(resp) {
			log.Printf("parend IVR %s already deleted", ivrId)
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to verify IVR %s before resetting identity resolution", ivrId), resp)
	}

	defaultConfig := buildDefaultIdentityResolutionIVRConfig()
	_, putResp, putErr := proxy.putArchitectIvrIdentityResolution(ctx, ivrId, defaultConfig)
	if putErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to reset identity resolution for IVR %s", ivrId), putResp)
	}

	log.Printf("reset identity resolution for IVR %s to default", ivrId)
	return nil
}
