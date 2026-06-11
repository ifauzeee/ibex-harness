package service

import "context"

func (m *memTokenRepo) RevokeToken(_ context.Context, orgID, tokenID, revokedBy string, reason *string) error {
	return m.revoke(revokeTokenInput{orgID: orgID, tokenID: tokenID, revokedBy: revokedBy, reason: reason})
}
