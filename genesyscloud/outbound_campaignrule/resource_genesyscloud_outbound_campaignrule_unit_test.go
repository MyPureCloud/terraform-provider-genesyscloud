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

// TestUnitSecondsToISO8601Duration covers the write side of for_duration. The API accepts a plain
// seconds duration and canonicalises it itself, so the provider only ever needs the PT<n>S form.
func TestUnitSecondsToISO8601Duration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{"fifteen minutes as seconds", 900, "PT900S"},
		{"one second", 1, "PT1S"},
		{"an hour", 3600, "PT3600S"},
		{"more than a day", 90000, "PT90000S"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, secondsToISO8601Duration(tt.seconds))
		})
	}
}

// TestUnitParseISO8601DurationSeconds pins the parser to the forms the API was observed to return,
// rather than to a reading of the ISO-8601 spec. Every "ok" case below was produced by the live API
// during the OBR-1683 probe: it canonicalises whatever it is given into Java's Duration.toString()
// form, rolling days into hours, so PT90000S comes back as PT25H and no day component ever appears.
func TestUnitParseISO8601DurationSeconds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		seconds int
		exact   bool
		ok      bool
	}{
		// Forms the API returned during the probe.
		{"minutes only", "PT15M", 900, true, true},
		{"seconds only", "PT900S", 900, true, true},
		{"hours and minutes", "PT2H10M", 7800, true, true},
		{"hours beyond a day", "PT25H", 90000, true, true},
		{"all three components", "PT1H1M1S", 3661, true, true},
		{"zero", "PT0S", 0, true, true},

		// Fractional seconds are accepted and preserved by the API but cannot be represented.
		{"fractional seconds lose precision", "PT1.5S", 1, false, true},
		{"fraction below a second", "PT0.5S", 0, false, true},
		{"zero fraction loses nothing", "PT1.0S", 1, true, true},
		{"fraction alongside other components", "PT1M0.250S", 60, false, true},

		// Rejected: the API refuses these too, or would never emit them.
		{"negative", "PT-5S", 0, false, false},
		{"bare designator", "PT", 0, false, false},
		{"day component", "P1D", 0, false, false},
		{"day and time components", "P1DT2H", 0, false, false},
		{"not a duration", "15m", 0, false, false},
		{"unknown unit", "PT1X", 0, false, false},
		{"empty", "", 0, false, false},
		{"components out of order", "PT1M1H", 0, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seconds, exact, ok := parseISO8601DurationSeconds(tt.input)
			assert.Equal(t, tt.ok, ok, "ok")
			if !tt.ok {
				return
			}
			assert.Equal(t, tt.seconds, seconds, "seconds")
			assert.Equal(t, tt.exact, exact, "exact")
		})
	}
}

// TestUnitForDurationRoundTrip is the property the seconds-block schema depends on: whatever the
// API does to the value in between, writing seconds and reading them back returns the same number.
// This is what makes a seconds block immune to the canonicalisation drift that a raw ISO-8601
// string field would suffer.
func TestUnitForDurationRoundTrip(t *testing.T) {
	t.Parallel()

	for _, seconds := range []int{1, 30, 90, 900, 3600, 3661, 7800, 90000} {
		t.Run(fmt.Sprintf("%d seconds", seconds), func(t *testing.T) {
			parsed, exact, ok := parseISO8601DurationSeconds(secondsToISO8601Duration(seconds))
			assert.True(t, ok, "the value we generate must parse back")
			assert.True(t, exact, "whole seconds must round-trip without loss")
			assert.Equal(t, seconds, parsed)
		})
	}
}

// TestUnitFlattenForDuration covers the read path, including the two cases that must not fail the
// read: a value the parser does not recognise, and one carrying precision the schema cannot hold.
// Failing either would break terraform import and tf export against a rule created elsewhere.
func TestUnitFlattenForDuration(t *testing.T) {
	t.Parallel()

	t.Run("nil gives no block", func(t *testing.T) {
		assert.Nil(t, flattenForDuration(nil))
	})

	t.Run("empty string gives no block", func(t *testing.T) {
		assert.Nil(t, flattenForDuration(platformclientv2.String("")))
	})

	t.Run("canonicalised value parses back to the configured seconds", func(t *testing.T) {
		assert.Equal(t,
			[]interface{}{map[string]interface{}{"seconds": 900}},
			flattenForDuration(platformclientv2.String("PT15M")))
	})

	t.Run("sub-second precision is truncated rather than dropped", func(t *testing.T) {
		assert.Equal(t,
			[]interface{}{map[string]interface{}{"seconds": 1}},
			flattenForDuration(platformclientv2.String("PT1.5S")))
	})

	t.Run("unparseable value gives no block rather than zero seconds", func(t *testing.T) {
		assert.Nil(t, flattenForDuration(platformclientv2.String("garbage")))
	})
}
