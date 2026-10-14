# Homelab Build-Out Roadmap

**Fork of khuedoan/homelab, adapted to gotbordom's own hardware (thinkcenter0, hpz3, optiplex0).** Evaluated 2026-10-14 against four goals: automated imaging of all inventory machines; a media server (movies/music/books); local AI training + chat; and game servers (Screeps). Status tags reflect what was actually verified in the repo, not assumed.

## Legend

- Exists: built and wired up, ready to use once the cluster is up
- Partial: something real exists but doesn't fully cover the goal
- Gap: nothing exists, needs to be built from scratch

## Phase 0 -- Fix the foundation (blocking, do first)

Nothing below matters until the cluster reliably comes up and stays healthy.

- [ ] Fix the subnet mismatch in metal/inventories/prod.yml -- control_plane_endpoint (192.168.1.100) and load_balancer_ip_pool (10.0.0.1/24) are leftovers from before the hardware migration; neither matches the master's actual IP (192.168.0.9). kube-vip needs all three on the same L2 segment.
- [ ] Re-enable a second real node (hpz3 or optiplex0) in the inventory. Only thinkcenter0 is active right now, so Rook-Ceph's size:2 replication (system/rook-ceph/values.yaml) has nowhere to put a second copy -- a disk failure today loses data, full stop.
- [ ] Dry-run the recovery story once: pull a node, wipe it, run make metal, confirm PXE reimage, k3s rejoin, ArgoCD resync, and Ceph rebalance all happen unattended. Don't trust the auto-magic-reimage plan until it's been exercised for real.
- [ ] Confirm optiplex0's disk: value -- the working tree currently disagrees with itself (nvme vs sda) across uncommitted edits.

(See Progress log below -- the config side of this phase is done; the dry-run is still open.)

## Already built -- ready to use

**Bare-metal imaging and cluster bring-up**

