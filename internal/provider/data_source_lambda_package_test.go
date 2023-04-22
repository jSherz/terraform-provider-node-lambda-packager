package provider_test

import (
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jSherz/terraform-provider-node-lambda-packager/internal/provider"
)

// Entrypoint file does not exist.
func TestAccPackageDataSourceCannotFindEntrypoint(t *testing.T) {
	const wdDoesNotExistConfig = `
data "node-lambda-packager_package" "test" {
  args              = ["--bundle"]
  entrypoint        = "/does/not/exist/handler/src/index.ts"
  working_directory = "/does/not/exist"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"node-lambda-packager": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config:      wdDoesNotExistConfig,
				ExpectError: regexp.MustCompile(`You specified the /does/not/exist/handler/src/index.ts entrypoint which was\s+resolved to /does/not/exist/handler/src/index.ts and we could not find the\s+file or do not have permission to view it: stat\s+/does/not/exist/handler/src/index.ts: no such file or directory`),
			},
		},
	})
}

// Unrecognized arguments.
func TestAccPackageDataSourceEsbuildBadArgs(t *testing.T) {
	const config = `
data "node-lambda-packager_package" "test" {
  args              = ["--bundle", "--lemons"]
  entrypoint        = "../../examples/src/handler/index.ts"
  working_directory = "../../examples"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"node-lambda-packager": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config:      config,
				ExpectError: regexp.MustCompile("Failed to parse esbuild options"),
			},
		},
	})
}

func TestAccPackageDataSourceEsbuildReturnedErrors(t *testing.T) {
	const config = `
data "node-lambda-packager_package" "test" {
  args              = ["--bundle", "--lemons"]
  entrypoint        = "../../examples/src/handler/index.ts"
  working_directory = "../../examples"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"node-lambda-packager": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config:      config,
				ExpectError: regexp.MustCompile("Failed to parse esbuild options"),
			},
		},
	})
}

func TestAccPackageDataSourceBuildsDeterministicZip(t *testing.T) {
	const config = `
data "node-lambda-packager_package" "test" {
  args              = ["--bundle"]
  entrypoint        = "../../examples/src/handler/index.ts"
  working_directory = "../../examples"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"node-lambda-packager": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: config,
				Check: func(state *terraform.State) error {
					hash := state.RootModule().Resources["data.node-lambda-packager_package.test"].Primary.Attributes["source_code_hash"]

					assert.Equal(t, "C6uJvTtKHacNdL1OJNa500w9pqArEHyfNlLE0n/jJ44=", hash)

					return nil
				},
			},
		},
	})
}
