//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/Rick1330/ibex-harness/infra/testing/testutil"
	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
	"github.com/google/uuid"
)

func TestAgentsRepository_GetByIDAndOrg(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		crossOrg   bool
		wantNil    bool
		wantStatus string
	}{
		{name: "active agent", status: "active", wantStatus: "active"},
		{name: "wrong org returns nil", status: "active", crossOrg: true, wantNil: true},
		{name: "paused agent", status: "paused", wantStatus: "paused"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dsn, cleanupPG := testutil.SetupPostgres(t)
			defer cleanupPG()

			db := testutil.OpenDB(t, dsn)
			defer db.Close()

			repo := repository.NewAgentsRepository(db, nil)
			orgA := testutil.SeedOrganization(t, db, "Org A "+tc.name, "org-a-"+uuid.NewString()[:8])
			orgB := testutil.SeedOrganization(t, db, "Org B "+tc.name, "org-b-"+uuid.NewString()[:8])
			userA := testutil.SeedUser(t, db, orgA, "user-"+uuid.NewString()[:8]+"@test.local", "User A")

			agentID := testutil.SeedAgentWithStatus(
				t, db, orgA, userA, "Agent "+tc.name, "agent-"+uuid.NewString()[:8], tc.status,
			)

			lookupOrg := orgA
			if tc.crossOrg {
				lookupOrg = orgB
			}

			rec, err := repo.GetByIDAndOrg(context.Background(), uuid.MustParse(agentID), uuid.MustParse(lookupOrg))
			if err != nil {
				t.Fatalf("GetByIDAndOrg: %v", err)
			}

			if tc.wantNil {
				if rec != nil {
					t.Fatalf("expected nil for cross-org lookup, got %+v", rec)
				}
				return
			}

			if rec == nil {
				t.Fatal("expected agent record")
			}
			if rec.ID != agentID || rec.OrgID != orgA || rec.Status != tc.wantStatus {
				t.Fatalf("record mismatch: %+v", rec)
			}
		})
	}
}
