package collector

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type StorageCollector struct {
	clientset *kubernetes.Clientset
	namespace string
}

func NewStorageCollector(clientset *kubernetes.Clientset, namespace string) *StorageCollector {
	return &StorageCollector{
		clientset: clientset,
		namespace: namespace,
	}
}

type PVCInfo struct {
	Name              string
	Namespace         string
	Status            corev1.PersistentVolumeClaimStatus
	Spec              corev1.PersistentVolumeClaimSpec
	VolumeName        string
	Capacity          corev1.ResourceList
	AccessModes       []corev1.PersistentVolumeAccessMode
	StorageClassName  *string
	VolumeMode        *corev1.PersistentVolumeMode
	Conditions        []metav1.Condition
	BoundPV           *PVInfo
}

type PVInfo struct {
	Name              string
	Status            corev1.PersistentVolumeStatus
	Spec              corev1.PersistentVolumeSpec
	Capacity          corev1.ResourceList
	AccessModes       []corev1.PersistentVolumeAccessMode
	StorageClassName  string
	VolumeMode        *corev1.PersistentVolumeMode
	ClaimRef          *corev1.ObjectReference
	Reason            string
	Message           string
}

func (c *StorageCollector) GetPVC(ctx context.Context, pvcName string) (*PVCInfo, error) {
	pvc, err := c.clientset.CoreV1().PersistentVolumeClaims(c.namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get PVC %s: %w", pvcName, err)
	}

	info := &PVCInfo{
		Name:             pvc.Name,
		Namespace:        pvc.Namespace,
		Status:           pvc.Status,
		Spec:             pvc.Spec,
		VolumeName:       pvc.Spec.VolumeName,
		Capacity:         pvc.Status.Capacity,
		AccessModes:      pvc.Spec.AccessModes,
		StorageClassName: pvc.Spec.StorageClassName,
		VolumeMode:       pvc.Spec.VolumeMode,
	}

	if pvc.Spec.VolumeName != "" {
		pv, err := c.clientset.CoreV1().PersistentVolumes().Get(ctx, pvc.Spec.VolumeName, metav1.GetOptions{})
		if err == nil {
			info.BoundPV = &PVInfo{
				Name:             pv.Name,
				Status:           pv.Status,
				Spec:             pv.Spec,
				Capacity:         pv.Spec.Capacity,
				AccessModes:      pv.Spec.AccessModes,
				StorageClassName: pv.Spec.StorageClassName,
				VolumeMode:       pv.Spec.VolumeMode,
				ClaimRef:         pv.Spec.ClaimRef,
				Reason:           pv.Status.Reason,
				Message:          pv.Status.Message,
			}
		}
	}

	return info, nil
}

func (c *StorageCollector) ListPVCs(ctx context.Context) ([]PVCInfo, error) {
	pvcs, err := c.clientset.CoreV1().PersistentVolumeClaims(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var infos []PVCInfo
	for _, pvc := range pvcs.Items {
		info := PVCInfo{
			Name:             pvc.Name,
			Namespace:        pvc.Namespace,
			Status:           pvc.Status,
			Spec:             pvc.Spec,
			VolumeName:       pvc.Spec.VolumeName,
			Capacity:         pvc.Status.Capacity,
			AccessModes:      pvc.Spec.AccessModes,
			StorageClassName: pvc.Spec.StorageClassName,
			VolumeMode:       pvc.Spec.VolumeMode,
		}
		infos = append(infos, info)
	}

	return infos, nil
}

func (c *StorageCollector) GetPendingPVCs(ctx context.Context) ([]PVCInfo, error) {
	pvcs, err := c.ListPVCs(ctx)
	if err != nil {
		return nil, err
	}

	var pending []PVCInfo
	for _, pvc := range pvcs {
		if pvc.Status.Phase == corev1.ClaimPending {
			pending = append(pending, pvc)
		}
	}

	return pending, nil
}

func (c *StorageCollector) GetPV(ctx context.Context, pvName string) (*PVInfo, error) {
	pv, err := c.clientset.CoreV1().PersistentVolumes().Get(ctx, pvName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get PV %s: %w", pvName, err)
	}

	return &PVInfo{
		Name:             pv.Name,
		Status:           pv.Status,
		Spec:             pv.Spec,
		Capacity:         pv.Spec.Capacity,
		AccessModes:      pv.Spec.AccessModes,
		StorageClassName: pv.Spec.StorageClassName,
		VolumeMode:       pv.Spec.VolumeMode,
		ClaimRef:         pv.Spec.ClaimRef,
		Reason:           pv.Status.Reason,
		Message:          pv.Status.Message,
	}, nil
}

func (c *StorageCollector) ListPVs(ctx context.Context) ([]PVInfo, error) {
	pvs, err := c.clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var infos []PVInfo
	for _, pv := range pvs.Items {
		info := PVInfo{
			Name:             pv.Name,
			Status:           pv.Status,
			Spec:             pv.Spec,
			Capacity:         pv.Spec.Capacity,
			AccessModes:      pv.Spec.AccessModes,
			StorageClassName: pv.Spec.StorageClassName,
			VolumeMode:       pv.Spec.VolumeMode,
			ClaimRef:         pv.Spec.ClaimRef,
			Reason:           pv.Status.Reason,
			Message:          pv.Status.Message,
		}
		infos = append(infos, info)
	}

	return infos, nil
}

func (c *StorageCollector) GetStorageClasses(ctx context.Context) ([]string, error) {
	// Note: StorageClass is in storage.k8s.io/v1, not core
	// This would need storagev1 client
	return []string{}, nil
}

func (c *StorageCollector) IsPVCBound(pvc *PVCInfo) bool {
	return pvc.Status.Phase == corev1.ClaimBound
}

func (c *StorageCollector) GetPVCEvents(ctx context.Context, pvcName string) ([]corev1.Event, error) {
	list, err := c.clientset.CoreV1().Events(c.namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=PersistentVolumeClaim", pvcName, c.namespace),
	})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}