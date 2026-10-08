package speechandtextanalytics_program

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func dataSourceSpeechAndTextAnalyticsProgramRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSttProgramProxy(sdkConfig)

	nextPage := ""
	for {
		programs, resp, err := proxy.listPrograms(ctx, nextPage, programsPageSize)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to list programs: %s", err), resp)
		}
		if programs == nil || programs.Entities == nil || len(*programs.Entities) == 0 {
			break
		}
		for _, program := range *programs.Entities {
			if program.Name != nil && *program.Name == name && program.Id != nil {
				d.SetId(*program.Id)
				_ = d.Set("name", name)
				return nil
			}
		}

		if programs.NextUri == nil || *programs.NextUri == "" {
			break
		}
		previousNextPage := nextPage
		nextPage, err = util.GetQueryParamValueFromUri(*programs.NextUri, "nextPage")
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to parse nextPage cursor from programs nextUri: %s", err), resp)
		}
		if nextPage == "" || nextPage == previousNextPage {
			break
		}
	}

	return diag.Errorf("No speech and text analytics program found with name '%s'", name)
}
