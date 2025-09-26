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
	columns string,
	rows string) string {
	return fmt.Sprintf(`resource "genesyscloud_business_rules_decision_table" "%s" {
		name = "%s"
		description = "%s"
		division_id = %s
		schema_id = %s
		%s
		%s
	}
	`, resourceLabel, name, description, divisionId, schemaId, columns, rows)
}

func generateColumns(queueResourceLabel string) string {
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
					schema_property_key = "customer_name"
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
					schema_property_key = "priority_level"
				}
				comparator = "GreaterThan"
			}
		}
		inputs {
			defaults_to {
				special = "Wildcard"
			}
			expression {
				contractual {
					schema_property_key = "score"
				}
				comparator = "GreaterThan"
			}
		}
		inputs {
			defaults_to {
				special = "Wildcard"
			}
			expression {
				contractual {
					schema_property_key = "created_date"
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
					schema_property_key = "last_updated"
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
					schema_property_key = "is_active"
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

// generateBasicRows generates basic rows for the original test (simpler than the comprehensive test)
func generateBasicRows(queueResourceLabel string) string {
	return `rows {
		inputs {
			schema_property_key = "customer_type"
			comparator = "Equals"
			literal {
				value = "VIP"
				type  = "string"
			}
		}
		inputs {
			schema_property_key = "priority_level"
			comparator = "GreaterThan"
			literal {
				value = "5"
				type  = "integer"
			}
		}
		outputs {
			schema_property_key = "transfer_queue"
			literal {
				value = genesyscloud_routing_queue.` + queueResourceLabel + `.id
				type  = "string"
			}
		}
		outputs {
			schema_property_key = "skill"
			literal {
				value = "VIP Support"
				type  = "string"
			}
		}
	}`
}

// generateRows generates rows using all literal types
func generateRows(queueResourceLabel string) string {
	return `rows {
		inputs {
			schema_property_key = "customer_type"
			comparator = "Equals"
			literal {
				value = "VIP"
				type  = "string"
			}
		}
		inputs {
			schema_property_key = "customer_name"
			comparator = "Equals"
			literal {
				value = "John Doe"
				type  = "string"
			}
		}
		inputs {
			schema_property_key = "priority_level"
			comparator = "GreaterThan"
			literal {
				value = "5"
				type  = "integer"
			}
		}
		inputs {
			schema_property_key = "score"
			comparator = "GreaterThan"
			literal {
				value = "85.5"
				type  = "number"
			}
		}
		inputs {
			schema_property_key = "created_date"
			comparator = "Equals"
			literal {
				value = "2023-01-15"
				type  = "date"
			}
		}
		inputs {
			schema_property_key = "last_updated"
			comparator = "Equals"
			literal {
				value = "2023-01-15T10:30:00.000Z"
				type  = "datetime"
			}
		}
		inputs {
			schema_property_key = "is_active"
			comparator = "Equals"
			literal {
				value = "true"
				type  = "boolean"
			}
		}
		outputs {
			schema_property_key = "transfer_queue"
			literal {
				value = genesyscloud_routing_queue.` + queueResourceLabel + `.id
				type  = "string"
			}
		}
		outputs {
			schema_property_key = "skill"
			literal {
				value = "Premium Support"
				type  = "string"
			}
		}
	}`
}

// generateRowsWithSpecials generates rows using special values for some types
func generateRowsWithSpecials(queueResourceLabel string) string {
	return `rows {
		inputs {
			schema_property_key = "customer_type"
			comparator = "Equals"
			literal {
				value = "Standard"
				type  = "string"
			}
		}
		inputs {
			schema_property_key = "customer_name"
			comparator = "Equals"
			literal {
				value = "Jane Smith"
				type  = "string"
			}
		}
		inputs {
			schema_property_key = "priority_level"
			comparator = "GreaterThan"
			literal {
				value = "Wildcard"
				type  = "special"
			}
		}
		inputs {
			schema_property_key = "score"
			comparator = "GreaterThan"
			literal {
				value = "Wildcard"
				type  = "special"
			}
		}
		inputs {
			schema_property_key = "created_date"
			comparator = "Equals"
			literal {
				value = "CurrentTime"
				type  = "special"
			}
		}
		inputs {
			schema_property_key = "last_updated"
			comparator = "Equals"
			literal {
				value = "CurrentTime"
				type  = "special"
			}
		}
		inputs {
			schema_property_key = "is_active"
			comparator = "Equals"
			literal {
				value = "Null"
				type  = "special"
			}
		}
		outputs {
			schema_property_key = "transfer_queue"
			literal {
				value = genesyscloud_routing_queue.` + queueResourceLabel + `.id
				type  = "string"
			}
		}
		outputs {
			schema_property_key = "skill"
			literal {
				value = "Standard Support"
				type  = "string"
			}
		}
	}`
}

// generateUpdatedRows generates updated rows for schema testing
func generateUpdatedRows(queueResourceLabel string) string {
	return `rows {
		inputs {
			schema_property_key = "customer_type"
			comparator = "Equals"
			literal {
				value = "Premium"
				type  = "string"
			}
		}
		inputs {
			schema_property_key = "customer_name"
			comparator = "Equals"
			literal {
				value = "Alice Johnson"
				type  = "string"
			}
		}
		inputs {
			schema_property_key = "priority_level"
			comparator = "GreaterThan"
			literal {
				value = "8"
				type  = "integer"
			}
		}
		inputs {
			schema_property_key = "score"
			comparator = "GreaterThan"
			literal {
				value = "92.3"
				type  = "number"
			}
		}
		inputs {
			schema_property_key = "created_date"
			comparator = "Equals"
			literal {
				value = "2023-02-20"
				type  = "date"
			}
		}
		inputs {
			schema_property_key = "last_updated"
			comparator = "Equals"
			literal {
				value = "2023-02-20T14:45:30.000Z"
				type  = "datetime"
			}
		}
		inputs {
			schema_property_key = "is_active"
			comparator = "Equals"
			literal {
				value = "false"
				type  = "boolean"
			}
		}
		outputs {
			schema_property_key = "transfer_queue"
			literal {
				value = genesyscloud_routing_queue.` + queueResourceLabel + `.id
				type  = "string"
			}
		}
		outputs {
			schema_property_key = "skill"
			literal {
				value = "Updated Support"
				type  = "string"
			}
		}
	}
	rows {
		inputs {
			schema_property_key = "customer_type"
			comparator = "Equals"
			literal {
				value = "VIP"
				type  = "string"
			}
		}
		inputs {
			schema_property_key = "customer_name"
			comparator = "Equals"
			literal {
				value = "Bob Wilson"
				type  = "string"
			}
		}
		inputs {
			schema_property_key = "priority_level"
			comparator = "GreaterThan"
			literal {
				value = "3"
				type  = "integer"
			}
		}
		inputs {
			schema_property_key = "score"
			comparator = "GreaterThan"
			literal {
				value = "67.8"
				type  = "number"
			}
		}
		inputs {
			schema_property_key = "created_date"
			comparator = "Equals"
			literal {
				value = "2023-03-10"
				type  = "date"
			}
		}
		inputs {
			schema_property_key = "last_updated"
			comparator = "Equals"
			literal {
				value = "2023-03-10T09:15:45.000Z"
				type  = "datetime"
			}
		}
		inputs {
			schema_property_key = "is_active"
			comparator = "Equals"
			literal {
				value = "true"
				type  = "boolean"
			}
		}
		outputs {
			schema_property_key = "transfer_queue"
			literal {
				value = genesyscloud_routing_queue.` + queueResourceLabel + `.id
				type  = "string"
			}
		}
		outputs {
			schema_property_key = "skill"
			literal {
				value = "Standard Support"
				type  = "string"
			}
		}
	}`
}

// generateRowsWithInvalidLiteral generates rows with invalid literal for testing validation
func generateRowsWithInvalidLiteral(queueResourceLabel, literalType, invalidValue string) string {
	return `rows {
		inputs {
			schema_property_key = "customer_type"
			comparator = "Equals"
			literal {
				value = "VIP"
				type  = "string"
			}
		}
		inputs {
			schema_property_key = "priority_level"
			comparator = "GreaterThan"
			literal {
				value = "` + invalidValue + `"
				type  = "` + literalType + `"
			}
		}
		outputs {
			schema_property_key = "transfer_queue"
			literal {
				value = genesyscloud_routing_queue.` + queueResourceLabel + `.id
				type  = "string"
			}
		}
	}`
}

// generateColumnsWithDefaults generates columns with defaults for testing default behavior
func generateColumnsWithDefaults(queueResourceLabel string) string {
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
					schema_property_key = "priority_level"
				}
				comparator = "GreaterThan"
			}
		}
		inputs {
			defaults_to {
				special = "Wildcard"
			}
			expression {
				contractual {
					schema_property_key = "score"
				}
				comparator = "GreaterThan"
			}
		}
		outputs {
			defaults_to {
				value = "Premium Support"
			}
			value {
				schema_property_key = "skill"
			}
		}
	}`)
}

// generateRowsWithDefaults generates rows that only specify some inputs/outputs, letting others use defaults
func generateRowsWithDefaults(queueResourceLabel string) string {
	return `rows {
		inputs {
			schema_property_key = "customer_type"
			comparator = "Equals"
			literal {
				value = "VIP"
				type  = "string"
			}
		}
		inputs {
			schema_property_key = "priority_level"
			comparator = "GreaterThan"
			literal {
				value = "5"
				type  = "integer"
			}
		}
		// Note: score input is omitted - will use default (Wildcard)
		outputs {
			schema_property_key = "skill"
			literal {
				value = "Premium Support"
				type  = "string"
			}
		}
		// Note: skill output is omitted - will use default (Premium Support)
	}`
}
