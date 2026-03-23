package speechandtextanalytics_topic

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func getAllTopics(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getSttTopicProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	const pageSize = 100
	for pageNum := 1; pageNum <= 50; pageNum++ {
		topics, resp, err := proxy.listTopics(ctx, pageSize, pageNum)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of topics: %s", err), resp)
		}
		if topics == nil || topics.Entities == nil || len(*topics.Entities) == 0 {
			break
		}

		for _, t := range *topics.Entities {
			if t.Id == nil {
				continue
			}
			blockLabel := *t.Id
			if t.Name != nil && t.Dialect != nil {
				blockLabel = fmt.Sprintf("%s_%s", *t.Name, *t.Dialect)
			} else if t.Name != nil {
				blockLabel = *t.Name
			}
			resources[*t.Id] = &resourceExporter.ResourceMeta{BlockLabel: blockLabel}
		}

		if topics.PageCount != nil && pageNum >= *topics.PageCount {
			break
		}
	}

	_ = provider.EnsureResourceContext(ctx, ResourceType)
	return resources, nil
}
