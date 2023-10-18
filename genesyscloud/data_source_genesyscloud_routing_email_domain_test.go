package genesyscloud

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccDataSourceRoutingEmailDomain(t *testing.T) {
	var (
		emailDomainResourceId = "email_domain_test"
		emailDomainId         = "terraform" + strconv.Itoa(rand.Intn(1000)) + ".com"
		emailDataResourceId   = "email_domain_data"
	)

	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	CleanupRoutingEmailDomains()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateRoutingEmailDomainResource(
					emailDomainResourceId,
					emailDomainId,
					falseValue,
					nullValue,
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
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		routingEmailDomains, _, getErr := routingAPI.GetRoutingEmailDomains(pageNum, pageSize, false, "")
		if getErr != nil {
			return
		}

		if routingEmailDomains.Entities == nil || len(*routingEmailDomains.Entities) == 0 {
			return
		}

		for _, routingEmailDomain := range *routingEmailDomains.Entities {
			if routingEmailDomain.Id != nil && strings.HasPrefix(*routingEmailDomain.Id, "terraform") {
				_, err := routingAPI.DeleteRoutingEmailDomain(*routingEmailDomain.Id)
				if err != nil {
					log.Printf("Failed to delete routing email domain %s: %s", *routingEmailDomain.Id, err)
					continue
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
}
