package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider(t *testing.T) {
	if err := New("0.1.0", make(map[string]*schema.Resource), make(map[string]*schema.Resource))().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
