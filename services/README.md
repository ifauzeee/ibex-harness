# services/

Deployable runtime components for IBEX Harness.

Planned services (see [docs/FILE_STRUCTURE.md](../docs/FILE_STRUCTURE.md)):

| Directory | Role |
| --- | --- |
| `proxy/` | Go — LLM proxy (latency-critical) |
| `auth/` | Go — authentication and token validation |
| `api/` | Python FastAPI — management plane |
| `memory/` | Python FastAPI — memory CRUD and vector search |
| `context/` | Python gRPC — context assembly |
| `embedder/` | Python FastAPI — embeddings |
| `worker/` | Python Celery — async jobs |
| `dashboard/` | Next.js — operator UI |

No service implementation code exists yet. Bootstrap the first skeleton via [docs/FILE_STRUCTURE.md](../docs/FILE_STRUCTURE.md) scaffolds.
