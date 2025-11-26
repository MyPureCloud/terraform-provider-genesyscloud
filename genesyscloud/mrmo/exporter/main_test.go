package exporter

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
)

func TestMrmoAccExport(t *testing.T) {
	creds, err := generateCredetials()
	if err != nil {
		t.Skipf("Failed to collect the required credentials: %s", err.Error())
	}

	t.Log("Authorizing client configuration")
	clientConfig, err := CreateClientConfig(*creds)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Creating attempt limit")
	attemptLimitId, err := createAttemptLimit(clientConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Created attempt limit '%s'", attemptLimitId)

	defer func() {
		t.Logf("Cleaning up attempt limit '%s'", attemptLimitId)
		err = deleteAttemptLimit(clientConfig, attemptLimitId)
		if err != nil {
			t.Logf("failed to cleanup attempt limit: %s", err.Error())
			return
		}
		t.Logf("Successfully deleted attempt limit")
	}()

	output, diags := Export(context.Background(), ExportInput{
		ResourceType: "genesyscloud_outbound_attempt_limit",
		EntityId:     attemptLimitId,
	}, *creds)

	if diags.HasError() {
		t.Fatalf("Expected no diagnostics errors, got: %v", diags)
	}

	if output == nil {
		t.Fatal("output is nil")
	}
	if output.ExportedResourceData == nil {
		t.Fatal("output.ExportedResourceData is nil")
	}
	if output.ExportedResourceData.Id() != attemptLimitId {
		t.Fatalf("Expected ID in ResourceData to be '%s', got '%s'", attemptLimitId, output.ExportedResourceData.Id())
	}
}

func generateCredetials() (*Credentials, error) {
	creds := &Credentials{
		ClientId:     os.Getenv("GENESYSCLOUD_OAUTHCLIENT_ID"),
		ClientSecret: os.Getenv("GENESYSCLOUD_OAUTHCLIENT_SECRET"),
		Region:       os.Getenv("GENESYSCLOUD_REGION"),
	}
	if creds.ClientId == "" || creds.ClientSecret == "" || creds.Region == "" {
		return nil, errors.New("could not generate credentials because one or more of the environment variables are not set")
	}
	return creds, nil
}

func createAttemptLimit(config *platformclientv2.Configuration) (id string, err error) {
	apiInstance := platformclientv2.NewOutboundApiWithConfig(config)

	var body = platformclientv2.Attemptlimits{
		Name:                  platformclientv2.String("mrmo test attempt limit " + uuid.NewString()),
		MaxAttemptsPerContact: platformclientv2.Int(10),
	}

	data, _, err := apiInstance.PostOutboundAttemptlimits(body)
	if err != nil {
		return "", err
	}

	return *data.Id, nil
}

func deleteAttemptLimit(config *platformclientv2.Configuration, id string) error {
	apiInstance := platformclientv2.NewOutboundApiWithConfig(config)
	_, err := apiInstance.DeleteOutboundAttemptlimit(id)
	return err
}
