# iso8601

Process iso8601 duration without using regex.

- Support duration range from `P9223372036854775807Y9223372036854775807M9223372036854775807W9223372036854775807DT9223372036854775807H9223372036854775807M9223372036854775807.999999999S` to `P-9223372036854775808Y-9223372036854775808M-9223372036854775808W-9223372036854775808DT-9223372036854775808H-9223372036854775808M-9223372036854775808.999999999S`

## Usage

```shell
go get github.com/NateScarlet/iso8601
```

```go
import github.com/NateScarlet/iso8601/pkg/iso8601

d, err := iso8601.ParseDuration("P1D")
// Duration{Days: 1}, nil
d.TimeDuration()
// time.Duration(24 * time.Hour)
d.String()
// "P1D"

iso8601.ParseDuration("P-1D")
// Duration{Days: -1}, nil

iso8601.ParseDuration("-P1D")
// Duration{Days: 1, Negative: true}, nil

iso8601.ParseDuration("P0.5D")
// Duration{Hours: 12}, nil

iso8601.ParseDuration("P0.5DT0.5H")
// nil, ErrInvalidDuration

iso8601.ParseDuration("P.D")
// nil, ErrInvalidDuration
```

## Benchmark

Athlon 64 X2 Dual core 5600+ 2.9Ghz

```text
BenchmarkDurationString-2     3603614        378 ns/op       48 B/op        1 allocs/op
BenchmarkParseDuration-2      3438430        332 ns/op        0 B/op        0 allocs/op
```
