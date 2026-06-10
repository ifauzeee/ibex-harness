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

type fakeLookup struct {
	row repository.TokenRow
	err error
}

func (f *fakeLookup) FindActiveByPrefix(_ context.Context, _ string) (repository.TokenRow, error) {
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
	exp := time.Now().UTC().Add(time.Hour)
	agentID := uuid.NewString()
	userID := uuid.NewString()

	tests := []struct {
		name    string
		token   string
		lookup  *fakeLookup
		wantErr error
	}{
		{
			name:    "malformed token",
			token:   "not-a-token",
			lookup:  &fakeLookup{},
			wantErr: token.ErrUnauthenticated,
		},
		{
			name:    "not found",
			token:   bearer,
			lookup:  &fakeLookup{err: sql.ErrNoRows},
			wantErr: token.ErrUnauthenticated,
		},
		{
			name:    "wrong hash",
			token:   bearer,
			lookup:  &fakeLookup{row: repository.TokenRow{Hash: "wrong", OrgID: uuid.NewString()}},
			wantErr: token.ErrUnauthenticated,
		},
		{
			name:    "db error",
			token:   bearer,
			lookup:  &fakeLookup{err: errors.New("db down")},
			wantErr: nil,
		},
		{
			name:  "ok with optional fields",
			token: bearer,
			lookup: &fakeLookup{row: repository.TokenRow{
				ID: tokenID.String(), OrgID: uuid.NewString(), Hash: hash, Permissions: 42,
				AgentID:   sql.NullString{String: agentID, Valid: true},
				UserID:    sql.NullString{String: userID, Valid: true},
				ExpiresAt: sql.NullTime{Time: exp, Valid: true},
			}},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			v := token.NewValidator(tc.lookup, argon2)
			resp, err := v.Validate(context.Background(), tc.token)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err: got %v want %v", err, tc.wantErr)
				}
				return
			}
			if tc.name == "db error" {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate: %v", err)
			}
			if resp.GetOrgId() == "" || resp.GetPermissions() != 42 {
				t.Fatalf("resp: %+v", resp)
			}
			if resp.GetAgentId() != agentID || resp.GetUserId() != userID {
				t.Fatalf("optional fields missing")
			}
		})
	}
}
