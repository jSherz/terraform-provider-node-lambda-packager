# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = "~> 1.3"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.60"
    }
    node-lambda-packager = {
      source  = "jsherz.com/tf/node-lambda-packager"
      version = "1.0.0"
    }
  }
}

provider "node-lambda-packager" {}

data "node-lambda-packager_package" "this" {
  args = [
    "--bundle",
    "--external:@aws-sdk*",
    "--external:@aws-lambda-powertools*",
    "--minify",
    "--platform=node",
    "--sourcemap",
    "--target=es2021",
    "--sourcemap=inline",
  ]

  entrypoint        = var.entrypoint
  working_directory = var.working_directory
}

resource "aws_lambda_function" "this" {
  function_name    = var.name
  role             = aws_iam_role.this.arn
  architectures    = ["arm64"]
  description      = var.description
  handler          = "index.handler"
  memory_size      = var.memory_size
  runtime          = "nodejs18.x"
  filename         = data.node-lambda-packager_package.this.filename
  source_code_hash = data.node-lambda-packager_package.this.source_code_hash

  reserved_concurrent_executions = var.reserved_concurrent_executions

  environment {
    variables = var.env_vars
  }

  layers = [
    # See: https://awslabs.github.io/aws-lambda-powertools-typescript/latest/
    "arn:aws:lambda:${data.aws_region.this.name}:094274105915:layer:AWSLambdaPowertoolsTypeScript:10",
  ]

  tracing_config {
    mode = "Active"
  }
}
