---
page_title: "Provider: node-lambda-packager"
subcategory: ""
description: |-
  Co-locate your Terraform code with your functions - uses esbuild to package Node.JS Lambdas.
---

# node-lambda-packager Provider

## Example Usage

You can set the `LAMBDA_PACKAGER_ESBUILD_BINARY` environment variable to help
the provider find esbuild.

For example:

```
/projects/lambda/node_modules/.bin/esbuild
```

```terraform
provider "node-lambda-packager" {}
```

Or you can set it in the provider:

```terraform
provider "node-lambda-packager" {
  esbuild_binary = "/projects/lambda/node_modules/.bin/esbuild"
}
```

## Schema

### Optional

- **esbuild_binary** (String, Optional) esbuild binary, if it's not in the PATH

  Can be set with the `LAMBDA_PACKAGER_ESBUILD_BINARY` variable.
