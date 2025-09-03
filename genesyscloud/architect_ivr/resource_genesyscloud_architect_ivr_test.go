package architect_ivr

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	didPool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceArchitectIvrConfigBasic(t *testing.T) {
	ivrConfigResourceLabel := "test-ivrconfig1"
	ivrFullResourcePath := ResourceType + "." + ivrConfigResourceLabel

	ivrConfigName := "terraform-ivrconfig-" + uuid.NewString()
	ivrConfigDescription := "Terraform IVR config"
	number1 := "+14175550011"
	number2 := "+14175550012"
	ivrConfigDnis := []string{number1, number2}
	didPoolResourceLabel := "test-didpool1"

	// did pool cleanup
	resp, err := didPool.DeleteDidPoolWithStartAndEndNumber(context.Background(), number1, number2, sdkConfig)
	if err != nil {
		respStr := "<nil>"
		if resp != nil {
			respStr = strconv.Itoa(resp.StatusCode)
		}
		t.Logf("Failed to delete DID pool: %s. API Response: %s", err.Error(), respStr)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceLabel: ivrConfigResourceLabel,
					Name:          ivrConfigName,
					Description:   ivrConfigDescription,
					Dnis:          nil, // No dnis
					DependsOn:     "",  // No depends_on
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ivrFullResourcePath, "name", ivrConfigName),
					resource.TestCheckResourceAttr(ivrFullResourcePath, "description", ivrConfigDescription),
					resource.TestCheckResourceAttr(ivrFullResourcePath, "dnis.#", "0"),
				),
			},
			{
				// Update with new DNIS
				Config: didPool.GenerateDidPoolResource(&didPool.DidPoolStruct{
					ResourceLabel:    didPoolResourceLabel,
					StartPhoneNumber: ivrConfigDnis[0],
					EndPhoneNumber:   ivrConfigDnis[1],
					Description:      util.NullValue, // No description
					Comments:         util.NullValue, // No comments
					PoolProvider:     util.NullValue, // No provider
				}) + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceLabel: ivrConfigResourceLabel,
					Name:          ivrConfigName,
					Description:   ivrConfigDescription,
					Dnis:          ivrConfigDnis,
					DependsOn:     didPool.ResourceType + "." + didPoolResourceLabel,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ivrFullResourcePath, "name", ivrConfigName),
					resource.TestCheckResourceAttr(ivrFullResourcePath, "description", ivrConfigDescription),
					resource.TestCheckResourceAttr(ivrFullResourcePath, "dnis.#", "2"),
					util.ValidateStringInArray(ivrFullResourcePath, "dnis", ivrConfigDnis[0]),
					util.ValidateStringInArray(ivrFullResourcePath, "dnis", ivrConfigDnis[1]),
				),
			},
			{
				// Import/Read
				ResourceName:      ivrFullResourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIvrConfigsDestroyed,
	})
}

