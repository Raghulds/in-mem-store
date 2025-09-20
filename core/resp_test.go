package core

import (
	"testing"
)

func TestReadLength(t *testing.T) {
	cases := []struct {
		in        string
		wantLen   int
		wantDelta int
	}{
		{"0\r\n", 0, 3},
		{"5\r\n", 5, 3},
		{"12\r\n", 12, 4},
		{"123\r\n", 123, 5},
	}

	for _, c := range cases {
		gotLen, gotDelta := readLength([]byte(c.in))
		if gotLen != c.wantLen || gotDelta != c.wantDelta {
			t.Fatalf("readLength(%q) = (%d,%d), want (%d,%d)", c.in, gotLen, gotDelta, c.wantLen, c.wantDelta)
		}
	}
}

func TestReadString(t *testing.T) {
	val, delta, err := readString([]byte("+OK\r\n"))
	if err != nil {
		t.Fatalf("readString error: %v", err)
	}
	if val != "OK" {
		t.Fatalf("readString value = %q, want %q", val, "OK")
	}
	if delta != 5 { // "+OK\r\n" -> starts at 1, stops before \r, returns pos+2
		t.Fatalf("readString delta = %d, want %d", delta, 5)
	}
}

func TestReadBulkString(t *testing.T) {
	val, delta, err := readBulkString([]byte("$5\r\nhello\r\n"))
	if err != nil {
		t.Fatalf("readBulkString error: %v", err)
	}
	if val != "hello" {
		t.Fatalf("readBulkString value = %q, want %q", val, "hello")
	}
	if delta != 5 {
		t.Fatalf("readBulkString delta = %d, want %d", delta, 5)
	}
}

func TestDecodeOne_String(t *testing.T) {
	val, _, err := DecodeOne([]byte("+PONG\r\n"))
	if err != nil {
		t.Fatalf("DecodeOne error: %v", err)
	}
	if s, ok := val.(string); !ok || s != "PONG" {
		t.Fatalf("DecodeOne value = %#v, want %q", val, "PONG")
	}
}

func TestDecode_BulkString(t *testing.T) {
	val, err := Decode([]byte("$5\r\nworld\r\n"))
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if s, ok := val.(string); !ok || s != "world" {
		t.Fatalf("Decode value = %#v, want %q", val, "world")
	}
}

func FuzzReadLength(f *testing.F) {
	seed := []string{"0\r\n", "1\r\n", "9\r\n", "10\r\n", "123\r\n"}
	for _, s := range seed {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		// keep only digits then add CRLF
		digits := make([]byte, 0, len(s))
		for i := 0; i < len(s); i++ {
			if s[i] >= '0' && s[i] <= '9' {
				digits = append(digits, s[i])
			}
		}
		digits = append(digits, '\r', '\n')
		_, delta := readLength(digits)
		if delta < 3 { // at least one digit plus CRLF
			t.Fatalf("delta too small: %d for %q", delta, string(digits))
		}
	})
}
