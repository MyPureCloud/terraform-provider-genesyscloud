package outbound_campaign

import (
	"log"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

const staleTestCampaignAge = 6 * time.Hour

var (
	generatedTestCampaignName = regexp.MustCompile(`^Test Campaign( - )?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	staleCampaignCleanupOnce  sync.Once
)

// cleanupStaleTestCampaignsOnce turns off and deletes campaigns left behind by earlier runs.
// Active campaigns hold outbound lines, and enough of them exhaust the edge group quota so
// later tests fail with RESOURCES_IN_USE. Only campaigns with a generated test name that are
// older than staleTestCampaignAge are touched, so a running test is never affected.
func cleanupStaleTestCampaignsOnce(t *testing.T) {
	staleCampaignCleanupOnce.Do(func() {
		config, err := provider.AuthorizeSdk()
		if err != nil {
			t.Logf("skipping stale campaign cleanup, SDK authorization failed: %v", err)
			return
		}
		api := platformclientv2.NewOutboundApiWithConfig(config)
		cutoff := time.Now().Add(-staleTestCampaignAge)

		var stale []platformclientv2.Campaign
		for page := 1; ; page++ {
			campaigns, _, err := api.GetOutboundCampaigns(100, page, "", "Test Campaign", nil, "", "", "", "", "", nil, "", "")
			if err != nil {
				t.Logf("stale campaign cleanup stopped, failed to list campaigns: %v", err)
				break
			}
			if campaigns.Entities == nil || len(*campaigns.Entities) == 0 {
				break
			}
			for _, campaign := range *campaigns.Entities {
				if campaign.Id != nil && campaign.Name != nil && campaign.DateCreated != nil &&
					generatedTestCampaignName.MatchString(*campaign.Name) && campaign.DateCreated.Before(cutoff) {
					stale = append(stale, campaign)
				}
			}
			if campaigns.PageCount == nil || page >= *campaigns.PageCount {
				break
			}
		}

		for _, campaign := range stale {
			if campaign.CampaignStatus != nil && *campaign.CampaignStatus == "on" {
				off := "off"
				campaign.CampaignStatus = &off
				if _, _, err := api.PutOutboundCampaign(*campaign.Id, campaign, false); err != nil {
					log.Printf("failed to turn off stale campaign %s (%s): %v", *campaign.Name, *campaign.Id, err)
					continue
				}
			}
			if _, _, err := api.DeleteOutboundCampaign(*campaign.Id); err != nil {
				log.Printf("failed to delete stale campaign %s (%s): %v", *campaign.Name, *campaign.Id, err)
				continue
			}
			log.Printf("deleted stale test campaign %s (%s)", *campaign.Name, *campaign.Id)
		}
	})
}
