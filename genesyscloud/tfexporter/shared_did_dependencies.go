package tfexporter

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Resource types involved in shared DID ordering. They are declared here instead of being imported from the
// user and architect_ivr packages so that the exporter does not take a new dependency on those packages.
const (
	sharedDidUserResourceType = "genesyscloud_user"
	sharedDidIvrResourceType  = "genesyscloud_architect_ivr"
)

var (
	// A canonical number is a leading '+' followed only by digits.
	sharedDidCanonicalNumber = regexp.MustCompile(`^\+[0-9]+$`)
	// Matches the depends_on markers written by the exporter, e.g. $dep$genesyscloud_flow.my_flow$dep$.
	sharedDidDependsOnMarker = regexp.MustCompile(`\$dep\$([^$]+)\$dep\$`)
	// Matches resource references inside interpolations, e.g. ${genesyscloud_flow.my_flow.id}.
	// Data source references (${data.x.y.id}) also match with "data" as the type and are skipped by the caller.
	sharedDidInterpolatedReference = regexp.MustCompile(`\$\{([a-z0-9_]+)\.([A-Za-z0-9_-]+)\.`)
)

// sharedDidSkippedDependency describes a user -> IVR dependency that was not added because it would create a cycle.
type sharedDidSkippedDependency struct {
	UserBlockLabel string
	IvrBlockLabel  string
	Number         string
}

// addSharedDidDependencies adds a depends_on entry to every exported genesyscloud_user whose phone numbers
// match a DNIS on an exported genesyscloud_architect_ivr. The IVR API rejects a DID that a user already has,
// while the user API accepts a DID that an IVR already has, so the IVR must be created first.
//
// Terraform cannot infer this ordering on its own because both resources only hold the number as a literal
// string. This pass runs independently of enable_dependency_resolution and of the export filter mode.
func (g *GenesysCloudResourceExporter) addSharedDidDependencies() {
	g.resourceTypesMapsMutex.Lock()
	added, skipped := addSharedDidDependenciesToConfigMaps(g.resourceTypesMaps)
	g.resourceTypesMapsMutex.Unlock()

	logSharedDidResults(g.ctx, added, skipped)
}

func logSharedDidResults(ctx context.Context, added int, skipped []sharedDidSkippedDependency) {
	if ctx == nil {
		ctx = context.Background()
	}
	if added > 0 {
		tflog.Info(ctx, fmt.Sprintf("[addSharedDidDependencies] Added %d depends_on entries from %s to %s for shared DIDs",
			added, sharedDidUserResourceType, sharedDidIvrResourceType))
	}
	for _, s := range skipped {
		tflog.Warn(ctx, fmt.Sprintf("[addSharedDidDependencies] Not adding depends_on from %s.%s to %s.%s for DID %s because it would create a dependency cycle. Terraform may create the user before the IVR for this DID.",
			sharedDidUserResourceType, s.UserBlockLabel, sharedDidIvrResourceType, s.IvrBlockLabel, s.Number))
	}
}

