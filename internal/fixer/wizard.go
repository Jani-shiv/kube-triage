package fixer

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"kube-triage/internal/analyzer"
)

type WizardModel struct {
	result        *analyzer.AnalysisResult
	currentStep   int
	confirmed     []bool
	executed      []bool
	quitting      bool
	aborted       bool
	allConfirmed  bool
}

func NewWizard(result *analyzer.AnalysisResult) *WizardModel {
	confirmed := make([]bool, len(result.Remediation))
	executed := make([]bool, len(result.Remediation))
	return &WizardModel{
		result:       result,
		currentStep:  0,
		confirmed:    confirmed,
		executed:     executed,
	}
}

func (m *WizardModel) Init() tea.Cmd {
	return nil
}

func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			m.aborted = true
			return m, tea.Quit
		case "y", "Y":
			if m.currentStep < len(m.result.Remediation) {
				m.confirmed[m.currentStep] = true
				if m.currentStep < len(m.result.Remediation)-1 {
					m.currentStep++
				} else {
					m.allConfirmed = true
				}
			}
		case "n", "N":
			if m.currentStep < len(m.result.Remediation) {
				m.confirmed[m.currentStep] = false
				if m.currentStep < len(m.result.Remediation)-1 {
					m.currentStep++
				} else {
					m.allConfirmed = true
				}
			}
		case "a", "A":
			for i := range m.confirmed {
				m.confirmed[i] = true
			}
			m.allConfirmed = true
		case "s", "S":
			m.aborted = true
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.allConfirmed && !m.quitting {
				m.quitting = true
				return m, tea.Quit
			}
		case "right", "l":
			if m.currentStep < len(m.result.Remediation)-1 {
				m.currentStep++
			}
		case "left", "h":
			if m.currentStep > 0 {
				m.currentStep--
			}
		}
	}
	return m, nil
}

func (m *WizardModel) View() string {
	if m.quitting {
		if m.aborted {
			return abortStyle.Render("🛑 Remediation aborted by user.")
		}
		if m.allConfirmed {
			var executedCount int
			for _, c := range m.confirmed {
				if c {
					executedCount++
				}
			}
			return successStyle.Render(fmt.Sprintf("✅ Remediation complete! %d/%d commands confirmed.", executedCount, len(m.result.Remediation)))
		}
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render(" kube-triage — Interactive Remediation Wizard "))
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render(" 🎯 Target Resource"))
	b.WriteString(bodyStyle.Render(fmt.Sprintf("   %s/%s (Namespace: %s)", m.result.ResourceType, m.result.ResourceName, m.result.Namespace)))
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render(" ⚠️  Severity Level"))
	severityLabel := getSeverityStyle(m.result.Severity).Render(" " + m.result.Severity + " ")
	b.WriteString(bodyStyle.Render("   " + severityLabel))
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render(" 🔍 Failure State"))
	b.WriteString(bodyStyle.Render("   " + m.result.Summary))
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render(" 🧠 Root Cause Analysis"))
	b.WriteString(bodyStyle.Render("   " + m.result.RootCause))
	b.WriteString("\n\n")

	b.WriteString(headerStyle.Render(" 🛠️ Remediation Steps"))
	b.WriteString("\n")

	for i, step := range m.result.Remediation {
		stepNum := fmt.Sprintf(" %d. ", i+1)
		prefix := "   "
		if i == m.currentStep {
			prefix = "▶ "
		}

		desc := step.Description
		if m.confirmed[i] {
			desc = confirmedStyle.Render("✅ " + desc)
		} else if i == m.currentStep {
			desc = currentStyle.Render("⏳ " + desc)
		} else {
			desc = pendingStyle.Render("⏸  " + desc)
		}

		b.WriteString(prefix + stepNum + desc)
		b.WriteString("\n")

		if step.Command != "" {
			cmdStyle := commandStyle
			if m.confirmed[i] {
				cmdStyle = commandConfirmedStyle
			}
			b.WriteString(cmdStyle.Render("      $ " + step.Command))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render(" [y] Confirm  [n] Skip  [a] Confirm All  [s] Abort  [←/→] Navigate  [Enter] Finish"))
	b.WriteString("\n")

	return b.String()
}

func RunWizard(result *analyzer.AnalysisResult) error {
	p := tea.NewProgram(NewWizard(result))
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	m := finalModel.(*WizardModel)
	if m.aborted {
		return fmt.Errorf("remediation aborted by user")
	}

	for i, step := range result.Remediation {
		if m.confirmed[i] && step.Command != "" {
			fmt.Printf("Executing: %s\n", step.Command)
		}
	}

	return nil
}

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFFF")).
		Background(lipgloss.Color("#1a1a2e")).
		Padding(0, 1).
		MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		MarginTop(1).
		MarginBottom(0)

	bodyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		MarginLeft(2)

	commandStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Background(lipgloss.Color("#0a0a1a")).
		Padding(0, 1).
		MarginLeft(6)

	commandConfirmedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Background(lipgloss.Color("#0a1a0a")).
		Padding(0, 1).
		MarginLeft(6)

	currentStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFF00"))

	confirmedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	pendingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		MarginTop(1)

	successStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		Padding(1, 2).
		MarginTop(1)

	abortStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF0000")).
		Padding(1, 2).
		MarginTop(1)
)

func getSeverityStyle(severity string) lipgloss.Style {
	s := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	switch severity {
	case "CRITICAL":
		return s.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#FF0000"))
	case "HIGH":
		return s.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#FF8800"))
	case "MEDIUM":
		return s.Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#FFFF00"))
	case "LOW":
		return s.Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#00FF00"))
	}
	return s
}