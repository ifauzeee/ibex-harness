package token_test

import (
	"context"
	"errors"
	"testing"

	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"github.com/Rick1330/ibex-harness/services/auth/internal/token"
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
	v := token.NewValidator(tc.lookup, argon2)
	resp, err := v.Validate(context.Background(), tc.token)
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
