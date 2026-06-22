package greetingmedia

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
)

const (
	// SubDirectory is the folder under the export directory where greeting audio files are written.
	SubDirectory = "greeting_audio"
	// DefaultFormatID is the Genesys Cloud greeting media format used when downloading audio.
	DefaultFormatID = "WAV_8000_16"
	// S3Enabled indicates greeting audio export supports S3-backed file paths.
	S3Enabled = true
)

// greetingMediaInfoFunc retrieves downloadable greeting media metadata for a greeting ID.
type greetingMediaInfoFunc func(api *platformclientv2.GreetingsApi, greetingID string) (*platformclientv2.Greetingmediainfo, *platformclientv2.APIResponse, error)

// OrganizationGreetingAudioResolver downloads organization greeting audio during export.
func OrganizationGreetingAudioResolver(greetingID, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}, resource resourceExporter.ResourceInfo) error {
	return greetingAudioResolver(greetingID, exportDirectory, subDirectory, configMap, meta, resource, func(api *platformclientv2.GreetingsApi, id string) (*platformclientv2.Greetingmediainfo, *platformclientv2.APIResponse, error) {
		return api.GetGreetingDownloads(id, DefaultFormatID)
	})
}

// GroupGreetingAudioResolver downloads group greeting audio during export.
func GroupGreetingAudioResolver(greetingID, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}, resource resourceExporter.ResourceInfo) error {
	return greetingAudioResolver(greetingID, exportDirectory, subDirectory, configMap, meta, resource, func(api *platformclientv2.GreetingsApi, id string) (*platformclientv2.Greetingmediainfo, *platformclientv2.APIResponse, error) {
		return api.GetGreetingGroupsDownloads(id, DefaultFormatID)
	})
}

func greetingAudioResolver(greetingID, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}, resource resourceExporter.ResourceInfo, getMediaInfo greetingMediaInfoFunc) error {
	_ = meta.(*provider.ProviderMeta)
	api := platformclientv2.NewGreetingsApi()

	mediaInfo, _, err := getMediaInfo(api, greetingID)
	if err != nil {
		return fmt.Errorf("failed to get greeting media info for %s: %w", greetingID, err)
	}
	if mediaInfo == nil || mediaInfo.MediaFileUri == nil || *mediaInfo.MediaFileUri == "" {
		log.Printf("No downloadable greeting audio found for %s", greetingID)
		return nil
	}

	fullPath := filepath.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}

	fileName := buildGreetingFileName(configMap, greetingID)
	exportFilename := filepath.Join(subDirectory, fileName)
	normalizedFilename := strings.ReplaceAll(exportFilename, "\\", "/")

	log.Printf("Downloading greeting audio for %s to %s", greetingID, filepath.Join(fullPath, fileName))
	if _, err := files.DownloadExportFile(fullPath, fileName, *mediaInfo.MediaFileUri); err != nil {
		return fmt.Errorf("failed to download greeting audio for %s: %w", greetingID, err)
	}

	configMap["audio_filename"] = normalizedFilename
	resource.State.Attributes["audio_filename"] = normalizedFilename

	fileContentVal := fmt.Sprintf(`${filesha256("%s")}`, exportFilename)
	configMap["audio_file_content_hash"] = fileContentVal

	ctx := context.Background()
	hash, hashErr := files.HashFileContent(ctx, filepath.Join(fullPath, fileName), S3Enabled)
	if hashErr != nil {
		log.Printf("Error calculating hash for greeting %s audio file %s: %s", greetingID, fileName, hashErr.Error())
	} else {
		resource.State.Attributes["audio_file_content_hash"] = hash
	}

	return nil
}

var fileNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func buildGreetingFileName(configMap map[string]interface{}, greetingID string) string {
	baseName := "greeting"
	if name, ok := configMap["name"].(string); ok && strings.TrimSpace(name) != "" {
		baseName = strings.TrimSpace(name)
	}
	baseName = fileNameSanitizer.ReplaceAllString(baseName, "_")
	return fmt.Sprintf("%s-%s.wav", baseName, greetingID)
}
