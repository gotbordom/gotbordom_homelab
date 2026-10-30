# Pause and restore kured

Use this to stop automatic node reboots during an outage, a failure test or risky maintenance.

**Status:** used for real on 2026-10-30, when kured started draining thinkcenter0 during the hpz3 failure test.

Kured reboots nodes between 09:00 and 12:00 UTC (see `system/kured/values.yaml`). It takes one reboot at a
time, and Ceph's disruption budgets can hold up a drain while Ceph is degraded. Once Ceph heals, a blocked
drain can continue, so a pause matters most just before you bring a node back.

## Pause

Deleting the DaemonSet isn't enough, because ArgoCD would restore it. Editing the `kured` Application doesn't
stick either, because the ApplicationSet overwrites it. A deny sync window on the project avoids both.

Check for existing windows. This should print a blank line:

```sh
kubectl -n argocd get appproject default -o jsonpath='{.spec.syncWindows}{"\n"}'
```

!!! warning

    If it prints anything else, stop. The commands below replace and remove the whole `syncWindows` list,
    so they'd overwrite those windows. Merge them by hand.

Add the window, then remove kured. Removing the DaemonSet also releases kured's reboot lock:

```sh
kubectl -n argocd patch appproject default --type merge -p '{"spec":{"syncWindows":[{"kind":"deny","schedule":"* * * * *","duration":"24h","applications":["kured"],"manualSync":false}]}}'
kubectl -n kured delete ds kured
```

If kured was draining a node, clear its marker and put the node back in service:

```sh
kubectl annotate node ${NODE_NAME} weave.works/kured-reboot-in-progress-
kubectl uncordon ${NODE_NAME}
```

Wait a minute, then confirm it stayed gone. ArgoCD should show kured as `OutOfSync` and `Missing`. Any pod
left on a dead node stays `Terminating` until that node is dealt with:

```sh
kubectl -n kured get ds,pods
kubectl -n argocd get application kured
```

The window stays in force until you remove it, because it reopens every minute. The change is only on the
cluster, not in git, so don't forget it.

## Restore

If you've changed kured's settings, merge that first and wait for ArgoCD to show the new revision. The deny
window can delay this, so don't wait more than a few minutes:

```sh
kubectl -n argocd get application kured -o jsonpath='{.status.sync.revision}{"\n"}'
```

Remove the window:

```sh
kubectl -n argocd patch appproject default --type json -p '[{"op":"remove","path":"/spec/syncWindows"}]'
```

After a minute or two, check it came back with its reboot window and no lock:

```sh
kubectl -n kured get ds kured -o jsonpath='{.spec.template.spec.containers[0].args}{"\n"}'
kubectl -n kured get ds kured -o jsonpath='lock={.metadata.annotations.weave\.works/kured-node-lock}{"\n"}'
kubectl -n argocd get application kured
```

The args should include `--start-time=09:00`, and the app should be `Synced` and `Healthy`.

Any reboot that's still pending will run in the next window. To see which nodes need one:

```sh
for p in $(kubectl -n kured get pods -o name); do echo "$p"; kubectl -n kured logs $p | grep -E 'Reboot is required|Reboot should not' | tail -1; done
```
