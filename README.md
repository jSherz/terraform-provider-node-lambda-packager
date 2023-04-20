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
