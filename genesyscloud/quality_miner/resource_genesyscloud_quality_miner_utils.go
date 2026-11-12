package quality_miner

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

func getQualityMinerFromResourceData(d *schema.ResourceData) platformclientv2.Miner {
	return platformclientv2.Miner{
		Name:      resourcedata.GetNillableValue[string](d, "name"),
		MinerType: resourcedata.GetNillableValue[string](d, "miner_type"),
		Language:  resourcedata.GetNillableValue[string](d, "language"),
	}
}

func setQualityMinerToResourceData(d *schema.ResourceData, miner *platformclientv2.Miner) {
	resourcedata.SetNillableValue(d, "name", miner.Name)
	resourcedata.SetNillableValue(d, "miner_type", miner.MinerType)
	resourcedata.SetNillableValue(d, "language", miner.Language)
}