func TestAccResourceArchitectIvrConfigDivision(t *testing.T) {
	ivrConfigResourceLabel1 := "test-ivrconfig1"
	ivrConfigName := "terraform-ivrconfig-" + uuid.NewString()
	ivrConfigDescription := "Terraform IVR config"
	number1 := "+14175550013"
	number2 := "+14175550014"
	divResourceLabel1 := "auth-division1"
	divResourceLabel2 := "auth-division2"
	divName1 := "TerraformDiv-" + uuid.NewString()
	divName2 := "TerraformDiv-" + uuid.NewString()
	ivrConfigDnis := []string{number1, number2}
	didPoolResourceLabel1 := "test-didpool1"

	fullResourceLabel := ResourceType + "." + ivrConfigResourceLabel1

	// did pool cleanup
	resp, err := didPool.DeleteDidPoolWithStartAndEndNumber(context.Background(), number1, number2, sdkConfig)
	if err != nil {
		respStr := "<nil>"
		if resp != nil {
			respStr = strconv.Itoa(resp.StatusCode)
		}
		t.Logf("Failed to delete DID pool: %s. API Response: %s", err.Error(), respStr)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateAuthDivisionResourceForIvrTests(
					divResourceLabel1,
					divName1,
					util.NullValue, // No description
					util.NullValue, // Not home division
				) + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceLabel: ivrConfigResourceLabel1,
					Name:          ivrConfigName,
					Description:   ivrConfigDescription,
					Dnis:          nil, // No dnis
					DependsOn:     "",  // No depends_on
					DivisionId:    "genesyscloud_auth_division." + divResourceLabel1 + ".id",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceLabel, "name", ivrConfigName),
					resource.TestCheckResourceAttr(fullResourceLabel, "description", ivrConfigDescription),
					resource.TestCheckResourceAttr(fullResourceLabel, "dnis.#", "0"),
					resource.TestCheckResourceAttrPair(fullResourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel1, "id"),
				),
			},
			{
				// Update with new DNIS and division
				Config: generateAuthDivisionResourceForIvrTests(
					divResourceLabel1,
					divName1,
					util.NullValue, // No description
					util.NullValue, // Not home division
				) + generateAuthDivisionResourceForIvrTests(
					divResourceLabel2,
					divName2,
					util.NullValue, // No description
					util.NullValue, // Not home division
				) + didPool.GenerateDidPoolResource(&didPool.DidPoolStruct{
					ResourceLabel:    didPoolResourceLabel1,
					StartPhoneNumber: ivrConfigDnis[0],
					EndPhoneNumber:   ivrConfigDnis[1],
					Description:      util.NullValue, // No description
					Comments:         util.NullValue, // No comments
					PoolProvider:     util.NullValue, // No provider
				}) + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceLabel: ivrConfigResourceLabel1,
					Name:          ivrConfigName,
					Description:   ivrConfigDescription,
					Dnis:          ivrConfigDnis,
					DependsOn:     "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceLabel1,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceLabel, "name", ivrConfigName),
					resource.TestCheckResourceAttr(fullResourceLabel, "description", ivrConfigDescription),
					resource.TestCheckResourceAttrPair(fullResourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel1, "id"),
					resource.TestCheckResourceAttr(fullResourceLabel, "dnis.#", "2"),
					util.ValidateStringInArray(fullResourceLabel, "dnis", ivrConfigDnis[0]),
					util.ValidateStringInArray(fullResourceLabel, "dnis", ivrConfigDnis[1]),
				),
			},
			{
				// Import/Read
				ResourceName:      fullResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: generateAuthDivisionResourceForIvrTests(
					divResourceLabel1,
					divName1,
					util.NullValue, // No description
					util.NullValue, // Not home division
				) + generateAuthDivisionResourceForIvrTests(
					divResourceLabel2,
					divName2,
					util.NullValue, // No description
					util.NullValue, // Not home division
				),
			},
		},
		CheckDestroy: testVerifyIvrConfigsDestroyed,
	})
}

