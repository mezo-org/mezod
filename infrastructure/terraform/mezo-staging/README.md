# Terraform: mezo-staging

This module contains Terraform configuration for the `mezo-staging` GCP project.

### Prerequisites

- The `mezo-staging` GCP project. The GCP project ID should be set in the `.env` file
- Service account with email `terraform@<project-id>.iam.gserviceaccount.com` (**Editor** role assigned)
- Terraform (at least v1.8.1). Recommended approach is using the [tfenv](https://github.com/tfutils/tfenv) version manager
- [dotenv](https://www.npmjs.com/package/dotenv) with a plugin loading env 
  variables to your shell ([example for oh-my-zsh](https://github.com/ohmyzsh/ohmyzsh/tree/master/plugins/dotenv)).
  This is necessary to load the environment variables from the `.env` file.
  An alternative is to set the environment variables from this file manually
- [1password CLI](https://developer.1password.com/docs/cli/get-started) installed and configured to use an account with access 
  to the vault holding SSL certificates (see template files living in the 
  `ssl-certificates` directory for details) for the project

### Authentication

Terraform impersonates the service account `terraform@<project-id>.iam.gserviceaccount.com` 
to perform operations on the GCP project. This is configured through the
`GOOGLE_IMPERSONATE_SERVICE_ACCOUNT` variable in the `.env` file.

In order to make it work, your personal GCP account needs to have the 
`roles/iam.serviceAccountTokenCreator` role assigned. You also need to
authenticate by doing:
```shell
gcloud auth application-default login
```

### Terraform state

Terraform requires a GCP bucket named `mezo-staging-terraform-backend-bucket` to 
store its state. If the bucket already exists, you can skip this step. 
Otherwise, you can create it by moving to the
`remote-state` directory and invoking:
```shell
terraform init && terraform apply
```

### Create and modify infrastructure resources

To create (or modify) the infrastructure resources, move to the `mezo-staging` root directory
follow the steps below.
1. Load SSL certificates from 1password (**this action needs to be done only once**):
    ```shell
    ./load-ssl-certificates.sh
    ```

2. Initialize Terraform (**this action needs to be done only once**):
    ```shell
    terraform init
    ``` 

3. Plan the changes:
    ```shell
    terraform plan
    ```
4. Apply the changes:
    ```shell
    terraform apply
    ```

### Supplementary non-managed resources

The `mezo-staging` GCP project requires some supplementary resources that are 
not managed by Terraform at the moment (they should be ported to Terraform in the future).

#### Cloudflare

Cloudflare is used as DNS provider for all domains used in the project.
Domains are managed manually in the Cloudflare dashboard.

Moreover, some `mezo-staging` services require Cloudflare proxy to be configured. 
In such a case, the following setup is used:
```asciidoc
[User] --- HTTPS ---> [Cloudflare] --- HTTPS ---> [GCP mezo-staging service]
```
To make it work:
- The Cloudflare proxy should be enabled for the given service domain
- A Cloudflare [edge certificate](https://developers.cloudflare.com/ssl/edge-certificates/) 
  should be created for the service domain. Note that Cloudflare automatically
  issues universal edge certificates for 
  [one-level subdomains](https://developers.cloudflare.com/ssl/troubleshooting/version-cipher-mismatch/#multi-level-subdomains) 
  but multi-level subdomains require an 
  [advanced edge certificate to be issued manually](https://developers.cloudflare.com/ssl/edge-certificates/advanced-certificate-manager/manage-certificates/)
- The service must be configured to accept HTTPS connections (e.g. through 
  a global HTTPS load balancer). Recommended setup is limiting connection
  sources to IPs corresponding to Cloudflare proxy servers
- A Cloudflare [origin certificate](https://developers.cloudflare.com/ssl/origin-configuration/origin-ca) 
  should be created and attached to the service 
  (e.g. as a pre-shared GCP SSL certificate tied to a global HTTPS load balancer).
  The `mezo-staging` Terraform module can automatically load origin certificates
  from 1password and attach them to appropriate 
  services (see `./load-ssl-certificates.sh` script)
  
A good example of a Cloudflare-proxied service exposed by the `mezo-staging` GCP project
is the Mezo Blockscout explorer available at `https://explorer.test.mezo.org`.





