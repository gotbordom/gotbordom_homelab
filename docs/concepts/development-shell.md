# Development shell

The development shell makes it easy to get all of the dependencies needed to interact with the homelab.

## Prerequisites

!!! info

    NixOS users can skip this step.

Install Nix using one of the following methods:

- [Official Nix installer](https://nixos.org/download)
- [Determinate Nix Installer](https://docs.determinate.systems/getting-started/#installer)

If you're using the official installer, add the following to your
`~/.config/nix/nix.conf` to enable [Flakes](https://nixos.wiki/wiki/Flakes):

```conf
experimental-features = nix-command flakes
```

## How to open it

Run the following command:

```sh
nix develop
```

It will open a shell with all the dependencies defined in `./flake.nix`:

```
[khuedoan@ryzentower:~/Documents/homelab]$ which kubectl
/nix/store/0558zzzqynzw7rx9dp2i7jymvznd1cqx-kubectl-1.30.1/bin/kubectl
```

!!! tip

    If you have [`direnv`](https://direnv.net) installed, you can run `direnv
    allow` once and it will automatically enter the Nix shell every time you
    `cd` into the project.

## Containerized alternative

If you don't have Nix (or can't reach its daemon -- e.g. inside a sandboxed
CI runner or a coding agent), `dev/Dockerfile` provides a Docker-based
alternative covering the tools exercised by static validation: Ansible
(syntax-check/lint), Helm (template/lint), Go (tests), `pre-commit`,
`shellcheck`, `yamllint`, and OpenTofu. It's not a full replacement for the
Nix shell -- things like `k9s`, `neovim`, and `kube3d` aren't included, since
they're for interactively operating a live cluster rather than validating
changes.

```sh
make dev-shell
```

This builds the image (if needed) and drops you into a shell with the repo
mounted at `/workspace`. Version pins in `dev/Dockerfile` are chosen to
satisfy the same ranges `test/tools_test.go` checks for.
