package analyzer

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"kube-triage/internal/analyzer/rules"
	"kube-triage/internal/collector"
)

type Engine struct {
	provider   string
	model      string
	endpoint   string
	ruleSet    *rules.RuleSet
	matcher    *rules.Matcher
}

func NewEngine(provider, model, endpoint string) *Engine {
	ruleSet, err := rules.LoadRules("")
	if err != nil {
		ruleSet = &rules.RuleSet{Rules: []rules.Rule{}}
	}
	matcher := rules.NewMatcher(ruleSet)

	return &Engine{
		provider: provider,
		model:    model,
		endpoint: endpoint,
		ruleSet:  ruleSet,
		matcher:  matcher,
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
	MatchedRules    []string
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
		Severity:     "LOW",
		Summary:      "No obvious failure patterns detected",
		RootCause:    "The resource appears to be in an unknown state. Manual investigation recommended.",
		Remediation: []RemediationStep{
			{
				Description: "Get detailed resource description",
				Command:     fmt.Sprintf("kubectl describe %s %s -n %s", data.ResourceType, data.ResourceName, data.Namespace),
				Type:        "command",
			},
		},
		Prevention: []string{
			"Implement comprehensive monitoring and alerting",
			"Regular cluster health checks",
		},
	}

	if data.PodData != nil && data.PodData.Pod != nil {
		podMatches := e.matcher.MatchPod(data.PodData.Pod)
		if len(podMatches) > 0 {
			return e.buildResultFromMatches(result, podMatches, data), nil
		}
	}

	if data.EventsData != nil && len(data.EventsData.PodEvents) > 0 {
		for _, event := range data.EventsData.PodEvents {
			eventMatches := e.matcher.MatchEvent(&event, data.PodData.Pod)
			if len(eventMatches) > 0 {
				return e.buildResultFromMatches(result, eventMatches, data), nil
			}
		}
	}

	if data.StorageData != nil {
		pvcMatches := e.matcher.MatchPVC(data.StorageData, data.Namespace, data.ResourceName)
		if len(pvcMatches) > 0 {
			return e.buildResultFromMatches(result, pvcMatches, data), nil
		}
	}

	if data.NodeData != nil {
		nodeMatches := e.matcher.MatchNode(data.NodeData, data.ResourceName)
		if len(nodeMatches) > 0 {
			return e.buildResultFromMatches(result, nodeMatches, data), nil
		}
	}

	return result, nil
}

func (e *Engine) buildResultFromMatches(base *AnalysisResult, matches []rules.MatchResult, data *collector.DiagnosticData) *AnalysisResult {
	if len(matches) == 0 {
		return base
	}

	bestMatch := matches[0]
	for _, m := range matches {
		if m.Score > bestMatch.Score {
			bestMatch = m
		}
	}

	rule := bestMatch.Rule
	ctx := bestMatch.Context

	base.Severity = rule.Severity
	base.Summary = rule.Name
	base.RootCause = e.renderTemplate(rule.DiagnosisTemplate, ctx)
	base.Remediation = e.renderRemediationSteps(rule.RemediationSteps, ctx)
	base.Prevention = rule.Prevention
	base.MatchedRules = []string{rule.ID}

	for _, m := range matches[1:] {
		base.MatchedRules = append(base.MatchedRules, m.Rule.ID)
	}

	return base
}

func (e *Engine) renderTemplate(tmpl string, ctx rules.MatchContext) string {
	t, err := template.New("diag").Parse(tmpl)
	if err != nil {
		return tmpl
	}

	var buf strings.Builder
	data := map[string]interface{}{
		"PodName":           ctx.PodName,
		"Namespace":         ctx.Namespace,
		"ContainerName":     ctx.ContainerName,
		"DeploymentName":    ctx.DeploymentName,
		"RestartCount":      ctx.RestartCount,
		"LimitMemory":       ctx.LimitMemory,
		"Image":             ctx.Image,
		"PVCName":           ctx.PVCName,
		"StorageClassName":  ctx.StorageClassName,
		"NodeName":          ctx.NodeName,
		"ConditionType":     ctx.ConditionType,
		"ServiceAccount":    ctx.ServiceAccount,
		"ProbeType":         ctx.ProbeType,
		"EventMessage":      ctx.EventMessage,
		"IsInitContainer":   ctx.IsInitContainer,
	}

	if err := t.Execute(&buf, data); err != nil {
		return tmpl
	}

	return buf.String()
}

func (e *Engine) renderRemediationSteps(steps []rules.RemediationStep, ctx rules.MatchContext) []RemediationStep {
	var result []RemediationStep
	for _, step := range steps {
		cmd := e.renderTemplate(step.Snippet, ctx)
		result = append(result, RemediationStep{
			Description: step.Description,
			Command:     cmd,
			Type:        step.Type,
		})
	}
	return result
}

func (e *Engine) ReloadRules() error {
	ruleSet, err := rules.LoadRules("")
	if err != nil {
		return err
	}
	e.ruleSet = ruleSet
	e.matcher = rules.NewMatcher(ruleSet)
	return nil
}

func (e *Engine) GetLoadedRules() []rules.Rule {
	if e.ruleSet == nil {
		return []rules.Rule{}
	}
	return e.ruleSet.GetEnabledRules()
}