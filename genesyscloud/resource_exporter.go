package genesyscloud

import (
	"context"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

var resourceExporters map[string]*ResourceExporter
var resourceExporterMapMutex = sync.RWMutex{}

func init() {
	resourceExporters = make(map[string]*ResourceExporter)
}

func RegisterExporter(exporterName string, resourceExporter *ResourceExporter) {
	resourceMapMutex.Lock()
	resourceExporters[exporterName] = resourceExporter
	resourceMapMutex.Unlock()
}

type ResourceMeta struct {
	// Name of the resource to be used in exports
	Name string

	// Prefix to add to the ID when reading state
	IdPrefix string
}

// ResourceIDMetaMap is a map of IDs to ResourceMeta
type ResourceIDMetaMap map[string]*ResourceMeta

// GetAllResourcesFunc is a method that returns all resource IDs
type GetAllResourcesFunc func(context.Context) (ResourceIDMetaMap, diag.Diagnostics)

// RefAttrSettings contains behavior settings for references
type RefAttrSettings struct {

	// Referenced resource type
	RefType string

	// Values that may be set that should not be treated as IDs
	AltValues []string
}

// Allows the definition of a custom resolver for an exporter.
type RefAttrCustomResolver struct {
	ResolverFunc func(map[string]interface{}, map[string]*ResourceExporter) error
}

// Allows the definition of a custom resolver for an exporter.
type CustomFlowResolver struct {
	ResolverFunc func(map[string]interface{}, string) error
}

type CustomFileWriterSettings struct {
	// Custom function for dumping data/media stored in an object in a sub directory along
	// with the exported config. For example: prompt audio files, csv data, jps/pngs
	RetrieveAndWriteFilesFunc func(string, string, string, map[string]interface{}, interface{}) error

	// Sub directory within export folder in which to write files retrieved by RetrieveAndWriteFilesFunc
	// For example, the user_prompt resource defines SubDirectory as "audio", so the prompt audio files will
	// be written to genesyscloud_tf_export.directory/audio/
	// The logic for retrieving and writing data to this dir should be defined in RetrieveAndWriteFilesFunc
	SubDirectory string
}

type JsonEncodeRefAttr struct {
	// The outer key
	Attr string

	// The RefAttr nested inside the json data
	NestedAttr string
}

// ResourceExporter is an interface to implement for resources that can be exported
type ResourceExporter struct {

	// Method to load all resource IDs for a given resource.
	// Returned map key should be the ID and the value should be a name to use for the resource.
	// Names will be sanitized with part of the ID appended, so it is not required that they be unique
	GetResourcesFunc GetAllResourcesFunc

	// A map of resource attributes to types that they reference
	// Attributes in nested objects can be defined with a '.' separator
	RefAttrs map[string]*RefAttrSettings

	// AllowZeroValues is a list of attributes that should allow zero values in the export.
	// By default zero values are removed from the config due to lack of "null" support in the plugin SDK
	AllowZeroValues []string

	// Some of our dependencies can not be exported properly because they have interdependencies between attributes.  You can
	// define a map of custom attribute resolvers with an exporter.  See resource_genesyscloud_routing_queue for an example of how to define this.
	// NOTE: CustomAttributeResolvers should be the exception and not the norm so use them when you have to do logic that will help you
	// resolve to the write reference
	CustomAttributeResolver map[string]*RefAttrCustomResolver

	// RemoveIfMissing is a map of attributes to a list of inner object attributes.
	// When all specified inner attributes are missing from an object, that object is removed
	RemoveIfMissing map[string][]string

	// Map of resource id->names. This is set after a call to loadSanitizedResourceMap
	SanitizedResourceMap ResourceIDMetaMap

	// List of attributes to exclude from config. This is set by the export configuration.
	ExcludedAttributes []string

	// Map of attributes that cannot be resolved. E.g. edge Ids which are locked to an org or properties that cannot be retrieved from the API
	UnResolvableAttributes map[string]*schema.Schema

	// List of attributes which can and should be exported in a jsonencode object rather than as a long escaped string of JSON data.
	JsonEncodeAttributes []string

	// Attributes that are jsonencode objects, and that contain nested RefAttrs
	EncodedRefAttrs map[*JsonEncodeRefAttr]*RefAttrSettings

	CustomFileWriter CustomFileWriterSettings

	CustomFlowResolver map[string]*CustomFlowResolver

	//This a place holder filter out specific resources from a filter.
	FilterResource func(ResourceIDMetaMap, string, []string) ResourceIDMetaMap
}

func (r *ResourceExporter) LoadSanitizedResourceMap(ctx context.Context, name string, filter []string) diag.Diagnostics {
	result, err := r.GetResourcesFunc(ctx)
	if err != nil {
		return err
	}

	if r.FilterResource != nil {
		result = r.FilterResource(result, name, filter)
	}

	r.SanitizedResourceMap = result
	sanitizeResourceNames(r.SanitizedResourceMap)
	return nil
}

func (r *ResourceExporter) GetRefAttrSettings(attribute string) *RefAttrSettings {
	if r.RefAttrs == nil {
		return nil
	}
	return r.RefAttrs[attribute]
}

func (r *ResourceExporter) GetNestedRefAttrSettings(attribute string) *RefAttrSettings {
	for key, val := range r.EncodedRefAttrs {
		if key.NestedAttr == attribute {
			return val
		}
	}
	return nil
}

func (r *ResourceExporter) ContainsNestedRefAttrs(attribute string) ([]string, bool) {
	var nestedAttributes []string
	for key, _ := range r.EncodedRefAttrs {
		if key.Attr == attribute {
			nestedAttributes = append(nestedAttributes, key.NestedAttr)
		}
	}
	return nestedAttributes, len(nestedAttributes) > 0
}

func (r *ResourceExporter) AllowForZeroValues(attribute string) bool {
	return StringInSlice(attribute, r.AllowZeroValues)
}

func (r *ResourceExporter) IsJsonEncodable(attribute string) bool {
	return StringInSlice(attribute, r.JsonEncodeAttributes)
}

func (r *ResourceExporter) AddExcludedAttribute(attribute string) {
	r.ExcludedAttributes = append(r.ExcludedAttributes, attribute)
}

func (r *ResourceExporter) IsAttributeExcluded(attribute string) bool {
	for _, excluded := range r.ExcludedAttributes {
		// Excluded if attributes match, or the specified attribute is nested in the excluded attribute
		if excluded == attribute || strings.HasPrefix(attribute, excluded+".") {
			return true
		}
	}
	return false
}

func (r *ResourceExporter) RemoveFieldIfMissing(attribute string, config map[string]interface{}) bool {
	if attrs, ok := r.RemoveIfMissing[attribute]; ok {
		// Check if all required inner attributes are missing
		missingAll := true
		for _, attr := range attrs {
			if val, foundInner := config[attr]; foundInner && val != nil {
				missingAll = false
				break
			}
		}
		return missingAll
	}
	return false
}

func GetResourceExporters() map[string]*ResourceExporter {

	RegisterExporter("genesyscloud_architect_datatable", architectDatatableExporter())
	RegisterExporter("genesyscloud_architect_datatable_row", architectDatatableRowExporter())
	RegisterExporter("genesyscloud_architect_emergencygroup", architectEmergencyGroupExporter())
	RegisterExporter("genesyscloud_architect_ivr", architectIvrExporter())
	RegisterExporter("genesyscloud_architect_schedules", architectSchedulesExporter())
	RegisterExporter("genesyscloud_architect_schedulegroups", architectScheduleGroupsExporter())
	RegisterExporter("genesyscloud_architect_user_prompt", architectUserPromptExporter())
	RegisterExporter("genesyscloud_auth_division", authDivisionExporter())
	RegisterExporter("genesyscloud_auth_role", authRoleExporter())
	RegisterExporter("genesyscloud_employeeperformance_externalmetrics_definitions", employeeperformanceExternalmetricsDefinitionExporter())
	RegisterExporter("genesyscloud_externalcontacts_contact", externalContactExporter())
	RegisterExporter("genesyscloud_flow", flowExporter())
	RegisterExporter("genesyscloud_flow_milestone", flowMilestoneExporter())
	RegisterExporter("genesyscloud_flow_outcome", flowOutcomeExporter())
	RegisterExporter("genesyscloud_group", groupExporter())
	RegisterExporter("genesyscloud_group_roles", groupRolesExporter())
	RegisterExporter("genesyscloud_idp_adfs", idpAdfsExporter())
	RegisterExporter("genesyscloud_idp_generic", idpGenericExporter())
	RegisterExporter("genesyscloud_idp_gsuite", idpGsuiteExporter())
	RegisterExporter("genesyscloud_idp_okta", idpOktaExporter())
	RegisterExporter("genesyscloud_idp_onelogin", idpOneloginExporter())
	RegisterExporter("genesyscloud_idp_ping", idpPingExporter())
	RegisterExporter("genesyscloud_idp_salesforce", idpSalesforceExporter())
	RegisterExporter("genesyscloud_integration", integrationExporter())
	RegisterExporter("genesyscloud_integration_action", integrationActionExporter())
	RegisterExporter("genesyscloud_integration_credential", credentialExporter())
	RegisterExporter("genesyscloud_journey_action_map", journeyActionMapExporter())
	RegisterExporter("genesyscloud_journey_action_template", journeyActionTemplateExporter())
	RegisterExporter("genesyscloud_journey_outcome", journeyOutcomeExporter())
	RegisterExporter("genesyscloud_journey_segment", journeySegmentExporter())
	RegisterExporter("genesyscloud_knowledge_knowledgebase", knowledgeKnowledgebaseExporter())
	RegisterExporter("genesyscloud_knowledge_document", knowledgeDocumentExporter())
	RegisterExporter("genesyscloud_knowledge_v1_document", knowledgeDocumentExporterV1())
	RegisterExporter("genesyscloud_knowledge_document_variation", knowledgeDocumentVariationExporter())
	RegisterExporter("genesyscloud_knowledge_category", knowledgeCategoryExporter())
	RegisterExporter("genesyscloud_knowledge_v1_category", knowledgeCategoryExporterV1())
	RegisterExporter("genesyscloud_knowledge_label", knowledgeLabelExporter())
	RegisterExporter("genesyscloud_location", locationExporter())
	RegisterExporter("genesyscloud_oauth_client", oauthClientExporter())
	RegisterExporter("genesyscloud_outbound_attempt_limit", outboundAttemptLimitExporter())
	RegisterExporter("genesyscloud_outbound_callanalysisresponseset", outboundCallAnalysisResponseSetExporter())
	RegisterExporter("genesyscloud_outbound_callabletimeset", outboundCallableTimesetExporter())
	RegisterExporter("genesyscloud_outbound_campaign", outboundCampaignExporter())
	RegisterExporter("genesyscloud_outbound_contact_list", outboundContactListExporter())
	RegisterExporter("genesyscloud_outbound_contactlistfilter", outboundContactListFilterExporter())
	RegisterExporter("genesyscloud_outbound_ruleset", outboundRulesetExporter())
	RegisterExporter("genesyscloud_outbound_messagingcampaign", outboundMessagingcampaignExporter())
	RegisterExporter("genesyscloud_outbound_sequence", outboundSequenceExporter())
	RegisterExporter("genesyscloud_outbound_dnclist", outboundDncListExporter())
	RegisterExporter("genesyscloud_outbound_campaignrule", outboundCampaignRuleExporter())
	RegisterExporter("genesyscloud_outbound_settings", outboundSettingsExporter())
	RegisterExporter("genesyscloud_outbound_wrapupcodemappings", outboundWrapupCodeMappingsExporter())
	RegisterExporter("genesyscloud_quality_forms_evaluation", evaluationFormExporter())
	RegisterExporter("genesyscloud_quality_forms_survey", surveyFormExporter())
	RegisterExporter("genesyscloud_recording_media_retention_policy", mediaRetentionPolicyExporter())
	RegisterExporter("genesyscloud_responsemanagement_library", responsemanagementLibraryExporter())
	RegisterExporter("genesyscloud_responsemanagement_response", responsemanagementResponseExporter())
	RegisterExporter("genesyscloud_routing_email_domain", routingEmailDomainExporter())
	RegisterExporter("genesyscloud_routing_email_route", routingEmailRouteExporter())
	RegisterExporter("genesyscloud_routing_language", routingLanguageExporter())
	RegisterExporter("genesyscloud_routing_queue", routingQueueExporter())
	RegisterExporter("genesyscloud_routing_settings", routingSettingsExporter())
	RegisterExporter("genesyscloud_routing_skill", routingSkillExporter())
	RegisterExporter("genesyscloud_routing_skill_group", resourceSkillGroupExporter())
	RegisterExporter("genesyscloud_routing_sms_address", routingSmsAddressExporter())
	RegisterExporter("genesyscloud_routing_utilization", routingUtilizationExporter())
	RegisterExporter("genesyscloud_routing_wrapupcode", routingWrapupCodeExporter())
	RegisterExporter("genesyscloud_script", scriptExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_did_pool", telephonyDidPoolExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_edge_group", edgeGroupExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_extension_pool", telephonyExtensionPoolExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_phone", phoneExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_site", siteExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_phonebasesettings", phoneBaseSettingsExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_trunkbasesettings", trunkBaseSettingsExporter())
	RegisterExporter("genesyscloud_telephony_providers_edges_trunk", trunkExporter())
	RegisterExporter("genesyscloud_user", userExporter())
	RegisterExporter("genesyscloud_user_roles", userRolesExporter())
	RegisterExporter("genesyscloud_webdeployments_configuration", webDeploymentConfigurationExporter())
	RegisterExporter("genesyscloud_webdeployments_deployment", webDeploymentExporter())
	RegisterExporter("genesyscloud_widget_deployment", widgetDeploymentExporter())

	//Make a Copy of the Map
	exportCopy := make(map[string]*ResourceExporter, len(resourceExporters))

	for k, v := range resourceExporters {
		exportCopy[k] = v
	}
	return exportCopy
}

func GetAvailableExporterTypes() []string {
	exporters := GetResourceExporters()
	types := make([]string, len(exporters))
	i := 0
	for k := range exporters {
		types[i] = k
		i++
	}
	return types
}

func escapeRune(s string) string {
	// Always replace with an underscore for readability. The appended hash will help ensure uniqueness
	return "_"
}

// Resource names must only contain alphanumeric chars, underscores, or dashes
// https://www.terraform.io/docs/language/syntax/configuration.html#identifiers
var unsafeNameChars = regexp.MustCompile(`[^0-9A-Za-z_-]`)

func sanitizeResourceNames(idMetaMap ResourceIDMetaMap) {
	for _, meta := range idMetaMap {
		meta.Name = SanitizeResourceName(meta.Name)
	}
}

func SanitizeResourceName(inputName string) string {
	name := unsafeNameChars.ReplaceAllStringFunc(inputName, escapeRune)
	if name != inputName {
		// Append a hash of the original name to ensure uniqueness for similar names
		// and that equivalent names are consistent across orgs
		algorithm := fnv.New32()
		algorithm.Write([]byte(inputName))
		name = name + "_" + strconv.FormatUint(uint64(algorithm.Sum32()), 10)
	}
	if unicode.IsDigit(rune(name[0])) {
		// Terraform does not allow names to begin with a number. Prefix with an underscore instead
		name = "_" + name
	}

	return name
}
