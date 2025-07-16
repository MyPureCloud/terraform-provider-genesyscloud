package scripts

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	"github.com/stretchr/testify/assert"
)

// MockS3Reader is a mock implementation of an S3 file reader
type MockS3Reader struct {
	content string
	pos     int
}

func (m *MockS3Reader) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.content) {
		return 0, io.EOF
	}
	n = copy(p, []byte(m.content[m.pos:]))
	m.pos += n
	return n, nil
}

func (m *MockS3Reader) Close() error {
	return nil
}

// TestCreateScriptFormDataWithS3Reader tests that the form data creation works correctly with S3 readers
func TestCreateScriptFormDataWithS3Reader(t *testing.T) {
	// Create a mock scripts proxy
	config := &platformclientv2.Configuration{
		BasePath:      "http://localhost",
		AccessToken:   "test-token",
		DefaultHeader: make(map[string]string),
	}

	proxy := newScriptsProxy(config)

	// Test the form data creation with an S3 reader
	scriptName := "test-script"
	scriptId := ""
	s3Reader := &MockS3Reader{content: `{"test": "script content"}`}

	// Create form data manually to simulate what would happen with S3
	formData := make(map[string]io.Reader)
	formData["file"] = s3Reader
	formData["scriptName"] = strings.NewReader(scriptName)
	if scriptId != "" {
		formData["scriptIdToReplace"] = strings.NewReader(scriptId)
	}

	// Test the S3Uploader with our form data
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + proxy.accessToken

	// Create a mock HTTP server to test the upload
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a POST request to the upload endpoint
		if r.Method != "POST" || !strings.Contains(r.URL.Path, "/uploads/v2/scripter") {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Verify it's a multipart form request
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			http.Error(w, "expected multipart form data", http.StatusBadRequest)
			return
		}

		// Verify Authorization header
		if !strings.Contains(r.Header.Get("Authorization"), "Bearer test-token") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse the multipart form
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		// Verify we have the expected form fields
		if r.MultipartForm.File["file"] == nil {
			http.Error(w, "missing file field", http.StatusBadRequest)
			return
		}

		if r.MultipartForm.Value["scriptName"] == nil {
			http.Error(w, "missing scriptName field", http.StatusBadRequest)
			return
		}

		// Verify the file field is actually a file (not a form field)
		fileHeaders := r.MultipartForm.File["file"]
		if len(fileHeaders) == 0 {
			http.Error(w, "file field is empty", http.StatusBadRequest)
			return
		}

		// Check that the file has a filename (indicating it's a form file, not a form field)
		if fileHeaders[0].Filename == "" {
			http.Error(w, "file field missing filename", http.StatusBadRequest)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"correlationId": "test-correlation-id"}`))
	}))
	defer mockServer.Close()

	// Update the proxy to use our mock server
	proxy.basePath = mockServer.URL

	// Create the S3Uploader with our form data
	s3Uploader := files.NewS3Uploader(nil, formData, nil, headers, "POST", mockServer.URL+"/uploads/v2/scripter")

	// Test the upload
	resp, err := s3Uploader.Upload()

	// Verify the results
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, string(resp), "test-correlation-id")
}

// TestS3UploaderWithS3Reader tests the S3Uploader with an S3 reader to ensure it creates form files correctly
func TestS3UploaderWithS3Reader(t *testing.T) {
	// Create form data with an S3 reader
	formData := make(map[string]io.Reader)
	formData["file"] = &MockS3Reader{content: `{"test": "script content"}`}
	formData["scriptName"] = strings.NewReader("test-script")

	// Create S3Uploader
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer test-token"

	// Create a mock HTTP server to test the upload
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a multipart form request
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			http.Error(w, "expected multipart form data", http.StatusBadRequest)
			return
		}

		// Parse the multipart form
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		// Verify we have the expected form fields
		if r.MultipartForm.File["file"] == nil {
			http.Error(w, "missing file field", http.StatusBadRequest)
			return
		}

		if r.MultipartForm.Value["scriptName"] == nil {
			http.Error(w, "missing scriptName field", http.StatusBadRequest)
			return
		}

		// Verify the file field is actually a file (not a form field)
		fileHeaders := r.MultipartForm.File["file"]
		if len(fileHeaders) == 0 {
			http.Error(w, "file field is empty", http.StatusBadRequest)
			return
		}

		// Check that the file has a filename (indicating it's a form file, not a form field)
		if fileHeaders[0].Filename == "" {
			http.Error(w, "file field missing filename", http.StatusBadRequest)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"correlationId": "test-correlation-id"}`))
	}))
	defer mockServer.Close()

	s3Uploader := files.NewS3Uploader(nil, formData, nil, headers, "POST", mockServer.URL)

	// Test the upload
	resp, err := s3Uploader.Upload()

	// Verify the results
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, string(resp), "test-correlation-id")
}

// TestS3UploaderWithLocalFile tests the S3Uploader with a local file for comparison
func TestS3UploaderWithLocalFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test-script-*.json")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write some content to the temp file
	content := `{"test": "local script content"}`
	_, err = tempFile.WriteString(content)
	assert.NoError(t, err)
	tempFile.Close()

	// Reopen the file for reading
	file, err := os.Open(tempFile.Name())
	assert.NoError(t, err)
	defer file.Close()

	// Create form data with the local file
	formData := make(map[string]io.Reader)
	formData["file"] = file
	formData["scriptName"] = strings.NewReader("test-script")

	// Create S3Uploader
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer test-token"

	// Create a mock HTTP server to test the upload
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's a multipart form request
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			http.Error(w, "expected multipart form data", http.StatusBadRequest)
			return
		}

		// Parse the multipart form
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		// Verify we have the expected form fields
		if r.MultipartForm.File["file"] == nil {
			http.Error(w, "missing file field", http.StatusBadRequest)
			return
		}

		if r.MultipartForm.Value["scriptName"] == nil {
			http.Error(w, "missing scriptName field", http.StatusBadRequest)
			return
		}

		// Verify the file field is actually a file (not a form field)
		fileHeaders := r.MultipartForm.File["file"]
		if len(fileHeaders) == 0 {
			http.Error(w, "file field is empty", http.StatusBadRequest)
			return
		}

		// Check that the file has a filename (indicating it's a form file, not a form field)
		if fileHeaders[0].Filename == "" {
			http.Error(w, "file field missing filename", http.StatusBadRequest)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"correlationId": "test-correlation-id"}`))
	}))
	defer mockServer.Close()

	s3Uploader := files.NewS3Uploader(nil, formData, nil, headers, "POST", mockServer.URL)

	// Test the upload
	resp, err := s3Uploader.Upload()

	// Verify the results
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, string(resp), "test-correlation-id")
}
