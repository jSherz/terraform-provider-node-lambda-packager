package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Entrypoint file does not exist
func TestAccPackageDataSourceCannotFindEntrypoint(t *testing.T) {
	const wdDoesNotExistConfig = `
data "lambdapackager_package" "test" {
  args              = ["--bundle"]
  entrypoint        = "/does/not/exist/handler/src/index.ts"
  working_directory = "/does/not/exist"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lambdapackager": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config:      wdDoesNotExistConfig,
				ExpectError: regexp.MustCompile("You specified the /does/not/exist/handler/src/index.ts entrypoint which was\\s+resolved to /does/not/exist/handler/src/index.ts and we could not find the\\s+file or do not have permission to view it: stat\\s+/does/not/exist/handler/src/index.ts: no such file or directory"),
			},
		},
	})
}

// Permission denied accessing entrypoint file
func TestAccPackageDataSourceCannotAccessEntrypoint(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "terraform-provider-lambdapackager-*")
	if err != nil {
		t.Fatalf("Failed to make temp dir for test setup: %s", err)
	}

	// No read permissions
	entrypoint := filepath.Join(tmpDir, "index.ts")
	err = os.WriteFile(entrypoint, []byte("foo"), os.FileMode(0664))
	if err != nil {
		t.Fatalf("Failed to write entrypoint in test setup: %s", err)
	}

	// No read, no execute permissions
	err = os.Chmod(tmpDir, os.FileMode(0222))
	if err != nil {
		t.Fatalf("Failed to set test dir permissions in test setup: %s", err)
	}

	config := fmt.Sprintf(`
data "lambdapackager_package" "test" {
  args              = ["--bundle"]
  entrypoint        = "%s"
  working_directory = "%s"
}
`, entrypoint, tmpDir)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"lambdapackager": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config:      config,
				ExpectError: regexp.MustCompile(fmt.Sprintf("You specified the\\s*%s\\s*entrypoint which was\\s+resolved to\\s*%s\\s*and we could not find the\\s+file or do not have permission to view it: stat\\s*%s:\\s*permission denied", entrypoint, entrypoint, entrypoint)),
			},
		},
	})
}

// Unrecognized arguments
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

//func TestAccExampleDataSource(t *testing.T) {
//	resource.Test(t, resource.TestCase{
//		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
//			"lambdapackager": providerserver.NewProtocol6WithError(New("test")()),
//		},
//		Steps: []resource.TestStep{
//			// Read testing
//			{
//				Config: testAccExampleDataSourceConfig,
//				Check: resource.ComposeAggregateTestCheckFunc(
//					resource.TestCheckResourceAttr("data.scaffolding_example.test", "id", "example-id"),
//				),
//			},
//		},
//	})
//}

const testAccExampleDataSourceConfig = `
data "lambdapackager_package" "test" {
  args              = ["--bundle"]
  entrypoint        = "/app/lambda/hander/src/index.ts"
  working_directory = "/app/lambda"
}
`

const testAccExampleDataSourceRelativePathConfig = `
data "lambdapackager_package" "test" {
  args              = ["--bundle"]
  entrypoint        = "lambda/handler/src/index.ts"
  working_directory = "lambda"
}
`
