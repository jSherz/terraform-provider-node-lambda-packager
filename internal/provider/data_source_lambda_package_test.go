package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPackageDataSourceGettingAbsolutePathFailed() {
}

// Entrypoint file does not exist
func TestAccPackageDataSourceCannotFindEntrypoint() {
}

// Permission denied accessing entrypoint file
func TestAccPackageDataSourceCannotAccessEntrypoint() {
}

// Unrecognized arguments
func TestAccPackageDataSourceEsbuildBadArgs() {
}

func TestAccPackageDataSourceGettingWorkingDirectoryAbsolutePathFailed() {
}

func TestAccPackageDataSourceEsbuildReturnedErrors() {
}

func TestAccPackageDataSourceEsbuildReturnedWarnings() {
}

func TestAccPackageDataSourceBuildsDeterministicZip() {
}

func TestAccExampleDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lambdapackager": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccExampleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaffolding_example.test", "id", "example-id"),
				),
			},
		},
	})
}

const testAccExampleDataSourceConfig = `
data "lambdapackager_package" "test" {
  args              = ["--bundle"]
  entrypoint        = "/app/lambda/hander/src/index.ts"
  working_directory = "/app/lambda"
}
`
