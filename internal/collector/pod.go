package collector

import (
	"context"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodCollector struct {
	clientset *kubernetes.Clientset
	namespace string
	previous  bool
	tailLines int64
}

func NewPodCollector(clientset *kubernetes.Clientset, namespace string, previous bool, tailLines int64) *PodCollector {
	return &PodCollector{
		clientset: clientset,
		namespace: namespace,
		previous:  previous,
		tailLines: tailLines,
	}
}

type PodDiagnosticData struct {
	Pod           *corev1.Pod
	ContainerLogs map[string]string
	Events        []corev1.Event
	OwnerRefs     []metav1.OwnerReference
}

func (c *PodCollector) Collect(ctx context.Context, podName string) (*PodDiagnosticData, error) {
	pod, err := c.clientset.CoreV1().Pods(c.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s: %w", podName, err)
	}

	containerLogs := make(map[string]string)
	for _, container := range pod.Spec.Containers {
		logs, err := c.getContainerLogs(ctx, pod, container.Name)
		if err != nil {
			containerLogs[container.Name] = fmt.Sprintf("ERROR fetching logs: %v", err)
		} else {
			containerLogs[container.Name] = logs
		}
	}

	if c.previous {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.RestartCount > 0 {
				logs, err := c.getContainerLogs(ctx, pod, containerStatus.Name, true)
				if err != nil {
					containerLogs[containerStatus.Name+"-previous"] = fmt.Sprintf("ERROR fetching previous logs: %v", err)
				} else {
					containerLogs[containerStatus.Name+"-previous"] = logs
				}
			}
		}
	}

	events, err := c.getPodEvents(ctx, pod)
	if err != nil {
		events = []corev1.Event{}
	}

	return &PodDiagnosticData{
		Pod:           pod,
		ContainerLogs: containerLogs,
		Events:        events,
		OwnerRefs:     pod.OwnerReferences,
	}, nil
}

func (c *PodCollector) getContainerLogs(ctx context.Context, pod *corev1.Pod, containerName string, previous ...bool) (string, error) {
	prev := false
	if len(previous) > 0 {
		prev = previous[0]
	}

	req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container:  containerName,
		Previous:   prev,
		TailLines:  &c.tailLines,
		Timestamps: true,
	})

	logs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer logs.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, logs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (c *PodCollector) getPodEvents(ctx context.Context, pod *corev1.Pod) ([]corev1.Event, error) {
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s,involvedObject.kind=Pod", pod.Name, pod.Namespace)
	list, err := c.clientset.CoreV1().Events(c.namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	var warningEvents []corev1.Event
	for _, event := range list.Items {
		if event.Type == "Warning" {
			warningEvents = append(warningEvents, event)
		}
	}

	return warningEvents, nil
}

func (c *PodCollector) GetContainerStatuses(pod *corev1.Pod) []ContainerStatusInfo {
	var statuses []ContainerStatusInfo
	for _, cs := range pod.Status.ContainerStatuses {
		info := ContainerStatusInfo{
			Name:         cs.Name,
			Ready:        cs.Ready,
			RestartCount: cs.RestartCount,
			Image:        cs.Image,
			ImageID:      cs.ImageID,
		}

		if cs.State.Terminated != nil {
			info.State = "Terminated"
			info.ExitCode = cs.State.Terminated.ExitCode
			info.Reason = cs.State.Terminated.Reason
			info.Message = cs.State.Terminated.Message
			info.StartedAt = cs.State.Terminated.StartedAt
			info.FinishedAt = cs.State.Terminated.FinishedAt
		} else if cs.State.Waiting != nil {
			info.State = "Waiting"
			info.Reason = cs.State.Waiting.Reason
			info.Message = cs.State.Waiting.Message
		} else if cs.State.Running != nil {
			info.State = "Running"
			info.StartedAt = cs.State.Running.StartedAt
		}

		statuses = append(statuses, info)
	}

	for _, cs := range pod.Status.InitContainerStatuses {
		info := ContainerStatusInfo{
			Name:         cs.Name,
			Ready:        cs.Ready,
			RestartCount: cs.RestartCount,
			Image:        cs.Image,
			ImageID:      cs.ImageID,
			IsInit:       true,
		}

		if cs.State.Terminated != nil {
			info.State = "Terminated"
			info.ExitCode = cs.State.Terminated.ExitCode
			info.Reason = cs.State.Terminated.Reason
			info.Message = cs.State.Terminated.Message
			info.StartedAt = cs.State.Terminated.StartedAt
			info.FinishedAt = cs.State.Terminated.FinishedAt
		} else if cs.State.Waiting != nil {
			info.State = "Waiting"
			info.Reason = cs.State.Waiting.Reason
			info.Message = cs.State.Waiting.Message
		} else if cs.State.Running != nil {
			info.State = "Running"
			info.StartedAt = cs.State.Running.StartedAt
		}

		statuses = append(statuses, info)
	}

	return statuses
}

type ContainerStatusInfo struct {
	Name         string
	Ready        bool
	RestartCount int32
	Image        string
	ImageID      string
	State        string
	ExitCode     int32
	Reason       string
	Message      string
	StartedAt    metav1.Time
	FinishedAt   metav1.Time
	IsInit       bool
}