package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	responsemanagementresponseresponsetextResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`content`: {
				Description: `Response text content.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`content_type`: {
				Description:  `Response text content type.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`text/plain`, `text/html`}, false),
			},
		},
	}
	responsemanagementresponseresponsesubstitutionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`description`: {
				Description: `Response substitution description.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`default_value`: {
				Description: `Response substitution default value.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}
	responsemanagementresponsemessagingtemplateResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`whats_app`: {
				Description: `Defines a messaging template for a WhatsApp messaging channel`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeSet,
				Elem:        responsemanagementresponsewhatsappdefinitionResource,
				Set: func(_ interface{}) int {
					return 0
				},
			},
		},
	}
	responsemanagementresponsewhatsappdefinitionResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The messaging template name.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`namespace`: {
				Description: `The messaging template namespace.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`language`: {
				Description: `The messaging template language configured for this template. This is a WhatsApp specific value. For example, 'en_US'`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
)

func ResourceResponsemanagementResponse() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud responsemanagement response`,

		CreateContext: CreateWithPooledClient(createResponsemanagementResponse),
		ReadContext:   ReadWithPooledClient(readResponsemanagementResponse),
		UpdateContext: UpdateWithPooledClient(updateResponsemanagementResponse),
		DeleteContext: DeleteWithPooledClient(deleteResponsemanagementResponse),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `Name of the responsemanagement response`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`library_ids`: {
				Description: `One or more libraries response is associated with. Changing the library IDs will result in the resource being recreated`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			`texts`: {
				Description: `One or more texts associated with the response.`,
				Required:    true,
				Type:        schema.TypeSet,
				Elem:        responsemanagementresponseresponsetextResource,
			},
			`interaction_type`: {
				Description:  `The interaction type for this response.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`chat`, `email`, `twitter`}, false),
			},
			`substitutions`: {
				Description: `Details about any text substitutions used in the texts for this response.`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        responsemanagementresponseresponsesubstitutionResource,
			},
			`substitutions_schema_id`: {
				Description: `Metadata about the text substitutions in json schema format.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`response_type`: {
				Description:  `The response type represented by the response.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`MessagingTemplate`, `CampaignSmsTemplate`, `CampaignEmailTemplate`}, false),
			},
			`messaging_template`: {
				Description: `An optional messaging template definition for responseType.MessagingTemplate.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeSet,
				Elem:        responsemanagementresponsemessagingtemplateResource,
				Set: func(_ interface{}) int {
					return 0
				},
			},
			`asset_ids`: {
				Description: `Assets used in the response`,
				Optional:    true,
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func getAllResponsemanagementResponse(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		libraries, _, getErr := responseManagementApi.GetResponsemanagementLibraries(pageNum, pageSize, "", "")

		if getErr != nil {
			return nil, diag.Errorf("Error requesting page of Responsemanagement library: %s", getErr)
		}
		if libraries.Entities == nil || len(*libraries.Entities) == 0 {
			break
		}

		for _, library := range *libraries.Entities {
			for pageNum := 1; ; pageNum++ {
				const pageSize = 100
				sdkresponseentitylisting, _, getErr := responseManagementApi.GetResponsemanagementResponses(*library.Id, pageNum, pageSize, "")
				if getErr != nil {
					return nil, diag.Errorf("Error requesting page of Responsemanagement Response: %s", getErr)
				}

				if sdkresponseentitylisting.Entities == nil || len(*sdkresponseentitylisting.Entities) == 0 {
					break
				}

				for _, entity := range *sdkresponseentitylisting.Entities {
					resources[*entity.Id] = &resourceExporter.ResourceMeta{Name: *entity.Name}
				}
			}
		}
	}

	return resources, nil
}

func ResponsemanagementResponseExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllResponsemanagementResponse),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			`library_ids`: {
				RefType: "genesyscloud_responsemanagement_library",
			},
			`asset_ids`: {
				RefType: "genesyscloud_responsemanagement_responseasset",
			},
		},
		JsonEncodeAttributes: []string{"substitutions_schema_id"},
	}
}

