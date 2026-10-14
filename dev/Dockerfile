# Containerized alternative to `nix develop` (../flake.nix), for anywhere
# Nix itself isn't available or its daemon isn't reachable -- e.g. a
# sandboxed CI runner or a coding agent working on this repo without a host
# Nix install. Not a full 1:1 replacement for the Nix shell: it covers the
# tools exercised by static validation (Ansible syntax-check/lint, Helm
# template/lint, Go tests, pre-commit, shellcheck, yamllint, OpenTofu) rather
# than every package there (k9s, neovim, kube3d, wireguard-tools, and a few
# others are left out -- this is for validating changes, not for actually
# operating a live cluster from inside the container).
#
# Version pins are chosen to satisfy the ranges test/tools_test.go checks
# for, and were verified against the real upstream download URLs when this
# was written.
FROM debian:bookworm-slim

ARG DOCKER_VERSION=25.0.5
ARG GO_VERSION=1.22.4
ARG GOTESTSUM_VERSION=1.12.0
ARG HELM_VERSION=3.16.2
ARG KUBECTL_VERSION=1.31.1
ARG KUSTOMIZE_VERSION=5.4.3
ARG TOFU_VERSION=1.8.3

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates \
      curl \
      git \
      jq \
      python3 \
      python3-pip \
      shellcheck \
      unzip \
      xz-utils \
    && rm -rf /var/lib/apt/lists/*

# Docker CLI only (no daemon) -- satisfies test/tools_test.go's version
# check and lets `docker` work against a mounted host socket
# (-v /var/run/docker.sock:/var/run/docker.sock) if that's ever needed;
# this container doesn't run dockerd itself.
RUN curl -fsSL "https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz" \
      | tar -C /usr/local/bin --strip-components=1 -xz docker/docker

# Go + gotestsum
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"
RUN go install "gotest.tools/gotestsum@v${GOTESTSUM_VERSION}"

# Helm
RUN curl -fsSL "https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz" | tar -C /tmp -xz \
    && mv /tmp/linux-amd64/helm /usr/local/bin/helm \
    && rm -rf /tmp/linux-amd64

# kubectl
RUN curl -fsSL -o /usr/local/bin/kubectl \
      "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl" \
    && chmod +x /usr/local/bin/kubectl

# kustomize
RUN curl -fsSL \
      "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_amd64.tar.gz" \
      | tar -C /usr/local/bin -xz

# OpenTofu
RUN curl -fsSL -o /tmp/tofu.zip \
      "https://github.com/opentofu/opentofu/releases/download/v${TOFU_VERSION}/tofu_${TOFU_VERSION}_linux_amd64.zip" \
    && unzip -o /tmp/tofu.zip -d /usr/local/bin tofu \
    && rm /tmp/tofu.zip

# Ansible, pre-commit, yamllint, ansible-lint, and the Python packages the
# metal/ playbooks and scripts/ tools import -- pinned to satisfy
# test/tools_test.go's version constraints.
RUN pip install --break-system-packages --no-cache-dir \
      "ansible-core>=2.16,<2.17" \
      ansible-lint \
      "pre-commit>=3.3.2,<4.0.0" \
      "yamllint>=1.27.1" \
      jinja2 \
      kubernetes \
      netaddr \
      pexpect \
      rich

RUN ansible-galaxy collection install \
      ansible.netcommon \
      ansible.posix \
      community.docker \
      community.general \
      kubernetes.core

WORKDIR /workspace
