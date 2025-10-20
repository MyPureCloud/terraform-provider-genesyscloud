package routing_email_domain

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRoutingEmailDomain(t *testing.T) {
	var (
		emailDomainResourceLabel = "email_domain_test"
		emailDomainId            = fmt.Sprintf("terraformdomain.%s.com", strings.Replace(uuid.NewString(), "-", "", -1))
		emailDataResourceLabel   = "email_domain_data"
	)

	if cleanupErr := CleanupRoutingEmailDomains("terraformdomain"); cleanupErr != nil {
		t.Logf("Failed to clean up routing email domains: %v", cleanupErr)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateRoutingEmailDomainResource(
					emailDomainResourceLabel,
					emailDomainId,
					util.FalseValue,
					util.NullValue,
				) + generateRoutingEmailDomainDataSource(emailDataResourceLabel, "genesyscloud_routing_email_domain."+emailDomainResourceLabel+".domain_id", "genesyscloud_routing_email_domain."+emailDomainResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_email_domain."+emailDataResourceLabel, "id", "genesyscloud_routing_email_domain."+emailDomainResourceLabel, "id"),
				),
			},
		},
	})
}

// Generates the data source string that will be used in doiung the lookuo
func generateRoutingEmailDomainDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_email_domain" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
