package tfexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/platform"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/files"

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
		return nil, fmt.Errorf("resources cannot be empty")
	}
	tfwriter := &TFStateFileWriter{
		ctx:              ctx,
		resources:        resources,
		d:                d,
		providerRegistry: providerRegistry,
	}

	return tfwriter, nil
}

func (t *TFStateFileWriter) writeTfState() diag.Diagnostics {

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

	platform := platform.GetPlatform()
	platformErr := platform.Validate()
	if platformErr != nil {
		log.Printf("Failed to validate platform: %v", platformErr)
		err := t.writeTfStateLegacy(stateFilePath)
		if err != nil {
			return nil
		}
	}

	// This outputs terraform state v3, and there is currently no public lib to generate v4 which is required for terraform 0.13+.
	// However, the state can be upgraded automatically by calling the terraform CLI. If this fails, just print a warning indicating
	// that the state likely needs to be upgraded manually.
	cliErrorPostscript := fmt.Sprintf(`The generated tfstate file will need to be upgraded manually by running the following in the state file's directory:
	'%s state replace-provider %s/-/genesyscloud %s/mypurecloud/genesyscloud'`, platform.Binary(), platform.GetProviderRegistry(), t.providerRegistry)

	if platform.IsDevelopmentPlatform() {
		cliErrorPostscript = `The current process is running via a debug server (debug binary detected), and so it is unable to run the proper command to replace the state. Please run this command outside of a debug session.` + cliErrorPostscript
		log.Print(cliErrorPostscript)
		return nil
	}

	replaceProviderOutput, err := platform.ExecuteCommand(t.ctx, []string{
		"state",
		"replace-provider",
		"-auto-approve",
		"-state=" + stateFilePath,
		// This is the provider determined by the platform (terraform vs tofu)
		fmt.Sprintf("%s/-/genesyscloud", platform.GetProviderRegistry()),
		// This is the platform that accounts for custom builds
		fmt.Sprintf("%s/mypurecloud/genesyscloud", t.providerRegistry),
	}...)
	log.Print(replaceProviderOutput.Stdout)

	if err != nil {
		cliErrorPostscript = fmt.Sprintf(`Failed to run the terraform CLI to upgrade the generated state file:
			Error: %v

			%s`, err, cliErrorPostscript)
		log.Print(cliErrorPostscript)
		// Don't fail everything even if this errors.
		err := t.writeTfStateLegacy(stateFilePath)
		if err != nil {
			return nil
		}
		return nil
	}
	return nil
}

func (t *TFStateFileWriter) writeTfStateLegacy(stateFilePath string) error {
	cliError := `Failed to run the terraform CLI to upgrade the generated state file.
	The generated tfstate file will need to be upgraded manually by running the
	following in the state file's directory:
	'terraform state replace-provider registry.terraform.io/-/genesyscloud registry.terraform.io/mypurecloud/genesyscloud'`

	tfpath, err := exec.LookPath("terraform")
	if err != nil {
		log.Println("Failed to find terraform path:", err)
		log.Println(cliError)
		return nil
	}

	// exec.CommandContext does not auto-resolve symlinks
	fileInfo, err := os.Lstat(tfpath)
	if err != nil {
		log.Println("Failed to Lstat terraform path:", err)
		log.Println(cliError)
		return nil
	}
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		tfpath, err = filepath.EvalSymlinks(tfpath)
		if err != nil {
			log.Println("Failed to resolve terraform path symlink:", err)
			log.Println(cliError)
			return nil
		}
	}

	cmd := exec.CommandContext(t.ctx, tfpath)
	cmd.Args = append(cmd.Args, []string{
		"state",
		"replace-provider",
		"-auto-approve",
		"-state=" + stateFilePath,
		"registry.terraform.io/-/genesyscloud",
		t.providerRegistry,
	}...)
	log.Printf("Running 'terraform state replace-provider' on %s", stateFilePath)
	if err = cmd.Run(); err != nil {
		log.Println("Failed to run command:", err)
		log.Println(cliError)
		return nil
	}
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
