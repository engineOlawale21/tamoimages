# Production edge and load-balancer requirements

Local Compose exposes services directly for development. Production must use a managed CDN/WAF and load balancer or ingress; this repository does not pretend a local proxy is equivalent to the production control plane.

```text
Internet -> CDN/WAF -> TLS load balancer/ingress -> route layer
                                              |-> auth-service pool
                                              `-> media-service pool
```

Required production behavior:

- Route `/api/v1/auth`, identity health endpoints, and future account paths to the auth-service pool.
- Route `/api/v1/media`, `/api/v1/uploads`, media health endpoints, and future catalogue paths to the media-service pool.
- Use readiness to add/remove targets and liveness only for restart decisions.
- Enable connection draining for graceful deployments.
- Generate or sanitize `X-Correlation-ID`; strip untrusted forwarding headers at the edge.
- Trust client-address forwarding only from the configured load-balancer network.
- Do not enable sticky sessions. Shared session and rate-limit state belongs in Redis.
- Terminate TLS with managed certificates and apply WAF, request-size, connection, and denial-of-service controls.
- Scale the auth and media pools independently and publish target health, latency, error, and rejection metrics.

The media service's in-memory catalogue is development-only and is a blocker for horizontal production scaling until PostgreSQL-backed persistence is enabled.
