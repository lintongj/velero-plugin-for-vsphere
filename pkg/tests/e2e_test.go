package tests

import (
	"context"
	"errors"
	"fmt"
	"github.com/vmware-tanzu/velero-plugin-for-vsphere/pkg/utils"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/builder"
	"github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned"
	"github.com/vmware/govmomi/cns"
	"github.com/vmware/govmomi/find"
	//"github.com/vmware/govmomi/view"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"testing"
	"time"
	cnstypes "github.com/vmware/govmomi/cns/types"
	vim25types "github.com/vmware/govmomi/vim25/types"
)

func TestProvisionPVUsingCnsAPI(t *testing.T) {
	path := os.Getenv("HOME") + "/.kube/config"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path/to/whatever does not exist
		t.Skipf("The KubeConfig file, %v, is not exist", path)
	}

	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		t.Fatalf("Failed to build k8s config from kubeconfig file: %+v ", err)
	}

	ctx := context.Background()
	client, err := GetGoVmomiClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to get govmomi client: %+v", err)
	}

	finder := find.NewFinder(client.Client)
	dss, err := finder.DatastoreList(ctx, "*")
	if err != nil {
		t.Fatal("Failed to list all datastores")
	}

	var dsList []vim25types.ManagedObjectReference
	for _, ds := range dss {
		dsList = append(dsList, ds.Reference())
	}

	t.Logf("Datastore List: %v", dss)

	cnsClient, err := GetCnsClient(ctx, config)
	if err != nil {
		t.Fatalf("Failed to get cns client: %+v", err)
	}

	// retrieve cluster id and vsphere user from VC
	params := make(map[string]interface{})
	err = utils.RetrieveVcConfigSecretFromConfig(config, params, nil)
	if err != nil {
		t.Fatalf("Failed to retrieve VC config secret: %+v", err)
	}

	cluster_id, ok := params["cluster-id"].(string)
	if !ok {
		t.Fatalf("Failed to retrieve cluster id")
	}

	user, ok := params["user"].(string)
	if !ok {
		t.Fatalf("Failed to retrieve vsphere user")
	}

	// create volume
	var cnsVolumeCreateSpecList []cnstypes.CnsVolumeCreateSpec
	cnsVolumeCreateSpec := cnstypes.CnsVolumeCreateSpec{
		Name:       "xyz",
		VolumeType: string(cnstypes.CnsVolumeTypeBlock),
		Datastores: dsList,
		Metadata: cnstypes.CnsVolumeMetadata{
			ContainerCluster: cnstypes.CnsContainerCluster{
				ClusterType: string(cnstypes.CnsClusterTypeKubernetes),
				ClusterId:   cluster_id,
				VSphereUser: user,
			},
		},
		BackingObjectDetails: &cnstypes.CnsBlockBackingDetails{
			CnsBackingObjectDetails: cnstypes.CnsBackingObjectDetails{
				CapacityInMb: 50,
			},
		},
	}
	cnsVolumeCreateSpecList = append(cnsVolumeCreateSpecList, cnsVolumeCreateSpec)
	t.Logf("Creating volume using the spec: %+v", cnsVolumeCreateSpec)
	createTask, err := cnsClient.CreateVolume(ctx, cnsVolumeCreateSpecList)
	if err != nil {
		t.Errorf("Failed to create volume. Error: %+v \n", err)
		t.Fatal(err)
	}
	createTaskInfo, err := cns.GetTaskInfo(ctx, createTask)
	if err != nil {
		t.Errorf("Failed to create volume. Error: %+v \n", err)
		t.Fatal(err)
	}
	createTaskResult, err := cns.GetTaskResult(ctx, createTaskInfo)
	if err != nil {
		t.Errorf("Failed to create volume. Error: %+v \n", err)
		t.Fatal(err)
	}
	if createTaskResult == nil {
		t.Fatalf("Empty create task results")
		t.FailNow()
	}
	createVolumeOperationRes := createTaskResult.GetCnsVolumeOperationResult()
	if createVolumeOperationRes.Fault != nil {
		t.Fatalf("Failed to create volume: fault=%+v", createVolumeOperationRes.Fault)
	}
	volumeId := createVolumeOperationRes.VolumeId.Id

	t.Logf("Volume created sucessfully. volumeId: %s", volumeId)

	var volumeIDList []cnstypes.CnsVolumeId
	volumeIDList = append(volumeIDList, cnstypes.CnsVolumeId{Id: volumeId})

	defer func () {
		// Test DeleteVolume API
		t.Logf("Deleting volume: %+v", volumeIDList)
		deleteTask, err := cnsClient.DeleteVolume(ctx, volumeIDList, true)
		if err != nil {
			t.Errorf("Failed to delete volume. Error: %+v \n", err)
			t.Fatal(err)
		}
		deleteTaskInfo, err := cns.GetTaskInfo(ctx, deleteTask)
		if err != nil {
			t.Errorf("Failed to delete volume. Error: %+v \n", err)
			t.Fatal(err)
		}
		deleteTaskResult, err := cns.GetTaskResult(ctx, deleteTaskInfo)
		if err != nil {
			t.Errorf("Failed to detach volume. Error: %+v \n", err)
			t.Fatal(err)
		}
		if deleteTaskResult == nil {
			t.Fatalf("Empty delete task results")
		}
		deleteVolumeOperationRes := deleteTaskResult.GetCnsVolumeOperationResult()
		if deleteVolumeOperationRes.Fault != nil {
			t.Fatalf("Failed to delete volume: fault=%+v", deleteVolumeOperationRes.Fault)
		}
		t.Logf("Volume deleted sucessfully")
	} ()

	// Test QueryVolume API
	var queryFilter cnstypes.CnsQueryFilter

	queryFilter.VolumeIds = volumeIDList
	t.Logf("Calling QueryVolume using queryFilter: %+v", queryFilter)
	queryResult, err := cnsClient.QueryVolume(ctx, queryFilter)
	if err != nil {
		t.Errorf("Failed to query volume. Error: %+v \n", err)
		t.Fatal(err)
	}
	t.Logf("Sucessfully Queried Volumes. queryResult: %+v", queryResult)
}

