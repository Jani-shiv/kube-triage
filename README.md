# kube-triage 🚀

<div align="center">

![kube-triage](https://img.shields.io/badge/kube--triage-v1.0.0-00FFFF?style=for-the-badge&logo=kubernetes&logoColor=white)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Build](https://img.shields.io/badge/Build-Passing-brightgreen?style=for-the-badge)

**AI-Powered Kubernetes Diagnostic Engine**  
*From `CrashLoopBackOff` to root cause in seconds — not hours.*

[![GitHub Stars](https://img.shields.io/github/stars/Jani-shiv/kube-triage?style=social)](https://github.com/Jani-shiv/kube-triage/stargazers)
[![GitHub Issues](https://img.shields.io/github/issues/Jani-shiv/kube-triage?style=social)](https://github.com/Jani-shiv/kube-triage/issues)
[![GitHub Forks](https://img.shields.io/github/forks/Jani-shiv/kube-triage?style=social)](https://github.com/Jani-shiv/kube-triage/network/members)

</div>

---

## 🎯 Why kube-triage?

| **Traditional Debugging** | **kube-triage** |
|---------------------------|-----------------|
| `kubectl describe pod` → 200 lines of YAML | **One-command diagnosis** |
| `kubectl logs` → scroll through noise | **Smart log correlation** |
| Check events, nodes, PVCs manually | **Auto-correlates all resources** |
| Guess root cause | **Deterministic rules + AI verification** |
| Craft `kubectl patch` by hand | **Copy-pasteable remediation** |

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🔍 **Multi-Resource Collection**
- Pods, Deployments, StatefulSets, DaemonSets, Jobs, CronJobs
- Services, Endpoints, Ingress, NetworkPolicies
- PVCs, PVs, StorageClasses
- Nodes (conditions, pressure, taints)

</td>
<td width="50%">

### 🧠 **Hybrid Analysis Engine**
- **Fast Path**: Deterministic rules (exit codes, events) → instant results
- **Deep Path**: LLM reasoning for complex multi-system failures
- Supports: NVIDIA Nemotron, OpenAI, Anthropic, Ollama (local)

</td>
</tr>
<tr>
<td>

### 🎨 **Beautiful Terminal UI**
- Dark-mode neo-brutalist theme
- Color-coded severity (🔴 CRITICAL → 🟢 LOW)
- Structured tables, copy-pasteable commands
- JSON output for automation

</td>
<td>

### 🛠️ **Production-Safe Remediation**
- Dry-run mode (`--dry-run`)
- Generates exact `kubectl patch`/`edit` commands
- Interactive fix wizard (coming soon)
- Prevention best practices included

</td>
</tr>
</table>

---

## 🚀 Quick Start

### Installation

```bash
# Homebrew (coming soon)
brew install jani-shiv/tap/kube-triage

# Go install
go install github.com/Jani-shiv/kube-triage/cmd@latest

# Download binary
# See Releases page
```

### Prerequisites

- **Kubernetes cluster** (Docker Desktop, kind, minikube, EKS/GKE/AKS)
- **kubectl** configured with cluster access
- **Go 1.22+** (for building from source)

> **Docker Desktop Users**: Settings → Kubernetes → ☑️ Enable Kubernetes → Apply & Restart

### Basic Usage

```bash
# Analyze a failing pod
kube-triage analyze pod my-app-7d9f8 -n production

# Analyze a deployment (auto-finds pods)
kube-triage analyze deployment my-app -n production

# Include previous container logs (for crashed containers)
kube-triage analyze pod my-app-7d9f8 -n production --previous

# Dry-run: show commands without executing
kube-triage analyze pod my-app-7d9f8 -n production --dry-run

# JSON output for CI/CD pipelines
kube-triage analyze pod my-app-7d9f8 -n production --output json
```

---

## 🎬 Demo Output

<div align="center">

```
╔═════════════════════════════════════════════════════════════════╗
║  kube-triage v1.0.0 — AI Kubernetes Diagnostic Engine           ║
╚═════════════════════════════════════════════════════════════════╝

 🎯 Target Resource : Pod/payment-api-5f78b9-8xk2l (Namespace: payments)
 🔍 Failure State   : CrashLoopBackOff (Restart Count: 47)
 ⚠️  Severity Level : ████████████ HIGH ████████████

 🧠 Root Cause Analysis:
   The container is exiting with exit code 137 (OOMKilled). The
   application exceeded its hard memory limit of 512Mi. Kubernetes
   sent SIGKILL to terminate the process. Current limit is too low
   for the workload's memory profile.

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

</div>

---

## 📋 Supported Failure Patterns (Phase 1)

| Rule ID | Pattern | Exit Code | Event Reason | Auto-Detected |
|---------|---------|-----------|--------------|---------------|
| **KT-001** | OOMKilled | 137 | `OOMKilled` | ✅ |
| **KT-002** | CrashLoopBackOff | - | `CrashLoopBackOff` | ✅ |
| **KT-003** | ImagePullBackOff | - | `ImagePullBackOff` / `ErrImagePull` | ✅ |
| **KT-004** | FailedMount | - | `FailedMount` | ✅ |
| **KT-005** | FailedScheduling | - | `FailedScheduling` | ✅ |
| **KT-006** | CreateContainerConfigError | - | `CreateContainerConfigError` | 🔄 |
| **KT-007** | RunContainerError | - | `RunContainerError` | 🔄 |
| **KT-008** | PVC Pending | - | `PersistentVolumeClaimPending` | 🔄 |
| **KT-009** | Node Pressure | - | `NodeMemoryPressure` / `DiskPressure` | 🔄 |
| **KT-010** | RBAC Denied | - | `Forbidden` | 🔄 |

> ✅ = Implemented | 🔄 = In Progress (Phase 2)

---

## ⚙️ Configuration

### Config File (`~/.kube-triage.yaml`)

```yaml
kubeconfig: ~/.kube/config
context: docker-desktop
namespace: default
llm-provider: auto        # auto, nvidia, openai, anthropic, ollama, none
llm-model: nemotron-3-ultra
llm-endpoint: ""          # Custom endpoint (for Ollama/local)
dry-run: false
previous: false
tail: 100
output: terminal          # terminal, json, yaml
```

### Environment Variables

```bash
export KUBE_TRIAGE_KUBECONFIG=~/.kube/config
export KUBE_TRIAGE_NAMESPACE=production
export KUBE_TRIAGE_LLM_PROVIDER=nvidia
export KUBE_TRIAGE_LLM_MODEL=nemotron-3-ultra
export NVIDIA_API_KEY=nvapi-xxxxxxxxxxxx     # For NVIDIA Nemotron
export OPENAI_API_KEY=sk-xxxxxxxxxxxx        # For OpenAI
export ANTHROPIC_API_KEY=sk-ant-xxxxxxxxxxxx # For Anthropic
export OLLAMA_HOST=http://localhost:11434    # For local Ollama
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        kube-triage CLI                          │
├─────────────────────────────────────────────────────────────────┤
│  cmd/                    │  internal/                            │
│  ├── main.go             │  ├── collector/   ← K8s API clients  │
│  └── (Cobra CLI)         │  │   ├── pod.go    ← Pods, logs      │
│                          │  │   ├── events.go ← Warning events   │
│                          │  │   ├── workload.go← Deploy/RS/STS   │
│                          │  │   ├── node.go   ← Node conditions  │
│                          │  │   ├── storage.go← PVC/PV           │
│                          │  │   └── network.go← Svc/EP/Ingress   │
│                          │  ├── analyzer/                     │
│                          │  │   ├── engine.go  ← Hybrid engine  │
│                          │  │   ├── rules/     ← Deterministic   │
│                          │  │   └── ai/        ← LLM clients    │
│                          │  ├── renderer/                     │
│                          │  │   ├── printer.go ← Lipgloss TUI  │
│                          │  │   └── tables.go  ← Structured    │
│                          │  └── fixer/        ← Remediation    │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🛣️ Roadmap

### ✅ Phase 1: Foundation (Complete)
- [x] Go module + Cobra CLI
- [x] 6 Collectors (Pod, Events, Workload, Node, Storage, Network)
- [x] Deterministic analyzer engine (5 rules)
- [x] Lipgloss terminal renderer
- [x] JSON output

### 🔄 Phase 2: Heuristic Rules Engine (In Progress)
- [ ] `heuristic-rules.yaml` declarative schema
- [ ] Rule loader & hot-reload
- [ ] 10+ built-in rules (KT-001 through KT-010)
- [ ] Rule testing framework

### 📋 Phase 3: AI Layer
- [ ] NVIDIA Nemotron integration (primary)
- [ ] Ollama local model support
- [ ] Context payload optimization
- [ ] Prompt templates for complex failures

### 📋 Phase 4: TUI & Release
- [ ] Interactive fix wizard (Bubble Tea)
- [ ] Dry-run with confirmation prompts
- [ ] goreleaser + Homebrew tap
- [ ] Documentation site

---

## 🤝 Contributing

```bash
# Fork → Clone → Branch
git checkout -b feat/amazing-feature

# Develop
go test ./...
go build ./...

# Commit (conventional commits)
git commit -m "feat: add amazing feature"

# Push → PR
git push origin feat/amazing-feature
```

### Development Setup

```bash
git clone https://github.com/Jani-shiv/kube-triage.git
cd kube-triage
go mod download
go build -o kube-triage ./cmd
./kube-triage --help
```

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- **Kubernetes SIG-CLI** for client-go
- **Charmbracelet** for Lipgloss, Bubbles, Bubble Tea
- **NVIDIA** for Nemotron 3 Ultra (free API tier)
- **Cobra & Viper** for CLI excellence

---

<div align="center">

**Made with ❤️ for the Kubernetes community**

[⬆ Back to Top](#kube-triage-)

</div>