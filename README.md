# terraform-provider-node-lambda-packager

This Terraform provider helps you to co-locate your Lambda application code
with your Infrastructure as Code (IaC). It uses esbuild to produce a bundle
that's then zipped up and made ready for Terraform to consume.

See `./examples` for practical usage.

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
