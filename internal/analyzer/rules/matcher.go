package rules

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

type Matcher struct {
	rules []Rule
}

func NewMatcher(ruleSet *RuleSet) *Matcher {
	return &Matcher{
		rules: ruleSet.GetEnabledRules(),
	}
}

func (m *Matcher) MatchPod(pod *corev1.Pod) []MatchResult {
	var results []MatchResult

	for _, container := range pod.Spec.Containers {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name != container.Name {
				continue
			}
			ctx := buildMatchContext(pod, &container, &cs, false)
			results = append(results, m.matchRules(ctx)...)
		}
	}

	for _, cs := range pod.Status.InitContainerStatuses {
		var container corev1.Container
		found := false
		for _, c := range pod.Spec.InitContainers {
			if c.Name == cs.Name {
				container = c
				found = true
				break
			}
		}
		if !found {
			continue
		}
		ctx := buildMatchContext(pod, &container, &cs, true)
		results = append(results, m.matchRules(ctx)...)
	}

	return results
}

func (m *Matcher) MatchEvent(event *corev1.Event, pod *corev1.Pod) []MatchResult {
	var results []MatchResult

	ctx := MatchContext{
		PodName:      event.InvolvedObject.Name,
		Namespace:    event.InvolvedObject.Namespace,
		EventMessage: event.Message,
		Event:        event,
	}

	if pod != nil {
		ctx.Pod = pod
		if len(pod.OwnerReferences) > 0 {
			for _, ref := range pod.OwnerReferences {
				if ref.Kind == "ReplicaSet" {
					ctx.DeploymentName = ref.Name
				}
				if ref.Kind == "Deployment" {
					ctx.DeploymentName = ref.Name
				}
			}
		}
	}

	results = append(results, m.matchRules(ctx)...)
	return results
}

func (m *Matcher) MatchPVC(pvc interface{}, namespace, name string) []MatchResult {
	ctx := MatchContext{
		Namespace: namespace,
		PVCName:   name,
		PVC:       pvc,
	}
	return m.matchRules(ctx)
}

func (m *Matcher) MatchNode(node interface{}, name string) []MatchResult {
	ctx := MatchContext{
		NodeName: name,
		Node:     node,
	}
	return m.matchRules(ctx)
}

func (m *Matcher) matchRules(ctx MatchContext) []MatchResult {
	var results []MatchResult

	for _, rule := range m.rules {
		if matched, score := m.evaluateRule(rule, ctx); matched {
			results = append(results, MatchResult{
				Rule:      &rule,
				Context:   ctx,
				MatchedAt: time.Now(),
				Score:     score,
			})
		}
	}

	return results
}

func (m *Matcher) evaluateRule(rule Rule, ctx MatchContext) (bool, int) {
	score := 0

	if rule.Trigger.ContainerState != nil {
		if matched, s := m.matchContainerState(rule.Trigger.ContainerState, ctx); matched {
			score += s
		} else {
			return false, 0
		}
	}

	if rule.Trigger.Event != nil {
		if matched, s := m.matchEvent(rule.Trigger.Event, ctx); matched {
			score += s
		} else {
			return false, 0
		}
	}

	if rule.Trigger.Resource != nil {
		if matched, s := m.matchResource(rule.Trigger.Resource, ctx); matched {
			score += s
		} else {
			return false, 0
		}
	}

	if rule.Trigger.NodeCondition != nil {
		if matched, s := m.matchNodeCondition(rule.Trigger.NodeCondition, ctx); matched {
			score += s
		} else {
			return false, 0
		}
	}

	return true, score
}

