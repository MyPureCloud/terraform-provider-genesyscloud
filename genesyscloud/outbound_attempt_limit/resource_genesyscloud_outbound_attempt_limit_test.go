package outbound_attempt_limit

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

// func init() {
// 	providerResources = make(map[string]*schema.Resource)
// 	providerResources["genesyscloud_outbound_attempt_limit"] = ResourceOutboundAttemptLimit()
// }

func TestAccResourceOutboundAttemptLimit(t *testing.T) {

	t.Parallel()
	var (
		resourceLabel = "attempt_limit"
		resourcePath  = ResourceType + "." + resourceLabel
		// Create
		name                  = "Test Limit " + uuid.NewString()
		maxAttemptsPerContact = "5"
		maxAttemptsPerNumber  = "5"
		timeZoneId            = "America/Chicago"
		resetPeriod           = "TODAY"

		recallEntryType1                = "busy"
		recallEntryNbrAttempts1         = ""
		recallEntryMinsBetweenAttempts1 = "7"

		// Update
		nameUpdated                  = "Test Limit " + uuid.NewString()
		maxAttemptsPerContactUpdated = "4"
		maxAttemptsPerNumberUpdated  = "3"
		timeZoneIdUpdated            = "Etc/GMT"
		resetPeriodUpdated           = "NEVER"

		updatedRecallEntryType1                = "no_answer"
		updatedRecallEntryNbrAttempts1         = "2"
		updatedRecallEntryMinsBetweenAttempts1 = "6"

		updatedRecallEntryType2                = "answering_machine"
		updatedRecallEntryNbrAttempts2         = "1"
		updatedRecallEntryMinsBetweenAttempts2 = "5"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, make(map[string]*schema.Resource)),
		Steps: []resource.TestStep{
			{
				Config: GenerateAttemptLimitResource(
					resourceLabel,
					name,
					maxAttemptsPerContact,
					maxAttemptsPerNumber,
					timeZoneId,
					resetPeriod,
					generateRecallEntries(
						generateRecallEntry(recallEntryType1, recallEntryMinsBetweenAttempts1, recallEntryNbrAttempts1),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", name),
					resource.TestCheckResourceAttr(resourcePath, "max_attempts_per_contact", maxAttemptsPerContact),
					resource.TestCheckResourceAttr(resourcePath, "max_attempts_per_number", maxAttemptsPerNumber),
					resource.TestCheckResourceAttr(resourcePath, "time_zone_id", timeZoneId),
					resource.TestCheckResourceAttr(resourcePath, "reset_period", resetPeriod),
					resource.TestCheckResourceAttr(resourcePath, "recall_entries.0.busy.0.minutes_between_attempts", recallEntryMinsBetweenAttempts1),
					resource.TestCheckResourceAttrSet(resourcePath, "recall_entries.0.busy.0.nbr_attempts"),
				),
			},
			{
				// Update
				Config: GenerateAttemptLimitResource(
					resourceLabel,
					nameUpdated,
					maxAttemptsPerContactUpdated,
					maxAttemptsPerNumberUpdated,
					timeZoneIdUpdated,
					resetPeriodUpdated,
					generateRecallEntries(
						generateRecallEntry(updatedRecallEntryType1, updatedRecallEntryMinsBetweenAttempts1, updatedRecallEntryNbrAttempts1),
						generateRecallEntry(updatedRecallEntryType2, updatedRecallEntryMinsBetweenAttempts2, updatedRecallEntryNbrAttempts2),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", nameUpdated),
					resource.TestCheckResourceAttr(resourcePath, "max_attempts_per_contact", maxAttemptsPerContactUpdated),
					resource.TestCheckResourceAttr(resourcePath, "max_attempts_per_number", maxAttemptsPerNumberUpdated),
					resource.TestCheckResourceAttr(resourcePath, "time_zone_id", timeZoneIdUpdated),
					resource.TestCheckResourceAttr(resourcePath, "reset_period", resetPeriodUpdated),
					resource.TestCheckNoResourceAttr(resourcePath, "recall_entries.0."+recallEntryType1+".%"),
					resource.TestCheckResourceAttr(resourcePath, "recall_entries.0."+updatedRecallEntryType1+".0.nbr_attempts", updatedRecallEntryNbrAttempts1),
					resource.TestCheckResourceAttr(resourcePath, "recall_entries.0."+updatedRecallEntryType1+".0.minutes_between_attempts", updatedRecallEntryMinsBetweenAttempts1),
					resource.TestCheckResourceAttr(resourcePath, "recall_entries.0."+updatedRecallEntryType2+".0.nbr_attempts", updatedRecallEntryNbrAttempts2),
					resource.TestCheckResourceAttr(resourcePath, "recall_entries.0."+updatedRecallEntryType2+".0.minutes_between_attempts", updatedRecallEntryMinsBetweenAttempts2),
				),
			},
			{
				// Import/Read
				ResourceName:      resourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyAttemptLimitDestroyed,
	})
}

func generateRecallEntries(nestedBlocks ...string) string {
	return fmt.Sprintf(`
	recall_entries {
		%s
	}
`, strings.Join(nestedBlocks, "\n"))
}

func generateRecallEntry(recallType string, minsBetweenAttempts string, nbrAttempts string) string {
	if nbrAttempts != "" {
		nbrAttempts = fmt.Sprintf("nbr_attempts = %s", nbrAttempts)
	}
	return fmt.Sprintf(`
		%s {
			minutes_between_attempts = %s
			%s
		}
`, recallType, minsBetweenAttempts, nbrAttempts)
}

func testVerifyAttemptLimitDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		attemptLimit, resp, err := outboundAPI.GetOutboundAttemptlimit(rs.Primary.ID)
		if attemptLimit != nil {
			return fmt.Errorf("attempt limit (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Attempt limit not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All attempt limits destroyed
	return nil
}
