#!/bin/bash
# record-demo.sh - Automated demo recording with asciinema
# Usage: ./record-demo.sh

set -e

CAST_FILE="kube-triage-demo.cast"
DEMO_DIR="/tmp/kube-triage-demo"

echo "🎬 Starting kube-triage demo recording..."
echo "Cast file: $CAST_FILE"

# Check asciinema
if ! command -v asciinema &> /dev/null; then
    echo "Installing asciinema..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install asciinema
    else
        sudo apt-get update && sudo apt-get install -y asciinema
    fi
fi

# Clean up previous demo resources
cleanup() {
    echo "🧹 Cleaning up demo resources..."
    kubectl delete pod demo-oom demo-crash demo-bad-image demo-exit1 --ignore-not-found=true 2>/dev/null
    kubectl delete deployment demo-oom-deploy --ignore-not-found=true 2>/dev/null
    rm -f /tmp/oom-pod.yaml /tmp/oom-deploy.yaml
}

trap cleanup EXIT

# Start recording
echo "🔴 Recording started. Press Ctrl+D to stop."
asciinema rec "$CAST_FILE" -c "
    echo '=== kube-triage v1.0.0 Demo ==='
    echo ''
    echo '📋 Version & Help'
    kube-triage version
    echo ''
    kube-triage --help
    echo ''
    echo '🔍 Scene 1: OOMKilled Detection (KT-001)'
    cat > /tmp/oom-pod.yaml << 'EOFPOD'
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
    command: ['/bin/sh', '-c']
    args:
      - |
        cat > /oom.go << 'EOGO'
        package main
        import (\"runtime\"; \"time\")
        func main() {
          var data []byte
          for { data = append(data, make([]byte, 1024*1024)...); runtime.GC(); time.Sleep(100*time.Millisecond) }
        }
        EOGO
        go run /oom.go
    resources:
      limits:
        memory: '10Mi'
      requests:
        memory: '5Mi'
EOFPOD
    kubectl apply -f /tmp/oom-pod.yaml
    sleep 15
    echo 'Pod status:'
    kubectl get pod demo-oom
    echo ''
    echo 'Analysis:'
    kube-triage analyze pod demo-oom -n default
    echo ''
    echo 'JSON output:'
    kube-triage analyze pod demo-oom -n default --output json | jq .
    echo ''
    echo '🔍 Scene 2: Deployment Auto-Fix'
    cat > /tmp/oom-deploy.yaml << 'EOFDEPLOY'
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
        command: ['/bin/sh', '-c']
        args:
          - |
            cat > /oom.go << 'EOGO'
            package main
            import (\"runtime\"; \"time\")
            func main() {
              var data []byte
              for { data = append(data, make([]byte, 1024*1024)...); runtime.GC(); time.Sleep(100*time.Millisecond) }
            }
            EOGO
            go run /oom.go
        resources:
          limits:
            memory: '10Mi'
          requests:
            memory: '5Mi'
EOFDEPLOY
    kubectl apply -f /tmp/oom-deploy.yaml
    sleep 20
    echo 'Deployment status:'
    kubectl get deployment demo-oom-deploy
    kubectl get pods -l app=demo-oom
    echo ''
    echo 'Analysis (deployment):'
    kube-triage analyze deployment demo-oom-deploy -n default
    echo ''
    echo 'Dry-run:'
    kube-triage analyze deployment demo-oom-deploy -n default --dry-run
    echo ''
    echo 'AUTO-APPROVE (actual fix):'
    kube-triage analyze deployment demo-oom-deploy -n default --auto-approve
    echo ''
    echo 'Verify fix:'
    kubectl get deployment demo-oom-deploy -o yaml | grep -A 5 'resources:'
    echo ''
    echo '🔍 Scene 3: CrashLoopBackOff (KT-002)'
    kubectl run demo-crash --image=busybox --restart=Always --command -- sh -c 'exit 1'
    sleep 10
    kube-triage analyze pod demo-crash -n default --auto-approve
    kubectl delete pod demo-crash --ignore-not-found=true
    echo ''
    echo '🔍 Scene 4: ImagePullBackOff (KT-003)'
    kubectl run demo-bad-image --image=nonexistent/registry/invalid:latest --restart=Never
    sleep 10
    kube-triage analyze pod demo-bad-image -n default
    kubectl delete pod demo-bad-image --ignore-not-found=true
    echo ''
    echo '🔍 Scene 5: Shell Completion'
    source <(kube-triage completion bash)
    echo 'Try: kube-triage analyze pod <TAB>'
    echo ''
    echo '✅ Demo complete!'
    echo 'Repo: https://github.com/Jani-shiv/kube-triage'
" 2>&1 | tee demo-output.log

echo ""
echo "✅ Recording saved to: $CAST_FILE"
echo "📺 Play with: asciinema play $CAST_FILE"
echo "☁️  Upload with: asciinema upload $CAST_FILE"
echo "📝 Output log: demo-output.log"