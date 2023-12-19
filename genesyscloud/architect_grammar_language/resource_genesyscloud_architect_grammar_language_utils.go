package architect_grammar_language

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

/*
The resource_genesyscloud_architect_grammar_language_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getArchitectGrammarLanguageFromResourceData maps data from schema ResourceData into a Genesys Cloud platformclientv2.Grammarlanguage
func getArchitectGrammarLanguageFromResourceData(d *schema.ResourceData) platformclientv2.Grammarlanguage {
	return platformclientv2.Grammarlanguage{
		GrammarId:         platformclientv2.String(d.Get("grammar_id").(string)),
		Language:          platformclientv2.String(d.Get("language").(string)),
		VoiceFileMetadata: buildGrammarLanguageFileMetadata(d.Get("voice_file_data").([]interface{})),
		DtmfFileMetadata:  buildGrammarLanguageFileMetadata(d.Get("dtmf_file_data").([]interface{})),
	}
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

func flattenGrammarLanguageFileMetadata(d *schema.ResourceData, fileMetadata *platformclientv2.Grammarlanguagefilemetadata, fileType FileType) []interface{} {
	if fileMetadata == nil {
		return nil
	}

	metadataMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(metadataMap, "file_name", fileMetadata.FileName)
	resourcedata.SetMapValueIfNotNil(metadataMap, "file_type", fileMetadata.FileType)

	if fileType == Voice {
		if voiceData := d.Get("voice_file_data").([]interface{}); len(voiceData) > 0 {
			voiceDataMap := voiceData[0].(map[string]interface{})
			if hash, ok := voiceDataMap["file_content_hash"].(string); ok {
				metadataMap["file_content_hash"] = hash
			}
		}
	}
	if fileType == Dtmf {
		if dtmfData := d.Get("dtmf_file_data").([]interface{}); len(dtmfData) > 0 {
			dtmfDataMap := dtmfData[0].(map[string]interface{})
			if hash, ok := dtmfDataMap["file_content_hash"].(string); ok {
				metadataMap["file_content_hash"] = hash
			}
		}
	}

	return []interface{}{metadataMap}
}

func ArchitectGrammarLanguageResolver(languageId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}) error {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getArchitectGrammarLanguageProxy(sdkConfig)

	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}

	ctx := context.Background()
	grammarId, languageCode := splitLanguageId(languageId)
	language, _, err := proxy.getArchitectGrammarLanguageById(ctx, grammarId, languageCode)
	if err != nil {
		return err
	}

	if language.VoiceFileMetadata != nil && language.VoiceFileUrl != nil {
		if language.VoiceFileMetadata.FileType != nil {
			downloadFiles(grammarId, fullPath, *language.Language, *language.VoiceFileUrl, *language.VoiceFileMetadata.FileType)
		} else {
			downloadFiles(grammarId, fullPath, *language.Language, *language.VoiceFileUrl, "")
		}
	}

	if language.DtmfFileMetadata != nil && language.DtmfFileUrl != nil {
		if language.DtmfFileMetadata.FileType != nil {
			downloadFiles(grammarId, fullPath, *language.Language, *language.DtmfFileUrl, *language.DtmfFileMetadata.FileType)
		} else {
			downloadFiles(grammarId, fullPath, *language.Language, *language.DtmfFileUrl, "")
		}
	}

	updateFilenamesInExportConfigMap(configMap, grammarId, *language, subDirectory)
	return nil
}

func downloadFiles(grammarId string, fullPath string, languageCode string, fileUrl string, fileTypeSdk string) error {
	fileType := ""
	if fileTypeSdk != "" {
		fileType = strings.ToLower(fileTypeSdk)
	}
	dtmfFileName := fmt.Sprintf("%s-dtmf-%s.%s", languageCode, grammarId, fileType)
	if err := files.DownloadExportFile(fullPath, dtmfFileName, fileUrl); err != nil {
		return err
	}
	return nil
}

func updateFilenamesInExportConfigMap(configMap map[string]interface{}, grammarId string, language platformclientv2.Grammarlanguage, subDir string) {
	if voiceFileData, ok := configMap["voice_file_data"].([]interface{}); ok {
		setExporterFileData(voiceFileData, grammarId, language, subDir, Voice)
	}
	if dtmfFileData, ok := configMap["dtmf_file_data"].([]interface{}); ok {
		setExporterFileData(dtmfFileData, grammarId, language, subDir, Dtmf)
	}
}

func setExporterFileData(fileDataMap []interface{}, grammarId string, language platformclientv2.Grammarlanguage, subDir string, fileType FileType) {
	//Set file name and content hash in the exporter map
	if fileData, ok := fileDataMap[0].(map[string]interface{}); ok {
		fileExtension := ""
		if fileType == Voice && language.VoiceFileMetadata.FileType != nil {
			fileExtension = strings.ToLower(*language.VoiceFileMetadata.FileType)
		}
		if fileType == Dtmf && language.DtmfFileMetadata.FileType != nil {
			fileExtension = strings.ToLower(*language.DtmfFileMetadata.FileType)
		}

		fileName := fmt.Sprintf("%s-%v-%s.%s", *language.Language, fileType, grammarId, fileExtension)
		fileData["file_name"] = path.Join(subDir, fileName)
		fileData["file_content_hash"] = fmt.Sprintf(`${filesha256("%s")}`, path.Join(subDir, fileName))
		if fileData["file_type"] == nil {
			fileData["file_type"] = ""
		}
	}
}
