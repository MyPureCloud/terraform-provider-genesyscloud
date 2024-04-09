package responsemanagement_responseasset

import (
	"os"
	"path"
	"terraform-provider-genesyscloud/genesyscloud/provider"
)

func responsemanagementResponseassetResolver(responseAssetId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}) error {
	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}

	responseAssetDataList, err := getResponseAssetData(responseAssetId, meta)

	return err
}

func getResponseAssetData(responseAssetId string, meta interface{}) ([]responseAssetData, error) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRespManagementRespAssetProxy(sdkConfig)

	data, _, err := proxy.responseManagementApi.GetResponsemanagementResponseasset(responseAssetId)
	if err != nil {
		return nil, err
	}

	for _, r := range *data.Responses {
		data.ContentLocation
	}
}
