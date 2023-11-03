package architect_ivr

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"strconv"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	didPool "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	"testing"
)

func TestAccResourceIvrConfigBasic(t *testing.T) {
	ivrConfigResource1 := "test-ivrconfig1"
	ivrConfigName := "terraform-ivrconfig-" + uuid.NewString()
	ivrConfigDescription := "Terraform IVR config"
	number1 := "+14175550011"
	number2 := "+14175550012"
	ivrConfigDnis := []string{number1, number2}
	didPoolResource1 := "test-didpool1"

	// did pool cleanup
	defer func() {
		if _, err := gcloud.AuthorizeSdk(); err != nil {
			return
		}
		ctx := context.TODO()
		_ = didPool.DeleteDidPoolWithStartAndEndNumber(ctx, number1, number2)
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceID:  ivrConfigResource1,
					Name:        ivrConfigName,
					Description: ivrConfigDescription,
					Dnis:        nil, // No dnis
					DependsOn:   "",  // No depends_on
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+ivrConfigResource1, "name", ivrConfigName),
					resource.TestCheckResourceAttr(resourceName+"."+ivrConfigResource1, "description", ivrConfigDescription),
					hasEmptyDnis(resourceName+"."+ivrConfigResource1),
				),
			},
			{
				// Update with new DNIS
				Config: didPool.GenerateDidPoolResource(&didPool.DidPoolStruct{
					ResourceID:       didPoolResource1,
					StartPhoneNumber: ivrConfigDnis[0],
					EndPhoneNumber:   ivrConfigDnis[1],
					Description:      gcloud.NullValue, // No description
					Comments:         gcloud.NullValue, // No comments
					PoolProvider:     gcloud.NullValue, // No provider
				}) + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceID:  ivrConfigResource1,
					Name:        ivrConfigName,
					Description: ivrConfigDescription,
					Dnis:        ivrConfigDnis,
					DependsOn:   "genesyscloud_telephony_providers_edges_did_pool." + didPoolResource1,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+ivrConfigResource1, "name", ivrConfigName),
					resource.TestCheckResourceAttr("genesyscloud_architect_ivr."+ivrConfigResource1, "description", ivrConfigDescription),
					gcloud.ValidateStringInArray("genesyscloud_architect_ivr."+ivrConfigResource1, "dnis", ivrConfigDnis[0]),
					gcloud.ValidateStringInArray("genesyscloud_architect_ivr."+ivrConfigResource1, "dnis", ivrConfigDnis[1]),
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
	ivrConfigDnis := []string{number1, number2}
	didPoolResource1 := "test-didpool1"

	fullResourceId := resourceName + "." + ivrConfigResource1

	// did pool cleanup
	defer func() {
		if _, err := gcloud.AuthorizeSdk(); err != nil {
			return
		}
		ctx := context.TODO()
		_ = didPool.DeleteDidPoolWithStartAndEndNumber(ctx, number1, number2)
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateAuthDivisionResourceForIvrTests(
					divResource1,
					divName1,
					gcloud.NullValue, // No description
					gcloud.NullValue, // Not home division
				) + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceID:  ivrConfigResource1,
					Name:        ivrConfigName,
					Description: ivrConfigDescription,
					Dnis:        nil, // No dnis
					DependsOn:   "",  // No depends_on
					DivisionId:  "genesyscloud_auth_division." + divResource1 + ".id",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceId, "name", ivrConfigName),
					resource.TestCheckResourceAttr(fullResourceId, "description", ivrConfigDescription),
					resource.TestCheckResourceAttrPair(fullResourceId, "division_id", "genesyscloud_auth_division."+divResource1, "id"),
					hasEmptyDnis(resourceName+"."+ivrConfigResource1),
				),
			},
			{
				// Update with new DNIS and division
				Config: generateAuthDivisionResourceForIvrTests(
					divResource1,
					divName1,
					gcloud.NullValue, // No description
					gcloud.NullValue, // Not home division
				) + generateAuthDivisionResourceForIvrTests(
					divResource2,
					divName2,
					gcloud.NullValue, // No description
					gcloud.NullValue, // Not home division
				) + didPool.GenerateDidPoolResource(&didPool.DidPoolStruct{
					ResourceID:       didPoolResource1,
					StartPhoneNumber: ivrConfigDnis[0],
					EndPhoneNumber:   ivrConfigDnis[1],
					Description:      gcloud.NullValue, // No description
					Comments:         gcloud.NullValue, // No comments
					PoolProvider:     gcloud.NullValue, // No provider
				}) + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceID:  ivrConfigResource1,
					Name:        ivrConfigName,
					Description: ivrConfigDescription,
					Dnis:        ivrConfigDnis,
					DependsOn:   "genesyscloud_telephony_providers_edges_did_pool." + didPoolResource1,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceId, "name", ivrConfigName),
					resource.TestCheckResourceAttr(fullResourceId, "description", ivrConfigDescription),
					resource.TestCheckResourceAttrPair(fullResourceId, "division_id", "genesyscloud_auth_division."+divResource1, "id"),
					gcloud.ValidateStringInArray(fullResourceId, "dnis", ivrConfigDnis[0]),
					gcloud.ValidateStringInArray(fullResourceId, "dnis", ivrConfigDnis[1]),
				),
			},
			{
				// Import/Read
				ResourceName:      fullResourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: generateAuthDivisionResourceForIvrTests(
					divResource1,
					divName1,
					gcloud.NullValue, // No description
					gcloud.NullValue, // Not home division
				) + generateAuthDivisionResourceForIvrTests(
					divResource2,
					divName2,
					gcloud.NullValue, // No description
					gcloud.NullValue, // Not home division
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

		didRangeLength    = 200 // Should be at least 50 to avoid index out of bounds errors below
		didPoolResourceId = "did_pool"
		startNumber       = 35375550120
		endNumber         = startNumber + didRangeLength
		startNumberStr    = fmt.Sprintf("+%v", startNumber)
		endNumberStr      = fmt.Sprintf("+%v", endNumber)
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
		ResourceID:       didPoolResourceId,
		StartPhoneNumber: startNumberStr,
		EndPhoneNumber:   endNumberStr,
		Description:      gcloud.NullValue, // No description
		Comments:         gcloud.NullValue, // No comments
		PoolProvider:     gcloud.NullValue, // No provider
	})

	// did pool cleanup
	defer func() {
		if _, err := gcloud.AuthorizeSdk(); err != nil {
			return
		}
		ctx := context.TODO()
		_ = didPool.DeleteDidPoolWithStartAndEndNumber(ctx, startNumberStr, endNumberStr)
	}()

	fullResourceId := resourceName + "." + resourceID

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: didPoolResource + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceID:  resourceID,
					Name:        name,
					Description: "",
					Dnis:        createStringArrayOfPhoneNumbers(startNumber, startNumber+20),
					DependsOn:   "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceId,
					DivisionId:  "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceId, "name", name),
					resource.TestCheckResourceAttr(fullResourceId, "dnis.#", "20"),
				),
			},
			{
				Config: didPoolResource + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceID:  resourceID,
					Name:        name,
					Description: "",
					Dnis:        createStringArrayOfPhoneNumbers(startNumber, startNumber+48),
					DependsOn:   "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceId,
					DivisionId:  "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceId, "name", name),
					resource.TestCheckResourceAttr(fullResourceId, "dnis.#", "48"),
				),
			},
			{
				Config: didPoolResource + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceID:  resourceID,
					Name:        name,
					Description: "",
					Dnis:        createStringArrayOfPhoneNumbers(startNumber, startNumber+12),
					DependsOn:   "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceId,
					DivisionId:  "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceId, "name", name),
					resource.TestCheckResourceAttr(fullResourceId, "dnis.#", "12"),
				),
			},
			{
				Config: didPoolResource + GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceID:  resourceID,
					Name:        name,
					Description: "",
					Dnis:        createStringArrayOfPhoneNumbers(startNumber, endNumber),
					DependsOn:   "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceId,
					DivisionId:  "",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceId, "name", name),
					resource.TestCheckResourceAttr(fullResourceId, "dnis.#", fmt.Sprintf("%v", len(allNumbers))),
				),
			},
			{
				// Import/Read
				ResourceName:      fullResourceId,
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
		if rs.Type != resourceName {
			continue
		}

		ivrConfig, resp, err := architectApi.GetArchitectIvr(rs.Primary.ID)
		if ivrConfig != nil && ivrConfig.State != nil && *ivrConfig.State == "deleted" {
			continue
		}

		if ivrConfig != nil {
			return fmt.Errorf("IVR config (%s) still exists", rs.Primary.ID)
		}

		if gcloud.IsStatus404(resp) {
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
	config, err := gcloud.AuthorizeSdk()
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
	resourceID string,
	name string,
	description string,
	home string) string {
	return fmt.Sprintf(`resource "genesyscloud_auth_division" "%s" {
		name = "%s"
		description = %s
		home = %s
	}
	`, resourceID, name, description, home)
}

/** Unit Test **/
func TestResourceArchitectRead(t *testing.T) {
	tId := uuid.NewString()
	tName := "My Unit Test IVR"
	tDescription := "My Unit Test IVR"
	tDnis := []string{"+920-555-2902", "+920-321-5463"}
	tIDnis := make([]interface{}, len(tDnis))
	for i, v := range tDnis {
		tIDnis[i] = v
	}
	tOpenHoursFlowId := uuid.NewString()
	tClosedHoursFlowId := uuid.NewString()
	tHolidayHoursFlowId := uuid.NewString()
	tScheduleGroupId := uuid.NewString()
	tDivisionId := uuid.NewString()

	archProxy := &architectIvrProxy{}

	archProxy.getArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		ivr := &platformclientv2.Ivr{
			Name:             &tName,
			Description:      &tDescription,
			Dnis:             &tDnis,
			OpenHoursFlow:    &platformclientv2.Domainentityref{Id: &tOpenHoursFlowId},
			ClosedHoursFlow:  &platformclientv2.Domainentityref{Id: &tClosedHoursFlowId},
			HolidayHoursFlow: &platformclientv2.Domainentityref{Id: &tHolidayHoursFlowId},
			ScheduleGroup:    &platformclientv2.Domainentityref{Id: &tScheduleGroupId},
			Division:         &platformclientv2.Writabledivision{Id: &tDivisionId},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return ivr, apiResponse, nil
	}
	internalProxy = archProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceArchitectIvrConfig().Schema

	//Setup a map of values
	resourceDataMap := buildIvrResourceMap(tId, tName, tDescription, tIDnis, tOpenHoursFlowId, tClosedHoursFlowId, tHolidayHoursFlowId, tScheduleGroupId, tDivisionId)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readIvrConfig(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tName, d.Get("name").(string))
	assert.Equal(t, tName, d.Get("description").(string))
	assert.Equal(t, tOpenHoursFlowId, d.Get("open_hours_flow_id").(string))
	assert.Equal(t, tClosedHoursFlowId, d.Get("closed_hours_flow_id").(string))
	assert.Equal(t, tHolidayHoursFlowId, d.Get("holiday_hours_flow_id").(string))
	assert.Equal(t, tScheduleGroupId, d.Get("schedule_group_id").(string))
	assert.Equal(t, tDivisionId, d.Get("division_id").(string))

}

func TestResourceArchitectDeleteStandard(t *testing.T) {
	tId := uuid.NewString()
	tName := "My Unit Test IVR"
	tDescription := "My Unit Test IVR"
	tDnis := []string{"+920-555-2902", "+920-321-5463"}
	tIDnis := make([]interface{}, len(tDnis))
	for i, v := range tDnis {
		tIDnis[i] = v
	}
	tOpenHoursFlowId := uuid.NewString()
	tClosedHoursFlowId := uuid.NewString()
	tHolidayHoursFlowId := uuid.NewString()
	tScheduleGroupId := uuid.NewString()
	tDivisionId := uuid.NewString()

	archProxy := &architectIvrProxy{}

	archProxy.deleteArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	archProxy.getArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
		err := fmt.Errorf("Unable to find targeted IVR: %s", id)
		return nil, apiResponse, err
	}

	internalProxy = archProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceArchitectIvrConfig().Schema

	//Setup a map of values
	resourceDataMap := buildIvrResourceMap(tId, tName, tDescription, tIDnis, tOpenHoursFlowId, tClosedHoursFlowId, tHolidayHoursFlowId, tScheduleGroupId, tDivisionId)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteIvrConfig(ctx, d, gcloud)
	assert.Nil(t, diag)
	assert.Equal(t, tId, d.Id())
}

