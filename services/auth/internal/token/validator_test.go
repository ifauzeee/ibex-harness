package token_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

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

type validatorRun struct {
	argon2          token.Argon2Params
	tc              validatorCase
	agentID, userID string
}

func runValidatorCase(t *testing.T, run validatorRun) {
	t.Helper()
	resp, err := token.NewValidator(run.tc.lookup, run.argon2).Validate(context.Background(), run.tc.token)
	if run.tc.wantErr != nil {
		assertValidatorError(t, err, run.tc.wantErr)
		return
	}
	if run.tc.expect == "db error" {
		if err == nil {
			t.Fatal("expected error")
		}
		return
	}
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if resp.GetOrgId() == "" {
		t.Fatalf("resp: %+v", resp)
	}
	if resp.GetPermissions() != 42 {
		t.Fatalf("perms: %d", resp.GetPermissions())
	}
	if resp.GetAgentId() != run.agentID {
		t.Fatal("agent id missing")
	}
	if resp.GetUserId() != run.userID {
		t.Fatal("user id missing")
	}
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
			runValidatorCase(t, validatorRun{argon2: argon2, tc: tc, agentID: agentID, userID: userID})
		})
	}
}
