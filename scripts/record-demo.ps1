<# 
.SYNOPSIS
    Record kube-triage demo with asciinema on Windows
.DESCRIPTION
    Automated demo recording script for Windows PowerShell
.EXAMPLE
    .\record-demo.ps1
#>

param(
    [string]$CastFile = "kube-triage-demo.cast"
)

Write-Host "🎬 Starting kube-triage demo recording..." -ForegroundColor Cyan
Write-Host "Cast file: $CastFile" -ForegroundColor Gray

# Check asciinema
if (-not (Get-Command asciinema -ErrorAction SilentlyContinue)) {
    Write-Host "Installing asciinema via Chocolatey..." -ForegroundColor Yellow
    if (Get-Command choco -ErrorAction SilentlyContinue) {
        choco install asciinema -y
    } else {
        Write-Error "Please install asciinema manually: https://asciinema.org/docs/installation"
        exit 1
    }
}

function Cleanup {
    Write-Host "🧹 Cleaning up demo resources..." -ForegroundColor Yellow
    kubectl delete pod demo-oom, demo-crash, demo-bad-image, demo-exit1 --ignore-not-found=true 2>$null
    kubectl delete deployment demo-oom-deploy --ignore-not-found=true 2>$null
    Remove-Item /tmp/oom-pod.yaml, /tmp/oom-deploy.yaml -ErrorAction SilentlyContinue
}

$cleanupAction = { Cleanup }
Set-StrictMode -Version Latest
$ErrorActionPreference = "Continue"

try {
    Write-Host "🔴 Recording started. Press Ctrl+D to stop." -ForegroundColor Red
    
    $scriptBlock = {
        Write-Host "=== kube-triage v1.0.0 Demo ===" -ForegroundColor Green
        Write-Host ""
        Write-Host "📋 Version & Help" -ForegroundColor Cyan
        kube-triage version
        Write-Host ""
        kube-triage --help
        Write-Host ""
        
        Write-Host "🔍 Scene 1: OOMKilled Detection (KT-001)" -ForegroundColor Cyan
        @"
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
"@ | Out-File /tmp/oom-pod.yaml -Encoding utf8

        kubectl apply -f /tmp/oom-pod.yaml
        Start-Sleep 15
        Write-Host "Pod status:" -ForegroundColor Yellow
        kubectl get pod demo-oom
        Write-Host ""
        Write-Host "Analysis:" -ForegroundColor Yellow
        kube-triage analyze pod demo-oom -n default
        Write-Host ""
        Write-Host "JSON output:" -ForegroundColor Yellow
        kube-triage analyze pod demo-oom -n default --output json | ConvertFrom-Json | ConvertTo-Json -Depth 10
        Write-Host ""
        
        Write-Host "🔍 Scene 2: Deployment Auto-Fix" -ForegroundColor Cyan
        @"
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
"@ | Out-File /tmp/oom-deploy.yaml -Encoding utf8

        kubectl apply -f /tmp/oom-deploy.yaml
        Start-Sleep 20
        Write-Host "Deployment status:" -ForegroundColor Yellow
        kubectl get deployment demo-oom-deploy
        kubectl get pods -l app=demo-oom
        Write-Host ""
        Write-Host "Analysis (deployment):" -ForegroundColor Yellow
        kube-triage analyze deployment demo-oom-deploy -n default
        Write-Host ""
        Write-Host "Dry-run:" -ForegroundColor Yellow
        kube-triage analyze deployment demo-oom-deploy -n default --dry-run
        Write-Host ""
        Write-Host "AUTO-APPROVE (actual fix):" -ForegroundColor Yellow
        kube-triage analyze deployment demo-oom-deploy -n default --auto-approve
        Write-Host ""
        Write-Host "Verify fix:" -ForegroundColor Yellow
        kubectl get deployment demo-oom-deploy -o yaml | Select-String "resources:" -Context 0,5
        Write-Host ""
        
        Write-Host "🔍 Scene 3: CrashLoopBackOff (KT-002)" -ForegroundColor Cyan
        kubectl run demo-crash --image=busybox --restart=Always --command -- sh -c "exit 1"
        Start-Sleep 10
        kube-triage analyze pod demo-crash -n default --auto-approve
        kubectl delete pod demo-crash --ignore-not-found=true
        Write-Host ""
        
        Write-Host "🔍 Scene 4: ImagePullBackOff (KT-003)" -ForegroundColor Cyan
        kubectl run demo-bad-image --image=nonexistent/registry/invalid:latest --restart=Never
        Start-Sleep 10
        kube-triage analyze pod demo-bad-image -n default
        kubectl delete pod demo-bad-image --ignore-not-found=true
        Write-Host ""
        
        Write-Host "🔍 Scene 5: Shell Completion" -ForegroundColor Cyan
        $completion = kube-triage completion powershell
        Invoke-Expression $completion
        Write-Host "Try: kube-triage analyze pod <TAB>" -ForegroundColor Yellow
        Write-Host ""
        
        Write-Host "✅ Demo complete!" -ForegroundColor Green
        Write-Host "Repo: https://github.com/Jani-shiv/kube-triage" -ForegroundColor Cyan
    }
    
    asciinema rec $CastFile -c $scriptBlock 2>&1 | Tee-Object -FilePath demo-output.log
    
} finally {
    Cleanup
}

Write-Host ""
Write-Host "✅ Recording saved to: $CastFile" -ForegroundColor Green
Write-Host "📺 Play with: asciinema play $CastFile" -ForegroundColor Cyan
Write-Host "☁️  Upload with: asciinema upload $CastFile" -ForegroundColor Cyan
Write-Host "📝 Output log: demo-output.log" -ForegroundColor Gray