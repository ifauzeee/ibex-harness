package auth

import "testing"

func assertAgentRecord(t *testing.T, got, want *AgentRecord) {
	t.Helper()
	if got.ID != want.ID || got.OrgID != want.OrgID || got.Status != want.Status {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
