package service

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
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

func tokenMissingForRevoke(p repository.CreateTokenParams, ok bool, orgID, tokenID string, revoked map[string]bool) bool {
	if !ok {
		return true
	}
	if p.OrgID != orgID {
		return true
	}
	return revoked[tokenID]
}

func (m *memTokenRepo) revoke(in revokeTokenInput) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.tokens[in.tokenID]
	if tokenMissingForRevoke(p, ok, in.orgID, in.tokenID, m.revoked) {
		return repository.ErrNotFound
	}
	m.revoked[in.tokenID] = true
	return nil
}

func (m *memTokenRepo) ListTokens(ctx context.Context, orgID, cursor string, limit int) ([]repository.TokenMetadata, string, error) {
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
	sortTokenRows(rows)
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
