package security_test

import (
	"testing"

	"github.com/robertd2000/go-image-processing-app/auth/internal/infrastructure/security"
)

func TestSHA1Hasher_Check(t *testing.T) {
	tests := []struct {
		name  string
		salt  string
		plain string
		hash  string
		want  bool
	}{
		{
			name:  "valid password",
			salt:  "mysalt",
			plain: "password123",
			hash: func() string {
				h := security.NewSHA1Hasher("mysalt")
				v, _ := h.Hash("password123")
				return v
			}(),
			want: true,
		},
		{
			name:  "invalid password",
			salt:  "mysalt",
			plain: "wrongpassword",
			hash: func() string {
				h := security.NewSHA1Hasher("mysalt")
				v, _ := h.Hash("password123")
				return v
			}(),
			want: false,
		},
		{
			name:  "invalid salt",
			salt:  "othersalt",
			plain: "password123",
			hash: func() string {
				h := security.NewSHA1Hasher("mysalt")
				v, _ := h.Hash("password123")
				return v
			}(),
			want: false,
		},
		{
			name:  "empty password",
			salt:  "mysalt",
			plain: "",
			hash: func() string {
				h := security.NewSHA1Hasher("mysalt")
				v, _ := h.Hash("")
				return v
			}(),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := security.NewSHA1Hasher(tt.salt)
			got := h.Compare(tt.plain, tt.hash)

			if got != tt.want {
				t.Errorf("Check() = %v, want %v", got, tt.want)
			}
		})
	}
}
