package tests

import (
	"context"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/astrolabe/pkg/ivd"
	v1api "github.com/vmware-tanzu/velero-plugin-for-vsphere/pkg/apis/veleroplugin/v1"
	pluginversioned "github.com/vmware-tanzu/velero-plugin-for-vsphere/pkg/generated/clientset/versioned"
	"github.com/vmware-tanzu/velero-plugin-for-vsphere/pkg/utils"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/cns"
	vim "github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vslm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"time"

	restclient "k8s.io/client-go/rest"
)

func GetVStorageObjectManager(ctx context.Context, config *restclient.Config) (*vslm.GlobalObjectManager, error) {
	client, err := GetGoVmomiClient(ctx, config)
	if err != nil {
		return nil, errors.Errorf("Failed to get a new govmomi client: %+v", err)
	}

	vslmClient, err := vslm.NewClient(ctx, client.Client)
	if err != nil {
		return nil, errors.Errorf("Failed to get a new vslm govmomi client: %+v", err)
	}
	vsom := vslm.NewGlobalObjectManager(vslmClient)

	return vsom, nil
}

func GetCnsClient(ctx context.Context, config *restclient.Config) (*cns.Client, error) {
	client, err := GetGoVmomiClient(ctx, config)
	if err != nil {
		return nil, errors.Errorf("Failed to get a new govmomi client: %+v", err)
	}

	cnsClient, err := cns.NewClient(ctx, client.Client)
	if err != nil {
		return nil, errors.Errorf("Failed to get a new cns govmomi client: %+v", err)
	}


	return cnsClient, nil
}

func GetGoVmomiClient(ctx context.Context, config *restclient.Config) (*govmomi.Client, error) {
	params := make(map[string]interface{})

	err := utils.RetrieveVcConfigSecretFromConfig(config, params, nil)
	if err != nil {
		return nil, errors.Errorf("Failed to retrieve VC config secret: %+v", err)
	}

	vcUrl, insecure, err := utils.GetVcConfigFromParams(params)
	if err != nil {
		return nil, errors.Errorf("Failed to get VC config from params: %+v", err)
	}

	client, err := govmomi.NewClient(ctx, vcUrl, insecure)
	if err != nil {
		return nil, errors.Errorf("Failed to get a new govmomi client: %+v", err)
	}

	return client, nil
}

func WatchUploadRequests(config *restclient.Config, veleroNs string, nRequests int) error {
	pluginClientset, err := pluginversioned.NewForConfig(config)
	if err != nil {
		return errors.Errorf("Failed to get velero plugin clientset from k8s config: %+v", err)
	}
	// add upload watcher
	watcher, err := pluginClientset.VeleropluginV1().Uploads(veleroNs).Watch(metav1.ListOptions{})
	if err != nil {
		return errors.Errorf("Failed to watch the customized resource of plugin: %+v", err)
	}
	defer watcher.Stop()

	var nResponses int

	err = wait.PollImmediate(time.Second, 10*time.Minute, func() (bool, error) {
		select {
		case e := <-watcher.ResultChan():
			updated, ok := e.Object.(*v1api.Upload)
			if !ok {
				return false, errors.Errorf("unexpected type %T", e.Object)
			}

			switch e.Type {
			case watch.Deleted:
				return false, errors.New("upload request was unexpectedly deleted")
			case watch.Modified:
				if updated.Status.Phase == v1api.UploadPhaseFailed {
					return false, errors.Errorf("upload request was failed with error message: %v", updated.Status.Message)
				}
				if updated.Status.Phase == v1api.UploadPhaseCompleted {
					nResponses += 1
					if nResponses == nRequests {
						break
					}
				}
			}
		}
		return true, nil
	})

	if err != nil {
		return err
	}

	return nil
}

func CheckVolumeNamesForPVCs(config *restclient.Config, AppNs string) ([]vim.ID, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Errorf("Failed to get k8s clientset from k8s config: %+v", err)
	}
	_, err = clientset.CoreV1().Namespaces().Get(AppNs, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Errorf("Failed to find the specified namespace, %v: %+v ", AppNs, err)
	}

	pvcList, err := clientset.CoreV1().PersistentVolumeClaims(AppNs).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Errorf("Failed to list PVCs in the specified namespace, %v: %+v", AppNs, err)
	}

	nPVCs := len(pvcList.Items)
	if nPVCs < 1 {
		return nil, errors.Errorf("Number of PVC available is less than expected: %v", len(pvcList.Items))
	}
	var volumeVimIDs []vim.ID
	for _, item := range pvcList.Items {
		pv, err := clientset.CoreV1().PersistentVolumes().Get(item.Spec.VolumeName, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Errorf("Got error: %v ", err)
		}
		if pv.Spec.CSI.Driver != "csi.vsphere.vmware.com" {
			return nil, errors.Errorf("Unexpected csi driver: %v", pv.Spec.CSI.Driver)
		}
		volumeVimIDs = append(volumeVimIDs, ivd.NewIDFromString(pv.Spec.CSI.VolumeHandle))
	}

	ctx := context.Background()
	vsom, err := GetVStorageObjectManager(ctx, config)
	if err != nil {
		return nil, errors.Errorf("Failed to get VStorageObjectManager: %+v", err)
	}

	vsoResults, err := vsom.RetrieveObjects(ctx, volumeVimIDs)
	if err != nil {
		return nil, errors.Errorf("Failed to retrieve the volume, %+v: %+v", volumeVimIDs, err)
	}
	for _, item := range vsoResults {
		//t.Logf("The name of volume, %v: %v", item.Id.Id, item.Name)
		if item.Name == "ivd-created" {
			return nil, errors.Errorf("Unexpected name, %v, of volume, %v", item.Name, item.Id.Id)
		}
	}

	return volumeVimIDs, nil
}
