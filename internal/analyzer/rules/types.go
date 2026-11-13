package rules

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

type RuleSet struct {
	Version        string    `yaml:"version"`
	Description    string    `yaml:"description"`
	Rules          []Rule    `yaml:"rules"`
	LoadedAt       time.Time `yaml:"-"`
	ConfigFilePath string    `yaml:"-"`
}

type Rule struct {
	ID               string                 `yaml:"id"`
	Name             string                 `yaml:"name"`
	Category         string                 `yaml:"category"`
	Severity         string                 `yaml:"severity"`
	Trigger          Trigger                `yaml:"trigger"`
	DiagnosisTemplate string                `yaml:"diagnosisTemplate"`
	RemediationSteps []RemediationStep      `yaml:"remediationSteps"`
	Prevention       []string               `yaml:"prevention"`
	References       []string               `yaml:"references"`
	Enabled          bool                   `yaml:"enabled,omitempty"`
}

type Trigger struct {
	ContainerState *ContainerStateTrigger `yaml:"containerState,omitempty"`
	Event          *EventTrigger          `yaml:"event,omitempty"`
	Resource       *ResourceTrigger       `yaml:"resource,omitempty"`
	NodeCondition  *NodeConditionTrigger  `yaml:"nodeCondition,omitempty"`
}

type ContainerStateTrigger struct {
	Terminated *TerminatedTrigger `yaml:"terminated,omitempty"`
	Waiting    *WaitingTrigger    `yaml:"waiting,omitempty"`
	Running    *RunningTrigger    `yaml:"running,omitempty"`
	IsInitContainer bool           `yaml:"isInitContainer,omitempty"`
}

type TerminatedTrigger struct {
	ExitCode int32    `yaml:"exitCode,omitempty"`
	Reason   string   `yaml:"reason,omitempty"`
	Signal   int32    `yaml:"signal,omitempty"`
}

type WaitingTrigger struct {
	Reason []string `yaml:"reason,omitempty"`
}

type RunningTrigger struct{}

type EventTrigger struct {
	Reason []string `yaml:"reason,omitempty"`
	Type   string   `yaml:"type,omitempty"`
}

type ResourceTrigger struct {
	Kind   string `yaml:"kind,omitempty"`
	Status string `yaml:"status,omitempty"`
}

type NodeConditionTrigger struct {
	Type   []string `yaml:"type,omitempty"`
	Status string   `yaml:"status,omitempty"`
}

type RemediationStep struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Snippet     string `yaml:"snippet"`
}

type MatchContext struct {
	PodName          string
	Namespace        string
	ContainerName    string
	DeploymentName   string
	RestartCount     int32
	LimitMemory      string
	Image            string
	PVCName          string
	StorageClassName string
	NodeName         string
	ConditionType    string
	ServiceAccount   string
	ProbeType        string
	EventMessage     string
	IsInitContainer  bool
	Pod              *corev1.Pod
	Event            *corev1.Event
	Node             interface{}
	PVC              interface{}
}

type MatchResult struct {
	Rule       *Rule
	Context    MatchContext
	MatchedAt  time.Time
	Score      int
}