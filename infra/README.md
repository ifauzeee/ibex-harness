# infra/

Deployment, local development infrastructure, and observability configuration.

Planned layout (see [docs/FILE_STRUCTURE.md](../docs/FILE_STRUCTURE.md)):

| Directory | Role |
| --- | --- |
| `compose/dev/` | Docker Compose for local dependencies |
| `compose/test/` | Compose for CI/integration tests |
| `docker/` | Shared Docker build assets |
| `helm/` | Kubernetes Helm charts |
| `terraform/` | Cloud infrastructure as code |
| `monitoring/` | Prometheus, Grafana, Loki, Tempo configs |
| `scripts/` | Operational helpers |

No compose or IaC files exist yet. Next milestone: `infra/compose/dev` per [docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md).
