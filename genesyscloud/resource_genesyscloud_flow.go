package genesyscloud

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"

	"gopkg.in/yaml.v2"
)

func getAllFlows(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	architectAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 50
		flows, _, err := architectAPI.GetFlows(nil, pageNum, pageSize, "", "", nil, "", "", "", "", "", "", "", "", false, true, "", "", nil)
		if err != nil {
			return nil, diag.Errorf("Failed to get page of flows: %v", err)
		}

		if flows.Entities == nil || len(*flows.Entities) == 0 {
			break
		}

		for _, flow := range *flows.Entities {
			resources[*flow.Id] = &resourceExporter.ResourceMeta{Name: *flow.Name}
		}
	}

	return resources, nil
}

func FlowExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllFlows),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
		UnResolvableAttributes: map[string]*schema.Schema{
			"filepath": ResourceFlow().Schema["filepath"],
		},
		CustomFlowResolver: map[string]*resourceExporter.CustomFlowResolver{
			"file_content_hash": {ResolverFunc: resourceExporter.FileContentHashResolver},
		},
	}
}

func ResourceFlow() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Flow`,

		CreateContext: CreateWithPooledClient(createFlow),
		UpdateContext: UpdateWithPooledClient(updateFlow),
		ReadContext:   ReadWithPooledClient(readFlow),
		DeleteContext: DeleteWithPooledClient(deleteFlow),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"filepath": {
				Description:  "YAML file path for flow configuration. Note: Changing the flow name will result in the creation of a new flow with a new GUID, while the original flow will persist in your org.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: ValidatePath,
			},
			"file_content_hash": {
				Description: "Hash value of the YAML file content. Used to detect changes.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"substitutions": {
				Description: "A substitution is a key value pair where the key is the value you want to replace, and the value is the value to substitute in its place.",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"force_unlock": {
				Description: `Will perform a force unlock on an architect flow before beginning the publication process.  NOTE: The force unlock publishes the 'draft'
				              architect flow and then publishes the flow named in this resource. This mirrors the behavior found in the archy CLI tool.`,
				Type:     schema.TypeBool,
				Optional: true,
			},
			"flow_name": {
				Description: `Genesys Cloud flow name. The value must be equal to the flow name set in the config yaml or to the substitution variable value.`,
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"flow_type": {
				Description: `Genesys Cloud flow type. The value must be equal to the flow type set in the config yaml.`,
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func readFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		flow, resp, err := architectAPI.GetFlow(d.Id(), false)
		if err != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read flow %s: %s", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read flow %s: %s", d.Id(), err))
		}
		d.Set("flow_name", *flow.Name)
		d.Set("flow_type", *flow.VarType)

		log.Printf("Read flow %s %s", d.Id(), *flow.Name)
		return nil
	})
}

func forceUnlockFlow(flowId string, sdkConfig *platformclientv2.Configuration) error {
	log.Printf("Attempting to perform an unlock on flow: %s", flowId)
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)
	_, _, err := architectAPI.PostFlowsActionsUnlock(flowId)

	if err != nil {
		return err
	}
	return nil
}

func isForceUnlockEnabled(d *schema.ResourceData) bool {
	forceUnlock := d.Get("force_unlock").(bool)
	log.Printf("ForceUnlock: %v, id %v", forceUnlock, d.Id())

	if forceUnlock && d.Id() != "" {
		return true
	}
	return false
}

func createFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating flow")
	return updateFlow(ctx, d, meta)
}

func updateFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	flowName := d.Get("flow_name").(string)
	flowType := d.Get("flow_type").(string)
	log.Printf("Updating flow name '%s' type '%s'", flowName, flowType)

	//Check to see if we need to force and unlock on an architect flow
	if isForceUnlockEnabled(d) {
		err := forceUnlockFlow(d.Id(), sdkConfig)
		if err != nil {
			setFileContentHashToNil(d)
			return diag.Errorf("Failed to unlock targeted flow %s with error %s", d.Id(), err)
		}
	}

	flowJob, response, err := architectAPI.PostFlowsJobs()

	if err != nil {
		setFileContentHashToNil(d)
		return diag.Errorf("Failed to update job %s", err)
	}

	if err == nil && response.Error != nil {
		setFileContentHashToNil(d)
		return diag.Errorf("Failed to register job. %s", err)
	}

	presignedUrl := *flowJob.PresignedUrl
	jobId := *flowJob.Id
	headers := *flowJob.Headers

	filePath := d.Get("filepath").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})

	reader, _, err := files.DownloadOrOpenFile(filePath)
	if err != nil {
		setFileContentHashToNil(d)
		return diag.Errorf(err.Error())
	}

	// DEVTOOLING-355: Underscore in flow names prevent accurate assignment of the flow config yaml with the terraform resource
	if err = compareFlowInfo(reader, flowName, flowType, substitutions); err != nil {
		return diag.Errorf(err.Error())
	}

	// reader was consumed, reset it to the beginning
	if reader, err = ResetReader(reader); err != nil {
		return diag.Errorf(err.Error())
	}

	s3Uploader := files.NewS3Uploader(reader, nil, substitutions, headers, "PUT", presignedUrl)
	_, err = s3Uploader.Upload()
	if err != nil {
		setFileContentHashToNil(d)
		return diag.Errorf(err.Error())
	}

	// Pre-define here before entering retry function, otherwise it will be overwritten
	flowID := ""

	retryErr := WithRetries(ctx, 16*time.Minute, func() *retry.RetryError {
		flowJob, response, err := architectAPI.GetFlowsJob(jobId, []string{"messages"})
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error retrieving job status. JobID: %s, error: %s ", jobId, response.ErrorMessage))
		}

		if *flowJob.Status == "Failure" {
			if flowJob.Messages == nil {
				return retry.NonRetryableError(fmt.Errorf("Flow publish failed. JobID: %s, no tracing messages available.", jobId))
			}
			messages := make([]string, 0)
			for _, m := range *flowJob.Messages {
				messages = append(messages, *m.Text)
			}
			return retry.NonRetryableError(fmt.Errorf("Flow publish failed. JobID: %s, tracing messages: %v ", jobId, strings.Join(messages, "\n\n")))
		}

		if *flowJob.Status == "Success" {
			flowID = *flowJob.Flow.Id
			return nil
		}

		time.Sleep(15 * time.Second) // Wait 15 seconds for next retry
		return retry.RetryableError(fmt.Errorf("Job (%s) could not finish in 16 minutes and timed out ", jobId))
	})

	if retryErr != nil {
		setFileContentHashToNil(d)
		return retryErr
	}

	if flowID == "" {
		setFileContentHashToNil(d)
		return diag.Errorf("Failed to get the flowId from Architect Job (%s).", jobId)
	}

	d.SetId(flowID)

	log.Printf("Updated flow %s. ", d.Id())
	return readFlow(ctx, d, meta)
}

func deleteFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	//Check to see if we need to force
	if isForceUnlockEnabled(d) {
		err := forceUnlockFlow(d.Id(), sdkConfig)
		if err != nil {
			return diag.Errorf("Failed to unlock targeted flow %s with error %s", d.Id(), err)
		}
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		resp, err := architectAPI.DeleteFlow(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Flow deleted
				log.Printf("Deleted Flow %s", d.Id())
				return nil
			}
			if resp.StatusCode == http.StatusConflict {
				return retry.RetryableError(fmt.Errorf("Error deleting flow %s: %s", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting flow %s: %s", d.Id(), err))
		}
		return nil
	})
}

func GenerateFlowResource(resourceID, srcFile, filecontent string, force_unlock bool, substitutions ...string) string {
	fullyQualifiedPath, _ := filepath.Abs(srcFile)

	if filecontent != "" {
		updateFile(srcFile, filecontent)
	}

	flowResourceStr := fmt.Sprintf(`resource "genesyscloud_flow" "%s" {
        filepath = %s
		file_content_hash =  filesha256(%s)
		force_unlock = %v
		%s
	}
	`, resourceID, strconv.Quote(srcFile), strconv.Quote(fullyQualifiedPath), force_unlock, strings.Join(substitutions, "\n"))

	return flowResourceStr
}

func updateFile(filepath, content string) {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	file.WriteString(content)
}

// setFileContentHashToNil This operation is required after a flow update fails because we want Terraform to detect changes
// in the file content hash and re-attempt an update, should the user re-run terraform apply without making changes to the file contents
func setFileContentHashToNil(d *schema.ResourceData) {
	_ = d.Set("file_content_hash", nil)
}

// Compares the flow name & type from the terraform resource and see if they match with the values set in the yaml config
func compareFlowInfo(reader io.Reader, flowNameAttrValue string, flowTypeAttrValue string, substitutions map[string]interface{}) error {
	// Unmarshal YAML content into a map
	var data map[interface{}]interface{}
	if err := yaml.NewDecoder(reader).Decode(&data); err != nil {
		return fmt.Errorf("error decoding YAML content: %v", err)
	}

	if err := checkFlowType(data, flowTypeAttrValue); err != nil {
		return err
	}

	if err := checkFlowName(data, flowNameAttrValue, substitutions); err != nil {
		return err
	}

	return nil
}

// checkFlowType checks if the provided flow type attribute matches the flow type in the YAML data.
func checkFlowType(data map[interface{}]interface{}, flowTypeAttrValue string) error {
	if flowTypeAttrValue == "" { // attr is optional
		return nil
	}

	yamlFlowType, err := getRootNodeKey(data)
	if err != nil {
		return fmt.Errorf("invalid flow config yaml: %s", err)
	}
	yamlFlowType = strings.TrimSuffix(yamlFlowType, ":")
	if !strings.EqualFold(yamlFlowType, flowTypeAttrValue) {
		return fmt.Errorf("flow type provided '%s' does not match the flow type set within yaml '%s'", flowTypeAttrValue, yamlFlowType)
	}

	return nil
}

// checkFlowName checks if the provided flow name attribute matches the flow name in the YAML data.
func checkFlowName(data map[interface{}]interface{}, flowNameAttrValue string, substitutions map[string]interface{}) error {
	if flowNameAttrValue == "" { // attr is optional
		return nil
	}

	yamlFlowType, err := getRootNodeKey(data)
	if err != nil {
		return fmt.Errorf("invalid flow config yaml: %s", err)
	}
	yamlFlowType = strings.TrimSuffix(yamlFlowType, ":")

	// Extract flow name from YAML
	var flowNameYaml string
	flow, ok := data[yamlFlowType].(map[interface{}]interface{})
	if ok {
		if name, ok := flow["name"].(string); ok {
			log.Printf("flow name property value: %s", name)
			flowNameYaml = name
		}
	}

	if flowNameYaml == "" {
		return fmt.Errorf("invalid flow name property value: '%s'", flowNameYaml)
	}

	if isSubVariable(flowNameYaml) { // Check if the flow name value in config YAML is a substitution variable
		flowNameYaml, _ = extractSubVariableStringValue(flowNameYaml)
		// Check if substitution key exists and its value matches the 'flow_name' attribute value
		if value, ok := substitutions[flowNameYaml]; ok {
			if flowNameAttrValue != value {
				return fmt.Errorf("'flow_name' attribute value '%s' does not match substitution key '%s' value '%s'", flowNameAttrValue, flowNameYaml, value)
			}
		} else {
			return fmt.Errorf("substitution key '%s' found in the flow config yaml does not exist in the flow resource substitutions config map", flowNameYaml)
		}
	} else { // Check if flow name in YAML matches the 'flow_name' attribute value
		if flowNameAttrValue != flowNameYaml {
			return fmt.Errorf("'flow_name' attribute value '%s' does not match the flow name set in the flow config yaml: '%s'", flowNameAttrValue, flowNameYaml)
		}
	}

	return nil
}

// Checks if the input string represents a substitution variable enclosed within double curly braces (e.g. {{variable}} )
func isSubVariable(input string) bool {
	re := regexp.MustCompile(`^\{\{([^{}]+)\}\}$`)
	return re.MatchString(input)
}

// Retrieves the value of the first node in a YAML config
func getRootNodeKey(data map[interface{}]interface{}) (string, error) {
	for key := range data {
		if keyStr, ok := key.(string); ok {
			log.Printf("Root node key found: %s", keyStr)
			return keyStr, nil
		}
	}
	return "", fmt.Errorf("could not find the root node key in the flow config yaml")
}

// Extracts the content inside double curly braces, assuming it represents a substitution variable (e.g. {{variable}} )
func extractSubVariableStringValue(input string) (string, error) {
	re := regexp.MustCompile(`\{\{(.+?)\}\}`)
	match := re.FindStringSubmatch(input)

	if len(match) < 2 {
		return "", fmt.Errorf("no match found")
	}

	content := match[1]
	return content, nil
}
