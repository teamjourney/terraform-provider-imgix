package imgix

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"testing"
)

var testProviders map[string]func() (*schema.Provider, error)
var testProvider *schema.Provider

func init() {
	testProvider = Provider()
	testProviders = map[string]func() (*schema.Provider, error){
		"imgix": func() (*schema.Provider, error) {
			return testProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderImpl(t *testing.T) {
	var _ *schema.Provider = Provider()
}
