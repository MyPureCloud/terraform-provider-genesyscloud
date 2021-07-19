package genesyscloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceRoutingEmailDomain(t *testing.T) {
	var (
		emailDomainResourceId   = "email_domain_test"
		emailDomainId  			= "jcc4234234.thoughtx.com"
		emailDataResourceId 	= "email_domain_data"
	)

	fmt.Printf(generateRoutingEmailDomainResource(
		emailDomainResourceId,
		emailDomainId,
		falseValue,
		nullValue,
	) )
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateRoutingEmailDomainResource(
					emailDomainResourceId,
					emailDomainId,
					falseValue,
					nullValue,
				) + generateRoutingEmailDomainDataSource(emailDataResourceId, "genesyscloud_routing_email_domain."+emailDomainResourceId+".domain_id", "genesyscloud_routing_email_domain."+emailDomainResourceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_email_domain."+emailDataResourceId , "id", "genesyscloud_routing_email_domain."+emailDomainResourceId, "id"),
				),
			},
		},
	})
}

//Generates the data source string that will be used in doiung the lookuo
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