package renderer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"

	"kube-triage/internal/analyzer"
)

type Printer struct {
	outputFormat string
}

func NewPrinter(outputFormat string) *Printer {
	return &Printer{
		outputFormat: outputFormat,
	}
}

func (p *Printer) Render(result *analyzer.AnalysisResult, dryRun bool) error {
	switch p.outputFormat {
	case "json":
		return p.renderJSON(result)
	case "yaml":
		return p.renderYAML(result)
	default:
		return p.renderTerminal(result, dryRun)
	}
}

func (p *Printer) renderTerminal(result *analyzer.AnalysisResult, dryRun bool) error {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Background(lipgloss.Color("#1a1a2e")).
		Padding(0, 1).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		MarginTop(1).
		MarginBottom(0)

	bodyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		MarginLeft(2)

	commandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Background(lipgloss.Color("#0a0a1a")).
		Padding(0, 1).
		MarginLeft(4)

	severityStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	switch result.Severity {
	case "CRITICAL":
		severityStyle = severityStyle.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#FF0000"))
	case "HIGH":
		severityStyle = severityStyle.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#FF8800"))
	case "MEDIUM":
		severityStyle = severityStyle.Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#FFFF00"))
	case "LOW":
		severityStyle = severityStyle.Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#00FF00"))
	}

	fmt.Println(titleStyle.Render(" kube-triage v1.0.0 — AI Kubernetes Diagnostic Engine "))
	fmt.Println()

	fmt.Println(headerStyle.Render(" 🎯 Target Resource"))
	fmt.Println(bodyStyle.Render(fmt.Sprintf("   %s/%s (Namespace: %s)", result.ResourceType, result.ResourceName, result.Namespace)))
	fmt.Println()

	fmt.Println(headerStyle.Render(" 🔍 Failure State"))
	fmt.Println(bodyStyle.Render(fmt.Sprintf("   %s", result.Summary)))
	fmt.Println()

	severityLabel := severityStyle.Render(" " + result.Severity + " ")
	fmt.Println(headerStyle.Render(" ⚠️  Severity Level"))
	fmt.Println(bodyStyle.Render(fmt.Sprintf("   %s", severityLabel)))
	fmt.Println()

	fmt.Println(headerStyle.Render(" 🧠 Root Cause Analysis"))
	fmt.Println(bodyStyle.Render("   " + result.RootCause))
	fmt.Println()

	fmt.Println(headerStyle.Render(" 🛠️ Recommended Remediation"))
	for i, step := range result.Remediation {
		fmt.Println(bodyStyle.Render(fmt.Sprintf("   %d. %s", i+1, step.Description)))
		if step.Command != "" {
			fmt.Println(commandStyle.Render("   $ " + step.Command))
		}
		if dryRun {
			fmt.Println(bodyStyle.Render("      [DRY-RUN] Command would be executed above"))
		}
	}
	fmt.Println()

	if len(result.Prevention) > 0 {
		fmt.Println(headerStyle.Render(" 🛡️ Prevention & Best Practices"))
		for _, tip := range result.Prevention {
			fmt.Println(bodyStyle.Render("   • " + tip))
		}
		fmt.Println()
	}

	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(" ─────────────────────────────────────────────────────────────"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  kube-triage — Analysis complete"))
	fmt.Println()

	return nil
}

func (p *Printer) renderJSON(result *analyzer.AnalysisResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func (p *Printer) renderYAML(result *analyzer.AnalysisResult) error {
	return fmt.Errorf("YAML output not yet implemented")
}