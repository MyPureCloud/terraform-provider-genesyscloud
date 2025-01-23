package files

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	testrunner "terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/stretchr/testify/assert"
)

// TestS3UploadSuccess will simulate a successful uploading of a file to S3 using the S3 uploader.  It does not care about the actual YAML file contents and simply mocks out what gets retruned
func TestS3UploadSuccess(t *testing.T) {
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

// TestS3UploadBadRequest tests the situation where the pre-signed URL call returns a bad status code
func TestS3UploadBadRequest(t *testing.T) {
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

// TestSubstitutions will test the substitution replacement in the S3Upload
func TestSubstitutions(t *testing.T) {
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

func TestScriptUploadSuccess(t *testing.T) {
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

func TestFileContentHashChanged(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test-content-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write initial content
	initialContent := []byte("initial content")
	if err := os.WriteFile(tmpFile.Name(), initialContent, 0644); err != nil {
		t.Fatalf("Failed to write initial content: %v", err)
	}

	tests := []struct {
		name         string
		setupFunc    func() error
		expectedDiff bool
	}{
		{
			name: "content_unchanged",
			setupFunc: func() error {
				// No changes to file
				return nil
			},
			expectedDiff: false,
		},
		{
			name: "content_changed",
			setupFunc: func() error {
				return os.WriteFile(tmpFile.Name(), []byte("changed content"), 0644)
			},
			expectedDiff: true,
		},
		{
			name: "content_unchanged_again",
			setupFunc: func() error {
				// No changes to file
				return nil
			},
			expectedDiff: false,
		},
		{
			name: "content_changed_again",
			setupFunc: func() error {
				return os.WriteFile(tmpFile.Name(), []byte("changed content again"), 0644)
			},
			expectedDiff: true,
		},
		{
			name: "final_content_changed",
			setupFunc: func() error {
				return os.WriteFile(tmpFile.Name(), []byte("final changed content"), 0644)
			},
			expectedDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := testrunner.GenerateTestProvider("test_resource",
				map[string]*schema.Schema{
					"filepath": {
						Type:     schema.TypeString,
						Required: true,
					},
					"file_content_hash": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
				customdiff.ComputedIf("file_content_hash", FileContentHashChanged("filepath", "file_content_hash")),
			)

			// Pre calculate hash
			priorHash, err := HashFileContent(tmpFile.Name())
			if err != nil {
				t.Fatalf("Failed to calculate hash: %v", err)
			}

			// Run setup for this test case
			if err := tt.setupFunc(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			diff, err := testrunner.GenerateTestDiff(
				provider,
				"test_resource",
				map[string]string{
					"filepath":          tmpFile.Name(),
					"file_content_hash": priorHash,
				},
				map[string]string{
					"filepath": tmpFile.Name(),
				},
			)

			if err != nil {
				t.Fatalf("Diff failed with error: %s", err)
			}

			if tt.expectedDiff {
				if diff == nil {
					t.Error("Expected a diff when file content changes, got nil")
				} else if !diff.Attributes["file_content_hash"].NewComputed {
					t.Error("file_content_hash is not marked as NewComputed when file content changes")
				}
			} else {
				if diff != nil && diff.Attributes["file_content_hash"].NewComputed {
					t.Error("Expected no diff when file content unchanged, but file_content_hash was marked as NewComputed")
				}
			}
		})
	}
}

func TestHashFileContent(t *testing.T) {
	// Create a temporary test file
	tempContent := []byte("test content")
	tempFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up after test

	// Write content to temp file
	if err := os.WriteFile(tempFile.Name(), tempContent, 0644); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Test successful case
	t.Run("successful hash", func(t *testing.T) {
		hash, err := HashFileContent(tempFile.Name())
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if hash == "" {
			t.Error("Expected non-empty hash")
		}
		// Known hash for "test content"
		expectedHash := "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})

	// Test non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		hash, err := HashFileContent("non_existent_file.txt")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
		if hash != "" {
			t.Errorf("Expected empty hash for error case, got %s", hash)
		}
	})
}

func TestGetCSVRecordCount(t *testing.T) {
	tests := []struct {
		name          string
		fileContent   string
		expectedCount int
		expectedError bool
	}{
		{
			name:          "Valid CSV with multiple records",
			fileContent:   "header1,header2\nvalue1,value2\nvalue3,value4",
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:          "CSV with only header",
			fileContent:   "header1,header2",
			expectedCount: 0,
			expectedError: false,
		},
		{
			name:          "Empty file",
			fileContent:   "",
			expectedCount: 0,
			expectedError: false,
		},
		{
			name:          "Malformed CSV",
			fileContent:   "header1,header2\nvalue1,value2,extra",
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary test file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.csv")

			err := os.WriteFile(tmpFile, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Run the function
			count, err := GetCSVRecordCount(tmpFile)

			// Check error
			if tt.expectedError && err == nil {
				t.Error("Expected an error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check count
			if !tt.expectedError && count != tt.expectedCount {
				t.Errorf("Expected count %d, got %d", tt.expectedCount, count)
			}
		})
	}
}

func TestGetCSVRecordCount_NonexistentFile(t *testing.T) {
	_, err := GetCSVRecordCount("nonexistent.csv")
	if err == nil {
		t.Error("Expected error for nonexistent file, got none")
	}
}
