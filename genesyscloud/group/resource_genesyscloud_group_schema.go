package group

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const resourceName = "genesyscloud_group"

var (
	groupPhoneType       = "PHONE"
	groupAddressResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"number": {
				Description:      "Phone number for this contact type. Must be in an E.164 number format.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validators.ValidatePhoneNumber,
			},
			"extension": {
				Description: "Phone extension.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description:  "Contact type of the address. (GROUPRING | GROUPPHONE)",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"GROUPRING", "GROUPPHONE"}, false),
			},
		},
	}
)

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceGroup())
	regInstance.RegisterDataSource(resourceName, DataSourceGroup())
	regInstance.RegisterExporter(resourceName, GroupExporter())
}

func GroupExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(GetAllGroups),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"owner_ids":  {RefType: "genesyscloud_user"},
			"member_ids": {RefType: "genesyscloud_user"},
		},
		CustomValidateExports: map[string][]string{
			"E164": {"addresses.number"},
		},
	}
}

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Directory Group",

		CreateContext: provider.CreateWithPooledClient(createGroup),
		ReadContext:   provider.ReadWithPooledClient(readGroup),
		UpdateContext: provider.UpdateWithPooledClient(updateGroup),
		DeleteContext: provider.DeleteWithPooledClient(deleteGroup),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Group description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description:  "Group type (official | social). This cannot be modified. Changing type attribute will cause the existing genesys_group object to dropped and recreated with a new ID.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "official",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"official", "social"}, false),
			},
			"visibility": {
				Description:  "Who can view this group (public | owners | members).",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "public",
				ValidateFunc: validation.StringInSlice([]string{"public", "owners", "members"}, false),
			},
			"rules_visible": {
				Description: "Are membership rules visible to the person requesting to view the group.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"addresses": {
				Description: "Contact numbers for this group.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        groupAddressResource,
			},
			"owner_ids": {
				Description: "IDs of owners of the group.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
			"member_ids": {
				Description: "IDs of members assigned to the group. If not set, this resource will not manage group members.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"roles_enabled": {
				Description: "Allow roles to be assigned to this group.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func DataSourceGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Groups. Select a group by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceGroupRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