func createResponsemanagementResponse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	interactionType := d.Get("interaction_type").(string)
	substitutionsSchema := d.Get("substitutions_schema_id").(string)
	responseType := d.Get("response_type").(string)
	messagingTemplate := d.Get("messaging_template").(*schema.Set)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	sdkresponse := platformclientv2.Response{
		Libraries:     BuildSdkDomainEntityRefArr(d, "library_ids"),
		Texts:         buildSdkresponsemanagementresponseResponsetextSlice(d.Get("texts").(*schema.Set)),
		Substitutions: buildSdkresponsemanagementresponseResponsesubstitutionSlice(d.Get("substitutions").(*schema.Set)),
		Assets:        buildSdkresponsemanagementresponseAddressableentityrefSlice(d.Get("asset_ids").(*schema.Set)),
	}

	if name != "" {
		sdkresponse.Name = &name
	}
	if interactionType != "" {
		sdkresponse.InteractionType = &interactionType
	}
	if substitutionsSchema != "" {
		sdkresponse.SubstitutionsSchema = &platformclientv2.Jsonschemadocument{Id: &substitutionsSchema}
	}
	if responseType != "" {
		sdkresponse.ResponseType = &responseType
	}
	// Need to check messaging template like this to avoid the responseType being giving a default value
	if messagingTemplate.Len() > 0 {
		sdkresponse.MessagingTemplate = buildSdkresponsemanagementresponseMessagingtemplate(messagingTemplate)
	}

	log.Printf("Creating Responsemanagement Response %s", name)
	responsemanagementResponse, _, err := responseManagementApi.PostResponsemanagementResponses(sdkresponse, "")
	if err != nil {
		return diag.Errorf("Failed to create Responsemanagement Response %s: %s", name, err)
	}
	d.SetId(*responsemanagementResponse.Id)

	log.Printf("Created Responsemanagement Response %s %s", name, *responsemanagementResponse.Id)
	return readResponsemanagementResponse(ctx, d, meta)
}

func updateResponsemanagementResponse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	interactionType := d.Get("interaction_type").(string)
	substitutionsSchema := d.Get("substitutions_schema_id").(string)
	responseType := d.Get("response_type").(string)
	messagingTemplate := d.Get("messaging_template").(*schema.Set)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	sdkresponse := platformclientv2.Response{
		Libraries:     BuildSdkDomainEntityRefArr(d, "library_ids"),
		Texts:         buildSdkresponsemanagementresponseResponsetextSlice(d.Get("texts").(*schema.Set)),
		Substitutions: buildSdkresponsemanagementresponseResponsesubstitutionSlice(d.Get("substitutions").(*schema.Set)),
		Assets:        buildSdkresponsemanagementresponseAddressableentityrefSlice(d.Get("asset_ids").(*schema.Set)),
	}

	if name != "" {
		sdkresponse.Name = &name
	}
	if interactionType != "" {
		sdkresponse.InteractionType = &interactionType
	}
	if substitutionsSchema != "" {
		sdkresponse.SubstitutionsSchema = &platformclientv2.Jsonschemadocument{Id: &substitutionsSchema}
	}
	if responseType != "" {
		sdkresponse.ResponseType = &responseType
	}
	// Need to check messaging template like this to avoid the responseType being giving a default value
	if messagingTemplate.Len() > 0 {
		sdkresponse.MessagingTemplate = buildSdkresponsemanagementresponseMessagingtemplate(messagingTemplate)
	}

	log.Printf("Updating Responsemanagement Response %s", name)
	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Responsemanagement Response version
		responsemanagementResponse, resp, getErr := responseManagementApi.GetResponsemanagementResponse(d.Id(), "")
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Responsemanagement Response %s: %s", d.Id(), getErr)
		}
		sdkresponse.Version = responsemanagementResponse.Version
		responsemanagementResponse, _, updateErr := responseManagementApi.PutResponsemanagementResponse(d.Id(), sdkresponse, "")
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Responsemanagement Response %s: %s", name, updateErr)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Responsemanagement Response %s", name)
	return readResponsemanagementResponse(ctx, d, meta)
}

func readResponsemanagementResponse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	log.Printf("Reading Responsemanagement Response %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkresponse, resp, getErr := responseManagementApi.GetResponsemanagementResponse(d.Id(), "")
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Responsemanagement Response %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Responsemanagement Response %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceResponsemanagementResponse())

		if sdkresponse.Name != nil {
			d.Set("name", *sdkresponse.Name)
		}
		if sdkresponse.Libraries != nil {
			d.Set("library_ids", SdkDomainEntityRefArrToList(*sdkresponse.Libraries))
		}
		if sdkresponse.Texts != nil {
			d.Set("texts", flattenSdkresponsemanagementresponseResponsetextSlice(*sdkresponse.Texts))
		}
		if sdkresponse.InteractionType != nil {
			d.Set("interaction_type", *sdkresponse.InteractionType)
		}
		if sdkresponse.Substitutions != nil {
			d.Set("substitutions", flattenSdkresponsemanagementresponseResponsesubstitutionSlice(*sdkresponse.Substitutions))
		}
		if sdkresponse.SubstitutionsSchema != nil && sdkresponse.SubstitutionsSchema.Id != nil {
			d.Set("substitutions_schema_id", *sdkresponse.SubstitutionsSchema.Id)
		}
		if sdkresponse.ResponseType != nil {
			d.Set("response_type", *sdkresponse.ResponseType)
		}
		if sdkresponse.MessagingTemplate != nil {
			d.Set("messaging_template", flattenSdkresponsemanagementresponseMessagingtemplate(sdkresponse.MessagingTemplate))
		}
		if sdkresponse.Assets != nil {
			d.Set("asset_ids", flattenSdkresponsemanagementresponseAddressableentityrefSlice(*sdkresponse.Assets))
		}

		log.Printf("Read Responsemanagement Response %s %s", d.Id(), *sdkresponse.Name)
		return cc.CheckState()
	})
}

