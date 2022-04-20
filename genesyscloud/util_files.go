package genesyscloud

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
		file, err := os.Open(path)
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

	return fmt.Sprintf("%s %s", path, hex.EncodeToString(hash.Sum(nil)))
}
