package aistudio_summary_setting

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_aistudio_summary_setting_test.go contains all of the test cases for running the resource
tests for aistudio_summary_setting.
*/

func TestAccResourceAistudioSummarySetting(t *testing.T) {
	t.Parallel()
	var ()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps:             []resource.TestStep{},
		CheckDestroy:      testVerifyAistudioSummarySettingDestroyed,
	})
}

func testVerifyAistudioSummarySettingDestroyed(state *terraform.State) error {
	return nil
}
