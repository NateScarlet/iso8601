package iso8601

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringDuration(t *testing.T) {
	for _, c := range []struct {
		duration Duration
		expected string
	}{
		{duration: Duration{}, expected: "P0D"},
		{duration: Duration{Years: 1}, expected: "P1Y"},
		{duration: Duration{Weeks: 1}, expected: "P1W"},
		{duration: Duration{Months: 1}, expected: "P1M"},
		{duration: Duration{Days: 1}, expected: "P1D"},
		{duration: Duration{Hours: 1}, expected: "PT1H"},
		{duration: Duration{Hours: 24}, expected: "PT24H"},
		{duration: Duration{Minutes: 1}, expected: "PT1M"},
		{duration: Duration{Seconds: 1}, expected: "PT1S"},
		{duration: Duration{Nanoseconds: 1}, expected: "PT0.000000001S"},
		{duration: Duration{Minutes: 1, Seconds: 1}, expected: "PT1M1S"},
		{duration: Duration{Negative: true}, expected: "-P0D"},
		{duration: Duration{Hours: -1}, expected: "PT-1H"},
		{duration: Duration{Minutes: -1}, expected: "PT-1M"},
		{duration: Duration{Seconds: -1}, expected: "PT-1S"},
		{duration: Duration{Minutes: -1, Seconds: -1}, expected: "PT-1M-1S"},
		{duration: Duration{
			Years:       math.MaxInt64,
			Months:      math.MaxInt64,
			Weeks:       math.MaxInt64,
			Days:        math.MaxInt64,
			Hours:       math.MaxInt64,
			Minutes:     math.MaxInt64,
			Seconds:     math.MaxInt64,
			Nanoseconds: int64(time.Second) - 1,
		}, expected: "P9223372036854775807Y9223372036854775807M9223372036854775807W9223372036854775807DT9223372036854775807H9223372036854775807M9223372036854775807.999999999S"},
		{duration: Duration{
			Years:       math.MinInt64,
			Months:      math.MinInt64,
			Weeks:       math.MinInt64,
			Days:        math.MinInt64,
			Hours:       math.MinInt64,
			Minutes:     math.MinInt64,
			Seconds:     math.MinInt64,
			Nanoseconds: -int64(time.Second) + 1,
			Negative:    true,
		}, expected: "-P-9223372036854775808Y-9223372036854775808M-9223372036854775808W-9223372036854775808DT-9223372036854775808H-9223372036854775808M-9223372036854775808.999999999S"},
	} {
		t.Run(c.expected, func(t *testing.T) {
			assert.Equal(t, c.expected, c.duration.String())
		})
	}
}

func TestParseDuration(t *testing.T) {
	for _, c := range []struct {
		s        string
		expected Duration
	}{
		{
			s: "P1Y1M1W1DT1H1M1.001S",
			expected: Duration{Years: 1, Months: 1, Weeks: 1, Days: 1,
				Hours: 1, Minutes: 1, Seconds: 1, Nanoseconds: 1e6},
		},
		{s: "P0D", expected: Duration{}},
		{s: "P1D", expected: Duration{Days: 1}},
		{s: "P-1D", expected: Duration{Days: -1}},
		{s: "P1M", expected: Duration{Months: 1}},
		{s: "P-1M", expected: Duration{Months: -1}},
		{s: "P1Y", expected: Duration{Years: 1}},
		{s: "P-1Y", expected: Duration{Years: -1}},
		{s: "PT1S", expected: Duration{Seconds: 1}},
		{s: "PT-1S", expected: Duration{Seconds: -1}},
		{s: "PT1H", expected: Duration{Hours: 1}},
		{s: "PT-1H", expected: Duration{Hours: -1}},
		{s: "PT1M", expected: Duration{Minutes: 1}},
		{s: "PT-1M", expected: Duration{Minutes: -1}},
		{s: "PT1H1M1S", expected: Duration{Hours: 1, Minutes: 1, Seconds: 1}},
		{s: "PT0.5H", expected: Duration{Minutes: 30}},
		{s: "PT0.001S", expected: Duration{Nanoseconds: 1e6}},
		{s: "-PT1H", expected: Duration{Hours: 1, Negative: true}},
		{s: "+PT1H", expected: Duration{Hours: 1}},
	} {
		t.Run(c.s, func(t *testing.T) {
			v, err := ParseDuration(c.s)
			require.NoError(t, err)
			assert.Equal(t, c.expected, v)
		})
	}
}

func BenchmarkDurationString(b *testing.B) {
	x := Duration{
		Years:       1234,
		Months:      -1234,
		Weeks:       1234,
		Days:        -1234,
		Hours:       1234,
		Minutes:     -1234,
		Nanoseconds: 1234,
		Negative:    true,
	}
	for i := 0; i < b.N; i++ {
		_ = x.String()
	}
}
func BenchmarkParseDuration(b *testing.B) {
	x := "P1Y23M34W56DT78H90M12.3456789S"
	for i := 0; i < b.N; i++ {
		_, err := ParseDuration(x)
		if err != nil {
			b.Fatal(err)
		}
	}
}
