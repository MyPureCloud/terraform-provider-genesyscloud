package external_contacts_organization

import (
	"context"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// TestAccExternalContactsOrganizationScanEndpoints validates DEVTOOLING-1826:
// division-aware orgs require GET /scan/organizations/divisionviews/all for discovery.
func TestAccExternalContactsOrganizationScanEndpoints(t *testing.T) {
	util.TestAccPreCheck(t)

	config, err := provider.AuthorizeSdk()
	if err != nil {
		t.Fatalf("failed to authorize SDK client: %v", err)
	}
	api := platformclientv2.NewExternalContactsApiWithConfig(config)
	ctx := context.Background()
	proxy := getExternalContactsOrganizationProxy(config)

	legacyCount, err := countScanOrganizations(api, false)
	if err != nil {
		t.Fatalf("legacy scan endpoint failed: %v", err)
	}

	divisionViewsCount, err := countScanOrganizations(api, true)
	if err != nil {
		t.Fatalf("divisionviews scan endpoint failed: %v", err)
	}

	orgs, _, err := proxy.getAllExternalContactsOrganization(ctx)
	if err != nil {
		t.Fatalf("proxy getAllExternalContactsOrganization failed: %v", err)
	}
	proxyCount := 0
	if orgs != nil {
		proxyCount = len(*orgs)
	}

	t.Logf("GET /scan/organizations count: %d", legacyCount)
	t.Logf("GET /scan/organizations/divisionviews/all count: %d", divisionViewsCount)
	t.Logf("proxy getAllExternalContactsOrganization count: %d", proxyCount)

	if proxyCount != divisionViewsCount {
		t.Fatalf("proxy count (%d) must match divisionviews endpoint count (%d)", proxyCount, divisionViewsCount)
	}

	if divisionViewsCount == 0 && legacyCount == 0 {
		t.Skip("no external organizations in this org; create at least one to fully validate export")
	}

	if divisionViewsCount > 0 && legacyCount == 0 {
		t.Log("org appears division-aware: legacy scan empty, divisionviews scan returned organizations")
	}
}

func countScanOrganizations(api *platformclientv2.ExternalContactsApi, divisionViews bool) (int, error) {
	const pageSize = 200
	total := 0
	cursor := ""

	for {
		var listing *platformclientv2.Cursororganizationlisting
		var err error

		if divisionViews {
			listing, _, err = api.GetExternalcontactsScanOrganizationsDivisionviewsAll(pageSize, cursor)
		} else {
			listing, _, err = api.GetExternalcontactsScanOrganizations(pageSize, cursor, "")
		}
		if err != nil {
			return total, err
		}
		if listing == nil || listing.Entities == nil || len(*listing.Entities) == 0 {
			break
		}
		total += len(*listing.Entities)
		if listing.Cursors == nil || listing.Cursors.After == nil || *listing.Cursors.After == "" {
			break
		}
		cursor = *listing.Cursors.After
	}

	return total, nil
}
