package outbound_ruleset

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func TestRuleSetSuccess(t *testing.T) {
	var (
		
		ruleSetId             = uuid.NewString()
	)

	ruleset := &platformclientv2.Ruleset{
		Id: &ruleSetId,
	}

	outboundAPIProxy := NewOutboundAPIProxy()
	
	outboundAPIProxy.CreateOutboundRulesets = createMockPostOutboundRulesetsFunc(ruleSetId, nil)
	outboundAPIProxy.UpdateOutboundRuleset = createMockPutOutboundRulesetFunc(nil)

	t.Log("Testing Post RuleSet function")
	ruleset, _, err := outboundAPIProxy.CreateOutboundRulesets(outboundAPIProxy, *ruleset)
	if err != nil {
		t.Errorf("Expected error to be nil, got '%v'", err)
	}

	

	t.Log("Testing Put RuleSet function")
	ruleset, _, err = outboundAPIProxy.UpdateOutboundRuleset(outboundAPIProxy, *ruleset, ruleSetId)
	if err != nil {
		t.Errorf("Expected error to be nil, got '%v'", err)
	}

	

}

type ruleSetErrorTestData struct {
	mockGetFunction  getOutboundRulesetFunc
	mockPutFunction  putOutboundRulesetFunc
	mockPostFunction postOutboundRulesetsFunc
	mockError        error
}

func TestRuleSetError(t *testing.T) {
	var (
		ruleSetId             = uuid.NewString()
		mockPostError     = fmt.Errorf("error on proxy.PostOutboundRulesets")
		mockPutError      = fmt.Errorf("error on proxy.PutOutboundRuleset")
	)

	ruleSet := platformclientv2.Ruleset{
		Id: &ruleSetId,
	}

	outboundAPIProxy := NewOutboundAPIProxy()

	

	testCases := []ruleSetErrorTestData{
		ruleSetErrorTestData{
			mockGetFunction:  createMockGetOutboundRulesetsFunc(ruleSetId, nil, nil),
			mockPostFunction: createMockPostOutboundRulesetsFunc(ruleSetId, nil),
			mockPutFunction:  createMockPutOutboundRulesetFunc(mockPutError),
			mockError:        mockPutError,
		},
	}

	t.Log("Testing error handling on proxy.PutOutboundRuleset")
	for _, test := range testCases {
		outboundAPIProxy.ReadOutboundRuleset = test.mockGetFunction
		outboundAPIProxy.CreateOutboundRulesets = test.mockPostFunction
		outboundAPIProxy.UpdateOutboundRuleset = test.mockPutFunction

		_, _, err := outboundAPIProxy.UpdateOutboundRuleset(outboundAPIProxy, ruleSet, ruleSetId)
		if err == nil {
			t.Errorf("Expected non nil error")
		}
		if !strings.Contains(fmt.Sprintf("%v", err), fmt.Sprintf("%v", test.mockError)) {
			t.Errorf("Expected to receive error containing '%v', got '%v'", test.mockError, err)
		}
	}

	testCases = []ruleSetErrorTestData{
		ruleSetErrorTestData{
			mockGetFunction:  createMockGetOutboundRulesetsFunc(ruleSetId, nil, nil),
			mockPostFunction: createMockPostOutboundRulesetsFunc(ruleSetId, mockPostError),
			mockPutFunction:  createMockPutOutboundRulesetFunc(nil),
			mockError:        mockPostError,
		},
	}

	t.Log("Testing error handling on proxy.PostOutboundRuleset")
	for _, test := range testCases {
		if test.mockError == mockPostError {
			continue
		}
		outboundAPIProxy.ReadOutboundRuleset = test.mockGetFunction
		outboundAPIProxy.CreateOutboundRulesets = test.mockPostFunction
		outboundAPIProxy.UpdateOutboundRuleset = test.mockPutFunction

		_, _, err := outboundAPIProxy.CreateOutboundRulesets(outboundAPIProxy,ruleSet)
		if err == nil {
			t.Errorf("Expected non nil error")
		}
		if !strings.Contains(fmt.Sprintf("%v", err), fmt.Sprintf("%v", test.mockError)) {
			t.Errorf("Expected to receive error containing '%v', got '%v'", test.mockError, err)
		}
	}
}

func createMockGetOutboundRulesetsFunc(rulesetId string, dnis []string, err error) getOutboundRulesetFunc {
	return func(*OutboundAPIProxy, string) (*platformclientv2.Ruleset, *platformclientv2.APIResponse, error) {
		if err != nil {
			return nil, nil, err
		}
		mockGetRuleSet := &platformclientv2.Ruleset{
			Id:   &rulesetId,
		}
		return mockGetRuleSet, nil, err
	}
}

func createMockPostOutboundRulesetsFunc(rulesetId string, err error) postOutboundRulesetsFunc {
	return func(a *OutboundAPIProxy, ruleset platformclientv2.Ruleset) (*platformclientv2.Ruleset, *platformclientv2.APIResponse, error) {
		if err != nil {
			return nil, nil, err
		}
		ruleset.Id = &rulesetId
		return &ruleset, nil, err
	}
}

func createMockPutOutboundRulesetFunc(err error) putOutboundRulesetFunc {
	return func(a *OutboundAPIProxy, ruleset platformclientv2.Ruleset, id string) (*platformclientv2.Ruleset, *platformclientv2.APIResponse, error) {
		if err != nil {
			return nil, nil, err
		}
		ruleset.Id = &id
		return &ruleset, nil, err
	}
}
