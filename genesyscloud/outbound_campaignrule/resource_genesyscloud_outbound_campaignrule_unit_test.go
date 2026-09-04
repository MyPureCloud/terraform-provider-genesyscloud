package outbound_campaignrule

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceOutboundCampaignruleCreate(t *testing.T) {
	tId := uuid.NewString()
	tName := "campaign rule name"
	testCampaignRule := generateCampaignRuleData(tId, tName)

	campaignRuleProxy := &outboundCampaignruleProxy{}
	campaignRuleProxy.getOutboundCampaignruleByIdAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.Campaignrule, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		campaignRule := &testCampaignRule

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return campaignRule, apiResponse, nil
	}

	campaignRuleProxy.createOutboundCampaignruleAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, campaignRule *platformclientv2.Campaignrule) (*platformclientv2.Campaignrule, *platformclientv2.APIResponse, error) {
		campaignRule.Id = &tId

		equal := cmp.Equal(testCampaignRule, *campaignRule)
		assert.Equal(t, true, equal, "campaignRule not equal to expected value in create: %s", cmp.Diff(testCampaignRule, *campaignRule))

		return campaignRule, nil, nil
	}

	internalProxy = campaignRuleProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceOutboundCampaignrule().Schema

	//Setup a map of values
	resourceDataMap := buildCampaignRuleResourceMap(tId, *testCampaignRule.Name, *testCampaignRule.Enabled, *testCampaignRule.MatchAnyConditions, *testCampaignRule.CampaignRuleEntities, *testCampaignRule.CampaignRuleConditions, *testCampaignRule.CampaignRuleActions)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := createOutboundCampaignRule(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceOutboundCampaignruleRead(t *testing.T) {
	tId := uuid.NewString()
	tName := "campaign rule name"
	testCampaignRule := generateCampaignRuleData(tId, tName)

	campaignRuleProxy := &outboundCampaignruleProxy{}

	campaignRuleProxy.getOutboundCampaignruleByIdAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.Campaignrule, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		campaignRule := &testCampaignRule

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return campaignRule, apiResponse, nil
	}

	internalProxy = campaignRuleProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceOutboundCampaignrule().Schema

	//Setup a map of values
	resourceDataMap := buildCampaignRuleResourceMap(tId, *testCampaignRule.Name, *testCampaignRule.Enabled, *testCampaignRule.MatchAnyConditions, *testCampaignRule.CampaignRuleEntities, *testCampaignRule.CampaignRuleConditions, *testCampaignRule.CampaignRuleActions)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readOutboundCampaignRule(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())

	campaignRule := getCampaignruleFromResourceData(d)
	campaignRule.Id = platformclientv2.String(d.Id())

	equal := cmp.Equal(testCampaignRule, campaignRule)
	assert.Equal(t, true, equal, "campaignRule not equal to expected value in read: %s", cmp.Diff(testCampaignRule, campaignRule))
}

func TestUnitResourceOutboundCampaignruleUpdate(t *testing.T) {
	tId := uuid.NewString()
	tName := "Updated campaign rule name"
	testCampaignRule := generateCampaignRuleData(tId, tName)

	campaignRulePoxy := &outboundCampaignruleProxy{}
	campaignRulePoxy.getOutboundCampaignruleByIdAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.Campaignrule, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		campaignRule := &testCampaignRule

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return campaignRule, apiResponse, nil
	}

	campaignRulePoxy.updateOutboundCampaignruleAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string, campaignRule *platformclientv2.Campaignrule) (*platformclientv2.Campaignrule, *platformclientv2.APIResponse, error) {
		campaignRule.Id = &tId

		equal := cmp.Equal(testCampaignRule, *campaignRule)
		assert.Equal(t, true, equal, "campaignRule not equal to expected value in update: %s", cmp.Diff(testCampaignRule, *campaignRule))

		return campaignRule, nil, nil
	}

	internalProxy = campaignRulePoxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceOutboundCampaignrule().Schema

	//Setup a map of values
	resourceDataMap := buildCampaignRuleResourceMap(tId, *testCampaignRule.Name, *testCampaignRule.Enabled, *testCampaignRule.MatchAnyConditions, *testCampaignRule.CampaignRuleEntities, *testCampaignRule.CampaignRuleConditions, *testCampaignRule.CampaignRuleActions)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateOutboundCampaignRule(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, *testCampaignRule.Name, d.Get("name").(string))
}

