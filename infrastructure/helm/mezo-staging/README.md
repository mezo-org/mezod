# Helm: mezo-staging

This module contains Helm artifacts for the `mezo-staging-gke-cluster` cluster
created by the corresponding [Terraform module](./../../terraform/mezo-staging/README.md).

### Usage

The [`mezo-staging` Terraform module](./../../terraform/mezo-staging/README.md)
manages cluster's Helm releases using the Terraform Helm provider. Please refer to the 
Terraform module documentation for more details.

This module is a complementary part holding custom values overriding the default values
of the deployed Helm charts.