func TestResourceArchitectDeleteSoftDelete(t *testing.T) {
	tId := uuid.NewString()
	tName := "My Unit Test IVR"
	tDescription := "My Unit Test IVR"
	tDnis := []string{"+920-555-2902", "+920-321-5463"}
	tIDnis := make([]interface{}, len(tDnis))
	for i, v := range tDnis {
		tIDnis[i] = v
	}
	tOpenHoursFlowId := uuid.NewString()
	tClosedHoursFlowId := uuid.NewString()
	tHolidayHoursFlowId := uuid.NewString()
	tScheduleGroupId := uuid.NewString()
	tDivisionId := uuid.NewString()

	archProxy := &architectIvrProxy{}

	archProxy.deleteArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	archProxy.getArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		state := "deleted"
		ivr := &platformclientv2.Ivr{
			Name:             &tName,
			Description:      &tDescription,
			Dnis:             &tDnis,
			OpenHoursFlow:    &platformclientv2.Domainentityref{Id: &tOpenHoursFlowId},
			ClosedHoursFlow:  &platformclientv2.Domainentityref{Id: &tClosedHoursFlowId},
			HolidayHoursFlow: &platformclientv2.Domainentityref{Id: &tHolidayHoursFlowId},
			ScheduleGroup:    &platformclientv2.Domainentityref{Id: &tScheduleGroupId},
			Division:         &platformclientv2.Writabledivision{Id: &tDivisionId},
			State:            &state,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return ivr, apiResponse, nil
	}

	internalProxy = archProxy
	defer func() { internalProxy = nil }()
	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceArchitectIvrConfig().Schema

	//Setup a map of values
	resourceDataMap := buildIvrResourceMap(tId, tName, tDescription, tIDnis, tOpenHoursFlowId, tClosedHoursFlowId, tHolidayHoursFlowId, tScheduleGroupId, tDivisionId)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteIvrConfig(ctx, d, gcloud)
	assert.Nil(t, diag)
	assert.Equal(t, tId, d.Id())

}

