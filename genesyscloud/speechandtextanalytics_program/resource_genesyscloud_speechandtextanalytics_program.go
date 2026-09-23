package speechandtextanalytics_program

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const programsPageSize = 100

func createProgram(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSttProgramProxy(sdkConfig)

	req := buildProgramRequest(d)
	log.Printf("Creating Speech & Text Analytics Program %s", d.Get("name").(string))
	program, resp, err := proxy.createProgram(ctx, req)
	if err != nil {
		input, _ := util.InterfaceToJson(*req)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create speech and text analytics program: %s\n(input: %+v)", err, input), resp)
	}

	if program == nil || program.Id == nil {
		return diag.Errorf("API returned success but no program ID for %s", d.Get("name").(string))
	}
	d.SetId(*program.Id)

	log.Printf("Created Speech & Text Analytics Program %s", d.Id())
	return readProgram(ctx, d, meta)
}

func readProgram(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSttProgramProxy(sdkConfig)

	log.Printf("Reading Speech & Text Analytics Program %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		program, resp, err := proxy.getProgram(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read speech and text analytics program %s: %s", d.Id(), err), resp))
		}

		flattenProgramToResourceData(d, program)
		return nil
	})
}

func updateProgram(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSttProgramProxy(sdkConfig)

	req := buildProgramRequest(d)
	log.Printf("Updating Speech & Text Analytics Program %s", d.Id())
	program, resp, err := proxy.updateProgram(ctx, d.Id(), req)
	if err != nil {
		input, _ := util.InterfaceToJson(*req)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update speech and text analytics program %s: %s\n(input: %+v)", d.Id(), err, input), resp)
	}
	if program == nil || program.Id == nil {
		return diag.Errorf("API returned success but no program ID for %s", d.Get("name").(string))
	}
	d.SetId(*program.Id)

	log.Printf("Updated Speech & Text Analytics Program %s", d.Id())
	return readProgram(ctx, d, meta)
}

func deleteProgram(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSttProgramProxy(sdkConfig)

	log.Printf("Deleting Speech & Text Analytics Program %s", d.Id())
	// forceDelete=true so a program that is still referenced does not block destroy.
	resp, err := proxy.deleteProgram(ctx, d.Id(), true)
	if err != nil {
		if util.IsStatus404(resp) {
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete speech and text analytics program %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getProgram(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted speech and text analytics program %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error verifying deletion of program %s: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("speech and text analytics program %s still exists", d.Id()), resp))
	})
}

func getAllPrograms(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getSttProgramProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	nextPage := ""
	for {
		programs, resp, err := proxy.listPrograms(ctx, nextPage, programsPageSize)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of programs: %s", err), resp)
		}
		if programs == nil || programs.Entities == nil || len(*programs.Entities) == 0 {
			break
		}

		for _, program := range *programs.Entities {
			if program.Id == nil {
				continue
			}
			blockLabel := *program.Id
			if program.Name != nil {
				blockLabel = *program.Name
			}
			resources[*program.Id] = &resourceExporter.ResourceMeta{BlockLabel: blockLabel}
		}

		if programs.NextUri == nil || *programs.NextUri == "" {
			break
		}
		previousNextPage := nextPage
		nextPage, err = util.GetQueryParamValueFromUri(*programs.NextUri, "nextPage")
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to parse nextPage cursor from programs nextUri: %s", err), resp)
		}
		if nextPage == "" || nextPage == previousNextPage {
			break
		}
	}

	return resources, nil
}
