package crypto

import "testing"

func TestConstantTimeEqualBytes(t *testing.T) {
	t.Parallel()

	if !ConstantTimeEqualBytes([]byte("abc"), []byte("abc")) {
		t.Fatal("expected equal byte slices to match")
	}
	if ConstantTimeEqualBytes([]byte("abc"), []byte("abd")) {
		t.Fatal("expected different byte slices to differ")
	}
	if ConstantTimeEqualBytes([]byte("a"), []byte("ab")) {
		t.Fatal("expected different lengths to differ")
	}
}
