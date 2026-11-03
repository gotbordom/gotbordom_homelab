# Backup and restore

## Prerequisites

Create an S3 bucket to store backups. You can use AWS S3, Minio, or
any other S3-compatible provider.

- For AWS S3, your bucket URL might look something like this:
  `https://s3.amazonaws.com/my-homelab-backup`.
- For Minio, your bucket URL might look something like this:
  `https://my-s3-host.example.com/homelab-backup`.

Follow your provider's documentation to create a service account with the
following policy (replace `my-homelab-backup` with your actual bucket name):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::my-homelab-backup",
        "arn:aws:s3:::my-homelab-backup/*"
      ]
    }
  ]
}
```

Save the access key and secret key to a secure location, such as a password
manager. While you're at it, generate a new password for Restic encryption and
save it there as well.

!!! example

    I use Minio for my homelab backups. Here's how I set it up:

    - Create a bucket named `homelab-backup`.
    - Create a service account under Identity -> Service Accounts -> Create
      Service Account:
        - Enable Restrict beyond user policy.
        - Paste the policy above.
        - Click Create and copy the access key and secret key
    - I also set up Minio replication to store backups in two locations: one in
      my house and one remotely.

## Add backup credentials to global secrets

Add the following to `external/terraform.tfvars`:

```hcl
extra_secrets = {
  restic-password = "xxxxxxxxxxxxxxxxxxxxxxxx"
  restic-s3-bucket = "https://s3.amazonaws.com/my-homelab-backup-xxxxxxxxxx"
  restic-s3-access-key = "xxxxxxxxxxxxxxxx"
  restic-s3-secret-key = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}
```

Then apply the changes:

```sh
make external
```

You may want to back up the `external/terraform.tfvars` file to a secure location as well.

## Add backup configuration for volumes

Backups are defined in git, next to the app they protect. Each app has a `templates/backup.yaml` with:

- An `ExternalSecret` that builds the Restic repository secret from the global secrets above.
- A VolSync `ReplicationSource` that sets the volume, schedule and retention.

See `platform/kanidm/templates/backup.yaml` for an example. To back up another volume, copy it, then change the volume
name and the schedule.

| Volume | Schedule (UTC) | Notes |
|---|---|---|
| `kanidm/data-kanidm-0` | 09:30 | SSO accounts and credentials |
| `jellyfin/jellyfin-settings` | 10:00 | Settings for all the media apps |

These aren't backed up: `jellyfin/jellyfin-data` (media and downloads) and `ollama/ollama-data` (models can be downloaded
again).

!!! warning

    VolSync backs up a whole volume, so keep large replaceable data on a separate volume from anything you want to back up.

A new `ReplicationSource` runs once as soon as it's created, then follows its schedule. It skips the backup if the volume
is empty.

## Check backups

```sh
kubectl get replicationsources -A
kubectl -n kanidm get replicationsource data-kanidm-0 -o jsonpath='{.status.latestMoverStatus.result} {.status.lastSyncTime}{"\n"}'
```

## Test a restore

Restore into a throwaway volume, so the app's real volume isn't touched:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata: {name: restore-test, namespace: kanidm}
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: standard-rwo
  resources: {requests: {storage: 1Gi}}
---
apiVersion: volsync.backube/v1alpha1
kind: ReplicationDestination
metadata: {name: restore-test, namespace: kanidm}
spec:
  trigger: {manual: restore-test-1}
  restic:
    repository: data-kanidm-0-backup-repository
    destinationPVC: restore-test
    copyMethod: Direct
```

When `status.latestMoverStatus.result` is `Successful`, mount `restore-test` in a temporary pod to check the files. This
was last done for Kanidm on 2026-11-01, including an SQLite integrity check of its database. Then delete the test
resources.

## Restore from backup

!!! danger

    This overwrites what's on the volume. Test on a throwaway volume first, and stop the app before restoring over its data.
    A real restore over a live volume hasn't been tried on this cluster yet.

Scale the app down, then create a `ReplicationDestination` as above, with `destinationPVC` set to the app's volume.

## Other backups

The nightly backups don't cover the files needed to recover the cluster itself, so also run `make recovery-backup`. See
[Recover from a dead laptop](runbooks/rebuild-from-dead-laptop.md).
