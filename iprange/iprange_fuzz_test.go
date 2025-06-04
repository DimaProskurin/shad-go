package iprange

import "testing"

func FuzzParseList(f *testing.F) {
	testCases := []string{
		"0.0.0.0",
		"0.0.0.0-10",
		"0-10.0-255.0.0-64",
		"192.168.1.0",
		"192.168.1.*",
		"*.*.*.*",
		"192.168.1.0-64",
		"255.255.15.*",
		"255.255.*.*",
		"255.255.16.0/24",
		"255.0.0.0/8",
		"255.255.0.0/16",
		"10.0.0.1, 10.0.0.5-10",
		"10.0.0.1, 255.255.0.0/16",
		"255.255.*.*, 255.0.0.0/8",
	}
	for _, tc := range testCases {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, in string) {
		_, err := ParseList(in)
		if err != nil {
			t.Skip()
		}
	})
}
