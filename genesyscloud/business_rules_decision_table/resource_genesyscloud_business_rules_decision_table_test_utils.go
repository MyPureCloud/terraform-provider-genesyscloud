package business_rules_decision_table

import (
	"fmt"

	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v172/platformclientv2"
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
		inputs {
			defaults_to {
				special = "Wildcard"
			}
			expression {
				contractual {
					schema_property_key = "optional_string_empty_type_value"
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
		outputs {
			defaults_to {
				special = "Null"
			}
			value {
				schema_property_key = "optional_string_empty_block"
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

// generateRows generates rows using all literal types
func generateRows(queueResourceLabel string) string {
	return `rows {
		inputs {
			literal {
				value = "VIP"
				type  = "string"
			}
		}
		inputs {
			literal {
				value = "John Doe"
				type  = "string"
			}
		}
		inputs {
			literal {
				value = "5"
				type  = "integer"
			}
		}
		inputs {
			literal {
				value = "85.5"
				type  = "number"
			}
		}
		inputs {
			literal {
				value = "2023-01-15"
				type  = "date"
			}
		}
		inputs {
			literal {
				value = "2023-01-15T10:30:00.000Z"
				type  = "datetime"
			}
		}
		inputs {
			literal {
				value = "true"
				type  = "boolean"
			}
		}
		inputs {
			literal {
				value = ""
				type  = ""
			}
		}
		outputs {
			literal {
				value = genesyscloud_routing_queue.` + queueResourceLabel + `.id
				type  = "string"
			}
		}
		outputs {
			literal {
				value = "Premium Support"
				type  = "string"
			}
		}
		outputs {
			literal {}
		}
	}`
}

// generateRowsWithSpecials generates rows using special values for some types
func generateRowsWithSpecials(queueResourceLabel string) string {
	return `rows {
		inputs {
			literal {
				value = "Standard"
				type  = "string"
			}
		}
		inputs {
			literal {
				value = "Jane Smith"
				type  = "string"
			}
		}
		inputs {
			literal {
				value = "Wildcard"
				type  = "special"
			}
		}
		inputs {
			literal {
				value = "Wildcard"
				type  = "special"
			}
		}
		inputs {
			literal {
				value = "CurrentTime"
				type  = "special"
			}
		}
		inputs {
			literal {
				value = "CurrentTime"
				type  = "special"
			}
		}
		inputs {
			literal {
				value = "Null"
				type  = "special"
			}
		}
		inputs {
			literal {
				value = "Null"
				type  = "special"
			}
		}
		outputs {
			literal {
				value = genesyscloud_routing_queue.` + queueResourceLabel + `.id
				type  = "string"
			}
		}
		outputs {
			literal {
				value = "Standard Support"
				type  = "string"
			}
		}
		outputs {
			literal {
				value = "Null"
				type  = "special"
			}
		}
	}`
}

// generateUpdatedRows generates updated rows for schema testing
func generateUpdatedRows(queueResourceLabel string) string {
	return `rows {
		inputs {
			literal {
				value = "Premium"
				type  = "string"
			}
		}
		inputs {
			literal {
				value = "Alice Johnson"
				type  = "string"
			}
		}
		inputs {
			literal {
				value = "8"
				type  = "integer"
			}
		}
		inputs {
			literal {
				value = "92.3"
				type  = "number"
			}
		}
		inputs {
			literal {
				value = "2023-02-20"
				type  = "date"
			}
		}
		inputs {
			literal {
				value = "2023-02-20T14:45:30.000Z"
				type  = "datetime"
			}
		}
		inputs {
			literal {
				value = "false"
				type  = "boolean"
			}
		}
		inputs {
			literal {
				value = "this was defaulted to empty value and type first row"
				type  = "string"
			}
		}
		outputs {
			literal {
				value = genesyscloud_routing_queue.` + queueResourceLabel + `.id
				type  = "string"
			}
		}
		outputs {
			literal {
				value = "Updated Support"
				type  = "string"
			}
		}
		outputs {
			literal {
				value = "this was defaulted to empty block first row"
				type  = "string"
			}
		}
	}
	rows {
		inputs {
			literal {
				value = "VIP"
				type  = "string"
			}
		}
		inputs {
			literal {
				value = "Bob Wilson"
				type  = "string"
			}
		}
		inputs {
			literal {
				value = "3"
				type  = "integer"
			}
		}
		inputs {
			literal {
				value = "67.8"
				type  = "number"
			}
		}
		inputs {
			literal {
				value = "2023-03-10"
				type  = "date"
			}
		}
		inputs {
			literal {
				value = "2023-03-10T09:15:45.000Z"
				type  = "datetime"
			}
		}
		inputs {
			literal {
				value = "true"
				type  = "boolean"
			}
		}
		inputs {
			literal {
				value = "this was defaulted to empty value and type second row"
				type  = "string"
			}
		}
		outputs {
			literal {
				value = genesyscloud_routing_queue.` + queueResourceLabel + `.id
				type  = "string"
			}
		}
		outputs {
			literal {
				value = "Standard Support"
				type  = "string"
			}
		}
		outputs {
			literal {
				value = "this was defaulted to empty block second row"
				type  = "string"
			}
		}
	}`
}

// generateRowsWithInvalidLiteral generates rows with invalid literal for testing validation
func generateRowsWithInvalidLiteral(queueResourceLabel, literalType, invalidValue string) string {
	return `rows {
		inputs {
			literal {
				value = "VIP"
				type  = "string"
			}
		}
		inputs {
			literal {
				value = "` + invalidValue + `"
				type  = "` + literalType + `"
			}
		}
		outputs {
			literal {
				value = genesyscloud_routing_queue.` + queueResourceLabel + `.id
				type  = "string"
			}
		}
	}`
}

func generateBusinessRulesSchemaResource(resourceLabel, name, description string) string {
	return fmt.Sprintf(`resource "genesyscloud_business_rules_schema" "%s" {
		enabled = true
		name = "%s"
		description = "%s"
		properties = jsonencode({
			"customer_type" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/enum"
					}
				],
				"title" : "customer_type",
				"description" : "Customer type for routing decisions",
				"enum" : ["VIP", "Standard", "Premium"],
				"_enumProperties" : {
					"VIP" : {
						"title" : "VIP Customer"
					},
					"Standard" : {
						"title" : "Standard Customer"
					},
					"Premium" : {
						"title" : "Premium Customer"
					}
				}
			},
			"customer_name" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/string"
					}
				],
				"title" : "customer_name",
				"description" : "Customer name",
				"minLength" : 1,
				"maxLength" : 100
			},
			"priority_level" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/integer"
					}
				],
				"title" : "priority_level",
				"description" : "Priority level (1-10)",
				"minimum" : 1,
				"maximum" : 10
			},
			"score" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/number"
					}
				],
				"title" : "score",
				"description" : "Customer score",
				"minimum" : 0.0,
				"maximum" : 100.0
			},
			"created_date" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/date"
					}
				],
				"title" : "created_date",
				"description" : "Customer creation date"
			},
			"last_updated" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/datetime"
					}
				],
				"title" : "last_updated",
				"description" : "Last update timestamp"
			},
			"is_active" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/boolean"
					}
				],
				"title" : "is_active",
				"description" : "Whether customer is active"
			},
			"transfer_queue" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/businessRulesQueue"
					}
				],
				"title" : "transfer_queue",
				"description" : "Transfer queue for routing"
			},
			"skill" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/string"
					}
				],
				"title" : "skill",
				"description" : "Skill for routing",
				"minLength" : 1,
				"maxLength" : 100
			},
			"optional_string_empty_block" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/string"
					}
				],
				"title" : "optional_string_empty_block",
				"description" : "Used to test optional empty literal block in rows",
				"minLength" : 1,
				"maxLength" : 100
			},
			"optional_string_empty_type_value" = {
				"allOf" : [
					{
						"$ref" : "#/definitions/string"
					}
				],
				"title" : "optional_string_empty_type_value",
				"description" : "Used to test literal block in rows with empty type and value",
				"minLength" : 1,
				"maxLength" : 100
			}
		})
	}
	`, resourceLabel, name, description)
}
