package user

import (
	"testing"
)

func TestNewUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "john_doe", false},
		{"too short", "ab", true},
		{"too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUsername(tt.input)

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "test@mail.com", false},
		{"invalid no @", "testmail.com", true},
		{"invalid empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEmail(tt.input)

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
