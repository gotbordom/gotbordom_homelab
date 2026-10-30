# External resources

!!! info

    These resources are optional, the homelab still works without them but will lack some features like trusted certificates and offsite backup

Although I try to keep the amount of external resources to the minimum, there's still need for a few of them.
Below is a list of external resources and why we need them (also see some [alternatives](#alternatives) below).

| Provider        | Resource  | Purpose                                                                                                     |
| --------        | --------  | -------                                                                                                     |
| Cloudflare      | DNS       | DNS and [DNS-01 challenge](https://letsencrypt.org/docs/challenge-types/#dns-01-challenge) for certificates |
| Cloudflare      | Tunnel    | Public services to the internet without port forwarding                                                     |
| ntfy            | Topic     | External notification service to receive alerts                                                             |

## Create credentials

You'll be asked to provide these credentials on first build.

### Terraform state

OpenTofu state is stored locally in `external/terraform.tfstate`, next to
`external/terraform.tfvars`. Both are gitignored and both contain secrets --
back them up together (e.g. in a password manager). If the state is lost,
`make external` will try to recreate everything; delete the old `homelab`
tunnel and `homelab_*` API tokens in Cloudflare by hand first.

To use a [remote backend](https://opentofu.org/docs/language/settings/backends/configuration/) instead, edit `external/versions.tf`.

### Cloudflare

- Use a domain whose DNS is on Cloudflare: either [add an existing domain](https://developers.cloudflare.com/fundamentals/setup/manage-domains/add-site/) and point its nameservers at Cloudflare (registration and website hosting can stay where they are), or buy one through Cloudflare Registrar
- Get Cloudflare email and account ID
- Global API key: <https://dash.cloudflare.com/profile/api-tokens>

<!-- TODO switch to API token instead of API key? -->
<!-- Terraform API token summary: -->

<!-- ``` -->
<!-- This API token will affect the below accounts and zones, along with their respective permissions -->

<!-- └── Khue Doan - Argo Tunnel:Edit, Account Settings:Read -->
<!--     └── khuedoan.com - Zone:Read, DNS:Edit -->

<!-- Client IP Address Filtering -->

<!-- └── Is in - 117.xxx.xxx.xxx, 2402:xxx:xxx:xxx:xxx:xxx:xxx:xxx -->
<!-- ``` -->

<!-- ### Create Minio keys -->

<!-- TODO: skip this for now -->

<!-- ### Create AWS API key -->

<!-- TODO: skip this for now -->

### ntfy

- Choose a topic name like <https://ntfy.sh/random_topic_name_here_a8sd7fkjxlkcjasdw33813> (treat it like you password)

## Alternatives

To avoid vendor lock-in, each external provider must have an equivalent alternative that is easy to replace:

- Local state:
    - Any other [OpenTofu backend](https://opentofu.org/docs/language/settings/backends/configuration/)
- Cloudflare DNS:
    - Update cert-manager and external-dns to use a different provider
    - [Alternate DNS setup](../../how-to-guides/alternate-dns-setup.md)
- Cloudflare Tunnel:
    - Use port forwarding if it's available
    - Create a small VPS in the cloud and utilize Wireguard to route traffic via it
    - Access everything via VPN
    - See also [awesome tunneling](https://github.com/anderspitman/awesome-tunneling)
- ntfy:
    - [Self-host your own ntfy server](https://docs.ntfy.sh/install)
    - Any other [integration supported by Grafana Alerting](https://grafana.com/docs/grafana/latest/alerting/alerting-rules/manage-contact-points/integrations/#list-of-supported-integrations)
