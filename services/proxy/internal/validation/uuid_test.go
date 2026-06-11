package validation

import "testing"

func runUUIDFieldCase(t *testing.T, tc uuidFieldCase) {
	t.Helper()
	field := tc.wantField
	if field == "" {
		field = "agent_id"
	}
	got := ValidateUUIDField(field, tc.value)
	if tc.wantNil {
		if got != nil {
			t.Fatalf("expected nil, got %+v", got)
		}
		return
	}
	if got == nil {
		t.Fatal("expected field error")
	}
	if got.Field != tc.wantField || got.Code != tc.wantCode {
		t.Fatalf("got %+v want field=%s code=%s", got, tc.wantField, tc.wantCode)
	}
}

func TestValidateUUIDField(t *testing.T) {
	t.Parallel()
	for _, tc := range uuidFieldCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runUUIDFieldCase(t, tc)
		})
	}
}