func TestUnitResourceOutboundCampaignruleDelete(t *testing.T) {
	tId := uuid.NewString()
	tName := "campaign rule name"
	testCampaignRule := generateCampaignRuleData(tId, tName)

	campaignRulePoxy := &outboundCampaignruleProxy{}

	campaignRulePoxy.deleteOutboundCampaignruleAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	campaignRulePoxy.getOutboundCampaignruleByIdAttr = func(ctx context.Context, proxy *outboundCampaignruleProxy, id string) (*platformclientv2.Campaignrule, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
		err := fmt.Errorf("Unable to find targeted IVR: %s", id)
		return nil, apiResponse, err
	}

	internalProxy = campaignRulePoxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceOutboundCampaignrule().Schema

	//Setup a map of values
	resourceDataMap := buildCampaignRuleResourceMap(tId, *testCampaignRule.Name, *testCampaignRule.Enabled, *testCampaignRule.MatchAnyConditions, *testCampaignRule.CampaignRuleEntities, *testCampaignRule.CampaignRuleConditions, *testCampaignRule.CampaignRuleActions)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteOutboundCampaignRule(ctx, d, gcloud)
	assert.Nil(t, diag)
	assert.Equal(t, tId, d.Id())
}

func generateCampaignRuleData(id string, name string) platformclientv2.Campaignrule {
	// Create campaign rule entity
	campaignsEntities := make([]platformclientv2.Domainentityref, 3)
	sequencesEntities := make([]platformclientv2.Domainentityref, 3)
	var smsCampaignsEntities []platformclientv2.Domainentityref
	var emailCampaignsEntities []platformclientv2.Domainentityref
	for i := 0; i < 3; i++ {
		campaignsEntities[i] = generateRandomDomainEntityRef()
		sequencesEntities[i] = generateRandomDomainEntityRef()
	}
	campaignRuleEntities := platformclientv2.Campaignruleentities{
		Campaigns:      &campaignsEntities,
		Sequences:      &sequencesEntities,
		SmsCampaigns:   &smsCampaignsEntities,
		EmailCampaigns: &emailCampaignsEntities,
	}

	// Create campaign rule conditions
	parameterCondition := platformclientv2.Campaignruleparameters{
		Operator:    platformclientv2.String("lessThan"),
		Value:       platformclientv2.String("0.5"),
		DialingMode: platformclientv2.String("preview"),
		Priority:    platformclientv2.String("2"),
	}
	campaignRuleCondition := platformclientv2.Campaignrulecondition{
		Id:            platformclientv2.String(uuid.NewString()),
		ConditionType: platformclientv2.String("campaignProgress"),
		Parameters:    &parameterCondition,
	}
	campaignRuleConditions := []platformclientv2.Campaignrulecondition{campaignRuleCondition}

	// Create campaign rule actions
	parameterAction := platformclientv2.Campaignruleparameters{
		Operator:    platformclientv2.String("lessThan"),
		Value:       platformclientv2.String("0.5"),
		DialingMode: platformclientv2.String("preview"),
		Priority:    platformclientv2.String("2"),
	}
	campaignsActions := make([]platformclientv2.Domainentityref, 3)
	sequencesActions := make([]platformclientv2.Domainentityref, 3)
	var smsCampaignsActions []platformclientv2.Domainentityref
	var emailCampaignsActions []platformclientv2.Domainentityref

	for i := 0; i < 3; i++ {
		campaignsActions[i] = generateRandomDomainEntityRef()
		sequencesActions[i] = generateRandomDomainEntityRef()
	}
	actionEntities := platformclientv2.Campaignruleactionentities{
		UseTriggeringEntity: platformclientv2.Bool(false),
		Campaigns:           &campaignsActions,
		Sequences:           &sequencesActions,
		SmsCampaigns:        &smsCampaignsActions,
		EmailCampaigns:      &emailCampaignsActions,
	}
	campaignRuleAction := platformclientv2.Campaignruleaction{
		Id:                         platformclientv2.String(uuid.NewString()),
		ActionType:                 platformclientv2.String("turnOnCampaign"),
		Parameters:                 &parameterAction,
		CampaignRuleActionEntities: &actionEntities,
	}
	campaignRuleActions := []platformclientv2.Campaignruleaction{campaignRuleAction}

	return platformclientv2.Campaignrule{
		Id:                     &id,
		Name:                   &name,
		Enabled:                platformclientv2.Bool(false),
		MatchAnyConditions:     platformclientv2.Bool(true),
		CampaignRuleEntities:   &campaignRuleEntities,
		CampaignRuleConditions: &campaignRuleConditions,
		CampaignRuleActions:    &campaignRuleActions,
	}
}

