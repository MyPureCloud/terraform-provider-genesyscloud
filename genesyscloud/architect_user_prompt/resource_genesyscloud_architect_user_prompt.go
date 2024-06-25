package architect_user_prompt

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getAllUserPrompts(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getArchitectUserPromptProxy(clientConfig)

	userPrompts, resp, err, _ := proxy.getAllArchitectUserPrompts(ctx, false, false, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("failed to get user prompts: %s", err), resp)
	}
	for _, userPrompt := range *userPrompts {
		resources[*userPrompt.Id] = &resourceExporter.ResourceMeta{Name: *userPrompt.Name}
	}

	return resources, nil
}

func createUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectUserPromptProxy(sdkConfig)

	prompt := platformclientv2.Prompt{
		Name: &name,
	}

	if description != "" {
		prompt.Description = &description
	}

	log.Printf("Creating user prompt %s", name)
	userPrompt, resp, err := proxy.createArchitectUserPrompt(ctx, prompt)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create user prompt %s: %s", name, err), resp)
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
			userPromptResource, resp, err := proxy.createArchitectUserPromptResource(ctx, *userPrompt.Id, promptResource)
			if err != nil {
				return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create user prompt resource %s: %s", name, err), resp)
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
				return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to upload user prompt resource %s", name), err)
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
	proxy := getArchitectUserPromptProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectUserPrompt(), constants.DefaultConsistencyChecks, resourceName)

	log.Printf("Reading User Prompt %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		userPrompt, resp, getErr, retryable := proxy.getArchitectUserPrompt(ctx, d.Id(), true, true, nil)
		if retryable {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read User Prompt %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("failed to read User Prompt %s | error: %s", d.Id(), getErr), resp))
		}

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
		return cc.CheckState(d)
	})
}

func updateUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectUserPromptProxy(sdkConfig)

	prompt := platformclientv2.Prompt{
		Name: &name,
	}

	if description != "" {
		prompt.Description = &description
	}

	log.Printf("Updating user prompt %s", name)
	_, resp, err := proxy.updateArchitectUserPrompt(ctx, d.Id(), prompt)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update user prompt %s: %s", name, err), resp)
	}

	diagErr := updatePromptResource(ctx, d, proxy, sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated User Prompt %s", d.Id())
	return readUserPrompt(ctx, d, meta)
}

func deleteUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectUserPromptProxy(sdkConfig)

	log.Printf("Deleting user prompt %s", name)
	if resp, err := proxy.deleteArchitectUserPrompt(ctx, d.Id(), true); err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete user prompt %s: %s", name, err), resp)
	}
	log.Printf("Deleted user prompt %s", name)

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err, retryable := proxy.getArchitectUserPrompt(ctx, d.Id(), false, false, nil)
		if retryable {
			if resp != nil && resp.StatusCode == 404 {
				// User prompt deleted
				log.Printf("Deleted user prompt %s", name)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error deleting user prompt %s | error: %s", name, err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("user prompt %s still exists", name), resp))
	})
}

func updatePromptResource(ctx context.Context, d *schema.ResourceData, proxy *architectUserPromptProxy, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	name := d.Get("name").(string)

	// Get the prompt so we can get existing prompt resources
	userPrompt, resp, err, _ := proxy.getArchitectUserPrompt(ctx, d.Id(), true, true, nil)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get user prompt %s: %s", d.Id(), err), resp)
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
				res, resp, err := proxy.updateArchitectUserPromptResource(ctx, *userPrompt.Id, resourceLanguage, promptResource)
				if err != nil {
					return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create user prompt resource %s: %s", name, err), resp)
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
				res, resp, err := proxy.createArchitectUserPromptResource(ctx, *userPrompt.Id, promptResource)
				if err != nil {
					return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create user prompt resource %s: %s", name, err), resp)
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
				return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to upload user prompt resource %s", name), err)
			}

			log.Printf("Successfully uploaded user prompt resource for language: %s", resourceLanguage)
		}
	}

	return nil
}

func getArchitectPromptAudioData(ctx context.Context, promptId string, meta interface{}) ([]PromptAudioData, error) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectUserPromptProxy(sdkConfig)
	data, _, err, _ := proxy.getArchitectUserPrompt(ctx, promptId, true, true, nil)
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
