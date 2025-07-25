package files

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	utilAws "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws"
	testrunner "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

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
		_, _ = w.Write([]byte(yamlFile))
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
	fmt.Fprint(&original, origYamlFile)
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
		_, _ = w.Write([]byte(scriptFile))
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

func TestUnitDownloadOrOpenFile(t *testing.T) {
	ctx := context.Background()
	// Test HTTP download
	t.Run("successful HTTP download", func(t *testing.T) {
		// Setup test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("test content"))
		}))
		defer server.Close()

		reader, file, err := DownloadOrOpenFile(ctx, server.URL, false)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if file != nil {
			t.Error("Expected file to be nil for HTTP downloads")
		}

		// Read content
		content, err := io.ReadAll(reader)
		if err != nil {
			t.Errorf("Failed to read content: %v", err)
		}
		if string(content) != "test content" {
			t.Errorf("Expected 'test content', got '%s'", string(content))
		}
	})

	t.Run("HTTP download failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		reader, file, err := DownloadOrOpenFile(ctx, server.URL, false)
		if err == nil {
			t.Error("Expected error for 404 response, got nil")
		}
		if reader != nil || file != nil {
			t.Error("Expected nil reader and file for failed request")
		}
	})

	// Test local file operations
	t.Run("successful local file read", func(t *testing.T) {
		// Create temporary test file
		tmpfile, err := os.CreateTemp("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())

		content := []byte("local file content")
		if _, err := tmpfile.Write(content); err != nil {
			t.Fatal(err)
		}
		tmpfile.Close()

		reader, file, err := DownloadOrOpenFile(ctx, tmpfile.Name(), false)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if file == nil {
			t.Error("Expected file to not be nil for local files")
		}
		defer file.Close()

		// Read content
		readContent, err := io.ReadAll(reader)
		if err != nil {
			t.Errorf("Failed to read content: %v", err)
		}
		if string(readContent) != "local file content" {
			t.Errorf("Expected 'local file content', got '%s'", string(readContent))
		}
	})

	t.Run("non-existent local file", func(t *testing.T) {
		path := filepath.Join(os.TempDir(), "nonexistent-file")
		reader, file, err := DownloadOrOpenFile(ctx, path, false)
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
		if !strings.Contains(err.Error(), fmt.Sprintf("could not open %s: no such file", path)) {
			t.Error("Expected 'no such file or directory' error")
		}
		if reader != nil || file != nil {
			t.Error("Expected nil reader and file for non-existent file")
		}
	})

	// Test that GetS3FileReader is used when the path is an S3 URI
	t.Run("S3 util function is used when the path is an S3 URI and supportS3 is true", func(t *testing.T) {
		originalGetS3FileReader := utilAws.GetS3FileReader
		defer func() {
			utilAws.GetS3FileReader = originalGetS3FileReader
		}()
		utilAws.GetS3FileReader = func(ctx context.Context, path string) (io.Reader, *os.File, error) {
			return nil, nil, fmt.Errorf("test error")
		}

		reader, file, err := DownloadOrOpenFile(ctx, "s3://test-bucket/test-key", true)
		if err == nil {
			t.Error("Expected error for S3 URI, got nil")
		}
		if reader != nil || file != nil {
			t.Error("Expected nil reader and file for S3 URI")
		}
	})
}

func TestUnitHashFileContent(t *testing.T) {
	ctx := context.Background()
	// Create a temporary test file
	tempContent := []byte("test content")
	tempFile, err := os.CreateTemp(testrunner.GetTestDataPath(), "test_file_*.txt")
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
		hash, err := HashFileContent(ctx, tempFile.Name(), false)
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
		hash, err := HashFileContent(ctx, "non_existent_file.txt", false)
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
		if hash != "" {
			t.Errorf("Expected empty hash for error case, got %s", hash)
		}
	})
}

func TestUnitGetCSVRecordCount(t *testing.T) {
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

func TestUnitGetCSVRecordCount_NonexistentFile(t *testing.T) {
	_, err := GetCSVRecordCount("nonexistent.csv")
	if err == nil {
		t.Error("Expected error for nonexistent file, got none")
	}
}
