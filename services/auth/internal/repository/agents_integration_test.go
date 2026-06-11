//go:build integration

package repository_test

import (
	"testing"

	"github.com/Rick1330/ibex-harness/infra/testing/repotest"
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
			repotest.WithAgentsScenario(t, tc.status, tc.crossOrg, func(t *testing.T, s repotest.AgentsScenario, lookupOrg string) {
				rec := s.LookupAgent(t, lookupOrg)
				if tc.wantNil {
					if rec != nil {
						t.Fatalf("expected nil for cross-org lookup, got %+v", rec)
					}
					return
				}
				if rec == nil {
					t.Fatal("expected agent record")
				}
				if rec.ID != s.AgentID || rec.OrgID != s.OrgA || rec.Status != tc.wantStatus {
					t.Fatalf("record mismatch: %+v", rec)
				}
			})
		})
	}
}