func (m *Matcher) matchContainerState(trigger *ContainerStateTrigger, ctx MatchContext) (bool, int) {
	if trigger.IsInitContainer && !ctx.IsInitContainer {
		return false, 0
	}

	if trigger.Terminated != nil {
		if ctx.Pod == nil {
			return false, 0
		}
		for _, cs := range ctx.Pod.Status.ContainerStatuses {
			if cs.Name != ctx.ContainerName {
				continue
			}
			if cs.State.Terminated == nil {
				return false, 0
			}
			if trigger.Terminated.ExitCode != 0 && cs.State.Terminated.ExitCode != trigger.Terminated.ExitCode {
				return false, 0
			}
			if trigger.Terminated.Reason != "" && cs.State.Terminated.Reason != trigger.Terminated.Reason {
				return false, 0
			}
			return true, 100
		}
		for _, cs := range ctx.Pod.Status.InitContainerStatuses {
			if cs.Name != ctx.ContainerName {
				continue
			}
			if cs.State.Terminated == nil {
				return false, 0
			}
			if trigger.Terminated.ExitCode != 0 && cs.State.Terminated.ExitCode != trigger.Terminated.ExitCode {
				return false, 0
			}
			if trigger.Terminated.Reason != "" && cs.State.Terminated.Reason != trigger.Terminated.Reason {
				return false, 0
			}
			return true, 100
		}
		return false, 0
	}

	if trigger.Waiting != nil {
		if ctx.Pod == nil {
			return false, 0
		}
		for _, cs := range ctx.Pod.Status.ContainerStatuses {
			if cs.Name != ctx.ContainerName {
				continue
			}
			if cs.State.Waiting == nil {
				return false, 0
			}
			for _, reason := range trigger.Waiting.Reason {
				if cs.State.Waiting.Reason == reason {
					return true, 100
				}
			}
			return false, 0
		}
		for _, cs := range ctx.Pod.Status.InitContainerStatuses {
			if cs.Name != ctx.ContainerName {
				continue
			}
			if cs.State.Waiting == nil {
				return false, 0
			}
			for _, reason := range trigger.Waiting.Reason {
				if cs.State.Waiting.Reason == reason {
					return true, 100
				}
			}
			return false, 0
		}
		return false, 0
	}

	return false, 0
}

func (m *Matcher) matchEvent(trigger *EventTrigger, ctx MatchContext) (bool, int) {
	if ctx.Event == nil {
		return false, 0
	}

	if trigger.Type != "" && ctx.Event.Type != trigger.Type {
		return false, 0
	}

	if len(trigger.Reason) > 0 {
		found := false
		for _, reason := range trigger.Reason {
			if ctx.Event.Reason == reason {
				found = true
				break
			}
		}
		if !found {
			return false, 0
		}
	}

	return true, 90
}

func (m *Matcher) matchResource(trigger *ResourceTrigger, ctx MatchContext) (bool, int) {
	if trigger.Kind == "PersistentVolumeClaim" && trigger.Status == "Pending" {
		if ctx.PVC == nil {
			return false, 0
		}
		return true, 95
	}
	return false, 0
}

func (m *Matcher) matchNodeCondition(trigger *NodeConditionTrigger, ctx MatchContext) (bool, int) {
	if ctx.Node == nil {
		return false, 0
	}
	return false, 0
}

func buildMatchContext(pod *corev1.Pod, container *corev1.Container, cs *corev1.ContainerStatus, isInit bool) MatchContext {
	ctx := MatchContext{
		PodName:        pod.Name,
		Namespace:      pod.Namespace,
		ContainerName:  container.Name,
		RestartCount:   cs.RestartCount,
		Image:          container.Image,
		IsInitContainer: isInit,
		Pod:            pod,
	}

	for _, ref := range pod.OwnerReferences {
		if ref.Kind == "ReplicaSet" {
			ctx.DeploymentName = ref.Name
		}
		if ref.Kind == "Deployment" {
			ctx.DeploymentName = ref.Name
		}
	}

	if cs.State.Terminated != nil {
		if cs.State.Terminated.ExitCode == 137 && cs.State.Terminated.Reason == "OOMKilled" {
			for _, c := range pod.Spec.Containers {
				if c.Name == container.Name && c.Resources.Limits != nil {
					if mem, ok := c.Resources.Limits[corev1.ResourceMemory]; ok {
						ctx.LimitMemory = mem.String()
					}
				}
			}
		}
	}

	return ctx
}