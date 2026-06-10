package auth

import (
	"context"
	"testing"
)

func TestAuthContext_roundTrip(t *testing.T) {
	t.Parallel()

	res := &ValidateResult{OrgID: "org-a", Permissions: 99}
	ctx := WithContext(context.Background(), res)

	got, ok := FromContext(ctx)
	if !ok {
		t.Fatal("expected auth context")
	}
	if got.OrgID != res.OrgID || got.Permissions != res.Permissions {
		t.Fatalf("got %+v, want %+v", got, res)
	}
}

func TestAuthContext_missing(t *testing.T) {
	t.Parallel()

	_, ok := FromContext(context.Background())
	if ok {
		t.Fatal("expected false without auth context")
	}
}

func TestAuthContext_nilValue(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), contextKey{}, (*ValidateResult)(nil))
	_, ok := FromContext(ctx)
	if ok {
		t.Fatal("expected false for nil auth result")
	}
}
