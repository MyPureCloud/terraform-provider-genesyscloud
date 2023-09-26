package architect_grammar

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v109/platformclientv2"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
)

/*
The genesyscloud_architect_grammar_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *architectGrammarProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, error)
type createArchitectGrammarLanguageFunc func(ctx context.Context, p *architectGrammarProxy, grammarId string, language *platformclientv2.Grammarlanguage) (*platformclientv2.Grammarlanguage, error)
type getAllArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy) (*[]platformclientv2.Grammar, error)
type getArchitectGrammarByIdFunc func(ctx context.Context, p *architectGrammarProxy, grammarId string) (grammar *platformclientv2.Grammar, responseCode int, err error)
type getArchitectGrammarIdByNameFunc func(ctx context.Context, p *architectGrammarProxy, name string) (grammarId string, retryable bool, err error)
type updateArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy, grammarId string, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, error)
type deleteArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy, grammarId string) (responseCode int, err error)
type uploadArchitectGrammarFunc func(ctx context.Context, p *architectGrammarProxy, grammarId string, grammar *platformclientv2.Grammar) (responseCode int, err error)

// architectGrammarProxy contains all of the methods that call genesys cloud APIs.
type architectGrammarProxy struct {
	clientConfig                       *platformclientv2.Configuration
	architectApi                       *platformclientv2.ArchitectApi
	createArchitectGrammarAttr         createArchitectGrammarFunc
	createArchitectGrammarLanguageAttr createArchitectGrammarLanguageFunc
	getAllArchitectGrammarAttr         getAllArchitectGrammarFunc
	getArchitectGrammarByIdAttr        getArchitectGrammarByIdFunc
	getArchitectGrammarIdByNameAttr    getArchitectGrammarIdByNameFunc
	updateArchitectGrammarAttr         updateArchitectGrammarFunc
	deleteArchitectGrammarAttr         deleteArchitectGrammarFunc
	uploadArchitectGrammarAttr         uploadArchitectGrammarFunc
}

// newArchitectGrammarProxy initializes the grammar proxy with all of the data needed to communicate with Genesys Cloud
func newArchitectGrammarProxy(clientConfig *platformclientv2.Configuration) *architectGrammarProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &architectGrammarProxy{
		clientConfig:                       clientConfig,
		architectApi:                       api,
		createArchitectGrammarAttr:         createArchitectGrammarFn,
		createArchitectGrammarLanguageAttr: createArchitectGrammarLanguageFn,
		getAllArchitectGrammarAttr:         getAllArchitectGrammarFn,
		getArchitectGrammarByIdAttr:        getArchitectGrammarByIdFn,
		getArchitectGrammarIdByNameAttr:    getArchitectGrammarIdByNameFn,
		updateArchitectGrammarAttr:         updateArchitectGrammarFn,
		deleteArchitectGrammarAttr:         deleteArchitectGrammarFn,
		uploadArchitectGrammarAttr:         uploadArchitectGrammarFn,
	}
}

// getArchitectGrammarProxy acts as a singleton for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getArchitectGrammarProxy(clientConfig *platformclientv2.Configuration) *architectGrammarProxy {
	if internalProxy == nil {
		internalProxy = newArchitectGrammarProxy(clientConfig)
	}

	return internalProxy
}

// createArchitectGrammar creates a Genesys Cloud Architect Grammar
func (p *architectGrammarProxy) createArchitectGrammar(ctx context.Context, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, error) {
	return p.createArchitectGrammarAttr(ctx, p, grammar)
}

// createArchitectGrammarLanguage creates a Genesys Cloud Architect Grammarlanguage for a grammar
func (p *architectGrammarProxy) createArchitectGrammarLanguage(ctx context.Context, grammarId string, language *platformclientv2.Grammarlanguage) (*platformclientv2.Grammarlanguage, error) {
	return p.createArchitectGrammarLanguageAttr(ctx, p, grammarId, language)
}

// getAllArchitectGrammar retrieves all Genesys Cloud Architect Grammar
func (p *architectGrammarProxy) getAllArchitectGrammar(ctx context.Context) (*[]platformclientv2.Grammar, error) {
	return p.getAllArchitectGrammarAttr(ctx, p)
}

// getArchitectGrammarById returns a single Genesys Cloud Architect Grammar by Id
func (p *architectGrammarProxy) getArchitectGrammarById(ctx context.Context, grammarId string) (grammar *platformclientv2.Grammar, statusCode int, err error) {
	return p.getArchitectGrammarByIdAttr(ctx, p, grammarId)
}

// getArchitectGrammarIdByName returns a single Genesys Cloud Architect Grammar by a name
func (p *architectGrammarProxy) getArchitectGrammarIdByName(ctx context.Context, name string) (grammarId string, retryable bool, err error) {
	return p.getArchitectGrammarIdByNameAttr(ctx, p, name)
}

// updateArchitectGrammar updates a Genesys Cloud Architect Grammar
func (p *architectGrammarProxy) updateArchitectGrammar(ctx context.Context, grammarId string, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, error) {
	return p.updateArchitectGrammarAttr(ctx, p, grammarId, grammar)
}

// deleteArchitectGrammar deletes a Genesys Cloud Architect Grammar by Id
func (p *architectGrammarProxy) deleteArchitectGrammar(ctx context.Context, grammarId string) (statusCode int, err error) {
	return p.deleteArchitectGrammarAttr(ctx, p, grammarId)
}

// deleteArchitectGrammar deletes a Genesys Cloud Architect Grammar by Id
func (p *architectGrammarProxy) uploadArchitectGrammar(ctx context.Context, grammarId string, grammar *platformclientv2.Grammar) (statusCode int, err error) {
	return p.uploadArchitectGrammarAttr(ctx, p, grammarId, grammar)
}

// createArchitectGrammarFn is an implementation function for creating a Genesys Cloud Architect Grammar
func createArchitectGrammarFn(ctx context.Context, p *architectGrammarProxy, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, error) {
	grammar, _, err := p.architectApi.PostArchitectGrammars(*grammar)
	if err != nil {
		return nil, fmt.Errorf("Failed to create grammar: %s", err)
	}

	return grammar, nil
}

// createArchitectGrammarLanguageFn is an implementation function for creating a Genesys Cloud Architect Grammarlanguage
func createArchitectGrammarLanguageFn(ctx context.Context, p *architectGrammarProxy, grammarId string, language *platformclientv2.Grammarlanguage) (*platformclientv2.Grammarlanguage, error) {
	language, _, err := p.architectApi.PostArchitectGrammarLanguages(grammarId, *language)
	if err != nil {
		return nil, fmt.Errorf("Failed to create grammar language: %s", err)
	}

	return language, nil
}

// getAllArchitectGrammarFn is the implementation for retrieving all Architect Grammars in Genesys Cloud
func getAllArchitectGrammarFn(ctx context.Context, p *architectGrammarProxy) (*[]platformclientv2.Grammar, error) {
	var allGrammars []platformclientv2.Grammar

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100

		grammars, _, err := p.architectApi.GetArchitectGrammars(pageNum, pageSize, "", "", []string{}, "", "", "", true)
		if err != nil {
			return nil, fmt.Errorf("Failed to get architect grammars: %v", err)
		}

		if grammars.Entities == nil || len(*grammars.Entities) == 0 {
			break
		}

		for _, grammar := range *grammars.Entities {
			log.Printf("Dealing with grammar id : %s", *grammar.Id)
			allGrammars = append(allGrammars, grammar)
		}
	}

	return &allGrammars, nil
}

// getArchitectGrammarByIdFn is an implementation of the function to get a Genesys Cloud Architect Grammar by Id
func getArchitectGrammarByIdFn(ctx context.Context, p *architectGrammarProxy, grammarId string) (grammar *platformclientv2.Grammar, statusCode int, err error) {
	grammar, resp, err := p.architectApi.GetArchitectGrammar(grammarId, true)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve grammar by id %s: %s", grammarId, err)
	}
	return grammar, resp.StatusCode, nil
}

// getArchitectGrammarIdBySearchFn is an implementation of the function to get a Genesys Cloud Architect Grammar by name
func getArchitectGrammarIdByNameFn(ctx context.Context, p *architectGrammarProxy, name string) (grammarId string, retryable bool, err error) {
	const pageNum = 1
	const pageSize = 100
	grammars, _, err := p.architectApi.GetArchitectGrammars(pageNum, pageSize, "", "", []string{}, name, "", "", true)
	if err != nil {
		return "", false, fmt.Errorf("Error searching architect grammar %s: %s", name, err)
	}

	if grammars.Entities == nil || len(*grammars.Entities) == 0 {
		return "", true, fmt.Errorf("No architect grammar found with name %s", name)
	}

	if len(*grammars.Entities) > 1 {
		return "", false, fmt.Errorf("Too many values returned in look for architect grammar.  Unable to choose 1 grammar.  Please refine search and continue.")
	}

	log.Printf("Retrieved the grammar id %s by name %s", *(*grammars.Entities)[0].Id, name)
	grammar := (*grammars.Entities)[0]
	return *grammar.Id, false, nil
}

// updateArchitectGrammarFn is an implementation of the function to update a Genesys Cloud Architect Grammar
func updateArchitectGrammarFn(ctx context.Context, p *architectGrammarProxy, grammarId string, grammar *platformclientv2.Grammar) (*platformclientv2.Grammar, error) {
	grammar, _, err := p.architectApi.PatchArchitectGrammar(grammarId, *grammar)
	if err != nil {
		return nil, fmt.Errorf("Failed to update grammar %s: %s", grammarId, err)
	}

	return grammar, nil
}

// deleteArchitectGrammarFn is an implementation function for deleting a Genesys Cloud Architect Grammar
func deleteArchitectGrammarFn(ctx context.Context, p *architectGrammarProxy, grammarId string) (statusCode int, err error) {
	_, resp, err := p.architectApi.DeleteArchitectGrammar(grammarId)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete grammar: %s", err)
	}

	return resp.StatusCode, nil
}

// deleteArchitectGrammarFn is an implementation function for deleting a Genesys Cloud Architect Grammar
func uploadArchitectGrammarFn(ctx context.Context, p *architectGrammarProxy, grammarId string, languageCode string, filename *string, uploadBody *platformclientv2.Grammarfileuploadrequest) error {
	uploadResponse, _, err := p.architectApi.PostArchitectGrammarLanguageFilesVoice(grammarId, languageCode, *uploadBody)
	if err != nil {
		return fmt.Errorf("Failed to get language file presignedUri: %s", err)
	}

	reader, file, err := files.DownloadOrOpenFile(*filename)
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		return err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(*filename))
	if err != nil {
		return err
	}

	if file != nil {
		io.Copy(part, file)
	} else {
		io.Copy(part, reader)
	}
	io.Copy(part, file)
	writer.Close()

	request, err := http.NewRequest(http.MethodPut, *uploadResponse.Url, body)
	if err != nil {
		return err
	}

	for key, value := range *uploadResponse.Headers {
		request.Header.Add(key, value)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	log.Printf("Content of upload: %s", content)

	return nil
}
