//go:build integration

package repotest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Rick1330/ibex-harness/infra/testing/testutil"
	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
	"github.com/google/uuid"
)

// AgentsScenario holds seeded orgs, user, and agent for repository integration tests.
type AgentsScenario struct {
	DB      *sql.DB
	Repo    *repository.AgentsRepository
	OrgA    string
	OrgB    string
	UserA   string
	AgentID string
}

// WithAgentsScenario seeds postgres and runs fn with a ready agents repository fixture.
func WithAgentsScenario(t *testing.T, status string, crossOrg bool, fn func(t *testing.T, s AgentsScenario, lookupOrg string)) {
	t.Helper()
	dsn, cleanupPG := testutil.SetupPostgres(t)
	t.Cleanup(cleanupPG)

	db := testutil.OpenDB(t, dsn)
	t.Cleanup(func() { _ = db.Close() })

	repo := repository.NewAgentsRepository(db, nil)
	label := uuid.NewString()[:8]
	orgA := testutil.SeedOrganization(t, db, "Org A "+label, "org-a-"+label)
	orgB := testutil.SeedOrganization(t, db, "Org B "+label, "org-b-"+label)
	userA := testutil.SeedUser(t, db, orgA, "user-"+label+"@test.local", "User A")
	agentID := testutil.SeedAgentWithStatus(t, db, orgA, userA, "Agent "+label, "agent-"+label, status)

	lookupOrg := orgA
	if crossOrg {
		lookupOrg = orgB
	}

	fn(t, AgentsScenario{DB: db, Repo: repo, OrgA: orgA, OrgB: orgB, UserA: userA, AgentID: agentID}, lookupOrg)
}

// LookupAgent calls GetByIDAndOrg for the scenario agent.
func (s AgentsScenario) LookupAgent(t *testing.T, lookupOrg string) *repository.AgentRecord {
	t.Helper()
	rec, err := s.Repo.GetByIDAndOrg(context.Background(), uuid.MustParse(s.AgentID), uuid.MustParse(lookupOrg))
	if err != nil {
		t.Fatalf("GetByIDAndOrg: %v", err)
	}
	return rec
}