func TestAccResourceArchitectIvrConfigDnisOverload(t *testing.T) {
	var (
		resourceLabel = "ivr"
		name          = "TF Test IVR " + uuid.NewString()

		didRangeLength       = 100 // Should be at least 50 to avoid index out of bounds errors below
		didPoolResourceLabel = "did_pool"
		startNumber          = 35375550120
		endNumber            = startNumber + didRangeLength
		startNumberStr       = fmt.Sprintf("+%v", startNumber)
		endNumberStr         = fmt.Sprintf("+%v", endNumber)
	)

	/*
		To avoid clashes, try to get final existing did number and create a pool outside that range
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

	didPoolResource := didPool.GenerateDidPoolResource(&didPool.DidPoolStruct{
		ResourceLabel:    didPoolResourceLabel,
		StartPhoneNumber: startNumberStr,
		EndPhoneNumber:   endNumberStr,
		Description:      util.NullValue, // No description
		Comments:         util.NullValue, // No comments
		PoolProvider:     util.NullValue, // No provider
	})

	// did pool cleanup
	resp, err := didPool.DeleteDidPoolWithStartAndEndNumber(context.Background(), startNumberStr, endNumberStr, sdkConfig)
	if err != nil {
		respStr := "<nil>"
		if resp != nil {
			respStr = strconv.Itoa(resp.StatusCode)
		}
		t.Logf("Failed to delete DID pool: %s. API Response: %s", err.Error(), respStr)
	}

	resourcePath := ResourceType + "." + resourceLabel

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: didPoolResource + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceLabel: resourceLabel,
					Name:          name,
					Description:   "",
					Dnis:          createStringArrayOfPhoneNumbers(startNumber, startNumber+20),
					DependsOn:     "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceLabel,
					DivisionId:    "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", name),
					resource.TestCheckResourceAttr(resourcePath, "dnis.#", "20"),
				),
			},
			{
				Config: didPoolResource + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceLabel: resourceLabel,
					Name:          name,
					Description:   "",
					Dnis:          createStringArrayOfPhoneNumbers(startNumber, startNumber+48),
					DependsOn:     "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceLabel,
					DivisionId:    "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", name),
					resource.TestCheckResourceAttr(resourcePath, "dnis.#", "48"),
				),
			},
			{
				Config: didPoolResource + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceLabel: resourceLabel,
					Name:          name,
					Description:   "",
					Dnis:          createStringArrayOfPhoneNumbers(startNumber, startNumber+12),
					DependsOn:     "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceLabel,
					DivisionId:    "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", name),
					resource.TestCheckResourceAttr(resourcePath, "dnis.#", "12"),
				),
			},
			{
				Config: didPoolResource + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceLabel: resourceLabel,
					Name:          name,
					Description:   "",
					Dnis:          createStringArrayOfPhoneNumbers(startNumber, endNumber),
					DependsOn:     "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceLabel,
					DivisionId:    "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", name),
					resource.TestCheckResourceAttr(resourcePath, "dnis.#", fmt.Sprintf("%v", len(allNumbers))),
				),
			},
			{
				// Import/Read
				ResourceName:      resourcePath,
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

func testVerifyIvrConfigsDestroyed(state *terraform.State) error {
	architectApi := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		ivrConfig, resp, err := architectApi.GetArchitectIvr(rs.Primary.ID)
		if ivrConfig != nil && ivrConfig.State != nil && *ivrConfig.State == "deleted" {
			continue
		}

		if ivrConfig != nil {
			return fmt.Errorf("IVR config (%s) still exists", rs.Primary.ID)
		}

		if util.IsStatus404(resp) {
			// IVR Config not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All IVR Config pool destroyed
	return nil
}

func createStringArrayOfPhoneNumbers(from, to int) []string {
	var slice []string
	for i := 0; i < to-from; i++ {
		slice = append(slice, fmt.Sprintf("+%v", from+i))
	}
	return slice
}

func getLastDidNumberAsInteger() (int, error) {
	config, err := provider.AuthorizeSdk()
	if err != nil {
		return 0, fmt.Errorf("failed to authorize client: %v", err)
	}
	api := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(config)

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
	const (
		varType  = "ASSIGNED_AND_UNASSIGNED"
		pageSize = 100
	)
	var result *platformclientv2.Didnumberentitylisting
	result, response, err := api.GetTelephonyProvidersEdgesDidpoolsDids(varType, []string{}, "", pageSize, pageNumber, "")
	if err != nil {
		return result, err
	}
	if response.Error != nil {
		return result, fmt.Errorf("Response error: %v", response.Error)
	}
	return result, nil
}

// TODO: When the auth division resource is moved to its own package, reference the generate function there and remove this one.
func generateAuthDivisionResourceForIvrTests(
	resourceLabel string,
	name string,
	description string,
	home string) string {
	return fmt.Sprintf(`resource "genesyscloud_auth_division" "%s" {
		name = "%s"
		description = %s
		home = %s
	}
	`, resourceLabel, name, description, home)
}
