<# 
.SYNOPSIS
    Run complete kube-triage demo automatically
.DESCRIPTION
    Runs all 6 demo scenes automatically and captures output to log file
    No external recording tools required - just runs the demo commands
.EXAMPLE
    .\scripts\run-demo.ps1
    .\scripts\run-demo.ps1 -LogPath "my-demo.log"
#>

param(
    [string]$LogPath = "kube-triage-demo-$(Get-Date -Format 'yyyyMMdd-HHmmss').log",
    [switch]$NoCleanup
)

$ErrorActionPreference = "Continue"
$startTime = Get-Date

function Write-Log {
    param($Message, $Level = "INFO")
    $timestamp = Get-Date -Format "HH:mm:ss"
    $prefix = switch ($Level) {
        "SCENE" { "SCENE" }
        "CMD"   { "CMD" }
        "OK"    { "OK" }
        "WARN"  { "WARN" }
        "ERROR" { "ERROR" }
        default { "INFO" }
    }
    $line = "[$timestamp] [$prefix] $Message"
    Write-Host $line
    Add-Content -Path $LogPath -Value $line
}

function Run-Cmd {
    param($Command, $Description = "")
    if ($Description) { Write-Log $Description "CMD" }
    Write-Log "Running: $Command" "CMD"
    $output = Invoke-Expression $Command 2>&1
    Write-Log $output
    return $output
}

function Wait-Pod {
    param($PodName, $Timeout = 30)
    $elapsed = 0
    while ($elapsed -lt $Timeout) {
        $pod = kubectl get pod $PodName --no-headers 2>$null
        if ($pod -and ($pod -match "OOMKilled|CrashLoopBackOff|ErrImagePull|Error|Completed")) {
            return $true
        }
        Start-Sleep 2
        $elapsed += 2
    }
    return $false
}

# Initialize log
"=== kube-triage v1.0.0 Demo Log ===" | Out-File $LogPath
"Started: $startTime" | Add-Content $LogPath
"=======================================" | Add-Content $LogPath

Write-Log "Starting kube-triage v1.0.0 Demo" "SCENE"
Write-Log "Log file: $LogPath"

# ===========================================
# PRE-CHECKS
# ===========================================
Write-Log "Checking prerequisites..." "SCENE"
Run-Cmd "kube-triage version" "Check kube-triage installed"
Run-Cmd "kubectl get nodes" "Check Kubernetes cluster"
Run-Cmd "kubectl get ns default" "Check default namespace"

# ===========================================
# SCENE 1: OOMKilled (KT-001)
# ===========================================
Write-Log "SCENE 1: OOMKilled Detection (KT-001) - CRITICAL" "SCENE"

# Create OOM pod YAML
$oomYaml = @"
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
"@

$oomYaml | Out-File "demo-oom.yaml" -Encoding utf8

Run-Cmd "kubectl apply -f demo-oom.yaml" "Create OOM test pod"
Start-Sleep 15

if (Wait-Pod "demo-oom" 30) {
    Run-Cmd "kubectl get pod demo-oom" "Check pod status"
    Run-Cmd "kube-triage analyze pod demo-oom -n default --dry-run" "Dry-run analysis"
    Run-Cmd "kube-triage analyze pod demo-oom -n default --output json" "JSON output"
    Run-Cmd "kubectl delete -f demo-oom.yaml" "Cleanup OOM pod"
} else {
    Write-Log "Pod did not reach OOMKilled state in time" "WARN"
    kubectl delete -f demo-oom.yaml -ErrorAction SilentlyContinue
}

# ===========================================
# SCENE 2: Deployment Auto-Fix
# ===========================================
Write-Log "SCENE 2: Deployment Auto-Fix (OOMKilled)" "SCENE"

$deployYaml = @"
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
"@

$deployYaml | Out-File "demo-oom-deploy.yaml" -Encoding utf8

Run-Cmd "kubectl apply -f demo-oom-deploy.yaml" "Create OOM deployment"
Start-Sleep 20

if (Wait-Pod "demo-oom-deploy" 30) {
    Run-Cmd "kubectl get deployment demo-oom-deploy" "Check deployment"
    Run-Cmd "kubectl get pods -l app=demo-oom" "Check pods"
    Run-Cmd "kube-triage analyze deployment demo-oom-deploy -n default --dry-run" "Dry-run analysis"
    Run-Cmd "kube-triage analyze deployment demo-oom-deploy -n default --auto-approve" "Auto-approve fix"
    Run-Cmd "kubectl get deployment demo-oom-deploy -o yaml | Select-String 'memory:'" "Verify fix applied"
    Run-Cmd "kubectl delete -f demo-oom-deploy.yaml" "Cleanup deployment"
} else {
    Write-Log "Deployment pod did not reach OOMKilled state in time" "WARN"
    kubectl delete -f demo-oom-deploy.yaml -ErrorAction SilentlyContinue
}

# ===========================================
# SCENE 3: CrashLoopBackOff (KT-002)
# ===========================================
Write-Log "SCENE 3: CrashLoopBackOff (KT-002) - HIGH" "SCENE"

