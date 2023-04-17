package genesyscloud

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

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
