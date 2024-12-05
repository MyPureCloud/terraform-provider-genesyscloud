package integration_credential

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	oauth "terraform-provider-genesyscloud/genesyscloud/oauth_client"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
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

	credentials, resp, err := ip.getAllIntegrationCreds(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all credentials error: %s", err), resp)
	}

	for _, cred := range *credentials {
		log.Printf("Dealing with credential id : %s", *cred.Id)
		if cred.Name != nil { // Credential is possible to have no name

			// Export integration credential only if it matches the expected format: DEVTOOLING-310
			regexPattern := regexp.MustCompile("Integration-.+")
			if !regexPattern.MatchString(*cred.Name) {
				log.Printf("integration credential name [%s] does not match the expected format [%s], not exporting integration credential id %s", *cred.Name, regexPattern.String(), *cred.Id)
				continue
			}
			// Verify that the integration entity itself exist before exporting the integration credentials associated to it: DEVTOOLING-282
			integrationId := strings.Split(*cred.Name, "Integration-")[1]
			_, resp, err := ip.getIntegrationById(ctx, integrationId)
			if err != nil {
				if util.IsStatus404(resp) {
					log.Printf("Integration id %s no longer exist, we are therefore not exporting the associated integration credential id %s", integrationId, *cred.Id)
					continue
				} else {
					log.Printf("Integration id %s exists but we got an unexpected error retrieving it: %v", integrationId, err)
				}
			}
			resources[*cred.Id] = &resourceExporter.ResourceMeta{BlockLabel: *cred.Name}
		}
	}
	return resources, nil
}

// createCredential is used by the integration credential resource to create Genesyscloud integration credential
func createCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	cred_type := d.Get("credential_type_name").(string)
	fields := buildCredentialFields(d)
	_, secretFieldPresent := fields["clientSecret"]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ip := getIntegrationCredsProxy(sdkConfig)

	//If if is a Genesys Cloud OAuth Client and the user has not provided a secret field we should look for the
	//item in the cache DEVTOOLING-448
	if cred_type == "pureCloudOAuthClient" && !secretFieldPresent {
		retrieveCachedOauthClientSecret(sdkConfig, fields)
	}

	createCredential := platformclientv2.Credential{
		Name: &name,
		VarType: &platformclientv2.Credentialtype{
			Name: &cred_type,
		},
		CredentialFields: &fields,
	}

	credential, resp, err := ip.createIntegrationCred(ctx, &createCredential)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create credential %s error: %s", name, err), resp)
	}

	d.SetId(*credential.Id)
	log.Printf("Created credential %s, %s", name, *credential.Id)
	return readCredential(ctx, d, meta)
}

func retrieveCachedOauthClientSecret(sdkConfig *platformclientv2.Configuration, fields map[string]string) {
	op := oauth.GetOAuthClientProxy(sdkConfig)
	if clientId, ok := fields["clientId"]; ok {
		oAuthClient := op.GetCachedOAuthClient(clientId)
		fields["clientSecret"] = *oAuthClient.Secret
		log.Printf("Successfully matched with OAuth Client Credential id %s", clientId)
	}
}

// readCredential is used by the integration credential resource to read a  credential from genesys cloud.
func readCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ip := getIntegrationCredsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationCredential(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading credential %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentCredential, resp, err := ip.getIntegrationCredById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read credential %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read credential %s | error: %s", d.Id(), err), resp))
		}

		_ = d.Set("name", *currentCredential.Name)
		_ = d.Set("credential_type_name", *currentCredential.VarType.Name)

		log.Printf("Read credential %s %s", d.Id(), *currentCredential.Name)

		return cc.CheckState(d)
	})
}

// updateCredential is used by the integration credential resource to update a credential in Genesys Cloud
func updateCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	cred_type := d.Get("credential_type_name").(string)
	fields := buildCredentialFields(d)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ip := getIntegrationCredsProxy(sdkConfig)

	if d.HasChanges("name", "credential_type_name", "fields") {
		log.Printf("Updating credential %s", name)

		_, resp, err := ip.updateIntegrationCred(ctx, d.Id(), &platformclientv2.Credential{
			Name: &name,
			VarType: &platformclientv2.Credentialtype{
				Name: &cred_type,
			},
			CredentialFields: &fields,
		})
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update credential %s error: %s", name, err), resp)
		}
	}
	log.Printf("Updated credential %s %s", name, d.Id())
	return readCredential(ctx, d, meta)
}

// deleteCredential is used by the integration credential resource to delete a credential from Genesys cloud.
func deleteCredential(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ip := getIntegrationCredsProxy(sdkConfig)

	resp, err := ip.deleteIntegrationCred(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete credential %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := ip.getIntegrationCredById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Integration credential deleted
				log.Printf("Deleted Integration credential %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting credential action %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("integration credential %s still exists", d.Id()), resp))
	})
}
