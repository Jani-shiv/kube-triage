package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	"kube-triage/internal/collector"
)

type PayloadBuilder struct {
	maxLogLines   int
	maxEvents     int
	maxContainers int
}

func NewPayloadBuilder(maxLogLines, maxEvents, maxContainers int) *PayloadBuilder {
	return &PayloadBuilder{
		maxLogLines:   maxLogLines,
		maxEvents:     maxEvents,
		maxContainers: maxContainers,
	}
}

type AIPayload struct {
	Timestamp    string                 `json:"timestamp"`
	Resource     ResourceInfo           `json:"resource"`
	Pod          *PodInfo               `json:"pod,omitempty"`
	Workload     *WorkloadInfo          `json:"workload,omitempty"`
	Events       []EventInfo            `json:"events,omitempty"`
	Node         *NodeInfo              `json:"node,omitempty"`
	Storage      *StorageInfo           `json:"storage,omitempty"`
	Network      *NetworkInfo           `json:"network,omitempty"`
	MatchedRules []string               `json:"matched_rules,omitempty"`
}

type ResourceInfo struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type PodInfo struct {
	Name              string             `json:"name"`
	Namespace         string             `json:"namespace"`
	Phase             string             `json:"phase"`
	NodeName          string             `json:"node_name"`
	QoSClass          string             `json:"qos_class"`
	RestartPolicy     string             `json:"restart_policy"`
	ServiceAccount    string             `json:"service_account"`
	Containers        []ContainerInfo    `json:"containers"`
	InitContainers    []ContainerInfo    `json:"init_containers"`
	Conditions        []ConditionInfo    `json:"conditions"`
	Volumes           []VolumeInfo       `json:"volumes"`
	OwnerReferences   []OwnerRefInfo     `json:"owner_references"`
}

type ContainerInfo struct {
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	ImageID      string            `json:"image_id"`
	Ready        bool              `json:"ready"`
	RestartCount int32             `json:"restart_count"`
	State        string            `json:"state"`
	ExitCode     int32             `json:"exit_code,omitempty"`
	Reason       string            `json:"reason,omitempty"`
	Message      string            `json:"message,omitempty"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	FinishedAt   *time.Time        `json:"finished_at,omitempty"`
	Resources    ResourceReqLimits `json:"resources"`
	Ports        []PortInfo        `json:"ports"`
	EnvVars      []EnvVarInfo      `json:"env_vars"`
	VolumeMounts []VolumeMountInfo `json:"volume_mounts"`
	LivenessProbe  *ProbeInfo      `json:"liveness_probe,omitempty"`
	ReadinessProbe *ProbeInfo      `json:"readiness_probe,omitempty"`
	StartupProbe   *ProbeInfo      `json:"startup_probe,omitempty"`
	Logs           string          `json:"logs,omitempty"`
	PreviousLogs   string          `json:"previous_logs,omitempty"`
}

type ResourceReqLimits struct {
	Requests map[string]string `json:"requests,omitempty"`
	Limits   map[string]string `json:"limits,omitempty"`
}

type PortInfo struct {
	Name          string `json:"name"`
	ContainerPort int32  `json:"container_port"`
	Protocol      string `json:"protocol"`
}

type EnvVarInfo struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
	ValueFrom string `json:"value_from,omitempty"`
}

type VolumeMountInfo struct {
	Name      string `json:"name"`
	MountPath string `json:"mount_path"`
	ReadOnly  bool   `json:"read_only"`
}

type ProbeInfo struct {
	Type                string `json:"type"`
	Path                string `json:"path,omitempty"`
	Port                int32  `json:"port,omitempty"`
	InitialDelaySeconds int32  `json:"initial_delay_seconds"`
	PeriodSeconds       int32  `json:"period_seconds"`
	TimeoutSeconds      int32  `json:"timeout_seconds"`
	FailureThreshold    int32  `json:"failure_threshold"`
	SuccessThreshold    int32  `json:"success_threshold"`
}

type ConditionInfo struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

type VolumeInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	ClaimName    string `json:"claim_name,omitempty"`
	ConfigMap    string `json:"config_map,omitempty"`
	Secret       string `json:"secret,omitempty"`
}

type OwnerRefInfo struct {
	APIVersion string `json:"api_version"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	UID        string `json:"uid"`
	Controller *bool  `json:"controller,omitempty"`
}

