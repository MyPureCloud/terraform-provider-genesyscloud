package architect_ivr

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

type IvrConfigStruct struct {
	ResourceID  string
	Name        string
	Description string
	Dnis        []string
	DependsOn   string
	DivisionId  string
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
	`, resourceName,
		ivrConfig.ResourceID,
		ivrConfig.Name,
		ivrConfig.Description,
		strings.Join(quotedDnsSlice, ","),
		ivrConfig.DependsOn,
		divisionId,
	)
}

// GenerateIvrDataSource generate an ivr data source as a string
func GenerateIvrDataSource(
	resourceID string,
	name string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceName, resourceID, name, dependsOnResource)
}

func buildArchitectIvrFromResourceData(d *schema.ResourceData) *platformclientv2.Ivr {
	ivrBody := platformclientv2.Ivr{
		Name:             platformclientv2.String(d.Get("name").(string)),
		OpenHoursFlow:    util.BuildSdkDomainEntityRef(d, "open_hours_flow_id"),
		ClosedHoursFlow:  util.BuildSdkDomainEntityRef(d, "closed_hours_flow_id"),
		HolidayHoursFlow: util.BuildSdkDomainEntityRef(d, "holiday_hours_flow_id"),
		ScheduleGroup:    util.BuildSdkDomainEntityRef(d, "schedule_group_id"),
		Dnis:             lists.BuildSdkStringList(d, "dnis"),
	}

	if description := d.Get("description").(string); description != "" {
		ivrBody.Description = &description
	}

	if divisionId := d.Get("division_id").(string); divisionId != "" {
		ivrBody.Division = &platformclientv2.Writabledivision{Id: &divisionId}
	}

	return &ivrBody
}