func TestVolumeNameIsRestoredAsExpected(t *testing.T) {
	path := os.Getenv("HOME") + "/.kube/config"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path/to/whatever does not exist
		t.Skipf("The KubeConfig file, %v, is not exist", path)
	}

	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		t.Fatalf("Failed to build k8s config from kubeconfig file: %+v ", err)
	}

	// Step 0: Check the existence and prerequisite of the app namespace
	t.Logf("Step 0: Check the existence and prerequisite of the app namespace")
	// Step 1: Check the volume name using the govmomi vslm API.
	t.Logf("Step 1: Check the names of volumes are not hard coded using the govmomi vslm API")
	AppNs := "demo-app"
	volumeIds, err := CheckVolumeNamesForPVCs(config, AppNs)
	if err != nil {
		t.Fatalf("Unexpected volume names were identified: %+v", err)
	}
	nPVCs := len(volumeIds)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to get k8s clientset from k8s config: %+v", err)
	}
	// Step 2: Backup the app namespace using Velero API and make the upload of the backup is completed as expected
	t.Logf("Step 2: Backup the app namespace using Velero API and make the upload of the backup is completed as expected")
	veleroNs := "velero"
	_, err = clientset.CoreV1().Namespaces().Get(AppNs, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to find the specified namespace, %v: %+v ", AppNs, err.Error())
	}

	deployment, err := clientset.AppsV1().Deployments(veleroNs).Get("velero", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get the velero deployment: %+v", err)
	}
	if deployment.Status.Replicas != deployment.Status.AvailableReplicas {
		t.Fatalf("The velero deployment is not up and running as expected")
	}

	daemonset, err := clientset.AppsV1().DaemonSets(veleroNs).Get("datamgr-for-vsphere-plugin", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get the data manager daemonset: %+v", err)
	}
	if daemonset.Status.NumberMisscheduled != 0 {
		t.Fatalf("The data manager daemonset is not up and running as expected")
	}

	veleroClientset, err := versioned.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to get velero clientset from k8s config: %+v", err)
	}

	backupName := AppNs + "-backup-" + fmt.Sprintf("%v", time.Now().Unix())
	t.Logf("backupName: %v", backupName)
	backupReq := builder.ForBackup(veleroNs, backupName).IncludedNamespaces(AppNs).SnapshotVolumes(true).Result()
	_, err = veleroClientset.VeleroV1().Backups(veleroNs).Create(backupReq)
	if err != nil {
		t.Fatalf("Failed to create velero backup: %+v", err)
	}

	t.Logf("Wait until the backup is completed")
	err = wait.PollImmediate(time.Second, 10*time.Minute, func() (bool, error) {
		backup, err := veleroClientset.VeleroV1().Backups(veleroNs).Get(backupName, metav1.GetOptions{})
		if err != nil {
			return false, errors.New(fmt.Sprintf("Failed to get velero backup, %v", backupName))
		}
		if backup.Status.Phase == v1.BackupPhaseCompleted {
			return true, nil
		} else if backup.Status.Phase == v1.BackupPhaseFailed || backup.Status.Phase == v1.BackupPhaseFailedValidation {
			return false, errors.New(fmt.Sprintf("Failed to backup, %v", backupName))
		} else {
			return false, nil
		}
	})

	if err != nil {
		t.Fatalf("Failed to complete the velero backup, %v: %+v", backupName, err)
	}

	err = WatchUploadRequests(config, veleroNs, nPVCs)
	if err != nil {
		t.Fatalf("Failed to upload all snapshots(%v): %v", nPVCs, err)
	}

	// Step 3: Delete the app namespace
	t.Logf("Step 3: Delete the app namespace")
	err = clientset.CoreV1().Namespaces().Delete(AppNs, &metav1.DeleteOptions{})
	if err != nil {
		t.Fatalf("Failed to delete the app namespace, %v: %v", AppNs, err)
	}
	t.Logf("Wait until the deletion is completed")
	err = wait.PollImmediate(time.Second, 10*time.Minute, func() (bool, error) {
		_, err = clientset.CoreV1().Namespaces().Get(AppNs, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			err = nil
			return true, nil
		} else {
			return false, err
		}
	})
	if err != nil {
		t.Fatalf("Failed to delete the app namespace, %v: %+v", AppNs, err)
	}

	// Step 4: Restore the app namespace from the previous backup
	t.Logf("Step 4: Restore the app namespace from the previous backup")
	restoreName := backupName + fmt.Sprintf("%v", time.Now().Unix())
	restoreReq := builder.ForRestore(veleroNs, restoreName).Backup(backupName).Result()
	_, err = veleroClientset.VeleroV1().Restores(veleroNs).Create(restoreReq)
	if err != nil {
		t.Fatalf("Failed to create the restore request, %v: %+v", restoreName, err)
	}

	t.Logf("Wait until the restore is completed")
	err = wait.PollImmediate(time.Second, 10*time.Minute, func() (bool, error) {
		restore, err := veleroClientset.VeleroV1().Restores(veleroNs).Get(restoreName, metav1.GetOptions{})
		if err != nil {
			return false, errors.New(fmt.Sprintf("Failed to get velero restore, %v", restoreName))
		}
		if restore.Status.Phase == v1.RestorePhaseCompleted {
			return true, nil
		} else if restore.Status.Phase == v1.RestorePhaseFailed || restore.Status.Phase == v1.RestorePhaseFailedValidation {
			return false, errors.New(fmt.Sprintf("Failed to restore, %v", restoreName))
		} else {
			return false, nil
		}
	})
	if err != nil {
		t.Fatalf("Failed to wait for the completion of restore, %v: %+v", restoreName, err)
	}

	// Step 5: Check the volume name using the govmomi vslm API and make sure the vso name is not hard coded.
	_, err = CheckVolumeNamesForPVCs(config, AppNs)
	if err != nil {
		t.Fatalf("Unexpected volume names were identified: %+v", err)
	}
}
