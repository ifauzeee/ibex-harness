package validation

import (
	"testing"
)

func TestValidateUUIDField(t *testing.T) {
	t.Parallel()

	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name      string
		value     string
		wantNil   bool
		wantCode  string
		wantField string
	}{
		{name: "valid", value: validUUID, wantNil: true},
		{name: "empty", value: "", wantCode: fieldCodeRequired, wantField: "agent_id"},
		{name: "whitespace only", value: "   ", wantCode: fieldCodeRequired, wantField: "agent_id"},
		{name: "malformed", value: "not-a-uuid", wantCode: fieldCodeInvalidFormat, wantField: "agent_id"},
		{name: "too short", value: "550e8400-e29b-41d4-a716", wantCode: fieldCodeInvalidFormat, wantField: "org_id"},
		{name: "non hex", value: "gggggggg-gggg-gggg-gggg-gggggggggggg", wantCode: fieldCodeInvalidFormat, wantField: "org_id"},
		{name: "valid trimmed", value: "  " + validUUID + "  ", wantNil: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
			if got.Field != tc.wantField {
				t.Fatalf("field: %q", got.Field)
			}
			if got.Code != tc.wantCode {
				t.Fatalf("code: %q, want %q", got.Code, tc.wantCode)
			}
		})
	}
}
