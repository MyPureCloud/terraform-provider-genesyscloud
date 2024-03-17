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
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	files "terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

func getAllUserPrompts(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	architectAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		userPrompts, _, getErr := architectAPI.GetArchitectPrompts(pageNum, pageSize, nil, "", "", "", "", false, false, nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of prompts: %v", getErr)
		}

		if userPrompts.Entities == nil || len(*userPrompts.Entities) == 0 {
			break
		}

		for _, userPrompt := range *userPrompts.Entities {
			resources[*userPrompt.Id] = &resourceExporter.ResourceMeta{Name: *userPrompt.Name}
		}
	}

	return resources, nil
}

func createUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	prompt := platformclientv2.Prompt{
		Name: &name,
	}

	if description != "" {
		prompt.Description = &description
	}

	log.Printf("Creating user prompt %s", name)
	userPrompt, _, err := architectApi.PostArchitectPrompts(prompt)
	if err != nil {
		return diag.Errorf("Failed to create user prompt %s: %s", name, err)
	}

	// Create the prompt resources
	if resources, ok := d.GetOk("resources"); ok && resources != nil {
		promptResources := resources.(*schema.Set).List()
		for _, promptResource := range promptResources {
			resourceMap := promptResource.(map[string]interface{})
			resourceLanguage := resourceMap["language"].(string)

			tag := make(map[string][]string)
			resourceFilenameStr := ""
			if filename, ok := resourceMap["filename"].(string); ok && filename != "" {
				tag["filename"] = []string{filename}
				resourceFilenameStr = filename
			}

			promptResource := platformclientv2.Promptassetcreate{
				Language: &resourceLanguage,
				Tags:     &tag,
			}

			if resourceTtsString, ok := resourceMap["tts_string"].(string); ok && resourceTtsString != "" {
				promptResource.TtsString = &resourceTtsString
			}

			if resourceText, ok := resourceMap["text"].(string); ok && resourceText != "" {
				promptResource.Text = &resourceText
			}

			log.Printf("Creating user prompt resource for language: %s", resourceLanguage)
			userPromptResource, _, err := architectApi.PostArchitectPromptResources(*userPrompt.Id, promptResource)
			if err != nil {
				return diag.Errorf("Failed to create user prompt resource %s: %s", name, err)
			}
			uploadUri := userPromptResource.UploadUri

			if resourceFilenameStr == "" {
				continue
			}

			if err := uploadPrompt(uploadUri, &resourceFilenameStr, sdkConfig); err != nil {
				d.SetId(*userPrompt.Id)
				diagErr := deleteUserPrompt(ctx, d, meta)
				if diagErr != nil {
					log.Printf("Error deleting user prompt resource %s: %v", *userPrompt.Id, diagErr)
				}
				d.SetId("")
				return diag.Errorf("Failed to upload user prompt resource %s: %s", name, err)
			}

			log.Printf("Successfully uploaded user prompt resource for language: %s", resourceLanguage)
		}
	}

	d.SetId(*userPrompt.Id)
	log.Printf("Created user prompt %s %s", name, *userPrompt.Id)
	return readUserPrompt(ctx, d, meta)
}

func readUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading User Prompt %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		userPrompt, resp, getErr := architectAPI.GetArchitectPrompt(d.Id(), true, true, nil)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read User Prompt %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read User Prompt %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectUserPrompt())

		resourcedata.SetNillableValue(d, "name", userPrompt.Name)
		resourcedata.SetNillableValue(d, "description", userPrompt.Description)

		if resourcesSet, ok := d.Get("resources").(*schema.Set); ok && resourcesSet != nil {
			promptResources := resourcesSet.List()
			for _, promptResource := range promptResources {
				resourceMap, ok := promptResource.(map[string]interface{})
				if !ok {
					continue
				}
				resourceFilename, ok := resourceMap["filename"].(string)
				if !ok || resourceFilename == "" {
					continue
				}
				APIResources := *userPrompt.Resources
				for _, APIResource := range APIResources {
					if APIResource.Tags == nil {
						continue
					}
					tags := *APIResource.Tags
					filenameTag, ok := tags["filename"]
					if !ok {
						continue
					}
					if len(filenameTag) > 0 {
						if filenameTag[0] == resourceFilename {
							if *APIResource.UploadStatus != "transcoded" {
								return retry.RetryableError(fmt.Errorf("prompt file not transcoded. User prompt ID: '%s'. Filename: '%s'", d.Id(), resourceFilename))
							}
						}
					}
				}
			}
		}

		_ = d.Set("resources", flattenPromptResources(d, userPrompt.Resources))

		log.Printf("Read Audio Prompt %s %s", d.Id(), *userPrompt.Id)
		return cc.CheckState()
	})
}

func updateUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	prompt := platformclientv2.Prompt{
		Name: &name,
	}

	if description != "" {
		prompt.Description = &description
	}

	log.Printf("Updating user prompt %s", name)
	_, _, err := architectApi.PutArchitectPrompt(d.Id(), prompt)
	if err != nil {
		return diag.Errorf("Failed to update user prompt %s: %s", name, err)
	}

	diagErr := updatePromptResource(d, architectApi, sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated User Prompt %s", d.Id())
	return readUserPrompt(ctx, d, meta)
}

func deleteUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Deleting user prompt %s", name)
	if _, err := architectApi.DeleteArchitectPrompt(d.Id(), true); err != nil {
		return diag.Errorf("Failed to delete user prompt %s: %s", name, err)
	}
	log.Printf("Deleted user prompt %s", name)

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := architectApi.GetArchitectPrompt(d.Id(), false, false, nil)
		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				// User prompt deleted
				log.Printf("Deleted user prompt %s", name)
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting user prompt %s: %s", name, err))
		}
		return retry.RetryableError(fmt.Errorf("User prompt %s still exists", name))
	})
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

func updatePromptResource(d *schema.ResourceData, architectApi *platformclientv2.ArchitectApi, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	name := d.Get("name").(string)

	// Get the prompt so we can get existing prompt resources
	userPrompt, _, err := architectApi.GetArchitectPrompt(d.Id(), true, true, nil)
	if err != nil {
		return diag.Errorf("Failed to get user prompt %s: %s", d.Id(), err)
	}

	// Update the prompt resources
	if resources, ok := d.GetOk("resources"); ok && resources != nil {
		promptResources := resources.(*schema.Set).List()
		for _, promptResource := range promptResources {
			var userPromptResource *platformclientv2.Promptasset
			languageExists := false

			resourceMap := promptResource.(map[string]interface{})
			resourceLanguage := resourceMap["language"].(string)

			tag := make(map[string][]string)
			tag["filename"] = []string{resourceMap["filename"].(string)}

			// Check if language resource already exists
			for _, v := range *userPrompt.Resources {
				if *v.Language == resourceLanguage {
					languageExists = true
					break
				}
			}

			if languageExists {
				// Update existing resource
				promptResource := platformclientv2.Promptasset{
					Language: &resourceLanguage,
					Tags:     &tag,
				}

				resourceTtsString := resourceMap["tts_string"]
				if resourceTtsString != nil || resourceTtsString.(string) != "" {
					strResourceTtsString := resourceTtsString.(string)
					promptResource.TtsString = &strResourceTtsString
				}

				resourceText := resourceMap["text"]
				if resourceText != nil || resourceText.(string) != "" {
					strResourceText := resourceText.(string)
					promptResource.Text = &strResourceText
				}

				log.Printf("Updating user prompt resource for language: %s", resourceLanguage)
				res, _, err := architectApi.PutArchitectPromptResource(*userPrompt.Id, resourceLanguage, promptResource)
				if err != nil {
					return diag.Errorf("Failed to create user prompt resource %s: %s", name, err)
				}

				userPromptResource = res
			} else {
				// Create new resource for language
				promptResource := platformclientv2.Promptassetcreate{
					Language: &resourceLanguage,
					Tags:     &tag,
				}

				resourceTtsString := resourceMap["tts_string"]
				if resourceTtsString != nil || resourceTtsString.(string) != "" {
					strResourceTtsString := resourceTtsString.(string)
					promptResource.TtsString = &strResourceTtsString
				}

				resourceText := resourceMap["text"]
				if resourceText != nil || resourceText.(string) != "" {
					strResourceText := resourceText.(string)
					promptResource.Text = &strResourceText
				}

				log.Printf("Creating user prompt resource for language: %s", resourceLanguage)
				res, _, err := architectApi.PostArchitectPromptResources(*userPrompt.Id, promptResource)
				if err != nil {
					return diag.Errorf("Failed to create user prompt resource %s: %s", name, err)
				}

				userPromptResource = res
			}

			uploadUri := userPromptResource.UploadUri

			resourceFilename := resourceMap["filename"]
			if resourceFilename == nil || resourceFilename.(string) == "" {
				continue
			}
			resourceFilenameStr := resourceFilename.(string)

			if err := uploadPrompt(uploadUri, &resourceFilenameStr, sdkConfig); err != nil {
				return diag.Errorf("Failed to upload user prompt resource %s: %s", name, err)
			}

			log.Printf("Successfully uploaded user prompt resource for language: %s", resourceLanguage)
		}
	}

	return nil
}

func getArchitectPromptAudioData(promptId string, meta interface{}) ([]PromptAudioData, error) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	apiInstance := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	data, _, err := apiInstance.GetArchitectPrompt(promptId, true, true, nil)
	if err != nil {
		return nil, err
	}

	var promptResourceData []PromptAudioData
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

	audioDataList, err := getArchitectPromptAudioData(promptId, meta)
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