// addSharedDidDependenciesToConfigMaps mutates the user config maps in resourceTypesMaps and returns the number of
// dependencies added together with any dependency that was skipped to avoid a cycle. Data sources are not part
// of resourceTypesMaps, so they never receive or provide a shared DID dependency.
func addSharedDidDependenciesToConfigMaps(resourceTypesMaps map[string]ResourceJSONMaps) (int, []sharedDidSkippedDependency) {
	ivrs := resourceTypesMaps[sharedDidIvrResourceType]
	users := resourceTypesMaps[sharedDidUserResourceType]
	if len(ivrs) == 0 || len(users) == 0 {
		return 0, nil
	}

	// Map each canonical DNIS to the IVR block labels that hold it.
	ivrLabelsByNumber := make(map[string][]string)
	for _, ivrLabel := range sortedConfigMapLabels(ivrs) {
		for _, number := range ivrDnisNumbers(ivrs[ivrLabel]) {
			ivrLabelsByNumber[number] = appendUniqueString(ivrLabelsByNumber[number], ivrLabel)
		}
	}
	if len(ivrLabelsByNumber) == 0 {
		return 0, nil
	}

	graph := buildSharedDidReferenceGraph(resourceTypesMaps)
	added := 0
	var skipped []sharedDidSkippedDependency

	for _, userLabel := range sortedConfigMapLabels(users) {
		userConfig := users[userLabel]
		userNode := sharedDidNode(sharedDidUserResourceType, userLabel)

		var newRefs []string
		for _, number := range userPhoneNumbers(userConfig) {
			for _, ivrLabel := range ivrLabelsByNumber[number] {
				ivrNode := sharedDidNode(sharedDidIvrResourceType, ivrLabel)
				if containsString(newRefs, ivrNode) {
					continue
				}
				// Adding user -> IVR closes a cycle if the IVR can already reach the user.
				if graph.reaches(ivrNode, userNode) {
					skipped = append(skipped, sharedDidSkippedDependency{UserBlockLabel: userLabel, IvrBlockLabel: ivrLabel, Number: number})
					continue
				}
				graph.addEdge(userNode, ivrNode)
				newRefs = append(newRefs, ivrNode)
			}
		}

		if len(newRefs) > 0 {
			added += mergeDependsOn(userConfig, newRefs)
		}
	}

	return added, skipped
}

// mergeDependsOn adds the given resource references to the config map's depends_on list using the exporter's
// depends_on marker format. Existing entries are kept and duplicates are not added. Returns the number added.
func mergeDependsOn(configMap map[string]interface{}, refs []string) int {
	existing := dependsOnEntries(configMap["depends_on"])
	present := make(map[string]bool, len(existing))
	for _, entry := range existing {
		present[dependsOnReference(entry)] = true
	}

	added := 0
	merged := append([]string{}, existing...)
	for _, ref := range refs {
		if present[ref] {
			continue
		}
		present[ref] = true
		merged = append(merged, fmt.Sprintf("$dep$%s$dep$", ref))
		added++
	}

	if added > 0 {
		configMap["depends_on"] = merged
	}
	return added
}

// dependsOnEntries returns the existing depends_on entries as strings. The exporter writes []string, but a
// []interface{} is accepted too.
func dependsOnEntries(value interface{}) []string {
	switch v := value.(type) {
	case []string:
		return append([]string{}, v...)
	case []interface{}:
		entries := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				entries = append(entries, s)
			}
		}
		return entries
	default:
		return nil
	}
}

// dependsOnReference strips the depends_on marker, so "$dep$a.b$dep$" and "a.b" compare as equal.
func dependsOnReference(entry string) string {
	if match := sharedDidDependsOnMarker.FindStringSubmatch(entry); len(match) == 2 {
		return match[1]
	}
	return entry
}

// ivrDnisNumbers returns the canonical DNIS numbers from an IVR config map.
func ivrDnisNumbers(configMap map[string]interface{}) []string {
	var numbers []string
	for _, value := range stringValues(configMap["dnis"]) {
		if number, ok := canonicalSharedDidNumber(value); ok {
			numbers = appendUniqueString(numbers, number)
		}
	}
	return numbers
}

// userPhoneNumbers returns the canonical phone numbers from a user config map (addresses.phone_numbers.number).
// Extension-only entries have no number and are ignored.
func userPhoneNumbers(configMap map[string]interface{}) []string {
	var numbers []string
	for _, address := range mapValues(configMap["addresses"]) {
		for _, phone := range mapValues(address["phone_numbers"]) {
			value, ok := phone["number"].(string)
			if !ok {
				continue
			}
			if number, ok := canonicalSharedDidNumber(value); ok {
				numbers = appendUniqueString(numbers, number)
			}
		}
	}
	sort.Strings(numbers)
	return numbers
}

