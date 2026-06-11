package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTokenService_CreateToken(t *testing.T) {
	t.Parallel()

	orgID := uuid.New().String()
	expires := time.Now().UTC().Add(24 * time.Hour)
	userID := uuid.New().String()
	agentID := uuid.New().String()

	tests := []struct {
		name    string
		req     *authv1.CreateTokenRequest
		wantErr error
	}{
		{
			name:    "invalid empty org",
			req:     &authv1.CreateTokenRequest{Name: "x"},
			wantErr: ErrInvalidArgument,
		},
		{
			name:    "invalid empty name",
			req:     &authv1.CreateTokenRequest{OrgId: orgID},
			wantErr: ErrInvalidArgument,
		},
		{
			name: "invalid token type",
			req: &authv1.CreateTokenRequest{
				OrgId: orgID, Name: "bad-type", Type: authv1.TokenType(99),
			},
			wantErr: ErrInvalidArgument,
		},
		{
			name: "happy path",
			req: &authv1.CreateTokenRequest{
				OrgId:       orgID,
				Name:        "ci-pat",
				Description: "desc",
				Permissions: 42,
				UserId:      &userID,
				AgentId:     &agentID,
				ExpiresAt:   timestamppb.New(expires),
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := newMemTokenRepo()
			svc := testTokenService(repo)

			result, err := svc.CreateToken(context.Background(), tc.req)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("CreateToken err: got %v want %v", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("CreateToken: %v", err)
			}
			if result.TokenID == "" || result.Plaintext == "" || result.Prefix == "" {
				t.Fatalf("incomplete result: %+v", result)
			}
			if _, ok := repo.tokens[result.TokenID]; !ok {
				t.Fatal("token not persisted in repo")
			}
		})
	}
}

func TestTokenService_CreateToken_repoError(t *testing.T) {
	t.Parallel()

	svc := testTokenService(errTokenRepo{})
	_, err := svc.CreateToken(context.Background(), &authv1.CreateTokenRequest{
		OrgId: uuid.NewString(), Name: "x", Type: authv1.TokenType_TOKEN_TYPE_PAT,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTokenService_RevokeToken(t *testing.T) {
	t.Parallel()

	orgID := uuid.New().String()
	tokenID := uuid.New().String()
	repo := newMemTokenRepo()
	repo.tokens[tokenID] = repository.CreateTokenParams{ID: tokenID, OrgID: orgID}
	svc := testTokenService(repo)

	err := svc.RevokeToken(context.Background(), orgID, tokenID, "", nil)
	if err != nil {
		t.Fatalf("RevokeToken: %v", err)
	}
	if !repo.revoked[tokenID] {
		t.Fatal("token not marked revoked")
	}

	err = svc.RevokeToken(context.Background(), orgID, uuid.NewString(), "", nil)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTokenService_ListTokens(t *testing.T) {
	t.Parallel()

	orgID := uuid.New().String()
	expires := time.Now().UTC().Add(time.Hour)
	revokedAt := time.Now().UTC().Add(-time.Minute)
	repo := newMemTokenRepo()
	repo.list = []repository.TokenMetadata{
		{
			ID: "t1", Name: "a", Prefix: "ibex_pat_a", Permissions: 1,
			CreatedAt: time.Now().UTC(),
			ExpiresAt: sql.NullTime{Time: expires, Valid: true},
		},
		{
			ID: "t2", Name: "b", Prefix: "ibex_pat_b", Permissions: 2,
			CreatedAt: time.Now().UTC(), IsRevoked: true,
			RevokedAt: sql.NullTime{Time: revokedAt, Valid: true},
		},
	}
	svc := testTokenService(repo)

	rows, next, err := svc.ListTokens(context.Background(), orgID, "", 10)
	if err != nil {
		t.Fatalf("ListTokens: %v", err)
	}
	if len(rows) != 2 || next != "" {
		t.Fatalf("list: len=%d next=%q", len(rows), next)
	}
}

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
