package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func getAllCredentials(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		credentials, _, err := integrationAPI.GetIntegrationsCredentials(pageNum, pageSize)
		if err != nil {
			return nil, diag.Errorf("Failed to get page of credentials: %v", err)
		}

		if credentials.Entities == nil || len(*credentials.Entities) == 0 {
			break
		}

		for _, cred := range *credentials.Entities {
			if cred.Name != nil { // Credential is possible to have no name
				resources[*cred.Id] = &resourceExporter.ResourceMeta{Name: *cred.Name}
			}
		}
	}

	return resources, nil
}

func CredentialExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllCredentials),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No Reference
		UnResolvableAttributes: map[string]*schema.Schema{
			"fields": ResourceCredential().Schema["fields"],
		},
	}
}

func ResourceCredential() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Credential",

		CreateContext: CreateWithPooledClient(createCredential),
		ReadContext:   ReadWithPooledClient(readCredential),
		UpdateContext: UpdateWithPooledClient(updateCredential),
		DeleteContext: DeleteWithPooledClient(deleteCredential),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Credential name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"credential_type_name": {
				Description: "Credential type name. Use [GET /api/v2/integrations/credentials/types](https://developer.genesys.cloud/api/rest/v2/integrations/#get-api-v2-integrations-credentials-types) to see the list of available integration credential types. ",
				Type:        schema.TypeString,
				Required:    true,
			},
			"fields": {
				Description: "Credential fields. Different credential types require different fields. Missing any correct required fields will result API request failure. Use [GET /api/v2/integrations/credentials/types](https://developer.genesys.cloud/api/rest/v2/integrations/#get-api-v2-integrations-credentials-types) to check out the specific credential type schema to find out what fields are required. ",
				Type:        schema.TypeMap,
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func createCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	name := d.Get("name").(string)
	cred_type := d.Get("credential_type_name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	createCredential := platformclientv2.Credential{
		Name: &name,
		VarType: &platformclientv2.Credentialtype{
			Name: &cred_type,
		},
		CredentialFields: buildCredentialFields(d),
	}

	credential, _, err := integrationAPI.PostIntegrationsCredentials(createCredential)

	if err != nil {
		return diag.Errorf("Failed to create credential %s : %s", name, err)
	}

	d.SetId(*credential.Id)

	log.Printf("Created credential %s, %s", name, *credential.Id)
	return readCredential(ctx, d, meta)
}

func readCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	log.Printf("Reading credential %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentCredential, resp, getErr := integrationAPI.GetIntegrationsCredential(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read credential %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read credential %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceCredential())
		d.Set("name", *currentCredential.Name)
		d.Set("credential_type_name", *currentCredential.VarType.Name)

		log.Printf("Read credential %s %s", d.Id(), *currentCredential.Name)

		return cc.CheckState()
	})
}

func updateCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	cred_type := d.Get("credential_type_name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	if d.HasChanges("name", "credential_type_name", "fields") {

		log.Printf("Updating credential %s", name)

		_, _, putErr := integrationAPI.PutIntegrationsCredential(d.Id(), platformclientv2.Credential{
			Name: &name,
			VarType: &platformclientv2.Credentialtype{
				Name: &cred_type,
			},
			CredentialFields: buildCredentialFields(d),
		})
		if putErr != nil {
			return diag.Errorf("Failed to update credential %s: %s", name, putErr)
		}
	}

	log.Printf("Updated credential %s %s", name, d.Id())
	return readCredential(ctx, d, meta)
}

func deleteCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	_, err := integrationAPI.DeleteIntegrationsCredential(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete the credential %s: %s", d.Id(), err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := integrationAPI.GetIntegrationsCredential(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Integration credential deleted
				log.Printf("Deleted Integration credential %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting credential action %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Integration credential %s still exists", d.Id()))
	})
}

func buildCredentialFields(d *schema.ResourceData) *map[string]string {
	results := make(map[string]string)
	if fields, ok := d.GetOk("fields"); ok {
		fieldMap := fields.(map[string]interface{})
		for k, v := range fieldMap {
			results[k] = v.(string)
		}
		return &results
	}
	return &results
}