func generateRandomDomainEntityRef() platformclientv2.Domainentityref {
	id := uuid.NewString()
	return platformclientv2.Domainentityref{
		Id: &id,
	}
}

// tCampaignRuleConditions interface{}, tCampaignRuleActions interface{}
func buildCampaignRuleResourceMap(tId string, tName string, tEnabled bool, tMatchAnyConditions bool, tCampaignRuleEntities platformclientv2.Campaignruleentities, tCampaignRuleConditions []platformclientv2.Campaignrulecondition, tCampaignRuleActions []platformclientv2.Campaignruleaction) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":                       tId,
		"name":                     tName,
		"enabled":                  tEnabled,
		"match_any_conditions":     tMatchAnyConditions,
		"campaign_rule_entities":   generateCampaignruleEntityInterface(&tCampaignRuleEntities),
		"campaign_rule_conditions": flattenCampaignRuleConditions(&tCampaignRuleConditions),
		"campaign_rule_actions":    flattenCampaignRuleAction(&tCampaignRuleActions, generateActionEntities),
	}
	return resourceDataMap
}

func generateCampaignruleEntityInterface(campaignRuleEntities *platformclientv2.Campaignruleentities) []interface{} {
	var (
		campaignRuleEntitiesMap = make(map[string]interface{})
		campaigns               []interface{}
		sequences               []interface{}
		smsCampaigns            []interface{}
		emailCampaigns          []interface{}
	)

	if campaignRuleEntities.Campaigns != nil {
		for _, v := range *campaignRuleEntities.Campaigns {
			campaigns = append(campaigns, *v.Id)
		}
	}

	if campaignRuleEntities.Sequences != nil {
		for _, v := range *campaignRuleEntities.Sequences {
			sequences = append(sequences, *v.Id)
		}
	}

	if campaignRuleEntities.SmsCampaigns != nil {
		for _, v := range *campaignRuleEntities.SmsCampaigns {
			smsCampaigns = append(smsCampaigns, *v.Id)
		}
	}

	if campaignRuleEntities.EmailCampaigns != nil {
		for _, v := range *campaignRuleEntities.EmailCampaigns {
			emailCampaigns = append(emailCampaigns, *v.Id)
		}
	}

	campaignRuleEntitiesMap["campaign_ids"] = campaigns
	campaignRuleEntitiesMap["sequence_ids"] = sequences
	campaignRuleEntitiesMap["sms_campaign_ids"] = smsCampaigns
	campaignRuleEntitiesMap["email_campaign_ids"] = emailCampaigns

	return []interface{}{campaignRuleEntitiesMap}
}

