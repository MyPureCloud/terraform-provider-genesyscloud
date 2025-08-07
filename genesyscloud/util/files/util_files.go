package files

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	utilAws "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

type S3Uploader struct {
	reader        io.Reader
	formData      map[string]io.Reader
	bodyBuf       *bytes.Buffer
	Writer        *multipart.Writer
	substitutions map[string]interface{}
	headers       map[string]string
	httpMethod    string
	presignedUrl  string
	client        http.Client

	UploadFunc            func(s *S3Uploader) ([]byte, error)
	UploadWithRetriesFunc func(ctx context.Context, s *S3Uploader, filePath string, timeout time.Duration) ([]byte, error)
}

func NewS3Uploader(reader io.Reader, formData map[string]io.Reader, substitutions map[string]interface{}, headers map[string]string, method, presignedUrl string) *S3Uploader {
	c := &http.Client{}
	var bodyBuf bytes.Buffer
	writer := multipart.NewWriter(&bodyBuf)
	s3Uploader := &S3Uploader{
		reader:        reader,
		formData:      formData,
		bodyBuf:       &bodyBuf,
		Writer:        writer,
		substitutions: substitutions,
		headers:       headers,
		httpMethod:    method,
		presignedUrl:  presignedUrl,
		client:        *c,

		UploadFunc:            UploadFn,
		UploadWithRetriesFunc: UploadWithRetriesFn,
	}
	return s3Uploader
}

func (s *S3Uploader) substituteValues() {
	// Attribute specific to the flows resource
	if len(s.substitutions) > 0 {
		fileContents := s.bodyBuf.String()
		for k, v := range s.substitutions {
			fileContents = strings.Replace(fileContents, fmt.Sprintf("{{%s}}", k), v.(string), -1)
		}

		s.bodyBuf.Reset()
		s.bodyBuf.WriteString(fileContents)
	}
}

func (s *S3Uploader) Upload() ([]byte, error) {
	return s.UploadFunc(s)
}

func (s *S3Uploader) UploadWithRetries(ctx context.Context, filePath string, timeout time.Duration) ([]byte, error) {
	return s.UploadWithRetriesFunc(ctx, s, filePath, timeout)
}

func UploadFn(s *S3Uploader) ([]byte, error) {
	if s.formData != nil {
		if err := s.createFormData(); err != nil {
			return nil, err
		}
		s.headers["Content-Type"] = s.Writer.FormDataContentType()
	} else {
		_, err := io.Copy(s.bodyBuf, s.reader)
		if err != nil {
			return nil, fmt.Errorf("failed to copy file content to the handler. Error: %s ", err)
		}
	}

	s.substituteValues()

	req, _ := http.NewRequest(s.httpMethod, s.presignedUrl, s.bodyBuf)
	for key, value := range s.headers {
		req.Header.Set(key, value)
	}

	resp, err := s.client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3 bucket with an error. Error: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to upload file to S3 bucket with an HTTP status code of %d", resp.StatusCode)
	}

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body when uploading file. %s", err)
	}

	return response, nil
}

func UploadWithRetriesFn(ctx context.Context, s *S3Uploader, filePath string, timeout time.Duration) ([]byte, error) {
	var response []byte

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("failed to read file information. Path: '%s' Error: %v", filePath, err)
	}

	uploadErr := util.WithRetries(ctx, timeout, func() *retry.RetryError {
		uploadStartTime := time.Now()
		response, err = s.Upload()
		if err != nil {
			uploadDuration := time.Since(uploadStartTime)
			log.Printf("failed to upload file %s after %d milliseconds (%v seconds). Error: %v", filePath, uploadDuration.Milliseconds(), uploadDuration.Seconds(), err)
			if fileInfo != nil {
				log.Printf("size of file '%s': %v bytes", filePath, fileInfo.Size())
			}
			return retry.RetryableError(err)
		}
		return nil
	})
	if uploadErr != nil {
		return nil, fmt.Errorf("%v", uploadErr)
	}
	return response, nil
}

func (s *S3Uploader) createFormData() error {
	defer s.Writer.Close()
	for key, r := range s.formData {
		var (
			fw  io.Writer
			err error
		)
		if r == nil {
			continue
		}
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}

		// For the "file" field, always create a form file, even for non-file readers (like S3)
		if key == "file" {
			// Try to get filename from the reader if it's a file
			filename := "script.json" // default filename
			if file, ok := r.(*os.File); ok {
				filename = file.Name()
			}
			fw, err = s.Writer.CreateFormFile(key, filename)
		} else {
			fw, err = s.Writer.CreateFormField(key)
		}
		if err != nil {
			return err
		}
		if _, err := io.Copy(fw, r); err != nil {
			return err
		}
	}
	return nil
}

