package files

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"
)

type S3Uploader struct {
	reader        io.Reader
	formData      map[string]io.Reader
	bodyBuf       *bytes.Buffer
	writer        *multipart.Writer
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
		writer:        writer,
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
	if s.substitutions != nil && len(s.substitutions) > 0 {
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
	if s.formData != nil && len(s.formData) > 0 {
		if err := s.createFormData(); err != nil {
			return nil, err
		}
		s.headers["Content-Type"] = s.writer.FormDataContentType()
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

	response, err := ioutil.ReadAll(resp.Body)
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
	defer s.writer.Close()
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
		if file, ok := r.(*os.File); ok {
			fw, err = s.writer.CreateFormFile(key, file.Name())
		} else {
			fw, err = s.writer.CreateFormField(key)
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

func DownloadOrOpenFile(path string) (io.Reader, *os.File, error) {
	var reader io.Reader
	var file *os.File

	_, err := os.Stat(path)
	if err != nil {
		_, err = url.ParseRequestURI(path)
		if err == nil {
			resp, err := http.Get(path)
			if err != nil {
				return nil, nil, err
			}
			if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
				return nil, nil, fmt.Errorf("HTTP Error downloading file: %v", resp.StatusCode)
			}
			reader = resp.Body
		} else {
			return nil, nil, fmt.Errorf("invalid file path or URL: %v", path)
		}
	} else {
		file, err = os.Open(path)
		if err != nil {
			return nil, nil, err
		}
		reader = file
	}

	return reader, file, nil
}

// DownloadExportFile Download file from uri to directory/fileName
func DownloadExportFile(directory, fileName, uri string) error {
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	out, err := os.Create(path.Join(directory, fileName))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// Hash file content, used in stateFunc for "filepath" type attributes
func hashFileContent(path string) string {
	reader, file, err := DownloadOrOpenFile(path)
	if err != nil {
		return err.Error()
	}
	if file != nil {
		defer file.Close()
	}

	hash := sha256.New()
	if file == nil {
		if _, err := io.Copy(hash, reader); err != nil {
			return err.Error()
		}
	} else {
		if _, err := io.Copy(hash, file); err != nil {
			return err.Error()
		}
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// Read and upload input file path to S3 pre-signed URL
func prepareAndUploadFile(filename string, substitutions map[string]interface{}, headers map[string]string, presignedUrl string) ([]byte, error) {
	bodyBuf := &bytes.Buffer{}

	reader, file, err := DownloadOrOpenFile(filename)
	if err != nil {
		return nil, err
	}
	if file != nil {
		defer file.Close()
	}

	_, err = io.Copy(bodyBuf, reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to copy file content to the handler. Error: %s ", err)
	}

	// Attribute specific to the flows resource
	if len(substitutions) > 0 {
		fileContents := bodyBuf.String()
		for k, v := range substitutions {
			fileContents = strings.Replace(fileContents, fmt.Sprintf("{{%s}}", k), v.(string), -1)
		}

		bodyBuf.Reset()
		bodyBuf.WriteString(fileContents)
	}

	req, _ := http.NewRequest("PUT", presignedUrl, bodyBuf)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to upload file to S3 bucket. Error: %s ", err)
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body when uploading file. %s", err)
	}

	return response, nil
}

func WriteToFile(bytes []byte, path string) diag.Diagnostics {
	err := os.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		return util.BuildDiagnosticError("File Writer", fmt.Sprintf("Error writing file with Path %s", path), err)
	}
	return nil
}
