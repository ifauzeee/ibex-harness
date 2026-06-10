# Coverage Gap Register

**Target:** ≥94% merged statement coverage (unit + integration) on **hand-written code only**.  
**Excluded:** `packages/proto/gen/go/**` (generated protobuf). Contract tests remain in `packages/proto/`.

**Baseline (pre–Phase 0 drive):** ~51% Codecov badge | ~37% unit total

**Post–Tier B delete (unit, hand-written):** **83.5%** (`go tool cover -func coverage-go-handwritten.out`)

| Package | Baseline % | Current unit % | Target % | Status |
|---------|------------|----------------|----------|--------|
| `services/auth/internal/repository` | 0 | 0 (integration) | ≥94 | covered in merged profile |
| `services/auth/cmd/auth` | 0 | ~49 | ≥94 | partial — `run()` happy path integration-adjacent |
| `services/proxy/cmd/proxy` | 0 | ~80 | ≥94 | partial — `runWithShutdown` + setup helpers |
| `packages/proto/gen/go/**` | — | **excluded** | n/a | not gated |
| `services/auth/internal/service` | 7.5 | 95.0 | ≥94 | done |
| `services/proxy/internal/auth` | 17.3 | 94.7 | ≥94 | done |
| `services/auth/internal/grpc` | 24 | 96.8 | ≥94 | done |
| `services/proxy/internal/validation` | 83.7 | **100** | ≥94 | done |
| `packages/ratelimit` | 69.8 | 93.0 | ≥94 | near target |
| `packages/healthcheck` | 67.3 | 88.9 | ≥94 | merged profile |
| `packages/config` | 19.6 | 86.7 | ≥94 | `MustLoad` subprocess test added |
| `services/proxy/internal/llm` | 70.4 | 88.9 | ≥94 | augmented |

**Tooling:** `make coverage-report`, `infra/scripts/coverage-filter.sh`, `infra/scripts/coverage-gate.sh`

**CI:** `coverage` job merges unit + integration (Postgres service), filters gen/go, enforces **94%** on hand-written scope.

**Local merged profile:**

```bash
make compose-test-up
POSTGRES_TEST_DSN=postgres://ibex:ibex@localhost:5433/ibex_test?sslmode=disable make coverage-report
bash infra/scripts/coverage-gate.sh coverage-go-merged.out
```