func deleteResponsemanagementResponse(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	responseManagementApi := platformclientv2.NewResponseManagementApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Responsemanagement Response")
		resp, err := responseManagementApi.DeleteResponsemanagementResponse(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Responsemanagement Response: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	time.Sleep(30 * time.Second)
	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := responseManagementApi.GetResponsemanagementResponse(d.Id(), "")
		if err != nil {
			if IsStatus404(resp) {
				// Responsemanagement Response deleted
				log.Printf("Deleted Responsemanagement Response %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Responsemanagement Response %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Responsemanagement Response %s still exists", d.Id()))
	})
}

func buildSdkresponsemanagementresponseResponsetextSlice(responsetext *schema.Set) *[]platformclientv2.Responsetext {
	if responsetext == nil {
		return nil
	}
	sdkResponsetextSlice := make([]platformclientv2.Responsetext, 0)
	responsetextList := responsetext.List()
	for _, configresponsetext := range responsetextList {
		var sdkResponsetext platformclientv2.Responsetext
		responsetextMap := configresponsetext.(map[string]interface{})
		if content := responsetextMap["content"].(string); content != "" {
			sdkResponsetext.Content = &content
		}
		if contentType := responsetextMap["content_type"].(string); contentType != "" {
			sdkResponsetext.ContentType = &contentType
		}

		sdkResponsetextSlice = append(sdkResponsetextSlice, sdkResponsetext)
	}
	return &sdkResponsetextSlice
}

func buildSdkresponsemanagementresponseResponsesubstitutionSlice(responsesubstitution *schema.Set) *[]platformclientv2.Responsesubstitution {
	if responsesubstitution == nil {
		return nil
	}
	sdkResponsesubstitutionSlice := make([]platformclientv2.Responsesubstitution, 0)
	responsesubstitutionList := responsesubstitution.List()
	for _, configresponsesubstitution := range responsesubstitutionList {
		var sdkResponsesubstitution platformclientv2.Responsesubstitution
		responsesubstitutionMap := configresponsesubstitution.(map[string]interface{})
		if description := responsesubstitutionMap["description"].(string); description != "" {
			sdkResponsesubstitution.Description = &description
		}
		if defaultValue := responsesubstitutionMap["default_value"].(string); defaultValue != "" {
			sdkResponsesubstitution.DefaultValue = &defaultValue
		}

		sdkResponsesubstitutionSlice = append(sdkResponsesubstitutionSlice, sdkResponsesubstitution)
	}
	return &sdkResponsesubstitutionSlice
}

func buildSdkresponsemanagementresponseWhatsappdefinition(whatsappdefinition *schema.Set) *platformclientv2.Whatsappdefinition {
	if whatsappdefinition == nil {
		return nil
	}
	var sdkWhatsappdefinition platformclientv2.Whatsappdefinition
	whatsappdefinitionList := whatsappdefinition.List()
	if len(whatsappdefinitionList) > 0 {
		whatsappdefinitionMap := whatsappdefinitionList[0].(map[string]interface{})
		if name := whatsappdefinitionMap["name"].(string); name != "" {
			sdkWhatsappdefinition.Name = &name
		}
		if namespace := whatsappdefinitionMap["namespace"].(string); namespace != "" {
			sdkWhatsappdefinition.Namespace = &namespace
		}
		if language := whatsappdefinitionMap["language"].(string); language != "" {
			sdkWhatsappdefinition.Language = &language
		}
	}

	return &sdkWhatsappdefinition
}

