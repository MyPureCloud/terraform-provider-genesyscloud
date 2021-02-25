package genesyscloud

import (
	"context"
	"strings"

	"github.com/MyPureCloud/platform-client-sdk-go/platformclientv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

// New initializes the provider schema
func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		return &schema.Provider{
			Schema: map[string]*schema.Schema{
				"oauthclient_id": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_OAUTHCLIENT_ID", nil),
					Description: "OAuthClient ID found on the OAuth page of Admin UI.",
				},
				"oauthclient_secret": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_OAUTHCLIENT_SECRET", nil),
					Description: "OAuthClient secret found on the OAuth page of Admin UI.",
					Sensitive:   true,
				},
				"aws_region": {
					Type:         schema.TypeString,
					Required:     true,
					DefaultFunc:  schema.EnvDefaultFunc("GENESYSCLOUD_REGION", nil),
					Description:  "AWS region where org exists. e.g. us-east-1",
					ValidateFunc: validation.StringInSlice(getAllowedRegions(), true),
				},
				"sdk_debug": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("GENESYSCLOUD_SDK_DEBUG", false),
					Description: "Enables debug tracing in the Genesys Cloud SDK.",
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"genesyscloud_user":          userResource(),
				"genesyscloud_routing_queue": routingQueueResource(),
				"genesyscloud_routing_skill": routingSkillResource(),
				"genesyscloud_auth_role":     authRoleResource(),
			},
			ConfigureContextFunc: configure,
		}
	}
}

type providerMeta struct {
}

func configure(context context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	oauthclientID := data.Get("oauthclient_id").(string)
	oauthclientSecret := data.Get("oauthclient_secret").(string)
	basePath := getRegionBasePath(data.Get("aws_region").(string))

	config := platformclientv2.GetDefaultConfiguration()
	config.BasePath = basePath
	config.SetDebug(data.Get("sdk_debug").(bool))

	err := config.AuthorizeClientCredentials(oauthclientID, oauthclientSecret)
	if err != nil {
		return nil, diag.Errorf("Failed to authorize Genesys Cloud client credentials")
	}

	return &providerMeta{}, nil
}

func getRegionMap() map[string]string {
	return map[string]string{
		"dca":            "https://api.inindca.com",
		"tca":            "https://api.inintca.com",
		"us-east-1":      "https://api.mypurecloud.com",
		"us-west-2":      "https://api.usw2.pure.cloud",
		"eu-west-1":      "https://api.mypurecloud.ie",
		"eu-west-2":      "https://api.euw2.pure.cloud",
		"ap-southeast-2": "https://api.mypurecloud.com.au",
		"ap-northeast-1": "https://api.mypurecloud.jp",
		"eu-central-1":   "https://api.mypurecloud.de",
		"ca-central-1":   "https://api.cac1.pure.cloud",
		"ap-northeast-2": "https://api.apne2.pure.cloud",
		"ap-south-1":     "https://api.aps1.pure.cloud",
	}
}

func getAllowedRegions() []string {
	regionMap := getRegionMap()
	regionKeys := make([]string, 0, len(regionMap))
	for k := range regionMap {
		regionKeys = append(regionKeys, k)
	}
	return regionKeys
}

func getRegionBasePath(region string) string {
	return getRegionMap()[strings.ToLower(region)]
}
