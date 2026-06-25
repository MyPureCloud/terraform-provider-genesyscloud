package greetingmedia

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
)

func TestBuildGreetingFileName(t *testing.T) {
	t.Parallel()

	configMap := map[string]interface{}{
		"name": "My Greeting!",
	}
	got := buildGreetingFileName(configMap, "abc-123")
	want := "My_Greeting_-abc-123.wav"
	if got != want {
		t.Fatalf("buildGreetingFileName() = %q, want %q", got, want)
	}
}

func TestGreetingAudioResolverDownloadsAndUpdatesConfig(t *testing.T) {
	t.Parallel()

	exportDir := t.TempDir()
	subDir := SubDirectory
	greetingID := "greeting-id-1"
	mediaURI := "https://example.com/greeting.wav"

	originalDownload := files.DownloadExportFile
	files.DownloadExportFile = func(directory, fileName, uri string) (*platformclientv2.APIResponse, error) {
		if uri != mediaURI {
			t.Fatalf("DownloadExportFile uri = %q, want %q", uri, mediaURI)
		}
		if err := os.MkdirAll(directory, os.ModePerm); err != nil {
			return nil, err
		}
		return nil, os.WriteFile(filepath.Join(directory, fileName), []byte("audio"), 0o644)
	}
	t.Cleanup(func() {
		files.DownloadExportFile = originalDownload
	})

	configMap := map[string]interface{}{
		"name": "Test Greeting",
	}
	resource := resourceExporter.ResourceInfo{
		State: &terraform.InstanceState{
			Attributes: map[string]string{},
		},
	}

	err := greetingAudioResolver(greetingID, exportDir, subDir, configMap, &provider.ProviderMeta{
		ClientConfig: platformclientv2.NewConfiguration(),
	}, resource, func(api *platformclientv2.GreetingsApi, id string) (*platformclientv2.Greetingmediainfo, *platformclientv2.APIResponse, error) {
		if id != greetingID {
			t.Fatalf("greeting id = %q, want %q", id, greetingID)
		}
		return &platformclientv2.Greetingmediainfo{
			MediaFileUri: platformclientv2.String(mediaURI),
		}, nil, nil
	})
	if err != nil {
		t.Fatalf("greetingAudioResolver() error = %v", err)
	}

	expectedFile := filepath.Join(subDir, "Test_Greeting-greeting-id-1.wav")
	if configMap["audio_filename"] != expectedFile {
		t.Fatalf("audio_filename = %v, want %v", configMap["audio_filename"], expectedFile)
	}
	if resource.State.Attributes["audio_filename"] != expectedFile {
		t.Fatalf("state audio_filename = %q, want %q", resource.State.Attributes["audio_filename"], expectedFile)
	}
	if configMap["audio_file_content_hash"] == "" {
		t.Fatal("expected audio_file_content_hash to be set in configMap")
	}
	if resource.State.Attributes["audio_file_content_hash"] == "" {
		t.Fatal("expected audio_file_content_hash to be set in state")
	}
}

func TestGreetingAudioResolverSkipsWhenNoMediaURI(t *testing.T) {
	t.Parallel()

	configMap := map[string]interface{}{}
	resource := resourceExporter.ResourceInfo{
		State: &terraform.InstanceState{
			Attributes: map[string]string{},
		},
	}

	err := greetingAudioResolver("greeting-id-2", t.TempDir(), SubDirectory, configMap, &provider.ProviderMeta{
		ClientConfig: platformclientv2.NewConfiguration(),
	}, resource, func(api *platformclientv2.GreetingsApi, id string) (*platformclientv2.Greetingmediainfo, *platformclientv2.APIResponse, error) {
		return &platformclientv2.Greetingmediainfo{}, nil, nil
	})
	if err != nil {
		t.Fatalf("greetingAudioResolver() error = %v", err)
	}
	if _, ok := configMap["audio_filename"]; ok {
		t.Fatal("expected audio_filename to remain unset when no media is available")
	}
}
