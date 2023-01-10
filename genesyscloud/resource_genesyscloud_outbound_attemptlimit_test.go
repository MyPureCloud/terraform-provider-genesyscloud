package genesyscloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v89/platformclientv2"
)

func TestAccResourceOutboundAttemptLimit(t *testing.T) {
	t.Parallel()
	var (
		resourceId = "attempt_limit"
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateAttemptLimitResource(
					resourceId,
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
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "max_attempts_per_contact", maxAttemptsPerContact),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "max_attempts_per_number", maxAttemptsPerNumber),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "time_zone_id", timeZoneId),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "reset_period", resetPeriod),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "recall_entries.0.busy.0.minutes_between_attempts", recallEntryMinsBetweenAttempts1),
					resource.TestCheckResourceAttrSet("genesyscloud_outbound_attempt_limit."+resourceId, "recall_entries.0.busy.0.nbr_attempts"),
				),
			},
			{
				// Update
				Config: generateAttemptLimitResource(
					resourceId,
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
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "max_attempts_per_contact", maxAttemptsPerContactUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "max_attempts_per_number", maxAttemptsPerNumberUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "time_zone_id", timeZoneIdUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "reset_period", resetPeriodUpdated),
					resource.TestCheckNoResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "recall_entries.0."+recallEntryType1+".%"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "recall_entries.0."+updatedRecallEntryType1+".0.nbr_attempts", updatedRecallEntryNbrAttempts1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "recall_entries.0."+updatedRecallEntryType1+".0.minutes_between_attempts", updatedRecallEntryMinsBetweenAttempts1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "recall_entries.0."+updatedRecallEntryType2+".0.nbr_attempts", updatedRecallEntryNbrAttempts2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_attempt_limit."+resourceId, "recall_entries.0."+updatedRecallEntryType2+".0.minutes_between_attempts", updatedRecallEntryMinsBetweenAttempts2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_attempt_limit." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyAttemptLimitDestroyed,
	})
}

func generateAttemptLimitResource(
	resourceId string,
	name string,
	maxAttemptsPerContact string,
	maxAttemptsPerNumber string,
	timeZoneId string,
	resetPeriod string,
	nestedBlocks ...string,
) string {
	if maxAttemptsPerContact != "" {
		maxAttemptsPerContact = fmt.Sprintf(`max_attempts_per_contact = %s`, maxAttemptsPerContact)
	}
	if maxAttemptsPerNumber != "" {
		maxAttemptsPerNumber = fmt.Sprintf(`max_attempts_per_number = %s`, maxAttemptsPerNumber)
	}
	if timeZoneId != "" {
		timeZoneId = fmt.Sprintf(`time_zone_id = "%s"`, timeZoneId)
	}
	if resetPeriod != "" {
		resetPeriod = fmt.Sprintf(`reset_period = "%s"`, resetPeriod)
	}
	return fmt.Sprintf(`
resource "genesyscloud_outbound_attempt_limit" "%s" {
	name = "%s"
	%s
	%s
	%s
	%s
	%s
}
	`, resourceId, name, maxAttemptsPerContact, maxAttemptsPerNumber, timeZoneId, resetPeriod, strings.Join(nestedBlocks, "\n"))
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
		if rs.Type != "genesyscloud_outbound_attempt_limit" {
			continue
		}

		attemptLimit, resp, err := outboundAPI.GetOutboundAttemptlimit(rs.Primary.ID)
		if attemptLimit != nil {
			return fmt.Errorf("attempt limit (%s) still exists", rs.Primary.ID)
		} else if isStatus404(resp) {
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
