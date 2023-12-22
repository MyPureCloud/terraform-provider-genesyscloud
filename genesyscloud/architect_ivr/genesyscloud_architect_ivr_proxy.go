package architect_ivr

import (
	"context"
	"fmt"
	"log"
	utillists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v116/platformclientv2"
)

/*
The genesyscloud_architect_ivr_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.

Each proxy implementation:

1.  Should provide a private package level variable that holds a instance of a proxy class.
2.  A New... constructor function to initialize the proxy object. This constructor should only be used within
    the proxy.
3.  A get private constructor function that the classes in the package can be used to retrieve
    the proxy. This proxy should check to see if the package level proxy instance is nil and
    should initialize it, otherwise it should return the instance
4.  Type definitions for each function that will be used in the proxy.  We use composition here
    so that we can easily provide mocks for testing.
5.  A struct for the proxy that holds an attribute for each function type.
6.  Wrapper methods on each of the elements on the struct.
7.  Function implementations for each function type definition.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *architectIvrProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createArchitectIvrFunc func(context.Context, *architectIvrProxy, platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error)
type getArchitectIvrFunc func(context.Context, *architectIvrProxy, string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error)
type updateArchitectIvrFunc func(context.Context, *architectIvrProxy, string, platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error)
type deleteArchitectIvrFunc func(context.Context, *architectIvrProxy, string) (*platformclientv2.APIResponse, error)
type getAllArchitectIvrsFunc func(context.Context, *architectIvrProxy, string) (*[]platformclientv2.Ivr, error)
type getArchitectIvrIdByNameFunc func(context.Context, *architectIvrProxy, string) (id string, retryable bool, err error)

type getTelephonyDidPoolsDidsFunc func(ctx context.Context, a *architectIvrProxy, number, varType string) (*[]platformclientv2.Didnumber, error)
type validatePhoneNumberFunc func(ctx context.Context, a *architectIvrProxy, number string) error

// architectIvrProxy contains all methods that call genesys cloud APIs.
type architectIvrProxy struct {
	clientConfig *platformclientv2.Configuration
	api          *platformclientv2.ArchitectApi

	createArchitectIvrAttr      createArchitectIvrFunc
	getArchitectIvrAttr         getArchitectIvrFunc
	updateArchitectIvrAttr      updateArchitectIvrFunc
	deleteArchitectIvrAttr      deleteArchitectIvrFunc
	getAllArchitectIvrsAttr     getAllArchitectIvrsFunc
	getArchitectIvrIdByNameAttr getArchitectIvrIdByNameFunc

	maxDnisPerRequest int

	// functions to perform basic put/post request without chunking logic
	updateArchitectIvrBasicAttr updateArchitectIvrFunc
	createArchitectIvrBasicAttr createArchitectIvrFunc

	// needed for number validation
	telephonyApi                 *platformclientv2.TelephonyProvidersEdgeApi
	getTelephonyDidPoolsDidsAttr getTelephonyDidPoolsDidsFunc
	validatePhoneNumberAttr      validatePhoneNumberFunc
}

// newArchitectIvrProxy initializes the proxy with all the data needed to communicate with Genesys Cloud
func newArchitectIvrProxy(clientConfig *platformclientv2.Configuration) *architectIvrProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	telephonyApi := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(clientConfig)
	return &architectIvrProxy{
		clientConfig: clientConfig,
		api:          api,

		createArchitectIvrAttr:      createArchitectIvrFn,
		getArchitectIvrAttr:         getArchitectIvrFn,
		updateArchitectIvrAttr:      updateArchitectIvrFn,
		deleteArchitectIvrAttr:      deleteArchitectIvrFn,
		getAllArchitectIvrsAttr:     getAllArchitectIvrsFn,
		getArchitectIvrIdByNameAttr: getArchitectIvrIdByNameFn,

		maxDnisPerRequest: maxDnisPerRequest,

		createArchitectIvrBasicAttr: createArchitectIvrBasicFn,
		updateArchitectIvrBasicAttr: updateArchitectIvrBasicFn,

		// needed for number validation
		telephonyApi:                 telephonyApi,
		getTelephonyDidPoolsDidsAttr: getTelephonyDidPoolsDidsFn,
		validatePhoneNumberAttr:      validatePhoneNumberFn,
	}
}

// getArchitectIvrProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getArchitectIvrProxy(clientConfig *platformclientv2.Configuration) *architectIvrProxy {
	if internalProxy == nil {
		internalProxy = newArchitectIvrProxy(clientConfig)
	}
	return internalProxy
}

// getAllArchitectIvrs retrieves all Genesys Cloud Architect IVRs
func (a *architectIvrProxy) getAllArchitectIvrs(ctx context.Context, name string) (*[]platformclientv2.Ivr, error) {
	return a.getAllArchitectIvrsAttr(ctx, a, name)
}

// getArchitectIvrIdByName retrieves a Genesys Cloud Architect IVR ID by name
func (a *architectIvrProxy) getArchitectIvrIdByName(ctx context.Context, name string) (string, bool, error) {
	return a.getArchitectIvrIdByNameAttr(ctx, a, name)
}

// createArchitectIvr creates a Genesys Cloud Architect IVR
func (a *architectIvrProxy) createArchitectIvr(ctx context.Context, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.createArchitectIvrAttr(ctx, a, ivr)
}

// createArchitectIvr retrieves a Genesys Cloud Architect IVR by ID (implements chunking logic)
func (a *architectIvrProxy) getArchitectIvr(ctx context.Context, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.getArchitectIvrAttr(ctx, a, id)
}

// updateArchitectIvr updates a Genesys Cloud Architect IVR (implements chunking logic)
func (a *architectIvrProxy) updateArchitectIvr(ctx context.Context, id string, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.updateArchitectIvrAttr(ctx, a, id, ivr)
}

// deleteArchitectIvr deletes a Genesys Cloud Architect IVR
func (a *architectIvrProxy) deleteArchitectIvr(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return a.deleteArchitectIvrAttr(ctx, a, id)
}

// createArchitectIvrBasicFn creates a Genesys Cloud Architect IVR (without chunking logic)
func (a *architectIvrProxy) createArchitectIvrBasic(ctx context.Context, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.createArchitectIvrBasicAttr(ctx, a, ivr)
}

// updateArchitectIvrBasic updates a Genesys Cloud Architect IVR (without chunking logic)
func (a *architectIvrProxy) updateArchitectIvrBasic(ctx context.Context, id string, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.updateArchitectIvrBasicAttr(ctx, a, id, ivr)
}

// getTelephonyDidPoolsDids retrieves all dids that match the number and varType (unassigned or assigned_and_unassigned)
func (a *architectIvrProxy) getTelephonyDidPoolsDids(ctx context.Context, number, varType string) (*[]platformclientv2.Didnumber, error) {
	return a.getTelephonyDidPoolsDidsAttr(ctx, a, number, varType)
}

// validatePhoneNumber validates that the phone number exists and is unassigned
func (a *architectIvrProxy) validatePhoneNumber(ctx context.Context, number string) error {
	return a.validatePhoneNumberAttr(ctx, a, number)
}

// createArchitectIvrFn is an implementation function for creating a Genesys Cloud Architect IVR
func createArchitectIvrFn(ctx context.Context, a *architectIvrProxy, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.uploadArchitectIvrWithChunkingLogic(ctx, true, "", ivr)
}

// getArchitectIvrFn is an implementation function for retrieving a Genesys Cloud Architect IVR by ID
func getArchitectIvrFn(_ context.Context, a *architectIvrProxy, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.api.GetArchitectIvr(id)
}

// updateArchitectIvrFn is an implementation function for updating a Genesys Cloud Architect IVR
func updateArchitectIvrFn(ctx context.Context, a *architectIvrProxy, id string, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.uploadArchitectIvrWithChunkingLogic(ctx, false, id, ivr)
}

// createArchitectIvrBasicFn is an implementation function for performing a basic post of a Genesys Cloud Architect IVR
// without any chunking logic for the dnis field
func createArchitectIvrBasicFn(_ context.Context, a *architectIvrProxy, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.api.PostArchitectIvrs(ivr)
}

// updateArchitectIvrBasicFn is an implementation function for performing a basic put of a Genesys Cloud Architect IVR
// without any chunking logic for the dnis field
func updateArchitectIvrBasicFn(_ context.Context, a *architectIvrProxy, id string, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.api.PutArchitectIvr(id, ivr)
}

// deleteArchitectIvrFn is an implementation function for deleting a Genesys Cloud Architect IVR
func deleteArchitectIvrFn(_ context.Context, a *architectIvrProxy, id string) (*platformclientv2.APIResponse, error) {
	return a.api.DeleteArchitectIvr(id)
}

// getAllArchitectIvrsFn is an implementation function for retrieving all Genesys Cloud Architect IVRs
func getAllArchitectIvrsFn(_ context.Context, a *architectIvrProxy, name string) (*[]platformclientv2.Ivr, error) {
	var (
		allIvrs   []platformclientv2.Ivr
		pageCount int
	)
	const pageSize = 100

	ivrs, _, err := a.api.GetArchitectIvrs(1, pageSize, "", "", name, "", "")
	if err != nil {
		return nil, fmt.Errorf("error requesting page of architect ivrs: %v", err)
	}
	pageCount = *ivrs.PageCount

	if ivrs.Entities != nil && len(*ivrs.Entities) > 0 {
		allIvrs = append(allIvrs, *ivrs.Entities...)
	}

	if pageCount < 2 {
		return &allIvrs, nil
	}

	for pageNum := 2; pageNum <= pageCount; pageNum++ {
		ivrs, _, err := a.api.GetArchitectIvrs(pageNum, pageSize, "", "", name, "", "")
		if err != nil {
			return nil, fmt.Errorf("error requesting page of architect ivrs: %v", err)
		}
		if ivrs.Entities == nil || len(*ivrs.Entities) == 0 {
			break
		}
		allIvrs = append(allIvrs, *ivrs.Entities...)
	}
	return &allIvrs, nil
}

// getArchitectIvrIdByNameFn is an implementation function for retrieving a Genesys Cloud Architect IVR ID by name
func getArchitectIvrIdByNameFn(ctx context.Context, a *architectIvrProxy, name string) (string, bool, error) {
	ivrs, err := getAllArchitectIvrsFn(ctx, a, name)
	if err != nil {
		return "", false, fmt.Errorf("failed to read ivrs: %v", err)
	}
	if ivrs == nil || len(*ivrs) == 0 {
		return "", true, fmt.Errorf("failed to find ivr with name '%s': %v", name, err)
	}
	for _, ivr := range *ivrs {
		if *ivr.Name == name {
			return *ivr.Id, false, nil
		}
	}
	return "", true, fmt.Errorf("failed to find ivr with name '%s': %v", name, err)
}

// uploadArchitectIvrWithChunkingLogic creates/updates an IVR. The function breaks the dnis field into chunks and uploads them in subsequent
// PUTs if the dnis array length is greater than a.maxDnisPerRequest
func (a *architectIvrProxy) uploadArchitectIvrWithChunkingLogic(ctx context.Context, post bool, id string, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	var (
		respIvr    *platformclientv2.Ivr
		resp       *platformclientv2.APIResponse
		err        error
		dnisChunks [][]string
	)

	if ivr.Dnis != nil && len(*ivr.Dnis) > 0 {
		ivr.Dnis, dnisChunks, err = a.getIvrDnisAndChunks(ctx, id, post, &ivr)
		if err != nil {
			return nil, resp, err
		}
	}

	// Perform initial post/put
	if post {
		respIvr, resp, err = a.createArchitectIvrBasic(ctx, ivr)
	} else {
		respIvr, resp, err = a.updateArchitectIvrBasic(ctx, id, ivr)
	}

	if err != nil {
		operation := "update"
		if post {
			operation = "create"
		}
		return respIvr, resp, fmt.Errorf("error performing %s inside function uploadArchitectIvrWithChunkingLogic: %v", operation, err)
	}

	id = *respIvr.Id

	// If there are chunks, call our function to perform each put request
	if len(dnisChunks) > 0 {
		respIvr, resp, err = a.uploadIvrDnisChunks(ctx, dnisChunks, id)
		if err != nil {
			// if the chunking logic failed on create - delete the IVR that was created above
			// Otherwise the CreateContext func will fail and terraform will assume the IVR was never created
			if post {
				if _, deleteErr := a.deleteArchitectIvr(ctx, id); deleteErr != nil {
					log.Printf("failed to delete ivr '%s' after dnis chunking logic failed: %v", id, deleteErr)
				}
			}
			return respIvr, resp, err
		}
	}

	return respIvr, resp, err
}

// uploadIvrDnisChunks loops through our chunks of dnis numbers and calls the uploadDnisChunk function for each.
func (a *architectIvrProxy) uploadIvrDnisChunks(ctx context.Context, dnisChunks [][]string, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	ivr, resp, getErr := a.getArchitectIvr(ctx, id)
	if getErr != nil {
		return ivr, resp, fmt.Errorf("error occured reading ivr '%s' in function uploadIvrDnisChunks: %v", id, getErr)
	}

	for i, chunk := range dnisChunks {
		time.Sleep(2 * time.Second)
		log.Printf("Uploading block %v of DID numbers to ivr config %s", i+1, id)
		// upload current chunk to IVR
		putIvr, resp, err := a.uploadDnisChunk(ctx, *ivr, chunk)
		if err != nil {
			return putIvr, resp, err
		}
		// Update ivr variable to represent the latest state of dnis field
		ivr = putIvr
	}

	return ivr, nil, nil
}

// uploadDnisChunk takes an IVR object and a chunk of dnis numbers as parameters, appends the dnis numbers from the chunk to the
// dnis numbers on the IVR object, and performs a basic PUT request
func (a *architectIvrProxy) uploadDnisChunk(ctx context.Context, ivr platformclientv2.Ivr, chunk []string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	var dnis []string

	if ivr.Dnis != nil && len(*ivr.Dnis) > 0 {
		dnis = append(dnis, *ivr.Dnis...)
	}

	dnis = append(dnis, chunk...)
	ivr.Dnis = &dnis

	log.Printf("Updating IVR config %s", *ivr.Id)
	putIvr, resp, putErr := a.updateArchitectIvrBasic(ctx, *ivr.Id, ivr)
	if putErr != nil {
		return putIvr, resp, fmt.Errorf("error occured updating ivr %s in function uploadDnisChunk: %v", *ivr.Id, putErr)
	}
	return putIvr, resp, nil
}

// getIvrDnisAndChunks returns the dnis array to attach to the ivr on the initial POST/PUT
// along with the chunks to PUT after, if necessary.
func (a *architectIvrProxy) getIvrDnisAndChunks(ctx context.Context, id string, post bool, ivr *platformclientv2.Ivr) (*[]string, [][]string, error) {
	var (
		dnisChunks                [][]string
		dnisSliceForInitialUpload *[]string
	)

	// Create
	if post {
		dnisChunks = utillists.ChunkStringSlice(*ivr.Dnis, a.maxDnisPerRequest)
		dnisSliceForInitialUpload = &dnisChunks[0]
		dnisChunks = dnisChunks[1:] // all chunks after index 0, if they exist
		return dnisSliceForInitialUpload, dnisChunks, nil
	}

	// Update
	// read the ivr to get current dnis array
	currentIvr, _, err := a.getArchitectIvr(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// slice to establish what we're adding
	dnisToAdd := utillists.SliceDifference(*ivr.Dnis, *currentIvr.Dnis)

	// chunk that if necessary
	if len(dnisToAdd) > a.maxDnisPerRequest {
		if err := a.validateDidNumbers(ctx, dnisToAdd); err != nil {
			return nil, nil, err
		}

		dnisChunks = utillists.ChunkStringSlice(dnisToAdd, a.maxDnisPerRequest)
		dnisForInitialCall := removeItemsToBeAddedFromOriginalList(*ivr.Dnis, dnisToAdd)
		// append the first chunk
		dnisForInitialCall = append(dnisForInitialCall, dnisChunks[0]...)
		// return dnis for initial upload along with any chunks that may exist after index 0
		return &dnisForInitialCall, dnisChunks[1:], nil
	}

	// no chunking logic necessary
	return ivr.Dnis, nil, nil
}

// removeItemsToBeAddedFromOriginalList is used to remove the new did numbers from the dnis slice that we collected from the schema
// We do this because we are keeping the new numbers separate in the chunks slice to be uploaded subsequently
func removeItemsToBeAddedFromOriginalList(allDnis []string, toBeAdded []string) []string {
	for _, number := range toBeAdded {
		allDnis = utillists.Remove(allDnis, number)
	}
	return allDnis
}

// validateDidNumbers implements a loop with custom retry logic to validate a list of phone numbers. Each individual number
// is passed to validatePhoneNumber to be validated
func (a *architectIvrProxy) validateDidNumbers(ctx context.Context, numbers []string) error {
	const maxRetries = 3
	for _, number := range numbers {
		for retryCount := 1; ; retryCount++ {
			err := a.validatePhoneNumber(ctx, number)
			if err == nil {
				break
			}
			if retryCount == maxRetries {
				return err
			}
			time.Sleep(3 * time.Second)
		}
	}
	return nil
}

// validatePhoneNumberFn is the implementation function that validates a phone number exists and is unassigned
func validatePhoneNumberFn(ctx context.Context, a *architectIvrProxy, number string) error {
	const assignedAndUnassigned = "ASSIGNED_AND_UNASSIGNED"

	dids, err := a.getTelephonyDidPoolsDids(ctx, number, assignedAndUnassigned)
	if err != nil {
		return fmt.Errorf("failed to query dids to validate n %s: %v", number, err)
	}
	if dids == nil || len(*dids) == 0 {
		return fmt.Errorf("could not find any numbers that match did %s", number)
	}
	for _, n := range *dids {
		if *n.Number == number {
			if *n.Assigned {
				return fmt.Errorf("phone number %s is already assigned. Owner type %s", number, *n.OwnerType)
			}

			// number found and it is unassigned
			return nil
		}
	}

	return fmt.Errorf("could not find did %s", number)
}

// getTelephonyDidPoolsDidsFn implementation function for retrieving all dids that match the number and varType
// (unassigned or assigned_and_unassigned)
func getTelephonyDidPoolsDidsFn(_ context.Context, a *architectIvrProxy, number, varType string) (*[]platformclientv2.Didnumber, error) {
	var (
		pageSize   = 100
		pageCount  int
		allNumbers []platformclientv2.Didnumber
	)
	data, _, err := a.telephonyApi.GetTelephonyProvidersEdgesDidpoolsDids(varType, nil, number, pageSize, 1, "")
	if err != nil || data.Entities == nil || len(*data.Entities) == 0 {
		return &allNumbers, err
	}

	allNumbers = append(allNumbers, *data.Entities...)

	pageCount = *data.PageCount
	for pageNum := 2; pageNum <= pageCount; pageNum++ {
		data, _, err := a.telephonyApi.GetTelephonyProvidersEdgesDidpoolsDids(varType, nil, number, pageSize, pageNum, "")
		if err != nil {
			return nil, err
		}
		if data.Entities == nil || len(*data.Entities) == 0 {
			break
		}
		allNumbers = append(allNumbers, *data.Entities...)
	}

	return &allNumbers, nil
}
