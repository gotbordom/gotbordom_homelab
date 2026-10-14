.POSIX:
.PHONY: *
.EXPORT_ALL_VARIABLES:

KUBECONFIG = $(shell pwd)/metal/kubeconfig.yaml
KUBE_CONFIG_PATH = $(KUBECONFIG)

default: metal system external smoke-test post-install clean

configure:
	./scripts/configure
	git status

metal:
	make -C metal

system:
	make -C system

external:
	make -C external

smoke-test:
	make -C test filter=Smoke

post-install:
	@./scripts/hacks

# TODO maybe there's a better way to manage backup with GitOps?
backup:
	./scripts/backup --action setup --namespace=actualbudget --pvc=actualbudget-data
	./scripts/backup --action setup --namespace=jellyfin --pvc=jellyfin-data

restore:
	./scripts/backup --action restore --namespace=actualbudget --pvc=actualbudget-data
	./scripts/backup --action restore --namespace=jellyfin --pvc=jellyfin-data

test:
	make -C test

clean:
	docker compose --project-directory ./metal/roles/pxe_server/files down

docs:
	mkdocs serve

git-hooks:
	pre-commit install

# Containerized alternative to `nix develop` (./flake.nix) -- for anywhere
# without Nix or without access to the Nix daemon. See dev/Dockerfile and
# docs/concepts/development-shell.md for what's covered and what isn't.
dev-build:
	docker build -t homelab-dev -f dev/Dockerfile dev/

dev-shell: dev-build
	docker run --rm -it \
		-v $(CURDIR):/workspace \
		-v homelab-dev-go:/root/go \
		-v homelab-dev-pip-cache:/root/.cache/pip \
		-w /workspace \
		homelab-dev bash
