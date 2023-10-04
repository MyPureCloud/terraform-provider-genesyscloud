package architect_ivr

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v112/platformclientv2"
	"strconv"
	"strings"
	"time"
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

// TODO : Remove if tests pass

// DeleteIvrStartingWithPrefix finds an ivr that has prefix in its name and deletes it
func DeleteIvrStartingWithPrefix(ctx context.Context, prefix string) error {
	config := platformclientv2.GetDefaultConfiguration()
	proxy := getArchitectIvrProxy(config)

	allIvrs, err := proxy.getAllArchitectIvrs(ctx, "")
	if err != nil {
		return fmt.Errorf("error reading architect ivrs: %v", err)
	}

	for _, ivr := range *allIvrs {
		if strings.HasPrefix(*ivr.Name, prefix) {
			_, err := proxy.deleteArchitectIvr(ctx, *ivr.Id)
			if err != nil {
				return fmt.Errorf("error deleting architect ivr %s: %v", *ivr.Id, err)
			}
			time.Sleep(5 * time.Second)
		}
	}
	return nil
}
