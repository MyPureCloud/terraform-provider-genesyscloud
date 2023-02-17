package genesyscloud

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
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

type PromptAudioData struct {
	Language string
	FileName string
	MediaUri string
}

var userPromptResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"language": {
			Description: "Language for the prompt resource. (eg. en-us)",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"af",
				"af-na",
				"af-za",
				"agq",
				"agq-cm",
				"ak",
				"ak-gh",
				"am",
				"am-et",
				"ar",
				"ar-001",
				"ar-ae",
				"ar-bh",
				"ar-dj",
				"ar-dz",
				"ar-eg",
				"ar-eh",
				"ar-er",
				"ar-il",
				"ar-iq",
				"ar-jo",
				"ar-km",
				"ar-kw",
				"ar-lb",
				"ar-ly",
				"ar-ma",
				"ar-mr",
				"ar-om",
				"ar-ps",
				"ar-qa",
				"ar-sa",
				"ar-sd",
				"ar-so",
				"ar-sy",
				"ar-td",
				"ar-tn",
				"ar-ye",
				"arn-cl",
				"as",
				"as-in",
				"asa",
				"asa-tz",
				"az",
				"az-cyrl",
				"az-cyrl-az",
				"az-latn",
				"az-latn-az",
				"ba-ru",
				"bas",
				"bas-cm",
				"be",
				"be-by",
				"bem",
				"bem-zm",
				"bez",
				"bez-tz",
				"bg",
				"bg-bg",
				"bm",
				"bm-ml",
				"bn",
				"bn-bd",
				"bn-in",
				"bo",
				"bo-cn",
				"bo-in",
				"br",
				"br-fr",
				"bs",
				"bs-cyrl",
				"bs-cyrl-ba",
				"bs-latn",
				"bs-latn-ba",
				"ca",
				"ca-ad",
				"ca-es",
				"cgg",
				"cgg-ug",
				"chr",
				"chr-us",
				"co-fr",
				"cs",
				"cs-cz",
				"cy",
				"cy-gb",
				"da",
				"da-dk",
				"dav",
				"dav-ke",
				"de",
				"de-at",
				"de-be",
				"de-ch",
				"de-de",
				"de-li",
				"de-lu",
				"dje",
				"dje-ne",
				"dsb-de",
				"dua",
				"dua-cm",
				"dv-mv",
				"dyo",
				"dyo-sn",
				"dz",
				"dz-bt",
				"ebu",
				"ebu-ke",
				"ee",
				"ee-gh",
				"ee-tg",
				"el",
				"el-cy",
				"el-gr",
				"en",
				"en-029",
				"en-150",
				"en-ag",
				"en-as",
				"en-au",
				"en-bb",
				"en-be",
				"en-bm",
				"en-bs",
				"en-bw",
				"en-bz",
				"en-ca",
				"en-cm",
				"en-dm",
				"en-fj",
				"en-fm",
				"en-gb",
				"en-gd",
				"en-gg",
				"en-gh",
				"en-gi",
				"en-gm",
				"en-gu",
				"en-gy",
				"en-hk",
				"en-ie",
				"en-im",
				"en-in",
				"en-je",
				"en-jm",
				"en-ke",
				"en-ki",
				"en-kn",
				"en-ky",
				"en-lc",
				"en-lr",
				"en-ls",
				"en-mg",
				"en-mh",
				"en-mp",
				"en-mt",
				"en-mu",
				"en-mw",
				"en-my",
				"en-na",
				"en-ng",
				"en-nz",
				"en-pg",
				"en-ph",
				"en-pk",
				"en-pr",
				"en-pw",
				"en-sb",
				"en-sc",
				"en-sg",
				"en-sl",
				"en-ss",
				"en-sz",
				"en-tc",
				"en-to",
				"en-tt",
				"en-tz",
				"en-ug",
				"en-um",
				"en-us",
				"en-vc",
				"en-vg",
				"en-vi",
				"en-vu",
				"en-ws",
				"en-za",
				"en-zm",
				"en-zw",
				"eo",
				"es",
				"es-419",
				"es-ar",
				"es-bo",
				"es-cl",
				"es-co",
				"es-cr",
				"es-cu",
				"es-do",
				"es-ea",
				"es-ec",
				"es-es",
				"es-gq",
				"es-gt",
				"es-hn",
				"es-ic",
				"es-mx",
				"es-ni",
				"es-pa",
				"es-pe",
				"es-ph",
				"es-pr",
				"es-py",
				"es-sv",
				"es-us",
				"es-uy",
				"es-ve",
				"et",
				"et-ee",
				"eu",
				"eu-es",
				"ewo",
				"ewo-cm",
				"fa",
				"fa-af",
				"fa-ir",
				"ff",
				"ff-sn",
				"fi",
				"fi-fi",
				"fil",
				"fil-ph",
				"fo",
				"fo-fo",
				"fr",
				"fr-be",
				"fr-bf",
				"fr-bi",
				"fr-bj",
				"fr-bl",
				"fr-ca",
				"fr-cd",
				"fr-cf",
				"fr-cg",
				"fr-ch",
				"fr-ci",
				"fr-cm",
				"fr-dj",
				"fr-dz",
				"fr-fr",
				"fr-ga",
				"fr-gf",
				"fr-gn",
				"fr-gp",
				"fr-gq",
				"fr-ht",
				"fr-km",
				"fr-lu",
				"fr-ma",
				"fr-mc",
				"fr-mf",
				"fr-mg",
				"fr-ml",
				"fr-mq",
				"fr-mr",
				"fr-mu",
				"fr-nc",
				"fr-ne",
				"fr-pf",
				"fr-re",
				"fr-rw",
				"fr-sc",
				"fr-sn",
				"fr-sy",
				"fr-td",
				"fr-tg",
				"fr-tn",
				"fr-vu",
				"fr-yt",
				"fy-nl",
				"ga",
				"ga-ie",
				"gd-gb",
				"gl",
				"gl-es",
				"gsw",
				"gsw-ch",
				"gsw-fr",
				"gu",
				"gu-in",
				"guz",
				"guz-ke",
				"gv",
				"gv-gb",
				"ha",
				"ha-latn",
				"ha-latn-gh",
				"ha-latn-ne",
				"ha-latn-ng",
				"ha-ng",
				"haw",
				"haw-us",
				"he",
				"he-il",
				"hi",
				"hi-in",
				"hr",
				"hr-ba",
				"hr-hr",
				"hsb-de",
				"hu",
				"hu-hu",
				"hy",
				"hy-am",
				"id",
				"id-id",
				"ig",
				"ig-ng",
				"ii",
				"ii-cn",
				"is",
				"is-is",
				"it",
				"it-ch",
				"it-it",
				"it-sm",
				"ja",
				"ja-jp",
				"jgo",
				"jgo-cm",
				"jmc",
				"jmc-tz",
				"ka",
				"ka-ge",
				"kab",
				"kab-dz",
				"kam",
				"kam-ke",
				"kde",
				"kde-tz",
				"kea",
				"kea-cv",
				"khq",
				"khq-ml",
				"ki",
				"ki-ke",
				"kk",
				"kk-cyrl",
				"kk-cyrl-kz",
				"kk-kz",
				"kl",
				"kl-gl",
				"kln",
				"kln-ke",
				"km",
				"km-kh",
				"kn",
				"kn-in",
				"ko",
				"ko-kp",
				"ko-kr",
				"kok",
				"kok-in",
				"ks",
				"ks-arab",
				"ks-arab-in",
				"ksb",
				"ksb-tz",
				"ksf",
				"ksf-cm",
				"kw",
				"kw-gb",
				"ky-kg",
				"lag",
				"lag-tz",
				"lb-lu",
				"lg",
				"lg-ug",
				"ln",
				"ln-ao",
				"ln-cd",
				"ln-cf",
				"ln-cg",
				"lo",
				"lo-la",
				"lt",
				"lt-lt",
				"lu",
				"lu-cd",
				"luo",
				"luo-ke",
				"luy",
				"luy-ke",
				"lv",
				"lv-lv",
				"mas",
				"mas-ke",
				"mas-tz",
				"mer",
				"mer-ke",
				"mfe",
				"mfe-mu",
				"mg",
				"mg-mg",
				"mgh",
				"mgh-mz",
				"mgo",
				"mgo-cm",
				"mi-nz",
				"mk",
				"mk-mk",
				"ml",
				"ml-in",
				"mn-cn",
				"mn-mn",
				"moh-ca",
				"mr",
				"mr-in",
				"ms",
				"ms-bn",
				"ms-my",
				"ms-sg",
				"mt",
				"mt-mt",
				"mua",
				"mua-cm",
				"my",
				"my-mm",
				"naq",
				"naq-na",
				"nb",
				"nb-no",
				"nd",
				"nd-zw",
				"ne",
				"ne-in",
				"ne-np",
				"nl",
				"nl-aw",
				"nl-be",
				"nl-cw",
				"nl-nl",
				"nl-sr",
				"nmg",
				"nmg-cm",
				"nn",
				"nn-no",
				"nso-za",
				"nus",
				"nus-sd",
				"nyn",
				"nyn-ug",
				"oc-fr",
				"om",
				"om-et",
				"om-ke",
				"or",
				"or-in",
				"pa",
				"pa-arab",
				"pa-arab-pk",
				"pa-guru",
				"pa-guru-in",
				"pa-in",
				"pl",
				"pl-pl",
				"prs-af",
				"ps",
				"ps-af",
				"pt",
				"pt-ao",
				"pt-br",
				"pt-cv",
				"pt-gw",
				"pt-mo",
				"pt-mz",
				"pt-pt",
				"pt-st",
				"pt-tl",
				"qut-gt",
				"quz-bo",
				"quz-ec",
				"quz-pe",
				"rm",
				"rm-ch",
				"rn",
				"rn-bi",
				"ro",
				"ro-md",
				"ro-ro",
				"rof",
				"rof-tz",
				"ru",
				"ru-by",
				"ru-kg",
				"ru-kz",
				"ru-md",
				"ru-ru",
				"ru-ua",
				"rw",
				"rw-rw",
				"rwk",
				"rwk-tz",
				"sa-in",
				"sah-ru",
				"saq",
				"saq-ke",
				"sbp",
				"sbp-tz",
				"se-fi",
				"se-no",
				"se-se",
				"seh",
				"seh-mz",
				"ses",
				"ses-ml",
				"sg",
				"sg-cf",
				"shi",
				"shi-latn",
				"shi-latn-ma",
				"shi-tfng",
				"shi-tfng-ma",
				"si",
				"si-lk",
				"sk",
				"sk-sk",
				"sl",
				"sl-si",
				"sma-no",
				"sma-se",
				"smj-no",
				"smj-se",
				"smn-fi",
				"sms-fi",
				"sn",
				"sn-zw",
				"so",
				"so-dj",
				"so-et",
				"so-ke",
				"so-so",
				"sq",
				"sq-al",
				"sq-mk",
				"sr",
				"sr-cyrl",
				"sr-cyrl-ba",
				"sr-cyrl-cs",
				"sr-cyrl-me",
				"sr-cyrl-rs",
				"sr-latn",
				"sr-latn-ba",
				"sr-latn-cs",
				"sr-latn-me",
				"sr-latn-rs",
				"sv",
				"sv-fi",
				"sv-se",
				"sw",
				"sw-ke",
				"sw-tz",
				"sw-ug",
				"swc",
				"swc-cd",
				"syr-sy",
				"ta",
				"ta-in",
				"ta-lk",
				"ta-my",
				"ta-sg",
				"te",
				"te-in",
				"teo",
				"teo-ke",
				"teo-ug",
				"tg-tj",
				"th",
				"th-th",
				"ti",
				"ti-er",
				"ti-et",
				"tk-tm",
				"tn-za",
				"to",
				"to-to",
				"tr",
				"tr-cy",
				"tr-tr",
				"tt-ru",
				"twq",
				"twq-ne",
				"tzm",
				"tzm-dz",
				"tzm-latn",
				"tzm-latn-ma",
				"ug-cn",
				"uk",
				"uk-ua",
				"ur",
				"ur-in",
				"ur-pk",
				"uz",
				"uz-arab",
				"uz-arab-af",
				"uz-cyrl",
				"uz-cyrl-uz",
				"uz-latn",
				"uz-latn-uz",
				"vai",
				"vai-latn",
				"vai-latn-lr",
				"vai-vaii",
				"vai-vaii-lr",
				"vi",
				"vi-vn",
				"vun",
				"vun-tz",
				"wo-sn",
				"yav",
				"yav-cm",
				"yo",
				"yo-ng",
				"zh",
				"zh-cn",
				"zh-hans",
				"zh-hans-cn",
				"zh-hans-hk",
				"zh-hans-mo",
				"zh-hans-sg",
				"zh-hant",
				"zh-hant-hk",
				"zh-hant-mo",
				"zh-hant-tw",
				"zh-hk",
				"zh-mo",
				"zh-sg",
				"zh-tw",
				"zu",
				"zu-za",
			}, false),
		},
		"tts_string": {
			Description: "Text to Speech (TTS) value for the prompt.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"text": {
			Description: "Text value for the prompt.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"filename": {
			Description: "Path or URL to the file to be uploaded as prompt.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	},
}

