package architect_api

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v99/platformclientv2"
)

func TestUploadIvrDnisChunksSuccess(t *testing.T) {
	var (
		chunksCombinedLength int
		chunks               [][]string
	)
	chunks = append(chunks, []string{"1", "2", "3", "4"})
	chunks = append(chunks, []string{"5", "6", "7", "8"})
	chunks = append(chunks, []string{"9", "10"})
	for _, c := range chunks {
		chunksCombinedLength += len(c)
	}

	ivrId := uuid.NewString()

	architectIvrProxy := NewArchitectIvrProxy()

	architectIvrProxy.GetArchitectIvr = createMockGetIvrFunc(ivrId, nil, nil)

	architectIvrProxy.PutArchitectIvr = createMockPutIvrFunc(nil)

	ivr, _, err := architectIvrProxy.UploadIvrDnisChunks(architectIvrProxy, chunks, ivrId)
	if err != nil {
		t.Errorf("Expected error to be nil, got '%v'", err)
	}

	if len(*ivr.Dnis) != chunksCombinedLength {
		t.Errorf("Expected length of returned Ivr.Dnis field to be %v, got %v", chunksCombinedLength, len(*ivr.Dnis))
	}
}

func TestUploadIvrDnisChunksError(t *testing.T) {
	var (
		ivrId               = "1234"
		mockGetError        error
		mockPutError        error
		dnisReturnedFromGet []string
		chunks              = [][]string{[]string{"123", "abc"}, []string{"iii", "zzz"}}
	)

	/* Handling errors on GET arch ivr */
	mockGetError = fmt.Errorf("error on proxy.GetArchitectIvr")
	architectProxy := NewArchitectIvrProxy()
	architectProxy.GetArchitectIvr = createMockGetIvrFunc(ivrId, dnisReturnedFromGet, mockGetError)
	architectProxy.PutArchitectIvr = createMockPutIvrFunc(nil)

	_, _, err := architectProxy.UploadIvrDnisChunks(architectProxy, chunks, ivrId)
	if err == nil {
		t.Errorf("Expected non nil error")
	}
	if !strings.Contains(fmt.Sprintf("%v", err), fmt.Sprintf("%v", mockGetError)) {
		t.Errorf("Expected to receive error containing '%v', got '%v'", mockGetError, err)
	}

	/* Handling errors on PUT arch ivr */
	mockPutError = fmt.Errorf("error on proxy.PutArchitectIvr")
	mockGetError = nil
	architectProxy.GetArchitectIvr = createMockGetIvrFunc(ivrId, dnisReturnedFromGet, mockGetError)
	architectProxy.PutArchitectIvr = createMockPutIvrFunc(mockPutError)

	_, _, err = architectProxy.UploadIvrDnisChunks(architectProxy, chunks, ivrId)
	if err == nil {
		t.Errorf("Expected non nil error")
	}
	if !strings.Contains(fmt.Sprintf("%v", err), fmt.Sprintf("%v", mockPutError)) {
		t.Errorf("Expected to receive error containing '%v', got '%v'", mockPutError, err)
	}
}

func createMockGetIvrFunc(ivrId string, dnis []string, err error) getArchitectIvrFunc {
	return func(*platformclientv2.ArchitectApi, string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
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

func createMockPutIvrFunc(err error) putArchitectIvrFunc {
	return func(api *platformclientv2.ArchitectApi, id string, ivr *platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		if err != nil {
			return nil, nil, err
		}
		return ivr, nil, nil
	}
}
