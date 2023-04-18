package genesyscloud

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

type ScriptUploader struct {
	FilePath      string
	ScriptName    string
	PostUrl       string
	AccessToken   string
	Substitutions map[string]interface{}

	BodyBuf *bytes.Buffer
	Writer  *multipart.Writer

	reader io.Reader

	Client  *http.Client
	Request *http.Request
}

func NewScriptUploaderObject(filePath, scriptName, apiBasePath, accessToken string, reader io.Reader, substitutions map[string]interface{}) ScriptUploader {
	var (
		bodyBuf = bytes.Buffer{}
		w       = multipart.NewWriter(&bodyBuf)
		client  = &http.Client{}
	)
	return ScriptUploader{
		FilePath:      filePath,
		ScriptName:    scriptName,
		PostUrl:       apiBasePath + "/uploads/v2/scripter",
		AccessToken:   accessToken,
		Substitutions: substitutions,

		BodyBuf: &bodyBuf,
		Writer:  w,

		reader: reader,

		Client: client,
	}
}

func (s *ScriptUploader) Upload() ([]byte, error) {
	if err := s.createScriptFormData(); err != nil {
		return nil, err
	}

	s.substituteValues()
	s.buildHttpRequest()

	resp, err := s.Client.Do(s.Request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failure uploading script '%s': %v", s.ScriptName, resp.Status)
	}
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body when uploading file. %s", err)
	}

	return response, nil
}

func (s *ScriptUploader) buildHttpRequest() {
	r, _ := http.NewRequest(http.MethodPost, s.PostUrl, s.BodyBuf)
	r.Header.Set("Authorization", "Bearer "+s.AccessToken)
	r.Header.Set("Content-Type", s.Writer.FormDataContentType())
	s.Request = r
}

func (s *ScriptUploader) createScriptFormData() error {
	readers := map[string]io.Reader{
		"file":       s.reader,
		"scriptName": strings.NewReader(s.ScriptName),
	}

	for key, r := range readers {
		var (
			fw  io.Writer
			err error
		)
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if file, ok := r.(*os.File); ok {
			fw, err = s.Writer.CreateFormFile(key, file.Name())
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
	s.Writer.Close()
	return nil
}

func (s *ScriptUploader) substituteValues() {
	// Attribute specific to the flows resource
	if len(s.Substitutions) > 0 {
		fileContents := s.BodyBuf.String()
		for k, v := range s.Substitutions {
			fileContents = strings.Replace(fileContents, fmt.Sprintf("{{%s}}", k), v.(string), -1)
		}

		s.BodyBuf.Reset()
		s.BodyBuf.WriteString(fileContents)
	}
}

type S3Uploader struct {
	reader        io.Reader
	substitutions map[string]interface{}
	headers       map[string]string
	presignedUrl  string
	client        http.Client
}

func NewS3Uploader(reader io.Reader, substitutions map[string]interface{}, headers map[string]string, presignedUrl string) *S3Uploader {
	c := &http.Client{}
	s3Uploader := &S3Uploader{
		reader:        reader,
		substitutions: substitutions,
		headers:       headers,
		presignedUrl:  presignedUrl,
		client:        *c,
	}

	log.Printf("%#v\n", s3Uploader)
	return s3Uploader
}

func (s *S3Uploader) substituteValues(bodyBuf *bytes.Buffer) {
	// Attribute specific to the flows resource
	if len(s.substitutions) > 0 {
		fileContents := bodyBuf.String()
		for k, v := range s.substitutions {
			fileContents = strings.Replace(fileContents, fmt.Sprintf("{{%s}}", k), v.(string), -1)
		}

		bodyBuf.Reset()
		bodyBuf.WriteString(fileContents)
	}
}

func (s *S3Uploader) Upload() ([]byte, error) {
	bodyBuf := &bytes.Buffer{}

	_, err := io.Copy(bodyBuf, s.reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content to the handler. Error: %s ", err)
	}

	s.substituteValues(bodyBuf)

	req, _ := http.NewRequest("PUT", s.presignedUrl, bodyBuf)
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

func downloadOrOpenFile(path string) (io.Reader, *os.File, error) {
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
			return nil, nil, fmt.Errorf("Invalid file path or URL: %v", path)
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

// Download file from uri to directory/fileName
func downloadExportFile(directory, fileName, uri string) error {
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
	reader, file, err := downloadOrOpenFile(path)
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

	reader, file, err := downloadOrOpenFile(filename)
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
