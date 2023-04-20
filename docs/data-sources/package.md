---
page_title: "package Data Source - terraform-provider-node-lambda-packager"
subcategory: ""
description: |-
  Bundles and zips up your Lambda code with the help of esbuild.
---

# Data Source `package`

The `node-lambda-packager_package` data source uses [esbuild] to bundle up
Lambda code and make it ready for use in `aws_lambda_function`.

[esbuild]: https://esbuild.github.io

## Example Usage

```terraform
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
```

## Argument Reference

- `args` - See `esbuild --help` for the options.
- `entrypoint` - The file containing your handler function.
- `working_directory` - The folder containing your package.json.

## Attributes Reference

The following attributes are exported.

- `filename` - Pass to `aws_lambda_function` - the zip bundle.
- `source_code_hash` - Pass to `aws_lambda_function` - should not change if
  your code does not.
