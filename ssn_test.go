package ssn

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSetDate(t *testing.T) {
	tests := []struct {
		time   string
		output string
	}{{
		"1975-09-30T10:00:00+02:00",
		"19750930-0000",
	}, {
		"2009-03-01T05:00:00Z",
		"20090301-0000",
	}, {
		"2011-05-30T17:00:00Z",
		"20110530-0000",
	}}
	for i, tc := range tests {
		t.Run(fmt.Sprint("Get SSN from date, test number:", i), func(t *testing.T) {
			refTime, err := time.Parse(time.RFC3339, tc.time)
			if err != nil {
				t.Error("Could not get reftime", err)
				t.FailNow()
			}
			var person SSN
			person.SetDate(refTime)
			if person.String() != tc.output {
				t.Error("Got ", person.String(), ", Want: ", tc.output)
			}
		})
	}
}

const util = "Test no: %v, got: %v want %v"

func TestFormat(t *testing.T) {
	pnr := SSN{1, 9, 7, 5, 0, 9, 3, 0, 1, 9, 3, 8}
	tests := []struct {
		cent   bool
		dash   bool
		output string
	}{
		{false, false, "7509301938"},
		{true, false, "197509301938"},
		{false, true, "750930-1938"},
		{true, true, "19750930-1938"},
	}
	for i, tc := range tests {
		result := pnr.Format(tc.cent, tc.dash)
		if result != tc.output {
			t.Errorf(util, i, tc.output, result)
		}
	}
}

func TestGetRandomTime(t *testing.T) {
	now := time.Now()
	year := time.Hour * 24 * 365
	from, to := year*100, year*20
	var last time.Time
	for i := 0; i < 20; i++ {
		tm := GetRandomTime(from, to)
		if tm == last {
			t.Error("Want different values, got ", tm, " and ", last)
		}
		if tm.Before(now.Add(-from)) {
			t.Error("Random ", tm, " should be after from ", now.Add(-from))
		}
		if !tm.Before(now.Add(-to)) {
			t.Error("Random ", tm, " should be before to ", now.Add(-to))
		}
		last = tm
	}
}

func TestGetChecksum(t *testing.T) {
	var tests = []SSN{
		{1, 9, 7, 5, 0, 9, 3, 0, 1, 9, 3, 8},
		{2, 0, 0, 9, 0, 3, 0, 1, 6, 6, 8, 1},
		{2, 0, 1, 1, 0, 5, 3, 0, 4, 9, 3, 3},
		{1, 9, 7, 2, 1, 1, 0, 1, 0, 5, 0, 4},
	}
	for i, tc := range tests {
		cs := GetChecksum(tc)
		exp := tc[11]
		if cs != exp {
			t.Errorf(util, i, cs, exp)
		}
	}
}

func TestSumDigits(t *testing.T) {
	var tests = []struct {
		in  int
		out int
	}{
		{
			121, 4,
		},
		{
			8, 8,
		},
		{
			19, 1,
		},
		{
			3564, 9,
		}}
	for i, tc := range tests {
		got := sumDigits(tc.in)
		if got != tc.out {
			t.Errorf(util, i, got, tc.out)
		}
	}
}

func TestNewSSNFromString(t *testing.T) {
	var tests = map[string]struct {
		input string
		ssn   *SSN
		err   error
	}{
		"Incorrect length": {
			"1975092-1938",
			nil,
			ErrFormat,
		},
		"Incorrect letters/symbols": {
			"198A0930-1938",
			nil,
			ErrFormat,
		},
		"Incorrect date": {
			"20101510-1234",
			nil,
			ErrDate,
		},
		"Incorrect checksum": {
			"20090301-6684",
			&SSN{2, 0, 0, 9, 0, 3, 0, 1, 6, 6, 8, 4},
			ErrChecksum,
		},
		"Correct SSN": {
			"20110530-4933",
			&SSN{2, 0, 1, 1, 0, 5, 3, 0, 4, 9, 3, 3},
			nil,
		},
	}
	for label, tc := range tests {
		t.Run(label, func(t *testing.T) {
			ssn, err := NewSSNFromString(tc.input)
			if (ssn == nil) == (tc.ssn == nil) {
				if ssn != nil {
					if *ssn != *tc.ssn {
						t.Errorf(util, "SSN values!", ssn, tc.ssn)
					}
				}
			} else {
				t.Errorf(util, "SSN types!", ssn, tc.ssn)
			}
			if err != tc.err {
				t.Errorf(util, "ERROR!", err, tc.err)
			}
		})
	}
}

