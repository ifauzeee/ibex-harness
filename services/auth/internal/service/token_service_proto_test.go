package service

import (
	"database/sql"
	"testing"
	"time"

	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
)

func sampleTokenMetadataRow() repository.TokenMetadata {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	return repository.TokenMetadata{
		ID: "tid", Name: "n", Prefix: "ibex_pat_x", Permissions: 99, CreatedAt: created,
		ExpiresAt: sql.NullTime{Time: created.Add(24 * time.Hour), Valid: true},
		RevokedAt: sql.NullTime{Time: created.Add(time.Hour), Valid: true},
		IsRevoked: true,
	}
}

func TestToProtoList_AllFields(t *testing.T) {
	t.Parallel()

	row := sampleTokenMetadataRow()
	out := ToProtoList([]repository.TokenMetadata{row})
	if len(out) != 1 {
		t.Fatalf("len: %d", len(out))
	}
	m := out[0]

	t.Run("identity fields", func(t *testing.T) {
		if m.GetTokenId() != row.ID || m.GetName() != row.Name || m.GetPrefix() != row.Prefix || m.GetPermissions() != row.Permissions {
			t.Fatalf("metadata fields: %+v", m)
		}
	})
	t.Run("revoked flag", func(t *testing.T) {
		if !m.GetIsRevoked() {
			t.Fatal("expected is_revoked true")
		}
	})
	t.Run("timestamps", func(t *testing.T) {
		if m.GetCreatedAt().AsTime() != row.CreatedAt {
			t.Fatalf("created_at: %v", m.GetCreatedAt().AsTime())
		}
		if m.GetExpiresAt().AsTime() != row.ExpiresAt.Time {
			t.Fatalf("expires_at: %v", m.GetExpiresAt().AsTime())
		}
		if m.GetRevokedAt().AsTime() != row.RevokedAt.Time {
			t.Fatalf("revoked_at: %v", m.GetRevokedAt().AsTime())
		}
	})
}
