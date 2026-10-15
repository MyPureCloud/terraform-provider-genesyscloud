package organization_presence_definition

import (
	"encoding/json"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

// apiPayload is a real /api/v2/presencedefinitions response captured from an org. Note that none of the
// entities include a "type" field, so VarType unmarshals as nil for all of them.
const apiPayload = `{
  "entities": [
    {"id":"6a3af858-942f-489d-9700-5f9bcdcdae9b","languageLabels":{"en_US":"Available"},"systemPresence":"Available","deactivated":false,"primary":true},
    {"id":"992c2437-1833-4737-9ff4-eafdc4a360c8","languageLabels":{"en_US":"Updated Java Presence"},"systemPresence":"Break","deactivated":false,"primary":false},
    {"id":"09dd6ff4-ff90-4d98-a3de-88a63db56e64","languageLabels":{"en":"Another Break Presence","en_US":"Another Break Presence"},"systemPresence":"Break","deactivated":false,"primary":false},
    {"id":"31fe3bac-dea6-44b7-bed7-47f91660a1a0","languageLabels":{"en_US":"Busy"},"systemPresence":"Busy","deactivated":false,"primary":true},
    {"id":"7ffa01cd-7866-4ad7-8faa-31e11b42f51f","languageLabels":{"en_US":"Idle"},"systemPresence":"Idle","deactivated":false,"primary":true},
    {"id":"5e5c5c66-ea97-4e7f-ac41-6424784829f2","languageLabels":{"en_US":"Away"},"systemPresence":"Away","deactivated":false,"primary":true},
    {"id":"bbdff279-7ae1-48ea-bade-7831d7234c64","languageLabels":{"en_US":"Meeting"},"systemPresence":"Meeting","deactivated":false,"primary":true},
    {"id":"2fa68d50-5edb-4e01-9390-522615f12243","languageLabels":{"en_US":"Away From Keyboard V2","es":"del teclado V2"},"systemPresence":"Away","deactivated":false,"primary":false},
    {"id":"3fd96123-badb-4f69-bc03-1b1ccc6d8014","languageLabels":{"en_US":"Meal"},"systemPresence":"Meal","deactivated":false,"primary":true},
    {"id":"d2390a99-8546-bad9-8f0a-219548e8aeb0","languageLabels":{"en_US":"Training"},"systemPresence":"Training","deactivated":false,"primary":true},
    {"id":"e08eaf1b-ee47-4fa9-a231-1200e284798f","languageLabels":{"en_US":"On Queue"},"systemPresence":"On Queue","deactivated":false,"primary":true},
    {"id":"227b37e2-f1d0-4dd0-8f50-badd7cf6d158","languageLabels":{"en_US":"Break"},"systemPresence":"Break","deactivated":false,"primary":true},
    {"id":"ccf3c10a-aa2c-4845-8e8d-f59fa48c58e5","languageLabels":{"en_US":"Offline"},"systemPresence":"Offline","deactivated":false,"primary":true}
  ],
  "pageSize": 20,
  "pageNumber": 1,
  "total": 13,
  "pageCount": 1
}`

func parsePayload(t *testing.T) *[]platformclientv2.Organizationpresencedefinition {
	t.Helper()
	var entityListing platformclientv2.Organizationpresencedefinitionentitylisting
	if err := json.Unmarshal([]byte(apiPayload), &entityListing); err != nil {
		t.Fatalf("failed to unmarshal api payload: %s", err)
	}
	if entityListing.Entities == nil {
		t.Fatal("no entities parsed from payload")
	}
	return entityListing.Entities
}

// TestUnitFilter_NoTypeFieldMatchesNothing documents the current behavior: because the live API payload
// carries no "type" field, VarType is nil on every entity and the User-type filter excludes all of them.
func TestUnitFilter_NoTypeFieldMatchesNothing(t *testing.T) {
	defs := parsePayload(t)

	for _, d := range *defs {
		if d.VarType != nil {
			t.Fatalf("expected VarType to be nil for payload entities, got %q", *d.VarType)
		}
	}

	matches := filterOrganizationPresenceDefinitions(defs, "Updated Java Presence", "", "")
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches because no entity has type=User, got %d", len(matches))
	}
}

// TestUnitFilter_UserTypeMatching verifies the matching, disambiguation, and ambiguity behavior when
// entities are correctly typed as "User".
func TestUnitFilter_UserTypeMatching(t *testing.T) {
	defs := parsePayload(t)
	// Force the non-primary (user-created) definitions to type "User" to simulate correct typing.
	userIds := map[string]bool{
		"992c2437-1833-4737-9ff4-eafdc4a360c8": true, // Updated Java Presence / Break
		"09dd6ff4-ff90-4d98-a3de-88a63db56e64": true, // Another Break Presence / Break
		"2fa68d50-5edb-4e01-9390-522615f12243": true, // Away From Keyboard V2 / Away
	}
	typed := make([]platformclientv2.Organizationpresencedefinition, 0, len(*defs))
	for _, d := range *defs {
		if d.Id != nil && userIds[*d.Id] {
			d.VarType = platformclientv2.String("User")
		}
		typed = append(typed, d)
	}

	tests := []struct {
		name           string
		lookupName     string
		systemPresence string
		wantCount      int
		wantId         string
	}{
		{"unique by name", "Updated Java Presence", "", 1, "992c2437-1833-4737-9ff4-eafdc4a360c8"},
		{"en_US label used", "Away From Keyboard V2", "", 1, "2fa68d50-5edb-4e01-9390-522615f12243"},
		{"system presence filter narrows", "Another Break Presence", "Break", 1, "09dd6ff4-ff90-4d98-a3de-88a63db56e64"},
		{"wrong system presence excludes", "Another Break Presence", "Away", 0, ""},
		{"name not present", "Does Not Exist", "", 0, ""},
		{"primary/system definition not matched", "Available", "", 0, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matches := filterOrganizationPresenceDefinitions(&typed, tc.lookupName, tc.systemPresence, "")
			if len(matches) != tc.wantCount {
				t.Fatalf("expected %d matches, got %d", tc.wantCount, len(matches))
			}
			if tc.wantCount == 1 && *matches[0].Id != tc.wantId {
				t.Fatalf("expected id %s, got %s", tc.wantId, *matches[0].Id)
			}
		})
	}
}

// TestUnitFilter_Ambiguous confirms multiple matches are surfaced so the read can error.
func TestUnitFilter_Ambiguous(t *testing.T) {
	same := []platformclientv2.Organizationpresencedefinition{
		{Id: platformclientv2.String("a"), VarType: platformclientv2.String("User"), SystemPresence: platformclientv2.String("Break"), LanguageLabels: &map[string]string{"en_US": "Lunch"}},
		{Id: platformclientv2.String("b"), VarType: platformclientv2.String("User"), SystemPresence: platformclientv2.String("Meal"), LanguageLabels: &map[string]string{"en_US": "Lunch"}},
	}
	matches := filterOrganizationPresenceDefinitions(&same, "Lunch", "", "")
	if len(matches) != 2 {
		t.Fatalf("expected 2 ambiguous matches, got %d", len(matches))
	}
	// system_presence disambiguates down to one.
	matches = filterOrganizationPresenceDefinitions(&same, "Lunch", "Meal", "")
	if len(matches) != 1 || *matches[0].Id != "b" {
		t.Fatalf("expected single match id=b, got %d matches", len(matches))
	}
}
