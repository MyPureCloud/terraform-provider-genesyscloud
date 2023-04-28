package architect_api

import (
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v99/platformclientv2"
)

type postArchitectIvrFunc func(*platformclientv2.ArchitectApi, *platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error)

type getArchitectIvrFunc func(*platformclientv2.ArchitectApi, string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error)

type putArchitectIvrFunc func(*platformclientv2.ArchitectApi, string, *platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error)

type deleteArchitectIvrFunc func(*platformclientv2.ArchitectApi, string) (*platformclientv2.APIResponse, error)

type getArchitectIvrsFunc func(*platformclientv2.ArchitectApi, int, int, string, string, string, string, string) (*platformclientv2.Ivrentitylisting, *platformclientv2.APIResponse, error)

type uploadIvrDnisChunksFunc func(*ArchitectIvrProxy, [][]string, string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error)

type ArchitectIvrProxy struct {
	Api *platformclientv2.ArchitectApi

	PostArchitectIvr   postArchitectIvrFunc
	GetArchitectIvr    getArchitectIvrFunc
	PutArchitectIvr    putArchitectIvrFunc
	DeleteArchitectIvr deleteArchitectIvrFunc
	GetArchitectIvrs   getArchitectIvrsFunc

	UploadIvrDnisChunks uploadIvrDnisChunksFunc
}

func NewArchitectIvrProxy() *ArchitectIvrProxy {
	var architectApi *platformclientv2.ArchitectApi
	return &ArchitectIvrProxy{
		Api: architectApi,
		PostArchitectIvr: func(api *platformclientv2.ArchitectApi, ivr *platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
			return api.PostArchitectIvrs(*ivr)
		},
		GetArchitectIvr: func(api *platformclientv2.ArchitectApi, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
			return api.GetArchitectIvr(id)
		},
		PutArchitectIvr: func(api *platformclientv2.ArchitectApi, id string, ivr *platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
			return api.PutArchitectIvr(id, *ivr)
		},
		DeleteArchitectIvr: func(api *platformclientv2.ArchitectApi, id string) (*platformclientv2.APIResponse, error) {
			return api.DeleteArchitectIvr(id)
		},
		GetArchitectIvrs: func(api *platformclientv2.ArchitectApi, pageNumber int, pageSize int, sortBy, sortOrder, name, dnis, scheduleGroup string) (*platformclientv2.Ivrentitylisting, *platformclientv2.APIResponse, error) {
			return api.GetArchitectIvrs(pageNumber, pageSize, sortBy, sortOrder, name, dnis, scheduleGroup)
		},
		UploadIvrDnisChunks: uploadIvrDnisChunks,
	}
}

func (a *ArchitectIvrProxy) ConfigureProxyApiInstance(c *platformclientv2.Configuration) {
	a.Api = platformclientv2.NewArchitectApiWithConfig(c)
}

func uploadIvrDnisChunks(a *ArchitectIvrProxy, dnisChunks [][]string, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	ivr, resp, getErr := a.GetArchitectIvr(a.Api, id)
	if getErr != nil {
		return ivr, resp, fmt.Errorf("Failed to read IVR config %s: %s", id, getErr)
	}

	for i, chunk := range dnisChunks {
		time.Sleep(2 * time.Second)
		log.Printf("Uploading block %v of DID numbers to ivr config %s", i+1, id)
		putIvr, resp, err := a.uploadDnisChunk(ivr, chunk)
		if err != nil {
			return putIvr, resp, err
		}
		ivr = putIvr
	}

	return ivr, nil, nil
}

func (a *ArchitectIvrProxy) uploadDnisChunk(ivr *platformclientv2.Ivr, chunk []string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	var dnis []string

	if ivr.Dnis != nil && len(*ivr.Dnis) > 0 {
		dnis = append(dnis, *ivr.Dnis...)
	}

	dnis = append(dnis, chunk...)
	ivr.Dnis = &dnis

	log.Printf("Updating IVR config %s", *ivr.Id)
	putIvr, resp, putErr := a.PutArchitectIvr(a.Api, *ivr.Id, ivr)
	if putErr != nil {
		return putIvr, resp, fmt.Errorf("Failed to update IVR config %s: %s", *ivr.Id, putErr)
	}

	return putIvr, resp, nil
}
