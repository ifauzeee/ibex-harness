package token_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
	"github.com/Rick1330/ibex-harness/services/auth/internal/token"
	"github.com/google/uuid"
)

func assertValidatorError(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("err: got %v want %v", err, want)
	}
}

func assertValidatorOK(t *testing.T, resp *authv1.ValidateTokenResponse, agentID, userID string) {
	t.Helper()
	if resp.GetOrgId() == "" || resp.GetPermissions() != 42 {
		t.Fatalf("resp: %+v", resp)
	}
	if resp.GetAgentId() != agentID || resp.GetUserId() != userID {
		t.Fatalf("optional fields missing")
	}
}

func runValidatorCase(t *testing.T, argon2 token.Argon2Params, tc validatorCase, agentID, userID string) {
	t.Helper()
	resp, err := token.NewValidator(tc.lookup, argon2).Validate(context.Background(), tc.token)
	if tc.wantErr != nil {
		assertValidatorError(t, err, tc.wantErr)
		return
	}
	if tc.expect == "db error" {
		if err == nil {
			t.Fatal("expected error")
		}
		return
	}
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	assertValidatorOK(t, resp, agentID, userID)
}

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
	for _, tc := range validatorCases(validatorFixture{bearer: bearer, hash: hash, agentID: agentID, userID: userID, row: row}) {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runValidatorCase(t, argon2, tc, agentID, userID)
		})
	}
}
