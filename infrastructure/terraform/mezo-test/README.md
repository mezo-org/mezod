# Terraform: mezo-test

This module contains Terraform configuration for the `mezo-test` GCP project.

### Prerequisites

- The `mezo-test` GCP project
- JSON key of the Terraform service account (with **Editor** role)
- Terraform (at least v1.8.1)

### Terraform authentication

Terraform uses a service account to authenticate with GCP. To make it possible, 
the `GOOGLE_CREDENTIALS` environment variable should point to the service 
account's JSON key file:

```shell
export GOOGLE_CREDENTIALS=<service-account-json-key-file>
```

### Terraform state

Terraform requires a GCP bucket named `mezo-test-terraform-backend-bucket` to 
store its state. If the bucket already exists, you can skip this step. 
Otherwise, you can create it by moving to the
`remote-state` directory and invoking:
```shell
terraform init && terraform apply
```

Type the project ID when prompted. Once done, the bucket will be created.
**This action needs to be done only once.**

### Create infrastructure resources

TODO