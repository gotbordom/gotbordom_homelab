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

test:
	make -C test

# Turn on when away from home, off at home (on sends LAN traffic through the router)
routes-on:
	sudo tailscale set --accept-routes=true

routes-off:
	sudo tailscale set --accept-routes=false

wake:
	@./scripts/wake $(node)

recovery-backup:
	@./scripts/recovery-backup $(dest)

clean:
	docker compose --project-directory ./metal/roles/pxe_server/files down

docs:
	mkdocs serve

git-hooks:
	pre-commit install
