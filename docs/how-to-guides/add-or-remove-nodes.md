# Add or remove nodes

Or how to scale vertically. To replace the same node with a clean OS, remove it and add it again.

## Add new nodes

!!! tip

    You can add multiple nodes at the same time

Add its details to the inventory **at the end of the group** (masters or workers):

```diff title="metal/inventories/prod.yml"
diff --git a/metal/inventories/prod.yml b/metal/inventories/prod.yml
index 7f6474a..1bb2cbc 100644
--- a/metal/inventories/prod.yml
+++ b/metal/inventories/prod.yml
@@ -8,3 +8,4 @@ metal:
     workers:
       hosts:
         metal3: {ansible_host: 192.168.1.113, mac: '00:23:24:d1:f5:69', disk: sda, network_interface: eno1}
+        metal4: {ansible_host: 192.168.1.114, mac: '00:11:22:33:44:55', disk: sda, network_interface: eno1}
```

!!! note

    `network_interface` can also be the NIC's MAC address (e.g. `'00:11:22:33:44:55'`), which is
    easier when you don't know the kernel interface name yet. If the name doesn't exist on the
    machine, the installer ignores it and the node silently falls back to DHCP. Likewise `disk`
    must be the real device: NVMe drives are `nvme0n1`, not `sda`.

    The node's firmware must be in UEFI mode with network boot (IPv4) ahead of the internal disk.

Install the OS and join the cluster:

```
make metal
```

That's it!

## Remove a node

!!! danger

    It is recommended to remove nodes one at a time

Remove it from the inventory:

```diff title="metal/inventories/prod.yml"
diff --git a/metal/inventories/prod.yml b/metal/inventories/prod.yml
index 7f6474a..d12b50a 100644
--- a/metal/inventories/prod.yml
+++ b/metal/inventories/prod.yml
@@ -4,7 +4,6 @@ metal:
       hosts:
         metal0: {ansible_host: 192.168.1.110, mac: '00:23:24:d1:f3:f0', disk: sda, network_interface: eno1}
         metal1: {ansible_host: 192.168.1.111, mac: '00:23:24:d1:f4:d6', disk: sda, network_interface: eno1}
-        metal2: {ansible_host: 192.168.1.112, mac: '00:23:24:e7:04:60', disk: sda, network_interface: eno1}
     workers:
       hosts:
         metal3: {ansible_host: 192.168.1.113, mac: '00:23:24:d1:f5:69', disk: sda, network_interface: eno1}
```

Drain the node:

```sh
kubectl drain ${NODE_NAME} --delete-emptydir-data --ignore-daemonsets --force
```

Remove the node from the cluster

```sh
kubectl delete node ${NODE_NAME}
```

Shutdown the node:

```
ssh root@${NODE_IP} poweroff
```
