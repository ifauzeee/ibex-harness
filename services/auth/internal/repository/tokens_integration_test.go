//go:build integration

package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/Rick1330/ibex-harness/infra/testing/testutil"
	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
	"github.com/Rick1330/ibex-harness/services/auth/internal/token"
	"github.com/google/uuid"
)

func setupTokensRepo(t *testing.T) (*repository.TokensRepository, *sql.DB, func()) {
	t.Helper()
	dsn, cleanupPG := testutil.SetupPostgres(t)
	db := testutil.OpenDB(t, dsn)
	repo := repository.NewTokensRepository(db, nil)
	return repo, db, func() {
		_ = db.Close()
		cleanupPG()
	}
}

func TestTokensRepository_FindActiveByPrefix(t *testing.T) {
	tests := []struct {
		name      string
		revoked   bool
		expired   bool
		wantFound bool
	}{
		{name: "happy path", wantFound: true},
		{name: "revoked excluded", revoked: true, wantFound: false},
		{name: "expired excluded", expired: true, wantFound: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo, db, cleanup := setupTokensRepo(t)
			defer cleanup()

			orgID := testutil.SeedOrganization(t, db, "Find Org "+tc.name, "find-"+uuid.NewString()[:8])
			tokenID := uuid.New()
			bearer := "ibex_pat_" + tokenID.String() + "_findsecret"
			prefix := "ibex_pat_" + tokenID.String()
			hash, err := token.HashForTest(bearer, token.DefaultArgon2Params())
			if err != nil {
				t.Fatalf("hash: %v", err)
			}

			var expiresAt *time.Time
			if tc.expired {
				past := time.Now().UTC().Add(-time.Hour)
				expiresAt = &past
			}

			_, err = repo.InsertTestToken(context.Background(), orgID, prefix, hash, tc.name, 7, tc.revoked, expiresAt)
			if err != nil {
				t.Fatalf("insert token: %v", err)
			}

			row, err := repo.FindActiveByPrefix(context.Background(), prefix)

			if tc.wantFound {
				if err != nil {
					t.Fatalf("FindActiveByPrefix: %v", err)
				}
				if row.OrgID != orgID || row.Permissions != 7 {
					t.Fatalf("row: org=%s perms=%d", row.OrgID, row.Permissions)
				}
				return
			}

			if !errors.Is(err, sql.ErrNoRows) {
				t.Fatalf("expected sql.ErrNoRows, got %v", err)
			}
		})
	}
}

func TestTokensRepository_CreateToken(t *testing.T) {
	repo, db, cleanup := setupTokensRepo(t)
	defer cleanup()

	orgID := testutil.SeedOrganization(t, db, "Create Org", "create-"+uuid.NewString()[:8])
	rowID := uuid.New()

	id, err := repo.CreateToken(context.Background(), repository.CreateTokenParams{
		ID:          rowID.String(),
		OrgID:       orgID,
		Name:        "integration-create",
		Description: "desc",
		Hash:        "hash-placeholder",
		Prefix:      "ibex_pat_" + rowID.String(),
		Permissions: 42,
	})
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}
	if id != rowID.String() {
		t.Fatalf("id: got %s want %s", id, rowID.String())
	}
}

func TestTokensRepository_RevokeToken_ErrNotFound(t *testing.T) {
	repo, db, cleanup := setupTokensRepo(t)
	defer cleanup()

	orgID := testutil.SeedOrganization(t, db, "Revoke Org", "revoke-"+uuid.NewString()[:8])

	err := repo.RevokeToken(context.Background(), orgID, uuid.NewString(), "", nil)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTokensRepository_RevokeToken_HappyPath(t *testing.T) {
	repo, db, cleanup := setupTokensRepo(t)
	defer cleanup()

	orgID := testutil.SeedOrganization(t, db, "Revoke OK Org", "revoke-ok-"+uuid.NewString()[:8])
	tokenID := uuid.New()
	bearer := "ibex_pat_" + tokenID.String() + "_revokeme"
	prefix := "ibex_pat_" + tokenID.String()
	hash, err := token.HashForTest(bearer, token.DefaultArgon2Params())
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	id, err := repo.InsertTestToken(context.Background(), orgID, prefix, hash, "revoke-me", 1, false, nil)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	err = repo.RevokeToken(context.Background(), orgID, id, "", nil)
	if err != nil {
		t.Fatalf("RevokeToken: %v", err)
	}

	_, err = repo.FindActiveByPrefix(context.Background(), prefix)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("revoked token should not be active: %v", err)
	}
}

func TestTokensRepository_ListTokens_CursorPagination(t *testing.T) {
	repo, db, cleanup := setupTokensRepo(t)
	defer cleanup()

	orgID := testutil.SeedOrganization(t, db, "List Org", "list-"+uuid.NewString()[:8])

	insert := func(name string) string {
		t.Helper()
		tokenID := uuid.New()
		bearer := "ibex_pat_" + tokenID.String() + "_" + name
		prefix := "ibex_pat_" + tokenID.String()
		hash, err := token.HashForTest(bearer, token.DefaultArgon2Params())
		if err != nil {
			t.Fatalf("hash %s: %v", name, err)
		}
		id, err := repo.InsertTestToken(context.Background(), orgID, prefix, hash, name, 1, false, nil)
		if err != nil {
			t.Fatalf("insert %s: %v", name, err)
		}
		return id
	}

	id1 := insert("token-a")
	id2 := insert("token-b")
	id3 := insert("token-c")

	page1, cursor, err := repo.ListTokens(context.Background(), orgID, "", 2)
	if err != nil {
		t.Fatalf("list page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1 len: got %d want 2", len(page1))
	}
	if cursor == "" {
		t.Fatal("expected next cursor after page1")
	}

	page2, next, err := repo.ListTokens(context.Background(), orgID, cursor, 2)
	if err != nil {
		t.Fatalf("list page2: %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("page2 len: got %d want 1", len(page2))
	}
	if next != "" {
		t.Fatalf("expected empty cursor after final page, got %q", next)
	}

	seen := map[string]bool{id1: false, id2: false, id3: false}
	for _, row := range append(page1, page2...) {
		if _, ok := seen[row.ID]; ok {
			seen[row.ID] = true
		}
	}
	for id, found := range seen {
		if !found {
			t.Fatalf("missing token id %s in paginated results", id)
		}
	}
}
