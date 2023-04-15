// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	diag2 "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"os"
	"os/exec"
	"terraform-provider-lambda-packager/internal/packager"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &EsbuildLambdaPackagerProvider{}

type EsbuildLambdaPackagerProvider struct {
	version string
}

type EsbuildLambdaPackagerProviderModel struct {
	EsbuildBinary types.String `tfsdk:"esbuild_binary"`
}

func (p *EsbuildLambdaPackagerProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "lambdapackager"
	resp.Version = p.version
}

func (p *EsbuildLambdaPackagerProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"esbuild_binary": &schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (p *EsbuildLambdaPackagerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data EsbuildLambdaPackagerProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	esbuildBinary := "esbuild"

	overrideEsbuildBinary := data.EsbuildBinary.ValueString()

	if overrideEsbuildBinary == "" {
		overrideEsbuildBinary = os.Getenv("LAMBDA_PACKAGER_ESBUILD_BINARY")
	}

	if overrideEsbuildBinary != "" {
		esbuildBinary = overrideEsbuildBinary

		overrideEsbuildBinaryStat, err := os.Stat(esbuildBinary)

		if err != nil {
			resp.Diagnostics.Append(diag2.NewErrorDiagnostic(
				"An override esbuild_binary was set and is invalid",
				fmt.Sprintf("You specified the %s esbuild_binary and we could not find the file or do not have permission to view it: %s", esbuildBinary, err),
			))
			return
		}

		if overrideEsbuildBinaryStat.IsDir() {
			resp.Diagnostics.Append(
				diag2.NewErrorDiagnostic("An override esbuild_binary was set and is invalid",
					fmt.Sprintf("You specified the %s esbuild_binary and it's a directory, not a file.", esbuildBinary),
				))
			return
		}
	} else {
		_, err := exec.LookPath(esbuildBinary)

		if err != nil {
			resp.Diagnostics.Append(
				diag2.NewErrorDiagnostic("Could not find the esbuild binary",
					fmt.Sprintf("Could not find the esbuild binary. Is it installed and in your path? Error: %s", err),
				))
			return
		}
	}

	client := packager.NewPackager(esbuildBinary)
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *EsbuildLambdaPackagerProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *EsbuildLambdaPackagerProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewLambdaPackageDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &EsbuildLambdaPackagerProvider{
			version: version,
		}
	}
}
