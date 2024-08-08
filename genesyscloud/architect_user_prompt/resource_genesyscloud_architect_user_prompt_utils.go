package architect_user_prompt

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	files "terraform-provider-genesyscloud/genesyscloud/util/files"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

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

		resourcedata.SetMapValueIfNotNil(promptResource, "language", sdkPromptAsset.Language)
		resourcedata.SetMapValueIfNotNil(promptResource, "tts_string", sdkPromptAsset.TtsString)
		resourcedata.SetMapValueIfNotNil(promptResource, "text", sdkPromptAsset.Text)

		if sdkPromptAsset.Tags != nil && len(*sdkPromptAsset.Tags) > 0 {
			t := *sdkPromptAsset.Tags
			promptResource["filename"] = t["filename"][0]
		}

		schemaResources, ok := d.Get("resources").(*schema.Set)
		if !ok {
			continue
		}

		for _, r := range schemaResources.List() {
			rMap, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			language, _ := rMap["language"].(string)
			if language != *sdkPromptAsset.Language {
				continue
			}
			if hash, _ := rMap["file_content_hash"].(string); hash != "" {
				promptResource["file_content_hash"] = hash
			}
		}

		resourceSet.Add(promptResource)
	}
	return resourceSet
}

// updateFilenamesInExportConfigMap replaces (or creates) the filenames key in configMap with the FileName fields in audioDataList
// which point towards the downloaded audio files stored in the export folder.
// Since a language can only appear once in a resources array, we can match resources[n]["language"] with audioDataList[n].Language
func updateFilenamesInExportConfigMap(configMap map[string]interface{}, audioDataList []PromptAudioData, subDir string) {
	resources, _ := configMap["resources"].([]interface{})
	if len(resources) == 0 {
		return
	}
	for _, resource := range resources {
		r, ok := resource.(map[string]interface{})
		if !ok {
			continue
		}
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

func GenerateUserPromptResource(userPrompt *UserPromptStruct) string {
	resourcesString := ``
	for _, p := range userPrompt.Resources {
		var fileContentHash string
		if p.FileContentHash != util.NullValue {
			fullyQualifiedPath, _ := testrunner.NormalizePath(p.FileContentHash)
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
	if err != nil {
		return err
	}

	if len(audioDataList) == 0 {
		log.Printf("No downloadable asset info found for prompt '%s'", promptId)
		return nil
	}

	for _, data := range audioDataList {
		log.Printf("Downloading file '%s' from '%s'", path.Join(fullPath, data.FileName), data.MediaUri)
		if err := files.DownloadExportFile(fullPath, data.FileName, data.MediaUri); err != nil {
			return err
		}
		log.Println("Successfully downloaded file")
	}
	updateFilenamesInExportConfigMap(configMap, audioDataList, subDirectory)
	return nil
}

func getArchitectPromptAudioData(ctx context.Context, promptId string, meta interface{}) ([]PromptAudioData, error) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectUserPromptProxy(sdkConfig)

	var promptResourceData []PromptAudioData

	data, _, err := proxy.getArchitectUserPrompt(ctx, promptId, true, true, nil)
	if err != nil {
		return nil, err
	}

	if data == nil || data.Resources == nil {
		return promptResourceData, nil
	}

	for _, r := range *data.Resources {
		var data PromptAudioData
		if r.MediaUri != nil && *r.MediaUri != "" {
			data.MediaUri = *r.MediaUri
			data.Language = *r.Language
			data.FileName = fmt.Sprintf("%s-%s.wav", *r.Language, promptId)
			promptResourceData = append(promptResourceData, data)
		}
	}

	return promptResourceData, nil
}

func buildUserPromptFromResourceData(d *schema.ResourceData) platformclientv2.Prompt {
	name := d.Get("name").(string)
	prompt := platformclientv2.Prompt{
		Name: &name,
	}
	if description, _ := d.Get("description").(string); description != "" {
		prompt.Description = &description
	}
	return prompt
}

func buildUserPromptResourceForCreate(resourceMap map[string]any) *platformclientv2.Promptassetcreate {
	resourceLanguage := resourceMap["language"].(string)

	tags := make(map[string][]string)
	if filename, _ := resourceMap["filename"].(string); filename != "" {
		tags["filename"] = []string{filename}
	}

	promptResource := platformclientv2.Promptassetcreate{
		Language: &resourceLanguage,
		Tags:     &tags,
	}

	if resourceTtsString, _ := resourceMap["tts_string"].(string); resourceTtsString != "" {
		promptResource.TtsString = &resourceTtsString
	}

	if resourceText, _ := resourceMap["text"].(string); resourceText != "" {
		promptResource.Text = &resourceText
	}

	return &promptResource
}

func buildUserPromptResourceForUpdate(resourceMap map[string]any) *platformclientv2.Promptasset {
	resourceLanguage := resourceMap["language"].(string)

	tags := make(map[string][]string)
	if filename, _ := resourceMap["filename"].(string); filename != "" {
		tags["filename"] = []string{filename}
	}

	promptResource := platformclientv2.Promptasset{
		Language: &resourceLanguage,
		Tags:     &tags,
	}

	if resourceTtsString, _ := resourceMap["tts_string"].(string); resourceTtsString != "" {
		promptResource.TtsString = &resourceTtsString
	}

	if resourceText, _ := resourceMap["text"].(string); resourceText != "" {
		promptResource.Text = &resourceText
	}

	return &promptResource
}
