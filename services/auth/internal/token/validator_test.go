package token_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
	"github.com/Rick1330/ibex-harness/services/auth/internal/token"
	"github.com/google/uuid"
)

type fakeLookup struct {
	row repository.TokenRow
	err error
}

func (f *fakeLookup) FindActiveByPrefix(ctx context.Context, _ string) (repository.TokenRow, error) {
	if f.err != nil {
		return repository.TokenRow{}, f.err
	}
	return f.row, nil
}

func TestValidator_Validate(t *testing.T) {
	t.Parallel()
	argon2 := token.DefaultArgon2Params()
	tokenID := uuid.New()
	bearer := "ibex_pat_" + tokenID.String() + "_secret"
	hash, err := token.HashForTest(bearer, argon2)
	if err != nil {
		t.Fatal(err)
	}
	agentID := uuid.NewString()
	userID := uuid.NewString()
	row := repository.TokenRow{
		ID: tokenID.String(), OrgID: uuid.NewString(), Hash: hash, Permissions: 42,
		AgentID:   sql.NullString{String: agentID, Valid: true},
		UserID:    sql.NullString{String: userID, Valid: true},
		ExpiresAt: sql.NullTime{Time: time.Now().UTC().Add(time.Hour), Valid: true},
	}
	for _, tc := range validatorCases(bearer, hash, agentID, userID, row) {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runValidatorCase(t, argon2, tc, agentID, userID)
		})
	}
}
