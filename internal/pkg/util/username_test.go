package util_test

import (
	"errors"
	"testing"

	"github.com/Rizwank123/chatApp/internal/pkg/util"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError error
	}{
		{
			name:        "valid simple username",
			input:       "rizwan",
			expected:    "rizwan",
			expectError: nil,
		},
		{
			name:        "valid username with mixed case and spaces",
			input:       "  Rizwan_99  ",
			expected:    "rizwan_99",
			expectError: nil,
		},
		{
			name:        "valid username with dot",
			input:       "john.doe",
			expected:    "john.doe",
			expectError: nil,
		},
		{
			name:        "too short",
			input:       "ab",
			expectError: util.ErrUsernameTooShort,
		},
		{
			name:        "too long",
			input:       "abcdefghijklmnopqrstuvwxyz012345",
			expectError: util.ErrUsernameTooLong,
		},
		{
			name:        "invalid leading underscore",
			input:       "_rizwan",
			expectError: util.ErrUsernameInvalidEdge,
		},
		{
			name:        "invalid trailing dot",
			input:       "rizwan.",
			expectError: util.ErrUsernameInvalidEdge,
		},
		{
			name:        "consecutive dots",
			input:       "riz..wan",
			expectError: util.ErrUsernameConsecutive,
		},
		{
			name:        "invalid characters",
			input:       "rizwan@123",
			expectError: util.ErrUsernameInvalidChars,
		},
		{
			name:        "reserved username admin",
			input:       "ADMIN",
			expectError: util.ErrUsernameReserved,
		},
		{
			name:        "reserved username support",
			input:       "support",
			expectError: util.ErrUsernameReserved,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			normalized, err := util.ValidateUsername(tc.input)
			if tc.expectError != nil {
				if err == nil {
					t.Fatalf("expected error containing %v, got nil", tc.expectError)
				}
				if !errors.Is(err, tc.expectError) {
					t.Fatalf("expected error %v, got %v", tc.expectError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if normalized != tc.expected {
					t.Fatalf("expected normalized username '%s', got '%s'", tc.expected, normalized)
				}
			}
		})
	}
}
