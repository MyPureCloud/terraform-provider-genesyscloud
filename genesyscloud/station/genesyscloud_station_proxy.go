package station

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *stationProxy

type getStationIdByNameFunc func(ctx context.Context, p *stationProxy, stationName string) (stationId string, retryable bool, resp *platformclientv2.APIResponse, err error)

// stationProxy contains all of the methods that call genesys cloud APIs.
type stationProxy struct {
	clientConfig           *platformclientv2.Configuration
	stationsApi            *platformclientv2.StationsApi
	getStationIdByNameAttr getStationIdByNameFunc
}

// newStationProxy initializes the Station proxy with all of the data needed to communicate with Genesys Cloud
func newStationProxy(clientConfig *platformclientv2.Configuration) *stationProxy {
	stationsApi := platformclientv2.NewStationsApiWithConfig(clientConfig)

	return &stationProxy{
		clientConfig:           clientConfig,
		stationsApi:            stationsApi,
		getStationIdByNameAttr: getStationIdByNameFn,
	}
}

// getStationProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getStationProxy(clientConfig *platformclientv2.Configuration) *stationProxy {
	if internalProxy == nil {
		internalProxy = newStationProxy(clientConfig)
	}
	return internalProxy
}

// getStationIdByName retrieves a Genesys Cloud Station ID by Name
func (p *stationProxy) getStationIdByName(ctx context.Context, stationName string) (stationId string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	return p.getStationIdByNameAttr(ctx, p, stationName)
}

// getStationIdByNameFn is an implementation function for retrieving a Station Id by Name
func getStationIdByNameFn(ctx context.Context, p *stationProxy, stationName string) (stationId string, retryable bool, resp *platformclientv2.APIResponse, err error) {
	const pageSize = 100
	stations, resp, err := p.stationsApi.GetStations(pageSize, 1, "", stationName, "", "", "", "")
	if err != nil {
		return "", false, resp, err
	}
	if stations.Entities == nil || len(*stations.Entities) == 0 {
		return "", true, resp, fmt.Errorf("failed to find ID of station '%s'", stationName)
	}

	for _, station := range *stations.Entities {
		if *station.Name == stationName {
			return *station.Id, false, resp, nil
		}
	}

	for pageNum := 2; pageNum <= *stations.PageCount; pageNum++ {
		stations, resp, err := p.stationsApi.GetStations(pageSize, pageNum, "", stationName, "", "", "", "")
		if err != nil {
			return "", false, resp, err
		}
		if stations.Entities == nil {
			return "", true, resp, fmt.Errorf("failed to find ID of station '%s'", stationName)
		}

		for _, station := range *stations.Entities {
			if *station.Name == stationName {
				return *station.Id, false, resp, nil
			}
		}
	}
	return "", true, resp, fmt.Errorf("failed to find ID of station '%s'", stationName)
}
