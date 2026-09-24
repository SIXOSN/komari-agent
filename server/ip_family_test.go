package server

import "testing"

func TestResolveIPFamilyLiterals(t *testing.T) {
	for _, test := range []struct {
		address string
		family  string
		valid   bool
	}{
		{"1.1.1.1", "ipv4", true},
		{"1.1.1.1", "ipv6", false},
		{"2001:db8::1", "ipv6", true},
		{"2001:db8::1", "ipv4", false},
	} {
		_, err := resolveIPFamily(test.address, test.family)
		if (err == nil) != test.valid {
			t.Errorf("resolveIPFamily(%q, %q): valid=%v, err=%v", test.address, test.family, test.valid, err)
		}
	}
}
