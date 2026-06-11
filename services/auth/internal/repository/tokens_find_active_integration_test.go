//go:build integration

package repository_test

import "testing"

func findActiveCases() []findActiveCase {
	return []findActiveCase{
		{name: "happy path", wantFound: true},
		{name: "revoked excluded", revoked: true, wantFound: false},
		{name: "expired excluded", expired: true, wantFound: false},
	}
}

func TestTokensRepository_FindActiveByPrefix(t *testing.T) {
	for _, tc := range findActiveCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runFindActiveCase(t, tc)
		})
	}
}
