package subscription

import (
	"errors"
	"testing"
	"time"
)

func TestParseMonthYear(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    time.Time
		wantErr error
	}{
		{
			name:  "valid month year",
			value: "07-2025",
			want:  time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "empty value",
			value:   "",
			wantErr: ErrInvalidDate,
		},
		{
			name:    "invalid separator",
			value:   "07/2025",
			wantErr: ErrInvalidDate,
		},
		{
			name:    "invalid month",
			value:   "13-2025",
			wantErr: ErrInvalidDate,
		},
		{
			name:    "zero month",
			value:   "00-2025",
			wantErr: ErrInvalidDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMonthYear(tt.value)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ParseMonthYear(%q) error = %v, want %v", tt.value, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseMonthYear(%q) error = %v", tt.value, err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("ParseMonthYear(%q) = %s, want %s", tt.value, got, tt.want)
			}
		})
	}
}

func TestFormatMonthYear(t *testing.T) {
	value := time.Date(2025, time.July, 20, 14, 30, 0, 0, time.Local)

	got := FormatMonthYear(value)

	if got != "07-2025" {
		t.Fatalf("FormatMonthYear() = %q, want %q", got, "07-2025")
	}
}
