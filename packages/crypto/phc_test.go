package crypto

import (
	"strings"
	"testing"
)

func TestPHC_formatAndParseRoundTrip(t *testing.T) {
	t.Parallel()

	p := TestParams()
	hash, err := HashSecret("vector-password", p)
	if err != nil {
		t.Fatalf("HashSecret: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Fatalf("prefix: %q", hash[:min(20, len(hash))])
	}

	ok, err := VerifySecret("vector-password", hash, p)
	if err != nil || !ok {
		t.Fatalf("verify: ok=%v err=%v", ok, err)
	}
}

func TestPHC_parseErrors(t *testing.T) {
	t.Parallel()

	cases := []string{
		"not-a-phc",
		"$argon2i$v=19$m=1,t=1,p=1$abc$def",
		"$argon2id$v=18$m=1,t=1,p=1$abc$def",
		"$argon2id$v=19$badparam$abc$def",
		"$argon2id$v=19$m=1,t=1,p=1$!!!$def",
	}
	for _, phc := range cases {
		if _, _, _, _, _, err := parsePHC(phc); err == nil {
			t.Fatalf("expected error for %q", phc)
		}
	}
}

func TestHashPasswordAndVerifyPassword(t *testing.T) {
	t.Parallel()

	p := TestParams()
	hash, err := HashPassword("user-password", p)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	ok, err := VerifyPassword("user-password", hash, p)
	if err != nil || !ok {
		t.Fatalf("VerifyPassword: ok=%v err=%v", ok, err)
	}
	ok, err = VerifyPassword("wrong", hash, p)
	if err != nil || ok {
		t.Fatalf("wrong password: ok=%v err=%v", ok, err)
	}
}

func TestPHC_usesFallbackParamsWhenZero(t *testing.T) {
	t.Parallel()

	p := TestParams()
	hash, err := HashSecret("pw", p)
	if err != nil {
		t.Fatal(err)
	}
	mem, time, par, salt, digest, err := parsePHC(hash)
	if err != nil {
		t.Fatal(err)
	}
	crafted := formatPHC(Argon2Params{}, salt, digest)
	if !strings.HasPrefix(crafted, "$argon2id$") {
		t.Fatalf("crafted: %q", crafted)
	}
	_ = mem
	_ = time
	_ = par
	ok, err := VerifySecret("pw", crafted, p)
	if err != nil || !ok {
		t.Fatalf("fallback verify: ok=%v err=%v", ok, err)
	}
}
