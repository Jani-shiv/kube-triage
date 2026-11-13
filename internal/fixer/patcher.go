package fixer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"kube-triage/internal/analyzer"
)

type Patcher struct {
	dryRun      bool
	autoApprove bool
	kubectlCmd  string
	tempFiles   []string
	mu          sync.Mutex
}

func NewPatcher(dryRun, autoApprove bool) *Patcher {
	return &Patcher{
		dryRun:      dryRun,
		autoApprove: autoApprove,
		kubectlCmd:  "kubectl",
		tempFiles:   []string{},
	}
}

type PatchResult struct {
	Command  string
	Success  bool
	Output   string
	Error    error
}

func (p *Patcher) ApplyRemediation(ctx context.Context, result *analyzer.AnalysisResult) ([]PatchResult, error) {
	if len(result.Remediation) == 0 {
		return []PatchResult{}, nil
	}

	var results []PatchResult

	for i, step := range result.Remediation {
		if step.Command == "" || step.Type != "command" {
			results = append(results, PatchResult{
				Command: step.Command,
				Success: true,
				Output:  "No command to execute",
			})
			continue
		}

		if p.dryRun {
			fmt.Printf("[DRY-RUN] Would execute: %s\n", step.Command)
			results = append(results, PatchResult{
				Command: step.Command,
				Success: true,
				Output:  "[DRY-RUN] Command displayed only",
			})
			continue
		}

		if !p.autoApprove {
			fmt.Printf("\nExecute step %d/%d?\n", i+1, len(result.Remediation))
			fmt.Printf("  %s\n", step.Description)
			fmt.Printf("  $ %s\n", step.Command)
			fmt.Print("  [y/N/a] (yes/No/All): ")

			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))

			if response == "a" || response == "all" {
				p.autoApprove = true
			} else if response != "y" && response != "yes" {
				fmt.Println("  Skipped.")
				results = append(results, PatchResult{
					Command: step.Command,
					Success: false,
					Output:  "Skipped by user",
				})
				continue
			}
		}

		fmt.Printf("Executing: %s\n", step.Command)
		output, err := p.ExecuteCommand(ctx, step.Command)

		results = append(results, PatchResult{
			Command: step.Command,
			Success: err == nil,
			Output:  output,
			Error:   err,
		})

		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			if output != "" {
				fmt.Printf("  Output: %s\n", output)
			}
		} else {
			fmt.Printf("  ✅ Success\n")
			if output != "" {
				fmt.Printf("  Output: %s\n", output)
			}
		}
	}

	return results, nil
}

func (p *Patcher) ExecuteCommand(ctx context.Context, command string) (string, error) {
	if p.dryRun {
		return fmt.Sprintf("[DRY-RUN] %s", command), nil
	}

	command = p.fixCommandForPlatform(command)
	cmd := p.createCommand(ctx, command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (p *Patcher) createCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd.exe", "/c", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}

func (p *Patcher) fixCommandForPlatform(command string) string {
	if runtime.GOOS != "windows" {
		return command
	}

	containsKubectl := strings.Contains(command, "kubectl patch")
	containsType := strings.Contains(command, "--type='json'")
	
	if containsKubectl && containsType {
		return p.fixKubectlPatchForWindows(command)
	}

	return command
}

func (p *Patcher) fixKubectlPatchForWindows(command string) string {
	if !strings.Contains(command, "kubectl patch") || !strings.Contains(command, "--type='json'") {
		return command
	}

	parts := strings.Split(command, "--type='json' -p='")
	if len(parts) < 2 {
		return command
	}

	beforePatch := parts[0]
	patchJSON := parts[1]
	patchJSON = strings.TrimSuffix(patchJSON, "'")

	beforeParts := strings.Fields(beforePatch)
	if len(beforeParts) < 4 {
		return command
	}

	resource := beforeParts[2]
	name := beforeParts[3]

	tmpFile, err := os.CreateTemp("", "kube-triage-patch-*.json")
	if err != nil {
		return command
	}

	if _, err := tmpFile.WriteString(patchJSON); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return command
	}
	tmpFile.Close()

	p.mu.Lock()
	p.tempFiles = append(p.tempFiles, tmpFile.Name())
	p.mu.Unlock()

	fixed := fmt.Sprintf(`kubectl patch %s %s --type=json --patch-file=%s`, resource, name, tmpFile.Name())
	return fixed
}

func (p *Patcher) GeneratePatchYAML(result *analyzer.AnalysisResult) (string, error) {
	var patches []string

	for _, step := range result.Remediation {
		if step.Command != "" && strings.Contains(step.Command, "kubectl patch") {
			patches = append(patches, step.Command)
		}
	}

	if len(patches) == 0 {
		return "", nil
	}

	yaml := "# kube-triage generated patches\n"
	yaml += "# Apply with: kubectl apply -f <this-file>.yaml\n\n"

	for _, patch := range patches {
		yaml += "# " + patch + "\n"
	}

	return yaml, nil
}

func (p *Patcher) PrintSummary(results []PatchResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("REMEDIATION SUMMARY")
	fmt.Println(strings.Repeat("=", 60))

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
			fmt.Printf("✅ %s\n", r.Command)
		} else {
			fmt.Printf("❌ %s\n", r.Command)
			if r.Error != nil {
				fmt.Printf("   Error: %v\n", r.Error)
			}
			if r.Output != "" {
				fmt.Printf("   Output: %s\n", r.Output)
			}
		}
	}

	fmt.Printf("\nTotal: %d | Success: %d | Failed: %d\n", len(results), successCount, len(results)-successCount)
}

func (p *Patcher) Cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, f := range p.tempFiles {
		os.Remove(f)
	}
	p.tempFiles = []string{}
}