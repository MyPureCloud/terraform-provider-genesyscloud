package analytics_dataretention_settings

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
The resource_genesyscloud_analytics_dataretention_settings_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getAnalyticsDataretentionSettingsFromResourceData maps the Terraform resource data into the request body
// sent to the PUT /api/v2/analytics/dataretention/settings endpoint.
//
// Note: dateCreated and dateModified are returned by the API but are read-only, so they are neither part of
// the schema nor of the update body. retention_days is Required, so d.Get always returns the configured value.
func getAnalyticsDataretentionSettingsFromResourceData(d *schema.ResourceData) platformclientv2.Updateanalyticsdataretentionrequest {
	return platformclientv2.Updateanalyticsdataretentionrequest{
		RetentionDays: platformclientv2.Int(d.Get("retention_days").(int)),
	}
}

// setAnalyticsDataretentionSettingsToResourceData maps the API response onto the Terraform resource data.
// It is shared by both the resource read and the data source read so the mapping stays consistent.
func setAnalyticsDataretentionSettingsToResourceData(d *schema.ResourceData, settings *platformclientv2.Analyticsdataretentionresponse) {
	resourcedata.SetNillableValue(d, "retention_days", settings.RetentionDays)
}