func BenchmarkSSN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		from, to := time.Hour*24*365*80, time.Hour*24*365*18
		var ssn SSN
		ssn.SetDate(GetRandomTime(from, to))
		ssn.SetLastDigits("ss?c")
	}
}

func TestSetLastDigits(t *testing.T) {
	baseSSN := SSN{1, 9, 7, 5, 0, 9, 2, 2, 1, 2, 3, 4}
	tests := map[string]struct {
		input string
		test  func(n SSN) bool
	}{
		"Safe": {
			"ss?c",
			func(n SSN) bool {
				if n[8] != 9 {
					return false
				}
				switch n[9] {
				case 9, 8:
				default:
					return false
				}
				return true
			},
		},
		"Random": {
			"????",
			func(n SSN) bool {
				if n != baseSSN {
					return true
				}
				return false
			},
		},
		"Female": {
			"**f*",
			func(n SSN) bool {
				if n[10]%2 == 0 {
					return true
				}
				return false
			},
		},
		"Male": {
			"**m*",
			func(n SSN) bool {
				if n[10]%2 == 1 {
					return true
				}
				return false
			},
		},
		"Checksum": {
			"***c",
			func(n SSN) bool {
				if n[11] == GetChecksum(n) {
					return true
				}
				return false
			},
		},
		"Selective preservation": {
			"5*5*",
			func(n SSN) bool {
				s := n.String()
				return strings.HasSuffix(s, "5254")
			},
		},
		"Selective preservation 2": {
			"*5*5",
			func(n SSN) bool {
				s := n.String()
				return strings.HasSuffix(s, "1535")
			},
		},
	}
	for label, tc := range tests {
		t.Run(label, func(t *testing.T) {
			n := baseSSN
			n.SetLastDigits(tc.input)
			if !tc.test(n) {
				t.Error("Failed", n)
			}
		})
	}
}

func TestDate(t *testing.T) {
	tt := []struct {
		ssn   string
		year  int
		month time.Month
		day   int
	}{
		{
			"19530105-9894",
			1953,
			time.Month(1),
			5,
		},
		{
			"19970712-9855",
			1997,
			time.Month(7),
			12,
		},
	}
	for _, tc := range tt {
		ssn, err := NewSSNFromString(tc.ssn)
		if err != nil {
			t.Error("Could not parse SSN from string")
			t.FailNow()
		}
		y, m, d := ssn.Date()
		if y != tc.year {
			t.Error("Wrong year:", y, "want", tc.year)
		}
		if m != tc.month {
			t.Error("Wrong month:", m, "want", tc.month)
		}
		if d != tc.day {
			t.Error("Wrong day:", d, "want", tc.day)
		}
	}
}

func assert(got, want interface{}, t *testing.T) {
	if got != want {
		t.Errorf("Got %v, Want %v", got, want)
	}
}

func TestAge(t *testing.T) {
	referenceTime, _ := time.Parse("20060102", "20000101")
	const hoursPerYear = 24 * 365.25
	tt := []struct {
		ssn      string
		ageYears int
	}{{
		"19720508-9894",
		27,
	}, {
		"19370704-9858",
		62,
	}, {
		"19811115-9870",
		18,
	}, {
		"19490801-9815",
		50,
	}}
	for testNo, tc := range tt {
		t.Run(fmt.Sprintf("Running testage no %v", testNo), func(t *testing.T) {
			ssn, _ := NewSSNFromString(tc.ssn)
			age := int(ssn.Age(referenceTime).Hours() / hoursPerYear)
			assert(age, tc.ageYears, t)
		})
	}
}

func TestSSN_Female(t *testing.T) {
	tests := []struct {
		name string
		n    string
		want bool
	}{
		{
			name: "female",
			n:    "19720525-6600",
			want: true,
		}, {
			name: "male",
			n:    "19541014-1674",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssn, _ := NewSSNFromString(tt.n)
			if got := ssn.Female(); got != tt.want {
				t.Errorf("SSN.Female() = %v, want %v", got, tt.want)
			}
		})
	}
}
