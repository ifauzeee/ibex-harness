//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/Rick1330/ibex-harness/infra/testing/testutil"
	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
	"github.com/google/uuid"
)

func TestTokensRepository_ListTokens_CursorPagination(t *testing.T) {
	repo, db := setupTokensRepo(t)
	orgID := testutil.SeedOrganization(t, db, "List Org", "list-"+uuid.NewString()[:8])
	id1 := insertNamedToken(t, repo, orgID, "token-a")
	id2 := insertNamedToken(t, repo, orgID, "token-b")
	id3 := insertNamedToken(t, repo, orgID, "token-c")

	page1, cursor, err := repo.ListTokens(context.Background(), orgID, "", 2)
	if err != nil || len(page1) != 2 || cursor == "" {
		t.Fatalf("page1: len=%d cursor=%q err=%v", len(page1), cursor, err)
	}
	page2, next, err := repo.ListTokens(context.Background(), orgID, cursor, 2)
	if err != nil || len(page2) != 1 || next != "" {
		t.Fatalf("page2: len=%d next=%q err=%v", len(page2), next, err)
	}
	assertTokenIDsPresent(t, []string{id1, id2, id3}, append(page1, page2...))
}

func assertTokenIDsPresent(t *testing.T, want []string, rows []repository.TokenMetadata) {
	t.Helper()
	seen := make(map[string]bool, len(want))
	for _, id := range want {
		seen[id] = false
	}
	for _, row := range rows {
		if _, ok := seen[row.ID]; ok {
			seen[row.ID] = true
		}
	}
	for id, found := range seen {
		if !found {
			t.Fatalf("missing token id %s", id)
		}
	}
}
