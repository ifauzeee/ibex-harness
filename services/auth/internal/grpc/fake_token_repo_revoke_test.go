package grpcserver

import "context"

func (f *fakeTokenRepo) RevokeToken(ctx context.Context, orgID, tokenID, revokedBy string, reason *string) error {
	if f.revokeFn != nil {
		return f.revokeFn(ctx, revokeTokenInput{
			orgID: orgID, tokenID: tokenID, revokedBy: revokedBy, reason: reason,
		})
	}
	return nil
}
