terraform {
  required_version = "~> 1.7"

  # State lives next to terraform.tfvars (both gitignored); back them up
  # together. See docs/installation/production/external-resources.md.
  backend "local" {}

  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.30.0"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.26.0"
    }

    http = {
      source  = "hashicorp/http"
      version = "~> 3.4.0"
    }
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

provider "kubernetes" {
  # Use KUBE_CONFIG_PATH environment variables
  # Or in cluster service account
}
