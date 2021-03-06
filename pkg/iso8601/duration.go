package iso8601

import (
	"errors"
	"strconv"
	"time"
)

const (
	// Day used when convert to time.Duration.
	Day = time.Hour * 24
	// Week used when convert to time.Duration.
	Week = Day * 7
	// Month used when convert to time.Duration.
	// Data from moment.js:
	// 400 years have 146097 days (taking into account leap year rules)
	Month = Day / 10 * 146097 / 4800 * 10
	// Year used when convert to time.Duration.
	Year = Month * 12

	maxInt64 int64 = 1<<63 - 1
	minInt64 int64 = -1 << 63
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

// ErrOverflow indicate value is overflowed.
var ErrOverflow = errors.New("iso8601: overflow")

// addInt handle overflow when add int64
func addInt(base int64, v int64) (int64, error) {
	if base > 0 {
		if v > maxInt64-base ||
			v < minInt64+base {
			return 0, ErrOverflow
		}
	} else {
		if v > maxInt64+base ||
			(v < minInt64-base) {
			return 0, ErrOverflow
		}
	}
	return base + v, nil
}

// multiplyInt handle overflow when multiple int64
func multiplyInt(base int64, v int64) (int64, error) {
	if base > 0 {
		if v > maxInt64/base ||
			v < minInt64/base {
			return 0, ErrOverflow
		}
	} else {
		if v > minInt64/base ||
			v < maxInt64/base {
			return 0, ErrOverflow
		}
	}
	return base * v, nil
}
func addNano(base int64, num int64, unit time.Duration) (int64, error) {
	var v int64
	var err error
	v, err = multiplyInt(int64(unit), num)
	if err != nil {
		return 0, err
	}
	return addInt(base, v)
}

// TimeDuration convert this to time.Duration.
// Will loose some precision because days, months, years can have different length.
func (d Duration) TimeDuration() (ret time.Duration, err error) {
	var nano int64
	nano, err = addNano(nano, d.Years, Year)
	if err != nil {
		return
	}
	nano, err = addNano(nano, d.Months, Month)
	if err != nil {
		return
	}
	nano, err = addNano(nano, d.Weeks, Week)
	if err != nil {
		return
	}
	nano, err = addNano(nano, d.Days, Day)
	if err != nil {
		return
	}
	nano, err = addNano(nano, d.Hours, time.Hour)
	if err != nil {
		return
	}
	nano, err = addNano(nano, d.Minutes, time.Minute)
	if err != nil {
		return
	}
	nano, err = addNano(nano, d.Seconds, time.Second)
	if err != nil {
		return
	}
	nano, err = addNano(nano, d.Nanoseconds, time.Nanosecond)
	if err != nil {
		return
	}
	if d.Negative {
		nano = -nano
	}
	return time.Duration(nano), nil
}

// MustTimeDuration execute TimeDuration and panic if error.
func (d Duration) MustTimeDuration() time.Duration {
	var ret, err = d.TimeDuration()
	if err != nil {
		panic(err)
	}
	return ret
}

// appendFrac append the fraction of v/10**prec (e.g., ".12345") into the
// tail of buf, omitting trailing zeros. It omits the decimal
// point too when the fraction is 0. It returns the index where the
// output bytes begin and the value v/10**prec.
func appendFrac(b []byte, v uint64, prec int) []byte {
	var buf [10]byte
	var w = len(buf)
	var print bool
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
	return append(b, buf[w:]...)
}

// AppendFormat is like String but appends the textual
// representation to b and returns the extended buffer.
func (d Duration) AppendFormat(b []byte) []byte {
	if d.Negative {
		b = append(b, '-')
	}

	b = append(b, 'P')
	var prefixWidth = len(b)

	// Y
	if d.Years != 0 {
		b = strconv.AppendInt(b, d.Years, 10)
		b = append(b, 'Y')
	}

	// M
	if d.Months != 0 {
		b = strconv.AppendInt(b, d.Months, 10)
		b = append(b, 'M')
	}

	// W
	if d.Weeks != 0 {
		b = strconv.AppendInt(b, d.Weeks, 10)
		b = append(b, 'W')
	}

	// D
	if d.Days != 0 {
		b = strconv.AppendInt(b, d.Days, 10)
		b = append(b, 'D')
	}

	// T
	if d.Hours != 0 ||
		d.Minutes != 0 ||
		d.Seconds != 0 ||
		d.Nanoseconds != 0 {
		b = append(b, 'T')
	}

	// H
	if d.Hours != 0 {
		b = strconv.AppendInt(b, d.Hours, 10)
		b = append(b, 'H')
	}

	// M
	if d.Minutes != 0 {
		b = strconv.AppendInt(b, d.Minutes, 10)
		b = append(b, 'M')
	}

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
		if neg {
			b = append(b, '-')
		}
		b = strconv.AppendUint(b, v, 10)
		b = appendFrac(b, f, 9)
		b = append(b, 'S')
	}

	if len(b) == prefixWidth {
		b = append(b, '0')
		b = append(b, 'D')
	}
	return b
}

func (d Duration) String() string {
	return string(d.AppendFormat(make([]byte, 0, 256)))
}

// NewDuration create duration from nanoseconds (e.g. time.Duration)
// Only use unit that below days, because days can have different length (e.g. DST).
// just declaring a Duration variable is enough to use duration.
func NewDuration(nanoseconds int64) *Duration {
	var ret = new(Duration)
	if nanoseconds < 0 {
		ret.Negative = true
		nanoseconds = -nanoseconds
	}
	ret.Hours = nanoseconds / int64(time.Hour)
	nanoseconds %= int64(time.Hour)
	ret.Minutes = nanoseconds / int64(time.Minute)
	nanoseconds %= int64(time.Minute)
	ret.Seconds = nanoseconds / int64(time.Second)
	nanoseconds %= int64(time.Second)
	ret.Nanoseconds = nanoseconds
	return ret
}

var errLeadingInt = errors.New("iso8601: bad [0-9]*") // never printed
func leadingNegative(s string) (x bool, rem string) {
	if s == "" {
		return false, s
	}
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
			return 0, "", ErrOverflow
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", ErrOverflow
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

// ParseDuration parse iso8601 duration string.
func ParseDuration(s string) (ret Duration, err error) {
	orig := s
	ret.Negative, s = leadingNegative(s)

	if s == "" || s[0] != 'P' {
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
			case 'H':
				ret.Hours += v
				ret.Minutes += int64(float64(f) * (float64(time.Hour/time.Minute) / scale))
			case 'M':
				ret.Minutes += v
				ret.Seconds += int64(float64(f) * (float64(time.Minute/time.Second) / scale))
			case 'S':
				ret.Seconds += v
				ret.Nanoseconds += int64(float64(f) * (float64(time.Second/time.Nanosecond) / scale))
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