type WorkloadInfo struct {
	Kind             string            `json:"kind"`
	Name             string            `json:"name"`
	Namespace        string            `json:"namespace"`
	Replicas         *int32            `json:"replicas,omitempty"`
	ReadyReplicas    int32             `json:"ready_replicas"`
	Selector         map[string]string `json:"selector,omitempty"`
	Strategy         string            `json:"strategy,omitempty"`
	Template         *PodTemplateInfo  `json:"template,omitempty"`
	Conditions       []ConditionInfo   `json:"conditions"`
}

type PodTemplateInfo struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Spec        *PodSpecInfo      `json:"spec,omitempty"`
}

type PodSpecInfo struct {
	ServiceAccountName string            `json:"service_account_name,omitempty"`
	Containers         []ContainerInfo   `json:"containers"`
	InitContainers     []ContainerInfo   `json:"init_containers"`
	Volumes            []VolumeInfo      `json:"volumes"`
	NodeSelector       map[string]string `json:"node_selector,omitempty"`
	Affinity           string            `json:"affinity,omitempty"`
	Tolerations        []TolerationInfo  `json:"tolerations,omitempty"`
}

type TolerationInfo struct {
	Key      string `json:"key,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
	Effect   string `json:"effect,omitempty"`
}

type EventInfo struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Source    string    `json:"source,omitempty"`
	FirstSeen time.Time `json:"first_seen,omitempty"`
	LastSeen  time.Time `json:"last_seen,omitempty"`
	Count     int32     `json:"count"`
	Object    ObjectRef `json:"object"`
}

type ObjectRef struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type NodeInfo struct {
	Name           string              `json:"name"`
	Conditions     []ConditionInfo     `json:"conditions"`
	Capacity       map[string]string   `json:"capacity,omitempty"`
	Allocatable    map[string]string   `json:"allocatable,omitempty"`
	Taints         []TaintInfo         `json:"taints,omitempty"`
	KubeletVersion string              `json:"kubelet_version,omitempty"`
	ContainerRuntime string            `json:"container_runtime,omitempty"`
}

type TaintInfo struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

type StorageInfo struct {
	PVCs []PVCInfo `json:"pvcs,omitempty"`
	PVs  []PVInfo  `json:"pvs,omitempty"`
}

type PVCInfo struct {
	Name             string            `json:"name"`
	Namespace        string            `json:"namespace"`
	Status           string            `json:"status"`
	StorageClassName string            `json:"storage_class_name,omitempty"`
	Capacity         map[string]string `json:"capacity,omitempty"`
	AccessModes      []string          `json:"access_modes,omitempty"`
	VolumeName       string            `json:"volume_name,omitempty"`
	VolumeMode       string            `json:"volume_mode,omitempty"`
}

type PVInfo struct {
	Name             string            `json:"name"`
	Status           string            `json:"status"`
	StorageClassName string            `json:"storage_class_name,omitempty"`
	Capacity         map[string]string `json:"capacity,omitempty"`
	AccessModes      []string          `json:"access_modes,omitempty"`
	ClaimRef         *ObjectRef        `json:"claim_ref,omitempty"`
	Reason           string            `json:"reason,omitempty"`
	Message          string            `json:"message,omitempty"`
}

type NetworkInfo struct {
	Services     []ServiceInfo     `json:"services,omitempty"`
	Endpoints    []EndpointsInfo   `json:"endpoints,omitempty"`
	Ingresses    []IngressInfo     `json:"ingresses,omitempty"`
	NetworkPolicies []NetworkPolicyInfo `json:"network_policies,omitempty"`
}

type ServiceInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	ClusterIP   string            `json:"cluster_ip"`
	Selector    map[string]string `json:"selector,omitempty"`
	Ports       []PortInfo        `json:"ports"`
	HasEndpoints bool             `json:"has_endpoints"`
	ReadyCount  int               `json:"ready_count"`
	NotReadyCount int             `json:"not_ready_count"`
}

type EndpointsInfo struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	Addresses  []string `json:"addresses"`
	NotReady   []string `json:"not_ready"`
}

type IngressInfo struct {
	Name        string   `json:"name"`
	Namespace   string   `json:"namespace"`
	ClassName   string   `json:"class_name,omitempty"`
	Rules       []string `json:"rules,omitempty"`
	TLS         []string `json:"tls,omitempty"`
	LoadBalancer string  `json:"load_balancer,omitempty"`
}

type NetworkPolicyInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	PodSelector map[string]string `json:"pod_selector"`
	PolicyTypes []string          `json:"policy_types"`
	Ingress     []string          `json:"ingress,omitempty"`
	Egress      []string          `json:"egress,omitempty"`
}

func (b *PayloadBuilder) Build(data *collector.DiagnosticData, matchedRules []string) (*AIPayload, error) {
	payload := &AIPayload{
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Resource: ResourceInfo{
			Type:      data.ResourceType,
			Name:      data.ResourceName,
			Namespace: data.Namespace,
		},
		MatchedRules: matchedRules,
	}

	if data.PodData != nil && data.PodData.Pod != nil {
		payload.Pod = b.buildPodInfo(data.PodData)
	}

	if data.WorkloadData != nil {
		payload.Workload = b.buildWorkloadInfo(data.WorkloadData)
	}

	if data.EventsData != nil {
		payload.Events = b.buildEventsInfo(data.EventsData)
	}

	if data.NodeData != nil {
		payload.Node = b.buildNodeInfo(data.NodeData)
	}

	if data.StorageData != nil {
		payload.Storage = b.buildStorageInfo(data.StorageData)
	}

	if data.NetworkData != nil {
		payload.Network = b.buildNetworkInfo(data.NetworkData)
	}

	return payload, nil
}

func (b *PayloadBuilder) buildPodInfo(podData *collector.PodDiagnosticData) *PodInfo {
	pod := podData.Pod
	info := &PodInfo{
		Name:           pod.Name,
		Namespace:      pod.Namespace,
		Phase:          string(pod.Status.Phase),
		NodeName:       pod.Spec.NodeName,
		QoSClass:       string(pod.Status.QOSClass),
		RestartPolicy:  string(pod.Spec.RestartPolicy),
		ServiceAccount: pod.Spec.ServiceAccountName,
	}

	for _, c := range pod.Spec.Containers {
		info.Containers = append(info.Containers, b.buildContainerInfo(&c, pod, podData, false))
	}
	for _, c := range pod.Spec.InitContainers {
		info.InitContainers = append(info.InitContainers, b.buildContainerInfo(&c, pod, podData, true))
	}

	for _, cond := range pod.Status.Conditions {
		info.Conditions = append(info.Conditions, ConditionInfo{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	for _, v := range pod.Spec.Volumes {
		vol := VolumeInfo{Name: v.Name}
		if v.PersistentVolumeClaim != nil {
			vol.Type = "PersistentVolumeClaim"
			vol.ClaimName = v.PersistentVolumeClaim.ClaimName
		} else if v.ConfigMap != nil {
			vol.Type = "ConfigMap"
			vol.ConfigMap = v.ConfigMap.Name
		} else if v.Secret != nil {
			vol.Type = "Secret"
			vol.Secret = v.Secret.SecretName
		}
		info.Volumes = append(info.Volumes, vol)
	}

	for _, ref := range pod.OwnerReferences {
		info.OwnerReferences = append(info.OwnerReferences, OwnerRefInfo{
			APIVersion: ref.APIVersion,
			Kind:       ref.Kind,
			Name:       ref.Name,
			UID:        string(ref.UID),
			Controller: ref.Controller,
		})
	}

	return info
}

func (b *PayloadBuilder) buildContainerInfo(c *corev1.Container, pod *corev1.Pod, podData *collector.PodDiagnosticData, isInit bool) ContainerInfo {
	name := c.Name
	info := ContainerInfo{
		Name:    name,
		Image:   c.Image,
		Ports:   make([]PortInfo, 0),
		EnvVars: make([]EnvVarInfo, 0),
	}

	for _, p := range c.Ports {
		info.Ports = append(info.Ports, PortInfo{
			Name:          p.Name,
			ContainerPort: p.ContainerPort,
			Protocol:      string(p.Protocol),
		})
	}

	for _, e := range c.Env {
		ev := EnvVarInfo{Name: e.Name, Value: e.Value}
		if e.ValueFrom != nil {
			if e.ValueFrom.FieldRef != nil {
				ev.ValueFrom = fmt.Sprintf("fieldRef:%s", e.ValueFrom.FieldRef.FieldPath)
			} else if e.ValueFrom.ConfigMapKeyRef != nil {
				ev.ValueFrom = fmt.Sprintf("configMapKeyRef:%s/%s", e.ValueFrom.ConfigMapKeyRef.Name, e.ValueFrom.ConfigMapKeyRef.Key)
			} else if e.ValueFrom.SecretKeyRef != nil {
				ev.ValueFrom = fmt.Sprintf("secretKeyRef:%s/%s", e.ValueFrom.SecretKeyRef.Name, e.ValueFrom.SecretKeyRef.Key)
			}
		}
		info.EnvVars = append(info.EnvVars, ev)
	}

	for _, vm := range c.VolumeMounts {
		info.VolumeMounts = append(info.VolumeMounts, VolumeMountInfo{
			Name:      vm.Name,
			MountPath: vm.MountPath,
			ReadOnly:  vm.ReadOnly,
		})
	}

	if c.Resources.Requests != nil {
		info.Resources.Requests = make(map[string]string)
		for k, v := range c.Resources.Requests {
			info.Resources.Requests[string(k)] = v.String()
		}
	}
	if c.Resources.Limits != nil {
		info.Resources.Limits = make(map[string]string)
		for k, v := range c.Resources.Limits {
			info.Resources.Limits[string(k)] = v.String()
		}
	}

	if c.LivenessProbe != nil {
		info.LivenessProbe = b.buildProbeInfo(c.LivenessProbe)
	}
	if c.ReadinessProbe != nil {
		info.ReadinessProbe = b.buildProbeInfo(c.ReadinessProbe)
	}
	if c.StartupProbe != nil {
		info.StartupProbe = b.buildProbeInfo(c.StartupProbe)
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == name {
			info.Ready = cs.Ready
			info.RestartCount = cs.RestartCount
			info.ImageID = cs.ImageID
			if cs.State.Terminated != nil {
				info.State = "Terminated"
				info.ExitCode = cs.State.Terminated.ExitCode
				info.Reason = cs.State.Terminated.Reason
				info.Message = cs.State.Terminated.Message
				t := cs.State.Terminated.StartedAt.Time
				info.StartedAt = &t
				t2 := cs.State.Terminated.FinishedAt.Time
				info.FinishedAt = &t2
			} else if cs.State.Waiting != nil {
				info.State = "Waiting"
				info.Reason = cs.State.Waiting.Reason
				info.Message = cs.State.Waiting.Message
			} else if cs.State.Running != nil {
				info.State = "Running"
				t := cs.State.Running.StartedAt.Time
				info.StartedAt = &t
			}
			break
		}
	}

	if logs, ok := podData.ContainerLogs[name]; ok {
		lines := strings.Split(logs, "\n")
		if len(lines) > b.maxLogLines {
			lines = lines[len(lines)-b.maxLogLines:]
		}
		info.Logs = strings.Join(lines, "\n")
	}

	if prevLogs, ok := podData.ContainerLogs[name+"-previous"]; ok {
		lines := strings.Split(prevLogs, "\n")
		if len(lines) > b.maxLogLines {
			lines = lines[len(lines)-b.maxLogLines:]
		}
		info.PreviousLogs = strings.Join(lines, "\n")
	}

	return info
}

func (b *PayloadBuilder) buildProbeInfo(p *corev1.Probe) *ProbeInfo {
	probe := &ProbeInfo{
		InitialDelaySeconds: p.InitialDelaySeconds,
		PeriodSeconds:       p.PeriodSeconds,
		TimeoutSeconds:      p.TimeoutSeconds,
		FailureThreshold:    p.FailureThreshold,
		SuccessThreshold:    p.SuccessThreshold,
	}

	if p.HTTPGet != nil {
		probe.Type = "HTTP"
		probe.Path = p.HTTPGet.Path
		probe.Port = p.HTTPGet.Port.IntVal
	} else if p.TCPSocket != nil {
		probe.Type = "TCP"
		probe.Port = p.TCPSocket.Port.IntVal
	} else if p.Exec != nil {
		probe.Type = "Exec"
		probe.Path = strings.Join(p.Exec.Command, " ")
	} else if p.GRPC != nil {
		probe.Type = "gRPC"
		probe.Port = p.GRPC.Port
	}

	return probe
}

func (b *PayloadBuilder) buildWorkloadInfo(w *collector.WorkloadInfo) *WorkloadInfo {
	info := &WorkloadInfo{
		Kind:       w.Kind,
		Name:       w.Name,
		Namespace:  w.Namespace,
		Replicas:   w.Replicas,
		ReadyReplicas: w.ReadyReplicas,
		Selector:   w.Selector,
		Conditions: make([]ConditionInfo, 0),
	}

	for _, cond := range w.Conditions {
		info.Conditions = append(info.Conditions, ConditionInfo{
			Type:    cond.Type,
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	if w.PodTemplate != nil {
		info.Template = &PodTemplateInfo{
			Labels:      w.PodTemplate.Labels,
			Annotations: w.PodTemplate.Annotations,
		}
		spec := w.PodTemplate.Spec
		info.Template.Spec = b.buildPodSpecInfo(&spec)
	}

	return info
}

func (b *PayloadBuilder) buildPodSpecInfo(spec *corev1.PodSpec) *PodSpecInfo {
	info := &PodSpecInfo{
		ServiceAccountName: spec.ServiceAccountName,
		Containers:         make([]ContainerInfo, 0),
		InitContainers:     make([]ContainerInfo, 0),
		Volumes:            make([]VolumeInfo, 0),
		NodeSelector:       spec.NodeSelector,
		Tolerations:        make([]TolerationInfo, 0),
	}

	for _, c := range spec.Containers {
		info.Containers = append(info.Containers, b.buildContainerInfoFromSpec(&c))
	}
	for _, c := range spec.InitContainers {
		info.InitContainers = append(info.InitContainers, b.buildContainerInfoFromSpec(&c))
	}

	for _, v := range spec.Volumes {
		vol := VolumeInfo{Name: v.Name}
		if v.PersistentVolumeClaim != nil {
			vol.Type = "PersistentVolumeClaim"
			vol.ClaimName = v.PersistentVolumeClaim.ClaimName
		} else if v.ConfigMap != nil {
			vol.Type = "ConfigMap"
			vol.ConfigMap = v.ConfigMap.Name
		} else if v.Secret != nil {
			vol.Type = "Secret"
			vol.Secret = v.Secret.SecretName
		}
		info.Volumes = append(info.Volumes, vol)
	}

	for _, t := range spec.Tolerations {
		info.Tolerations = append(info.Tolerations, TolerationInfo{
			Key:      t.Key,
			Operator: string(t.Operator),
			Value:    t.Value,
			Effect:   string(t.Effect),
		})
	}

	if spec.Affinity != nil {
		info.Affinity = "configured"
	}

	return info
}

func (b *PayloadBuilder) buildContainerInfoFromSpec(c *corev1.Container) ContainerInfo {
	info := ContainerInfo{
		Name:    c.Name,
		Image:   c.Image,
		Ports:   make([]PortInfo, 0),
		EnvVars: make([]EnvVarInfo, 0),
	}

	for _, p := range c.Ports {
		info.Ports = append(info.Ports, PortInfo{
			Name:          p.Name,
			ContainerPort: p.ContainerPort,
			Protocol:      string(p.Protocol),
		})
	}

	for _, e := range c.Env {
		ev := EnvVarInfo{Name: e.Name, Value: e.Value}
		if e.ValueFrom != nil {
			if e.ValueFrom.FieldRef != nil {
				ev.ValueFrom = fmt.Sprintf("fieldRef:%s", e.ValueFrom.FieldRef.FieldPath)
			} else if e.ValueFrom.ConfigMapKeyRef != nil {
				ev.ValueFrom = fmt.Sprintf("configMapKeyRef:%s/%s", e.ValueFrom.ConfigMapKeyRef.Name, e.ValueFrom.ConfigMapKeyRef.Key)
			} else if e.ValueFrom.SecretKeyRef != nil {
				ev.ValueFrom = fmt.Sprintf("secretKeyRef:%s/%s", e.ValueFrom.SecretKeyRef.Name, e.ValueFrom.SecretKeyRef.Key)
			}
		}
		info.EnvVars = append(info.EnvVars, ev)
	}

	for _, vm := range c.VolumeMounts {
		info.VolumeMounts = append(info.VolumeMounts, VolumeMountInfo{
			Name:      vm.Name,
			MountPath: vm.MountPath,
			ReadOnly:  vm.ReadOnly,
		})
	}

	if c.Resources.Requests != nil {
		info.Resources.Requests = make(map[string]string)
		for k, v := range c.Resources.Requests {
			info.Resources.Requests[string(k)] = v.String()
		}
	}
	if c.Resources.Limits != nil {
		info.Resources.Limits = make(map[string]string)
		for k, v := range c.Resources.Limits {
			info.Resources.Limits[string(k)] = v.String()
		}
	}

	if c.LivenessProbe != nil {
		info.LivenessProbe = b.buildProbeInfo(c.LivenessProbe)
	}
	if c.ReadinessProbe != nil {
		info.ReadinessProbe = b.buildProbeInfo(c.ReadinessProbe)
	}
	if c.StartupProbe != nil {
		info.StartupProbe = b.buildProbeInfo(c.StartupProbe)
	}

	return info
}

func (b *PayloadBuilder) buildEventsInfo(eventsData *collector.EventsData) []EventInfo {
	var events []EventInfo

	addEvents := func(evts []corev1.Event, source string) {
		for _, e := range evts {
			if len(events) >= b.maxEvents {
				return
			}
			events = append(events, EventInfo{
				Type:      e.Type,
				Reason:    e.Reason,
				Message:   e.Message,
				Source:    source,
				FirstSeen: e.FirstTimestamp.Time,
				LastSeen:  e.LastTimestamp.Time,
				Count:     e.Count,
				Object: ObjectRef{
					Kind:      e.InvolvedObject.Kind,
					Name:      e.InvolvedObject.Name,
					Namespace: e.InvolvedObject.Namespace,
				},
			})
		}
	}

	addEvents(eventsData.PodEvents, "pod")
	addEvents(eventsData.NamespaceEvents, "namespace")
	addEvents(eventsData.RelatedEvents, "related")

	return events
}

func (b *PayloadBuilder) buildNodeInfo(nodeData *collector.NodeInfo) *NodeInfo {
	info := &NodeInfo{
		Name:            nodeData.Name,
		KubeletVersion:  nodeData.KubeletVersion,
		ContainerRuntime: nodeData.ContainerRuntime,
		Conditions:      make([]ConditionInfo, 0),
		Capacity:        make(map[string]string),
		Allocatable:     make(map[string]string),
		Taints:          make([]TaintInfo, 0),
	}

	for _, cond := range nodeData.Conditions {
		info.Conditions = append(info.Conditions, ConditionInfo{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	for k, v := range nodeData.Capacity {
		info.Capacity[string(k)] = v.String()
	}
	for k, v := range nodeData.Allocatable {
		info.Allocatable[string(k)] = v.String()
	}

	for _, t := range nodeData.Taints {
		info.Taints = append(info.Taints, TaintInfo{
			Key:    t.Key,
			Value:  t.Value,
			Effect: string(t.Effect),
		})
	}

	return info
}

func (b *PayloadBuilder) buildStorageInfo(storageData *collector.PVCInfo) *StorageInfo {
	info := &StorageInfo{
		PVCs: make([]PVCInfo, 0),
		PVs:  make([]PVInfo, 0),
	}

	info.PVCs = append(info.PVCs, PVCInfo{
		Name:             storageData.Name,
		Namespace:        storageData.Namespace,
		Status:           string(storageData.Status.Phase),
		StorageClassName: "",
		Capacity:         make(map[string]string),
		AccessModes:      make([]string, 0),
		VolumeName:       storageData.VolumeName,
		VolumeMode:       "",
	})

	if storageData.Spec.StorageClassName != nil {
		info.PVCs[0].StorageClassName = *storageData.Spec.StorageClassName
	}

	for k, v := range storageData.Capacity {
		info.PVCs[0].Capacity[string(k)] = v.String()
	}
	for _, am := range storageData.AccessModes {
		info.PVCs[0].AccessModes = append(info.PVCs[0].AccessModes, string(am))
	}
	if storageData.VolumeMode != nil {
		info.PVCs[0].VolumeMode = string(*storageData.VolumeMode)
	}

	if storageData.BoundPV != nil {
		pv := storageData.BoundPV
		info.PVs = append(info.PVs, PVInfo{
			Name:             pv.Name,
			Status:           string(pv.Status.Phase),
			StorageClassName: pv.StorageClassName,
			Capacity:         make(map[string]string),
			AccessModes:      make([]string, 0),
			Reason:           pv.Reason,
			Message:          pv.Message,
		})
		for k, v := range pv.Capacity {
			info.PVs[0].Capacity[string(k)] = v.String()
		}
		for _, am := range pv.AccessModes {
			info.PVs[0].AccessModes = append(info.PVs[0].AccessModes, string(am))
		}
		if pv.ClaimRef != nil {
			info.PVs[0].ClaimRef = &ObjectRef{
				Kind:      pv.ClaimRef.Kind,
				Name:      pv.ClaimRef.Name,
				Namespace: pv.ClaimRef.Namespace,
			}
		}
	}

	return info
}

func (b *PayloadBuilder) buildNetworkInfo(networkData *collector.ServiceInfo) *NetworkInfo {
	info := &NetworkInfo{
		Services:        make([]ServiceInfo, 0),
		Endpoints:       make([]EndpointsInfo, 0),
		Ingresses:       make([]IngressInfo, 0),
		NetworkPolicies: make([]NetworkPolicyInfo, 0),
	}

	svc := ServiceInfo{
		Name:       networkData.Name,
		Namespace:  networkData.Namespace,
		Type:       string(networkData.Type),
		ClusterIP:  networkData.ClusterIP,
		Selector:   networkData.Selector,
		Ports:      make([]PortInfo, 0),
	}

	for _, p := range networkData.Ports {
		svc.Ports = append(svc.Ports, PortInfo{
			Name:          p.Name,
			ContainerPort: p.Port,
			Protocol:      string(p.Protocol),
		})
	}

	if networkData.Endpoints != nil {
		svc.HasEndpoints = len(networkData.Endpoints.Addresses) > 0
		svc.ReadyCount = len(networkData.Endpoints.Addresses)
		svc.NotReadyCount = len(networkData.Endpoints.NotReadyAddresses)
	}

	info.Services = append(info.Services, svc)

	if networkData.Endpoints != nil {
		ep := EndpointsInfo{
			Name:      networkData.Endpoints.Name,
			Namespace: networkData.Endpoints.Namespace,
			Addresses: make([]string, 0),
			NotReady:  make([]string, 0),
		}
		for _, addr := range networkData.Endpoints.Addresses {
			ep.Addresses = append(ep.Addresses, addr.IP)
		}
		for _, addr := range networkData.Endpoints.NotReadyAddresses {
			ep.NotReady = append(ep.NotReady, addr.IP)
		}
		info.Endpoints = append(info.Endpoints, ep)
	}

	return info
}

func (b *PayloadBuilder) ToJSON(payload *AIPayload) (string, error) {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (b *PayloadBuilder) BuildPrompt(payload *AIPayload) string {
	jsonStr, _ := b.ToJSON(payload)

	prompt := fmt.Sprintf(`You are kube-triage, an expert Kubernetes SRE. Analyze the following cluster state and provide a root cause analysis with specific remediation commands.

CLUSTER STATE (JSON):
%s

MATCHED DETERMINISTIC RULES: %v

Please provide:
1. ROOT CAUSE: One paragraph explaining the underlying issue
2. REMEDIATION: Numbered list of exact kubectl commands to fix the issue
3. PREVENTION: 2-3 best practices to prevent recurrence

Be concise and actionable. Focus on the most likely root cause.`, jsonStr, payload.MatchedRules)

	return prompt
}