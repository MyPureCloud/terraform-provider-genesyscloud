package greeting_user

import (
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

func TestUserGreetingFetchConcurrency(t *testing.T) {
	provider.SdkClientPool = nil
	if got := userGreetingFetchConcurrency(); got != defaultUserGreetingConcurrency {
		t.Fatalf("expected default concurrency %d, got %d", defaultUserGreetingConcurrency, got)
	}
}
