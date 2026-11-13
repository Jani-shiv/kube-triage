package analyzer

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"kube-triage/internal/analyzer/ai"
	"kube-triage/internal/analyzer/rules"
	"kube-triage/internal/collector"
)

type Engine struct {
	provider     string
	model        string
	endpoint     string
	ruleSet      *rules.RuleSet
	matcher      *rules.Matcher
	aiManager    *ai.Manager
	payloadBuilder *ai.PayloadBuilder
}

func NewEngine(provider, model, endpoint string) *Engine {
	ruleSet, err := rules.LoadRules("")
	if err != nil {
		ruleSet = &rules.RuleSet{Rules: []rules.Rule{}}
	}
	matcher := rules.NewMatcher(ruleSet)
	aiManager := ai.NewManager(provider, model, endpoint)
	payloadBuilder := ai.NewPayloadBuilder(100, 50, 20)

	return &Engine{
		provider:       provider,
		model:          model,
		endpoint:       endpoint,
		ruleSet:        ruleSet,
		matcher:        matcher,
		aiManager:      aiManager,
		payloadBuilder: payloadBuilder,
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
	AIEnhanced      bool
	AIProvider      string
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

	result.AIEnhanced = true
	return e.analyzeWithAI(ctx, result, data)
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

func (e *Engine) analyzeWithAI(ctx context.Context, base *AnalysisResult, data *collector.DiagnosticData) (*AnalysisResult, error) {
	payload, err := e.payloadBuilder.Build(data, base.MatchedRules)
	if err != nil {
		return base, nil
	}

	prompt := e.payloadBuilder.BuildPrompt(payload)

	aiResponse, err := e.aiManager.Analyze(ctx, prompt)
	if err != nil {
		return base, nil
	}

	return e.parseAIResponse(base, aiResponse, e.aiManager.GetPrimaryProvider()), nil
}

func (e *Engine) parseAIResponse(base *AnalysisResult, aiResponse, provider string) *AnalysisResult {
	base.AIEnhanced = true
	base.AIProvider = provider

	lines := strings.Split(aiResponse, "\n")
	
	var rootCauseLines []string
	var remediationLines []string
	var preventionLines []string
	
	currentSection := ""
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		
		if strings.Contains(lower, "root cause") || strings.Contains(lower, "root_cause") {
			currentSection = "rootcause"
			continue
		}
		if strings.Contains(lower, "remediation") || strings.Contains(lower, "remediation:") {
			currentSection = "remediation"
			continue
		}
		if strings.Contains(lower, "prevention") || strings.Contains(lower, "prevention:") {
			currentSection = "prevention"
			continue
		}
		
		if trimmed == "" {
			continue
		}
		
		switch currentSection {
		case "rootcause":
			rootCauseLines = append(rootCauseLines, trimmed)
		case "remediation":
			remediationLines = append(remediationLines, trimmed)
		case "prevention":
			preventionLines = append(preventionLines, trimmed)
		}
	}
	
	if len(rootCauseLines) > 0 {
		base.RootCause = strings.Join(rootCauseLines, " ")
	}
	
	if len(remediationLines) > 0 {
		base.Remediation = []RemediationStep{}
		for i, line := range remediationLines {
			cmd := strings.TrimSpace(strings.TrimPrefix(line, fmt.Sprintf("%d.", i+1)))
			cmd = strings.TrimSpace(strings.TrimPrefix(cmd, "-"))
			if cmd != "" {
				base.Remediation = append(base.Remediation, RemediationStep{
					Description: fmt.Sprintf("AI-suggested remediation step %d", i+1),
					Command:     cmd,
					Type:        "command",
				})
			}
		}
	}
	
	if len(preventionLines) > 0 {
		base.Prevention = preventionLines
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

func (e *Engine) GetAIProviders() []string {
	if e.aiManager == nil {
		return []string{}
	}
	return e.aiManager.GetAvailableProviders()
}