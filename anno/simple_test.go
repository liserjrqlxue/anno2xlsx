package anno

import "testing"

func TestReverse(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Hello, world", "dlrow ,olleH"},
		{"Hello, 世界", "界世 ,olleH"},
		{"", ""},
		{"ACGTacgtN", "NacgtACGT"},
	}
	for _, c := range cases {
		got := ReverseComplement(c.in)
		if got != c.want {
			t.Errorf("ReverseComplement(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
