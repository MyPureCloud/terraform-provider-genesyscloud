package outbound

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	outboundContactListFilterClauseResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`filter_type`: {
				Description:  `How to join predicates together.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`AND`, `OR`}, false),
			},
			`predicates`: {
				Description: `Conditions to filter the contacts by.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        outboundContactListFilterPredicateResource,
			},
		},
	}
	outboundContactListFilterPredicateResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`column`: {
				Description: `Contact list column from the contact list filter's contact list.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`column_type`: {
				Description:  `The type of data in the contact column.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`numeric`, `alphabetic`}, false),
			},
			`operator`: {
				Description:  `The operator for this contact list filter predicate.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`EQUALS`, `LESS_THAN`, `LESS_THAN_EQUALS`, `GREATER_THAN`, `GREATER_THAN_EQUALS`, `CONTAINS`, `BEGINS_WITH`, `ENDS_WITH`, `BEFORE`, `AFTER`, `BETWEEN`, `IN`}, false),
			},
			`value`: {
				Description: `Value with which to compare the contact's data. This could be text, a number, or a relative time. A value for relative time should follow the format PxxDTyyHzzM, where xx, yy, and zz specify the days, hours and minutes. For example, a value of P01DT08H30M corresponds to 1 day, 8 hours, and 30 minutes from now. To specify a time in the past, include a negative sign before each numeric value. For example, a value of P-01DT-08H-30M corresponds to 1 day, 8 hours, and 30 minutes in the past. You can also do things like P01DT00H-30M, which would correspond to 23 hours and 30 minutes from now (1 day - 30 minutes).`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`var_range`: {
				Description: `A range of values. Required for operators BETWEEN and IN.`,
				Optional:    true,
				MaxItems:    1,
				Type:        schema.TypeSet,
				Elem:        outboundContactListFilterRangeResource,
			},
			`inverted`: {
				Description: `Inverts the result of the predicate (i.e., if the predicate returns true, inverting it will return false).`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeBool,
			},
		},
	}
	outboundContactListFilterRangeResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`min`: {
				Description: `The minimum value of the range. Required for the operator BETWEEN.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`max`: {
				Description: `The maximum value of the range. Required for the operator BETWEEN.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`min_inclusive`: {
				Description: `Whether or not to include the minimum in the range.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`max_inclusive`: {
				Description: `Whether or not to include the maximum in the range.`,
				Optional:    true,
				Type:        schema.TypeBool,
			},
			`in_set`: {
				Description: `A set of values that the contact data should be in. Required for the IN operator.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
)

func getAllOutboundContactListFilters(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		contactListFilterConfigs, _, getErr := outboundAPI.GetOutboundContactlistfilters(pageSize, pageNum, true, "", "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of contact list filter configs: %v", getErr)
		}

		if contactListFilterConfigs.Entities == nil || len(*contactListFilterConfigs.Entities) == 0 {
			break
		}

		for _, contactListFilterConfig := range *contactListFilterConfigs.Entities {
			resources[*contactListFilterConfig.Id] = &resourceExporter.ResourceMeta{Name: *contactListFilterConfig.Name}
		}
	}

	return resources, nil
}

func OutboundContactListFilterExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: gcloud.GetAllWithPooledClient(getAllOutboundContactListFilters),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"contact_list_id": {RefType: "genesyscloud_outbound_contact_list"},
		},
	}
}