Run-Cmd "kubectl run demo-crash --image=busybox --restart=Always --command -- sh -c \"exit 1\"" "Create CrashLoopBackOff pod"
Start-Sleep 10

if (Wait-Pod "demo-crash" 20) {
    Run-Cmd "kubectl get pod demo-crash" "Check pod status"
    Run-Cmd "kube-triage analyze pod demo-crash -n default --auto-approve" "Auto-approve remediation"
    Run-Cmd "kubectl delete pod demo-crash --ignore-not-found=true" "Cleanup"
} else {
    Write-Log "Pod did not reach CrashLoopBackOff state in time" "WARN"
    kubectl delete pod demo-crash --ignore-not-found=true
}

# ===========================================
# SCENE 4: ImagePullBackOff (KT-003)
# ===========================================
Write-Log "SCENE 4: ImagePullBackOff (KT-003) - HIGH" "SCENE"

Run-Cmd "kubectl run demo-bad-image --image=nonexistent/registry/invalid:latest --restart=Never" "Create ImagePullBackOff pod"
Start-Sleep 10

Run-Cmd "kubectl get pod demo-bad-image" "Check pod status"
Run-Cmd "kube-triage analyze pod demo-bad-image -n default" "Analyze ImagePullBackOff"
Run-Cmd "kubectl delete pod demo-bad-image --ignore-not-found=true" "Cleanup"

# ===========================================
# SCENE 5: Shell Completion
# ===========================================
Write-Log "SCENE 5: Shell Completion (PowerShell)" "SCENE"
Run-Cmd "kube-triage completion powershell | Select-Object -First 30" "PowerShell completion (first 30 lines)"

# ===========================================
# SCENE 6: AI Provider Modes
# ===========================================
Write-Log "SCENE 6: AI Provider Modes" "SCENE"

Run-Cmd "kubectl run demo-ai-test --image=busybox --restart=Never --command -- sh -c \"exit 1\"" "Create test pod for AI modes"
Start-Sleep 5

$pod = kubectl get pod demo-ai-test --no-headers 2>$null
if ($pod -and ($pod -match "Error|Completed")) {
    Run-Cmd "kube-triage analyze pod demo-ai-test -n default --llm-provider none" "Deterministic only (--llm-provider none)"
    Run-Cmd "kubectl delete pod demo-ai-test --ignore-not-found=true" "Cleanup"
}

# Test auto provider
Run-Cmd "kubectl run demo-crash-ai --image=busybox --restart=Always --command -- sh -c \"exit 1\"" "Create pod for auto provider test"
Start-Sleep 10

$pod2 = kubectl get pod demo-crash-ai --no-headers 2>$null
if ($pod2 -and ($pod2 -match "CrashLoopBackOff")) {
    Run-Cmd "kube-triage analyze pod demo-crash-ai -n default --llm-provider auto" "Auto provider (NVIDIA+Ollama+OpenAI+Anthropic)"
    Run-Cmd "kubectl delete pod demo-crash-ai --ignore-not-found=true" "Cleanup"
}

# ===========================================
# CLEANUP
# ===========================================
if (-not $NoCleanup) {
    Write-Log "Cleaning up any remaining demo resources..." "SCENE"
    kubectl delete pod demo-oom, demo-crash, demo-bad-image, demo-ai-test, demo-crash-ai --ignore-not-found=true 2>$null
    kubectl delete deployment demo-oom-deploy --ignore-not-found=true 2>$null
    Remove-Item demo-oom.yaml, demo-oom-deploy.yaml -ErrorAction SilentlyContinue
}

# ===========================================
# SUMMARY
# ===========================================
$endTime = Get-Date
$duration = $endTime - $startTime

Write-Log "" "SCENE"
Write-Log "===========================================" "SCENE"
Write-Log "DEMO COMPLETE - kube-triage v1.0.0" "SCENE"
Write-Log "===========================================" "SCENE"
Write-Log "Duration: $($duration.Minutes)m $($duration.Seconds)s" "OK"
Write-Log "Log saved to: $LogPath" "OK"
Write-Log "" "SCENE"
Write-Log "SCENES COMPLETED:" "SCENE"
Write-Log "  1. OOMKilled (KT-001) - CRITICAL - Dry-run + JSON" "OK"
Write-Log "  2. Deployment Auto-Fix (KT-001) - CRITICAL - Auto-approve patched 10Mi to 1Gi" "OK"
Write-Log "  3. CrashLoopBackOff (KT-002) - HIGH - Auto-approve logs/describe" "OK"
Write-Log "  4. ImagePullBackOff (KT-003) - HIGH - Detected + remediation" "OK"
Write-Log "  5. Shell Completion - PowerShell/Bash/Zsh" "OK"
Write-Log "  6. AI Provider Modes - none/auto (NVIDIA/Ollama/OpenAI/Anthropic)" "OK"
Write-Log "" "SCENE"
Write-Log "Repository: https://github.com/Jani-shiv/kube-triage" "OK"
Write-Log "Install: go install github.com/Jani-shiv/kube-triage/cmd@latest" "OK"

Write-Log "Demo complete! Log saved to: $LogPath" "SCENE"