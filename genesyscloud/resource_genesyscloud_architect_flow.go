package genesyscloud

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v46/platformclientv2"
)

var apiurl string = `https://uzr7bby4z7.execute-api.us-east-1.amazonaws.com/deploy/`
var apikey string = "IoFDHjO6LS9JNnunoUyKw2oXAjrso2kQ5a9A6JcD"

func getAllFlows(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	architectAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		flows, _, err := architectAPI.GetFlows(nil, pageNum, 25, "", "", nil, "", "", "", "", "", "", "", "", false, true, "", "", nil)
		if err != nil {
			return nil, diag.Errorf("Failed to get page of flows: %v", err)
		}

		if flows.Entities == nil || len(*flows.Entities) == 0 {
			break
		}

		for _, flow := range *flows.Entities {
			resources[*flow.Id] = &ResourceMeta{Name: *flow.Name}
		}
	}

	return resources, nil
}

func architectFlowExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllFlows),
		RefAttrs:         map[string]*RefAttrSettings{},
	}
}

func resourceFlow() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Flow",

		CreateContext: createWithPooledClient(createFlow),
		ReadContext:   readWithPooledClient(readFlow),
		UpdateContext: updateWithPooledClient(updateFlow),
		DeleteContext: deleteWithPooledClient(deleteFlow),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "Description for the flow. This won't affect the flow at all. Configuration of the flow should be in the provided file path. ",
				Type:        schema.TypeString,
				Required:    true,
			},
			"filepath": {
				Description: "YAML file path for flow configuration. ",
				Type:        schema.TypeString,
				Required:    true,
				StateFunc: func(v interface{}) string {
					return hashFileContent(v.(string))
				},
			},
			"debug": {
				Description: "Boolean value for debug mode. ",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"force_unlock": {
				Description: "Whether to force unlocking the flow. ",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"recreate": {
				Description: "Whether to recreate the flow. ",
				Type:        schema.TypeBool,
				Required:    true,
			},
		},
	}
}

func createFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clientID := os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID")
	clientSecret := os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET")
	location := getLocationForArchy(os.Getenv("GENESYSCLOUD_REGION"))
	debug := strconv.FormatBool(d.Get("debug").(bool))
	forceUnlock := strconv.FormatBool(d.Get("force_unlock").(bool))
	recreate := strconv.FormatBool(d.Get("recreate").(bool))
	filepath := d.Get("filepath").(string)

	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return diag.Errorf("Failed to read configuration file from this path: %s. %s", filepath, err)
	}
	filecontent := string(file)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"command": "publish",
		"files": map[string]interface{}{
			"main.yaml": filecontent,
		},
		"settings": map[string]interface{}{
			"clientSecret": &clientSecret,
			"clientId":     &clientID,
			"debug":        &debug,
			"forceUnlock":  &forceUnlock,
			"location":     &location,
			"recreate":     &recreate,
		},
	})

	body, err := sendRequest(requestBody)

	if err != nil {
		return diag.Errorf(err.Error())
	}

	var result map[string]interface{}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return diag.Errorf("Failed to unmarshal response body when creating flow. %s", err)
	}

	if result["statusCode"] == nil || int(result["statusCode"].(float64)) != http.StatusOK {
		return diag.Errorf("Create flow Request failed. Result: %s", string(body))
	}

	flowID := getFlowID((result["body"]).(map[string]interface{})["stdout"].(string))

	d.SetId(flowID)

	log.Printf("Created flow %s. ", d.Id())
	return readFlow(ctx, d, meta)
}

func readFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading flow %s", d.Id())
	asdf := d.Id()
	log.Printf(asdf)
	flow, resp, err := architectAPI.GetFlow(d.Id(), false)
	if err != nil {
		if isStatus404(resp) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read flow %s: %s", d.Id(), err)
	}

	description := fmt.Sprintf("Flow name: %s, Flow type: %s", *flow.Name, *flow.VarType)

	d.Set("description", description)

	log.Printf("Read flow %s %s", d.Id(), *flow.Name)
	return nil
}

func updateFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clientID := os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID")
	clientSecret := os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET")
	location := getLocationForArchy(os.Getenv("GENESYSCLOUD_REGION"))
	debug := strconv.FormatBool(d.Get("debug").(bool))
	forceUnlock := strconv.FormatBool(d.Get("force_unlock").(bool))
	recreate := strconv.FormatBool(d.Get("recreate").(bool))
	filepath := d.Get("filepath").(string)

	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return diag.Errorf("Failed to read configuration file from this path: %s. %s", filepath, err)
	}
	filecontent := string(file)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"command": "update",
		"files": map[string]interface{}{
			"main.yaml": filecontent,
		},
		"settings": map[string]interface{}{
			"clientSecret": &clientSecret,
			"clientId":     &clientID,
			"debug":        &debug,
			"forceUnlock":  &forceUnlock,
			"location":     &location,
			"recreate":     &recreate,
		},
	})

	body, err := sendRequest(requestBody)

	if err != nil {
		return diag.Errorf(err.Error())
	}

	var result map[string]interface{}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return diag.Errorf("Failed to unmarshal response body when updating flow. %s", err)
	}
	if result["statusCode"] == nil || int(result["statusCode"].(float64)) != http.StatusOK {
		return diag.Errorf("Update flow Request failed. Result: %s", string(body))
	}

	flowID := getFlowID((result["body"]).(map[string]interface{})["stdout"].(string))

	d.SetId(flowID)

	log.Printf("Updated flow %s", d.Id())
	return readFlow(ctx, d, meta)
}

func deleteFlow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	_, err := architectAPI.DeleteFlow(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete the flow %s: %s", d.Id(), err)
	}
	log.Printf("Deleted flow %s", d.Id())
	return nil
}

func hashFileContent(path string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func sendRequest(body []byte) ([]byte, error) {
	req, _ := http.NewRequest("POST", apiurl, bytes.NewBuffer(body))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apikey)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to make API request to update flow. %s", err)
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body when updating flow. %s", err)
	}

	return response, nil
}

func getFlowID(response string) string {
	regex := regexp.MustCompile(`Flow Id: (\b[a-fA-F0-9]{8}\b-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-\b[a-fA-F0-9]{12}\b)`)
	regexResult := regex.FindAllStringSubmatch(response, -1)
	var flowID string

	if len(regexResult) > 0 {
		flowID = regexResult[0][1]
	}

	return flowID
}

func getLocationForArchy(region string) string {
	switch region {
	// TODO: Can add more regions here when Archy support them
	case "":
		return ""
	default:
		return "dev"
	}
}
