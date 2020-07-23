package timefmt

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type parseError struct {
	source, format string
	err            error
}

func (err *parseError) Error() string {
	return fmt.Sprintf("failed to parse %q with %q: %s", err.source, err.format, err.err)
}

// Parse time string using the format.
func Parse(source, format string) (t time.Time, err error) {
	year, month, day, hour, min, sec, nsec, loc := 0, 1, 1, 0, 0, 0, 0, time.UTC
	defer func() {
		if err != nil {
			err = &parseError{source, format, err}
		}
	}()
	var j, diff int
	for i, l := 0, len(source); i < len(format); i++ {
		if b := format[i]; b == '%' {
			i++
			b = format[i]
			switch b {
			case 'Y':
				if j+4 > l {
					err = errors.New("cannot parse %Y")
					return
				}
				if year, err = strconv.Atoi(string(source[j : j+4])); err != nil {
					return
				}
				j += 4
			case 'y':
				if j+2 > l {
					err = errors.New("cannot parse %y")
					return
				}
				if year, err = strconv.Atoi(string(source[j : j+2])); err != nil {
					return
				}
				year += (time.Now().Year() / 100) * 100
				j += 2
			case 'm':
				if month, diff, err = parseNumber(source[j:], 2, 'm'); err != nil {
					return
				}
				j += diff
			case 'B':
				if month, diff = lookup(source[j:], longMonthNames); month == 0 {
					err = errors.New("cannot parse %B")
					return
				}
				j += diff
			case 'b':
				if month, diff = lookup(source[j:], shortMonthNames); month == 0 {
					err = errors.New("cannot parse %b")
					return
				}
				j += diff
			case 'A':
				var week int
				if week, diff = lookup(source[j:], longWeekNames); week == 0 {
					err = errors.New("cannot parse %A")
					return
				}
				j += diff
			case 'a':
				var week int
				if week, diff = lookup(source[j:], shortWeekNames); week == 0 {
					err = errors.New("cannot parse %a")
					return
				}
				j += diff
			case 'w':
				if j >= l || source[j] < '0' || '6' < source[j] {
					err = errors.New("cannot parse %w")
					return
				}
				j++
			case 'd':
				if day, diff, err = parseNumber(source[j:], 2, 'd'); err != nil {
					return
				}
				j += diff
			case 'H':
				if hour, diff, err = parseNumber(source[j:], 2, 'H'); err != nil {
					return
				}
				j += diff
			case 'M':
				if min, diff, err = parseNumber(source[j:], 2, 'M'); err != nil {
					return
				}
				j += diff
			case 'S':
				if sec, diff, err = parseNumber(source[j:], 2, 'S'); err != nil {
					return
				}
				j += diff
			case 'f':
				var msec int
				if msec, diff, err = parseNumber(source[j:], 6, 'f'); err != nil {
					return
				}
				j += diff
				nsec = msec * 1000
				for diff < 6 {
					nsec *= 10
					diff++
				}
			default:
				err = fmt.Errorf("unexpected format: %q", format[i-1:i+1])
				return
			}
		} else if j >= len(source) || b != source[j] {
			err = fmt.Errorf("expected %q", b)
			return
		} else {
			j++
		}
	}
	if j < len(source) {
		err = fmt.Errorf("unconverted string: %q", source[j:])
		return
	}
	return time.Date(year, time.Month(month), day, hour, min, sec, nsec, loc), nil
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func parseNumber(source string, max int, format byte) (int, int, error) {
	if len(source) > 0 && isDigit(source[0]) {
		for i := 1; i < max; i++ {
			if i >= len(source) || !isDigit(source[i]) {
				val, err := strconv.Atoi(string(source[:i]))
				return val, i, err
			}
		}
		val, err := strconv.Atoi(string(source[:max]))
		return val, max, err
	}
	return 0, 0, fmt.Errorf("cannot parse %%%c", format)
}

func lookup(source string, candidates []string) (int, int) {
L:
	for i, xs := range candidates {
		for j, x := range []byte(xs) {
			if j >= len(source) {
				continue L
			}
			if y := source[j]; x != y && x|('a'-'A') != y|('a'-'A') {
				continue L
			}
		}
		return i + 1, len(xs)
	}
	return 0, 0
}
