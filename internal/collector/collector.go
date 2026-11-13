package collector

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Collector struct {
	clientset *kubernetes.Clientset
	namespace string
	previous  bool
	tailLines int64

	podCollector      *PodCollector
	eventsCollector   *EventsCollector
	workloadCollector *WorkloadCollector
	nodeCollector     *NodeCollector
	storageCollector  *StorageCollector
	networkCollector  *NetworkCollector
}

func NewCollector(clientset *kubernetes.Clientset, namespace string, previous bool, tailLines int64) *Collector {
	return &Collector{
		clientset:         clientset,
		namespace:         namespace,
		previous:          previous,
		tailLines:         tailLines,
		podCollector:      NewPodCollector(clientset, namespace, previous, tailLines),
		eventsCollector:   NewEventsCollector(clientset, namespace),
		workloadCollector: NewWorkloadCollector(clientset, namespace),
		nodeCollector:     NewNodeCollector(clientset),
		storageCollector:  NewStorageCollector(clientset, namespace),
		networkCollector:  NewNetworkCollector(clientset, namespace),
	}
}

type DiagnosticData struct {
	ResourceType     string
	ResourceName     string
	Namespace        string
	PodData          *PodDiagnosticData
	EventsData       *EventsData
	WorkloadData     *WorkloadInfo
	NodeData         *NodeInfo
	StorageData      *PVCInfo
	NetworkData      *ServiceInfo
	CollectedAt      string
}

func (c *Collector) Collect(ctx context.Context, resourceType, resourceName string) (*DiagnosticData, error) {
	data := &DiagnosticData{
		ResourceType: resourceType,
		ResourceName: resourceName,
		Namespace:    c.namespace,
		CollectedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	switch resourceType {
	case "pod", "Pod":
		podData, err := c.podCollector.Collect(ctx, resourceName)
		if err != nil {
			return nil, fmt.Errorf("pod collection failed: %w", err)
		}
		data.PodData = podData

		eventsData, err := c.eventsCollector.Collect(ctx, resourceType, resourceName)
		if err != nil {
			return nil, fmt.Errorf("events collection failed: %w", err)
		}
		data.EventsData = eventsData

		workloadData, err := c.workloadCollector.Collect(ctx, resourceType, resourceName)
		if err != nil {
			return nil, fmt.Errorf("workload collection failed: %w", err)
		}
		data.WorkloadData = workloadData

		if podData != nil && podData.Pod != nil && podData.Pod.Spec.NodeName != "" {
			nodeData, err := c.nodeCollector.Collect(ctx, podData.Pod.Spec.NodeName)
			if err == nil {
				data.NodeData = nodeData
			}
		}

	case "deployment", "Deployment",
		"replicaset", "ReplicaSet",
		"statefulset", "StatefulSet",
		"daemonset", "DaemonSet",
		"job", "Job",
		"cronjob", "CronJob":

		workloadData, err := c.workloadCollector.Collect(ctx, resourceType, resourceName)
		if err != nil {
			return nil, fmt.Errorf("workload collection failed: %w", err)
		}
		data.WorkloadData = workloadData

		eventsData, err := c.eventsCollector.Collect(ctx, resourceType, resourceName)
		if err != nil {
			return nil, fmt.Errorf("events collection failed: %w", err)
		}
		data.EventsData = eventsData

		if workloadData.PodTemplate != nil {
			pods, err := c.workloadCollector.GetPodsForWorkload(ctx, workloadData)
			if err == nil && len(pods) > 0 {
				podData, err := c.podCollector.Collect(ctx, pods[0].Name)
				if err == nil {
					// Find the deployment name for remediation commands
					deploy, err := c.workloadCollector.GetDeploymentForPod(ctx, pods[0].Name)
					if err == nil && deploy != nil {
						podData.Pod.OwnerReferences = append(podData.Pod.OwnerReferences, metav1.OwnerReference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Name:       deploy.Name,
							UID:        deploy.UID,
						})
					}
					data.PodData = podData
				}
			}
		}

	case "service", "Service":
		networkData, err := c.networkCollector.GetService(ctx, resourceName)
		if err != nil {
			return nil, fmt.Errorf("service collection failed: %w", err)
		}
		data.NetworkData = networkData

		eventsData, err := c.eventsCollector.Collect(ctx, resourceType, resourceName)
		if err != nil {
			return nil, fmt.Errorf("events collection failed: %w", err)
		}
		data.EventsData = eventsData

	case "persistentvolumeclaim", "pvc", "PersistentVolumeClaim":
		storageData, err := c.storageCollector.GetPVC(ctx, resourceName)
		if err != nil {
			return nil, fmt.Errorf("PVC collection failed: %w", err)
		}
		data.StorageData = storageData

		eventsData, err := c.eventsCollector.GetEventsForPVC(ctx, resourceName)
		if err != nil {
			return nil, fmt.Errorf("events collection failed: %w", err)
		}
		data.EventsData = &EventsData{PodEvents: eventsData}

	case "node", "Node":
		nodeData, err := c.nodeCollector.Collect(ctx, resourceName)
		if err != nil {
			return nil, fmt.Errorf("node collection failed: %w", err)
		}
		data.NodeData = nodeData

		eventsData, err := c.eventsCollector.GetEventsForNode(ctx, resourceName)
		if err != nil {
			return nil, fmt.Errorf("events collection failed: %w", err)
		}
		data.EventsData = &EventsData{PodEvents: eventsData}

	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	return data, nil
}

func (c *Collector) GetPodCollector() *PodCollector {
	return c.podCollector
}

func (c *Collector) GetEventsCollector() *EventsCollector {
	return c.eventsCollector
}

func (c *Collector) GetWorkloadCollector() *WorkloadCollector {
	return c.workloadCollector
}

func (c *Collector) GetNodeCollector() *NodeCollector {
	return c.nodeCollector
}

func (c *Collector) GetStorageCollector() *StorageCollector {
	return c.storageCollector
}

func (c *Collector) GetNetworkCollector() *NetworkCollector {
	return c.networkCollector
}