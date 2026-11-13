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
- **Fast Path**: 15 deterministic rules (exit codes, events) → instant results
- **Deep Path**: LLM reasoning for complex multi-system failures
- **Providers**: NVIDIA Nemotron (primary), Ollama (local), OpenAI, Anthropic
- **Auto-detection**: Priority chain with graceful fallback

</td>
</tr>
<tr>
<td>

### 🎨 **Beautiful Terminal UI**
- Dark-mode neo-brutalist theme (Lipgloss)
- Color-coded severity (🔴 CRITICAL → 🟢 LOW)
- Structured tables, copy-pasteable commands
- JSON output for CI/CD automation

</td>
<td>

### 🛠️ **Production-Safe Remediation**
- Dry-run mode (`--dry-run`)
- Generates exact `kubectl patch`/`edit` commands
- Deployment name resolution (Pod → RS → Deploy)
- Prevention best practices included

</td>
</tr>
</table>

---

## 🚀 Quick Start

### Installation

```bash
# Go install (latest)
go install github.com/Jani-shiv/kube-triage/cmd@latest

# Build from source
git clone https://github.com/Jani-shiv/kube-triage.git
cd kube-triage
go build -o kube-triage ./cmd
./kube-triage --help

# Download binary
# See Releases page (coming soon)
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

# Analyze a deployment (auto-finds pods, resolves deployment name)
kube-triage analyze deployment my-app -n production

# Include previous container logs (for crashed containers)
kube-triage analyze pod my-app-7d9f8 -n production --previous

# Dry-run: show commands without executing
kube-triage analyze pod my-app-7d9f8 -n production --dry-run

# JSON output for CI/CD pipelines
kube-triage analyze pod my-app-7d9f8 -n production --output json

# Use specific AI provider
kube-triage analyze pod my-app-7d9f8 -n production --llm-provider nvidia
kube-triage analyze pod my-app-7d9f8 -n production --llm-provider ollama
kube-triage analyze pod my-app-7d9f8 -n production --llm-provider none
```

---

## 🎬 Demo Output

<div align="center">

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

## 📋 Supported Failure Patterns (Phases 1-2 Complete)

| Rule ID | Pattern | Category | Severity | Trigger |
|---------|---------|----------|----------|---------|
| **KT-001** | OOMKilled Container Termination | resource | 🔴 CRITICAL | exitCode=137, reason=OOMKilled |
| **KT-002** | CrashLoopBackOff | lifecycle | 🟠 HIGH | waiting.reason=CrashLoopBackOff |
| **KT-003** | ImagePullBackOff / ErrImagePull | image | 🟠 HIGH | waiting.reason=ImagePullBackOff\|ErrImagePull |
| **KT-004** | FailedMount Volume | storage | 🟠 HIGH | event.reason=FailedMount |
| **KT-005** | FailedScheduling | scheduling | 🟠 HIGH | event.reason=FailedScheduling |
| **KT-006** | CreateContainerConfigError | configuration | 🟠 HIGH | waiting.reason=CreateContainerConfigError |
| **KT-007** | RunContainerError | runtime | 🟠 HIGH | waiting.reason=RunContainerError |
| **KT-008** | PersistentVolumeClaim Pending | storage | 🟠 HIGH | resource.kind=PVC, status=Pending |
| **KT-009** | Node Memory/Disk/PID Pressure | node | 🟠 HIGH | nodeCondition (Memory/Disk/PID) |
| **KT-010** | RBAC Permission Denied | security | 🟠 HIGH | event.reason=Forbidden |
| **KT-011** | Liveness/Readiness Probe Failure | probes | 🟡 MEDIUM | event.reason=Liveness/ReadinessProbeFailed |
| **KT-012** | OOMKilled - Init Container | resource | 🔴 CRITICAL | initContainer exitCode=137 |
| **KT-013** | Container Exited with Code 1 | application | 🟡 MEDIUM | terminated.exitCode=1 |
| **KT-014** | Container Exited with Code 139 (Segfault) | application | 🟠 HIGH | terminated.exitCode=139 |
| **KT-015** | Container Exited with Code 143 (SIGTERM) | lifecycle | 🟢 LOW | terminated.exitCode=143 |

> ✅ All 15 rules implemented and tested

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
# Core
export KUBE_TRIAGE_KUBECONFIG=~/.kube/config
export KUBE_TRIAGE_NAMESPACE=production
export KUBE_TRIAGE_LLM_PROVIDER=auto
export KUBE_TRIAGE_LLM_MODEL=nemotron-3-ultra

