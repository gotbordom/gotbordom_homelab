variable "cloudflare_api_token" {
  type        = string
  description = "Scoped Cloudflare API token (see docs/installation/production/external-resources.md for the required permissions)"
  sensitive   = true
}

variable "cloudflare_account_id" {
  type = string
}

variable "ntfy" {
  type = object({
    url   = string
    topic = string
  })

  sensitive = true
}

variable "extra_secrets" {
  type        = map(string)
  description = "Key-value pairs of extra secrets that cannot be randomly generated (e.g. third party API tokens)"
  sensitive   = true
  default     = {}
}
