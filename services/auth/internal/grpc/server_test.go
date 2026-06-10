package grpcserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Rick1330/ibex-harness/packages/logger"
	"github.com/Rick1330/ibex-harness/packages/permissions"
	authv1 "github.com/Rick1330/ibex-harness/packages/proto/gen/go/ibex/auth/v1"
	"github.com/Rick1330/ibex-harness/services/auth/internal/repository"
	"github.com/Rick1330/ibex-harness/services/auth/internal/service"
	"github.com/Rick1330/ibex-harness/services/auth/internal/token"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeTokenValidator struct {
	fn func(context.Context, string) (*authv1.ValidateTokenResponse, error)
}

func (f *fakeTokenValidator) Validate(ctx context.Context, accessToken string) (*authv1.ValidateTokenResponse, error) {
	return f.fn(ctx, accessToken)
}

type fakeTokenRepo struct {
	createFn func(context.Context, repository.CreateTokenParams) (string, error)
	revokeFn func(context.Context, string, string, string, *string) error
	listFn   func(context.Context, string, string, int) ([]repository.TokenMetadata, string, error)
}

func (f *fakeTokenRepo) CreateToken(ctx context.Context, p repository.CreateTokenParams) (string, error) {
	if f.createFn != nil {
		return f.createFn(ctx, p)
	}
	return p.ID, nil
}

func (f *fakeTokenRepo) RevokeToken(ctx context.Context, orgID, tokenID, revokedBy string, reason *string) error {
	if f.revokeFn != nil {
		return f.revokeFn(ctx, orgID, tokenID, revokedBy, reason)
	}
	return nil
}

func (f *fakeTokenRepo) ListTokens(ctx context.Context, orgID, cursor string, limit int) ([]repository.TokenMetadata, string, error) {
	if f.listFn != nil {
		return f.listFn(ctx, orgID, cursor, limit)
	}
	return nil, "", nil
}

func newTestServer(
	validator tokenValidator,
	tokenRepo serviceTokenRepo,
	agents AgentStore,
) *Server {
	tokenSvc := service.NewTokenService(tokenRepo, token.DefaultArgon2Params(), logger.Discard("auth"))
	return NewServer(validator, tokenSvc, agents, testAuthRegistry())
}

// serviceTokenRepo mirrors service.tokenRepo for test wiring.
type serviceTokenRepo interface {
	CreateToken(ctx context.Context, p repository.CreateTokenParams) (string, error)
	RevokeToken(ctx context.Context, orgID, tokenID, revokedBy string, reason *string) error
	ListTokens(ctx context.Context, orgID, cursor string, limit int) ([]repository.TokenMetadata, string, error)
}

func adminCtx(t *testing.T, orgID string) context.Context {
	t.Helper()
	return ContextWithCaller(context.Background(), CallerContext{
		OrgID:       orgID,
		TokenID:     uuid.NewString(),
		UserID:      uuid.NewString(),
		Permissions: permissions.Admin,
	})
}

func TestServer_ValidateToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fn       func(context.Context, string) (*authv1.ValidateTokenResponse, error)
		wantCode codes.Code
	}{
		{
			name: "unauthenticated",
			fn: func(context.Context, string) (*authv1.ValidateTokenResponse, error) {
				return nil, token.ErrUnauthenticated
			},
			wantCode: codes.Unauthenticated,
		},
		{
			name: "internal error",
			fn: func(context.Context, string) (*authv1.ValidateTokenResponse, error) {
				return nil, errors.New("db down")
			},
			wantCode: codes.Internal,
		},
		{
			name: "ok",
			fn: func(context.Context, string) (*authv1.ValidateTokenResponse, error) {
				return &authv1.ValidateTokenResponse{OrgId: "org-1", Permissions: 7}, nil
			},
			wantCode: codes.OK,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := newTestServer(&fakeTokenValidator{fn: tc.fn}, &fakeTokenRepo{}, &fakeAgentsStore{})

			resp, err := s.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{AccessToken: "tok"})

			if tc.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("ValidateToken: %v", err)
				}
				if resp.GetOrgId() != "org-1" || resp.GetPermissions() != 7 {
					t.Fatalf("response: %+v", resp)
				}
				return
			}
			if status.Code(err) != tc.wantCode {
				t.Fatalf("code: got %v want %v", status.Code(err), tc.wantCode)
			}
		})
	}
}

