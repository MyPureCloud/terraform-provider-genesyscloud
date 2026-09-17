package architect_ivr_identity_resolution

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	architectIvr "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const (
	homeDivisionDataSourceLabel = "home"
	homeDivisionIdConfigRef     = "data.genesyscloud_auth_division_home.home.id"
	homeDivisionStateRef        = "data.genesyscloud_auth_division_home.home"
)

func TestAccResourceArchitectIvrIdentityResolution(t *testing.T) {
	var (
		identityResolutionResourceLabel = "test-identity-resolution"
		ivrResourceLabel                = "test-ivr"
		ivrName                         = "Terraform Test IVR-" + uuid.NewString()
		homeDivisionConfig              = gcloud.GenerateAuthDivisionHomeDataSource(homeDivisionDataSourceLabel)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		CheckDestroy:      testVerifyArchitectIvrIdentityResolutionDestroyed,
		Steps: []resource.TestStep{
			{
				Config: homeDivisionConfig + architectIvr.GenerateIvrConfigResource(&architectIvr.IvrConfigStruct{
					ResourceLabel: ivrResourceLabel,
					Name:          ivrName,
					Description:   "IR test IVR",
					Dnis:          nil,
					DependsOn:     "",
				}) + generateArchitectIvrIdentityResolutionResource(
					identityResolutionResourceLabel,
					"genesyscloud_architect_ivr."+ivrResourceLabel+".id",
					"true",
					homeDivisionIdConfigRef,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"genesyscloud_architect_ivr_identity_resolution."+identityResolutionResourceLabel, "ivr_id",
						"genesyscloud_architect_ivr."+ivrResourceLabel, "id",
					),
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr_identity_resolution."+identityResolutionResourceLabel, "resolve_identities", "true"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_architect_ivr_identity_resolution."+identityResolutionResourceLabel, "division_id",
						"data.genesyscloud_auth_division_home."+homeDivisionDataSourceLabel, "id",
					),
					verifyIdentityResolutionConfig("genesyscloud_architect_ivr."+ivrResourceLabel, true, homeDivisionStateRef),
				),
			},
			{
				Config: architectIvr.GenerateIvrConfigResource(&architectIvr.IvrConfigStruct{
					ResourceLabel: ivrResourceLabel,
					Name:          ivrName,
					Description:   "IR test IVR",
					Dnis:          nil,
					DependsOn:     "",
				}) + generateArchitectIvrIdentityResolutionResource(
					identityResolutionResourceLabel,
					"genesyscloud_architect_ivr."+ivrResourceLabel+".id",
					"false",
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr_identity_resolution."+identityResolutionResourceLabel, "resolve_identities", "false"),
					verifyIdentityResolutionStateDivisionCleared("genesyscloud_architect_ivr_identity_resolution."+identityResolutionResourceLabel),
					verifyIdentityResolutionConfig("genesyscloud_architect_ivr."+ivrResourceLabel, false, ""),
				),
			},
			{
				// Explicit "*" is equivalent to omitted division_id (both mean the unassigned / STAR division).
				Config: architectIvr.GenerateIvrConfigResource(&architectIvr.IvrConfigStruct{
					ResourceLabel: ivrResourceLabel,
					Name:          ivrName,
					Description:   "IR test IVR",
					Dnis:          nil,
					DependsOn:     "",
				}) + generateArchitectIvrIdentityResolutionResource(
					identityResolutionResourceLabel,
					"genesyscloud_architect_ivr."+ivrResourceLabel+".id",
					"false",
					`"*"`,
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: architectIvr.GenerateIvrConfigResource(&architectIvr.IvrConfigStruct{
					ResourceLabel: ivrResourceLabel,
					Name:          ivrName,
					Description:   "IR test IVR",
					Dnis:          nil,
					DependsOn:     "",
				}),
				Check: resource.ComposeTestCheckFunc(
					verifyIdentityResolutionDefault("genesyscloud_architect_ivr." + ivrResourceLabel),
				),
			},
		},
	})
}

func generateArchitectIvrIdentityResolutionResource(resourceLabel, ivrId, resolveIdentities, divisionId string) string {
	divisionBlock := ""
	if divisionId != "" {
		divisionBlock = fmt.Sprintf("\n    division_id = %s", divisionId)
	}

	return fmt.Sprintf(`resource "genesyscloud_architect_ivr_identity_resolution" "%s" {
  		ivr_id = %s
  		resolve_identities = %s%s
	}`, resourceLabel, ivrId, resolveIdentities, divisionBlock)
}