func generateActionEntities(entities *platformclientv2.Campaignruleactionentities) []interface{} {
	var (
		campaigns      []interface{}
		sequences      []interface{}
		smsCampaigns   []interface{}
		emailCampaigns []interface{}
		entitiesMap    = make(map[string]interface{})
	)

	if entities == nil {
		return nil
	}

	if entities.Campaigns != nil {
		for _, campaign := range *entities.Campaigns {
			campaigns = append(campaigns, *campaign.Id)
		}
	}

	if entities.Sequences != nil {
		for _, sequence := range *entities.Sequences {
			sequences = append(sequences, *sequence.Id)
		}
	}

	if entities.SmsCampaigns != nil {
		for _, v := range *entities.SmsCampaigns {
			smsCampaigns = append(smsCampaigns, *v.Id)
		}
	}

	if entities.EmailCampaigns != nil {
		for _, v := range *entities.EmailCampaigns {
			emailCampaigns = append(emailCampaigns, *v.Id)
		}
	}

	entitiesMap["campaign_ids"] = campaigns
	entitiesMap["sequence_ids"] = sequences
	entitiesMap["sms_campaign_ids"] = smsCampaigns
	entitiesMap["email_campaign_ids"] = emailCampaigns
	entitiesMap["use_triggering_entity"] = *entities.UseTriggeringEntity

	return []interface{}{entitiesMap}
}

// TestUnitCanonicalTimeOfDay covers the normalisation behind the time-of-day drift fix. The
// campaign rules API accepts HH:mm, HH:mm:ss and HH:mm:ss.fff but always stores HH:mm:ss, so any
// other spelling left in a config sits against the stored value and every plan reports a change
// that never settles. Rewrites that discard nothing are normalised; a non-zero fraction is left
// alone so that the loss the API inflicts stays visible to the operator.
func TestUnitCanonicalTimeOfDay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"already canonical", "12:00:00", "12:00:00"},
		{"zero milliseconds", "12:00:00.000", "12:00:00"},
		{"single zero fraction digit", "12:00:00.0", "12:00:00"},
		{"two zero fraction digits", "12:00:00.00", "12:00:00"},
		{"trailing dot with no digits", "12:00:00.", "12:00:00"},
		{"hours and minutes only", "12:00", "12:00:00"},
		{"hours and minutes with zero fraction", "12:00.000", "12:00:00"},
		{"midnight without seconds", "00:00", "00:00:00"},
		{"non-zero fraction is preserved", "12:00:00.500", "12:00:00.500"},
		{"small non-zero fraction is preserved", "12:00:00.010", "12:00:00.010"},
		{"long non-zero fraction is preserved", "12:00:00.123456", "12:00:00.123456"},
		{"empty string", "", ""},
		{"unrecognised text", "not-a-time", "not-a-time"},
		{"too many colons", "12:00:00:00", "12:00:00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, canonicalTimeOfDay(tt.input))
		})
	}
}

// TestUnitNormalizeTimeOfDay covers the schema.SchemaStateFunc adapter. StateFunc runs during
// plan, so the non-string guard matters more than the value it returns.
func TestUnitNormalizeTimeOfDay(t *testing.T) {
	t.Parallel()

	t.Run("drops a zero fraction", func(t *testing.T) {
		assert.Equal(t, "12:00:00", normalizeTimeOfDay("12:00:00.000"))
	})

	t.Run("pads a value with no seconds", func(t *testing.T) {
		assert.Equal(t, "09:30:00", normalizeTimeOfDay("09:30"))
	})

	t.Run("preserves a non-zero fraction", func(t *testing.T) {
		assert.Equal(t, "12:00:00.500", normalizeTimeOfDay("12:00:00.500"))
	})

	t.Run("returns empty for a non-string without panicking", func(t *testing.T) {
		assert.NotPanics(t, func() {
			assert.Equal(t, "", normalizeTimeOfDay(42))
		})
	})

	t.Run("returns empty for nil without panicking", func(t *testing.T) {
		assert.NotPanics(t, func() {
			assert.Equal(t, "", normalizeTimeOfDay(nil))
		})
	})
}

