package consulacl_test

import (
	"fmt"
	"github.com/ashald/terraform-provider-consulacl/consulacl"
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"os"
)

const resourcePolicyBinding = `
resource "consulacl_policy_binding" "test" {
  accessor = "00000000-0000-0000-0000-000000000002"
  policy = "global-management"
}
`

func TestIntegrationPolicyBinding(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: false,
		Providers:  testProviders,
		PreCheck:   func() { testResourcePolicyBindingPreConfig(t) },
		Steps: []resource.TestStep{
			{
				Config: resourcePolicyBinding,
				Check: resource.ComposeTestCheckFunc(
					testResourcePolicyBinding("00000000-0000-0000-0000-000000000002", "global-management", true),
				),
			},
			{
				Config: "locals {}",
				Check: resource.ComposeTestCheckFunc(
					testResourcePolicyBinding("00000000-0000-0000-0000-000000000002", "global-management", false),
				),
			},
		},
	})
}

func testResourcePolicyBindingPreConfig(t *testing.T) {
	ok := false

	if v := os.Getenv("CONSUL_TOKEN"); v != "" {
		ok = true
	}
	if v := os.Getenv("CONSUL_HTTP_TOKEN"); v != "" {
		ok = true
	}
	if !ok {
		t.Fatal("Either CONSUL_TOKEN or CONSUL_HTTP_TOKEN must be set for integration tests")
	}

	rp := consulacl.Provider()

	raw := map[string]interface{}{}

	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testResourcePolicyBinding(accessor, policy string, expected bool) resource.TestCheckFunc {
	states := map[bool]string{true: "pesent", false: "absent"}
	return func(s *terraform.State) error {
		aclToken, _, err := testClient.ACL().TokenRead(accessor, nil)
		if err != nil {
			return err
		}

		found := false
		for _, policyLink := range aclToken.Policies {
			if policyLink.Name == policy {
				found = true
				break
			}
		}

		if found != expected {
			return fmt.Errorf(
				"A binding between token %q and policy %s was expected to be %s but was %s",
				accessor,
				policy,
				states[expected],
				states[expected],
			)
		}

		return nil
	}
}