func TestResourceArchitectCreate(t *testing.T) {
	tId := uuid.NewString()
	tName := "My Unit Test IVR"
	tDescription := "My Unit Test IVR"
	tDnis := []string{"+920-555-2902", "+920-321-5463"}
	tIDnis := make([]interface{}, len(tDnis))
	for i, v := range tDnis {
		tIDnis[i] = v
	}
	tOpenHoursFlowId := uuid.NewString()
	tClosedHoursFlowId := uuid.NewString()
	tHolidayHoursFlowId := uuid.NewString()
	tScheduleGroupId := uuid.NewString()
	tDivisionId := uuid.NewString()

	archProxy := &architectIvrProxy{}
	archProxy.getArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		ivr := &platformclientv2.Ivr{
			Id:               &tId,
			Name:             &tName,
			Description:      &tDescription,
			Dnis:             &tDnis,
			OpenHoursFlow:    &platformclientv2.Domainentityref{Id: &tOpenHoursFlowId},
			ClosedHoursFlow:  &platformclientv2.Domainentityref{Id: &tClosedHoursFlowId},
			HolidayHoursFlow: &platformclientv2.Domainentityref{Id: &tHolidayHoursFlowId},
			ScheduleGroup:    &platformclientv2.Domainentityref{Id: &tScheduleGroupId},
			Division:         &platformclientv2.Writabledivision{Id: &tDivisionId},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return ivr, apiResponse, nil
	}

	archProxy.createArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tName, *ivr.Name,"ivr.Name check failed in create createArchitectIvrAttr"))
		assert.Equal(t, tDescription, *ivr.Description, "ivr.Description check failed in create createArchitectIvrAttr"))
		//assert.EqualValues(t, tDnis, ivr.Dnis)
		assert.Equal(t, tOpenHoursFlowId, *ivr.OpenHoursFlow.Id, "ivr.OpenHoursFlow.Id check failed in create createArchitectIvrAttr")
		assert.Equal(t, tClosedHoursFlowId, *ivr.ClosedHoursFlow.Id, "ivr.ClosedHoursFlow.Id check failed in create createArchitectIvrAttr")
		assert.Equal(t, tHolidayHoursFlowId, *ivr.HolidayHoursFlow.Id, "ivr.HolidayHoursFlow.Id check failed in create createArchitectIvrAttr")
		assert.Equal(t, tScheduleGroupId, *ivr.ScheduleGroup.Id, "ivr.ScheduleGroup.Id check failed in create createArchitectIvrAttr")
		assert.Equal(t, tDivisionId, *ivr.Division.Id, "ivr.Division.Id check failed in create createArchitectIvrAttr"))

		ivr.Id = &tId

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &ivr, apiResponse, nil
	}

	internalProxy = archProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceArchitectIvrConfig().Schema

	//Setup a map of values
	resourceDataMap := buildIvrResourceMap(tId, tName, tDescription, tIDnis, tOpenHoursFlowId, tClosedHoursFlowId, tHolidayHoursFlowId, tScheduleGroupId, tDivisionId)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := createIvrConfig(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func buildIvrResourceMap(tId string, tName string, tDescription string, tIDnis []interface{}, tOpenHoursFlowId string, tClosedHoursFlowId string, tHolidayHoursFlowId string, tScheduleGroupId string, tDivisionId string) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":                    tId,
		"name":                  tName,
		"description":           tDescription,
		"dnis":                  tIDnis,
		"open_hours_flow_id":    tOpenHoursFlowId,
		"closed_hours_flow_id":  tClosedHoursFlowId,
		"holiday_hours_flow_id": tHolidayHoursFlowId,
		"schedule_group_id":     tScheduleGroupId,
		"division_id":           tDivisionId,
	}
	return resourceDataMap
}
