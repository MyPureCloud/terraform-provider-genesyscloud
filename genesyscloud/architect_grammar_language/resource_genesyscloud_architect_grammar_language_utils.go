package architect_grammar_language

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_architect_grammar_language_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getArchitectGrammarLanguageFromResourceData maps data from schema ResourceData into a Genesys Cloud platformclientv2.Grammarlanguage
func getArchitectGrammarLanguageFromResourceData(d *schema.ResourceData) platformclientv2.Grammarlanguage {
	grammarLanguage := platformclientv2.Grammarlanguage{
		GrammarId: platformclientv2.String(d.Get("grammar_id").(string)),
		Language:  platformclientv2.String(d.Get("language").(string)),
	}

	if voiceFileDataList, ok := d.Get("voice_file_data").([]interface{}); ok {
		grammarLanguage.VoiceFileMetadata = buildGrammarLanguageFileMetadata(voiceFileDataList)
	}

	if dtmfFileDataList, ok := d.Get("dtmf_file_data").([]interface{}); ok {
		grammarLanguage.DtmfFileMetadata = buildGrammarLanguageFileMetadata(dtmfFileDataList)
	}

	return grammarLanguage
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

type grammarLanguageDownloader struct {
	configMap             map[string]interface{}
	exportFilesFolderPath string
	grammarId             string
	language              *platformclientv2.Grammarlanguage
	exportFileName        string
	subDirectory          string
	fileUrl               string
	fileExtension         string
	fileType              FileType
}

func ArchitectGrammarLanguageResolver(languageId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}) error {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectGrammarLanguageProxy(sdkConfig)

	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}

	grammarId, languageCode := splitLanguageId(languageId)
	language, _, err := proxy.getArchitectGrammarLanguageById(context.Background(), grammarId, languageCode)
	if err != nil {
		return err
	}

	downloader := grammarLanguageDownloader{
		configMap:             configMap,
		exportFilesFolderPath: fullPath,
		grammarId:             grammarId,
		language:              language,
		subDirectory:          subDirectory,
	}

	return downloader.downloadVoiceAndDtmfFileData()
}

func (d *grammarLanguageDownloader) downloadVoiceAndDtmfFileData() error {
	if err := d.downloadFileData(Voice); err != nil {
		return err
	}
	return d.downloadFileData(Dtmf)
}

func (d *grammarLanguageDownloader) downloadFileData(fileType FileType) error {
	var (
		url         *string
		fileDataKey string
	)

	d.fileType = fileType

	if d.fileType == Voice {
		url = d.language.VoiceFileUrl
		fileDataKey = "voice_file_data"
	} else {
		url = d.language.DtmfFileUrl
		fileDataKey = "dtmf_file_data"
	}

	if url != nil {
		if err := d.downloadLanguageFileAndUpdateConfigMap(*url); err != nil {
			return fmt.Errorf("error downloading %s %s language file for grammar '%s': %v", fileDataKey, *d.language.Language, d.grammarId, err)
		}
	} else {
		// If there are no files to download, we don't need this block in the export resource
		d.configMap[fileDataKey] = nil
	}

	return nil
}

func (d *grammarLanguageDownloader) downloadLanguageFileAndUpdateConfigMap(url string) error {
	d.fileUrl = url
	d.setExportFileName()
	if err := files.DownloadExportFile(d.exportFilesFolderPath, d.exportFileName, d.fileUrl); err != nil {
		return err
	}
	d.updatePathsInExportConfigMap()
	return nil
}

func (d *grammarLanguageDownloader) setExportFileName() {
	d.setLanguageFileExtension()
	fileTypeStr := "dtmf"
	if d.fileType == Voice {
		fileTypeStr = "voice"
	}
	d.exportFileName = fmt.Sprintf("%s-%s-%s.%s", *d.language.Language, fileTypeStr, d.grammarId, d.fileExtension)
}

func (d *grammarLanguageDownloader) setLanguageFileExtension() {
	var fileExtension string
	if d.fileType == Voice {
		if d.language.VoiceFileMetadata != nil && d.language.VoiceFileMetadata.FileType != nil {
			fileExtension = strings.ToLower(*d.language.VoiceFileMetadata.FileType)
		}
	} else {
		if d.language.DtmfFileMetadata != nil && d.language.DtmfFileMetadata.FileType != nil {
			fileExtension = strings.ToLower(*d.language.DtmfFileMetadata.FileType)
		}
	}
	if fileExtension == "" {
		log.Printf("no file type found when exporting grammar language '%s'. Defaulting to .grxml (grammar ID: '%s', language: '%s')", *d.language.Id, *d.language.GrammarId, *d.language.Language)
		fileExtension = "grxml"
	}
	d.fileExtension = fileExtension
}

// updatePathsInExportConfigMap updates fields filename and file_content_hash to point to the files we downloaded to the export directory
func (d *grammarLanguageDownloader) updatePathsInExportConfigMap() {
	var (
		fileDataMapKey string
		filePath       = path.Join(d.subDirectory, d.exportFileName)
	)

	switch d.fileType {
	case Voice:
		fileDataMapKey = "voice_file_data"
	default:
		fileDataMapKey = "dtmf_file_data"
	}

	if fileDataList, ok := d.configMap[fileDataMapKey].([]interface{}); ok {
		if fileDataMap, ok := fileDataList[0].(map[string]interface{}); ok {
			fileDataMap["file_name"] = filePath
			fileDataMap["file_content_hash"] = fmt.Sprintf(`${filesha256("%s")}`, filePath)
			if fileDataMap["file_type"] == nil {
				fileDataMap["file_type"] = ""
			}
		}
	}
}
