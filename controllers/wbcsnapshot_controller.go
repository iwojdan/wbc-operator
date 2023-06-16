/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	snapshotv1alpha1 "github.com/iwojdan/wbc-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// WbcSnapshotReconciler reconciles a WbcSnapshot object
type WbcSnapshotReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=snapshot.wbc.com,resources=wbcsnapshots,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=snapshot.wbc.com,resources=wbcsnapshots/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=snapshot.wbc.com,resources=wbcsnapshots/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WbcSnapshot object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *WbcSnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	fmt.Println("jestem_w_reconciler_function")
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	fmt.Println("jestem tuż przed snapshot z githuba pustego")
	snapshot := &snapshotv1alpha1.WbcSnapshot{}
	res2B, _ := json.Marshal(snapshot) //marshal json into bytes[]
	fmt.Println(string(res2B))
	fmt.Println("tutaj koniec WbcSnapshot")

	snapshotReadErr := r.Client.Get(ctx, req.NamespacedName, snapshot)
	res3B, _ := json.Marshal(snapshot)
	fmt.Println("Jestem linijka tuż po r.Client.Get i teraz snapshot się zmineił na", string(res3B))
	//if snapshotReadErr != nil {
	//		panic(fmt.Sprint("error 1:", snapshotReadErr))
	//	}
	//Exit early if something went wrong parsing the wbcsnapshot object
	//if snapshotReadErr != nil || len(snapshot.Name) == 0 {
	//	r.Log.Error(snapshotReadErr, "Error encountered reading WbcSnapshot")
	//	return ctrl.Result{}, snapshotReadErr
	//}
	fmt.Println("snapshotReadErr", snapshotReadErr)
	if snapshotReadErr != nil {
		fmt.Println("snapshotReadErr jest rózny od zera")
	}

	snapshotCreator := &corev1.Pod{}
	res4B, _ := json.Marshal(snapshotCreator)
	fmt.Println("snapshotCreator", string(res4B))
	snapshotCreatorReadErr := r.Get(
		ctx, types.NamespacedName{Name: "snapshot-creator", Namespace: req.Namespace}, snapshotCreator,
	)
	fmt.Println("snapshotCreatorReadErr", snapshotCreatorReadErr)
	// Exit early if the WBCSnapshot creator is already running
	if snapshotCreatorReadErr == nil {
		r.Log.Error(snapshotCreatorReadErr, "Snapshot Creator already running!")
		return ctrl.Result{}, snapshotCreatorReadErr
	}
	fmt.Println("Jestem za snapshotCreatorReadErr tworzymy zaraz newPVname")
	newPvName := "wbc-snapshot-pv-" + strconv.FormatInt(time.Now().Unix(), 10) + "-" + snapshot.Spec.SourceVolumeName
	newPvcName := "wbc-snapshot-pvc-" + strconv.FormatInt(time.Now().Unix(), 10) + "-" + snapshot.Spec.SourceClaimName

	// Create a new Persistent Volume
	newPersistentVol := corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newPvName,
			Namespace: req.Namespace,
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName: "manual",
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Capacity: corev1.ResourceList{
				corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("10Gi"),
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: snapshot.Spec.HostPath,
				},
			},
		},
	}
	pvCreateErr := r.Create(ctx, &newPersistentVol)
	fmt.Println("pvCreateErr", pvCreateErr)
	if pvCreateErr != nil {
		r.Log.Error(pvCreateErr, "Error encountered creating new pv")
		return ctrl.Result{}, pvCreateErr
	}
	fmt.Println("jestem przed log created new snapshot persistent Volume")
	fmt.Println("req.NamespacedName", req.NamespacedName)
	//_ = r.Log.WithValues("Created New Snapshot Persistent Volume", req.NamespacedName) to jest zle zadeklarowane i powoduje panic error w operatorze
	fmt.Println("New Snapshot Persistent volume", newPersistentVol)
	manualStorageClass := "manual"
	fmt.Println("manualStorageClass", manualStorageClass)
	fmt.Println("Jestem przed tworzeniem PersistentVolClaim")
	//Create a new Persistent Volume Claim connected to the new Persistent Volume
	newPersistentVolClaim := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newPvcName,
			Namespace: req.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &manualStorageClass,
			VolumeName:       newPvName,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("10Gi"),
				},
			},
		},
	}
	fmt.Println("Jestem za newPersistentVolClaim a przed pvcCreateErr")
	// client created PVC in the cluster
	pvcCreateErr := r.Create(ctx, &newPersistentVolClaim)
	fmt.Println("pvcCreateErr", pvcCreateErr)

	if pvcCreateErr != nil {
		r.Log.Error(pvcCreateErr, "Error encountered creating new pvc")
		return ctrl.Result{}, pvcCreateErr
	}
	//_ = r.Log.WithValues("Created New Snapshot Persistent Volume Claim", req.NamespacedName)
	fmt.Println("New Snapshot Persistent volume Claim", newPersistentVolClaim)
	//start a Pod that is hooked up to the snapshot PVC and the original PVC
	// The Pod simply copies a file from the old PVC to the new PVC - creating a primitive snapshot
	// corev1 api is used to create pod dynamicly
	snapshotCreatorPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snapshot-creator",
			Namespace: req.Namespace,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			Volumes: []corev1.Volume{
				{
					Name: "wbc-snapshot-" + snapshot.Spec.SourceClaimName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: snapshot.Spec.SourceClaimName,
						},
					},
				},
				{
					Name: "wbc-snapshot-" + newPvcName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: newPvcName,
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:  "busybox",
					Image: "k8s.gcr.io/busybox",
					Command: []string{
						"/bin/sh",
						"-c",
						//	"ls /tmp/dest"},
						//"mkdir /tmp/source",
						"echo -e \"abcd\" > /tmp/source/test",
						"cp /tmp/source/test /tmp/dest/"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "wbc-snapshot-" + snapshot.Spec.SourceClaimName,
							MountPath: "/tmp/source",
						},
						{
							Name:      "wbc-snapshot-" + newPvcName,
							MountPath: "/tmp/dest",
						},
					},
				},
			},
		},
	}

	podCreateErr := r.Create(ctx, &snapshotCreatorPod)
	if podCreateErr != nil {
		r.Log.Error(podCreateErr, "Error encountered creating snapshotting pod")
		return ctrl.Result{}, podCreateErr
	}
	fmt.Println("snapshot-creator pod", snapshotCreatorPod)

	//_ = r.Log.WithValues("Instantiating snapshot-creator pod", req.NamespacedName)

	/**
	Challenge:
	1. Can you use the "Owns" method on the controller to watch the Pod resources?
	2. Can you ensure that a snapshot (and the relative PVC and PV are removed appropiately?
	3. Can you update the status fields of the wbcSnapshot objects appropriately?)
	**/

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WbcSnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1alpha1.WbcSnapshot{}).
		Complete(r)
}
