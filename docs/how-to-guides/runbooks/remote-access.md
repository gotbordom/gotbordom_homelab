# Work on the cluster remotely

Use this to reach the cluster over Tailscale, and to troubleshoot when you can't.

**Status:** tested from a phone hotspot and cafe Wi-Fi on 2026-10-30. `kubectl` works. SSO login works, but web
apps stall intermittently (#19).

## How it works

The `homelab-router` pod is a Tailscale subnet router for `192.168.0.0/24`. Each device connecting from outside
needs Tailscale signed in to your tailnet. The servers don't need anything extra.

## Connect

From the repo root:

```sh
make routes-on
nix develop
export KUBECONFIG=$PWD/metal/kubeconfig.yaml
kubectl get nodes
```

`nix develop` prints nothing and opens a new shell. `KUBECONFIG` needs exporting in each new shell.

## Back at home

```sh
make routes-off
```

!!! warning

    On Linux, accepting routes sends all `192.168.0.x` traffic through the router, even at home. Leave routes
    off at home, especially for PXE imaging and failure tests. Otherwise your access to each node depends on a
    single pod.

## What to avoid while away

Avoid risky changes. A node that loses power can't be restarted remotely (#21, #25), and if it runs the router
you may lose access until the pod moves. If the failure needs the
[dead-node procedure](dead-node-recovery.md), you can't run it without access.

Do these from home: GPU driver installs, BIOS changes, PXE imaging and failure tests.

## If you can't connect

| Symptom | Likely cause | Check or fix |
|---|---|---|
| `connection refused` to `localhost:8080` | `KUBECONFIG` isn't set | `export KUBECONFIG=$PWD/metal/kubeconfig.yaml` |
| `kubectl: command not found` | Not in the Nix shell | `nix develop` |
| `nix develop` is refused | Not in the `nix-users` group yet | `sg nix-users -c "nix --extra-experimental-features 'nix-command flakes' develop"` |
| `kubectl` times out | Routes are off, or the network overlaps | See the checks below |
| Everything times out and the router is offline | Router pod or home network is down | See the checks below |
| Web apps stall but `kubectl` works | Known issue | #19 |

Check the route. It should say `dev tailscale0`:

```sh
tailscale status
ip route get 192.168.0.239
```

If it names your Wi-Fi interface, run `make routes-on`. If it still does, the local network may also use
`192.168.0.x`, so switch to a phone hotspot.

If `homelab-router` shows as offline in `tailscale status`, there's little you can do remotely, unless you have
another way into your home network. Check it when you're next home:

```sh
kubectl -n tailscale get pods -o wide
kubectl -n tailscale logs deploy/tailscale --tail=30
```

## Keys

`homelab-router`'s key was due to expire on 2027-04-28, and remote access stops if it expires. To prevent that, go to
[the Machines page](https://login.tailscale.com/admin/machines), open the `...` menu on `homelab-router` and
choose **Disable key expiry**.

To check expiry dates from a connected device:

```sh
tailscale status --json | python3 -c 'import json,sys
d=json.load(sys.stdin)
for n in [d["Self"]]+list(d.get("Peer",{}).values()): print(n.get("HostName"),n.get("KeyExpiry"))'
```

No date for a device means its key doesn't expire.
