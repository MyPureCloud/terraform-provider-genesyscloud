package architect_grammar_language

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

type FileType int

const (
	Dtmf FileType = iota
	Voice
)

/*
The genesyscloud_architect_grammar_language_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *architectGrammarLanguageProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createArchitectGrammarLanguageFunc func(ctx context.Context, p *architectGrammarLanguageProxy, language *platformclientv2.Grammarlanguage) (*platformclientv2.Grammarlanguage, error)
type getArchitectGrammarLanguageByIdFunc func(ctx context.Context, p *architectGrammarLanguageProxy, grammarId string, languageCode string) (language *platformclientv2.Grammarlanguage, responseCode int, err error)
type updateArchitectGrammarLanguageFunc func(ctx context.Context, p *architectGrammarLanguageProxy, grammarId string, languageCode string, language *platformclientv2.Grammarlanguage) (*platformclientv2.Grammarlanguage, error)
type deleteArchitectGrammarLanguageFunc func(ctx context.Context, p *architectGrammarLanguageProxy, grammarId string, languageCode string) (responseCode int, err error)
type getAllArchitectGrammarLanguageFunc func(ctx context.Context, p *architectGrammarLanguageProxy) (*[]platformclientv2.Grammarlanguage, error)

// architectGrammarLanguageProxy contains all of the methods that call genesys cloud APIs.
type architectGrammarLanguageProxy struct {
	clientConfig                        *platformclientv2.Configuration
	architectApi                        *platformclientv2.ArchitectApi
	createArchitectGrammarLanguageAttr  createArchitectGrammarLanguageFunc
	getArchitectGrammarLanguageByIdAttr getArchitectGrammarLanguageByIdFunc
	updateArchitectGrammarLanguageAttr  updateArchitectGrammarLanguageFunc
	deleteArchitectGrammarLanguageAttr  deleteArchitectGrammarLanguageFunc
	getAllArchitectGrammarLanguageAttr  getAllArchitectGrammarLanguageFunc
}

// newArchitectGrammarLanguageProxy initializes the grammar Language proxy with all of the data needed to communicate with Genesys Cloud
func newArchitectGrammarLanguageProxy(clientConfig *platformclientv2.Configuration) *architectGrammarLanguageProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &architectGrammarLanguageProxy{
		clientConfig:                        clientConfig,
		architectApi:                        api,
		createArchitectGrammarLanguageAttr:  createArchitectGrammarLanguageFn,
		getArchitectGrammarLanguageByIdAttr: getArchitectGrammarLanguageByIdFn,
		updateArchitectGrammarLanguageAttr:  updateArchitectGrammarLanguageFn,
		deleteArchitectGrammarLanguageAttr:  deleteArchitectGrammarLanguageFn,
		getAllArchitectGrammarLanguageAttr:  getAllArchitectGrammarLanguageFn,
	}
}

// getArchitectGrammarLanguageProxy acts as a singleton for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getArchitectGrammarLanguageProxy(clientConfig *platformclientv2.Configuration) *architectGrammarLanguageProxy {
	if internalProxy == nil {
		internalProxy = newArchitectGrammarLanguageProxy(clientConfig)
	}

	return internalProxy
}

// createArchitectGrammarLanguage creates a Genesys Cloud Architect Grammar Language
func (p *architectGrammarLanguageProxy) createArchitectGrammarLanguage(ctx context.Context, language *platformclientv2.Grammarlanguage) (*platformclientv2.Grammarlanguage, error) {
	return p.createArchitectGrammarLanguageAttr(ctx, p, language)
}

// getArchitectGrammarLanguageById returns a single Genesys Cloud Architect Grammar Language by Id
func (p *architectGrammarLanguageProxy) getArchitectGrammarLanguageById(ctx context.Context, grammarId string, languageCode string) (language *platformclientv2.Grammarlanguage, statusCode int, err error) {
	return p.getArchitectGrammarLanguageByIdAttr(ctx, p, grammarId, languageCode)
}

// updateArchitectGrammarLanguage updates a Genesys Cloud Architect Grammar Language
func (p *architectGrammarLanguageProxy) updateArchitectGrammarLanguage(ctx context.Context, grammarId string, languageCode string, language *platformclientv2.Grammarlanguage) (*platformclientv2.Grammarlanguage, error) {
	return p.updateArchitectGrammarLanguageAttr(ctx, p, grammarId, languageCode, language)
}

// deleteArchitectGrammarLanguage deletes a Genesys Cloud Architect Grammar Language by Id
func (p *architectGrammarLanguageProxy) deleteArchitectGrammarLanguage(ctx context.Context, grammarId string, languageCode string) (statusCode int, err error) {
	return p.deleteArchitectGrammarLanguageAttr(ctx, p, grammarId, languageCode)
}

// getAllArchitectGrammarLanguage retrieves all Genesys Cloud Architect Grammar Languages
func (p *architectGrammarLanguageProxy) getAllArchitectGrammarLanguage(ctx context.Context) (*[]platformclientv2.Grammarlanguage, error) {
	return p.getAllArchitectGrammarLanguageAttr(ctx, p)
}

// createArchitectGrammarLanguageFn is an implementation function for creating a Genesys Cloud Architect Grammar Language
func createArchitectGrammarLanguageFn(_ context.Context, p *architectGrammarLanguageProxy, language *platformclientv2.Grammarlanguage) (*platformclientv2.Grammarlanguage, error) {
	languageSdk, _, err := p.architectApi.PostArchitectGrammarLanguages(*language.GrammarId, *language)
	if err != nil {
		return nil, err
	}

	// Upload grammar voice file
	if language.VoiceFileMetadata != nil && language.VoiceFileMetadata.FileName != nil {
		if err = uploadGrammarLanguageFile(p, language, *language.VoiceFileMetadata.FileName, Voice); err != nil {
			return nil, fmt.Errorf("failed to upload language voice file for grammar '%s': %s", *language.GrammarId, err)
		}
	}

	// Upload grammar dtmf file
	if language.DtmfFileMetadata != nil && language.DtmfFileMetadata.FileName != nil {
		if err := uploadGrammarLanguageFile(p, language, *language.DtmfFileMetadata.FileName, Dtmf); err != nil {
			return nil, fmt.Errorf("failed to upload language dtmf file for grammar '%s': %s", *language.GrammarId, err)
		}
	}

	return languageSdk, nil
}

// getArchitectGrammarLanguageByIdFn is an implementation of the function to get a Genesys Cloud Architect Grammar Language by Id
func getArchitectGrammarLanguageByIdFn(_ context.Context, p *architectGrammarLanguageProxy, grammarId string, languageCode string) (language *platformclientv2.Grammarlanguage, statusCode int, err error) {
	language, resp, err := p.architectApi.GetArchitectGrammarLanguage(grammarId, languageCode)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return language, resp.StatusCode, nil
}

// updateArchitectGrammarLanguageFn is an implementation of the function to update a Genesys Cloud Architect Grammar Language
func updateArchitectGrammarLanguageFn(_ context.Context, p *architectGrammarLanguageProxy, grammarId string, languageCode string, language *platformclientv2.Grammarlanguage) (*platformclientv2.Grammarlanguage, error) {
	languageUpdate := platformclientv2.Grammarlanguageupdate{
		VoiceFileMetadata: language.VoiceFileMetadata,
		DtmfFileMetadata:  language.DtmfFileMetadata,
	}

	languageSdk, _, err := p.architectApi.PatchArchitectGrammarLanguage(grammarId, languageCode, languageUpdate)
	if err != nil {
		return nil, err
	}

	// Upload grammar voice file
	if language.VoiceFileMetadata != nil && language.VoiceFileMetadata.FileName != nil {
		if err := uploadGrammarLanguageFile(p, language, *language.VoiceFileMetadata.FileName, Voice); err != nil {
			return nil, fmt.Errorf("failed to upload language voice file for grammar '%s': %s", *language.GrammarId, err)
		}
	}

	// Upload grammar dtmf file
	if language.DtmfFileMetadata != nil && language.DtmfFileMetadata.FileName != nil {
		if err := uploadGrammarLanguageFile(p, language, *language.DtmfFileMetadata.FileName, Dtmf); err != nil {
			return nil, fmt.Errorf("failed to upload language dtmf file for grammar '%s': %s", *language.GrammarId, err)
		}
	}

	return languageSdk, nil
}

// deleteArchitectGrammarLanguageFn is an implementation function for deleting a Genesys Cloud Architect Grammar Language
func deleteArchitectGrammarLanguageFn(_ context.Context, p *architectGrammarLanguageProxy, grammarId string, languageCode string) (statusCode int, err error) {
	resp, err := p.architectApi.DeleteArchitectGrammarLanguage(grammarId, languageCode)
	return resp.StatusCode, err
}

// uploadGrammarLanguageFile is a function for uploading a grammar language file to Genesys cloud
func uploadGrammarLanguageFile(p *architectGrammarLanguageProxy, language *platformclientv2.Grammarlanguage, filePath string, fileType FileType) error {
	var (
		uploadResponse *platformclientv2.Uploadurlresponse
		err            error
		grammarId      = *language.GrammarId
		languageCode   = *language.Language

		uploadBody platformclientv2.Grammarfileuploadrequest
	)
	if fileType == Voice {
		uploadBody.FileType = language.VoiceFileMetadata.FileType
		uploadResponse, _, err = p.architectApi.PostArchitectGrammarLanguageFilesVoice(grammarId, languageCode, uploadBody)
	}
	if fileType == Dtmf {
		uploadBody.FileType = language.DtmfFileMetadata.FileType
		uploadResponse, _, err = p.architectApi.PostArchitectGrammarLanguageFilesDtmf(grammarId, languageCode, uploadBody)
	}
	if err != nil {
		return fmt.Errorf("failed to get language file presignedUri: %s for file %s", err, filePath)
	}

	reader, _, err := files.DownloadOrOpenFile(filePath)
	if err != nil {
		return fmt.Errorf("error downloading file '%s': %v", filePath, err)
	}

	s3Uploader := files.NewS3Uploader(reader, nil, nil, *uploadResponse.Headers, http.MethodPut, *uploadResponse.Url)

	if _, uploadErr := s3Uploader.UploadWithRetries(context.Background(), filePath, 20*time.Second); uploadErr != nil {
		return fmt.Errorf("failed to upload language file for grammar '%s': %v", *language.GrammarId, uploadErr)
	}

	return nil
}

// getAllArchitectGrammarLanguageFn is the implementation for retrieving all Architect Grammars in Genesys Cloud
func getAllArchitectGrammarLanguageFn(_ context.Context, p *architectGrammarLanguageProxy) (*[]platformclientv2.Grammarlanguage, error) {
	var allLanguages []platformclientv2.Grammarlanguage

	grammars, _, err := p.architectApi.GetArchitectGrammars(1, 100, "", "", []string{}, "", "", "", true)
	if err != nil {
		return nil, fmt.Errorf("failed to get architect grammar languages: %v", err)
	}
	if grammars.Entities == nil || len(*grammars.Entities) == 0 {
		return &allLanguages, nil
	}

	for _, grammar := range *grammars.Entities {
		if grammar.Languages != nil {
			for _, language := range *grammar.Languages {
				allLanguages = append(allLanguages, language)
			}
		}
	}

	for pageNum := 2; pageNum <= *grammars.PageCount; pageNum++ {
		const pageSize = 100

		grammars, _, err := p.architectApi.GetArchitectGrammars(pageNum, pageSize, "", "", []string{}, "", "", "", true)
		if err != nil {
			return nil, fmt.Errorf("failed to get architect grammar languages: %v", err)
		}

		if grammars.Entities == nil || len(*grammars.Entities) == 0 {
			break
		}

		for _, grammar := range *grammars.Entities {
			if grammar.Languages != nil {
				for _, language := range *grammar.Languages {
					allLanguages = append(allLanguages, language)
				}
			}
		}
	}

	return &allLanguages, nil
}
