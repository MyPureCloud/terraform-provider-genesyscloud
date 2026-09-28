package knowledge_document

import (
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
)

func strPtr(s string) *string { return &s }

// TestUnitFindExactLabelMatch covers DEVTOOLING-1821.
// The knowledge label name query is a partial/prefix match, so searching for "Home" also returns
// labels like "Auto and Home". Previously the resolver took the first entity, which attached the
// wrong label to the document and caused the label_names consistency mismatch during a replace.
// findExactLabelMatch must return the entity whose name matches exactly.
func TestUnitFindExactLabelMatch(t *testing.T) {
	entities := []platformclientv2.Labelresponse{
		{Id: strPtr("id-auto-and-home"), Name: strPtr("Auto and Home")},
		{Id: strPtr("id-home"), Name: strPtr("Home")},
		{Id: strPtr("id-second-home"), Name: strPtr("Second Home")},
	}

	tests := []struct {
		name      string
		query     string
		wantID    string
		wantMatch bool
	}{
		{name: "exact_home_not_auto_and_home", query: "Home", wantID: "id-home", wantMatch: true},
		{name: "exact_auto_and_home", query: "Auto and Home", wantID: "id-auto-and-home", wantMatch: true},
		{name: "exact_second_home", query: "Second Home", wantID: "id-second-home", wantMatch: true},
		{name: "no_match_returns_nil", query: "Nonexistent", wantMatch: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findExactLabelMatch(&entities, tt.query)
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

// TestUnitFindExactCategoryMatch covers the same substring-collision issue for category_name.
func TestUnitFindExactCategoryMatch(t *testing.T) {
	entities := []platformclientv2.Categoryresponse{
		{Id: strPtr("id-home"), Name: strPtr("Home")},
		{Id: strPtr("id-home-insurance"), Name: strPtr("Home Insurance")},
	}

	if got := findExactCategoryMatch(&entities, "Home"); got == nil || got.Id == nil || *got.Id != "id-home" {
		t.Fatalf("expected exact category match id-home for \"Home\", got %v", got)
	}
	if got := findExactCategoryMatch(&entities, "Home Insurance"); got == nil || got.Id == nil || *got.Id != "id-home-insurance" {
		t.Fatalf("expected exact category match id-home-insurance for \"Home Insurance\", got %v", got)
	}
	if got := findExactCategoryMatch(&entities, "Auto"); got != nil {
		t.Fatalf("expected no category match for \"Auto\", got %v", got)
	}
}
