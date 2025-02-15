# terraform-provider-node-lambda-packager

![Terraform](https://img.shields.io/badge/terraform-%235835CC.svg?style=for-the-badge&logo=terraform&logoColor=white) ![Terraform Provider Downloads](https://img.shields.io/terraform/provider/dt/4103.svg?style=for-the-badge)

This Terraform provider helps you to co-locate your Lambda application code
with your Infrastructure as Code (IaC). It uses esbuild to produce a bundle
that's then zipped up and made ready for Terraform to consume.

## Usage

```terraform
#
# 1. Use the data source to package up your NodeJS-based Lambda function.
#
data "node-lambda-packager_package" "create_user" {
  #
  # 2. Set esbuild arguments as appropriate.
  #
  # See: https://esbuild.github.io/api/#build-api
  #
  args = [
    "--bundle",
    "--minify",
    "--platform=node",
    "--sourcemap",
    "--target=es2024",
    "--sourcemap=inline",
  ]

  #
  # 3. The entrypoint file is relative to the working directory and you can
  #    leave the working directory empty if it's already correct.
  #
  entrypoint        = "src/handlers/v1/admin/users/create/index.ts"
  working_directory = ""
}

#
# 4. Define your Lambda function as normal!
#
resource "aws_lambda_function" "create_user" {
  function_name = "create-user"
  role          = aws_iam_role.create_user.arn
  architectures = ["arm64"]
  handler       = "index.handler"
  memory_size   = 1024
  runtime       = "nodejs22.x"

  #
  # 5. Here's where our data source is used:
  #
  filename         = data.node-lambda-packager_package.create_user.filename
  source_code_hash = data.node-lambda-packager_package.create_user.source_code_hash
}
```

See [jSherz/automate-aws-access-removal] for a more in-depth example of this
provider in use.

[jSherz/automate-aws-access-removal]: https://github.com/jSherz/automate-aws-access-removal

## Build

```bash
make build
```

## Install for local testing

```bash
make install
```

## Test

```bash
make testacc
```

## Linting

For install steps, see [golang-ci-lint on GitHub]. Find the correct version to
use in the `./github/workflows/golangci-lint.yml` file.

[golang-ci-lint on GitHub]: https://github.com/golangci/golangci-lint

```bash
golangci-lint run
```

### Fixing formatting or input problems

You can run a single linter in fix mode as follows:

```bash
golangci-lint run --no-config --enable gofumpt --fix
```

[gci] and [gofumpt] have to be run in order to have both agree about the import
order and formatting:

```bash
golangci-lint run --no-config --enable gci --fix
golangci-lint run --no-config --enable gofumpt --fix
```

[gci]: https://github.com/daixiang0/gci
[gofumpt]: https://github.com/mvdan/gofumpt
