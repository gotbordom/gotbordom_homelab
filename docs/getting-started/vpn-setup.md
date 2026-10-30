# VPN setup

This homelab uses [Tailscale](https://tailscale.com) for remote access. A
self-hosted [WireGuard](https://www.wireguard.com) server used to be an option
too; it's in the backlog for now (see the WireGuard decision record in
`docs/reference/architecture/decision-records.md` for why, and when to bring it
back).

## Tailscale (requires third-party account)

Get an [auth key](https://tailscale.com/kb/1085/auth-keys) from [Tailscale admin console](https://login.tailscale.com/admin/authkeys):

- Description: homelab
- Reusable: optionally set this to true

Add it to `external/terraform.tfvars` as an extra secret:

```hcl
extra_secrets = {
  tailscale-auth-key = "tskey-auth-myauthkeyhere"
}
```

You may want to back up the `external/terraform.tfvars` file to a secure location.

Apply the secret:

```sh
make external
```

Finally, [enable subnet routes](https://tailscale.com/kb/1019/subnets#step-3-enable-subnet-routes-from-the-admin-console) for `homelab-router`
(it advertises `192.168.0.0/24`, set by `TS_ROUTES` in `apps/tailscale/values.yaml` -- change it if the home subnet changes)
from the [admin console](https://login.tailscale.com/admin/machines).

You can now connect to your homelab via Tailscale and [invite user to your Tailscale network](https://tailscale.com/kb/1371/invite-users).
