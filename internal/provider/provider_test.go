// Copyright IBM Corp. 2026

package provider

import (
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// acctestSuffix returns a process-unique numeric suffix for generating
// collision-free resource names within a test run.
func acctestSuffix() int64 { return time.Now().UnixNano() }

// testAccProtoV6ProviderFactories wires the in-process provider under the
// "directus" name for acceptance tests. The provider reads DIRECTUS_URL and
// DIRECTUS_TOKEN from the environment, so test configs need no explicit
// provider block credentials.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"directus": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck fails fast when the live-instance credentials required by
// acceptance tests are absent. Acceptance tests only run when TF_ACC is set.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("DIRECTUS_URL") == "" {
		t.Fatal("DIRECTUS_URL must be set for acceptance tests")
	}
	if os.Getenv("DIRECTUS_TOKEN") == "" {
		t.Fatal("DIRECTUS_TOKEN must be set for acceptance tests")
	}
}
