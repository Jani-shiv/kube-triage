package analyzer

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"kube-triage/internal/collector"
)

type Engine struct {
	provider string
	model    string
	endpoint string
}

func NewEngine(provider, model, endpoint string) *Engine {
	return &Engine{
		provider: provider,
		model:    model,
		endpoint: endpoint,
	}
}

type AnalysisResult struct {
	ResourceType    string
	ResourceName    string
	Namespace       string
	Severity        string
	Summary         string
	RootCause       string
	Remediation     []RemediationStep
	Prevention      []string
	CollectedAt     string
}

type RemediationStep struct {
	Description string
	Command     string
	Type        string
}

func (e *Engine) Analyze(ctx context.Context, data *collector.DiagnosticData) (*AnalysisResult, error) {
	result := &AnalysisResult{
		ResourceType: data.ResourceType,
		ResourceName: data.ResourceName,
		Namespace:    data.Namespace,
		CollectedAt:  data.CollectedAt,
	}

	if data.PodData != nil && data.PodData.Pod != nil {
		pod := data.PodData.Pod

		if len(pod.Status.ContainerStatuses) > 0 {
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Terminated != nil && cs.State.Terminated.ExitCode == 137 && cs.State.Terminated.Reason == "OOMKilled" {
					result.Severity = "CRITICAL"
					result.Summary = "Container OOMKilled - exceeded memory limit"
					result.RootCause = "The container exceeded its hard memory limit. Kubernetes sent SIGKILL (exit code 137) to terminate the process."
					result.Remediation = []RemediationStep{
						{
							Description: "Increase memory limit in deployment spec",
							Command:     "kubectl patch deployment " + data.ResourceName + " --type='json' -p='[{\"op\": \"replace\", \"path\": \"/spec/template/spec/containers/0/resources/limits/memory\", \"value\": \"1Gi\"}]'",
							Type:        "command",
						},
					}
					result.Prevention = []string{
						"Set appropriate memory requests and limits based on profiling",
						"Configure Prometheus alerts for memory usage approaching limits",
						"Implement application-level memory monitoring",
					}
					return result, nil
				}

				if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
					result.Severity = "HIGH"
					result.Summary = "Container in CrashLoopBackOff - restarting repeatedly"
					result.RootCause = "The container is failing and restarting repeatedly. Check container logs for the actual error."
					result.Remediation = []RemediationStep{
						{
							Description: "Check container logs for error details",
							Command:     "kubectl logs " + data.ResourceName + " -n " + data.Namespace + " --previous",
							Type:        "command",
						},
					}
					result.Prevention = []string{
						"Configure proper liveness and readiness probes",
						"Set appropriate startup probes for slow-starting applications",
					}
					return result, nil
				}

				if cs.State.Waiting != nil && cs.State.Waiting.Reason == "ImagePullBackOff" {
					result.Severity = "HIGH"
					result.Summary = "Image pull failed - cannot retrieve container image"
					result.RootCause = "Kubernetes cannot pull the container image. Check image name, registry credentials, and network connectivity."
					result.Remediation = []RemediationStep{
						{
							Description: "Verify image name and tag",
							Command:     "kubectl describe pod " + data.ResourceName + " -n " + data.Namespace,
							Type:        "command",
						},
					}
					result.Prevention = []string{
						"Use imagePullSecrets for private registries",
						"Verify image exists and is accessible",
					}
					return result, nil
				}
			}
		}

		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionFalse {
				result.Severity = "MEDIUM"
				result.Summary = "Pod not ready"
				result.RootCause = "Pod readiness probe is failing or not configured."
				result.Remediation = []RemediationStep{
					{
						Description: "Check pod events and readiness probe configuration",
						Command:     "kubectl describe pod " + data.ResourceName + " -n " + data.Namespace,
						Type:        "command",
					},
				}
				result.Prevention = []string{
					"Configure proper readiness probes",
					"Ensure application handles startup gracefully",
				}
				return result, nil
			}
		}
	}

	if data.EventsData != nil && len(data.EventsData.PodEvents) > 0 {
		for _, event := range data.EventsData.PodEvents {
			if event.Reason == "FailedMount" {
				result.Severity = "HIGH"
				result.Summary = "Volume mount failed"
				result.RootCause = "Failed to mount volume: " + event.Message
				result.Remediation = []RemediationStep{
					{
						Description: "Check PVC status and storage provisioner",
						Command:     "kubectl get pvc -n " + data.Namespace,
						Type:        "command",
					},
				}
				return result, nil
			}
			if event.Reason == "FailedScheduling" {
				result.Severity = "HIGH"
				result.Summary = "Pod scheduling failed"
				result.RootCause = "Scheduler cannot place pod: " + event.Message
				result.Remediation = []RemediationStep{
					{
						Description: "Check node resources and taints",
						Command:     "kubectl describe nodes",
						Type:        "command",
					},
				}
				return result, nil
			}
		}
	}

	result.Severity = "LOW"
	result.Summary = "No obvious failure patterns detected"
	result.RootCause = "The resource appears to be in an unknown state. Manual investigation recommended."
	result.Remediation = []RemediationStep{
		{
			Description: "Get detailed resource description",
			Command:     "kubectl describe " + data.ResourceType + " " + data.ResourceName + " -n " + data.Namespace,
			Type:        "command",
		},
	}
	result.Prevention = []string{
		"Implement comprehensive monitoring and alerting",
		"Regular cluster health checks",
	}

	return result, nil
}