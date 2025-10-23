package user

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_user"

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	// Framework-only registration (SDKv2 removed)
	l.RegisterFrameworkResource(ResourceType, NewUserFrameworkResource)
	l.RegisterFrameworkDataSource(ResourceType, NewUserFrameworkDataSource)
	l.RegisterExporter(ResourceType, UserExporter())
}

// SDKv2 schema resources removed - migrated to Framework implementation

// customizeDiffAddressRemoval removed - SDKv2 function no longer needed

// ResourceUser removed - migrated to Framework implementation

// DataSourceUser removed - migrated to Framework implementation

func UserExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(GetAllUsersSDK),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"manager":                                   {RefType: ResourceType},
			"division_id":                               {RefType: "genesyscloud_auth_division"},
			"routing_skills.skill_id":                   {RefType: "genesyscloud_routing_skill"},
			"routing_languages.language_id":             {RefType: "genesyscloud_routing_language"},
			"locations.location_id":                     {RefType: "genesyscloud_location"},
			"addresses.phone_numbers.extension_pool_id": {RefType: "genesyscloud_telephony_providers_edges_extension_pool"},
		},
		RemoveIfMissing: map[string][]string{
			"routing_skills":         {"skill_id"},
			"routing_languages":      {"language_id"},
			"locations":              {"location_id"},
			"voicemail_userpolicies": {"alert_timeout_seconds"},
		},
		AllowEmptyArrays: []string{"routing_skills", "routing_languages"},
		AllowZeroValues:  []string{"routing_skills.proficiency", "routing_languages.proficiency"},
	}
}
