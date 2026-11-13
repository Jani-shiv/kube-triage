package collector

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type EventsCollector struct {
	clientset *kubernetes.Clientset
	namespace string
}

func NewEventsCollector(clientset *kubernetes.Clientset, namespace string) *EventsCollector {
	return &EventsCollector{
		clientset: clientset,
		namespace: namespace,
	}
}

type EventsData struct {
	PodEvents       []corev1.Event
	NamespaceEvents []corev1.Event
	RelatedEvents   []corev1.Event
}

func (c *EventsCollector) Collect(ctx context.Context, resourceType, resourceName string) (*EventsData, error) {
	data := &EventsData{}

	podEvents, err := c.getResourceEvents(ctx, "Pod", resourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod events: %w", err)
	}
	data.PodEvents = podEvents

	nsEvents, err := c.getNamespaceWarningEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace events: %w", err)
	}
	data.NamespaceEvents = nsEvents

	relatedEvents, err := c.getRelatedEvents(ctx, resourceType, resourceName)
	if err != nil {
		relatedEvents = []corev1.Event{}
	}
	data.RelatedEvents = relatedEvents

	return data, nil
}

func (c *EventsCollector) getResourceEvents(ctx context.Context, kind, name string) ([]corev1.Event, error) {
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=%s", name, c.namespace, kind)
	list, err := c.clientset.CoreV1().Events(c.namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	var events []corev1.Event
	for _, event := range list.Items {
		if event.Type == "Warning" {
			events = append(events, event)
		}
	}

	return events, nil
}

func (c *EventsCollector) getNamespaceWarningEvents(ctx context.Context) ([]corev1.Event, error) {
	list, err := c.clientset.CoreV1().Events(c.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var events []corev1.Event
	for _, event := range list.Items {
		if event.Type == "Warning" {
			events = append(events, event)
		}
	}

	return events, nil
}

func (c *EventsCollector) getRelatedEvents(ctx context.Context, resourceType, resourceName string) ([]corev1.Event, error) {
	var related []corev1.Event

	switch resourceType {
	case "pod", "Pod":
		podEvents, _ := c.getResourceEvents(ctx, "Pod", resourceName)
		related = append(related, podEvents...)

	case "deployment", "Deployment":
		replicasets, err := c.clientset.AppsV1().ReplicaSets(c.namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", resourceName),
		})
		if err == nil {
			for _, rs := range replicasets.Items {
				rsEvents, _ := c.getResourceEvents(ctx, "ReplicaSet", rs.Name)
				related = append(related, rsEvents...)

				pods, _ := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
					LabelSelector: fmt.Sprintf("controller-revision-hash=%s", rs.Labels["controller-revision-hash"]),
				})
				for _, pod := range pods.Items {
					podEvents, _ := c.getResourceEvents(ctx, "Pod", pod.Name)
					related = append(related, podEvents...)
				}
			}
		}

	case "statefulset", "StatefulSet":
		pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("statefulset.kubernetes.io/pod-name=%s", resourceName),
		})
		if err == nil {
			for _, pod := range pods.Items {
				podEvents, _ := c.getResourceEvents(ctx, "Pod", pod.Name)
				related = append(related, podEvents...)
			}
		}

	case "daemonset", "DaemonSet":
		pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("daemonset=%s", resourceName),
		})
		if err == nil {
			for _, pod := range pods.Items {
				podEvents, _ := c.getResourceEvents(ctx, "Pod", pod.Name)
				related = append(related, podEvents...)
			}
		}

	case "job", "Job":
		pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("job-name=%s", resourceName),
		})
		if err == nil {
			for _, pod := range pods.Items {
				podEvents, _ := c.getResourceEvents(ctx, "Pod", pod.Name)
				related = append(related, podEvents...)
			}
		}

	case "cronjob", "CronJob":
		jobs, err := c.clientset.BatchV1().Jobs(c.namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("cronjob-name=%s", resourceName),
		})
		if err == nil {
			for _, job := range jobs.Items {
				jobEvents, _ := c.getResourceEvents(ctx, "Job", job.Name)
				related = append(related, jobEvents...)

				pods, _ := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
					LabelSelector: fmt.Sprintf("job-name=%s", job.Name),
				})
				for _, pod := range pods.Items {
					podEvents, _ := c.getResourceEvents(ctx, "Pod", pod.Name)
					related = append(related, podEvents...)
				}
			}
		}
	}

	return related, nil
}

func (c *EventsCollector) GetEventsForPVC(ctx context.Context, pvcName string) ([]corev1.Event, error) {
	return c.getResourceEvents(ctx, "PersistentVolumeClaim", pvcName)
}

func (c *EventsCollector) GetEventsForNode(ctx context.Context, nodeName string) ([]corev1.Event, error) {
	return c.getResourceEvents(ctx, "Node", nodeName)
}