//go:build unit
// +build unit

package architect_ivr

import (
	"context"
	"fmt"
	"strings"
	utillists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestUnitUploadIvrDnisChunksSuccess(t *testing.T) {
	var (
		maxDnisPerRequest = 4
		dnis              = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
		ivrId             = uuid.NewString()
	)

	ivr := &platformclientv2.Ivr{
		Dnis: &dnis,
	}

	architectIvrProxy := newArchitectIvrProxy(nil)
	architectIvrProxy.maxDnisPerRequest = maxDnisPerRequest

	architectIvrProxy.getArchitectIvrAttr = createMockGetIvrFunc(ivrId, nil, nil)
	architectIvrProxy.updateArchitectIvrBasicAttr = createMockPutIvrFunc(nil)
	architectIvrProxy.createArchitectIvrBasicAttr = createMockPostIvrFunc(ivrId, nil)

	t.Log("Testing Post Ivr function")
	ivr, _, err := architectIvrProxy.createArchitectIvr(nil, *ivr)
	if err != nil {
		t.Errorf("Expected error to be nil, got '%v'", err)
	}

	if ivr.Dnis == nil {
		t.Errorf("Dnis array returned on ivr is nil")
	} else if len(*ivr.Dnis) != len(dnis) {
		t.Errorf("Expected length of returned Ivr.Dnis field to be %v, got %v", len(dnis), len(*ivr.Dnis))
	}

	ivr = &platformclientv2.Ivr{
		Dnis: &dnis,
		Id:   &ivrId,
	}

	t.Log("Testing Put Ivr function")

	ivr, _, err = architectIvrProxy.updateArchitectIvr(nil, ivrId, *ivr)
	if err != nil {
		t.Errorf("Expected error to be nil, got '%v'", err)
	}

	if ivr.Dnis == nil {
		t.Errorf("Dnis array on returned ivr is nil")
	} else if len(*ivr.Dnis) != len(dnis) {
		t.Errorf("Expected length of returned Ivr.Dnis field to be %v, got %v", len(dnis), len(*ivr.Dnis))
	}

}

type architectIvrUploadErrorTestData struct {
	mockGetFunction  getArchitectIvrFunc
	mockPutFunction  updateArchitectIvrFunc
	mockPostFunction createArchitectIvrFunc
	mockError        error
}

type getIvrDnisAndChunksFuncTestCase struct {
	ivrFromSchema       *platformclientv2.Ivr
	maxPerRequest       int
	post                bool
	dnisReturnedFromGet *[]string

	expectedInitialUploadDnis *[]string
	expectedChunks            [][]string
}

func TestUnitUploadIvrDnisChunksError(t *testing.T) {
	var (
		ivrId             = uuid.NewString()
		mockGetError      = fmt.Errorf("error on proxy.GetArchitectIvr")
		mockPostError     = fmt.Errorf("error on proxy.PostArchitectIvr")
		mockPutError      = fmt.Errorf("error on proxy.PutArchitectIvr")
		maxDnisPerRequest = 2
		dnis              = []string{"123", "abc", "iii", "zzz"}
	)

	ivr := platformclientv2.Ivr{
		Dnis: &dnis,
	}

	architectProxy := newArchitectIvrProxy(nil)
	architectProxy.maxDnisPerRequest = maxDnisPerRequest

	// Will be called on create after a chunk update fails because the ivr will need to be manually taken down, in that case
	architectProxy.deleteArchitectIvrAttr = createMockDeleteIvrFunc(nil)

	testCases := []architectIvrUploadErrorTestData{
		architectIvrUploadErrorTestData{
			mockGetFunction:  createMockGetIvrFunc(ivrId, nil, mockGetError),
			mockPostFunction: createMockPostIvrFunc(ivrId, nil),
			mockPutFunction:  createMockPutIvrFunc(nil),
			mockError:        mockGetError,
		},
		architectIvrUploadErrorTestData{
			mockGetFunction:  createMockGetIvrFunc(ivrId, nil, nil),
			mockPostFunction: createMockPostIvrFunc(ivrId, mockPostError),
			mockPutFunction:  createMockPutIvrFunc(nil),
			mockError:        mockPostError,
		},
		architectIvrUploadErrorTestData{
			mockGetFunction:  createMockGetIvrFunc(ivrId, nil, nil),
			mockPostFunction: createMockPostIvrFunc(ivrId, nil),
			mockPutFunction:  createMockPutIvrFunc(mockPutError),
			mockError:        mockPutError,
		},
	}

	t.Log("Testing error handling on proxy.PostArchitectIvr")
	for _, test := range testCases {
		if test.mockError == mockGetError {
			continue
		}
		architectProxy.getArchitectIvrAttr = test.mockGetFunction
		architectProxy.createArchitectIvrBasicAttr = test.mockPostFunction
		architectProxy.updateArchitectIvrBasicAttr = test.mockPutFunction

		_, _, err := architectProxy.createArchitectIvr(nil, ivr)
		if err == nil {
			t.Errorf("Expected non nil error")
		}
		if !strings.Contains(fmt.Sprintf("%v", err), fmt.Sprintf("%v", test.mockError)) {
			t.Errorf("Expected to receive error containing '%v', got '%v'", test.mockError, err)
		}
	}

	t.Log("Testing error handling on proxy.PutArchitectIvr")
	for _, test := range testCases {
		if test.mockError == mockPostError {
			continue
		}
		architectProxy.getArchitectIvrAttr = test.mockGetFunction
		architectProxy.createArchitectIvrBasicAttr = test.mockPostFunction
		architectProxy.updateArchitectIvrBasicAttr = test.mockPutFunction

		_, _, err := architectProxy.updateArchitectIvr(nil, ivrId, ivr)
		if err == nil {
			t.Errorf("Expected non nil error")
		}
		if !strings.Contains(fmt.Sprintf("%v", err), fmt.Sprintf("%v", test.mockError)) {
			t.Errorf("Expected to receive error containing '%v', got '%v'", test.mockError, err)
		}
	}
}

func TestUnitRemoveItemsToBeAddedFromOriginalList(t *testing.T) {
	type TestCase struct {
		originalList []string
		toAddList    []string
		expected     []string
	}

	testCases := []TestCase{
		{
			originalList: []string{"a", "b", "c", "d", "e", "f"},
			toAddList:    []string{"a", "d", "f"},
			expected:     []string{"b", "c", "e"},
		},
		{
			originalList: []string{"a", "b", "c", "d", "e", "f"},
			toAddList:    []string{},
			expected:     []string{"a", "b", "c", "d", "e", "f"},
		},
		{
			originalList: []string{"a", "b", "c", "d", "e", "f"},
			toAddList:    []string{"a", "b", "c", "d", "e", "f"},
			expected:     []string{},
		},
		{
			originalList: []string{"a", "b", "c", "d", "e", "f"},
			toAddList:    []string{"a", "f"},
			expected:     []string{"b", "c", "d", "e"},
		},
	}

	for _, testCase := range testCases {
		resultingList := removeItemsToBeAddedFromOriginalList(testCase.originalList, testCase.toAddList)
		if !utillists.AreEquivalent(resultingList, testCase.expected) {
			t.Errorf("expected %v, got %v", testCase.expected, resultingList)
		}
	}
}

func TestUnitGetIvrDnisAndChunks(t *testing.T) {
	var (
		ivrId = uuid.NewString()
	)

	testCases := []getIvrDnisAndChunksFuncTestCase{
		// Updates
		{
			ivrFromSchema: &platformclientv2.Ivr{
				Dnis: &[]string{"1", "2", "3", "4"},
			},
			post:                      false,
			maxPerRequest:             3,
			dnisReturnedFromGet:       &[]string{"1", "2", "3"},
			expectedInitialUploadDnis: &[]string{"1", "2", "3", "4"},
			expectedChunks:            [][]string{},
		},
		{
			ivrFromSchema: &platformclientv2.Ivr{
				Dnis: &[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			},
			maxPerRequest:             3,
			post:                      false,
			dnisReturnedFromGet:       &[]string{"1", "2", "3", "4"},
			expectedInitialUploadDnis: &[]string{"1", "2", "3", "4", "5", "6", "7"},
			expectedChunks: [][]string{
				{"8", "9", "10"},
			},
		},
		{
			ivrFromSchema: &platformclientv2.Ivr{
				Dnis: &[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
			},
			maxPerRequest:             3,
			post:                      false,
			dnisReturnedFromGet:       &[]string{"1", "2", "3", "4", "5"},
			expectedInitialUploadDnis: &[]string{"1", "2", "3", "4", "5", "6", "7", "8"},
			expectedChunks: [][]string{
				{"9", "10", "11"},
				{"12"},
			},
		},
		{
			ivrFromSchema: &platformclientv2.Ivr{
				Dnis: &[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
			},
			maxPerRequest:             3,
			post:                      false,
			dnisReturnedFromGet:       &[]string{"1"},
			expectedInitialUploadDnis: &[]string{"1", "2", "3", "4"},
			expectedChunks: [][]string{
				{"5", "6", "7"},
				{"8", "9", "10"},
				{"11", "12"},
			},
		},
		// Creates
		{
			ivrFromSchema: &platformclientv2.Ivr{
				Dnis: &[]string{"1", "2", "3"},
			},
			maxPerRequest:             3,
			post:                      true,
			dnisReturnedFromGet:       nil,
			expectedInitialUploadDnis: &[]string{"1", "2", "3"},
			expectedChunks:            [][]string{},
		},
		{
			ivrFromSchema: &platformclientv2.Ivr{
				Dnis: &[]string{"1", "2", "3", "4"},
			},
			maxPerRequest:             3,
			post:                      true,
			dnisReturnedFromGet:       nil,
			expectedInitialUploadDnis: &[]string{"1", "2", "3"},
			expectedChunks:            [][]string{{"4"}},
		},
		{
			ivrFromSchema: &platformclientv2.Ivr{
				Dnis: &[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"},
			},
			maxPerRequest:             3,
			post:                      true,
			dnisReturnedFromGet:       nil,
			expectedInitialUploadDnis: &[]string{"1", "2", "3"},
			expectedChunks: [][]string{
				{"4", "5", "6"},
				{"7", "8", "9"},
				{"10", "11", "12"},
			},
		},
	}

	for _, testCase := range testCases {
		architectIvrProxy := newArchitectIvrProxy(nil)
		architectIvrProxy.maxDnisPerRequest = testCase.maxPerRequest

		if !testCase.post {
			architectIvrProxy.getArchitectIvrAttr = createMockGetIvrFunc(ivrId, *testCase.dnisReturnedFromGet, nil)
		}

		initialUploadDnis, chunks, err := architectIvrProxy.getIvrDnisAndChunks(context.TODO(), ivrId, testCase.post, testCase.ivrFromSchema)
		if err != nil {
			t.Errorf("expected nil error, got: %v", err)
		}

		if !utillists.AreEquivalent(*initialUploadDnis, *testCase.expectedInitialUploadDnis) {
			t.Errorf("expected initial dnis slice to be %v, got %v", *testCase.expectedInitialUploadDnis, *initialUploadDnis)
		}
		if len(chunks) != len(testCase.expectedChunks) {
			t.Errorf("expected length of chunks array to be %v, got %v", len(testCase.expectedChunks), len(chunks))
		}
		for i := range chunks {
			if !utillists.AreEquivalent(chunks[i], testCase.expectedChunks[i]) {
				t.Errorf("expected chunk item to be %v, got %v", chunks[i], testCase.expectedChunks[i])
			}
		}
	}
}

func createMockGetIvrFunc(ivrId string, dnis []string, err error) getArchitectIvrFunc {
	return func(context.Context, *architectIvrProxy, string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		if err != nil {
			return nil, nil, err
		}
		mockGetIvr := &platformclientv2.Ivr{
			Id:   &ivrId,
			Dnis: &dnis,
		}
		return mockGetIvr, nil, err
	}
}

func createMockPostIvrFunc(ivrId string, err error) createArchitectIvrFunc {
	return func(ctx context.Context, p *architectIvrProxy, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		if err != nil {
			return nil, nil, err
		}
		ivr.Id = &ivrId
		return &ivr, nil, err
	}
}

func createMockPutIvrFunc(err error) updateArchitectIvrFunc {
	return func(ctx context.Context, p *architectIvrProxy, id string, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		if err != nil {
			return nil, nil, err
		}
		ivr.Id = &id
		return &ivr, nil, err
	}
}

func createMockDeleteIvrFunc(errToReturn error) deleteArchitectIvrFunc {
	return func(context.Context, *architectIvrProxy, string) (*platformclientv2.APIResponse, error) {
		return nil, errToReturn
	}
}
