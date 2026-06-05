package idgen

import (
	"testing"
)

func TestParseID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"valid", "1234567890123456789", 1234567890123456789, false},
		{"empty", "", 0, true},
		{"zero", "0", 0, true},
		{"negative", "-1", 0, true},
		{"non_decimal", "12a", 0, true},
		{"leading_space", " 1", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseID(%q) err = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("ParseID(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatID(t *testing.T) {
	t.Parallel()
	if got := FormatID(9223372036854775807); got != "9223372036854775807" {
		t.Fatalf("FormatID = %q, want max int64 string", got)
	}
}
