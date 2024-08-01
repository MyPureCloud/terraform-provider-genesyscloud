package recording_media_retention_policy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_recording_media_retention_policy_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.

Each proxy implementation:

1.  Should provide a private package level variable that holds a instance of a proxy class.
2.  A New... constructor function  to initialize the proxy object.  This constructor should only be used within
    the proxy.
3.  A get private constructor function that the classes in the package can be used to to retrieve
    the proxy.  This proxy should check to see if the package level proxy instance is nil and
    should initialize it, otherwise it should return the instance
4.  Type definitions for each function that will be used in the proxy.  We use composition here
    so that we can easily provide mocks for testing.
5.  A struct for the proxy that holds an attribute for each function type.
6.  Wrapper methods on each of the elements on the struct.
7.  Function implementations for each function type definition.

*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *policyProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllPoliciesFunc func(ctx context.Context, p *policyProxy) (*[]platformclientv2.Policy, *platformclientv2.APIResponse, error)
type createPolicyFunc func(ctx context.Context, p *policyProxy, policyCreate *platformclientv2.Policycreate) (*platformclientv2.Policy, *platformclientv2.APIResponse, error)
type getPolicyByIdFunc func(ctx context.Context, p *policyProxy, policyId string) (policy *platformclientv2.Policy, response *platformclientv2.APIResponse, err error)
type getPolicyByNameFunc func(ctx context.Context, p *policyProxy, policyName string) (policy *platformclientv2.Policy, retryable bool, response *platformclientv2.APIResponse, err error)
type updatePolicyFunc func(ctx context.Context, p *policyProxy, policyId string, policy *platformclientv2.Policy) (*platformclientv2.Policy, *platformclientv2.APIResponse, error)
type deletePolicyFunc func(ctx context.Context, p *policyProxy, policyId string) (response *platformclientv2.APIResponse, err error)
type getFormsEvaluationFunc func(ctx context.Context, p *policyProxy, formId string) (*platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error)
type getEvaluationFormRecentVerIdFunc func(ctx context.Context, p *policyProxy, formId string) (string, *platformclientv2.APIResponse, error)
type getQualityFormsSurveyByNameFunc func(ctx context.Context, p *policyProxy, surveyName string) (*platformclientv2.Publishedsurveyformreference, *platformclientv2.APIResponse, error)

// integrationProxy contains all of the methods that call genesys cloud APIs.
type policyProxy struct {
	clientConfig                     *platformclientv2.Configuration
	qualityApi                       *platformclientv2.QualityApi
	recordingApi                     *platformclientv2.RecordingApi
	getAllPoliciesAttr               getAllPoliciesFunc
	createPolicyAttr                 createPolicyFunc
	getPolicyByIdAttr                getPolicyByIdFunc
	getPolicyByNameAttr              getPolicyByNameFunc
	updatePolicyAttr                 updatePolicyFunc
	deletePolicyAttr                 deletePolicyFunc
	getFormsEvaluationAttr           getFormsEvaluationFunc
	getEvaluationFormRecentVerIdAttr getEvaluationFormRecentVerIdFunc
	getQualityFormsSurveyByNameAttr  getQualityFormsSurveyByNameFunc
}

// newPolicyProxy initializes the Policy proxy with all of the data needed to communicate with Genesys Cloud
func newPolicyProxy(clientConfig *platformclientv2.Configuration) *policyProxy {
	qApi := platformclientv2.NewQualityApiWithConfig(clientConfig)
	rApi := platformclientv2.NewRecordingApiWithConfig(clientConfig)
	return &policyProxy{
		clientConfig:                     clientConfig,
		qualityApi:                       qApi,
		recordingApi:                     rApi,
		getAllPoliciesAttr:               getAllPoliciesFn,
		createPolicyAttr:                 createPolicyFn,
		getPolicyByIdAttr:                getPolicyByIdFn,
		getPolicyByNameAttr:              getPolicyByNameFn,
		updatePolicyAttr:                 updatePolicyFn,
		deletePolicyAttr:                 deletePolicyFn,
		getFormsEvaluationAttr:           getFormsEvaluationFn,
		getEvaluationFormRecentVerIdAttr: getEvaluationFormRecentVerIdFn,
		getQualityFormsSurveyByNameAttr:  getQualityFormsSurveyByNameFn,
	}
}

// getPolicyProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getPolicyProxy(clientConfig *platformclientv2.Configuration) *policyProxy {
	if internalProxy == nil {
		internalProxy = newPolicyProxy(clientConfig)
	}
	return internalProxy
}

// getAllPolicies retrieves all Genesys Cloud Recording Media Retention Policies
func (p *policyProxy) getAllPolicies(ctx context.Context) (*[]platformclientv2.Policy, *platformclientv2.APIResponse, error) {
	return p.getAllPoliciesAttr(ctx, p)
}

