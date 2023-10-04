package architect_ivr

import (
	"fmt"
	"strconv"
	"strings"
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
