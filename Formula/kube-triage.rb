class KubeTriage < Formula
  desc "AI-powered Kubernetes diagnostic engine"
  homepage "https://github.com/Jani-shiv/kube-triage"
  license "MIT"
  version "1.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Jani-shiv/kube-triage/releases/download/v1.0.0/kube-triage_1.0.0_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_ARM64"
    else
      url "https://github.com/Jani-shiv/kube-triage/releases/download/v1.0.0/kube-triage_1.0.0_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/Jani-shiv/kube-triage/releases/download/v1.0.0/kube-triage_1.0.0_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_ARM64"
    else
      url "https://github.com/Jani-shiv/kube-triage/releases/download/v1.0.0/kube-triage_1.0.0_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_AMD64"
    end
  end

  on_windows do
    url "https://github.com/Jani-shiv/kube-triage/releases/download/v1.0.0/kube-triage_1.0.0_windows_amd64.zip"
    sha256 "REPLACE_WITH_ACTUAL_SHA256_WINDOWS"
  end

  def install
    bin.install "kube-triage"
  end

  def caveats
    <<~EOS
      kube-triage requires a Kubernetes cluster and kubectl configured.

      Quick start:
        kube-triage analyze pod <pod-name> -n <namespace>

      For AI-powered analysis, set one of these environment variables:
        export NVIDIA_API_KEY=nvapi-xxxxx     # NVIDIA Nemotron (free tier)
        export OLLAMA_HOST=http://localhost:11434  # Local Ollama
        export OPENAI_API_KEY=sk-xxxxx        # OpenAI
        export ANTHROPIC_API_KEY=sk-ant-xxxxx # Anthropic

      Config file: ~/.kube-triage.yaml
    EOS
  end

  test do
    system "#{bin}/kube-triage", "version"
  end
end