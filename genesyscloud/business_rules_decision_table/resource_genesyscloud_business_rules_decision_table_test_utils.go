package business_rules_decision_table

import (
	"fmt"

	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// generateBusinessRulesDecisionTableResource generates a basic business rules decision table resource
func generateBusinessRulesDecisionTableResource(
	resourceLabel string,
	name string,
	description string,
	divisionId string,
	schemaId string,
	columns string) string {
	return fmt.Sprintf(`resource "genesyscloud_business_rules_decision_table" "%s" {
		name = "%s"
		description = "%s"
		division_id = %s
		schema_id = %s
		%s
	}
	`, resourceLabel, name, description, divisionId, schemaId, columns)
}

// generateBusinessRulesDecisionTableResourceWithQueues generates a decision table resource with complex columns and queue references
func generateBusinessRulesDecisionTableResourceWithQueues(
	resourceLabel string,
	name string,
	description string,
	divisionId string,
	schemaId string,
	queueResourceLabel string) string {
	return generateBusinessRulesDecisionTableResource(
		resourceLabel,
		name,
		description,
		divisionId,
		schemaId,
		generateComplexColumnsWithQueues(queueResourceLabel))
}

// generateComplexColumnsWithQueues generates complex columns with real queue references
func generateComplexColumnsWithQueues(queueResourceLabel string) string {
	return fmt.Sprintf(`columns {
		inputs {
			defaults_to {
				special = "Wildcard"
			}
			expression {
				contractual {
					schema_property_key = "customer_type"
				}
				comparator = "Equals"
			}
		}
		inputs {
			defaults_to {
				special = "Wildcard"
			}
			expression {
				contractual {
					schema_property_key = "priority"
				}
				comparator = "Equals"
			}
		}
		outputs {
			defaults_to {
				value = genesyscloud_routing_queue.%s.id
			}
			value {
				schema_property_key = "transfer_queue"
				properties {
					schema_property_key = "queue"
					properties {
						schema_property_key = "id"
					}
				}
			}
		}
		outputs {
			defaults_to {
				special = "Null"
			}
			value {
				schema_property_key = "skill"
			}
		}
	}`, queueResourceLabel)
}

// generateRoutingQueueResource generates a routing queue resource for testing
func generateRoutingQueueResource(resourceLabel, name string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
		name = "%s"
		division_id = data.genesyscloud_auth_division_home.home.id
		description = "Test queue for decision table testing"
	}
	`, resourceLabel, name)
}

func businessRulesDecisionTableFtIsEnabled() (bool, *platformclientv2.APIResponse, *platformclientv2.APIResponse) {
	log.Printf("DEBUG: Checking if business rules decision tables is enabled")
	clientConfig := platformclientv2.GetDefaultConfiguration()
	businessRulesApi := platformclientv2.NewBusinessRulesApiWithConfig(clientConfig)
	queueApi := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	// to create a decision table queue view permission is required so lets ensure we can get queues by calling get all queues
	_, decisionTableResp, businessRulesDecisionTableErr := businessRulesApi.GetBusinessrulesDecisiontables("", "", nil, "")
	_, queueResp, queueErr := queueApi.GetRoutingQueues(1, 100, "", "", nil, nil, nil, "", false, nil)
	if businessRulesDecisionTableErr != nil || queueErr != nil {
		if businessRulesDecisionTableErr != nil {
			log.Printf("Error getting business rules decision tables: %v", businessRulesDecisionTableErr)
		}
		if queueErr != nil {
			log.Printf("Error getting routing queues: %v", queueErr)
		}
		return false, decisionTableResp, queueResp
	}

	return decisionTableResp.StatusCode == 200 && queueResp.StatusCode == 200, decisionTableResp, queueResp
}
