package architect_user_prompt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"terraform-provider-genesyscloud/genesyscloud/util"
	files "terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

type PromptAudioData struct {
	Language string
	FileName string
	MediaUri string
}

type UserPromptStruct struct {
	ResourceID  string
	Name        string
	Description string
	Resources   []*UserPromptResourceStruct
}

type UserPromptResourceStruct struct {
	Language        string
	Tts_string      string
	Text            string
	Filename        string
	FileContentHash string
}

func flattenPromptResources(d *schema.ResourceData, promptResources *[]platformclientv2.Promptasset) *schema.Set {
	if promptResources == nil || len(*promptResources) == 0 {
		return nil
	}
	resourceSet := schema.NewSet(schema.HashResource(userPromptResource), []interface{}{})
	for _, sdkPromptAsset := range *promptResources {
		promptResource := make(map[string]interface{})

		if sdkPromptAsset.Language != nil {
			promptResource["language"] = *sdkPromptAsset.Language
		}
		if sdkPromptAsset.TtsString != nil {
			promptResource["tts_string"] = *sdkPromptAsset.TtsString
		}
		if sdkPromptAsset.Text != nil {
			promptResource["text"] = *sdkPromptAsset.Text
		}

		if sdkPromptAsset.Tags != nil && len(*sdkPromptAsset.Tags) > 0 {
			t := *sdkPromptAsset.Tags
			promptResource["filename"] = t["filename"][0]
		}

		if schemaResources, ok := d.Get("resources").(*schema.Set); ok {
			schemaResourcesList := schemaResources.List()
			for _, r := range schemaResourcesList {
				if rMap, ok := r.(map[string]interface{}); ok {
					if fmt.Sprintf("%v", rMap["language"]) != *sdkPromptAsset.Language {
						continue
					}
					if hash, ok := rMap["file_content_hash"].(string); ok && hash != "" {
						promptResource["file_content_hash"] = hash
					}
				}
			}
		}

		resourceSet.Add(promptResource)
	}
	return resourceSet
}

// Replace (or create) the filenames key in configMap with the FileName fields in audioDataList
// which point towards the downloaded audio files stored in the export folder.
// Since a language can only appear once in a resources array, we can match resources[n]["language"] with audioDataList[n].Language
func updateFilenamesInExportConfigMap(configMap map[string]interface{}, audioDataList []PromptAudioData, subDir string) {
	if resources, ok := configMap["resources"].([]interface{}); ok && len(resources) > 0 {
		for _, resource := range resources {
			if r, ok := resource.(map[string]interface{}); ok {
				fileName := ""
				languageStr := r["language"].(string)
				for _, data := range audioDataList {
					if data.Language == languageStr {
						fileName = data.FileName
						break
					}
				}
				if fileName != "" {
					r["filename"] = path.Join(subDir, fileName)
					r["file_content_hash"] = fmt.Sprintf(`${filesha256("%s")}`, path.Join(subDir, fileName))
				}
			}
		}
	}
}

func GenerateUserPromptResource(userPrompt *UserPromptStruct) string {
	resourcesString := ``
	for _, p := range userPrompt.Resources {
		var fileContentHash string
		if p.FileContentHash != util.NullValue {
			fullyQualifiedPath, _ := filepath.Abs(p.FileContentHash)
			fileContentHash = fmt.Sprintf(`filesha256("%s")`, fullyQualifiedPath)
		} else {
			fileContentHash = util.NullValue
		}
		resourcesString += fmt.Sprintf(`resources {
			language          = "%s"
			tts_string        = %s
			text              = %s
			filename          = %s
			file_content_hash = %s
		}
        `,
			p.Language,
			p.Tts_string,
			p.Text,
			p.Filename,
			fileContentHash,
		)
	}

	return fmt.Sprintf(`resource "genesyscloud_architect_user_prompt" "%s" {
		name = "%s"
		description = %s
		%s
	}
	`, userPrompt.ResourceID,
		userPrompt.Name,
		userPrompt.Description,
		resourcesString,
	)
}

func ArchitectPromptAudioResolver(promptId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}) error {
	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}

	audioDataList, err := getArchitectPromptAudioData(context.TODO(), promptId, meta)
	if err != nil || len(audioDataList) == 0 {
		return err
	}

	for _, data := range audioDataList {
		if err := files.DownloadExportFile(fullPath, data.FileName, data.MediaUri); err != nil {
			return err
		}
	}
	updateFilenamesInExportConfigMap(configMap, audioDataList, subDirectory)
	return nil
}

func uploadPrompt(uploadUri *string, filename *string, sdkConfig *platformclientv2.Configuration) error {
	reader, file, err := files.DownloadOrOpenFile(*filename)
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		return err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(*filename))
	if err != nil {
		return err
	}

	if file != nil {
		io.Copy(part, file)
	} else {
		io.Copy(part, reader)
	}
	io.Copy(part, file)
	writer.Close()
	request, err := http.NewRequest(http.MethodPost, *uploadUri, body)
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Add("Authorization", sdkConfig.AccessToken)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	log.Printf("Content of upload: %s", content)

	return nil
}
