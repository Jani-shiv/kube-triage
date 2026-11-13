# Quick Start Guide

## Installation

### Linux/macOS (One-liner)
```bash
curl -fsSL https://raw.githubusercontent.com/Jani-shiv/kube-triage/main/scripts/install.sh | bash
```

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/Jani-shiv/kube-triage/main/scripts/install.ps1 | iex
```

### Homebrew (macOS/Linux)
```bash
brew tap Jani-shiv/tap
brew install kube-triage
```

### Manual Download
Download the latest release from [GitHub Releases](https://github.com/Jani-shiv/kube-triage/releases).

## Prerequisites
- Kubernetes cluster (Docker Desktop, kind, minikube, EKS/GKE/AKS)
- `kubectl` configured with cluster access

> **Docker Desktop Users**: Settings → Kubernetes → ☑️ Enable Kubernetes → Apply & Restart

## Quick Start

### 1. Analyze a failing pod
```bash
kube-triage analyze pod my-app-7d9f8 -n production
```

### 2. Analyze a deployment (auto-finds pods)
```bash
kube-triage analyze deployment my-app -n production
```

### 3. Dry-run (preview commands)
```bash
kube-triage analyze pod my-app-7d9f8 -n production --dry-run
```

### 4. JSON output for CI/CD
```bash
kube-triage analyze pod my-app-7d9f8 -n production --output json
```

### 5. Auto-approve remediation
```bash
kube-triage analyze deployment my-app -n production --auto-approve
```

## AI Providers (Optional)

Set up AI for complex failure analysis:

```bash
# NVIDIA Nemotron 3 Ultra (Primary - Free tier)
export NVIDIA_API_KEY=nvapi-xxxxxxxxxxxx

# Ollama (Local)
export OLLAMA_HOST=http://localhost:11434

# OpenAI
export OPENAI_API_KEY=sk-xxxxxxxxxxxx

# Anthropic
export ANTHROPIC_API_KEY=sk-ant-xxxxxxxxxxxx
```

## Config File (~/.kube-triage.yaml)
```yaml
kubeconfig: ~/.kube/config
context: docker-desktop
namespace: default
llm-provider: auto
dry-run: false
previous: false
tail: 100
output: terminal
```

## Example Output

```
╔══════════════════════════════════════════════════════════════════╗
║  kube-triage v1.0.0 — AI Kubernetes Diagnostic Engine           ║
╚══════════════════════════════════════════════════════════════════╝

 🎯 Target Resource : Pod/payment-api-5f78b9-8xk2l (Namespace: payments)
 🔍 Failure State   : CrashLoopBackOff (Restart Count: 47)
 ⚠️  Severity Level : ████████████ HIGH ████████████

 🧠 Root Cause Analysis:
   The container is exiting with exit code 137 (OOMKilled). The
   application exceeded its hard memory limit of 512Mi. Kubernetes
   sent SIGKILL to terminate the process.

 🛠️ Recommended Remediation:
   1. Increase memory limit to 1Gi
      $ kubectl patch deployment payment-api --type='json' \
        -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/resources/limits/memory", "value": "1Gi"}]'

   2. Add memory request for QoS guarantees
      $ kubectl patch deployment payment-api --type='json' \
        -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/resources/requests/memory", "value": "512Mi"}]'

 🛡️ Prevention & Best Practices:
   • Set memory requests ≈ limits for Guaranteed QoS class
   • Configure Prometheus alerts: container_memory_usage_bytes > 0.8 * limit
   • Profile application with pprof to identify memory leaks
   • Implement graceful degradation under memory pressure
```

## Supported Failure Patterns

| Rule | Pattern | Severity |
|------|---------|----------|
| KT-001 | OOMKilled (exit 137) | 🔴 CRITICAL |
| KT-002 | CrashLoopBackOff | 🟠 HIGH |
| KT-003 | ImagePullBackOff / ErrImagePull | 🟠 HIGH |
| KT-004 | FailedMount Volume | 🟠 HIGH |
| KT-005 | FailedScheduling | 🟠 HIGH |
| KT-006 | CreateContainerConfigError | 🟠 HIGH |
| KT-007 | RunContainerError | 🟠 HIGH |
| KT-008 | PVC Pending | 🟠 HIGH |
| KT-009 | Node Pressure (Memory/Disk/PID) | 🟠 HIGH |
| KT-010 | RBAC Forbidden | 🟠 HIGH |
| KT-011 | Probe Failure | 🟡 MEDIUM |
| KT-012 | Init Container OOMKilled | 🔴 CRITICAL |
| KT-013 | Exit Code 1 | 🟡 MEDIUM |
| KT-014 | Segfault (Exit 139) | 🟠 HIGH |
| KT-015 | SIGTERM (Exit 143) | 🟢 LOW |

## Shell Completion

```bash
# Bash
source <(kube-triage completion bash)

# Zsh
source <(kube-triage completion zsh)

# PowerShell
kube-triage completion powershell | Out-String | Invoke-Expression
```

## Links
- [GitHub Repository](https://github.com/Jani-shiv/kube-triage)
- [Issues](https://github.com/Jani-shiv/kube-triage/issues)
- [Releases](https://github.com/Jani-shiv/kube-triage/releases)