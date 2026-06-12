package grpctest

import (
	"errors"
	"testing"
)

// ErrCase runs an operation expected to succeed or return a specific error.
type ErrCase struct {
	Name    string
	Run     func() error
	WantErr error
}

// RunErrCases executes table-driven error-mapping tests.
func RunErrCases(t *testing.T, cases []ErrCase) {
	t.Helper()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := tc.Run()
			if tc.WantErr != nil {
				if !errors.Is(err, tc.WantErr) {
					t.Fatalf("err = %v, want %v", err, tc.WantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
