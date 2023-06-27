package architect_api

import (
	"fmt"
	"log"
	utillists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

type postArchitectIvrFunc func(*ArchitectIvrProxy, platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error)

type getArchitectIvrFunc func(*ArchitectIvrProxy, string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error)

type putArchitectIvrFunc func(*ArchitectIvrProxy, string, platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error)

type deleteArchitectIvrFunc func(*ArchitectIvrProxy, string) (*platformclientv2.APIResponse, error)

type getArchitectIvrsFunc func(*ArchitectIvrProxy, int, int, string, string, string, string, string) (*platformclientv2.Ivrentitylisting, *platformclientv2.APIResponse, error)

type ArchitectIvrProxy struct {
	Api *platformclientv2.ArchitectApi

	PostArchitectIvr   postArchitectIvrFunc
	GetArchitectIvr    getArchitectIvrFunc
	PutArchitectIvr    putArchitectIvrFunc
	DeleteArchitectIvr deleteArchitectIvrFunc
	GetArchitectIvrs   getArchitectIvrsFunc

	maxDnisPerRequest int

	// functions to perform basic put/post request without chunking logic
	putArchitectIvrBasic  putArchitectIvrFunc
	postArchitectIvrBasic postArchitectIvrFunc
}

func NewArchitectIvrProxy() *ArchitectIvrProxy {
	var architectApi *platformclientv2.ArchitectApi
	return &ArchitectIvrProxy{
		Api: architectApi,

		PostArchitectIvr:   postArchitectIvr,
		GetArchitectIvr:    getArchitectIvr,
		PutArchitectIvr:    putArchitectIvr,
		DeleteArchitectIvr: deleteArchitectIvr,
		GetArchitectIvrs:   getArchitectIvrs,

		maxDnisPerRequest: 50,

		postArchitectIvrBasic: postArchitectIvrBasic,
		putArchitectIvrBasic:  putArchitectIvrBasic,
	}
}

func (a *ArchitectIvrProxy) ConfigureProxyApiInstance(c *platformclientv2.Configuration) {
	a.Api = platformclientv2.NewArchitectApiWithConfig(c)
}

func postArchitectIvr(a *ArchitectIvrProxy, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	isPost := true
	return a.uploadArchitectIvr(isPost, "", ivr)
}

func getArchitectIvr(a *ArchitectIvrProxy, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.Api.GetArchitectIvr(id)
}

func putArchitectIvr(a *ArchitectIvrProxy, id string, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	isPost := false
	return a.uploadArchitectIvr(isPost, id, ivr)
}

func postArchitectIvrBasic(a *ArchitectIvrProxy, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.Api.PostArchitectIvrs(ivr)
}

func putArchitectIvrBasic(a *ArchitectIvrProxy, id string, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	return a.Api.PutArchitectIvr(id, ivr)
}

func deleteArchitectIvr(a *ArchitectIvrProxy, id string) (*platformclientv2.APIResponse, error) {
	return a.Api.DeleteArchitectIvr(id)
}

func getArchitectIvrs(a *ArchitectIvrProxy, pageNumber int, pageSize int, sortBy, sortOrder, name, dnis, scheduleGroup string) (*platformclientv2.Ivrentitylisting, *platformclientv2.APIResponse, error) {
	return a.Api.GetArchitectIvrs(pageNumber, pageSize, sortBy, sortOrder, name, dnis, scheduleGroup)
}

func (a *ArchitectIvrProxy) uploadArchitectIvr(post bool, id string, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	var (
		respIvr    *platformclientv2.Ivr
		resp       *platformclientv2.APIResponse
		err        error
		dnisChunks [][]string
	)

	if ivr.Dnis != nil {
		dnisChunks = utillists.ChunkStringSlice(*ivr.Dnis, a.maxDnisPerRequest)
		if len(dnisChunks) == 1 {
			ivr.Dnis = &dnisChunks[0]
		} else {
			ivr.Dnis = nil
		}
	}

	if post {
		respIvr, resp, err = a.postArchitectIvrBasic(a, ivr)
	} else {
		respIvr, resp, err = a.putArchitectIvrBasic(a, id, ivr)
	}

	if err != nil {
		return respIvr, resp, err
	}

	id = *respIvr.Id

	if len(dnisChunks) > 1 {
		respIvr, resp, err = a.uploadIvrDnisChunks(dnisChunks, id)
		if err != nil {
			return respIvr, resp, err
		}
	}

	return respIvr, resp, err
}

func (a *ArchitectIvrProxy) uploadIvrDnisChunks(dnisChunks [][]string, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	ivr, resp, getErr := a.GetArchitectIvr(a, id)
	if getErr != nil {
		return ivr, resp, fmt.Errorf("Failed to read IVR config %s: %s", id, getErr)
	}

	for i, chunk := range dnisChunks {
		time.Sleep(2 * time.Second)
		log.Printf("Uploading block %v of DID numbers to ivr config %s", i+1, id)
		putIvr, resp, err := a.uploadDnisChunk(*ivr, chunk)
		if err != nil {
			return putIvr, resp, err
		}
		ivr = putIvr
	}

	return ivr, nil, nil
}

func (a *ArchitectIvrProxy) uploadDnisChunk(ivr platformclientv2.Ivr, chunk []string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
	var dnis []string

	if ivr.Dnis != nil && len(*ivr.Dnis) > 0 {
		dnis = append(dnis, *ivr.Dnis...)
	}

	dnis = append(dnis, chunk...)
	ivr.Dnis = &dnis

	log.Printf("Updating IVR config %s", *ivr.Id)
	putIvr, resp, putErr := a.putArchitectIvrBasic(a, *ivr.Id, ivr)
	if putErr != nil {
		return putIvr, resp, fmt.Errorf("Failed to update IVR config %s: %s", *ivr.Id, putErr)
	}
	return putIvr, resp, nil
}