func verifyIdentityResolutionStateDivisionCleared(resourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("failed to find resource %s in state", resourcePath)
		}

		if divisionId, ok := resourceState.Primary.Attributes["division_id"]; ok && divisionId != "" {
			return fmt.Errorf("expected division_id to be cleared from state for %s, still have %q", resourcePath, divisionId)
		}

		return nil
	}
}

func verifyIdentityResolutionConfig(ivrResourcePath string, resolveIdentities bool, divisionStateRef string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ivrResource, ok := state.RootModule().Resources[ivrResourcePath]
		if !ok {
			return fmt.Errorf("failed to find ivr %s in state", ivrResourcePath)
		}

		expectedDivisionId := ""
		if divisionStateRef != "" {
			divisionResource, ok := state.RootModule().Resources[divisionStateRef]
			if !ok {
				return fmt.Errorf("failed to find division %s in state", divisionStateRef)
			}
			expectedDivisionId = divisionResource.Primary.ID
		}

		architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)
		config, _, err := architectApi.GetArchitectIvrIdentityresolution(ivrResource.Primary.ID)
		if err != nil {
			return err
		}

		if config.ResolveIdentities == nil {
			return fmt.Errorf("identity resolution config missing for ivr %s", ivrResource.Primary.ID)
		}

		if *config.ResolveIdentities != resolveIdentities {
			return fmt.Errorf("expected resolve_identities=%t for ivr %s, got %t", resolveIdentities, ivrResource.Primary.ID, *config.ResolveIdentities)
		}

		if expectedDivisionId != "" {
			if config.Division == nil || config.Division.Id == nil {
				return fmt.Errorf("expected division_id=%s for ivr %s, got none", expectedDivisionId, ivrResource.Primary.ID)
			}
			if *config.Division.Id != expectedDivisionId {
				return fmt.Errorf("expected division_id=%s for ivr %s, got %s", expectedDivisionId, ivrResource.Primary.ID, *config.Division.Id)
			}
		} else if config.Division != nil && config.Division.Id != nil {
			divisionId := *config.Division.Id
			if !isUnassignedDivisionId(divisionId) {
				return fmt.Errorf("expected the unassigned division (* or empty) for ivr %s, got division_id=%s", ivrResource.Primary.ID, divisionId)
			}
		}

		return nil
	}
}

func verifyIdentityResolutionDefault(ivrResourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ivrResource, ok := state.RootModule().Resources[ivrResourcePath]
		if !ok {
			return fmt.Errorf("failed to find ivr %s in state", ivrResourcePath)
		}

		architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)
		config, _, err := architectApi.GetArchitectIvrIdentityresolution(ivrResource.Primary.ID)
		if err != nil {
			return err
		}

		if !isDefaultIdentityResolutionConfig(config) {
			resolveIdentities := "nil"
			if config.ResolveIdentities != nil {
				resolveIdentities = fmt.Sprintf("%t", *config.ResolveIdentities)
			}
			return fmt.Errorf("expected default identity resolution config for ivr %s, got resolve_identities=%s", ivrResource.Primary.ID, resolveIdentities)
		}

		return nil
	}
}

// testVerifyArchitectIvrIdentityResolutionDestroyed verifies destroy reset IR config to the
// platform default. Parent 404 is accepted because the final test destroy may also remove the IVR.
func testVerifyArchitectIvrIdentityResolutionDestroyed(state *terraform.State) error {
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		ivrId := rs.Primary.Attributes["ivr_id"]
		if ivrId == "" {
			ivrId = rs.Primary.ID
		}

		config, resp, err := architectApi.GetArchitectIvrIdentityresolution(ivrId)
		if util.IsStatus404(resp) {
			continue
		}
		if err != nil {
			return fmt.Errorf("unexpected error verifying identity resolution destroy for ivr %s: %s", ivrId, err)
		}
		if !isDefaultIdentityResolutionConfig(config) {
			return fmt.Errorf("expected default identity resolution config after destroy for ivr %s", ivrId)
		}
	}

	return nil
}
