# Velero vSphere operator CLI
## Install Velero & Plugins

**Tips**: users are expected to create a WCP supervisor namespace either via vSphere UI or API before running the command
below to install Velero. Otherwise, the installed velero pod will be stuck at `Pending` state.

```
Install Velero Instance

Usage:
  velero-vsphere install [flags]

Flags:
      --backup-location-config mapStringString     configuration to use for the backup storage location. Format is key1=value1,key2=value2
      --bucket string                              name of the object storage bucket where backups should be stored
      --enable-cbt-in-guests                       whether or not to enable the ChangeBlockTracking option in guest clusters. Optional.
  -h, --help                                       help for install
      --image string                               image to use for the Velero server pods. Optional. (default "velero/velero:v1.5.1")
  -n, --namespace string                           the namespace where to install Velero. Optional. (default "velero")
      --no-default-backup-location                 flag indicating if a default backup location should be created. Must be used as confirmation if --bucket or --provider are not provided. Optional.
      --no-secret                                  flag indicating if a secret should be created. Must be used as confirmation if --secret-file is not provided. Optional.
      --plugins stringArray                        plugin container images to install into the Velero Deployment
      --provider string                            provider name for backup and volume storage
      --secret-file string                         file containing credentials for backup and volume provider. If not specified, --no-secret must be used for confirmation. Optional.
      --snapshot-location-config mapStringString   configuration to use for the volume snapshot location. Format is key1=value1,key2=value2
      --upgrade-option string                      upgrade option: manual or automatic. Optional. (default "Manual")
      --use-private-registry                       whether or not to pull instance images from a private registry. Optional
      --use-volume-snapshots                       whether or not to create snapshot location automatically. Optional (default true)
      --version string                             version for velero to be installed. Optional. (default "v1.5.1")

Global Flags:
      --enable-leader-election   Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.
      --kubeconfig string        Paths to a kubeconfig. Only required if out-of-cluster.
      --master --kubeconfig      (Deprecated: switch to --kubeconfig) The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.
      --webhook-port int         Webhook server port (set to 0 to disable)
```
Below are some examples.
1. Installing Velero with default backup location and snapshot location
    ```
    velero-vsphere install \
           --namespace $NAMESPACE \
           --version v1.5.1 \
           --provider aws \
           --plugins velero/velero-plugin-for-aws:v1.1.0,dpcpinternal/velero-plugin-for-vsphere:master-8b811eb-31.Aug.2020.00.17.55 \
           --bucket $BUCKET \
           --secret-file ~/.aws/credentials \
           --snapshot-location-config region=$REGION \
           --backup-location-config region=$REGION
    ```
2. Installing Velero without default backup location and snapshot location
    ```
    velero-vsphere install \
        --version v1.5.1 \
        --plugins dpcpinternal/velero-plugin-for-vsphere:master-8b811eb-31.Aug.2020.00.17.55 \
        --no-secret \
        --use-volume-snapshots=false \
        --no-default-backup-location
    ```
3. Install Velero in the an Air-gap environment
    ```
    velero-vsphere install \
        --namespace velero \
        --image <private registry name>/velero:v1.5.1 --use-private-registry \
        --provider aws \
        --plugins <private registry name>/velero-plugin-for-aws:v1.1.0,<private registry name>/velero-plugin-for-vsphere:master-ee03655-23.Oct.2020.00.33.51 \
        --bucket $BUCKET \
        --secret-file ~/.minio/credentials \
        --snapshot-location-config region=$REGION \
        --backup-location-config region=$REGION,s3ForcePathStyle="true",s3Url=$S3URL
    ```

## Uninstall Velero & Plugins
```
Uninstall Velero Instance

Usage:
  velero-vsphere uninstall [flags]

Flags:
  -h, --help               help for uninstall
  -n, --namespace string   The namespace of Velero instance. Optional. (default "velero")
```

Below is an example,
```
velero-vsphere uninstall -n non-velero
```
**Tips**: users are expected to remove the corresponding WCP supervisor namespace either via vSphere UI or API after
running the command above to uninstall Velero.

## Configure Velero & Plugins
Currently, only toggling CBT configuration in guest clusters (vSphere specific) are supported in the **configure** command.
Going forward, there might be more to be added.

```
Configure backup option(s)

Usage:
  velero-vsphere configure [flags]

Flags:
      --enable-cbt-in-guests   Whether or not to enable the ChangeBlockTracking option in guest clusters. Optional.
  -h, --help                   help for configure
  -n, --namespace string       The namespace of Velero instance. Optional. (default "velero")
```

Below are some examples. To enable CBT for a running Velero and vSphere plugins,
```
velero-vsphere configure --enable-cbt-in-guests -n non-velero
```
To disable CBT, the following command is expected.
```
velero-vsphere configure --enable-cbt-in-guests=false -n non-velero
```

More details on configuring CBT can be found [here](ChangeBlockTracking.md).
