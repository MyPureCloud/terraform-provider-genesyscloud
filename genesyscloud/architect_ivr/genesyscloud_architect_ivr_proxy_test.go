//go:build unit
// +build unit

package architect_ivr

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestUploadIvrDnisChunksSuccess(t *testing.T) {
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

func TestUploadIvrDnisChunksError(t *testing.T) {
	var (
		ivrId             = uuid.NewString()
		mockGetError      = fmt.Errorf("error on proxy.GetArchitectIvr")
		mockPostError     = fmt.Errorf("error on proxy.PostArchitectIvr")
		mockPutError      = fmt.Errorf("error on proxy.PutArchitectIvr")
		dnis              = []string{"123", "abc", "iii", "zzz"}
		maxDnisPerRequest = 2
	)

	ivr := platformclientv2.Ivr{
		Dnis: &dnis,
	}

	architectProxy := newArchitectIvrProxy(nil)
	architectProxy.maxDnisPerRequest = maxDnisPerRequest

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
