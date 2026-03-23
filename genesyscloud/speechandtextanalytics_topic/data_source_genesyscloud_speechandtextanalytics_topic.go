package speechandtextanalytics_topic

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func dataSourceSpeechAndTextAnalyticsTopicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	dialect := d.Get("dialect").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSttTopicProxy(sdkConfig)

	const pageSize = 100
	for pageNum := 1; pageNum <= 50; pageNum++ {
		topics, resp, err := proxy.listTopics(ctx, pageSize, pageNum)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to list topics: %s", err), resp)
		}
		if topics == nil || topics.Entities == nil || len(*topics.Entities) == 0 {
			break
		}
		for _, t := range *topics.Entities {
			if t.Name != nil && t.Dialect != nil && *t.Name == name && *t.Dialect == dialect && t.Id != nil {
				d.SetId(*t.Id)
				_ = d.Set("name", name)
				_ = d.Set("dialect", dialect)
				return nil
			}
		}
		if topics.PageCount != nil && pageNum >= *topics.PageCount {
			break
		}
	}

	return diag.Errorf("No speech and text analytics topic found with name '%s' and dialect '%s'", name, dialect)
}
