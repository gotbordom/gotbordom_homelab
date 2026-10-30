# Expose services to the internet

!!! info

    This tutorial is for Cloudflare Tunnel users, please skip if you use port forwarding.

Nothing is public by default: every app's DNS record points at its private LAN
address, and the tunnel 404s any hostname it doesn't explicitly list. Exposing an
app takes two deliberate changes.

!!! warning

    Only expose apps that have their own strong authentication, or put them
    behind SSO (Kanidm) or Cloudflare Access first. Admin tools like
    Transmission, Radarr, Sonarr and Prowlarr should stay private (LAN/Tailscale).

Apply the `./external` layer to create a tunnel if you haven't already, then:

1. Add the hostname to the tunnel's ingress rules in `system/cloudflared/values.yaml`
   (above the final `http_status:404` rule):

```yaml
- hostname: myapp.example.com
  service: https://ingress-nginx-controller.ingress-nginx
  originRequest:
    noTLSVerify: true
```

2. Add the following annotations to your `Ingress` object (replace `example.com` with your domain):

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/target: "homelab-tunnel.example.com"
    external-dns.alpha.kubernetes.io/cloudflare-proxied: "true"
# ...
```
