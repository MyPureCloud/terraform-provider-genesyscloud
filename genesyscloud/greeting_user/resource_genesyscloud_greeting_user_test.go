package greeting_user

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	userResource "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceUserGreeting(t *testing.T) {

	var (
		resourceLabel = "greeting"
		name1         = "Test Greeting " + uuid.NewString()
		type1         = "NAME"
		ownerType1    = "USER"
		audioTts1     = "This is a test greeting"

		name2      = "Test Greeting " + uuid.NewString()
		type2      = "NAME"
		ownerType2 = "USER"
		audioTts2  = "This is an updated test greeting"

		randomizer        = uuid.NewString()
		userName          = "TestUser" + randomizer
		userEmail         = randomizer + "@website.com"
		userResourceLabel = "sample_user"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: userResource.GenerateBasicUserResource(
					userResourceLabel,
					userEmail,
					userName,
				) + GenerateGreetingUser(
					resourceLabel,
					name1,
					type1,
					ownerType1,
					"genesyscloud_user."+userResourceLabel+".id",
					audioTts1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "type", type1),
					resource.TestCheckResourceAttrSet(ResourceType+"."+resourceLabel, "owner_type"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "audio_tts", audioTts1),
					resource.TestCheckResourceAttrSet(ResourceType+"."+resourceLabel, "user_id"),
				),
			},
			{
				Config: userResource.GenerateBasicUserResource(
					userResourceLabel,
					userEmail,
					userName,
				) + GenerateGreetingUser(
					resourceLabel,
					name2,
					type2,
					ownerType2,
					"genesyscloud_user."+userResourceLabel+".id",
					audioTts2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "type", type2),
					resource.TestCheckResourceAttrSet(ResourceType+"."+resourceLabel, "owner_type"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "audio_tts", audioTts2),
					resource.TestCheckResourceAttrSet(ResourceType+"."+resourceLabel, "user_id"),
				),
			},
			{
				ResourceName:            ResourceType + "." + resourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_id"},
			},
		},
		CheckDestroy: testVerifyGreetingDestroyed,
	})
}

func TestUnitUserGreetingFetchConcurrency(t *testing.T) {
	prevPool := provider.SdkClientPool
	t.Cleanup(func() {
		provider.SdkClientPool = prevPool
	})

	provider.SdkClientPool = nil
	if got := userGreetingFetchConcurrency(); got != 1 {
		t.Fatalf("userGreetingFetchConcurrency() = %d, want 1 when pool is nil", got)
	}
}

func TestUnitIsGreetingsPermissionDenied(t *testing.T) {
	if isGreetingsPermissionDenied(nil) {
		t.Fatal("expected false for nil response")
	}
	if isGreetingsPermissionDenied(&platformclientv2.APIResponse{StatusCode: http.StatusNotFound}) {
		t.Fatal("expected false for non-403 response")
	}
	if !isGreetingsPermissionDenied(&platformclientv2.APIResponse{StatusCode: http.StatusForbidden}) {
		t.Fatal("expected true for 403 response")
	}
}

func testVerifyGreetingDestroyed(state *terraform.State) error {
	greetingAPI := platformclientv2.NewGreetingsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}
		greeting, resp, err := greetingAPI.GetGreeting(rs.Primary.ID)
		if greeting != nil {
			return fmt.Errorf("greeting (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			continue
		} else {
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	return nil
}
