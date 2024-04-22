# Terraform: mezo-staging

This module contains Terraform configuration for the `mezo-staging` GCP project.

### Prerequisites

- The `mezo-staging` GCP project
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

Terraform requires a GCP bucket named `mezo-staging-terraform-backend-bucket` to 
store its state. If the bucket already exists, you can skip this step. 
Otherwise, you can create it by moving to the
`remote-state` directory and invoking:
```shell
terraform init && terraform apply
```

Type the project ID when prompted. Once done, the bucket will be created.
**This action needs to be done only once.**

### Create infrastructure resources

To create the infrastructure resources, move to the `mezo-staging` root directory
follow the steps below. Type the project ID when prompted.
1. Initialize Terraform (**this action needs to be done only once**):
    ```shell
    terraform init
    ``` 

2. Plan the changes:
    ```shell
    terraform plan
    ```

3. Apply the changes:
    ```shell
    terraform apply
    ```
   
In order to avoid typing the project ID every time, you can put it in an 
environment variable `TF_VAR_project_id`:
```shell
export TF_VAR_project_id=<project-id>
```
or set it using [another way supported by Terraform.](https://developer.hashicorp.com/terraform/language/values/variables#assigning-values-to-root-module-variables)
