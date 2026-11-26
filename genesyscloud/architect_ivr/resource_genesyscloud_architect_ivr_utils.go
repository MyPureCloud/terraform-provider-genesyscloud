package architect_ivr

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

type IvrConfigStruct struct {
	ResourceLabel string
	Name          string
	Description   string
	Dnis          []string
	DependsOn     string
	DivisionId    string
}

// GenerateIvrConfigResource returns an ivr resource as a string based on the IvrConfigStruct struct
func GenerateIvrConfigResource(ivrConfig *IvrConfigStruct) string {
	var quotedDnsSlice []string
	for _, val := range ivrConfig.Dnis {
		quotedDnsSlice = append(quotedDnsSlice, strconv.Quote(val))
	}

	divisionId := ""
	if ivrConfig.DivisionId != "" {
		divisionId = ivrConfig.DivisionId
	} else {
		divisionId = "null"
	}

	return fmt.Sprintf(`resource "%s" "%s" {
		name        = "%s"
		description = "%s"
		dnis        = [%s]
		depends_on  = [%s]
		division_id = %s
	}
	`, ResourceType,
		ivrConfig.ResourceLabel,
		ivrConfig.Name,
		ivrConfig.Description,
		strings.Join(quotedDnsSlice, ","),
		ivrConfig.DependsOn,
		divisionId,
	)
}

// GenerateIvrDataSource generate an ivr data source as a string
func GenerateIvrDataSource(
	resourceLabel string,
	name string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, ResourceType, resourceLabel, name, dependsOnResource)
}

func buildArchitectIvrFromResourceData(d *schema.ResourceData) *platformclientv2.Ivr {
	ivrBody := platformclientv2.Ivr{
		Name:             platformclientv2.String(d.Get("name").(string)),
		OpenHoursFlow:    util.BuildSdkDomainEntityRef(d, "open_hours_flow_id"),
		ClosedHoursFlow:  util.BuildSdkDomainEntityRef(d, "closed_hours_flow_id"),
		HolidayHoursFlow: util.BuildSdkDomainEntityRef(d, "holiday_hours_flow_id"),
		ScheduleGroup:    util.BuildSdkDomainEntityRef(d, "schedule_group_id"),
		Dnis:             buildDnisFromResourceData(d),
	}

	if description := d.Get("description").(string); description != "" {
		ivrBody.Description = &description
	}

	if divisionId := d.Get("division_id").(string); divisionId != "" {
		ivrBody.Division = &platformclientv2.Writabledivision{Id: &divisionId}
	}

	return &ivrBody
}

// buildDnisFromResourceData reads the dnis list from the resource data. It returns nil when no numbers
// are configured so the request body is the same as it was when dnis was a set.
func buildDnisFromResourceData(d *schema.ResourceData) *[]string {
	dnisList, ok := d.Get("dnis").([]interface{})
	if !ok || len(dnisList) == 0 {
		return nil
	}
	dnis := lists.InterfaceListToStrings(dnisList)
	return &dnis
}

// orderDnisLikeCurrent returns the numbers from the API as a list that keeps the order of the numbers
// already in the resource data (config or prior state). Numbers that are not in the resource data yet are
// appended in sorted order. The API does not return dnis in a stable order, so without this the list would
// flip to the API order and produce a perpetual plan diff. Returns nil when there are no numbers.
func orderDnisLikeCurrent(d *schema.ResourceData, apiDnis []string) []interface{} {
	inApi := make(map[string]bool, len(apiDnis))
	for _, number := range apiDnis {
		inApi[number] = true
	}

	ordered := make([]interface{}, 0, len(apiDnis))
	seen := make(map[string]bool, len(apiDnis))

	if current, ok := d.Get("dnis").([]interface{}); ok {
		for _, v := range current {
			number, isString := v.(string)
			if isString && inApi[number] && !seen[number] {
				ordered = append(ordered, number)
				seen[number] = true
			}
		}
	}

	var remaining []string
	for _, number := range apiDnis {
		if !seen[number] {
			remaining = append(remaining, number)
			seen[number] = true
		}
	}
	sort.Strings(remaining)
	for _, number := range remaining {
		ordered = append(ordered, number)
	}

	if len(ordered) == 0 {
		return nil
	}
	return ordered
}
