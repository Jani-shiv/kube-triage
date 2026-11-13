# kube-triage v1.0.0 - Full Demo Recording Script

## Setup (Before Recording)

```bash
# 1. Enable Kubernetes in Docker Desktop
# Settings → Kubernetes → Enable Kubernetes → Apply & Restart

# 2. Verify cluster
kubectl get nodes
kubectl version --short

# 3. Install kube-triage (from release or build)
go install github.com/Jani-shiv/kube-triage/cmd@latest
# OR: curl -fsSL https://raw.githubusercontent.com/Jani-shiv/kube-triage/main/scripts/install.sh | bash

# 4. Verify install
kube-triage version
kube-triage --help
```

---

## Demo Scene 1: OOMKilled Detection (KT-001) - 60 seconds

### Setup
```bash
# Create OOM test pod
cat > /tmp/oom-pod.yaml << 'EOF'
apiVersion: v1
kind: Pod
metadata:
  name: demo-oom
  namespace: default
spec:
  restartPolicy: Never
  containers:
  - name: oom
    image: golang:1.22-alpine
    command: ["/bin/sh", "-c"]
    args:
      - |
        cat > /oom.go << 'EOGO'
        package main
        import ("runtime"; "time")
        func main() {
          var data []byte
          for { data = append(data, make([]byte, 1024*1024)...); runtime.GC(); time.Sleep(100*time.Millisecond) }
        }
        EOGO
        go run /oom.go
    resources:
      limits:
        memory: "10Mi"
      requests:
        memory: "5Mi"
EOF

kubectl apply -f /tmp/oom-pod.yaml
sleep 15
```

### Recording Commands
```bash
# Show pod status
kubectl get pod demo-oom

# Run kube-triage analysis
kube-triage analyze pod demo-oom -n default

# Show JSON output
kube-triage analyze pod demo-oom -n default --output json

# Clean up
kubectl delete -f /tmp/oom-pod.yaml
```

### Key Points to Highlight
- "Notice the CRITICAL severity - exit code 137 = OOMKilled"
- "Root cause explains Kubernetes sent SIGKILL"
- "Remediation includes exact kubectl patch commands"
- "Deployment name resolved automatically"

---

## Demo Scene 2: Deployment Analysis with Auto-Fix - 90 seconds

### Setup
```bash
# Create deployment that will OOM
cat > /tmp/oom-deploy.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-oom-deploy
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: demo-oom
  template:
    metadata:
      labels:
        app: demo-oom
    spec:
      containers:
      - name: oom
        image: golang:1.22-alpine
        command: ["/bin/sh", "-c"]
        args:
          - |
            cat > /oom.go << 'EOGO'
            package main
            import ("runtime"; "time")
            func main() {
              var data []byte
              for { data = append(data, make([]byte, 1024*1024)...); runtime.GC(); time.Sleep(100*time.Millisecond) }
            }
            EOGO
            go run /oom.go
        resources:
          limits:
            memory: "10Mi"
          requests:
            memory: "5Mi"
EOF

kubectl apply -f /tmp/oom-deploy.yaml
sleep 20
```

### Recording Commands
```bash
# Show deployment and pods
kubectl get deployment demo-oom-deploy
kubectl get pods -l app=demo-oom

# Analyze deployment (finds pods automatically)
kube-triage analyze deployment demo-oom-deploy -n default

# Dry-run mode
kube-triage analyze deployment demo-oom-deploy -n default --dry-run

# AUTO-APPROVE - Actually fixes it!
kube-triage analyze deployment demo-oom-deploy -n default --auto-approve

# Verify fix
kubectl get deployment demo-oom-deploy -o yaml | grep -A 5 "resources:"
```

### Key Points to Highlight
- "Analyzes deployment, finds pods, traces to deployment"
- "Dry-run shows commands without executing"
- "Auto-approve actually patches the deployment"
- "Memory limit increased from 10Mi → 1Gi, request 5Mi → 512Mi"

---

## Demo Scene 3: CrashLoopBackOff (KT-002) - 45 seconds

### Setup
```bash
kubectl run demo-crash --image=busybox --restart=Always --command -- sh -c "exit 1"
sleep 10
```

### Recording Commands
```bash
kubectl get pod demo-crash

kube-triage analyze pod demo-crash -n default --auto-approve

kubectl delete pod demo-crash --ignore-not-found=true
```

---

## Demo Scene 4: ImagePullBackOff (KT-003) - 30 seconds

### Recording Commands
```bash
kubectl run demo-bad-image --image=nonexistent/registry/invalid:latest --restart=Never
sleep 10

kube-triage analyze pod demo-bad-image -n default

kubectl delete pod demo-bad-image --ignore-not-found=true
```

---

