package routing_email_domain

import (
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccDataSourceRoutingEmailDomain(t *testing.T) {
	var (
		emailDomainResourceId = "email_domain_test"
		emailDomainId         = fmt.Sprintf("terraformdomain.%s.com", strings.Replace(uuid.NewString(), "-", "", -1))
		emailDataResourceId   = "email_domain_data"
	)

	CleanupRoutingEmailDomains()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateRoutingEmailDomainResource(
					emailDomainResourceId,
					emailDomainId,
					util.FalseValue,
					util.NullValue,
				) + generateRoutingEmailDomainDataSource(emailDataResourceId, "genesyscloud_routing_email_domain."+emailDomainResourceId+".domain_id", "genesyscloud_routing_email_domain."+emailDomainResourceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_email_domain."+emailDataResourceId, "id", "genesyscloud_routing_email_domain."+emailDomainResourceId, "id"),
				),
			},
		},
	})
}

// Generates the data source string that will be used in doiung the lookuo
func generateRoutingEmailDomainDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_email_domain" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}

func CleanupRoutingEmailDomains() {
	sdkConfig, _ := provider.AuthorizeSdk()
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		routingEmailDomains, _, getErr := routingAPI.GetRoutingEmailDomains(pageSize, pageNum, false, "")
		if getErr != nil {
			log.Printf("failed to get page %v of routing email domains: %v", pageNum, getErr)
			return
		}

		if routingEmailDomains.Entities == nil || len(*routingEmailDomains.Entities) == 0 {
			break
		}

		for _, routingEmailDomain := range *routingEmailDomains.Entities {
			if routingEmailDomain.Name != nil && strings.HasPrefix(*routingEmailDomain.Name, "terraformdomain") {
				_, err := routingAPI.DeleteRoutingEmailDomain(*routingEmailDomain.Id)
				if err != nil {
					log.Printf("Failed to delete routing email domain %s: %s", *routingEmailDomain.Id, err)
					return
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
}