func buildSdkresponsemanagementresponseMessagingtemplate(messagingtemplate *schema.Set) *platformclientv2.Messagingtemplate {
	if messagingtemplate == nil {
		return nil
	}
	var sdkMessagingtemplate platformclientv2.Messagingtemplate
	messagingtemplateList := messagingtemplate.List()
	if len(messagingtemplateList) > 0 {
		messagingtemplateMap := messagingtemplateList[0].(map[string]interface{})
		if whatsApp := messagingtemplateMap["whats_app"]; whatsApp != nil {
			sdkMessagingtemplate.WhatsApp = buildSdkresponsemanagementresponseWhatsappdefinition(whatsApp.(*schema.Set))
		}
	}

	return &sdkMessagingtemplate
}

func buildSdkresponsemanagementresponseAddressableentityrefSlice(addressableentityref *schema.Set) *[]platformclientv2.Addressableentityref {
	if addressableentityref == nil {
		return nil
	}
	strList := lists.SetToStringList(addressableentityref)
	if strList == nil {
		return nil
	}
	addressableentityrefRefs := make([]platformclientv2.Addressableentityref, len(*strList))
	for i, id := range *strList {
		tempId := id
		addressableentityrefRefs[i] = platformclientv2.Addressableentityref{Id: &tempId}
	}
	return &addressableentityrefRefs
}

func flattenSdkresponsemanagementresponseResponsetextSlice(responsetexts []platformclientv2.Responsetext) *schema.Set {
	if len(responsetexts) == 0 {
		return nil
	}

	responsetextSet := schema.NewSet(schema.HashResource(responsemanagementresponseresponsetextResource), []interface{}{})
	for _, responsetext := range responsetexts {
		responsetextMap := make(map[string]interface{})

		if responsetext.Content != nil {
			responsetextMap["content"] = *responsetext.Content
		}
		if responsetext.ContentType != nil {
			responsetextMap["content_type"] = *responsetext.ContentType
		}

		responsetextSet.Add(responsetextMap)
	}

	return responsetextSet
}

func flattenSdkresponsemanagementresponseResponsesubstitutionSlice(responsesubstitutions []platformclientv2.Responsesubstitution) *schema.Set {
	if len(responsesubstitutions) == 0 {
		return nil
	}

	responsesubstitutionSet := schema.NewSet(schema.HashResource(responsemanagementresponseresponsesubstitutionResource), []interface{}{})
	for _, responsesubstitution := range responsesubstitutions {
		responsesubstitutionMap := make(map[string]interface{})

		if responsesubstitution.Description != nil {
			responsesubstitutionMap["description"] = *responsesubstitution.Description
		}
		if responsesubstitution.DefaultValue != nil {
			responsesubstitutionMap["default_value"] = *responsesubstitution.DefaultValue
		}

		responsesubstitutionSet.Add(responsesubstitutionMap)
	}

	return responsesubstitutionSet
}

func flattenSdkresponsemanagementresponseWhatsappdefinition(whatsappdefinition *platformclientv2.Whatsappdefinition) *schema.Set {
	if whatsappdefinition == nil {
		return nil
	}

	whatsappdefinitionSet := schema.NewSet(schema.HashResource(responsemanagementresponsewhatsappdefinitionResource), []interface{}{})
	whatsappdefinitionMap := make(map[string]interface{})

	if whatsappdefinition.Name != nil {
		whatsappdefinitionMap["name"] = *whatsappdefinition.Name
	}
	if whatsappdefinition.Namespace != nil {
		whatsappdefinitionMap["namespace"] = *whatsappdefinition.Namespace
	}
	if whatsappdefinition.Language != nil {
		whatsappdefinitionMap["language"] = *whatsappdefinition.Language
	}

	whatsappdefinitionSet.Add(whatsappdefinitionMap)

	return whatsappdefinitionSet
}

func flattenSdkresponsemanagementresponseMessagingtemplate(messagingtemplate *platformclientv2.Messagingtemplate) *schema.Set {
	if messagingtemplate == nil {
		return nil
	}

	messagingtemplateSet := schema.NewSet(schema.HashResource(responsemanagementresponsemessagingtemplateResource), []interface{}{})
	messagingtemplateMap := make(map[string]interface{})

	if messagingtemplate.WhatsApp != nil {
		messagingtemplateMap["whats_app"] = flattenSdkresponsemanagementresponseWhatsappdefinition(messagingtemplate.WhatsApp)
	}

	messagingtemplateSet.Add(messagingtemplateMap)

	return messagingtemplateSet
}

func flattenSdkresponsemanagementresponseAddressableentityrefSlice(addressableentityrefs []platformclientv2.Addressableentityref) *schema.Set {
	addressableentityrefList := make([]interface{}, len(addressableentityrefs))
	for i, v := range addressableentityrefs {
		addressableentityrefList[i] = *v.Id
	}
	return schema.NewSet(schema.HashString, addressableentityrefList)
}
