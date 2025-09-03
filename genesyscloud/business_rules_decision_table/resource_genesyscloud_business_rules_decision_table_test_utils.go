package business_rules_decision_table

import (
	"fmt"
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
