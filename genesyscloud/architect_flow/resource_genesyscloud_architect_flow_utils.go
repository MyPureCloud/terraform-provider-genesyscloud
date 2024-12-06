package architect_flow

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func isForceUnlockEnabled(d *schema.ResourceData) bool {
	forceUnlock := d.Get("force_unlock").(bool)
	log.Printf("ForceUnlock: %v, id %v", forceUnlock, d.Id())

	if forceUnlock && d.Id() != "" {
		return true
	}
	return false
}

func GenerateFlowResource(resourceLabel, srcFile, fileContent string, forceUnlock bool, substitutions ...string) string {
	fullyQualifiedPath, _ := testrunner.NormalizePath(srcFile)

	if fileContent != "" {
		updateFile(srcFile, fileContent)
	}

	flowResourceStr := fmt.Sprintf(`resource "genesyscloud_flow" "%s" {
        filepath = %s
		file_content_hash =  filesha256(%s)
		force_unlock = %v
		%s
	}
	`, resourceLabel, strconv.Quote(srcFile), strconv.Quote(fullyQualifiedPath), forceUnlock, strings.Join(substitutions, "\n"))

	return flowResourceStr
}

// setFileContentHashToNil This operation is required after a flow update fails because we want Terraform to detect changes
// in the file content hash and re-attempt an update, should the user re-run terraform apply without making changes to the file contents
func setFileContentHashToNil(d *schema.ResourceData) {
	_ = d.Set("file_content_hash", nil)
}