- PXE boot plus a per-host kickstart generated dynamically from metal/inventories/prod.yml (MAC, IP, hostname, disk, interface), wiping the disk, installing Fedora Server, and carving a raw partition for Ceph.
- k3s install and join via Ansible (metal/roles/k3s), kube-vip for the HA control-plane VIP, Cilium as CNI.
- ArgoCD app-of-apps via a directory-generator ApplicationSet watching system/*, platform/*, apps/* in this fork -- drop a new chart under apps/ and it deploys itself.

**Shared platform resources**

- Rook-Ceph: shared block and file storage (system/rook-ceph)
- Zot: private container registry (platform/zot)
- Gitea + Woodpecker: git hosting and CI for anything you build yourself (platform/gitea, platform/woodpecker)
- external-secrets + global-secrets: centralized secret management (platform/external-secrets, platform/global-secrets)
- cert-manager + external-dns + cloudflared: automatic TLS, DNS, and secure internet exposure (system/cert-manager, system/external-dns, system/cloudflared)
- Tailscale / Wireguard: remote access VPN (apps/tailscale, apps/wireguard)

**Identity and SSO skeleton**

- Kanidm as the identity provider (platform/kanidm), Dex as the OIDC broker fanning it out to apps (platform/dex) -- Grafana and Gitea already log in through it.
- scripts/onboard-user plus docs/getting-started/user-onboarding.md: one command creates a person and a credential-reset link for passkeys or password+TOTP.

**Analytics and observability**

- kube-prometheus-stack (system/monitoring-system): Prometheus and Alertmanager, already piping alerts to ntfy.
- Loki (system/loki): log aggregation.
- Grafana (platform/grafana): dashboards auto-provisioned via sidecar discovery, SSO'd through Dex -- this is your log-in-and-check-the-lab window, already real.

**Media server -- movies and TV**

- apps/jellyfin bundles Jellyfin, Radarr, Sonarr, Prowlarr, Transmission, and Jellyseerr, with ingress and persistent storage already wired. docs/how-to-guides/media-management.md walks through setup. This is complete.

**Your own apps**

- The Gitea to Woodpecker to Zot to apps/ to ArgoCD pipeline is a full build-your-own-app-and-ship-it path, already proven by the blog, excalidraw, and pairdrop apps living in this repo today. Adding one of your own apps is writing a chart under apps/<name>/, nothing more.

## Partial -- exists, needs extension

- **Permission tiers.** Kanidm and Dex give real SSO, but scripts/onboard-user puts every new person into a single "editor" group -- the script's own comment says "TODO better group management." There is no admin/viewer/guest tiering today; you would design and wire up Kanidm groups mapped to per-app roles yourself.
- **AI chat / inference.** apps/ollama runs Ollama plus Open WebUI and works today, but there is zero GPU plumbing anywhere in the repo: no driver install in metal/roles/prerequisites, no nvidia-container-toolkit, no containerd runtimeClass, no device plugin. It only runs on CPU until that is built, and it should be scoped to hpz3 specifically via node labels and taints so a future GPU-less node doesn't break.
- **Music and books in the media stack.** Jellyfin can serve a music library out of the box, but there is no acquisition automation for it the way Radarr/Sonarr cover movies and TV (no Lidarr). Books are not covered by anything here at all.

## Not built -- greenfield

- **AI training.** Nothing in the repo does this -- apps/ollama is inference-only. This needs a real framework (Jupyter, PyTorch, or similar) as a new app chart, plus the same GPU plumbing the chat bot needs.
- **Book server.** No ebook or audiobook app exists. This means adding something like Calibre-Web or Kavita for ebooks, and/or Audiobookshelf, as new apps/ charts, following the same pattern as Jellyfin.
- **Game servers (Screeps).** Nothing exists. Screeps is not a common Helm-charted app -- a private Screeps server is a Node.js process plus MongoDB, so this would be a bespoke apps/screeps chart you build, pointed at an image pushed through the Gitea/Woodpecker/Zot pipeline you already have.

## Testing roadmap

**What exists today** (test/, Terratest in Go):

- tools_test.go: checks your dev-shell tool versions match what's pinned
- external_test.go: terraform init/validate on external/
- integration_test.go: confirms ArgoCD's own ingress returns 200
- smoke_test.go: six core apps (argocd, gitea, grafana, homepage, kanidm, zot) return 200 over their ingress

**What's missing, roughly in priority order:**

- [ ] Smoke coverage for the workloads you actually care about -- jellyfin and ollama aren't in smoke_test.go's app list at all
- [ ] A Ceph health/redundancy check (assert pool size and OSD count match expectations) -- directly relevant to the Phase 0 redundancy gap
- [ ] A real backup/restore test -- scripts/backup is wired for actualbudget and jellyfin only, and nothing currently proves a restore actually recovers data
- [ ] Any test of the metal/ layer itself (PXE, kickstart, node join) -- today that path is only ever exercised by hand
- [ ] A GPU scheduling test once that work lands (pod actually gets the GPU on hpz3, doesn't get scheduled onto a GPU-less node)

Note: the upstream project's own docs/reference/roadmap.md still lists "Automated testing" and "Chaos testing" under Unplanned -- this isn't a gap specific to your fork, it's an acknowledged upstream gap too.

(See Progress log below -- jellyfin/ollama smoke coverage and the Ceph redundancy check have since been added.)

## Suggested sequencing

1. Phase 0 (subnet fix, second node, dry-run recovery) -- unblocks everything else and validates the auto-magic-recovery promise you're actually counting on
2. Expand smoke tests to cover jellyfin and ollama now, cheaply, while the apps are fresh
3. Media server -- effectively done once Phase 0 lands, just walk docs/how-to-guides/media-management.md
4. Tiered permissions in Kanidm/Dex -- needed before you let other people register at different access levels
5. GPU plumbing on hpz3 (driver, nvidia-container-toolkit, device plugin, node labels/taints), validated first against Ollama before touching training
6. AI training app on top of the now-working GPU path
7. Screeps and a book server as self-contained apps -- lowest-risk, most parallelizable work, can happen anytime alongside the above

## Progress log

**2026-10-14 -- Phase 0, config side done.** Subnet mismatch fixed (control_plane_endpoint and load_balancer_ip_pool now derive from one metal_subnet_prefix var instead of two independently hardcoded, now-wrong values). hpz3 and optiplex0 both re-enabled as workers (disk: sda confirmed for both), giving Rook-Ceph three nodes to replicate across. Added TestInventorySubnetConsistency and TestCephReplicationFitsNodeCount / TestCephMonCountFitsNodeCount (static, no live cluster needed) plus smoke coverage for jellyfin/ollama, all wired into static-checks CI. Decided not to build disk auto-detection (a %pre kickstart script that picks the right device automatically) for now -- manual disk: field is an acceptable tradeoff while there's no test harness for the metal/ provisioning layer itself; revisit if it bites again.

Opened as PR #1 (branch `phase0` -> `master`) with a validation checklist to run once home: confirm the guessed subnet range doesn't collide with DHCP, run the new tests for the first time for real, run `make metal` across all three nodes, verify the jellyfin/ollama ingress-name guess, check Ceph health, do the actual dry-run, and physically confirm optiplex0's disk.

Still open from Phase 0: the actual dry-run (pull a node, wipe it, run make metal, confirm it all heals) -- that's a physical-hardware step for the user to run, not something done from this environment. Everything past Phase 0 (permission tiers, GPU plumbing, AI training, game servers, book server) is unstarted.

Live doc: https://claude.ai/artifact/TbXYEro1f7f3v6FrE3hjeC
