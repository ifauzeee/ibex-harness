package service

import (
	"database/sql"
	"testing"
	"time"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
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

func protoListFromSample() (repository.TokenMetadata, *authv1.TokenMetadata) {
	row := sampleTokenMetadataRow()
	out := ToProtoList([]repository.TokenMetadata{row})
	if len(out) != 1 {
		panic("expected one proto row")
	}
	return row, out[0]
}

func TestToProtoList_identity(t *testing.T) {
	t.Parallel()
	row, m := protoListFromSample()
	if m.GetTokenId() != row.ID || m.GetName() != row.Name {
		t.Fatalf("metadata: %+v", m)
	}
}

func TestToProtoList_revoked(t *testing.T) {
	t.Parallel()
	_, m := protoListFromSample()
	if !m.GetIsRevoked() {
		t.Fatal("expected is_revoked true")
	}
}

func TestToProtoList_timestamps(t *testing.T) {
	t.Parallel()
	row, m := protoListFromSample()
	if m.GetCreatedAt().AsTime() != row.CreatedAt {
		t.Fatalf("created_at: %v", m.GetCreatedAt().AsTime())
	}
	if m.GetExpiresAt().AsTime() != row.ExpiresAt.Time {
		t.Fatalf("expires_at: %v", m.GetExpiresAt().AsTime())
	}
	if m.GetRevokedAt().AsTime() != row.RevokedAt.Time {
		t.Fatalf("revoked_at: %v", m.GetRevokedAt().AsTime())
	}
}
