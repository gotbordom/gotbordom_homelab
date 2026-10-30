# Recover from control plane loss

Use this when thinkcenter0, the only control plane, is down.

**Status:** untested. This is a draft based on how k3s works and hasn't been run on this cluster. Update it after the
thinkcenter0 failure test (#23). Read each step before running it.

## What to expect

- `kubectl` and the API VIP (`192.168.0.239`) stop working, and SSH to thinkcenter0 fails.
- Nothing can be rescheduled, so the [dead-node procedure](dead-node-recovery.md) can't be used.
- thinkcenter0's Ceph mon and OSD go down. The other 2 mons should keep quorum, and the data should stay available.
- Pods on hpz3 and optiplex0 should keep running, but this hasn't been tested. Anything that needs the API or a pod on
  thinkcenter0 may fail.

## Before this happens

!!! warning

    k3s takes etcd snapshots automatically, at about 00:00 and 12:00 UTC. They're stored **only on thinkcenter0**,
    along with the token needed to restore them, so a dead disk loses both. There's currently no off-node copy (#24).

To list snapshots:

```sh
kubectl get etcdsnapshotfiles.k3s.cattle.io
```

To copy them and the token somewhere safe. Treat them as secrets, and never commit them:

```sh
scp -r root@192.168.0.9:/var/lib/rancher/k3s/server/db/snapshots ./etcd-snapshots
scp root@192.168.0.9:/var/lib/rancher/k3s/server/token ./k3s-token
```

## Case 1: the machine is off or hung

The disk should be fine, so this is the simple case. Don't restore anything.

1. Power it on or restart it. Check that it boots from disk. Keep the PXE server off.
2. Wait for the API, then check the cluster:

    ```sh
    kubectl get nodes
    kubectl -n rook-ceph exec deploy/rook-ceph-operator -- ceph --conf /var/lib/rook/rook-ceph/rook-ceph.config -s
    kubectl -n argocd get applications
    ```

3. If the API doesn't come back, check k3s on the node:

    ```sh
    ssh root@192.168.0.9 'systemctl status k3s --no-pager; journalctl -u k3s -n 50 --no-pager'
    ```

## Case 2: the machine runs, but etcd is damaged

!!! danger

    Restoring rolls the whole cluster back to the snapshot's time. Anything since then is lost from cluster state.
    Try Case 1 first, and copy `/var/lib/rancher/k3s/server/db` somewhere safe before you start.

On thinkcenter0:

```sh
systemctl stop k3s
cp -a /var/lib/rancher/k3s/server/db /root/k3s-db-backup
ls /var/lib/rancher/k3s/server/db/snapshots
k3s server --cluster-reset --cluster-reset-restore-path=/var/lib/rancher/k3s/server/db/snapshots/<snapshot-name>
systemctl start k3s
```

Follow the messages from `--cluster-reset`, since they take priority over this page. If k3s doesn't start with the same
settings, check `/etc/rancher/k3s/config.yaml` and `/etc/systemd/system/k3s.service`.

Then restart k3s on the workers and check the cluster as in Case 1:

```sh
ssh root@192.168.0.12 systemctl restart k3s
ssh root@192.168.0.13 systemctl restart k3s
```

## Case 3: the disk is dead

**If you have an off-node snapshot and token:** reinstall thinkcenter0, restore the token and snapshot to their usual
locations before k3s starts, then follow Case 2. This is the least tested path, so expect to adapt it.

**If you don't:** the cluster state is gone, so the cluster has to be rebuilt:

1. Reinstall the control plane and rejoin the workers with `make metal`.
2. Restore the rest with `make system`, `make external` and `make post-install`.
3. Volumes will need restoring from backups, and **no volume backups exist yet** (#26). Data on the surviving Ceph OSDs
   may be recoverable, but that needs expert recovery work, so don't wipe those disks until you've decided.

## Afterwards

Follow up on anything that didn't match this page, and update it. Check that no node is `SchedulingDisabled`, that
Ceph is `HEALTH_OK`, and whether kured is holding a stale lock (see [Pause and restore kured](pause-kured.md)).
