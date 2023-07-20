package genesyscloud

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
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
	if _, err := AuthorizeSdk(); err != nil {
		t.Fatal(err)
	}
	deleteIvrStartingWith("terraform-ivrconfig-")
	if err := deleteDidPoolWithNumber(number1); err != nil {
		t.Fatalf("error deleting did pool start number: %v", err)
	}
	if err := deleteDidPoolWithNumber(number2); err != nil {
		t.Fatalf("error deleting did pool end number: %v", err)
	}
	ivrConfigDnis := []string{number1, number2}
	didPoolResource1 := "test-didpool1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
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
					ValidateStringInArray("genesyscloud_architect_ivr."+ivrConfigResource1, "dnis", ivrConfigDnis[0]),
					ValidateStringInArray("genesyscloud_architect_ivr."+ivrConfigResource1, "dnis", ivrConfigDnis[1]),
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
	if _, err := AuthorizeSdk(); err != nil {
		t.Fatal(err)
	}
	deleteIvrStartingWith("terraform-ivrconfig-")
	if err := deleteDidPoolWithNumber(number1); err != nil {
		t.Fatalf("error deleting did pool start number: %v", err)
	}
	if err := deleteDidPoolWithNumber(number2); err != nil {
		t.Fatalf("error deleting did pool end number: %v", err)
	}
	ivrConfigDnis := []string{number1, number2}
	didPoolResource1 := "test-didpool1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
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
					ValidateStringInArray("genesyscloud_architect_ivr."+ivrConfigResource1, "dnis", ivrConfigDnis[0]),
					ValidateStringInArray("genesyscloud_architect_ivr."+ivrConfigResource1, "dnis", ivrConfigDnis[1]),
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

		didRangeLength    = 200 // Should be atleast 50 to avoid index out of bounds errors below
		didPoolResourceId = "did_pool"
		startNumber       = 35375550120
		endNumber         = startNumber + didRangeLength
		startNumberStr    = fmt.Sprintf("+%v", startNumber)
		endNumberStr      = fmt.Sprintf("+%v", endNumber)
	)

	/*
		To avoid clashes, try to get final existing did number and create a pool outside of that range
		If err is not nil, use the hardcoded phone number variables
	*/
	lastNumber, err := getLastDidNumberAsInteger()
	if err == nil {
		startNumber = lastNumber + 5
		endNumber = startNumber + didRangeLength
		startNumberStr = fmt.Sprintf("+%v", startNumber)
		endNumberStr = fmt.Sprintf("+%v", endNumber)
	} else {
		log.Printf("Failed to get last did number for ivr tests: %v", err)
	}

	allNumbers := createStringArrayOfPhoneNumbers(startNumber, endNumber)

	didPoolResource := generateDidPoolResource(&didPoolStruct{
		didPoolResourceId,
		startNumberStr,
		endNumberStr,
		nullValue, // No description
		nullValue, // No comments
		nullValue, // No provider
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: didPoolResource + generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  resourceID,
					name:        name,
					description: "",
					dnis:        createStringArrayOfPhoneNumbers(startNumber, startNumber+20),
					depends_on:  "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceId,
					divisionId:  "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+resourceID, "dnis.#", "20"),
				),
			},
			{
				Config: didPoolResource + generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  resourceID,
					name:        name,
					description: "",
					dnis:        createStringArrayOfPhoneNumbers(startNumber, startNumber+48),
					depends_on:  "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceId,
					divisionId:  "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+resourceID, "dnis.#", "48"),
				),
			},
			{
				Config: didPoolResource + generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  resourceID,
					name:        name,
					description: "",
					dnis:        createStringArrayOfPhoneNumbers(startNumber, startNumber+12),
					depends_on:  "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceId,
					divisionId:  "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+resourceID, "dnis.#", "12"),
				),
			},
			{
				Config: didPoolResource + generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  resourceID,
					name:        name,
					description: "",
					dnis:        createStringArrayOfPhoneNumbers(startNumber, endNumber),
					depends_on:  "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceId,
					divisionId:  "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+resourceID, "dnis.#", fmt.Sprintf("%v", len(allNumbers))),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_architect_ivr." + resourceID,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: didPoolResource, // Extra step to ensure take-down is done correctly
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
		name        = "%s"
		description = "%s"
		dnis        = [%s]
		depends_on  = [%s]
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

		if IsStatus404(resp) {
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

func createStringArrayOfPhoneNumbers(from, to int) []string {
	var slice []string
	for i := 0; i < to-from; i++ {
		slice = append(slice, fmt.Sprintf("+%v", from+i))
	}
	return slice
}

func getLastDidNumberAsInteger() (int, error) {
	config := platformclientv2.GetDefaultConfiguration()
	api := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(config)
	if err := config.AuthorizeClientCredentials(os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"), os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET")); err != nil {
		return 0, err
	}

	// Get the page count
	result, err := getDidNumbers(api, 1)
	if err != nil {
		return 0, err
	}

	// Get last page
	lastPage, err := getDidNumbers(api, *result.PageCount)
	if err != nil {
		return 0, err
	}

	var lastNumberString string
	if lastPage.Entities != nil && len(*lastPage.Entities) > 0 {
		lastItem := (*lastPage.Entities)[len(*lastPage.Entities)-1]
		lastNumberString = *lastItem.Number
	}

	if lastNumberString == "" {
		return 0, fmt.Errorf("Failed to retrieve last did number")
	}

	lastNumberString = strings.Replace(lastNumberString, "+", "", -1)

	lastNumberInt, err := strconv.Atoi(lastNumberString)
	if err != nil {
		return lastNumberInt, err
	}

	return lastNumberInt, nil
}

func getDidNumbers(api *platformclientv2.TelephonyProvidersEdgeApi, pageNumber int) (*platformclientv2.Didnumberentitylisting, error) {
	var (
		varType  = "ASSIGNED_AND_UNASSIGNED"
		pageSize = 100
		result   *platformclientv2.Didnumberentitylisting
	)
	result, response, err := api.GetTelephonyProvidersEdgesDidpoolsDids(varType, []string{}, "", pageSize, pageNumber, "")
	if err != nil {
		return result, err
	}
	if response.Error != nil {
		return result, fmt.Errorf("Response error: %v", response.Error)
	}
	return result, nil
}
