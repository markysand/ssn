// Package ssn contains tools to generate and handle ssn-s.
package ssn

import (
	"errors"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SSN is a representation of a 12 digit swedish social security number
type SSN [12]int

// GetRandomTime gets a random time in the past
func GetRandomTime(from, to time.Duration) time.Time {
	t1 := time.Now()
	diff := from - to
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
	ErrFormat   = errors.New("Input does not match simple regexp guard YYYYMMDD-XXXX")
	ErrDate     = errors.New("Could not parse date")
	ErrChecksum = errors.New("Checksum is incorrect")
)

// NewSSNFromString makes a ssn type object from a string and at the same time validates that string
// to format, date, checksum and will send errors accordingly
func NewSSNFromString(s string) (*SSN, error) {
	var re = regexp.MustCompile(`^[0-9]{8}-[0-9]{4}$`)
	ok := re.MatchString(s)
	if !ok {
		return nil, ErrFormat
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
		return nil, ErrChecksum
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
		n[9] = rand.Intn(1) + 8
	} else {
		if ss[1] == '?' {
			n[8] = rand.Intn(9)
		}
		if ss[2] == '?' {
			n[9] = rand.Intn(9)
		}
	}
	switch ss[3] {
	case '?':
		n[10] = rand.Intn(9)
	case 'f':
		n[10] = rand.Intn(4) * 2
	case 'm':
		n[10] = rand.Intn(4)*2 + 1
	}
	if ss[4] == 'c' {
		n[11] = GetChecksum(*n)
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