// DownloadOrOpenFile is a function that downloads or opens a file from a given path.
// Note: supportS3 lets us know if the resource is prepared to handle S3 paths (e.g. architect_flow). Once all resources support S3 paths, we can remove this parameter.
func DownloadOrOpenFile(ctx context.Context, path string, supportS3 bool) (io.Reader, *os.File, error) {
	var reader io.Reader
	var file *os.File

	// Check if the path is an S3 URI
	if utilAws.IsS3Path(path) && supportS3 {
		return utilAws.GetS3FileReader(ctx, path)
	}

	// Check if the path has a protocol scheme to call as an HTTP request
	if u, err := url.ParseRequestURI(path); err == nil && u.Scheme != "" {
		resp, err := http.Get(path)
		if err != nil {
			return nil, nil, err
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return nil, nil, fmt.Errorf("HTTP Error downloading file: %v", resp.StatusCode)
		}
		reader = resp.Body
	} else {
		file, err = os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil, fmt.Errorf("could not %w", err)
			}
			return nil, nil, fmt.Errorf("error opening local file \"%s\": %v", path, err)
		}
		reader = file
	}

	return reader, file, nil
}

// DownloadExportFile is a variable that holds the function for downloading export files.
// By default it points to downloadExportFile, but can be replaced with a mock implementation
// during testing. This pattern enables unit testing of code that depends on file downloads
// without actually downloading files.
var DownloadExportFile = downloadExportFile

func downloadExportFile(directory, fileName, uri string) (*platformclientv2.APIResponse, error) {
	return downloadExportFileWithAccessToken(directory, fileName, uri, "")
}

var DownloadExportFileWithAccessToken = downloadExportFileWithAccessToken

func downloadExportFileWithAccessToken(directory, fileName, uri, accessToken string) (*platformclientv2.APIResponse, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	if accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	apiResp, apiErr := platformclientv2.NewAPIResponse(resp, nil)

	if apiErr != nil {
		return apiResp, apiErr
	}
	defer resp.Body.Close()

	if err := os.MkdirAll(directory, 0755); err != nil {
		return apiResp, fmt.Errorf("failed to create directory: %w", err)
	}

	out, err := os.Create(filepath.Join(directory, fileName))
	if err != nil {
		return apiResp, err
	}
	defer func(out *os.File) {
		if err := out.Close(); err != nil {
			log.Printf("failed to close file: %s", err.Error())
		}
	}(out)

	_, err = io.Copy(out, resp.Body)
	return apiResp, err
}

// HashFileContent Hash file content, used in stateFunc for "filepath" type attributes
// Note: supportS3 lets us know if the resource is prepared to handle S3 paths (e.g. architect_flow). Once all resources support S3 paths, we can remove this parameter.
func HashFileContent(ctx context.Context, path string, supportS3 bool) (string, error) {
	reader, file, err := DownloadOrOpenFile(ctx, path, supportS3)
	if err != nil {
		return "", fmt.Errorf("unable to open file: %v", err.Error())
	}
	if file != nil {
		defer file.Close()
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", fmt.Errorf("unable to copy file content: %v", err.Error())
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func WriteToFile(bytes []byte, path string) diag.Diagnostics {
	err := os.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		return util.BuildDiagnosticError("File Writer", fmt.Sprintf("Error writing file with Path %s", path), err)
	}
	return nil
}

// getCSVRecordCount retrieves the number of records in a CSV file (i.e., number of lines in a file minus the header)
func GetCSVRecordCount(filepath string) (int, error) {
	// Open file up and read the record count
	reader, file, err := DownloadOrOpenFile(context.Background(), filepath, true)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Count the number of records in the CSV file
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = 0
	recordCount := 0
	for {
		_, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		recordCount++
	}

	// Subtract 1 to account for header row
	if recordCount > 0 {
		recordCount--
	}

	return recordCount, nil
}

// Get a string path to the target export directory
func GetDirPath(directory string) (string, diag.Diagnostics) {
	if strings.HasPrefix(directory, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", diag.Errorf("Failed to evaluate home directory: %v", err)
		}
		directory = strings.Replace(directory, "~", homeDir, 1)
	}
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return "", diag.FromErr(err)
	}

	return directory, nil
}
