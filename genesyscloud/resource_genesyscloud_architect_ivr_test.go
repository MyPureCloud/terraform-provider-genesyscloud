package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v94/platformclientv2"
)

type ivrConfigStruct struct {
	resourceID  string
	name        string
	description string
	dnis        []string
	depends_on  string
	divisionId  string
}

func deleteIvrStartingWith(name string) {
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		ivrs, _, getErr := archAPI.GetArchitectIvrs(pageNum, pageSize, "", "", "", "", "")
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
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  ivrConfigResource1,
					name:        ivrConfigName,
					description: ivrConfigDescription,
					dnis:        nil, // No dnis
					depends_on:  "",  // No depends_on
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
					resourceID:  ivrConfigResource1,
					name:        ivrConfigName,
					description: ivrConfigDescription,
					dnis:        ivrConfigDnis,
					depends_on:  "genesyscloud_telephony_providers_edges_did_pool." + didPoolResource1,
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

func TestAccResourceIvrConfigDivision(t *testing.T) {
	ivrConfigResource1 := "test-ivrconfig1"
	ivrConfigName := "terraform-ivrconfig-" + uuid.NewString()
	ivrConfigDescription := "Terraform IVR config"
	number1 := "+14175550011"
	number2 := "+14175550012"
	divResource1 := "auth-division1"
	divResource2 := "auth-division2"
	divName1 := "TerraformDiv-" + uuid.NewString()
	divName2 := "TerraformDiv-" + uuid.NewString()
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
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateAuthDivisionResource(
					divResource1,
					divName1,
					nullValue, // No description
					nullValue, // Not home division
				) + generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  ivrConfigResource1,
					name:        ivrConfigName,
					description: ivrConfigDescription,
					dnis:        nil, // No dnis
					depends_on:  "",  // No depends_on
					divisionId:  "genesyscloud_auth_division." + divResource1 + ".id",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+ivrConfigResource1, "name", ivrConfigName),
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+ivrConfigResource1, "description", ivrConfigDescription),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_ivr."+ivrConfigResource1, "division_id", "genesyscloud_auth_division."+divResource1, "id"),
					hasEmptyDnis("genesyscloud_architect_ivr."+ivrConfigResource1),
				),
			},
			{
				// Update with new DNIS and division
				Config: generateAuthDivisionResource(
					divResource1,
					divName1,
					nullValue, // No description
					nullValue, // Not home division
				) + generateAuthDivisionResource(
					divResource2,
					divName2,
					nullValue, // No description
					nullValue, // Not home division
				) + generateDidPoolResource(&didPoolStruct{
					didPoolResource1,
					ivrConfigDnis[0],
					ivrConfigDnis[1],
					nullValue, // No description
					nullValue, // No comments
					nullValue, // No provider
				}) + generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  ivrConfigResource1,
					name:        ivrConfigName,
					description: ivrConfigDescription,
					dnis:        ivrConfigDnis,
					depends_on:  "genesyscloud_telephony_providers_edges_did_pool." + didPoolResource1,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+ivrConfigResource1, "name", ivrConfigName),
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+ivrConfigResource1, "description", ivrConfigDescription),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_ivr."+ivrConfigResource1, "division_id", "genesyscloud_auth_division."+divResource1, "id"),
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
			{
				Config: generateAuthDivisionResource(
					divResource1,
					divName1,
					nullValue, // No description
					nullValue, // Not home division
				) + generateAuthDivisionResource(
					divResource2,
					divName2,
					nullValue, // No description
					nullValue, // Not home division
				),
			},
		},
		CheckDestroy: testVerifyIvrConfigsDestroyed,
	})
}

