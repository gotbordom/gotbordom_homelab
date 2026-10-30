# Recover from a dead worker node

Use this when a worker has lost power or died and apps with volumes are stuck.

**Status:** used for real in the hpz3 failure test on 2026-10-30. All 4 volumes moved, and a checksummed
test file was intact on its new node. For the control plane, see
[Recover from control plane loss](control-plane-loss.md).

## What you'll see

Ceph keeps serving data on its own, but pods that use a block volume don't recover without help:

- The node is `NotReady`, and its old pods are stuck `Terminating`.
- Replacement pods are stuck in `ContainerCreating` with a `Multi-Attach` error.
- StatefulSet pods such as `kanidm-0` have no replacement at all, so SSO is down.
- `kubectl get volumeattachments` still lists the dead node.

Waiting doesn't fix this. In the hpz3 test, the volumes were still attached after 70 minutes.

## Should you wait instead?

If the node will be back within minutes, plug it in or power it on and skip the rest. Its pods and volumes
come back on the same node.

## Recover

!!! danger

    Only continue if the node is **physically off**. This force-deletes its pods and detaches its volumes.
    If it's actually running, two nodes could write to the same volume and corrupt it.

Check that the node really is off. This must print `DOWN`, and a failed ping alone isn't proof, so confirm
the power state if you can:

```sh
export NODE_NAME=hpz3 NODE_IP=192.168.0.12
ping -c3 -W2 ${NODE_IP} >/dev/null && echo "UP - STOP" || echo DOWN
```

Check that kured isn't draining another node, since that can strand even more pods:

```sh
kubectl get nodes
kubectl -n kured get ds kured -o jsonpath='{.metadata.annotations.weave\.works/kured-node-lock}{"\n"}'
```

If a healthy node shows `SchedulingDisabled` or the lock names a node, follow
[Pause and restore kured](pause-kured.md) first.

Add the taint:

```sh
kubectl taint node ${NODE_NAME} node.kubernetes.io/out-of-service=nodeshutdown:NoExecute
```

Check recovery. Volumes should move within minutes. Dex may take several more minutes to start after Kanidm:

```sh
kubectl get volumeattachments -o custom-columns=PV:.spec.source.persistentVolumeName,NODE:.spec.nodeName,ATTACHED:.status.attached
kubectl get pods -A -o wide | grep -vE 'Running|Completed'
kubectl -n argocd get applications
```

Pods tied to the dead node, such as its Ceph `mon` and `osd`, stay `Pending`. That's expected.

## Bring the node back

!!! warning

    Leave the taint on until the node is back and clean. It keeps workloads off the node while it boots.

1. Power the node on. It may need the power button (see the AC Recovery and Wake-on-LAN issues, #21 and #25).
2. Wait for `Ready`, then check it has no old pods or volume mounts:

    ```sh
    kubectl get node ${NODE_NAME}
    kubectl get pods -A -o wide --field-selector spec.nodeName=${NODE_NAME}
    ssh root@${NODE_IP} 'rbd showmapped; mount | grep -c rbd'
    ```

    The pod list and `rbd showmapped` should be empty, and the mount count should be `0`.

3. Remove the taint. The `-` on the end removes it:

    ```sh
    kubectl taint node ${NODE_NAME} node.kubernetes.io/out-of-service=nodeshutdown:NoExecute-
    ```

4. Confirm Ceph heals. In the hpz3 test, it reached `HEALTH_OK` about a minute after the OSD rejoined:

    ```sh
    kubectl -n rook-ceph exec deploy/rook-ceph-operator -- ceph --conf /var/lib/rook/rook-ceph/rook-ceph.config -s
    ```

    You want `HEALTH_OK`, all placement groups `active+clean`, and 3 mons in quorum.

Moved pods stay on their new nodes. That's fine.
