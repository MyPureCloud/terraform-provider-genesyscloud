package tfexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/platform"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
This files contains all of the code used to create an export's Terraform state file.  The TFStateFileWriter struct encapsulates all the logic to write a Terraform state file.
The other functions in this file deal with how to generate the TFVars we create during the export.
*/
type TFStateFileWriter struct {
	ctx              context.Context
	resources        []resourceExporter.ResourceInfo
	d                *schema.ResourceData
	providerRegistry string
}

func NewTFStateWriter(ctx context.Context, resources []resourceExporter.ResourceInfo, d *schema.ResourceData, providerRegistry string) (*TFStateFileWriter, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}
	if d == nil {
		return nil, fmt.Errorf("schema.ResourceData cannot be nil")
	}
	if len(resources) == 0 {
		return nil, fmt.Errorf("no resources found to export")
	}
	tfWriter := &TFStateFileWriter{
		ctx:              ctx,
		resources:        resources,
		d:                d,
		providerRegistry: providerRegistry,
	}

	return tfWriter, nil
}

func (t *TFStateFileWriter) writeTfState() diag.Diagnostics {

	platformInstance := platform.GetPlatform()
	platformErr := platformInstance.Validate()
	if platformErr != nil {
		log.Printf("Failed to validate platform: %v. Will use default values to run final state commands.", platformErr)
	}

	stateFilePath, diagErr := getFilePath(t.d, defaultTfStateFile)
	if diagErr != nil {
		return diagErr
	}

	tfstate := terraform.NewState()
	if tfstate == nil {
		return diag.Errorf("failed to create new terraform state")
	}

	// Ensure the root module and resources map are initialized
	rootModule := tfstate.RootModule()
	if rootModule == nil {
		return diag.Errorf("failed to get root module")
	}
	if rootModule.Resources == nil {
		rootModule.Resources = make(map[string]*terraform.ResourceState)
	}

	for _, resource := range t.resources {
		resourceKey := ""
		if resource.BlockType != "" {
			resourceKey = resource.BlockType + "."
		}
		resourceKey += resource.Type + "." + resource.BlockLabel
		if resourceKey == ".." || resourceKey == "." { // This would catch the worst case of all empty strings
			return diag.Errorf("invalid resource key generated for resource: %+v", resource)
		}

		resourceState := &terraform.ResourceState{
			Type:     resource.Type,
			Primary:  resource.State,
			Provider: "provider.genesyscloud",
		}
		rootModule.Resources[resourceKey] = resourceState
	}

	data, err := json.MarshalIndent(tfstate, "", "  ")
	if err != nil {
		return diag.Errorf("Failed to encode state as JSON: %v", err)
	}

	log.Printf("Writing export state file to %s", stateFilePath)
	if diagErr := files.WriteToFile(data, stateFilePath); diagErr != nil {
		return diagErr
	}

	// This outputs terraform state v3, and there is currently no public lib to generate v4 which is required for terraform 0.13+.
	// However, the state can be upgraded automatically by calling the terraform CLI. If this fails, just print a warning indicating
	// that the state likely needs to be upgraded manually.
	cliErrorPostscript := fmt.Sprintf(`The generated tfstate file will need to be upgraded manually by running the following in the state file's directory:
	'%s state replace-provider %s/-/genesyscloud %s/mypurecloud/genesyscloud'`, platformInstance.Binary(), platformInstance.GetProviderRegistry(), t.providerRegistry)

	if platformInstance.IsDevelopmentPlatform() {
		cliErrorPostscript = `The current process is running via a debug server (debug binary detected), and so it is unable to run the proper command to replace the state. Please run this command outside of a debug session.` + cliErrorPostscript
		log.Print(cliErrorPostscript)
		return nil
	}

	replaceProviderOutput, err := platformInstance.ExecuteCommand(t.ctx, []string{
		"state",
		"replace-provider",
		"-auto-approve",
		"-state=" + stateFilePath,
		// This is the provider determined by the platform (terraform vs tofu)
		fmt.Sprintf("%s/-/genesyscloud", platformInstance.GetProviderRegistry()),
		// This is the platform that accounts for custom builds
		fmt.Sprintf("%s/mypurecloud/genesyscloud", t.providerRegistry),
	}...)
	if err != nil {
		cliErrorPostscript = fmt.Sprintf(`Failed to run the terraform CLI to upgrade the generated state file:
			Error: %v

			%s`, err, cliErrorPostscript)
		log.Print(cliErrorPostscript)
		// Don't fail everything even if this errors.

		return nil
	}

	log.Print(replaceProviderOutput.Stdout)
	return nil
}

func generateTfVarsContent(vars map[string]interface{}) string {
	tfVarsContent := ""
	for k, v := range vars {
		vStr := v
		if v == nil {
			vStr = "null"
		} else if s, ok := v.(string); ok {
			vStr = fmt.Sprintf(`"%s"`, s)
		} else if m, ok := v.(map[string]interface{}); ok {
			vStr = fmt.Sprintf(`{
	%s
}`, strings.Replace(generateTfVarsContent(m), "\n", "\n\t", -1))
		}
		newLine := ""
		if tfVarsContent != "" {
			newLine = "\n"
		}
		tfVarsContent = fmt.Sprintf("%v%s%s = %v", tfVarsContent, newLine, k, vStr)
	}

	return tfVarsContent
}

func writeTfVars(tfVars map[string]interface{}, path string) diag.Diagnostics {
	tfVarsStr := generateTfVarsContent(tfVars)
	tfVarsStr = fmt.Sprintf("// This file has been autogenerated. The following properties could not be retrieved from the API or would not make sense in a different org e.g. Edge IDs"+
		"\n// The variables contained in this file have been given default values and should be edited as necessary\n\n%s", tfVarsStr)

	log.Printf("Writing export tfvars file to %s", path)
	return files.WriteToFile([]byte(tfVarsStr), path)
}
