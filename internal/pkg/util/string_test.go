package util

import "testing"

func TestByteCountIEC(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{999, "999 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{1024*1024 + 512*1024, "1.5 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
	}

	for _, tt := range tests {
		got := ByteCountIEC(tt.in)
		if got != tt.want {
			t.Fatalf("ByteCountIEC(%d) = %q; want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{12, "12"},
		{999, "999"},
		{1000, "1 000"},
		{1234, "1 234"},
		{1234567, "1 234 567"},
		{-9876543, "-9 876 543"},
		{2147483647, "2 147 483 647"},
		{-2147483647, "-2 147 483 647"},
	}

	for _, tt := range tests {
		got := FormatInt(tt.in)
		if got != tt.want {
			t.Fatalf("FormatInt(%d) = %q; want %q", tt.in, got, tt.want)
		}
	}
}
