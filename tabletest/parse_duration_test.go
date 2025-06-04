package tabletest

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		in       string
		expected time.Duration
	}{
		{in: "1s"},
		{in: "1m"},
		{in: "1h"},
		{in: "1µs"},
		{in: "1us"},
		{in: "2h20m20s"},
		{in: "1.14s"},
		{in: "-1.14s"},
		{in: "0"},
		{in: "bad"},
		{in: ""},
		{in: "999999999999999999999999999999999999999s"},
		{in: "1.999999999999999999999999999999999999999s"},
		{in: "1ds"},
		{in: ".s"},
		{in: "1"},
		{in: "99999999999s"},
		{in: "9223372036854775808s"},
		{in: "1.9223372036854775808s"},
		{in: "9223372036854.999ms"},
		{in: "9223372036854775.000us999ns"},
	}

	for _, tt := range testCases {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			dur, err := ParseDuration(tt.in)
			expected, expectedErr := time.ParseDuration(tt.in)
			if expectedErr != nil {
				require.Error(t, err)
			}
			if expectedErr == nil {
				require.NoError(t, err)
			}
			require.Equal(t, expected, dur)
		})
	}
}