// createPolicy creates a Genesys Cloud Recording Media Retention Policy
func (p *policyProxy) createPolicy(ctx context.Context, policyCreate *platformclientv2.Policycreate) (*platformclientv2.Policy, *platformclientv2.APIResponse, error) {
	return p.createPolicyAttr(ctx, p, policyCreate)
}

// getPolicyById gets a Genesys Cloud Recording Media Retention Policy by id
func (p *policyProxy) getPolicyById(ctx context.Context, policyId string) (policy *platformclientv2.Policy, response *platformclientv2.APIResponse, err error) {
	return p.getPolicyByIdAttr(ctx, p, policyId)
}

// getPolicyByName gets a Genesys Cloud Recording Media Retention Policy by name
func (p *policyProxy) getPolicyByName(ctx context.Context, policyName string) (policy *platformclientv2.Policy, retryable bool, response *platformclientv2.APIResponse, err error) {
	return p.getPolicyByNameAttr(ctx, p, policyName)
}

// updatePolicy updates a Genesys Cloud Recording Media Retention Policy
func (p *policyProxy) updatePolicy(ctx context.Context, policyId string, policy *platformclientv2.Policy) (*platformclientv2.Policy, *platformclientv2.APIResponse, error) {
	return p.updatePolicyAttr(ctx, p, policyId, policy)
}

// deletePolicy deletes a Genesys Cloud Recording Media Retention Policy
func (p *policyProxy) deletePolicy(ctx context.Context, policyId string) (response *platformclientv2.APIResponse, err error) {
	return p.deletePolicyAttr(ctx, p, policyId)
}

// getFormsEvaluation gets a Genesys Cloud Evaluation Form by id
func (p *policyProxy) getFormsEvaluation(ctx context.Context, formId string) (*platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error) {
	return p.getFormsEvaluationAttr(ctx, p, formId)
}

// getFormsEvaluation gets the most recent unpublished version id of a Genesys Cloud Evaluation Form
func (p *policyProxy) getEvaluationFormRecentVerId(ctx context.Context, formId string) (string, *platformclientv2.APIResponse, error) {
	return p.getEvaluationFormRecentVerIdAttr(ctx, p, formId)
}

// getQualityFormsSurveyByName gets a Genesys Cloud Survey Form by name
func (p *policyProxy) getQualityFormsSurveyByName(ctx context.Context, surveyName string) (*platformclientv2.Publishedsurveyformreference, *platformclientv2.APIResponse, error) {
	return p.getQualityFormsSurveyByNameAttr(ctx, p, surveyName)
}

// getAllIntegrationCredsFn is the implementation for getting all media retention policy in Genesys Cloud
func getAllPoliciesFn(ctx context.Context, p *policyProxy) (*[]platformclientv2.Policy, *platformclientv2.APIResponse, error) {
	var allPolicies []platformclientv2.Policy
	const pageSize = 100

	retentionPolicies, resp, err := callGetAllPoliciesApi(pageSize, 1, p.qualityApi.Configuration)
	if err != nil {
		return nil, resp, err
	}

	if retentionPolicies.Entities == nil || len(*retentionPolicies.Entities) == 0 {
		return &allPolicies, resp, nil
	}

	allPolicies = append(allPolicies, *retentionPolicies.Entities...)

	for pageNum := 2; pageNum <= *retentionPolicies.PageCount; pageNum++ {
		retentionPolicies, resp, err = callGetAllPoliciesApi(pageSize, pageNum, p.qualityApi.Configuration)
		if err != nil {
			return nil, resp, err
		}

		if retentionPolicies.Entities == nil || len(*retentionPolicies.Entities) == 0 {
			break
		}

		allPolicies = append(allPolicies, *retentionPolicies.Entities...)
	}

	return &allPolicies, resp, nil
}

// createPolicyFn is the implementation for creating a media retention policy in Genesys Cloud
func createPolicyFn(ctx context.Context, p *policyProxy, policyCreate *platformclientv2.Policycreate) (*platformclientv2.Policy, *platformclientv2.APIResponse, error) {
	policy, resp, err := p.recordingApi.PostRecordingMediaretentionpolicies(*policyCreate)
	if err != nil {
		return nil, resp, err
	}
	return policy, resp, nil
}

// getPolicyByIdFn is the implementation for getting a media retention policy in Genesys Cloud by id
func getPolicyByIdFn(ctx context.Context, p *policyProxy, policyId string) (policy *platformclientv2.Policy, response *platformclientv2.APIResponse, err error) {
	policy, resp, err := p.recordingApi.GetRecordingMediaretentionpolicy(policyId)
	if err != nil {
		return nil, resp, err
	}
	return policy, resp, nil
}

