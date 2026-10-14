# User onboarding

=== "For user"

    ## Create user

    Ask an admin to create your account, provide the following information:

    - [ ] Full name (John Doe)
    - [ ] Select a username (`johndoe`)
    - [ ] Email address (`johndoe@example.com`)
    - [ ] Which access tier you need: `admin`, `editor`, or `viewer` (ask if unsure -- `viewer` is the safe default)

    ## Install companion apps

    For all users:

    - [ ] A password manager (I personally recommend [Bitwarden](https://bitwarden.com/download))
    - [ ] A [Matrix chat client](https://matrix.org/clients) (optional, you can use the web version)

    For technical users:

    - [ ] [Docker](https://docs.docker.com/engine/install)
    - [ ] [Nix](https://nixos.org/download) and [direnv](https://direnv.net) (optional, but highly recommended)
    - [ ] [Lens](https://k8slens.dev) (optional, you can use the included `kubectl` or `k9s` command in the tools container)

=== "For admin"

    Run the following script, with the tier as the 4th argument (defaults to `editor` if omitted):

    ```sh
    ./scripts/onboard-user johndoe "John Doe" "johndoe@example.com" viewer
    ```

    Let the user scan the QR code or follow the link to set up passkeys or password + TOTP.

    ### Access tiers

    - `admin`: full access, including Grafana's Admin org role
    - `editor`: the original default -- Grafana Editor, access to everything Dex fronts
    - `viewer`: can log in and view dashboards (Grafana Viewer), but not change anything

    Only Grafana actually enforces different behavior per tier today (via `role_attribute_path`
    in `platform/grafana/values.yaml`). Apps without native OIDC group support (Jellyfin, Ollama/Open WebUI)
    don't yet distinguish between tiers -- anyone who can log in via Dex has the same access to those.
