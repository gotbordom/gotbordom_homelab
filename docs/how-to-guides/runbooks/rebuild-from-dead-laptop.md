# Recover from a dead laptop

Use this to regain control of the cluster from a replacement machine. The cluster keeps running without the
laptop, so there's no rush.

**Status:** untested. It needs a real test (#24).

## What you need

| Item | Why | If it's lost |
|---|---|---|
| SSH key for the nodes (`~/.ssh/id_ed25519`) | Root access to nodes | See [No SSH key](#no-ssh-key) |
| `external/terraform.tfvars` and `external/terraform.tfstate` | Cloudflare and other external resources | See [No Terraform files](#no-terraform-files) |
| `metal/kubeconfig.yaml` | Cluster admin access | Get a new copy from thinkcenter0 |
| Tailscale login | Remote access only | Sign in again |
| GitHub login | Repo access and merges | Recover through GitHub |

This repo is public, so these files must never be committed. Store them in a password manager or similar.

## Set up the new machine

1. Install [Nix](https://nixos.org/download/) with flakes enabled, then clone the repo:

    ```sh
    git clone https://github.com/gotbordom/gotbordom_homelab
    cd gotbordom_homelab
    ```

2. Restore the SSH key, then set its permissions:

    ```sh
    chmod 600 ~/.ssh/id_ed25519
    ```

3. Restore `external/terraform.tfvars` and `external/terraform.tfstate`, keeping the two from the same backup.

4. If you're away from home, install Tailscale and run `sudo tailscale up --accept-routes`.

## Regain cluster access

!!! danger

    Don't run `make metal` or `make -C metal boot` to recover. The boot stage sends Wake-on-LAN packets and runs a
    PXE server, and a node that network-boots from it can be **reinstalled**. Use only the steps below.

Check the host key prompt on first connection against what you expect, then fetch a new kubeconfig:

```sh
ssh root@192.168.0.9 cat /etc/rancher/k3s/k3s.yaml | sed 's/127.0.0.1/192.168.0.239/' > metal/kubeconfig.yaml
chmod 600 metal/kubeconfig.yaml
nix develop
export KUBECONFIG=$PWD/metal/kubeconfig.yaml
kubectl get nodes
```

You should see all 3 nodes `Ready`. Check that apps and Ceph are healthy:

```sh
kubectl -n argocd get applications
kubectl -n rook-ceph exec deploy/rook-ceph-operator -- ceph --conf /var/lib/rook/rook-ceph/rook-ceph.config -s
```

If Terraform files were restored, confirm that the state matches reality. This should report no changes:

```sh
cd external
tofu init
tofu plan
```

Read a plan carefully before you apply it.

## Clean up

- Remove the dead laptop's device from your tailnet.
- If the laptop may have been stolen rather than broken, rotate what it could access: its SSH key on all nodes, the
  kubeconfig, the Cloudflare token and other values in `terraform.tfvars`, and your GitHub sessions and tokens.

## If something is missing

### No SSH key

Try any other machine with an authorized key. Otherwise, you need physical access to a node's console. If you can't log
in there either, a node has to be reimaged to install a new key, which is a large job on the control plane. See
[Add or remove nodes](../add-or-remove-nodes.md) and [Recover from control plane loss](control-plane-loss.md).

### No kubeconfig

Fetch a new one using the steps above. You only need SSH access to thinkcenter0.

### No Terraform files

The cluster keeps running, but OpenTofu no longer knows what it manages.

- `terraform.tfvars` can be rebuilt from `terraform.tfvars.example` with new credentials.
- Without `terraform.tfstate`, a plain `tofu apply` may try to create resources that already exist. Import them with
  `tofu import`, or remove them from Cloudflare first, which interrupts whatever depends on them. Read the plan before
  applying.