// getPolicyByNameFn is the implementation for getting a media retention policy in Genesys Cloud by name
func getPolicyByNameFn(ctx context.Context, p *policyProxy, policyName string) (policy *platformclientv2.Policy, retryable bool, response *platformclientv2.APIResponse, err error) {
	policies, resp, err := getAllPoliciesFn(ctx, p)
	if err != nil {
		return nil, false, resp, err
	}

	if policies == nil || len(*policies) == 0 {
		return nil, true, resp, fmt.Errorf("no media retention policy found with name %s", policyName)
	}

	for _, policy := range *policies {
		if *policy.Name == policyName {
			return &policy, false, resp, nil
		}
	}

	return nil, true, resp, fmt.Errorf("unable to find media retention policy with name %s", policyName)
}

// updatePolicyFn is the implementation for updating a media retention policy in Genesys Cloud
func updatePolicyFn(ctx context.Context, p *policyProxy, policyId string, policyBody *platformclientv2.Policy) (*platformclientv2.Policy, *platformclientv2.APIResponse, error) {
	policy, resp, err := p.recordingApi.PutRecordingMediaretentionpolicy(policyId, *policyBody)
	if err != nil {
		return nil, resp, err
	}
	return policy, resp, nil
}

// deletePolicyFn is the implementation for deleting a media retention policy in Genesys Cloud
func deletePolicyFn(ctx context.Context, p *policyProxy, policyId string) (response *platformclientv2.APIResponse, err error) {
	resp, err := p.recordingApi.DeleteRecordingMediaretentionpolicy(policyId)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// getFormsEvaluationFn is the implementation for getting an evaluation form in Genesys Cloud
func getFormsEvaluationFn(ctx context.Context, p *policyProxy, formId string) (*platformclientv2.Evaluationformresponse, *platformclientv2.APIResponse, error) {
	form, resp, err := p.qualityApi.GetQualityFormsEvaluation(formId)
	if err != nil {
		return nil, resp, err
	}
	return form, resp, nil
}

// getEvaluationFormRecentVerIdFn is the implementation for getting the most recent version if of an evaluation form in Genesys Cloud
func getEvaluationFormRecentVerIdFn(ctx context.Context, p *policyProxy, formId string) (string, *platformclientv2.APIResponse, error) {
	formVersions, resp, err := p.qualityApi.GetQualityFormsEvaluationVersions(formId, 25, 1, "desc")
	if err != nil {
		return "", resp, err
	}
	if formVersions.Entities == nil || len(*formVersions.Entities) == 0 {
		return "", resp, fmt.Errorf("no versions found for form %s", formId)
	}
	return *(*formVersions.Entities)[0].Id, resp, nil
}

// getQualityFormsSurveyByNameFn is the implementation for getting a survey form in Genesys Cloud
func getQualityFormsSurveyByNameFn(ctx context.Context, p *policyProxy, surveyName string) (*platformclientv2.Publishedsurveyformreference, *platformclientv2.APIResponse, error) {
	const pageNum = 1
	const pageSize = 100
	forms, resp, err := p.qualityApi.GetQualityFormsSurveys(pageSize, pageNum, "", "", "", "", surveyName, "desc")
	if err != nil {
		return nil, resp, err
	}
	if forms.Entities == nil || len(*forms.Entities) == 0 {
		return nil, resp, fmt.Errorf("no survey forms found with name %s", surveyName)
	}
	surveyFormReference := platformclientv2.Publishedsurveyformreference{Name: &surveyName, ContextId: (*forms.Entities)[0].ContextId}
	return &surveyFormReference, resp, nil
}

// We need to call /api/v2/recording/mediaretentionpolicies manually to avoid setting optional boolean and integer parameters than filter results
func callGetAllPoliciesApi(pageSize, pageNumber int, config *platformclientv2.Configuration) (*platformclientv2.Policyentitylisting, *platformclientv2.APIResponse, error) {
	apiClient := &config.APIClient

	// create path and map variables
	path := config.BasePath + "/api/v2/recording/mediaretentionpolicies"

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)
	formParams := url.Values{}
	var postBody interface{}
	var postFileName string
	var fileBytes []byte

	// oauth required
	if config.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + config.AccessToken
	}
	// add default headers if any
	for key := range config.DefaultHeader {
		headerParams[key] = config.DefaultHeader[key]
	}

	queryParams["pageSize"] = apiClient.ParameterToString(pageSize, "")
	queryParams["pageNumber"] = apiClient.ParameterToString(pageNumber, "")

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *platformclientv2.Policyentitylisting
	response, err := apiClient.CallAPI(path, http.MethodGet, postBody, headerParams, queryParams, formParams, postFileName, fileBytes)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = fmt.Errorf(response.ErrorMessage)
	} else {
		err = json.Unmarshal(response.RawBody, &successPayload)
	}
	return successPayload, response, err
}
