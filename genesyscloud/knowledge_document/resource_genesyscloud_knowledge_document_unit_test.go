package knowledge_document

import (
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

func strPtr(s string) *string { return &s }

// TestUnitFindExactLabelMatch covers DEVTOOLING-1821.
// The knowledge label name query is a partial/prefix, case-insensitive match, so searching for
// "Home" also returns labels like "Auto and Home". Previously the resolver took the first entity,
// which attached the wrong label and caused the label_names consistency mismatch during a replace.
//
// findExactLabelMatch must:
//   - prefer a case-sensitive exact match,
//   - fall back to a UNIQUE case-insensitive match (API name filter is case-insensitive),
//   - report ambiguous=true when multiple differently-cased candidates exist and none is exact.
func TestUnitFindExactLabelMatch(t *testing.T) {
	tests := []struct {
		name          string
		entities      []platformclientv2.Labelresponse
		query         string
		wantID        string
		wantMatch     bool
		wantAmbiguous bool
	}{
		{
			name: "exact_home_not_auto_and_home",
			entities: []platformclientv2.Labelresponse{
				{Id: strPtr("id-auto-and-home"), Name: strPtr("Auto and Home")},
				{Id: strPtr("id-home"), Name: strPtr("Home")},
				{Id: strPtr("id-second-home"), Name: strPtr("Second Home")},
			},
			query: "Home", wantID: "id-home", wantMatch: true,
		},
		{
			name: "exact_auto_and_home",
			entities: []platformclientv2.Labelresponse{
				{Id: strPtr("id-auto-and-home"), Name: strPtr("Auto and Home")},
				{Id: strPtr("id-home"), Name: strPtr("Home")},
			},
			query: "Auto and Home", wantID: "id-auto-and-home", wantMatch: true,
		},
		{
			name: "case_insensitive_unique_fallback",
			entities: []platformclientv2.Labelresponse{
				{Id: strPtr("id-home"), Name: strPtr("Home")},
			},
			query: "home", wantID: "id-home", wantMatch: true,
		},
		{
			name: "case_insensitive_ambiguous_reports_ambiguous",
			entities: []platformclientv2.Labelresponse{
				{Id: strPtr("id-home-upper"), Name: strPtr("Home")},
				{Id: strPtr("id-home-lower"), Name: strPtr("home")},
			},
			query: "HOME", wantMatch: false, wantAmbiguous: true,
		},
		{
			name: "no_match_returns_nil",
			entities: []platformclientv2.Labelresponse{
				{Id: strPtr("id-home"), Name: strPtr("Home")},
			},
			query: "Nonexistent", wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ambiguous := findExactLabelMatch(&tt.entities, tt.query)
			if ambiguous != tt.wantAmbiguous {
				t.Fatalf("query %q: ambiguous=%v, want %v", tt.query, ambiguous, tt.wantAmbiguous)
			}
			if !tt.wantMatch {
				if got != nil {
					t.Fatalf("expected no match for %q, got %v", tt.query, *got.Id)
				}
				return
			}
			if got == nil || got.Id == nil {
				t.Fatalf("expected match %q for query %q, got nil", tt.wantID, tt.query)
			}
			if *got.Id != tt.wantID {
				t.Fatalf("query %q: expected id %q, got %q", tt.query, tt.wantID, *got.Id)
			}
		})
	}
}

// TestUnitFindExactCategoryMatch covers the same substring/case-collision issue for category_name.
func TestUnitFindExactCategoryMatch(t *testing.T) {
	entities := []platformclientv2.Categoryresponse{
		{Id: strPtr("id-home"), Name: strPtr("Home")},
		{Id: strPtr("id-home-insurance"), Name: strPtr("Home Insurance")},
	}

	if got, amb := findExactCategoryMatch(&entities, "Home"); amb || got == nil || got.Id == nil || *got.Id != "id-home" {
		t.Fatalf("expected exact category match id-home for \"Home\", got %v ambiguous=%v", got, amb)
	}
	if got, amb := findExactCategoryMatch(&entities, "Home Insurance"); amb || got == nil || got.Id == nil || *got.Id != "id-home-insurance" {
		t.Fatalf("expected exact category match id-home-insurance for \"Home Insurance\", got %v ambiguous=%v", got, amb)
	}
	// Case-insensitive unique fallback.
	if got, amb := findExactCategoryMatch(&entities, "home insurance"); amb || got == nil || got.Id == nil || *got.Id != "id-home-insurance" {
		t.Fatalf("expected case-insensitive fallback id-home-insurance, got %v ambiguous=%v", got, amb)
	}
	// No match.
	if got, amb := findExactCategoryMatch(&entities, "Auto"); got != nil || amb {
		t.Fatalf("expected no category match for \"Auto\", got %v ambiguous=%v", got, amb)
	}
	// Ambiguous case-insensitive.
	ambEntities := []platformclientv2.Categoryresponse{
		{Id: strPtr("id-a"), Name: strPtr("Home")},
		{Id: strPtr("id-b"), Name: strPtr("home")},
	}
	if got, amb := findExactCategoryMatch(&ambEntities, "HOME"); got != nil || !amb {
		t.Fatalf("expected ambiguous for \"HOME\", got %v ambiguous=%v", got, amb)
	}
}
