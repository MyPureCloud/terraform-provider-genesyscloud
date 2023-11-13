package integration_credential

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

/*
The resource_genesyscloud_integration_credential.go contains all of the methods that perform the core logic for a resource.
In general a resource should have a approximately 5 methods in it:

1.  A getAll.... function that the CX as Code exporter will use during the process of exporting Genesys Cloud.
2.  A create.... function that the resource will use to create a Genesys Cloud object (e.g. genesycloud_integration_credential)
3.  A read.... function that looks up a single resource.
4.  An update... function that updates a single resource.
5.  A delete.... function that deletes a single resource.

Two things to note:

 1. All code in these methods should be focused on getting data in and out of Terraform.  All code that is used for interacting
    with a Genesys API should be encapsulated into a proxy class contained within the package.

 2. In general, to keep this file somewhat manageable, if you find yourself with a number of helper functions move them to a

utils function in the package.  This will keep the code manageable and easy to work through.
*/

// getAllCredentials retrieves all of the integration credentials via Terraform in the Genesys Cloud and is used for the exporter
func getAllCredentials(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	ip := getIntegrationCredsProxy(clientConfig)

	credentials, err := ip.getAllIntegrationCreds(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get all credentials: %v", err)
	}

	for _, cred := range *credentials {
		log.Printf("Dealing with credential id : %s", *cred.Id)
		if cred.Name != nil { // Credential is possible to have no name
			resources[*cred.Id] = &resourceExporter.ResourceMeta{Name: *cred.Name}
		}
	}

	return resources, nil
}

// createCredential is used by the integration credential resource to create Genesyscloud integration credential
func createCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	cred_type := d.Get("credential_type_name").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ip := getIntegrationCredsProxy(sdkConfig)

	createCredential := platformclientv2.Credential{
		Name: &name,
		VarType: &platformclientv2.Credentialtype{
			Name: &cred_type,
		},
		CredentialFields: buildCredentialFields(d),
	}

	credential, err := ip.createIntegrationCred(ctx, &createCredential)
	if err != nil {
		return diag.Errorf("Failed to create credential %s : %s", name, err)
	}

	d.SetId(*credential.Id)

	log.Printf("Created credential %s, %s", name, *credential.Id)
	return readCredential(ctx, d, meta)
}

// readCredential is used by the integration credential resource to read a  credential from genesys cloud.
func readCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ip := getIntegrationCredsProxy(sdkConfig)

	log.Printf("Reading credential %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentCredential, resp, err := ip.getIntegrationCredById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read credential %s: %s", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read credential %s: %s", d.Id(), err))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationCredential())
		d.Set("name", *currentCredential.Name)
		d.Set("credential_type_name", *currentCredential.VarType.Name)

		log.Printf("Read credential %s %s", d.Id(), *currentCredential.Name)

		return cc.CheckState()
	})
}

// updateCredential is used by the integration credential resource to update a credential in Genesys Cloud
func updateCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	cred_type := d.Get("credential_type_name").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ip := getIntegrationCredsProxy(sdkConfig)

	if d.HasChanges("name", "credential_type_name", "fields") {
		log.Printf("Updating credential %s", name)

		_, err := ip.updateIntegrationCred(ctx, d.Id(), &platformclientv2.Credential{
			Name: &name,
			VarType: &platformclientv2.Credentialtype{
				Name: &cred_type,
			},
			CredentialFields: buildCredentialFields(d),
		})
		if err != nil {
			return diag.Errorf("Failed to update credential %s: %s", name, err)
		}
	}

	log.Printf("Updated credential %s %s", name, d.Id())
	return readCredential(ctx, d, meta)
}

// deleteCredential is used by the integration credential resource to delete a credential from Genesys cloud.
func deleteCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ip := getIntegrationCredsProxy(sdkConfig)

	_, err := ip.deleteIntegrationCred(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete the credential %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := ip.getIntegrationCredById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Integration credential deleted
				log.Printf("Deleted Integration credential %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting credential action %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("integration credential %s still exists", d.Id()))
	})
}