// TestUnitSuppressTimeOfDayDiff covers the diff comparison. It is deliberately symmetric: the
// campaign rules API trims the fraction, but the sibling callable timeset API appends one, so the
// comparison has to hold whichever side carries the odd spelling.
func TestUnitSuppressTimeOfDayDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		old      string
		new      string
		suppress bool
	}{
		{"config adds zero milliseconds", "12:00:00", "12:00:00.000", true},
		{"state carries zero milliseconds", "12:00:00.000", "12:00:00", true},
		{"config omits seconds", "12:00:00", "12:00", true},
		{"state omits seconds", "12:00", "12:00:00", true},
		{"identical values", "12:00:00", "12:00:00", true},
		{"both empty", "", "", true},
		{"non-zero fraction stays visible", "12:00:00", "12:00:00.500", false},
		{"a real change stays visible", "12:00:00", "13:00:00", false},
		{"a real change carrying fractions stays visible", "12:00:00.000", "13:00:00.000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.suppress, suppressTimeOfDayDiff("", tt.old, tt.new, nil))
		})
	}
}

// nestedSchema walks a path of attribute names down through nested block schemas, descending into
// each Elem along the way, and returns the leaf schema.
func nestedSchema(t *testing.T, root map[string]*schema.Schema, path ...string) *schema.Schema {
	t.Helper()

	current := root
	for i, name := range path {
		s, ok := current[name]
		if !ok {
			t.Fatalf("attribute %q not found at position %d of %v", name, i, path)
		}
		if i == len(path)-1 {
			return s
		}
		res, ok := s.Elem.(*schema.Resource)
		if !ok {
			t.Fatalf("attribute %q at position %d of %v has no nested resource to descend into", name, i, path)
		}
		current = res.Schema
	}

	return nil
}

// TestUnitTimeOfDayFieldsAreWiredForDrift is the guard that keeps the drift fix in place. The
// helper tests above pass whether or not the schema actually references those helpers, so without
// this a later edit could quietly drop StateFunc or DiffSuppressFunc from a field and reintroduce
// the drift. Asserting through the retrieved hooks catches a wrong function as well as a missing
// one. Passing nil for *schema.ResourceData is safe because suppressTimeOfDayDiff ignores it.
func TestUnitTimeOfDayFieldsAreWiredForDrift(t *testing.T) {
	t.Parallel()

	root := ResourceOutboundCampaignrule().Schema

	fields := []struct {
		name string
		path []string
	}{
		{
			name: "time_of_day.threshold_value",
			path: []string{"condition_groups", "conditions", "date_time_parameters", "time_of_day", "threshold_value"},
		},
		{
			name: "time_of_day.interval.min",
			path: []string{"condition_groups", "conditions", "date_time_parameters", "time_of_day", "interval", "min"},
		},
		{
			name: "time_of_day.interval.max",
			path: []string{"condition_groups", "conditions", "date_time_parameters", "time_of_day", "interval", "max"},
		},
	}

	for _, field := range fields {
		t.Run(field.name, func(t *testing.T) {
			leaf := nestedSchema(t, root, field.path...)

			if !assert.NotNil(t, leaf.StateFunc, "StateFunc is not wired; time-of-day values will drift again") {
				return
			}
			if !assert.NotNil(t, leaf.DiffSuppressFunc, "DiffSuppressFunc is not wired; time-of-day values will drift again") {
				return
			}

			assert.Equal(t, "12:00:00", leaf.StateFunc("12:00:00.000"), "a zero fraction should be dropped before reaching state")
			assert.Equal(t, "12:00:00", leaf.StateFunc("12:00"), "missing seconds should be padded before reaching state")
			assert.Equal(t, "12:00:00.500", leaf.StateFunc("12:00:00.500"), "a real fraction should survive into state so the loss is visible")

			assert.True(t, leaf.DiffSuppressFunc("", "12:00:00", "12:00:00.000", nil), "a zero fraction should be suppressed")
			assert.True(t, leaf.DiffSuppressFunc("", "12:00:00", "12:00", nil), "a missing seconds part should be suppressed")
			assert.False(t, leaf.DiffSuppressFunc("", "12:00:00", "12:00:00.500", nil), "a real fraction should stay visible")
			assert.False(t, leaf.DiffSuppressFunc("", "12:00:00", "13:00:00", nil), "a real change should stay visible")
		})
	}
}
