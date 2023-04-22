package provider_test

import (
	"regexp"
	"terraform-provider-lambda-packager/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
	t.Skipped()
}

func TestAccPackageDataSourceGettingWorkingDirectoryAbsolutePathFailed(t *testing.T) {
	t.Skipped()
}

func TestAccPackageDataSourceEsbuildReturnedErrors(t *testing.T) {
	t.Skipped()
}

func TestAccPackageDataSourceEsbuildReturnedWarnings(t *testing.T) {
	t.Skipped()
}

func TestAccPackageDataSourceBuildsDeterministicZip(t *testing.T) {
	t.Skipped()
}