func getAllUserPrompts(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	architectAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		userPrompts, _, getErr := architectAPI.GetArchitectPrompts(pageNum, pageSize, nil, "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of prompts: %v", getErr)
		}

		if userPrompts.Entities == nil || len(*userPrompts.Entities) == 0 {
			break
		}

		for _, userPrompt := range *userPrompts.Entities {
			resources[*userPrompt.Id] = &ResourceMeta{Name: *userPrompt.Name}
		}
	}

	return resources, nil
}

func architectUserPromptExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllUserPrompts),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
		CustomFileWriter: CustomFileWriterSettings{
			RetrieveAndWriteFilesFunc: ArchitectPromptAudioResolver,
			SubDirectory:              "audio_prompts",
		},
	}
}

func resourceArchitectUserPrompt() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud User Audio Prompt",

		CreateContext: createWithPooledClient(createUserPrompt),
		ReadContext:   readWithPooledClient(readUserPrompt),
		UpdateContext: updateWithPooledClient(updateUserPrompt),
		DeleteContext: deleteWithPooledClient(deleteUserPrompt),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the user audio prompt. Note: If the name of the user prompt is changed, this will cause the Prompt to be dropped and recreated with a new ID. This will generate a new ID for the prompt and will invalidate any Architect flows referencing it. ",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "Description of the user audio prompt.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"resources": {
				Description: "Audio of TTS resources for the audio prompt.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				ConfigMode:  schema.SchemaConfigModeAttr,
				Elem:        userPromptResource,
			},
		},
	}
}

func createUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	prompt := platformclientv2.Prompt{
		Name:        &name,
		Description: &description,
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
			tag["filename"] = []string{resourceMap["filename"].(string)}

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
			userPromptResource, _, err := architectApi.PostArchitectPromptResources(*userPrompt.Id, promptResource)
			if err != nil {
				return diag.Errorf("Failed to create user prompt resource %s: %s", name, err)
			}
			uploadUri := userPromptResource.UploadUri

			resourceFilename := resourceMap["filename"]
			if resourceFilename.(string) == "" {
				continue
			}
			resourceFilenameStr := resourceFilename.(string)

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
	sdkConfig := meta.(*providerMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading User Prompt %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		userPrompt, resp, getErr := architectAPI.GetArchitectPrompt(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read User Prompt %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read User Prompt %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceArchitectUserPrompt())
		if userPrompt.Name != nil {
			d.Set("name", *userPrompt.Name)
		} else {
			d.Set("name", nil)
		}

		if userPrompt.Description != nil {
			d.Set("description", *userPrompt.Description)
		} else {
			d.Set("description", nil)
		}

		if resources, ok := d.GetOk("resources"); ok && resources != nil {
			promptResources := resources.(*schema.Set).List()
			for _, promptResource := range promptResources {
				resourceMap := promptResource.(map[string]interface{})
				resourceFilename := resourceMap["filename"]
				if resourceFilename.(string) == "" {
					continue
				}
				APIResources := *userPrompt.Resources
				isTranscoded := false
				for _, APIResource := range APIResources {
					if APIResource.Tags != nil {
						tags := *APIResource.Tags
						if len(tags["filename"]) > 0 {
							if tags["filename"][0] == resourceFilename {
								if *APIResource.UploadStatus == "transcoded" {
									isTranscoded = true
								}
							}
						}
					}
				}
				if !isTranscoded {
					return resource.RetryableError(fmt.Errorf("prompt file not transcoded"))
				}
			}
		}

		d.Set("resources", flattenPromptResources(userPrompt.Resources))

		log.Printf("Read Audio Prompt %s %s", d.Id(), *userPrompt.Id)
		return cc.CheckState()
	})
}

func updateUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	prompt := platformclientv2.Prompt{
		Name:        &name,
		Description: &description,
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Deleting user prompt %s", name)
	if _, err := architectApi.DeleteArchitectPrompt(d.Id(), true); err != nil {
		return diag.Errorf("Failed to delete user prompt %s: %s", name, err)
	}
	log.Printf("Deleted user prompt %s", name)

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := architectApi.GetArchitectPrompt(d.Id())
		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				// User prompt deleted
				log.Printf("Deleted user prompt %s", name)
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting user prompt %s: %s", name, err))
		}
		return resource.RetryableError(fmt.Errorf("User prompt %s still exists", name))
	})
}

func uploadPrompt(uploadUri *string, filename *string, sdkConfig *platformclientv2.Configuration) error {
	reader, file, err := downloadOrOpenFile(*filename)
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

func flattenPromptResources(promptResources *[]platformclientv2.Promptasset) *schema.Set {
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

		resourceSet.Add(promptResource)
	}
	return resourceSet
}

func updatePromptResource(d *schema.ResourceData, architectApi *platformclientv2.ArchitectApi, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	name := d.Get("name").(string)

	// Get the prompt so we can get existing prompt resources
	userPrompt, _, err := architectApi.GetArchitectPrompt(d.Id())
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	apiInstance := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	data, _, err := apiInstance.GetArchitectPrompt(promptId)
	if err != nil {
		return nil, err
	}

	promptResourceData := []PromptAudioData{}
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

// Download audio file from mediaUri to directory/fileName
func downloadAudioFile(directory string, fileName string, mediaUri string) error {
	resp, err := http.Get(mediaUri)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	out, err := os.Create(fmt.Sprintf("%s/%s", directory, fileName))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
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
				}
			}
		}
	}
}
