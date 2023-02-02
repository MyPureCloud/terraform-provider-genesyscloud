package genesyscloud

import (
	"context"
	"fmt"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

type ResourceMeta struct {
	// Name of the resoruce to be used in exports
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
}

func (r *ResourceExporter) loadSanitizedResourceMap(ctx context.Context, name string, filter []string) diag.Diagnostics {
	result, err := r.GetResourcesFunc(ctx)
	if err != nil {
		return err
	}

	if subStringInSlice(fmt.Sprintf("%v::", name), filter) {
		result = filterResources(result, name, filter)
	}

	r.SanitizedResourceMap = result
	sanitizeResourceNames(r.SanitizedResourceMap)
	return nil
}

func filterResources(result ResourceIDMetaMap, name string, filter []string) ResourceIDMetaMap {
	names := make([]string, 0)
	for _, f := range filter {
		n := fmt.Sprintf("%v::", name)
		if strings.Contains(f, n) {
			names = append(names, strings.Replace(f, n, "", 1))
		}
	}

	newResult := make(ResourceIDMetaMap)
	for _, name := range names {
		for k, v := range result {
			if v.Name == name {
				newResult[k] = v
			}
		}
	}
	return newResult
}

func (r *ResourceExporter) getRefAttrSettings(attribute string) *RefAttrSettings {
	if r.RefAttrs == nil {
		return nil
	}
	return r.RefAttrs[attribute]
}

func (r *ResourceExporter) getNestedRefAttrSettings(attribute string) *RefAttrSettings {
	for key, val := range r.EncodedRefAttrs {
		if key.NestedAttr == attribute {
			return val
		}
	}
	return nil
}

func (r *ResourceExporter) containsNestedRefAttrs(attribute string) ([]string, bool) {
	var nestedAttributes []string
	for key, _ := range r.EncodedRefAttrs {
		if key.Attr == attribute {
			nestedAttributes = append(nestedAttributes, key.NestedAttr)
		}
	}
	return nestedAttributes, len(nestedAttributes) > 0
}

func (r *ResourceExporter) allowZeroValues(attribute string) bool {
	return stringInSlice(attribute, r.AllowZeroValues)
}

func (r *ResourceExporter) isJsonEncodable(attribute string) bool {
	return stringInSlice(attribute, r.JsonEncodeAttributes)
}

func (r *ResourceExporter) addExcludedAttribute(attribute string) {
	r.ExcludedAttributes = append(r.ExcludedAttributes, attribute)
}

func (r *ResourceExporter) isAttributeExcluded(attribute string) bool {
	for _, excluded := range r.ExcludedAttributes {
		// Excluded if attributes match, or the specified attribute is nested in the excluded attribute
		if excluded == attribute || strings.HasPrefix(attribute, excluded+".") {
			return true
		}
	}
	return false
}

func (r *ResourceExporter) removeIfMissing(attribute string, config map[string]interface{}) bool {
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

func getResourceExporters(filter []string) map[string]*ResourceExporter {
	exporters := map[string]*ResourceExporter{
		// Add new resources that can be exported here
		"genesyscloud_architect_datatable":                             architectDatatableExporter(),
		"genesyscloud_architect_datatable_row":                         architectDatatableRowExporter(),
		"genesyscloud_architect_emergencygroup":                        architectEmergencyGroupExporter(),
		"genesyscloud_architect_ivr":                                   architectIvrExporter(),
		"genesyscloud_architect_schedules":                             architectSchedulesExporter(),
		"genesyscloud_architect_schedulegroups":                        architectScheduleGroupsExporter(),
		"genesyscloud_architect_user_prompt":                           architectUserPromptExporter(),
		"genesyscloud_auth_division":                                   authDivisionExporter(),
		"genesyscloud_auth_role":                                       authRoleExporter(),
		"genesyscloud_employeeperformance_externalmetrics_definitions": employeeperformanceExternalmetricsDefinitionExporter(),
		"genesyscloud_flow":                                            flowExporter(),
		"genesyscloud_flow_milestone":                                  flowMilestoneExporter(),
		"genesyscloud_flow_outcome":                                    flowOutcomeExporter(),
		"genesyscloud_group":                                           groupExporter(),
		"genesyscloud_group_roles":                                     groupRolesExporter(),
		"genesyscloud_idp_adfs":                                        idpAdfsExporter(),
		"genesyscloud_idp_generic":                                     idpGenericExporter(),
		"genesyscloud_idp_gsuite":                                      idpGsuiteExporter(),
		"genesyscloud_idp_okta":                                        idpOktaExporter(),
		"genesyscloud_idp_onelogin":                                    idpOneloginExporter(),
		"genesyscloud_idp_ping":                                        idpPingExporter(),
		"genesyscloud_idp_salesforce":                                  idpSalesforceExporter(),
		"genesyscloud_integration":                                     integrationExporter(),
		"genesyscloud_integration_action":                              integrationActionExporter(),
		"genesyscloud_integration_credential":                          credentialExporter(),
		"genesyscloud_journey_outcome":                                 journeyOutcomeExporter(),
		"genesyscloud_journey_segment":                                 journeySegmentExporter(),
		"genesyscloud_knowledge_knowledgebase":                         knowledgeKnowledgebaseExporter(),
		"genesyscloud_knowledge_document":                              knowledgeDocumentExporter(),
		"genesyscloud_knowledge_category":                              knowledgeCategoryExporter(),
		"genesyscloud_location":                                        locationExporter(),
		"genesyscloud_oauth_client":                                    oauthClientExporter(),
		"genesyscloud_outbound_attempt_limit":                          outboundAttemptLimitExporter(),
		"genesyscloud_outbound_callanalysisresponseset":                outboundCallAnalysisResponseSetExporter(),
		"genesyscloud_outbound_callabletimeset":                        outboundCallableTimesetExporter(),
		"genesyscloud_outbound_campaign":                               outboundCampaignExporter(),
		"genesyscloud_outbound_contact_list":                           outboundContactListExporter(),
		"genesyscloud_outbound_contactlistfilter":                      outboundContactListFilterExporter(),
		"genesyscloud_outbound_ruleset":                                outboundRulesetExporter(),
		"genesyscloud_outbound_messagingcampaign":                      outboundMessagingcampaignExporter(),
		"genesyscloud_outbound_sequence":                               outboundSequenceExporter(),
		"genesyscloud_outbound_dnclist":                                outboundDncListExporter(),
		"genesyscloud_outbound_campaignrule":                           outboundCampaignRuleExporter(),
		"genesyscloud_outbound_settings":                               outboundSettingsExporter(),
		"genesyscloud_outbound_wrapupcodemappings":                     outboundWrapupCodeMappingsExporter(),
		"genesyscloud_processautomation_trigger":                       processAutomationTriggerExporter(),
		"genesyscloud_quality_forms_evaluation":                        evaluationFormExporter(),
		"genesyscloud_quality_forms_survey":                            surveyFormExporter(),
		"genesyscloud_recording_media_retention_policy":                mediaRetentionPolicyExporter(),
		"genesyscloud_responsemanagement_library":                      responsemanagementLibraryExporter(),
		"genesyscloud_routing_email_domain":                            routingEmailDomainExporter(),
		"genesyscloud_routing_email_route":                             routingEmailRouteExporter(),
		"genesyscloud_routing_language":                                routingLanguageExporter(),
		"genesyscloud_routing_queue":                                   routingQueueExporter(),
		"genesyscloud_routing_settings":                                routingSettingsExporter(),
		"genesyscloud_routing_skill":                                   routingSkillExporter(),
		"genesyscloud_routing_skill_group":                             resourceSkillGroupExporter(),
		"genesyscloud_routing_utilization":                             routingUtilizationExporter(),
		"genesyscloud_routing_wrapupcode":                              routingWrapupCodeExporter(),
		"genesyscloud_telephony_providers_edges_did_pool":              telephonyDidPoolExporter(),
		"genesyscloud_telephony_providers_edges_edge_group":            edgeGroupExporter(),
		"genesyscloud_telephony_providers_edges_extension_pool":        telephonyExtensionPoolExporter(),
		"genesyscloud_telephony_providers_edges_phone":                 phoneExporter(),
		"genesyscloud_telephony_providers_edges_site":                  siteExporter(),
		"genesyscloud_telephony_providers_edges_phonebasesettings":     phoneBaseSettingsExporter(),
		"genesyscloud_telephony_providers_edges_trunkbasesettings":     trunkBaseSettingsExporter(),
		"genesyscloud_telephony_providers_edges_trunk":                 trunkExporter(),
		"genesyscloud_user":                                            userExporter(),
		"genesyscloud_user_roles":                                      userRolesExporter(),
		"genesyscloud_webdeployments_configuration":                    webDeploymentConfigurationExporter(),
		"genesyscloud_webdeployments_deployment":                       webDeploymentExporter(),
		"genesyscloud_widget_deployment":                               widgetDeploymentExporter(),
	}

	// Include all if no filters
	if len(filter) > 0 {
		for resType := range exporters {
			if !stringInSlice(resType, formatFilter(filter)) {
				delete(exporters, resType)
			}
		}
	}
	return exporters
}

// Removes the ::resource_name from the resource_types list
func formatFilter(filter []string) []string {
	newFilter := make([]string, 0)
	for _, str := range filter {
		newFilter = append(newFilter, strings.Split(str, "::")[0])
	}
	return newFilter
}

func getAvailableExporterTypes() []string {
	exporters := getResourceExporters(nil)
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
		meta.Name = sanitizeResourceName(meta.Name)
	}
}

func sanitizeResourceName(inputName string) string {
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
