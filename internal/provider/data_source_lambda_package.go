// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package provider builds the Terraform provider. See the README for usage.
package provider

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/evanw/esbuild/pkg/cli"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LambdaPackageDataSource struct{}

type LambdaPackageDataSourceModel struct {
	Args             types.List   `tfsdk:"args"`
	Entrypoint       types.String `tfsdk:"entrypoint"`
	WorkingDirectory types.String `tfsdk:"working_directory"`
	Filename         types.String `tfsdk:"filename"`
	SourceCodeHash   types.String `tfsdk:"source_code_hash"`
}

const descriptionPrefix = "Uses ebsuild to package a Lambda function and make it ready for"

const description = descriptionPrefix + " AWS."

const markdownDescription = descriptionPrefix + " [aws_lambda_function](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_function).\n\n" +
	`## Example Usage

BACKTICKBACKTICKBACKTICKterraform
data "node-lambda-packager_package" "my_function" {
  args = [
    "--bundle",
    # Here's an example package that is provided by the runtime (NodeJS 18)
    "--external:@aws-sdk*",
    # Here's an example package that might be provided by a layer
    "--external:@aws-lambda-powertools*",
    "--minify",
    "--platform=node",
    "--sourcemap",
    # This must be appropriate for the Node version you choose
    "--target=es2021",
    "--sourcemap=inline",
  ]

  # Try co-locating your Lambda code and Terraform in one project
  entrypoint        = "../lambda/src/handlers/user-deleted-listener/index.ts"
  working_directory = "../lambda"
}

resource "aws_lambda_function" "my_function" {
  function_name    = "my-function"
  role             = aws_iam_role.this.arn
  handler          = "index.handler"
  memory_size      = 128
  runtime          = "nodejs18.x"
  filename         = data.node-lambda-packager_package.my_function.filename
  source_code_hash = data.node-lambda-packager_package.my_function.source_code_hash
}
BACKTICKBACKTICKBACKTICK
`

var _ datasource.DataSource = &LambdaPackageDataSource{}

func NewLambdaPackageDataSource() datasource.DataSource {
	return &LambdaPackageDataSource{}
}

func (d *LambdaPackageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package"
}

func (d *LambdaPackageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         description,
		MarkdownDescription: strings.ReplaceAll(markdownDescription, "BACKTICK", "`"),
		Blocks:              map[string]schema.Block{},
		DeprecationMessage:  "",

		Attributes: map[string]schema.Attribute{
			//nolint:exhaustruct // too many props to be useful
			"args": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Arguments to pass to esbuild.",
			},
			//nolint:exhaustruct // too many props to be useful
			"entrypoint": schema.StringAttribute{
				Required:    true,
				Description: "Path to lambda function entrypoint.",
			},
			//nolint:exhaustruct // too many props to be useful
			"working_directory": schema.StringAttribute{
				Required:    true,
				Description: "Typically the folder containing the package.json at the root of your Lambda project.",
			},
			//nolint:exhaustruct // too many props to be useful
			"filename": schema.StringAttribute{
				Computed:    true,
				Description: "Path to the packaged lambda zip.",
			},
			//nolint:exhaustruct // too many props to be useful
			"source_code_hash": schema.StringAttribute{
				Computed:    true,
				Description: "Source code hash of the built package.",
			},
		},
	}
}

func formatLocation(location *api.Location) string {
	return fmt.Sprintf("file: %s, line %s - suggestion: %s", location.File, location.LineText, location.Suggestion)
}

func (d *LambdaPackageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
}

//nolint:cyclop,funlen,gocognit,gocyclo // needs refactoring after tests are complete
func (d *LambdaPackageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LambdaPackageDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	entrypointPath := data.Entrypoint.ValueString()

	fullEntrypointPath, err := filepath.Abs(entrypointPath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not find find the full entrypoint path",
			fmt.Sprintf("You specified the %s entrypoint which could not be resolved into an absolute path: %s", entrypointPath, err),
		)

		return
	}

	_, err = os.Stat(fullEntrypointPath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not find entrypoint file",
			fmt.Sprintf("You specified the %s entrypoint which was resolved to %s and we could not find the file or do not have permission to view it: %s", entrypointPath, fullEntrypointPath, err),
		)

		return
	}

	rawArgs := data.Args.Elements()

	args := make([]string, 0, len(rawArgs))
	for _, rawArg := range rawArgs {
		//nolint:forcetypeassert // we define these as strings
		args = append(args, (rawArg.(types.String)).ValueString())
	}

	args = append(args, fullEntrypointPath)

	buildArgs, err := cli.ParseBuildOptions(args)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse build options",
			fmt.Sprintf("Error: %s", err),
		)
	}

	workingDirectory := data.WorkingDirectory.ValueString()

	absWorkingDirectory, err := filepath.Abs(workingDirectory)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not find find the full working directory path",
			fmt.Sprintf("You specified the %s working_directory which could not be resolved into an absolute path: %s", entrypointPath, err),
		)

		return
	}

	buildArgs.AbsWorkingDir = absWorkingDirectory

	result := api.Build(buildArgs)

	if len(result.Errors) >= 1 {
		for _, message := range result.Errors {
			detail := ""

			if message.Location != nil {
				detail = detail + formatLocation(message.Location) + "\n\n"
			}

			for _, note := range message.Notes {
				if note.Location != nil {
					detail = detail + "\n" + fmt.Sprintf("%s: %s", formatLocation(note.Location), note.Text)
				} else {
					detail = detail + "\n" + note.Text
				}
			}

			resp.Diagnostics.AddError(
				fmt.Sprintf("esbuild error: %s", message.Text),
				detail,
			)
		}

		return
	}

	if len(result.Warnings) >= 1 {
		for _, message := range result.Warnings {
			detail := ""

			if message.Location != nil {
				detail = detail + formatLocation(message.Location) + "\n\n"
			}

			for _, note := range message.Notes {
				if note.Location != nil {
					detail = detail + "\n" + fmt.Sprintf("%s: %s", formatLocation(note.Location), note.Text)
				} else {
					detail = detail + "\n" + note.Text
				}
			}

			resp.Diagnostics.AddWarning(
				fmt.Sprintf("esbuild error: %s", message.Text),
				detail,
			)
		}
	}

	packageFile, err := os.CreateTemp("", "esbuild-lambda-packager-*")
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create temporary file for packaged Lambda",
			fmt.Sprintf("Error: %s", err),
		)

		return
	}

	defer func(packageFile *os.File) {
		_ = packageFile.Close()
	}(packageFile)

	zipWriter := zip.NewWriter(packageFile)

	//nolint:exhaustruct // too many props to be useful
	indexDotJs, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:     "index.js",
		Method:   zip.Deflate,
		Modified: time.Date(2020, 8, 31, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create ZIP with our Lambda bundle",
			fmt.Sprintf("Error: %s", err),
		)

		return
	}

	_, err = indexDotJs.Write(result.OutputFiles[0].Contents)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to compress Lambda bundle into zip",
			fmt.Sprintf("Error: %s", err),
		)

		return
	}

	err = zipWriter.Close()

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to close Lambda package zip",
			fmt.Sprintf("Error: %s", err),
		)

		return
	}

	res, err := os.ReadFile(packageFile.Name())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read back Lambda package zip to hash it",
			fmt.Sprintf("Error: %s", err),
		)

		return
	}

	hash := sha256.Sum256(res)
	encodedHash := base64.StdEncoding.EncodeToString(hash[:])

	data.Filename = types.StringValue(packageFile.Name())
	data.SourceCodeHash = types.StringValue(encodedHash)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
