package customer_intent

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

// getCustomerIntentFromResourceData extracts a Customerintentresponse from resource data
func getCustomerIntentFromResourceData(d *schema.ResourceData) platformclientv2.Customerintentresponse {
	categoryId := d.Get("category_id").(string)
	return platformclientv2.Customerintentresponse{
		Name:        platformclientv2.String(d.Get("name").(string)),
		Description: platformclientv2.String(d.Get("description").(string)),
		ExpiryTime:  platformclientv2.Int(d.Get("expiry_time").(int)),
		Category: &platformclientv2.Addressableentityref{
			Id: &categoryId,
		},
	}
}

// getSourceIntentsFromResourceData extracts source intents from resource data
func getSourceIntentsFromResourceData(d *schema.ResourceData) []platformclientv2.Sourceintent {
	if sourceIntentsSet, ok := d.GetOk("source_intents"); ok {
		return buildSourceIntentsFromSet(sourceIntentsSet.(*schema.Set))
	}
	return []platformclientv2.Sourceintent{}
}

// buildSourceIntentsFromSet converts a schema.Set to a slice of Sourceintent
func buildSourceIntentsFromSet(sourceIntentsSet *schema.Set) []platformclientv2.Sourceintent {
	sourceIntents := make([]platformclientv2.Sourceintent, 0)
	for _, item := range sourceIntentsSet.List() {
		sourceIntentMap := item.(map[string]interface{})
		sourceIntent := platformclientv2.Sourceintent{}

		if sourceIntentId, ok := sourceIntentMap["source_intent_id"].(string); ok && sourceIntentId != "" {
			sourceIntent.SourceIntentId = &sourceIntentId
		}
		if sourceIntentName, ok := sourceIntentMap["source_intent_name"].(string); ok && sourceIntentName != "" {
			sourceIntent.SourceIntentName = &sourceIntentName
		}
		if sourceType, ok := sourceIntentMap["source_type"].(string); ok && sourceType != "" {
			sourceIntent.SourceType = &sourceType
		}
		if sourceId, ok := sourceIntentMap["source_id"].(string); ok && sourceId != "" {
			sourceIntent.SourceId = &sourceId
		}
		if sourceName, ok := sourceIntentMap["source_name"].(string); ok && sourceName != "" {
			sourceIntent.SourceName = &sourceName
		}

		sourceIntents = append(sourceIntents, sourceIntent)
	}
	return sourceIntents
}

// flattenSourceIntents converts a slice of Customersourceintent to a schema.Set
func flattenSourceIntents(sourceIntents []platformclientv2.Customersourceintent) *schema.Set {
	sourceIntentSet := schema.NewSet(schema.HashResource(&schema.Resource{
		Schema: map[string]*schema.Schema{
			"source_intent_id": {
				Type: schema.TypeString,
			},
			"source_intent_name": {
				Type: schema.TypeString,
			},
			"source_type": {
				Type: schema.TypeString,
			},
			"source_id": {
				Type: schema.TypeString,
			},
			"source_name": {
				Type: schema.TypeString,
			},
		},
	}), []interface{}{})

	for _, customerSourceIntent := range sourceIntents {
		if customerSourceIntent.SourceIntent == nil {
			continue
		}
		sourceIntent := customerSourceIntent.SourceIntent
		sourceIntentMap := make(map[string]interface{})

		if sourceIntent.SourceIntentId != nil {
			sourceIntentMap["source_intent_id"] = *sourceIntent.SourceIntentId
		}
		if sourceIntent.SourceIntentName != nil {
			sourceIntentMap["source_intent_name"] = *sourceIntent.SourceIntentName
		}
		if sourceIntent.SourceType != nil {
			sourceIntentMap["source_type"] = *sourceIntent.SourceType
		}
		if sourceIntent.SourceId != nil {
			sourceIntentMap["source_id"] = *sourceIntent.SourceId
		}
		if sourceIntent.SourceName != nil {
			sourceIntentMap["source_name"] = *sourceIntent.SourceName
		}

		sourceIntentSet.Add(sourceIntentMap)
	}

	return sourceIntentSet
}
