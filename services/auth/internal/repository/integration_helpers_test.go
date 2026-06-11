//go:build integration

package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Rick1330/ibex-harness/infra/testing/testutil"
	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
	"github.com/google/uuid"
)

type agentsScenario struct {
	DB      *sql.DB
	Repo    *repository.AgentsRepository
	OrgA    string
	OrgB    string
	UserA   string
	AgentID string
}

func withAgentsScenario(t *testing.T, status string, crossOrg bool, fn func(t *testing.T, s agentsScenario, lookupOrg string)) {
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

	fn(t, agentsScenario{DB: db, Repo: repo, OrgA: orgA, OrgB: orgB, UserA: userA, AgentID: agentID}, lookupOrg)
}

func (s agentsScenario) lookupAgent(t *testing.T, lookupOrg string) *repository.AgentRecord {
	t.Helper()
	rec, err := s.Repo.GetByIDAndOrg(context.Background(), uuid.MustParse(s.AgentID), uuid.MustParse(lookupOrg))
	if err != nil {
		t.Fatalf("GetByIDAndOrg: %v", err)
	}
	return rec
}
