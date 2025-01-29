package files

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUnitS3UploadSuccess will simulate a successful uploading of a file to S3 using the S3 uploader.  It does not care about the actual YAML file contents and simply mocks out what gets retruned
func TestUnitS3UploadSuccess(t *testing.T) {
	///Setting up variables
	presignedURL := "/s3/presigned"
	substitutions := make(map[string]interface{})
	headers := make(map[string]string)
	yamlFile := `inboundCall:
					name: SimpleFinancialIvr
					description: SimpleFinancialIvr
					division: Home0349a372-0480-4879-b98e-8aace58b1bb7
					startUpRef: "/inboundCall/menus/menu[Main Menu_10]"
					defaultLanguage: en-us
					supportedLanguages:
						en-us:
							defaultLanguageSkill:
							noValue: true
							textToSpeech:
							defaultEngine:
								voice: Jill`
	fileReader := strings.NewReader(yamlFile)

	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that the request method and URL path match
		if r.Method != "PUT" || r.URL.Path != presignedURL {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Return a mock JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(yamlFile))
	}))
	defer mockServer.Close()

	// Replace the client's transport with the mock server's transport
	s3Uploader := NewS3Uploader(fileReader, nil, substitutions, headers, "PUT", fmt.Sprintf("%s%s", mockServer.URL, presignedURL))
	results, err := s3Uploader.Upload()

	if err != nil {
		t.Fatal(err)
	}

	r := string(results)

	if r != yamlFile {
		t.Errorf(`expected %s got %s`, r, yamlFile)
	}
}

// TestUnitS3UploadBadRequest tests the situation where the pre-signed URL call returns a bad status code
func TestUnitS3UploadBadRequest(t *testing.T) {
	//Setting up variables
	presignedURL := "/s3/presigned"
	substitutions := make(map[string]interface{})
	headers := make(map[string]string)
	yamlFile := `inboundCall:
					name: SimpleFinancialIvr
					description: SimpleFinancialIvr
					division: Home0349a372-0480-4879-b98e-8aace58b1bb7
					startUpRef: "/inboundCall/menus/menu[Main Menu_10]"
					defaultLanguage: en-us
					supportedLanguages:
						en-us:
							defaultLanguageSkill:
							noValue: true
							textToSpeech:
							defaultEngine:
								voice: Jill`
	fileReader := strings.NewReader(yamlFile)

	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a mock JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer mockServer.Close()

	// Replace the client's transport with the mock server's transport
	s3Uploader := NewS3Uploader(fileReader, nil, substitutions, headers, "PUT", fmt.Sprintf("%s%s", mockServer.URL, presignedURL))
	_, err := s3Uploader.Upload()

	expectedResult := fmt.Sprintf("failed to upload file to S3 bucket with an HTTP status code of %d", http.StatusBadRequest)
	if err != nil {
		assert.Equal(t, err.Error(), expectedResult)
	}
}

// TestUnitSubstitutions will test the substitution replacement in the S3Upload
func TestUnitSubstitutions(t *testing.T) {
	///Setting up variables
	presignedURL := "/s3/presigned"
	substitutions := make(map[string]interface{})
	substitutions["name"] = "SimpleFinancialIvr"
	headers := make(map[string]string)
	origYamlFile := `inboundCall:
						name: {{name}}`
	expcYamlFile := `inboundCall:
						name: SimpleFinancialIvr`
	fileReader := strings.NewReader(origYamlFile)

	s3Uploader := NewS3Uploader(fileReader, nil, substitutions, headers, "PUT", fmt.Sprintf("%s%s", "", presignedURL))

	var original bytes.Buffer
	fmt.Fprintf(&original, origYamlFile)
	s3Uploader.bodyBuf = &original
	s3Uploader.substituteValues()

	assert.Equal(t, string(expcYamlFile), original.String())

}

func TestUnitScriptUploadSuccess(t *testing.T) {
	var (
		urlPath     = "/uploads/v2/scripter"
		scriptName  = "testScript"
		accessToken = "1234abcd"
		scriptFile  = `
		{
			"id": "123",
			"name": "test"
		}
		`
	)

	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that the request method and URL path match
		if r.Method != "POST" || r.URL.Path != urlPath {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if _, ok := r.Header["Content-Type"]; !ok {
			t.Errorf("Expected Content-Type header to be set in http request")
		}
		if !strings.Contains(r.Header["Content-Type"][0], "multipart/form-data") {
			http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
		}

		if _, ok := r.Header["Authorization"]; !ok {
			t.Errorf("Expected Authorization header to be set in http request")
		}
		if !strings.Contains(r.Header["Authorization"][0], accessToken) {
			http.Error(w, "unauthorizied", http.StatusUnauthorized)
		}

		buf := new(strings.Builder)
		_, err := io.Copy(buf, r.Body)
		if err != nil {
			t.Errorf("%v", err)
		}

		requestBodyItems := []string{scriptName, "form-data"}
		for _, v := range requestBodyItems {
			if !strings.Contains(buf.String(), v) {
				t.Errorf("Expected to find %s in request body", v)
			}
		}

		// Return a mock JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(scriptFile))
	}))

	defer mockServer.Close()

	formData := make(map[string]io.Reader, 0)
	formData["file"] = strings.NewReader("test.json")
	formData["scriptName"] = strings.NewReader(scriptName)

	headers := make(map[string]string, 0)
	headers["Authorization"] = "Bearer " + accessToken

	s3Uploader := NewS3Uploader(nil, formData, nil, headers, "POST", mockServer.URL+urlPath)

	results, err := s3Uploader.Upload()
	if err != nil {
		t.Fatal(err)
	}

	resultsStr := string(results)
	if resultsStr != scriptFile {
		t.Errorf(`expected %s got %s`, scriptFile, resultsStr)
	}
}
