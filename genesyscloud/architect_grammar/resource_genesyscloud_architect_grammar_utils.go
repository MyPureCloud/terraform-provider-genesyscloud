package architect_grammar

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v109/platformclientv2"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
The resource_genesyscloud_architect_grammar_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getArchitectGrammarFromResourceData maps data from schema ResourceData object to a platformclientv2.Grammar
func getArchitectGrammarFromResourceData(d *schema.ResourceData) platformclientv2.Grammar {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	return platformclientv2.Grammar{
		Name:        &name,
		Description: &description,
		Languages:   buildGrammarLanguages(d.Get("languages").([]interface{})),
	}
}

// buildGrammarLanguages maps a []interface{} into a Genesys Cloud *[]platformclientv2.Grammarlanguage
func buildGrammarLanguages(languages []interface{}) *[]platformclientv2.Grammarlanguage {
	languagesSlice := make([]platformclientv2.Grammarlanguage, 0)

	for _, language := range languages {
		var sdkLanguage platformclientv2.Grammarlanguage
		languageMap, ok := language.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkLanguage.Language, languageMap, "language")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLanguage.VoiceFileUrl, languageMap, "voice_file_url")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkLanguage.VoiceFileMetadata, languageMap, "voice_file_metadata", buildGrammarLanguageFileMetadata)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLanguage.DtmfFileUrl, languageMap, "dtmf_file_url")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkLanguage.DtmfFileMetadata, languageMap, "dtmf_file_metadata", buildGrammarLanguageFileMetadata)

		languagesSlice = append(languagesSlice, sdkLanguage)
	}

	return &languagesSlice
}

func buildGrammarLanguageFileMetadata(fileMetadata []interface{}) *platformclientv2.Grammarlanguagefilemetadata {
	if fileMetadata == nil || len(fileMetadata) <= 0 {
		return nil
	}

	var sdkMetadata platformclientv2.Grammarlanguagefilemetadata
	metadataMap, ok := fileMetadata[0].(map[string]interface{})
	if !ok {
		return nil
	}

	resourcedata.BuildSDKStringValueIfNotNil(&sdkMetadata.FileName, metadataMap, "file_name")
	sdkMetadata.FileSizeBytes = platformclientv2.Int(metadataMap["file_size_bytes"].(int))
	resourcedata.BuildSDKStringValueIfNotNil(&sdkMetadata.DateUploaded, metadataMap, "date_uploaded")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkMetadata.FileType, metadataMap, "file_type")

	return &sdkMetadata
}

// flattenGrammarLanguages maps a Genesys Cloud *[]platformclientv2.Grammarlanguage into a []interface{}
func flattenGrammarLanguages(languages *[]platformclientv2.Grammarlanguage) []interface{} {
	if len(*languages) == 0 {
		return nil
	}

	var languageList []interface{}
	for _, language := range *languages {
		languageMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(languageMap, "language", language.Language)
		resourcedata.SetMapValueIfNotNil(languageMap, "voice_file_url", language.VoiceFileUrl)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(languageMap, "voice_file_metadata", language.VoiceFileMetadata, flattenGrammarLanguageFileMetadata)
		resourcedata.SetMapValueIfNotNil(languageMap, "dtmf_file_url", language.DtmfFileUrl)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(languageMap, "dtmf_file_metadata", language.DtmfFileMetadata, flattenGrammarLanguageFileMetadata)

		languageList = append(languageList, languageMap)
	}

	return languageList
}

func flattenGrammarLanguageFileMetadata(fileMetadata *platformclientv2.Grammarlanguagefilemetadata) []interface{} {
	if fileMetadata == nil {
		return nil
	}

	metadataMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(metadataMap, "file_name", fileMetadata.FileName)
	resourcedata.SetMapValueIfNotNil(metadataMap, "file_size_bytes", fileMetadata.FileSizeBytes)
	resourcedata.SetMapValueIfNotNil(metadataMap, "date_uploaded", fileMetadata.DateUploaded)
	resourcedata.SetMapValueIfNotNil(metadataMap, "file_type", fileMetadata.FileType)

	return []interface{}{metadataMap}
}