# AI Providers (priority: NVIDIA → Ollama → OpenAI → Anthropic)
export NVIDIA_API_KEY=nvapi-xxxxxxxxxxxx     # Primary (free tier at build.nvidia.com)
export NVIDIA_MODEL=nvidia/nemotron-3-ultra
export OLLAMA_HOST=http://localhost:11434    # Local fallback
export OLLAMA_MODEL=nemotron3-ultra
export OPENAI_API_KEY=sk-xxxxxxxxxxxx        # OpenAI fallback
export OPENAI_MODEL=gpt-4o
export ANTHROPIC_API_KEY=sk-ant-xxxxxxxxxxxx # Anthropic fallback
export ANTHROPIC_MODEL=claude-3-5-sonnet-20241022
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
│                          │  │   ├── rules/     ← 15 rules       │
│                          │  │   └── ai/        ← 4 LLM clients  │
│                          │  ├── renderer/                     │
│                          │  │   ├── printer.go ← Lipgloss TUI  │
│                          │  │   └── tables.go  ← Structured    │
│                          │  └── fixer/        ← Remediation    │
└─────────────────────────────────────────────────────────────────┘
```

### Hybrid Analysis Flow

```
User: kube-triage analyze pod failing-app -n prod
         │
         ▼
┌──────────────────────────────────────────┐
│ Collector: Pod + Events + Workload + Node │
└──────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────┐
│ Deterministic Rules (KT-001 to KT-015)   │
│   ├─ Match? → Return result (instant)    │
│   └─ No match? → Build AI Payload        │
└──────────────────────────────────────────┘
         │
         ▼ (if no deterministic match)
┌──────────────────────────────────────────┐
│ AI Provider Chain (auto-detected)        │
│   1. NVIDIA Nemotron (NVIDIA_API_KEY)    │
│   2. Ollama Local (OLLAMA_HOST)          │
│   3. OpenAI (OPENAI_API_KEY)             │
│   4. Anthropic (ANTHROPIC_API_KEY)       │
└──────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────┐
│ Render: Terminal UI / JSON / YAML        │
└──────────────────────────────────────────┘
```

---

## 🛣️ Roadmap

### ✅ Phase 1: Foundation (Complete)
- [x] Go module + Cobra CLI with full flag/config support
- [x] 6 Collectors (Pod, Events, Workload, Node, Storage, Network)
- [x] Deterministic analyzer engine structure
- [x] Lipgloss terminal renderer with severity colors
- [x] JSON output format

### ✅ Phase 2: Heuristic Rules Engine (Complete)
- [x] `heuristic-rules.yaml` declarative schema (v1)
- [x] Rule loader with validation & hot-reload support
- [x] 15 built-in rules (KT-001 through KT-015)
- [x] Template rendering for dynamic remediation commands
- [x] Deployment name resolution (Pod → ReplicaSet → Deployment)

### ✅ Phase 3: AI Layer (Complete)
- [x] NVIDIA Nemotron 3 Ultra client (primary, free tier)
- [x] Ollama local client with health check & model listing
- [x] OpenAI GPT-4o compatible client
- [x] Anthropic Claude 3.5 Sonnet client
- [x] Provider auto-detection with priority fallback chain
- [x] Context payload builder (cluster state → structured JSON)
- [x] Prompt templates for complex multi-system failures
- [x] Hybrid engine: deterministic first → LLM fallback

### 🔄 Phase 4: TUI & Release (In Progress)
- [ ] Interactive fix wizard (Bubble Tea)
- [ ] Dry-run with confirmation prompts
- [ ] goreleaser configuration
- [ ] Homebrew tap formula
- [ ] Installation scripts (Linux/macOS/Windows)
- [ ] Documentation site
- [ ] Unit/integration tests (kind-based)

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

### Testing Rules

```bash
# Create test scenarios
kubectl apply -f test/fixtures/oom-pod.yaml
kube-triage analyze pod oom-pod -n default

# Verify deterministic detection
kube-triage analyze pod oom-pod -n default --llm-provider none
```

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- **Kubernetes SIG-CLI** for client-go
- **Charmbracelet** for Lipgloss, Bubbles, Bubble Tea
- **NVIDIA** for Nemotron 3 Ultra (free API tier at build.nvidia.com)
- **Ollama** for local LLM runtime
- **Cobra & Viper** for CLI excellence

---

<div align="center">

**Made with ❤️ for the Kubernetes community**

[⬆ Back to Top](#kube-triage-)

</div>