func TestServer_CreateToken(t *testing.T) {
	t.Parallel()

	orgID := uuid.NewString()

	tests := []struct {
		name     string
		ctx      context.Context
		req      *authv1.CreateTokenRequest
		wantCode codes.Code
	}{
		{
			name:     "unauthenticated",
			ctx:      context.Background(),
			req:      &authv1.CreateTokenRequest{OrgId: orgID, Name: "x"},
			wantCode: codes.Unauthenticated,
		},
		{
			name: "permission denied wrong org",
			ctx: ContextWithCaller(context.Background(), CallerContext{
				OrgID: uuid.NewString(), Permissions: permissions.Admin,
			}),
			req:      &authv1.CreateTokenRequest{OrgId: orgID, Name: "x"},
			wantCode: codes.PermissionDenied,
		},
		{
			name:     "invalid argument",
			ctx:      adminCtx(t, orgID),
			req:      &authv1.CreateTokenRequest{OrgId: orgID},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "ok",
			ctx:  adminCtx(t, orgID),
			req: &authv1.CreateTokenRequest{
				OrgId: orgID, Name: "pat", Permissions: permissions.AgentDefault,
			},
			wantCode: codes.OK,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := newTestServer(
				&fakeTokenValidator{},
				&fakeTokenRepo{},
				&fakeAgentsStore{},
			)

			resp, err := s.CreateToken(tc.ctx, tc.req)

			if tc.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("CreateToken: %v", err)
				}
				if resp.GetTokenId() == "" || resp.GetPlaintext() == "" {
					t.Fatalf("incomplete response: %+v", resp)
				}
				return
			}
			if status.Code(err) != tc.wantCode {
				t.Fatalf("code: got %v want %v err=%v", status.Code(err), tc.wantCode, err)
			}
		})
	}
}

func TestServer_RevokeToken(t *testing.T) {
	t.Parallel()

	orgID := uuid.NewString()
	tokenID := uuid.NewString()
	selfTokenID := uuid.NewString()

	tests := []struct {
		name     string
		ctx      context.Context
		req      *authv1.RevokeTokenRequest
		repo     *fakeTokenRepo
		wantCode codes.Code
	}{
		{
			name:     "unauthenticated",
			ctx:      context.Background(),
			req:      &authv1.RevokeTokenRequest{OrgId: orgID, TokenId: tokenID},
			wantCode: codes.Unauthenticated,
		},
		{
			name: "cross tenant not found",
			ctx: ContextWithCaller(context.Background(), CallerContext{
				OrgID: uuid.NewString(), Permissions: permissions.Admin,
			}),
			req:      &authv1.RevokeTokenRequest{OrgId: orgID, TokenId: tokenID},
			wantCode: codes.NotFound,
		},
		{
			name: "permission denied",
			ctx: ContextWithCaller(context.Background(), CallerContext{
				OrgID: orgID, TokenID: uuid.NewString(), Permissions: permissions.ReadOnly,
			}),
			req:      &authv1.RevokeTokenRequest{OrgId: orgID, TokenId: tokenID},
			wantCode: codes.PermissionDenied,
		},
		{
			name: "not found in repo",
			ctx:  adminCtx(t, orgID),
			req:  &authv1.RevokeTokenRequest{OrgId: orgID, TokenId: tokenID},
			repo: &fakeTokenRepo{
				revokeFn: func(context.Context, string, string, string, *string) error {
					return repository.ErrNotFound
				},
			},
			wantCode: codes.NotFound,
		},
		{
			name: "internal error",
			ctx:  adminCtx(t, orgID),
			req:  &authv1.RevokeTokenRequest{OrgId: orgID, TokenId: tokenID},
			repo: &fakeTokenRepo{
				revokeFn: func(context.Context, string, string, string, *string) error {
					return errors.New("db down")
				},
			},
			wantCode: codes.Internal,
		},
		{
			name: "self revoke ok",
			ctx: ContextWithCaller(context.Background(), CallerContext{
				OrgID: orgID, TokenID: selfTokenID, Permissions: permissions.ReadOnly,
			}),
			req:      &authv1.RevokeTokenRequest{OrgId: orgID, TokenId: selfTokenID},
			repo:     &fakeTokenRepo{},
			wantCode: codes.OK,
		},
		{
			name:     "admin revoke ok",
			ctx:      adminCtx(t, orgID),
			req:      &authv1.RevokeTokenRequest{OrgId: orgID, TokenId: tokenID},
			repo:     &fakeTokenRepo{},
			wantCode: codes.OK,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := tc.repo
			if repo == nil {
				repo = &fakeTokenRepo{}
			}
			s := newTestServer(&fakeTokenValidator{}, repo, &fakeAgentsStore{})

			_, err := s.RevokeToken(tc.ctx, tc.req)

			if tc.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("RevokeToken: %v", err)
				}
				return
			}
			if status.Code(err) != tc.wantCode {
				t.Fatalf("code: got %v want %v err=%v", status.Code(err), tc.wantCode, err)
			}
		})
	}
}