func ResourceOutboundContactListFilter() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound Contact List Filter`,

		CreateContext: gcloud.CreateWithPooledClient(createOutboundContactListFilter),
		ReadContext:   gcloud.ReadWithPooledClient(readOutboundContactListFilter),
		UpdateContext: gcloud.UpdateWithPooledClient(updateOutboundContactListFilter),
		DeleteContext: gcloud.DeleteWithPooledClient(deleteOutboundContactListFilter),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the list.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`contact_list_id`: {
				Description: `The contact list the filter is based on.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`clauses`: {
				Description: `Groups of conditions to filter the contacts by.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem:        outboundContactListFilterClauseResource,
			},
			`filter_type`: {
				Description:  `How to join clauses together.`,
				Optional:     true,
				Computed:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`AND`, `OR`}, false),
			},
		},
	}
}

func createOutboundContactListFilter(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	filterType := d.Get("filter_type").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkContactListFilter := platformclientv2.Contactlistfilter{
		ContactList: gcloud.BuildSdkDomainEntityRef(d, "contact_list_id"),
		Clauses:     buildSdkOutboundContactListFilterClauseSlice(d.Get("clauses").([]interface{})),
	}

	if name != "" {
		sdkContactListFilter.Name = &name
	}
	if filterType != "" {
		sdkContactListFilter.FilterType = &filterType
	}

	log.Printf("Creating Outbound Contact List Filter %s", name)
	outboundContactListFilter, _, err := outboundApi.PostOutboundContactlistfilters(sdkContactListFilter)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Contact List Filter %s: %s", name, err)
	}

	d.SetId(*outboundContactListFilter.Id)

	log.Printf("Created Outbound Contact List Filter %s %s", name, *outboundContactListFilter.Id)
	return readOutboundContactListFilter(ctx, d, meta)
}

func updateOutboundContactListFilter(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	filterType := d.Get("filter_type").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkContactListFilter := platformclientv2.Contactlistfilter{
		ContactList: gcloud.BuildSdkDomainEntityRef(d, "contact_list_id"),
		Clauses:     buildSdkOutboundContactListFilterClauseSlice(d.Get("clauses").([]interface{})),
	}

	if name != "" {
		sdkContactListFilter.Name = &name
	}
	if filterType != "" {
		sdkContactListFilter.FilterType = &filterType
	}

	log.Printf("Updating Outbound Contact List Filter %s", name)
	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Contact list filter version
		outboundContactListFilter, resp, getErr := outboundApi.GetOutboundContactlistfilter(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Contact List Filter %s: %s", d.Id(), getErr)
		}
		sdkContactListFilter.Version = outboundContactListFilter.Version
		outboundContactListFilter, _, updateErr := outboundApi.PutOutboundContactlistfilter(d.Id(), sdkContactListFilter)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Outbound Contact List Filter %s: %s", name, updateErr)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Contact List Filter %s", name)
	return readOutboundContactListFilter(ctx, d, meta)
}

func readOutboundContactListFilter(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound Contact List Filter %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkContactListFilter, resp, getErr := outboundApi.GetOutboundContactlistfilter(d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read Outbound Contact List Filter %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read Outbound Contact List Filter %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundContactListFilter())
		if sdkContactListFilter.Name != nil {
			_ = d.Set("name", *sdkContactListFilter.Name)
		}
		if sdkContactListFilter.ContactList != nil && sdkContactListFilter.ContactList.Id != nil {
			_ = d.Set("contact_list_id", *sdkContactListFilter.ContactList.Id)
		}
		if sdkContactListFilter.Clauses != nil {
			_ = d.Set("clauses", flattenSdkOutboundContactListFilterClauseSlice(*sdkContactListFilter.Clauses))
		}
		if sdkContactListFilter.FilterType != nil {
			_ = d.Set("filter_type", *sdkContactListFilter.FilterType)
		}

		log.Printf("Read Outbound Contact List Filter %s %s", d.Id(), *sdkContactListFilter.Name)
		return cc.CheckState()
	})
}

func deleteOutboundContactListFilter(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Contact List Filter")
		resp, err := outboundApi.DeleteOutboundContactlistfilter(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Contact List Filter: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := outboundApi.GetOutboundContactlistfilter(d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Outbound Contact list filter deleted
				log.Printf("Deleted Outbound Contact List Filter %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting Outbound Contact List Filter %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Outbound Contact List Filter %s still exists", d.Id()))
	})
}

func buildSdkOutboundContactListFilterRange(contactListFilterRange *schema.Set) *platformclientv2.Contactlistfilterrange {
	var sdkContactListFilterRange platformclientv2.Contactlistfilterrange
	contactListFilterRangeList := contactListFilterRange.List()
	contactListFilterRangeMap := contactListFilterRangeList[0].(map[string]interface{})
	if min := contactListFilterRangeMap["min"].(string); min != "" {
		sdkContactListFilterRange.Min = &min
	}
	if max := contactListFilterRangeMap["max"].(string); max != "" {
		sdkContactListFilterRange.Max = &max
	}
	sdkContactListFilterRange.MinInclusive = platformclientv2.Bool(contactListFilterRangeMap["min_inclusive"].(bool))
	sdkContactListFilterRange.MaxInclusive = platformclientv2.Bool(contactListFilterRangeMap["max_inclusive"].(bool))
	inSet := make([]string, 0)
	for _, v := range contactListFilterRangeMap["in_set"].([]interface{}) {
		inSet = append(inSet, v.(string))
	}
	sdkContactListFilterRange.InSet = &inSet
	return &sdkContactListFilterRange
}

func buildSdkOutboundContactListFilterPredicateSlice(contactListFilterPredicate []interface{}) *[]platformclientv2.Contactlistfilterpredicate {
	if contactListFilterPredicate == nil || len(contactListFilterPredicate) == 0 {
		return nil
	}
	sdkContactListFilterPredicateSlice := make([]platformclientv2.Contactlistfilterpredicate, 0)
	for _, configContactListFilterPredicate := range contactListFilterPredicate {
		if contactListFilterPredicateMap, ok := configContactListFilterPredicate.(map[string]interface{}); ok {
			var sdkContactListFilterPredicate platformclientv2.Contactlistfilterpredicate
			if column := contactListFilterPredicateMap["column"].(string); column != "" {
				sdkContactListFilterPredicate.Column = &column
			}
			if columnType := contactListFilterPredicateMap["column_type"].(string); columnType != "" {
				sdkContactListFilterPredicate.ColumnType = &columnType
			}
			if operator := contactListFilterPredicateMap["operator"].(string); operator != "" {
				sdkContactListFilterPredicate.Operator = &operator
			}
			if value := contactListFilterPredicateMap["value"].(string); value != "" {
				sdkContactListFilterPredicate.Value = &value
			}
			if varRangeSet := contactListFilterPredicateMap["var_range"].(*schema.Set); varRangeSet != nil && len(varRangeSet.List()) > 0 {
				sdkContactListFilterPredicate.VarRange = buildSdkOutboundContactListFilterRange(varRangeSet)
			}
			sdkContactListFilterPredicate.Inverted = platformclientv2.Bool(contactListFilterPredicateMap["inverted"].(bool))
			sdkContactListFilterPredicateSlice = append(sdkContactListFilterPredicateSlice, sdkContactListFilterPredicate)
		}
	}
	return &sdkContactListFilterPredicateSlice
}

func buildSdkOutboundContactListFilterClauseSlice(contactListFilterClause []interface{}) *[]platformclientv2.Contactlistfilterclause {
	if contactListFilterClause == nil || len(contactListFilterClause) == 0 {
		return nil
	}
	sdkContactListFilterClauseSlice := make([]platformclientv2.Contactlistfilterclause, 0)
	for _, configContactListFilterClause := range contactListFilterClause {
		var sdkContactListFilterClause platformclientv2.Contactlistfilterclause
		contactListFilterClauseMap := configContactListFilterClause.(map[string]interface{})
		if filterType := contactListFilterClauseMap["filter_type"].(string); filterType != "" {
			sdkContactListFilterClause.FilterType = &filterType
		}
		if predicates := contactListFilterClauseMap["predicates"]; predicates != nil {
			sdkContactListFilterClause.Predicates = buildSdkOutboundContactListFilterPredicateSlice(predicates.([]interface{}))
		}

		sdkContactListFilterClauseSlice = append(sdkContactListFilterClauseSlice, sdkContactListFilterClause)
	}
	return &sdkContactListFilterClauseSlice
}

func flattenSdkOutboundContactListFilterRange(contactListFilterRange *platformclientv2.Contactlistfilterrange) *schema.Set {
	if contactListFilterRange == nil {
		return nil
	}

	contactListFilterRangeSet := schema.NewSet(schema.HashResource(outboundContactListFilterRangeResource), []interface{}{})
	contactListFilterRangeMap := make(map[string]interface{})

	if contactListFilterRange.Min != nil {
		contactListFilterRangeMap["min"] = *contactListFilterRange.Min
	}
	if contactListFilterRange.Max != nil {
		contactListFilterRangeMap["max"] = *contactListFilterRange.Max
	}
	if contactListFilterRange.MinInclusive != nil {
		contactListFilterRangeMap["min_inclusive"] = *contactListFilterRange.MinInclusive
	}
	if contactListFilterRange.MaxInclusive != nil {
		contactListFilterRangeMap["max_inclusive"] = *contactListFilterRange.MaxInclusive
	}
	if contactListFilterRange.InSet != nil {
		// Changed []string to []interface{} to prevent type conversion panic
		inSet := make([]interface{}, 0)
		for _, v := range *contactListFilterRange.InSet {
			inSet = append(inSet, v)
		}
		contactListFilterRangeMap["in_set"] = inSet
	}
	if len(contactListFilterRangeMap) == 0 {
		return nil
	}
	contactListFilterRangeSet.Add(contactListFilterRangeMap)

	return contactListFilterRangeSet
}

func flattenSdkOutboundContactListFilterPredicateSlice(contactListFilterPredicates []platformclientv2.Contactlistfilterpredicate) []interface{} {
	if len(contactListFilterPredicates) == 0 {
		return nil
	}

	contactListFilterPredicateList := make([]interface{}, 0)
	for _, contactListFilterPredicate := range contactListFilterPredicates {
		contactListFilterPredicateMap := make(map[string]interface{})

		if contactListFilterPredicate.Column != nil {
			contactListFilterPredicateMap["column"] = *contactListFilterPredicate.Column
		}
		if contactListFilterPredicate.ColumnType != nil {
			contactListFilterPredicateMap["column_type"] = *contactListFilterPredicate.ColumnType
		}
		if contactListFilterPredicate.Operator != nil {
			contactListFilterPredicateMap["operator"] = *contactListFilterPredicate.Operator
		}
		if contactListFilterPredicate.Value != nil {
			contactListFilterPredicateMap["value"] = *contactListFilterPredicate.Value
		}
		if contactListFilterPredicate.VarRange != nil {
			contactListFilterPredicateMap["var_range"] = flattenSdkOutboundContactListFilterRange(contactListFilterPredicate.VarRange)
		}
		if contactListFilterPredicate.Inverted != nil {
			contactListFilterPredicateMap["inverted"] = *contactListFilterPredicate.Inverted
		}
		contactListFilterPredicateList = append(contactListFilterPredicateList, contactListFilterPredicateMap)
	}
	return contactListFilterPredicateList
}

func flattenSdkOutboundContactListFilterClauseSlice(contactListFilterClauses []platformclientv2.Contactlistfilterclause) []interface{} {
	if len(contactListFilterClauses) == 0 {
		return nil
	}

	contactListFilterClauseList := make([]interface{}, 0)
	for _, contactListFilterClause := range contactListFilterClauses {
		contactListFilterClauseMap := make(map[string]interface{})

		if contactListFilterClause.FilterType != nil {
			contactListFilterClauseMap["filter_type"] = *contactListFilterClause.FilterType
		}
		if contactListFilterClause.Predicates != nil {
			contactListFilterClauseMap["predicates"] = flattenSdkOutboundContactListFilterPredicateSlice(*contactListFilterClause.Predicates)
		}
		contactListFilterClauseList = append(contactListFilterClauseList, contactListFilterClauseMap)
	}
	return contactListFilterClauseList
}
