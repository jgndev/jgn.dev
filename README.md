# jgn's personal site and blog

This repository contains the source code and terraform configuration for the jgn.dev website using Terraform and Google Cloud Platform.

## Prerequisites

1. A Terraform Cloud account
2. A Google Cloud Platform account
3. Git

## Setup Instructions

1. Fork this repository.

2. Sign up for Terraform Cloud and create a new organization if you haven't already.

3. In Terraform Cloud, create a new workspace named "jgn-dev" or a name of your choosing..

4. In the Terraform Cloud workspace, set up the following environment variables:
    - GOOGLE_CREDENTIALS (sensitive): Contents of your Google Cloud service account key JSON file
    - TF_VAR_project_id: Your GCP project ID
    - TF_VAR_org_id: Your GCP organization ID
    - TF_VAR_billing_account: Your GCP billing account ID

5. Update the `terraform.tf` file in this repository with your Terraform Cloud organization name:

   ```hcl
   terraform {
     cloud {
       organization = "your-org-name"
       workspaces {
         name = "jgn-website"
       }
     }
     # ... rest of the configuration
   }