func TestAccResourceIvrConfigDnisOverload(t *testing.T) {
	var (
		resourceID = "ivr"
		name       = "TF Test IVR " + uuid.NewString()
	)

	//t.Skip("Skipping if not using specific org that has these numbers")

	// TODO:
	// deleteIvrReservingDids()
	// defer recreateIvrReservingDids()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  resourceID,
					name:        name,
					description: "",
					dnis:        getSectionOfDnis(0, 20),
					depends_on:  "",
					divisionId:  "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+resourceID, "dnis.#", "21"),
				),
			},
			{
				Config: generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  resourceID,
					name:        name,
					description: "",
					dnis:        getSectionOfDnis(0, 48),
					depends_on:  "",
					divisionId:  "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+resourceID, "dnis.#", "49"),
				),
			},
			{
				Config: generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  resourceID,
					name:        name,
					description: "",
					dnis:        getSectionOfDnis(0, 12),
					depends_on:  "",
					divisionId:  "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+resourceID, "dnis.#", "13"),
				),
			},
			{
				Config: generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  resourceID,
					name:        name,
					description: "",
					dnis:        getSectionOfDnis(0, len(getAllDnisNumbersForIvrConfig())-1),
					depends_on:  "",
					divisionId:  "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+resourceID, "dnis.#", fmt.Sprintf("%v", len(getAllDnisNumbersForIvrConfig()))),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_architect_ivr." + resourceID,
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

	divisionId := ""
	if ivrConfig.divisionId != "" {
		divisionId = "division_id = " + ivrConfig.divisionId
	}

	return fmt.Sprintf(`resource "genesyscloud_architect_ivr" "%s" {
		name = "%s"
		description = "%s"
		dnis = [%s]
		depends_on=[%s]
		%s
	}
	`, ivrConfig.resourceID,
		ivrConfig.name,
		ivrConfig.description,
		strings.Join(dnisStrs, ","),
		ivrConfig.depends_on,
		divisionId,
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

func getSectionOfDnis(fromIndex int, toIndex int) []string {
	all := getAllDnisNumbersForIvrConfig()
	var toReturn []string

	if fromIndex > len(all)-1 || toIndex > len(all)-1 {
		fmt.Println("index out of range")
	}

	for i := fromIndex; i <= len(all); i++ {
		toReturn = append(toReturn, all[i])
		if i == toIndex {
			break
		}
	}

	return toReturn
}

func getAllDnisNumbersForIvrConfig() []string {
	all := []string{
		"+13177000697",
		"+13177000698",
		"+13177000699",
		"+13177000700",
		"+13177000701",
		"+13177000702",
		"+13177000703",
		"+13177000704",
		"+13177000705",
		"+13177000706",
		"+13177000707",
		"+13177000708",
		"+13177000709",
		"+13177000710",
		"+13177000711",
		"+13177000712",
		"+13177000713",
		"+13177000714",
		"+13177000715",
		"+13177000716",
		"+13177000717",
		"+13177000718",
		"+13177000719",
		"+13177000720",
		"+13177000721",
		"+13177000722",
		"+13177000723",
		"+13177000724",
		"+13177000725",
		"+13177000726",
		"+13177000727",
		"+13177000728",
		"+13177000729",
		"+13177000730",
		"+13177000731",
		"+13177000732",
		"+13177000733",
		"+13177000734",
		"+13177000735",
		"+13177000736",
		"+13177000737",
		"+13177000738",
		"+13177000739",
		"+13177000740",
		"+13177000741",
		"+13177000742",
		"+13177000743",
		"+13177000744",
		"+13177000745",
		"+13177000746",
		"+13177000747",
		"+13177000748",
		"+13177000749",
		"+13177000750",
		"+13177000752",
		"+13177000753",
		"+13177000754",
		"+13177000755",
		"+13177000756",
		"+13177000757",
		"+13177000758",
		"+13177000759",
		"+13177000760",
		"+13177000761",
		"+13177000762",
		"+13177000763",
		"+13177000764",
		"+13177000765",
		"+13177000766",
		"+13177000767",
		"+13177000768",
		"+13177000769",
		"+13177000770",
		"+13177000771",
		"+13177000772",
		"+13177000773",
		"+13177000774",
		"+13177000775",
		"+13177000776",
		"+13177000777",
		"+13177000778",
		"+13177000779",
		"+13177000780",
		"+13177000781",
		"+13177000782",
		"+13177000783",
		"+13177000784",
		"+13177000785",
		"+13177000786",
		"+13177000787",
		"+13177000788",
		"+13177000789",
		"+13177000790",
		"+13177000791",
		"+13177000792",
		"+13177000793",
		"+13177000794",
		"+13177000795",
		"+13177000796",
		"+13177000797",
		"+13177000798",
		"+13177000799",
		"+13177000800",
		"+13177000801",
		"+13177000802",
		"+13177000803",
		"+13177000804",
		"+13177000805",
		"+13177000806",
		"+13177000807",
		"+13177000808",
		"+13177000809",
		"+13177000810",
		"+13177000811",
		"+13177000812",
		"+13177000813",
		"+13177000814",
		"+13177000815",
		"+13177000816",
		"+13177000817",
		"+13177000818",
		"+13177000819",
		"+13177000820",
		"+13177000821",
		"+13177000822",
		"+13177000823",
		"+13177000824",
		"+13177000825",
		"+13177000826",
		"+13177000827",
		"+13177000828",
		"+13177000829",
		"+13177000830",
		"+13177000831",
		"+13177000832",
		"+13177000833",
		"+13177000834",
		"+13177000835",
		"+13177000836",
		"+13177000837",
		"+13177000838",
		"+13177000839",
		"+13177000840",
		"+13177000841",
		"+13177000842",
		"+13177000843",
		"+13177000844",
		"+13177000845",
		"+13177000846",
		"+13177000847",
		"+13177000848",
		"+13177000849",
		"+13177000850",
		"+13177000851",
		"+13177000852",
		"+13177000853",
		"+13177000854",
		"+13177000855",
		"+13177000856",
		"+13177000857",
		"+13177000858",
		"+13177000859",
		"+13177000860",
		"+13177000861",
		"+13177000862",
		"+13177000864",
		"+13177000865",
		"+13177000866",
		"+13177000867",
		"+13177000868",
		"+13177000869",
		"+13177000870",
		"+13177000871",
		"+13177000872",
		"+13177000873",
		"+13177000874",
		"+13177000875",
		"+13177000876",
		"+13177000877",
		"+13177000878",
		"+13177000879",
		"+13177000880",
		"+13177000881",
		"+13177000882",
		"+13177000883",
		"+13177000884",
		"+13177000885",
		"+13177000886",
		"+13177000887",
		"+13177000888",
		"+13177000889",
		"+13177000890",
		"+13177000891",
		"+13177000892",
		"+13177000893",
		"+13177000894",
		"+13177000895",
		"+13177000896",
		"+13177000897",
		"+13177000898",
		"+13177000899",
		"+13177000900",
		"+13177000901",
		"+13177000902",
		"+13177000903",
		"+13177000904",
		"+13177000905",
		"+13177000906",
		"+13177000907",
		"+13177000908",
		"+13177000909",
		"+13177000910",
		"+13177000911",
		"+13177000912",
		"+13177000913",
		"+13177000914",
		"+13177000915",
		"+13177000916",
		"+13177000917",
		"+13177000918",
		"+13177000919",
		"+13177000920",
		"+13177000921",
		"+13177000922",
		"+13177000923",
		"+13177000924",
		"+13177000925",
		"+13177000926",
		"+13177000927",
		"+13177000928",
		"+13177000929",
		"+13177000930",
		"+13177000931",
		"+13177000932",
		"+13177000933",
		"+13177000934",
		"+13177000935",
		"+13177000936",
		"+13177000937",
		"+13177000938",
		"+13177000939",
		"+13177000940",
		"+13177000941",
		"+13177000942",
		"+13177000943",
		"+13177000944",
		"+13177000945",
		"+13177000946",
		"+13177000947",
		"+13177000948",
		"+13177000949",
		"+13177000950",
		"+13177000951",
		"+13177000952",
		"+13177000953",
		"+13177000954",
		"+13177000955",
		"+13177000956",
		"+13177000957",
		"+13177000958",
		"+13177000959",
		"+13177000960",
		"+13177000961",
		"+13177000962",
		"+13177000963",
		"+13177000964",
		"+13177000965",
		"+13177000966",
		"+13177000967",
		"+13177000968",
		"+13177000969",
		"+13177000970",
		"+13177000971",
		"+13177000972",
		"+13177000973",
		"+13177000974",
		"+13177000975",
		"+13177000976",
		"+13177000977",
		"+13177000978",
		"+13177000979",
		"+13177000980",
		"+13177000981",
		"+13177000982",
		"+13177000983",
		"+13177000984",
		"+13177000985",
		"+13177000986",
		"+13177000987",
		"+13177000988",
		"+13177000989",
		"+13177000990",
		"+13177000991",
		"+13177000992",
		"+13177000993",
		"+13177000994",
		"+13177000995",
		"+13177000996",
		"+13177000997",
		"+13177000998",
		"+13177000999",
		"+13177001000",
	}
	return all
}
