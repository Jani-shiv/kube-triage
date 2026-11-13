package collector

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type WorkloadCollector struct {
	clientset *kubernetes.Clientset
	namespace string
}

func NewWorkloadCollector(clientset *kubernetes.Clientset, namespace string) *WorkloadCollector {
	return &WorkloadCollector{
		clientset: clientset,
		namespace: namespace,
	}
}

type WorkloadInfo struct {
	Kind       string
	Name       string
	Namespace  string
	OwnerRefs  []metav1.OwnerReference
	Selector   map[string]string
	Replicas   *int32
	ReadyReplicas int32
	Conditions []metav1.Condition
	PodTemplate *corev1.PodTemplateSpec
}

func (c *WorkloadCollector) Collect(ctx context.Context, resourceType, resourceName string) (*WorkloadInfo, error) {
	switch resourceType {
	case "deployment", "Deployment":
		return c.getDeployment(ctx, resourceName)
	case "replicaset", "ReplicaSet":
		return c.getReplicaSet(ctx, resourceName)
	case "statefulset", "StatefulSet":
		return c.getStatefulSet(ctx, resourceName)
	case "daemonset", "DaemonSet":
		return c.getDaemonSet(ctx, resourceName)
	case "job", "Job":
		return c.getJob(ctx, resourceName)
	case "cronjob", "CronJob":
		return c.getCronJob(ctx, resourceName)
	case "pod", "Pod":
		return c.getPodWorkloadInfo(ctx, resourceName)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func (c *WorkloadCollector) getDeployment(ctx context.Context, name string) (*WorkloadInfo, error) {
	deploy, err := c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	replicas := deploy.Spec.Replicas
	readyReplicas := deploy.Status.ReadyReplicas

	var conditions []metav1.Condition
	for _, cond := range deploy.Status.Conditions {
		conditions = append(conditions, metav1.Condition{
			Type:    string(cond.Type),
			Status:  metav1.ConditionStatus(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	return &WorkloadInfo{
		Kind:          "Deployment",
		Name:          deploy.Name,
		Namespace:     deploy.Namespace,
		OwnerRefs:     deploy.OwnerReferences,
		Selector:      deploy.Spec.Selector.MatchLabels,
		Replicas:      replicas,
		ReadyReplicas: readyReplicas,
		Conditions:    conditions,
		PodTemplate:   &deploy.Spec.Template,
	}, nil
}

func (c *WorkloadCollector) getReplicaSet(ctx context.Context, name string) (*WorkloadInfo, error) {
	rs, err := c.clientset.AppsV1().ReplicaSets(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	replicas := rs.Spec.Replicas
	readyReplicas := rs.Status.ReadyReplicas

	return &WorkloadInfo{
		Kind:          "ReplicaSet",
		Name:          rs.Name,
		Namespace:     rs.Namespace,
		OwnerRefs:     rs.OwnerReferences,
		Selector:      rs.Spec.Selector.MatchLabels,
		Replicas:      replicas,
		ReadyReplicas: readyReplicas,
		PodTemplate:   &rs.Spec.Template,
	}, nil
}

func (c *WorkloadCollector) getStatefulSet(ctx context.Context, name string) (*WorkloadInfo, error) {
	sts, err := c.clientset.AppsV1().StatefulSets(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	replicas := sts.Spec.Replicas
	readyReplicas := sts.Status.ReadyReplicas

	return &WorkloadInfo{
		Kind:          "StatefulSet",
		Name:          sts.Name,
		Namespace:     sts.Namespace,
		OwnerRefs:     sts.OwnerReferences,
		Selector:      sts.Spec.Selector.MatchLabels,
		Replicas:      replicas,
		ReadyReplicas: readyReplicas,
		PodTemplate:   &sts.Spec.Template,
	}, nil
}

func (c *WorkloadCollector) getDaemonSet(ctx context.Context, name string) (*WorkloadInfo, error) {
	ds, err := c.clientset.AppsV1().DaemonSets(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	desired := ds.Status.DesiredNumberScheduled
	ready := ds.Status.NumberReady

	return &WorkloadInfo{
		Kind:          "DaemonSet",
		Name:          ds.Name,
		Namespace:     ds.Namespace,
		OwnerRefs:     ds.OwnerReferences,
		Selector:      ds.Spec.Selector.MatchLabels,
		Replicas:      &desired,
		ReadyReplicas: ready,
		PodTemplate:   &ds.Spec.Template,
	}, nil
}

func (c *WorkloadCollector) getJob(ctx context.Context, name string) (*WorkloadInfo, error) {
	job, err := c.clientset.BatchV1().Jobs(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var conditions []metav1.Condition
	for _, cond := range job.Status.Conditions {
		conditions = append(conditions, metav1.Condition{
			Type:    string(cond.Type),
			Status:  metav1.ConditionStatus(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	return &WorkloadInfo{
		Kind:       "Job",
		Name:       job.Name,
		Namespace:  job.Namespace,
		OwnerRefs:  job.OwnerReferences,
		Selector:   job.Spec.Selector.MatchLabels,
		Conditions: conditions,
		PodTemplate: &job.Spec.Template,
	}, nil
}

func (c *WorkloadCollector) getCronJob(ctx context.Context, name string) (*WorkloadInfo, error) {
	cj, err := c.clientset.BatchV1().CronJobs(c.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &WorkloadInfo{
		Kind:       "CronJob",
		Name:       cj.Name,
		Namespace:  cj.Namespace,
		OwnerRefs:  cj.OwnerReferences,
		Selector:   cj.Spec.JobTemplate.Spec.Selector.MatchLabels,
		PodTemplate: &cj.Spec.JobTemplate.Spec.Template,
	}, nil
}

func (c *WorkloadCollector) getPodWorkloadInfo(ctx context.Context, podName string) (*WorkloadInfo, error) {
	pod, err := c.clientset.CoreV1().Pods(c.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	for _, ownerRef := range pod.OwnerReferences {
		switch ownerRef.Kind {
		case "ReplicaSet":
			return c.getReplicaSet(ctx, ownerRef.Name)
		case "StatefulSet":
			return c.getStatefulSet(ctx, ownerRef.Name)
		case "DaemonSet":
			return c.getDaemonSet(ctx, ownerRef.Name)
		case "Job":
			return c.getJob(ctx, ownerRef.Name)
		}
	}

	return &WorkloadInfo{
		Kind:      "Pod",
		Name:      pod.Name,
		Namespace: pod.Namespace,
		OwnerRefs: pod.OwnerReferences,
	}, nil
}

func (c *WorkloadCollector) GetPodsForWorkload(ctx context.Context, workloadInfo *WorkloadInfo) ([]corev1.Pod, error) {
	if workloadInfo.Selector == nil || len(workloadInfo.Selector) == 0 {
		return nil, fmt.Errorf("no selector found for workload")
	}

	labelSelector := labels.Set(workloadInfo.Selector).AsSelector().String()
	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func (c *WorkloadCollector) GetDeploymentForPod(ctx context.Context, podName string) (*appsv1.Deployment, error) {
	pod, err := c.clientset.CoreV1().Pods(c.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	for _, ownerRef := range pod.OwnerReferences {
		if ownerRef.Kind == "ReplicaSet" {
			rs, err := c.clientset.AppsV1().ReplicaSets(c.namespace).Get(ctx, ownerRef.Name, metav1.GetOptions{})
			if err != nil {
				continue
			}

			for _, rsOwnerRef := range rs.OwnerReferences {
				if rsOwnerRef.Kind == "Deployment" {
					return c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, rsOwnerRef.Name, metav1.GetOptions{})
				}
			}
		}
	}

	return nil, fmt.Errorf("no deployment found for pod %s", podName)
}

func (c *WorkloadCollector) GetReplicaSetsForDeployment(ctx context.Context, deploymentName string) ([]appsv1.ReplicaSet, error) {
	deploy, err := c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	labelSelector := labels.Set(deploy.Spec.Selector.MatchLabels).AsSelector().String()
	rss, err := c.clientset.AppsV1().ReplicaSets(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	return rss.Items, nil
}

func (c *WorkloadCollector) GetPodsForDeployment(ctx context.Context, deploymentName string) ([]corev1.Pod, error) {
	rss, err := c.GetReplicaSetsForDeployment(ctx, deploymentName)
	if err != nil {
		return nil, err
	}

	var allPods []corev1.Pod
	for _, rs := range rss {
		pods, err := c.GetPodsForReplicaSet(ctx, rs.Name)
		if err != nil {
			continue
		}
		allPods = append(allPods, pods...)
	}

	return allPods, nil
}

func (c *WorkloadCollector) GetPodsForReplicaSet(ctx context.Context, rsName string) ([]corev1.Pod, error) {
	rs, err := c.clientset.AppsV1().ReplicaSets(c.namespace).Get(ctx, rsName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return c.GetPodsForWorkload(ctx, &WorkloadInfo{
		Selector: rs.Spec.Selector.MatchLabels,
	})
}