func TestServer_ListTokens(t *testing.T) {
	t.Parallel()

	orgID := uuid.NewString()
	created := time.Now().UTC()

	tests := []struct {
		name     string
		ctx      context.Context
		repo     *fakeTokenRepo
		wantCode codes.Code
		wantLen  int
	}{
		{
			name:     "unauthenticated",
			ctx:      context.Background(),
			wantCode: codes.Unauthenticated,
		},
		{
			name: "permission denied",
			ctx: ContextWithCaller(context.Background(), CallerContext{
				OrgID: uuid.NewString(), Permissions: permissions.Admin,
			}),
			wantCode: codes.PermissionDenied,
		},
		{
			name: "internal error",
			ctx:  adminCtx(t, orgID),
			repo: &fakeTokenRepo{
				listFn: func(context.Context, string, string, int) ([]repository.TokenMetadata, string, error) {
					return nil, "", errors.New("db down")
				},
			},
			wantCode: codes.Internal,
		},
		{
			name: "ok",
			ctx:  adminCtx(t, orgID),
			repo: &fakeTokenRepo{
				listFn: func(context.Context, string, string, int) ([]repository.TokenMetadata, string, error) {
					return []repository.TokenMetadata{
						{ID: "t1", Name: "a", Prefix: "p1", Permissions: 1, CreatedAt: created},
					}, "next-cursor", nil
				},
			},
			wantCode: codes.OK,
			wantLen:  1,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := tc.repo
			if repo == nil {
				repo = &fakeTokenRepo{}
			}
			s := newTestServer(&fakeTokenValidator{}, repo, &fakeAgentsStore{})

			resp, err := s.ListTokens(tc.ctx, &authv1.ListTokensRequest{OrgId: orgID, Limit: 10})

			if tc.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("ListTokens: %v", err)
				}
				if len(resp.GetTokens()) != tc.wantLen {
					t.Fatalf("tokens len: got %d want %d", len(resp.GetTokens()), tc.wantLen)
				}
				if tc.wantLen > 0 && resp.GetNextCursor() != "next-cursor" {
					t.Fatalf("next cursor: %q", resp.GetNextCursor())
				}
				return
			}
			if status.Code(err) != tc.wantCode {
				t.Fatalf("code: got %v want %v", status.Code(err), tc.wantCode)
			}
		})
	}
}
