// Package ssn contains tools to generate and handle ssn-s.
package ssn

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// SSN is a representation of a 12 digit swedish social security number
type SSN [12]int

// GetRandomTime gets a random time
// Durations count backwards from Now
func GetRandomTime(from, to time.Duration) time.Time {
	t1 := time.Now()
	diff := from - to
	if diff <= 0 {
		return t1.Add(-from)
	}
	randomDiff := time.Duration(rand.Int63n(int64(diff)))
	t2 := t1.Add(-randomDiff - to)
	return t2
}

func getDigit(i int) (digit, next int) {
	return i % 10, i / 10
}

func sumDigits(i int) int {
	for i > 9 {
		dig, next := getDigit(i)
		i = dig + next
	}
	return i
}

// Error codes for parsing of SSNs
var (
	ErrFormat   = errors.New("Input does not match YYYYMMDD-XXXX or YYYYMMDDXXXX")
	ErrDate     = errors.New("Could not parse date")
	ErrChecksum = errors.New("Checksum is incorrect")
)

// NewSSNFromString makes a ssn type object from a string and at the same time validates that string
// to format, date, checksum and will send errors accordingly
func NewSSNFromString(s string) (*SSN, error) {
	var re = regexp.MustCompile(`^[0-9]{8}-?[0-9]{4}$`)
	ok := re.MatchString(s)
	if !ok {
		return nil, ErrFormat
	}
	if len(s) == 12 {
		s = s[0:8] + "-" + s[8:12]
	}
	tm, err := time.Parse("20060102", s[0:8])
	if err != nil {
		return nil, ErrDate
	}
	var ssn SSN
	ssn.SetDate(tm)
	for i := 8; i < 12; i++ {
		ssn[i], err = strconv.Atoi(string(s[i+1]))
		if err != nil {
			panic("Error parsing digit, probably got letter")
		}
	}
	if GetChecksum(ssn) != ssn[11] {
		return &ssn, ErrChecksum
	}
	return &ssn, nil
}

func safeString(s, def string) string {
	l1, l2 := len(s), len(def)
	if l1 >= l2 {
		return s[:l2]
	}
	return s + def[l1:l2]
}

func trySetDigitFromRune(r rune, i *int) {
	switch r {
	case '*':
	case '?':
		*i = rand.Intn(10)
	default:
		if x, err := strconv.Atoi(string(r)); err == nil {
			*i = x
		}
	}
}

// SetLastDigits will set the last digits (not checksum)
// ? = random
// * = keep current
// m = random male
// f = random female
// s = safe (980-999) last digits
// c = get checksum
func (n *SSN) SetLastDigits(s string) {
	ss := []rune(safeString(s, "****"))
	if (ss[0] == 's') || (ss[1] == 's') {
		n[8] = 9
		n[9] = rand.Intn(2) + 8
	} else {
		trySetDigitFromRune(ss[0], &n[8])
		trySetDigitFromRune(ss[1], &n[9])
	}
	switch ss[2] {
	case 'f':
		n[10] = rand.Intn(5) * 2
	case 'm':
		n[10] = rand.Intn(5)*2 + 1
	default:
		trySetDigitFromRune(ss[2], &n[10])
	}
	switch ss[3] {
	case 'c':

		n[11] = GetChecksum(*n)
	case '*':
	default:
		trySetDigitFromRune(ss[3], &n[11])
	}
}

// SetDate will set the time/date part of the SSN from a time.Time struct
func (n *SSN) SetDate(t time.Time) {
	y := t.Year()
	n[3], y = getDigit(y)
	n[2], y = getDigit(y)
	n[1], y = getDigit(y)
	n[0], _ = getDigit(y)
	m := int(t.Month())
	n[5], m = getDigit(m)
	n[4], _ = getDigit(m)
	d := t.Day()
	n[7], d = getDigit(d)
	n[6], _ = getDigit(d)
}

// String returns SSN in standard YYYYMMDD-XXXX formats
func (n SSN) String() string {
	return n.Format(true, true)
}

// Format will return an SSN in custom formats
func (n SSN) Format(century, dash bool) string {
	var i int
	if !century {
		i = 2
	}
	var b strings.Builder
	for i < len(n) {
		b.WriteString(strconv.Itoa(n[i]))
		if i == 7 && dash {
			b.WriteString("-")
		}
		i++
	}
	return b.String()
}

// GetChecksum returns the Luhn algoritm checksum for the ssn
func GetChecksum(n SSN) int {
	var sum int
	for i := 2; i < 11; i++ {
		sum += sumDigits(((i+1)%2 + 1) * n[i])
	}
	result := (10 - sum%10) % 10
	return result
}

func newRandomSSN() *SSN {
	var ssn SSN
	t := GetRandomTime(time.Hour*24*365*100, 0)
	ssn.SetDate(t)
	ssn.SetLastDigits("???c")
	return &ssn
}

// NewRandomSSN will return a SSN of a 0-100 year old
func NewRandomSSN() *SSN {
	ssn := newRandomSSN()
	ssn.SetLastDigits("???c")
	return ssn
}

// NewSafeRandomSSN will return a safe SSN of a 0-100 year old
func NewSafeRandomSSN() *SSN {
	ssn := newRandomSSN()
	ssn.SetLastDigits("ss?c")
	return ssn
}

func intSliceToInt(is []int) (sum int) {
	for i, k := len(is)-1, 1; i >= 0; i, k = i-1, k*10 {
		sum += k * is[i]
	}
	return
}

func (n SSN) Date() (year int, month time.Month, day int) {
	return intSliceToInt(n[0:4]), time.Month(intSliceToInt(n[4:6])), intSliceToInt(n[6:8])
}

func intSliceToString(is []int) string {
	var b strings.Builder
	for _, n := range is {
		b.WriteString(strconv.Itoa(n))
	}
	return b.String()
}

func (n SSN) Time() time.Time {
	t, err := time.Parse("20060102", intSliceToString(n[0:8]))
	if err != nil {
		panic(fmt.Sprint("SSN format invalid, cannot be parsed to Time", n))
	}
	return t
}

func (n SSN) Age(now time.Time) time.Duration {
	return now.Sub(n.Time())
}

func (n SSN) Female() bool {
	return n[10]%2 == 0
}
