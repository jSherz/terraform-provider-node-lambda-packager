// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/evanw/esbuild/pkg/cli"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"path/filepath"
	"time"
)

var _ datasource.DataSource = &LambdaPackageDataSource{}

func NewLambdaPackageDataSource() datasource.DataSource {
	return &LambdaPackageDataSource{}
}

type LambdaPackageDataSource struct {
}

type LambdaPackageDataSourceModel struct {
	Args             types.List   `tfsdk:"args"`
	Entrypoint       types.String `tfsdk:"entrypoint"`
	WorkingDirectory types.String `tfsdk:"working_directory"`
	Filename         types.String `tfsdk:"filename"`
	SourceCodeHash   types.String `tfsdk:"source_code_hash"`
}

func (d *LambdaPackageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package"
}

func (d *LambdaPackageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example data source",

		Attributes: map[string]schema.Attribute{
			"args": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Arguments to pass to esbuild.",
			},
			"entrypoint": schema.StringAttribute{
				Required:    true,
				Description: "Path to lambda function entrypoint.",
			},
			"working_directory": schema.StringAttribute{
				Required:    true,
				Description: "Typically the folder containing the package.json at the root of your Lambda project.",
			},
			"filename": schema.StringAttribute{
				Computed:    true,
				Description: "Path to the packaged lambda zip.",
			},
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

	statRes, err := os.Stat(fullEntrypointPath)
	fmt.Println(statRes)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not find entrypoint file",
			fmt.Sprintf("You specified the %s entrypoint which was resolved to %s and we could not find the file or do not have permission to view it: %s", entrypointPath, fullEntrypointPath, err),
		)
		return
	}

	rawArgs := data.Args.Elements()
	var args []string
	for _, rawArg := range rawArgs {
		args = append(args, (rawArg.(types.String)).ValueString())
	}

	//args = append(args, fmt.Sprintf("--outfile=%s", outputFilePath))
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
