//go:build integration

package repository_test

import (
	"testing"

	"github.com/Rick1330/ibex-harness/infra/testing/testutil"
	"github.com/google/uuid"
)

func TestTokensRepository_ListTokens_CursorPagination(t *testing.T) {
	repo, db := setupTokensRepo(t)
	orgID := testutil.SeedOrganization(t, db, "List Org", "list-"+uuid.NewString()[:8])
	id1 := insertNamedToken(t, repo, orgID, "token-a")
	id2 := insertNamedToken(t, repo, orgID, "token-b")
	id3 := insertNamedToken(t, repo, orgID, "token-c")

	page1, cursor := listTokensPage(t, repo, listPageQuery{orgID: orgID, limit: 2, wantLen: 2, wantCursor: true})
	page2, _ := listTokensPage(t, repo, listPageQuery{orgID: orgID, cursor: cursor, limit: 2, wantLen: 1})
	assertTokenIDsPresent(t, []string{id1, id2, id3}, append(page1, page2...))
}
