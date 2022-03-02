package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
	"strconv"
	"strings"
	"testing"
	"time"
)

type ivrConfigStruct struct {
	resourceID  string
	name        string
	description string
	dnis        []string
	depends_on  string
}

func deleteIvrStartingWith(name string) {
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		ivrs, _, getErr := archAPI.GetArchitectIvrs(pageNum, pageSize, "", "", "")
		if getErr != nil {
			return
		}

		if ivrs.Entities == nil || len(*ivrs.Entities) == 0 {
			break
		}

		for _, ivr := range *ivrs.Entities {
			if strings.HasPrefix(*ivr.Name, name) {
				archAPI.DeleteArchitectIvr(*ivr.Id)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func TestAccResourceIvrConfigBasic(t *testing.T) {
	ivrConfigResource1 := "test-ivrconfig1"
	ivrConfigName := "terraform-ivrconfig-" + uuid.NewString()
	ivrConfigDescription := "Terraform IVR config"
	number1 := "+14175550011"
	number2 := "+14175550012"
	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	deleteIvrStartingWith("terraform-ivrconfig-")
	deleteDidPoolWithNumber(number1)
	deleteDidPoolWithNumber(number2)
	ivrConfigDnis := []string{number1, number2}
	didPoolResource1 := "test-didpool1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIvrConfigResource(&ivrConfigStruct{
					ivrConfigResource1,
					ivrConfigName,
					ivrConfigDescription,
					nil, // No dnis
					"",  // No depends_on
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+ivrConfigResource1, "name", ivrConfigName),
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+ivrConfigResource1, "description", ivrConfigDescription),
					hasEmptyDnis("genesyscloud_architect_ivr."+ivrConfigResource1),
				),
			},
			{
				// Update with new DNIS
				Config: generateDidPoolResource(&didPoolStruct{
					didPoolResource1,
					ivrConfigDnis[0],
					ivrConfigDnis[1],
					nullValue, // No description
					nullValue, // No comments
					nullValue, // No provider
				}) + generateIvrConfigResource(&ivrConfigStruct{
					ivrConfigResource1,
					ivrConfigName,
					ivrConfigDescription,
					ivrConfigDnis,
					"genesyscloud_telephony_providers_edges_did_pool." + didPoolResource1,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+ivrConfigResource1, "name", ivrConfigName),
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+ivrConfigResource1, "description", ivrConfigDescription),
					validateStringInArray("genesyscloud_architect_ivr."+ivrConfigResource1, "dnis", ivrConfigDnis[0]),
					validateStringInArray("genesyscloud_architect_ivr."+ivrConfigResource1, "dnis", ivrConfigDnis[1]),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_architect_ivr." + ivrConfigResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIvrConfigsDestroyed,
	})
}

func generateIvrConfigResource(ivrConfig *ivrConfigStruct) string {
	dnisStrs := make([]string, len(ivrConfig.dnis))
	for i, val := range ivrConfig.dnis {
		dnisStrs[i] = fmt.Sprintf("\"%s\"", val)
	}

	return fmt.Sprintf(`resource "genesyscloud_architect_ivr" "%s" {
		name = "%s"
		description = "%s"
		dnis = [%s]
		depends_on=[%s]
	}
	`, ivrConfig.resourceID,
		ivrConfig.name,
		ivrConfig.description,
		strings.Join(dnisStrs, ","),
		ivrConfig.depends_on,
	)
}

func testVerifyIvrConfigsDestroyed(state *terraform.State) error {
	architectApi := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_architect_ivr" {
			continue
		}

		ivrConfig, resp, err := architectApi.GetArchitectIvr(rs.Primary.ID)
		if ivrConfig != nil && ivrConfig.State != nil && *ivrConfig.State == "deleted" {
			continue
		}

		if ivrConfig != nil {
			return fmt.Errorf("IVR config (%s) still exists", rs.Primary.ID)
		}

		if isStatus404(resp) {
			// IVR Config not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All IVR Config pool destroyed
	return nil
}

func hasEmptyDnis(ivrResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		ivrResource, ok := state.RootModule().Resources[ivrResourceName]
		if !ok {
			return fmt.Errorf("Failed to find ivr config %s in state", ivrResourceName)
		}
		ivrID := ivrResource.Primary.ID

		dnisCountStr, ok := ivrResource.Primary.Attributes["dnis.#"]
		if !ok {
			return fmt.Errorf("No dnis found for %s in state", ivrID)
		}

		dnisCount, err := strconv.Atoi(dnisCountStr)
		if err != nil {
			return fmt.Errorf("Error while converting dnis count")
		}

		if dnisCount > 0 {
			return fmt.Errorf("Dnis is not empty.")
		}

		return nil
	}
}
