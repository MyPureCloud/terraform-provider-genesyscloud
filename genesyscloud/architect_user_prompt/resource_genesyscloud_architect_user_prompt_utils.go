package architect_user_prompt

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	files "terraform-provider-genesyscloud/genesyscloud/util/files"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

type PromptAudioData struct {
	Language string
	FileName string
	MediaUri string
}

type UserPromptStruct struct {
	ResourceLabel string
	Name          string
	Description   string
	Resources     []*UserPromptResourceStruct
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
func updateFilenamesInExportConfigMap(configMap map[string]interface{}, audioDataList []PromptAudioData, subDir string, exportDir string, res resourceExporter.ResourceInfo) {
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
			fileNameVal := path.Join(subDir, fileName)
			fileContentVal := fmt.Sprintf(`${filesha256("%s")}`, path.Join(subDir, fileName))
			r["filename"] = fileNameVal
			r["file_content_hash"] = fileContentVal

			if resourceID := findResourceID(res, languageStr); resourceID != "" {
				res.State.Attributes[fmt.Sprintf("resources.%s.%s", resourceID, "filename")] = fileNameVal
				res.State.Attributes[fmt.Sprintf("resources.%s.%s", resourceID, "file_content_hash")] = fileContentVal
				fullPath := path.Join(exportDir, subDir)
				hash, er := files.HashFileContent(path.Join(fullPath, fileName))
				if er != nil {
					log.Printf("Error Calculating Hash '%s' ", er)
				} else {
					res.State.Attributes[fmt.Sprintf("resources.%s.%s", resourceID, "file_content_hash")] = hash
				}
			}

		}
	}
}

// Find the resourceID from the state, return early if found
func findResourceID(resource resourceExporter.ResourceInfo, valt string) string {
	pattern := regexp.MustCompile(`^resources\.(\d+)\.l.*$`)
	for key, value := range resource.State.Attributes {
		if matches := pattern.FindStringSubmatch(key); matches != nil && value == valt {
			return matches[1]
		}
	}
	return ""
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
	`, userPrompt.ResourceLabel,
		userPrompt.Name,
		userPrompt.Description,
		resourcesString,
	)
}

func ArchitectPromptAudioResolver(promptId, exportDirectory, subDirectory string, configMap map[string]any, meta any, resource resourceExporter.ResourceInfo) error {
	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}

	ctx := context.Background()
	allResources, err := getUserPromptResources(ctx, promptId, meta)
	if err != nil {
		return err
	}

	if allResources == nil || len(*allResources) == 0 {
		log.Printf("Found no resources for prompt '%s'. Exiting resolver function.", promptId)
		return nil
	}

	log.Printf("Collecting audio data (mediaUri, language, filename) for resources in prompt '%s'", promptId)
	audioDataList, err := getArchitectPromptAudioData(ctx, promptId, *allResources)
	if err != nil {
		return err
	}
	log.Printf("Found %v resources with downloadable content for prompt '%s'", len(audioDataList), promptId)

	for _, data := range audioDataList {
		log.Printf("Downloading file '%s' from mediaUri", path.Join(fullPath, data.FileName))
		if err := files.DownloadExportFile(fullPath, data.FileName, data.MediaUri); err != nil {
			return err
		}
		log.Println("Successfully downloaded file")
	}
	if len(audioDataList) > 0 {
		log.Printf("Updating filename fields in the resource config to point to newly downloaded data.")
		updateFilenamesInExportConfigMap(configMap, audioDataList, subDirectory, exportDirectory, resource)
	}

	cleanupFilenamesWhereThereIsNoDownloadableData(ctx, promptId, configMap, *allResources)
	return nil
}

// cleanupFilenamesWhereThereIsNoDownloadableData Finds instances where resources.filename has a value
// even though there is no audio file to download, and then it removes the filename key.
func cleanupFilenamesWhereThereIsNoDownloadableData(ctx context.Context, promptId string, configMap map[string]any, existingResources []platformclientv2.Promptasset) {
	log.Printf("Gathering prompt resources whose 'filename' field reference a non-existent file.")
	languagesWithNoFile := getUserPromptResourceLanguagesWithNoAssociatedFiles(ctx, promptId, existingResources)
	if len(languagesWithNoFile) == 0 {
		return
	}
	resources, _ := configMap["resources"].([]any)
	if len(resources) == 0 {
		return
	}
	for _, r := range resources {
		rMap, ok := r.(map[string]any)
		if !ok {
			continue
		}
		for _, language := range languagesWithNoFile {
			if language != rMap["language"].(string) {
				continue
			}
			if filename, _ := rMap["filename"].(string); filename != "" {
				log.Printf("Removing filename '%s' for language '%s' because file does not exist", filename, language)
				rMap["filename"] = nil
			}
		}
	}
}

// getUserPromptResourceLanguagesWithNoAssociatedFiles Collects all the languages associated with a prompt that
// do not have any downloadable content associated with them i.e. no mediaUri or uploadStatus != transcoded
func getUserPromptResourceLanguagesWithNoAssociatedFiles(ctx context.Context, promptId string, allResources []platformclientv2.Promptasset) []string {
	var languagesWithNoAssociatedFiles []string
	for _, r := range allResources {
		hasAssociatedAudioFile := (r.MediaUri != nil && *r.MediaUri != "") && (r.UploadStatus != nil && *r.UploadStatus == "transcoded")
		if hasAssociatedAudioFile {
			continue
		}
		languagesWithNoAssociatedFiles = append(languagesWithNoAssociatedFiles, *r.Language)
	}
	return languagesWithNoAssociatedFiles
}

func getUserPromptResources(ctx context.Context, promptId string, meta any) (*[]platformclientv2.Promptasset, error) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectUserPromptProxy(sdkConfig)

	log.Printf("Reading all resources for user prompt '%s'", promptId)
	allResources, _, err := proxy.getArchitectUserPromptResources(ctx, promptId)
	if err != nil {
		return nil, fmt.Errorf("failed to read resources for prompt '%s': %v", promptId, err)
	}
	resourceCount := 0
	if allResources != nil {
		resourceCount = len(*allResources)
	}
	log.Printf("Successfully read %v resources associated with prompt  '%s'", resourceCount, promptId)

	return allResources, nil
}

func getArchitectPromptAudioData(ctx context.Context, promptId string, allPromptResources []platformclientv2.Promptasset) ([]PromptAudioData, error) {
	var promptResourceData []PromptAudioData

	for _, r := range allPromptResources {
		if r.MediaUri == nil || *r.MediaUri == "" {
			continue
		}
		if r.UploadStatus == nil || *r.UploadStatus != "transcoded" {
			continue
		}
		var promptAudioData PromptAudioData
		promptAudioData.MediaUri = *r.MediaUri
		promptAudioData.Language = *r.Language
		promptAudioData.FileName = fmt.Sprintf("%s-%s.wav", *r.Language, promptId)
		promptResourceData = append(promptResourceData, promptAudioData)
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
