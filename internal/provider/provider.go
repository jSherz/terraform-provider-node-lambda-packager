// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type NodeLambdaPackagerProvider struct {
	version string
}

type NodeLambdaPackagerProviderModel struct{}

var _ provider.Provider = &NodeLambdaPackagerProvider{
	version: "0.0.2",
}

func (p *NodeLambdaPackagerProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "node-lambda-packager"
	resp.Version = p.version
}

func (p *NodeLambdaPackagerProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	description := "Helps you to co-locate your Lambda application code with your IaC, using esbuild to bundle your Lambda code."

	resp.Schema = schema.Schema{
		Attributes:          map[string]schema.Attribute{},
		Blocks:              map[string]schema.Block{},
		Description:         description,
		MarkdownDescription: description + "\n\n## Example Usage\n\n" + "```terraform\n" + "provider \"node-lambda-packager\" {}\n```",
		DeprecationMessage:  "",
	}
}

func (p *NodeLambdaPackagerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data NodeLambdaPackagerProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *NodeLambdaPackagerProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *NodeLambdaPackagerProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewLambdaPackageDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NodeLambdaPackagerProvider{
			version: version,
		}
	}
}
