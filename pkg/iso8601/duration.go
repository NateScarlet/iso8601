package iso8601

import (
	"errors"
	"time"
)

const (
	// Day used when parse duration day.
	Day = time.Hour * 24
	// Week used when parse duration year.
	Week = Day * 7
	// Month used when parse duration year.
	// Data from moment.js:
	// 400 years have 146097 days (taking into account leap year rules)
	Month = Day / 10 * 146097 / 4800 * 10
	// Year used when parse duration year.
	Year = Month * 12
)

// Duration contains iso8601 duration data.
// https://en.wikipedia.org/wiki/ISO_8601#Durations
type Duration struct {
	Years   int64
	Months  int64
	Weeks   int64
	Days    int64
	Hours   int64
	Minutes int64
	Seconds int64
	// Nanoseconds should never greater than time.Seconds-1
	// or less than -time.Seconds+1
	Nanoseconds int64
	Negative    bool
}

// Direction return -1 if negative else 1.
func (d Duration) Direction() int64 {
	if d.Negative {
		return -1
	}
	return 1
}

// TimeDuration convert this to time.Duration.
// Will loose some precision because days, months, years can have different length.
func (d Duration) TimeDuration() time.Duration {
	return time.Duration(
		(d.Years*int64(Year) +
			d.Weeks*int64(Week) +
			d.Days*int64(Day) +
			d.Hours*int64(time.Hour) +
			d.Minutes*int64(time.Minute) +
			d.Seconds*int64(time.Second) +
			d.Nanoseconds*int64(time.Nanosecond)) *
			d.Direction())
}

func (d Duration) String() string {
	// longest supported value: -P-9223372036854775808Y-9223372036854775808M-9223372036854775808W-9223372036854775808DT-9223372036854775808H-9223372036854775808M-9223372036854775808.999999999S
	var buf = [256]byte{}
	var w = len(buf)

	// S
	if d.Seconds != 0 ||
		d.Nanoseconds != 0 {
		var v, f = uint64(d.Seconds), uint64(d.Nanoseconds)
		var neg bool
		if d.Seconds < 0 {
			neg = true
			v = -v
		}
		if d.Nanoseconds < 0 {
			if neg {
				f = -f
			} else {
				v--
				f = uint64(int64(time.Second) + d.Nanoseconds)
			}
		}
		w--
		buf[w] = 'S'
		var u uint64
		w, u = fmtFrac(buf[:w], f, 9)
		v += u
		w = fmtUint(buf[:w], v)
		if neg {
			w--
			buf[w] = '-'
		}
	}

	// M
	if d.Minutes != 0 {
		w--
		buf[w] = 'M'
		w = fmtInt(buf[:w], d.Minutes)
	}

	// H
	if d.Hours != 0 {
		w--
		buf[w] = 'H'
		w = fmtInt(buf[:w], d.Hours)
	}

	// T
	if w != len(buf) {
		w--
		buf[w] = 'T'
	}

	// D
	if d.Days != 0 {
		w--
		buf[w] = 'D'
		w = fmtInt(buf[:w], d.Days)
	}

	// W
	if d.Weeks != 0 {
		w--
		buf[w] = 'W'
		w = fmtInt(buf[:w], d.Weeks)
	}

	// M
	if d.Months != 0 {
		w--
		buf[w] = 'M'
		w = fmtInt(buf[:w], d.Months)
	}

	// Y
	if d.Years != 0 {
		w--
		buf[w] = 'Y'
		w = fmtInt(buf[:w], d.Years)
	}

	if w == len(buf) {
		w--
		buf[w] = 'D'
		w--
		buf[w] = '0'
	}
	w--
	buf[w] = 'P'

	if d.Negative {
		w--
		buf[w] = '-'
	}
	return string(buf[w:])
}

// fmtUint formats v into the tail of buf.
// It returns the index where the output begins.
func fmtUint(buf []byte, v uint64) int {
	w := len(buf)
	if v == 0 {
		w--
		buf[w] = '0'
	} else {
		for v > 0 {
			w--
			buf[w] = byte(v%10) + '0'
			v /= 10
		}
	}
	return w
}

// fmtInt formats v into the tail of buf.
// It returns the index where the output begins.
func fmtInt(buf []byte, v int64) int {
	var w = len(buf)
	var u = uint64(v)
	var neg bool
	if v < 0 {
		neg = true
		u = -u
	}
	w = fmtUint(buf[:w], u)
	if neg {
		w--
		buf[w] = '-'
	}
	return w
}

