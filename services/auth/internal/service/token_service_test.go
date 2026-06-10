package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Rick1330/ibex-harness/packages/logger"
	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
	"github.com/Rick1330/ibex-harness/services/auth/internal/token"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type memTokenRepo struct {
	mu      sync.Mutex
	tokens  map[string]repository.CreateTokenParams
	revoked map[string]bool
	list    []repository.TokenMetadata
}

func newMemTokenRepo() *memTokenRepo {
	return &memTokenRepo{
		tokens:  make(map[string]repository.CreateTokenParams),
		revoked: make(map[string]bool),
	}
}

func (m *memTokenRepo) CreateToken(_ context.Context, p repository.CreateTokenParams) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[p.ID] = p
	return p.ID, nil
}

func (m *memTokenRepo) RevokeToken(_ context.Context, orgID, tokenID, _ string, _ *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.tokens[tokenID]
	if !ok || p.OrgID != orgID || m.revoked[tokenID] {
		return repository.ErrNotFound
	}
	m.revoked[tokenID] = true
	return nil
}

func (m *memTokenRepo) ListTokens(_ context.Context, orgID, cursor string, limit int) ([]repository.TokenMetadata, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.list) > 0 {
		return m.list, "", nil
	}
	var rows []repository.TokenMetadata
	for id, p := range m.tokens {
		if p.OrgID != orgID {
			continue
		}
		meta := repository.TokenMetadata{
			ID:          id,
			Name:        p.Name,
			Prefix:      p.Prefix,
			Permissions: p.Permissions,
			CreatedAt:   time.Now().UTC(),
			IsRevoked:   m.revoked[id],
		}
		if p.ExpiresAt != nil {
			meta.ExpiresAt = sql.NullTime{Time: *p.ExpiresAt, Valid: true}
		}
		if m.revoked[id] {
			meta.RevokedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
		}
		rows = append(rows, meta)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].CreatedAt.Equal(rows[j].CreatedAt) {
			return rows[i].ID > rows[j].ID
		}
		return rows[i].CreatedAt.After(rows[j].CreatedAt)
	})
	if cursor != "" {
		cursorTS, cursorID, err := decodeMemTokenCursor(cursor)
		if err != nil {
			return nil, "", err
		}
		filtered := rows[:0]
		for _, row := range rows {
			if row.CreatedAt.Before(cursorTS) || (row.CreatedAt.Equal(cursorTS) && row.ID < cursorID) {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	if limit <= 0 {
		limit = 50
	}
	var next string
	if len(rows) > limit {
		last := rows[limit-1]
		next = encodeMemTokenCursor(last.CreatedAt, last.ID)
		rows = rows[:limit]
	}
	return rows, next, nil
}

func encodeMemTokenCursor(createdAt time.Time, id string) string {
	return fmt.Sprintf("%d|%s", createdAt.UTC().UnixNano(), id)
}

func decodeMemTokenCursor(cursor string) (time.Time, string, error) {
	parts := strings.SplitN(cursor, "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("invalid cursor %q", cursor)
	}
	var nano int64
	if _, err := fmt.Sscanf(parts[0], "%d", &nano); err != nil {
		return time.Time{}, "", err
	}
	return time.Unix(0, nano).UTC(), parts[1], nil
}

func testTokenService(repo tokenRepo) *TokenService {
	return NewTokenService(repo, token.DefaultArgon2Params(), logger.Discard("auth"))
}

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

type errTokenRepo struct{}

func (errTokenRepo) CreateToken(context.Context, repository.CreateTokenParams) (string, error) {
	return "", errors.New("db down")
}

func (errTokenRepo) RevokeToken(context.Context, string, string, string, *string) error {
	return errors.New("db down")
}

func (errTokenRepo) ListTokens(context.Context, string, string, int) ([]repository.TokenMetadata, string, error) {
	return nil, "", errors.New("db down")
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

func TestToProtoList_AllFields(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	expires := created.Add(24 * time.Hour)
	revoked := created.Add(time.Hour)

	rows := []repository.TokenMetadata{
		{
			ID: "tid", Name: "n", Prefix: "ibex_pat_x", Permissions: 99,
			CreatedAt: created,
			ExpiresAt: sql.NullTime{Time: expires, Valid: true},
			RevokedAt: sql.NullTime{Time: revoked, Valid: true},
			IsRevoked: true,
		},
	}

	out := ToProtoList(rows)
	if len(out) != 1 {
		t.Fatalf("len: %d", len(out))
	}
	m := out[0]
	if m.GetTokenId() != "tid" || m.GetName() != "n" || m.GetPrefix() != "ibex_pat_x" || m.GetPermissions() != 99 {
		t.Fatalf("metadata fields: %+v", m)
	}
	if !m.GetIsRevoked() {
		t.Fatal("expected is_revoked true")
	}
	if m.GetCreatedAt().AsTime() != created {
		t.Fatalf("created_at: %v", m.GetCreatedAt().AsTime())
	}
	if m.GetExpiresAt().AsTime() != expires {
		t.Fatalf("expires_at: %v", m.GetExpiresAt().AsTime())
	}
	if m.GetRevokedAt().AsTime() != revoked {
		t.Fatalf("revoked_at: %v", m.GetRevokedAt().AsTime())
	}
}
