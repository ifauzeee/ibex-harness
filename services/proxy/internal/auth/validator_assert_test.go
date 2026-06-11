package auth

import (
	"errors"
	"testing"
)

func assertValidatorError(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
}

func assertValidatorResult(t *testing.T, got, want *ValidateResult) {
	t.Helper()
	if got.OrgID != want.OrgID || got.Permissions != want.Permissions {
		t.Fatalf("result: %+v, want %+v", got, want)
	}
	if got.AgentID != want.AgentID || got.UserID != want.UserID || got.TokenID != want.TokenID {
		t.Fatalf("optional fields: %+v, want %+v", got, want)
	}
}