// canonicalSharedDidNumber removes formatting characters and accepts the result only if it is a '+' followed by
// digits. Anything else (empty values, variable references, extensions) is not matched.
func canonicalSharedDidNumber(value string) (string, bool) {
	number := sanitizeE164Number(strings.TrimSpace(value))
	if !sharedDidCanonicalNumber.MatchString(number) {
		return "", false
	}
	return number, true
}

// sharedDidReferenceGraph is a directed graph of "type.label" nodes. An edge a -> b means a must wait for b.
type sharedDidReferenceGraph map[string]map[string]bool

func (graph sharedDidReferenceGraph) addEdge(from, to string) {
	if from == to {
		return
	}
	if graph[from] == nil {
		graph[from] = make(map[string]bool)
	}
	graph[from][to] = true
}

// reaches reports whether there is a path from start to target.
func (graph sharedDidReferenceGraph) reaches(start, target string) bool {
	if start == target {
		return true
	}
	visited := map[string]bool{start: true}
	stack := []string{start}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for next := range graph[node] {
			if next == target {
				return true
			}
			if !visited[next] {
				visited[next] = true
				stack = append(stack, next)
			}
		}
	}
	return false
}

// buildSharedDidReferenceGraph builds the dependency graph that Terraform will see for the exported resources:
// interpolated references such as ${genesyscloud_flow.x.id} and existing depends_on entries.
func buildSharedDidReferenceGraph(resourceTypesMaps map[string]ResourceJSONMaps) sharedDidReferenceGraph {
	graph := make(sharedDidReferenceGraph)
	for resourceType, configMaps := range resourceTypesMaps {
		for label, configMap := range configMaps {
			from := sharedDidNode(resourceType, label)
			collectReferences(configMap, func(to string) {
				graph.addEdge(from, to)
			})
		}
	}
	return graph
}

// collectReferences walks a config value and calls add for every resource it references.
func collectReferences(value interface{}, add func(string)) {
	switch v := value.(type) {
	case string:
		for _, match := range sharedDidDependsOnMarker.FindAllStringSubmatch(v, -1) {
			ref := match[1]
			if !strings.HasPrefix(ref, "data.") {
				add(ref)
			}
		}
		for _, match := range sharedDidInterpolatedReference.FindAllStringSubmatch(v, -1) {
			if match[1] == "data" || match[1] == "var" || match[1] == "local" {
				continue
			}
			add(sharedDidNode(match[1], match[2]))
		}
	case []string:
		for _, item := range v {
			collectReferences(item, add)
		}
	case []interface{}:
		for _, item := range v {
			collectReferences(item, add)
		}
	case map[string]interface{}:
		for _, item := range v {
			collectReferences(item, add)
		}
	case util.JsonMap:
		// Top-level config maps use the named util.JsonMap type, which does not match map[string]interface{}.
		collectReferences(map[string]interface{}(v), add)
	}
}

func sharedDidNode(resourceType, label string) string {
	return resourceType + "." + label
}

func sortedConfigMapLabels(configMaps ResourceJSONMaps) []string {
	labels := make([]string, 0, len(configMaps))
	for label := range configMaps {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	return labels
}

func stringValues(value interface{}) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []interface{}:
		values := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				values = append(values, s)
			}
		}
		return values
	default:
		return nil
	}
}

func mapValues(value interface{}) []map[string]interface{} {
	switch v := value.(type) {
	case []interface{}:
		values := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				values = append(values, m)
			}
		}
		return values
	case []map[string]interface{}:
		return v
	case map[string]interface{}:
		return []map[string]interface{}{v}
	case util.JsonMap:
		return []map[string]interface{}{v}
	default:
		return nil
	}
}

func appendUniqueString(values []string, value string) []string {
	if containsString(values, value) {
		return values
	}
	return append(values, value)
}

func containsString(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
