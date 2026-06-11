package service

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Rick1330/ibex-harness/packages/logger"
	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
	"github.com/Rick1330/ibex-harness/services/auth/internal/token"
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

type revokeTokenInput struct {
	orgID     string
	tokenID   string
	revokedBy string
	reason    *string
}

func (m *memTokenRepo) RevokeToken(_ context.Context, orgID, tokenID, revokedBy string, reason *string) error {
	return m.revoke(revokeTokenInput{orgID: orgID, tokenID: tokenID, revokedBy: revokedBy, reason: reason})
}

func (m *memTokenRepo) revoke(in revokeTokenInput) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.tokens[in.tokenID]
	if !ok || p.OrgID != in.orgID || m.revoked[in.tokenID] {
		return repository.ErrNotFound
	}
	m.revoked[in.tokenID] = true
	return nil
}

func (m *memTokenRepo) ListTokens(_ context.Context, orgID, cursor string, limit int) ([]repository.TokenMetadata, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.list) > 0 {
		return m.list, "", nil
	}
	rows := m.sortedRowsForOrg(orgID)
	var err error
	rows, err = filterTokensAfterCursor(rows, cursor)
	if err != nil {
		return nil, "", err
	}
	rows, next := paginateTokenRows(rows, limit)
	return rows, next, nil
}

func (m *memTokenRepo) sortedRowsForOrg(orgID string) []repository.TokenMetadata {
	var rows []repository.TokenMetadata
	for id, p := range m.tokens {
		if p.OrgID != orgID {
			continue
		}
		rows = append(rows, m.metadataForToken(id, p))
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].CreatedAt.Equal(rows[j].CreatedAt) {
			return rows[i].ID > rows[j].ID
		}
		return rows[i].CreatedAt.After(rows[j].CreatedAt)
	})
	return rows
}

func (m *memTokenRepo) metadataForToken(id string, p repository.CreateTokenParams) repository.TokenMetadata {
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
	return meta
}

func filterTokensAfterCursor(rows []repository.TokenMetadata, cursor string) ([]repository.TokenMetadata, error) {
	if cursor == "" {
		return rows, nil
	}
	cursorTS, cursorID, err := decodeMemTokenCursor(cursor)
	if err != nil {
		return nil, err
	}
	filtered := rows[:0]
	for _, row := range rows {
		if row.CreatedAt.Before(cursorTS) || (row.CreatedAt.Equal(cursorTS) && row.ID < cursorID) {
			filtered = append(filtered, row)
		}
	}
	return filtered, nil
}

func paginateTokenRows(rows []repository.TokenMetadata, limit int) ([]repository.TokenMetadata, string) {
	if limit <= 0 {
		limit = 50
	}
	if len(rows) <= limit {
		return rows, ""
	}
	last := rows[limit-1]
	return rows[:limit], encodeMemTokenCursor(last.CreatedAt, last.ID)
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

type errTokenRepo struct{}

func (errTokenRepo) CreateToken(context.Context, repository.CreateTokenParams) (string, error) {
	return "", fmt.Errorf("db down")
}

func (errTokenRepo) RevokeToken(context.Context, string, string, string, *string) error {
	return fmt.Errorf("db down")
}

func (errTokenRepo) ListTokens(context.Context, string, string, int) ([]repository.TokenMetadata, string, error) {
	return nil, "", fmt.Errorf("db down")
}
