package architect_grammar

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v112/platformclientv2"
	"log"
	"os"
	"path"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"
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
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkLanguage.VoiceFileMetadata, languageMap, "voice_file_data", buildGrammarLanguageFileMetadata)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkLanguage.DtmfFileMetadata, languageMap, "dtmf_file_data", buildGrammarLanguageFileMetadata)

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
	resourcedata.BuildSDKStringValueIfNotNil(&sdkMetadata.FileType, metadataMap, "file_type")

	// Get the current date time, helpful in the UI
	currentTime := time.Now().UTC()
	formattedTime := currentTime.Format("2006-01-02T15:04:05.999Z")
	parsedTime, err := time.Parse("2006-01-02T15:04:05.999Z", formattedTime)
	if err != nil {
		log.Printf("Unable to get current date time %s", err)
	} else {
		sdkMetadata.DateUploaded = &parsedTime
	}

	return &sdkMetadata
}

// flattenGrammarLanguages maps a Genesys Cloud *[]platformclientv2.Grammarlanguage into a []interface{}
func flattenGrammarLanguages(d *schema.ResourceData, languages *[]platformclientv2.Grammarlanguage) []interface{} {
	if len(*languages) == 0 {
		return nil
	}

	var languageList []interface{}
	for _, language := range *languages {
		languageMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(languageMap, "language", language.Language)
		if language.VoiceFileMetadata != nil {
			languageMap["voice_file_data"] = flattenGrammarLanguageFileMetadata(d, language.VoiceFileMetadata, *language.Language, "voice")
		}
		if language.DtmfFileMetadata != nil {
			languageMap["dtmf_file_data"] = flattenGrammarLanguageFileMetadata(d, language.DtmfFileMetadata, *language.Language, "dtmf")
		}

		languageList = append(languageList, languageMap)
	}

	return languageList
}

func flattenGrammarLanguageFileMetadata(d *schema.ResourceData, fileMetadata *platformclientv2.Grammarlanguagefilemetadata, languageCode string, fileType string) []interface{} {
	if fileMetadata == nil {
		return nil
	}

	metadataMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(metadataMap, "file_name", fileMetadata.FileName)
	resourcedata.SetMapValueIfNotNil(metadataMap, "file_type", fileMetadata.FileType)
	if schemaResource, ok := d.Get("languages").([]interface{}); ok {
		for _, language := range schemaResource {
			if languageMap, ok := language.(map[string]interface{}); ok {
				if fmt.Sprintf("%v", languageMap["language"]) != languageCode {
					continue
				}
				if fileType == "voice" {
					if voiceData, ok := languageMap["voice_file_data"].([]interface{}); ok {
						voiceDataMap := voiceData[0].(map[string]interface{})
						if hash, ok := voiceDataMap["file_content_hash"].(string); ok {
							metadataMap["file_content_hash"] = hash
						}
					}
				} else if fileType == "dtmf" {
					if dtmfData, ok := languageMap["dtmf_file_data"].([]interface{}); ok {
						dtmfDataMap := dtmfData[0].(map[string]interface{})
						if hash, ok := dtmfDataMap["file_content_hash"].(string); ok {
							metadataMap["file_content_hash"] = hash
						}
					}
				}
			}
		}
	}

	return []interface{}{metadataMap}
}

func ArchitectGrammarResolver(grammarId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}) error {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getArchitectGrammarProxy(sdkConfig)

	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}

	ctx := context.Background()
	grammar, _, err := proxy.getArchitectGrammarById(ctx, grammarId)
	if err != nil {
		return err
	}

	if grammar.Languages == nil {
		return nil
	}

	for _, language := range *grammar.Languages {
		if language.VoiceFileMetadata != nil && language.VoiceFileUrl != nil {
			fileType := ""
			if language.VoiceFileMetadata.FileType != nil {
				fileType = strings.ToLower(*language.VoiceFileMetadata.FileType)
			}
			voiceFileName := fmt.Sprintf("%s-voice-%s.%s", *language.Language, grammarId, fileType)
			if err := files.DownloadExportFile(fullPath, voiceFileName, *language.VoiceFileUrl); err != nil {
				return err
			}
		}

		if language.DtmfFileMetadata != nil && language.DtmfFileUrl != nil {
			fileType := ""
			if language.DtmfFileMetadata.FileType != nil {
				fileType = strings.ToLower(*language.DtmfFileMetadata.FileType)
			}
			dtmfFileName := fmt.Sprintf("%s-dtmf-%s.%s", *language.Language, grammarId, fileType)
			if err := files.DownloadExportFile(fullPath, dtmfFileName, *language.DtmfFileUrl); err != nil {
				return err
			}
		}
	}
	updateFilenamesInExportConfigMap(configMap, grammarId, *grammar.Languages, subDirectory)
	return nil
}

func updateFilenamesInExportConfigMap(configMap map[string]interface{}, grammarId string, languagesSdk []platformclientv2.Grammarlanguage, subDir string) {
	if languagesExporter, ok := configMap["languages"].([]interface{}); ok {
		// Loop through each language in the exporter map to find current language
		for _, languageExporter := range languagesExporter {
			if language, ok := languageExporter.(map[string]interface{}); ok {
				// Get current language code
				if languageCode, ok := language["language"].(string); ok {
					for _, languageSdk := range languagesSdk {
						// Check if this language in the exporter map is the same as the one we sent in
						if languageCode == *languageSdk.Language {
							if voiceFileData, ok := language["voice_file_data"].([]interface{}); ok {
								setExporterFileData(voiceFileData, grammarId, languageSdk, subDir, "voice")
							}
							if dtmfFileData, ok := language["dtmf_file_data"].([]interface{}); ok {
								setExporterFileData(dtmfFileData, grammarId, languageSdk, subDir, "dtmf")
							}
						}
					}
				}
			}
		}
	}
}

func setExporterFileData(fileDataMap []interface{}, grammarId string, language platformclientv2.Grammarlanguage, subDir string, fileType string) {
	//Set file name and content hash in the exporter map
	if fileData, ok := fileDataMap[0].(map[string]interface{}); ok {
		fileExtension := ""
		if fileType == "voice" && language.VoiceFileMetadata.FileType != nil {
			fileExtension = strings.ToLower(*language.VoiceFileMetadata.FileType)
		} else if fileType == "dtmf" && language.DtmfFileMetadata.FileType != nil {
			fileExtension = strings.ToLower(*language.DtmfFileMetadata.FileType)
		}

		fileName := fmt.Sprintf("%s-%s-%s.%s", *language.Language, fileType, grammarId, fileExtension)
		fileData["file_name"] = path.Join(subDir, fileName)
		fileData["file_content_hash"] = fmt.Sprintf(`${filesha256("%s")}`, path.Join(subDir, fileName))
		if fileData["file_type"] == nil {
			fileData["file_type"] = ""
		}
	}
}
