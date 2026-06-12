package repository

import (
	"testing"
	"time"
)

func TestDecodeTokenCursor(t *testing.T) {
	t.Parallel()

	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	valid := encodeTokenCursor(ts, "tok-id")

	tests := []struct {
		name    string
		cursor  string
		wantTS  time.Time
		wantID  string
		wantErr bool
	}{
		{name: "empty", cursor: ""},
		{name: "valid", cursor: valid, wantTS: ts, wantID: "tok-id"},
		{name: "missing pipe", cursor: "123", wantErr: true},
		{name: "empty timestamp", cursor: "|id", wantErr: true},
		{name: "empty id", cursor: "123|", wantErr: true},
		{name: "bad timestamp", cursor: "abc|id", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotTS, gotID, err := decodeTokenCursor(tc.cursor)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("decodeTokenCursor: %v", err)
			}
			if !gotTS.Equal(tc.wantTS) {
				t.Fatalf("timestamp: got %v want %v", gotTS, tc.wantTS)
			}
			if gotID != tc.wantID {
				t.Fatalf("id: got %q want %q", gotID, tc.wantID)
			}
		})
	}
}