// fmtFrac formats the fraction of v/10**prec (e.g., ".12345") into the
// tail of buf, omitting trailing zeros. It omits the decimal
// point too when the fraction is 0. It returns the index where the
// output bytes begin and the value v/10**prec.
func fmtFrac(buf []byte, v uint64, prec int) (nw int, nv uint64) {
	// Omit trailing zeros up to and including decimal point.
	w := len(buf)
	print := false
	for i := 0; i < prec; i++ {
		digit := v % 10
		print = print || digit != 0
		if print {
			w--
			buf[w] = byte(digit) + '0'
		}
		v /= 10
	}
	if print {
		w--
		buf[w] = '.'
	}
	return w, v
}

// FormatDuration to iso8601 format.
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "P0D"
	}
	var buf [32]byte
	w := len(buf)

	u := uint64(d)
	neg := d < 0
	if neg {
		u = -u
	}

	w--
	buf[w] = 'S'

	w, u = fmtFrac(buf[:w], u, 9)

	// u is now integer seconds
	w = fmtUint(buf[:w], u%60)
	u /= 60
	if buf[w] == '0' && buf[w+1] == 'S' {
		w += 2
	}

	// u is now integer minutes
	if u > 0 {
		w--
		buf[w] = 'M'
		w = fmtUint(buf[:w], u%60)
		u /= 60
		if buf[w] == '0' {
			w += 2
		}

		// u is now integer hours
		// Stop at hours because days can be different lengths.
		if u > 0 {
			w--
			buf[w] = 'H'
			w = fmtUint(buf[:w], u)
		}
	}
	w--
	buf[w] = 'T'
	w--
	buf[w] = 'P'
	if neg {
		w--
		buf[w] = '-'
	}

	return string(buf[w:])
}

var errLeadingInt = errors.New("iso8601: bad [0-9]*") // never printed
func leadingNegative(s string) (x bool, rem string) {
	i := 0
	c := s[0]
	if c == '-' || c == '+' {
		i++
		x = c == '-'
	}
	return x, s[i:]
}

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<63-1)/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x int64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + int64(c) - '0'
		if y < 0 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

// ErrInvalidDuration returned when parse failed.
type ErrInvalidDuration struct {
	String string
}

func (err ErrInvalidDuration) Error() string {
	return "iso8601: invalid duration " + err.String
}

// ParseDuration iso8601 duration string.
func ParseDuration(s string) (ret Duration, err error) {
	orig := s
	ret.Negative, s = leadingNegative(s)

	if s[0] != 'P' {
		err = ErrInvalidDuration{String: orig}
		return
	}
	s = s[1:]

	var afterT bool
	for s != "" {
		if s[0] == 'T' {
			s = s[1:]
			afterT = true
			continue
		}
		var v, f int64
		var scale float64 = 1
		var neg bool
		var pre, post bool
		neg, s = leadingNegative(s)

		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return
		}
		pre = pl != len(s) // whether we consumed anything before a period
		if neg {
			v = -v
		}

		// Consume (\.[0-9]*)?
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
			if neg {
				f = -f
			}
		}
		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			err = ErrInvalidDuration{String: orig}
			return
		}

		// Consume unit.
		if s == "" {
			err = ErrInvalidDuration{String: orig}
			return
		}
		var u = s[0]
		s = s[1:]
		if !afterT {
			switch u {
			case 'Y':
				ret.Years += v
				ret.Months += int64(float64(f) * (float64(Year/Month) / scale))
			case 'M':
				ret.Months += v
				ret.Weeks += int64(float64(f) * (float64(Month/Week) / scale))
			case 'W':
				ret.Weeks += v
				ret.Days += int64(float64(f) * (float64(Week/Day) / scale))
			case 'D':
				ret.Days += v
				ret.Hours += int64(float64(f) * (float64(Day/time.Hour) / scale))
			default:
				// unknown unit
				err = ErrInvalidDuration{String: orig}
				return
			}
		} else {
			switch u {
			case 'S':
				ret.Seconds += v
				ret.Nanoseconds += int64(float64(f) * (float64(time.Second/time.Nanosecond) / scale))
			case 'M':
				ret.Minutes += v
				ret.Seconds += int64(float64(f) * (float64(time.Minute/time.Second) / scale))
			case 'H':
				ret.Hours += v
				ret.Minutes += int64(float64(f) * (float64(time.Hour/time.Minute) / scale))
			default:
				// unknown unit
				err = ErrInvalidDuration{String: orig}
				return
			}
		}

		if post && s != "" {
			// must end after fraction used.
			err = ErrInvalidDuration{String: orig}
			return
		}
	}
	return
}
