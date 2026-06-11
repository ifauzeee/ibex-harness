package config

import "testing"

func TestListenAddress(t *testing.T) {
	t.Parallel()

	if got := ListenAddress("8080"); got != ":8080" {
		t.Fatalf("ListenAddress: %q", got)
	}
}
