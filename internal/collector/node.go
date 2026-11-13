package collector

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NodeCollector struct {
	clientset *kubernetes.Clientset
}

func NewNodeCollector(clientset *kubernetes.Clientset) *NodeCollector {
	return &NodeCollector{clientset: clientset}
}

type NodeInfo struct {
	Name           string
	Conditions     []corev1.NodeCondition
	Addresses      []corev1.NodeAddress
	Capacity       corev1.ResourceList
	Allocatable    corev1.ResourceList
	Taints         []corev1.Taint
	Images         []corev1.ContainerImage
	KubeletVersion string
	ContainerRuntime string
}

func (c *NodeCollector) Collect(ctx context.Context, nodeName string) (*NodeInfo, error) {
	node, err := c.clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}

	info := &NodeInfo{
		Name:        node.Name,
		Conditions:  node.Status.Conditions,
		Addresses:   node.Status.Addresses,
		Capacity:    node.Status.Capacity,
		Allocatable: node.Status.Allocatable,
		Taints:      node.Spec.Taints,
		Images:      node.Status.Images,
	}

	if node.Status.NodeInfo.KubeletVersion != "" {
		info.KubeletVersion = node.Status.NodeInfo.KubeletVersion
	}
	if node.Status.NodeInfo.ContainerRuntimeVersion != "" {
		info.ContainerRuntime = node.Status.NodeInfo.ContainerRuntimeVersion
	}

	return info, nil
}

func (c *NodeCollector) List(ctx context.Context) ([]NodeInfo, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var nodeInfos []NodeInfo
	for _, node := range nodes.Items {
		info := &NodeInfo{
			Name:        node.Name,
			Conditions:  node.Status.Conditions,
			Addresses:   node.Status.Addresses,
			Capacity:    node.Status.Capacity,
			Allocatable: node.Status.Allocatable,
			Taints:      node.Spec.Taints,
			Images:      node.Status.Images,
		}
		if node.Status.NodeInfo.KubeletVersion != "" {
			info.KubeletVersion = node.Status.NodeInfo.KubeletVersion
		}
		if node.Status.NodeInfo.ContainerRuntimeVersion != "" {
			info.ContainerRuntime = node.Status.NodeInfo.ContainerRuntimeVersion
		}
		nodeInfos = append(nodeInfos, *info)
	}

	return nodeInfos, nil
}

func (c *NodeCollector) GetNodePressureConditions(node *NodeInfo) map[string]bool {
	pressures := make(map[string]bool)
	for _, cond := range node.Conditions {
		switch cond.Type {
		case corev1.NodeMemoryPressure:
			pressures["MemoryPressure"] = cond.Status == corev1.ConditionTrue
		case corev1.NodeDiskPressure:
			pressures["DiskPressure"] = cond.Status == corev1.ConditionTrue
		case corev1.NodePIDPressure:
			pressures["PIDPressure"] = cond.Status == corev1.ConditionTrue
		case corev1.NodeNetworkUnavailable:
			pressures["NetworkUnavailable"] = cond.Status == corev1.ConditionTrue
		}
	}
	return pressures
}

func (c *NodeCollector) HasTaint(node *NodeInfo, key, value, effect string) bool {
	for _, taint := range node.Taints {
		if taint.Key == key && (value == "" || taint.Value == value) && (effect == "" || string(taint.Effect) == effect) {
			return true
		}
	}
	return false
}