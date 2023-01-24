package genesyscloud

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

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