## Demo Scene 4: Multiple Output Formats - 30 seconds

### Recording Commands
```bash
# Recreate OOM pod
kubectl apply -f /tmp/oom-pod.yaml
sleep 15

# Terminal (default)
kube-triage analyze pod demo-oom -n default

# JSON for CI/CD
kube-triage analyze pod demo-oom -n default --output json | jq .

# Cleanup
kubectl delete -f /tmp/oom-pod.yaml
```

---

## Demo Scene 5: Shell Completion - 20 seconds

### Recording Commands
```bash
# Bash
source <(kube-triage completion bash)
kube-triage analyze pod <TAB>

# PowerShell
kube-triage completion powershell | Out-String | Invoke-Expression
kube-triage analyze pod <TAB>
```

---

## Demo Scene 6: AI Provider Modes - 30 seconds

### Recording Commands
```bash
# Disable AI (deterministic only)
kube-triage analyze pod demo-oom -n default --llm-provider none

# Auto (NVIDIA → Ollama → OpenAI → Anthropic)
kube-triage analyze pod demo-oom -n default --llm-provider auto

# Show JSON with AI metadata
kube-triage analyze pod demo-oom -n default --output json | jq '.AIEnhanced, .AIProvider'
```

---

## Complete Demo Flow (5-6 minutes total)

| Scene | Duration | Focus |
|-------|----------|-------|
| Intro | 30s | Version, help, architecture |
| OOMKilled (Pod) | 60s | Core detection |
| Deployment Auto-Fix | 90s | End-to-end remediation |
| CrashLoopBackOff | 45s | Another rule |
| ImagePullBackOff | 30s | Image issues |
| Output Formats | 30s | JSON/Terminal |
| Shell Completion | 20s | Developer experience |
| AI Providers | 30s | Hybrid engine |
| Outro | 30s | Links, install, thanks |

**Total: ~6 minutes**

---

## Recording Tips

### asciinema (Recommended for terminal demos)
```bash
# Install
brew install asciinema  # macOS
sudo apt install asciinema  # Linux

# Record
asciinema rec kube-triage-demo.cast

# Play back
asciinema play kube-triage-demo.cast

# Upload to asciinema.org
asciinema upload kube-triage-demo.cast
```

### OBS Studio (Full screen recording)
- Resolution: 1920x1080 or 1280x720
- Frame rate: 30 fps
- Audio: Commentary mic
- Scenes: Terminal, Browser (GitHub), Slides

### Key moments to emphasize with zoom/highlight:
1. **Severity color coding** (CRITICAL=red, HIGH=orange, MEDIUM=yellow, LOW=green)
2. **Exact kubectl patch commands** - copy-pasteable
3. **Deployment name in command** - shows auto-resolution
4. **JSON output structure** - CI/CD integration
5. **Before/after resource limits** - proves fix worked

---

## Post-Recording Editing

1. **Trim** dead time (kubectl apply waits)
2. **Add captions** for key concepts
3. **Speed up** repetitive commands 2x
4. **Add overlay** with rule IDs (KT-001, KT-002, etc.)
5. **End screen** with GitHub repo link

---

## Demo Data Cleanup

```bash
# Clean up all test resources
kubectl delete pod demo-oom demo-crash demo-bad-image demo-exit1 --ignore-not-found=true
kubectl delete deployment demo-oom-deploy --ignore-not-found=true
rm /tmp/oom-pod.yaml /tmp/oom-deploy.yaml
```

---

## Talking Points Script

> "kube-triage bridges the gap between raw Kubernetes errors and actionable fixes. Instead of reading 200 lines of `kubectl describe`, you get a root cause and exact remediation commands in seconds."

> "The hybrid engine uses 15 deterministic rules for instant detection of common failures like OOMKilled, CrashLoopBackOff, ImagePullBackOff - no AI latency, no API costs."

> "For complex multi-system failures, it falls back to LLMs - NVIDIA Nemotron, Ollama local, OpenAI, Anthropic - with automatic provider detection and graceful fallback."

> "The fixer doesn't just show commands - it can execute them with `--auto-approve`, handles Windows kubectl JSON patching automatically, and supports dry-run for safety."

> "Install in seconds: `go install`, Homebrew, or our install scripts. Works with any Kubernetes cluster - Docker Desktop, kind, EKS, GKE, AKS."

---

## Upload Checklist

- [ ] Record all scenes
- [ ] Edit to ~5 minutes
- [ ] Add captions/annotations
- [ ] Upload to YouTube (unlisted or public)
- [ ] Add to GitHub repo README
- [ ] Share on Twitter/LinkedIn with #kubernetes #SRE #DevTools
- [ ] Submit to CNCF landscape / Awesome-Kubernetes lists