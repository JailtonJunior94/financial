package vos_test

import (
	"testing"
	"time"

	"github.com/jailtonjunior94/financial/pkg/domain/vos"
)

func TestNewReferenceMonth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   string
		wantErr bool
		wantStr string
	}{
		{name: "valid jan", input: "2024-01", wantStr: "2024-01"},
		{name: "valid dec", input: "2024-12", wantStr: "2024-12"},
		{name: "valid feb leap year", input: "2024-02", wantStr: "2024-02"},
		{name: "year boundary 2000", input: "2000-06", wantStr: "2000-06"},
		{name: "empty string", input: "", wantErr: true},
		{name: "invalid month 13", input: "2024-13", wantErr: true},
		{name: "invalid month 00", input: "2024-00", wantErr: true},
		{name: "wrong format YYYY/MM", input: "2024/01", wantErr: true},
		{name: "wrong format DD-MM-YYYY", input: "01-01-2024", wantErr: true},
		{name: "partial input", input: "2024", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rm, err := vos.NewReferenceMonth(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("NewReferenceMonth(%q): expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewReferenceMonth(%q): unexpected error: %v", tc.input, err)
			}
			if rm.String() != tc.wantStr {
				t.Errorf("NewReferenceMonth(%q).String() = %q, want %q", tc.input, rm.String(), tc.wantStr)
			}
		})
	}
}

func TestNewReferenceMonthFromYearMonth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		year    int
		month   int
		wantErr bool
		wantStr string
	}{
		{name: "valid", year: 2024, month: 6, wantStr: "2024-06"},
		{name: "january", year: 2025, month: 1, wantStr: "2025-01"},
		{name: "december", year: 2025, month: 12, wantStr: "2025-12"},
		{name: "year too low", year: 1899, month: 1, wantErr: true},
		{name: "year too high", year: 2101, month: 1, wantErr: true},
		{name: "month zero", year: 2024, month: 0, wantErr: true},
		{name: "month 13", year: 2024, month: 13, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rm, err := vos.NewReferenceMonthFromYearMonth(tc.year, tc.month)
			if tc.wantErr {
				if err == nil {
					t.Errorf("NewReferenceMonthFromYearMonth(%d, %d): expected error, got nil", tc.year, tc.month)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewReferenceMonthFromYearMonth(%d, %d): unexpected error: %v", tc.year, tc.month, err)
			}
			if rm.String() != tc.wantStr {
				t.Errorf("got %q, want %q", rm.String(), tc.wantStr)
			}
		})
	}
}

func TestNewReferenceMonthFromDate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		date    time.Time
		wantStr string
	}{
		{name: "mid month", date: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC), wantStr: "2024-06"},
		{name: "first day", date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), wantStr: "2024-01"},
		{name: "last day dec", date: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC), wantStr: "2024-12"},
		{name: "feb 29 leap", date: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), wantStr: "2024-02"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rm := vos.NewReferenceMonthFromDate(tc.date)
			if rm.String() != tc.wantStr {
				t.Errorf("NewReferenceMonthFromDate(%s) = %q, want %q",
					tc.date.Format("2006-01-02"), rm.String(), tc.wantStr)
			}
			if rm.Year() != tc.date.Year() {
				t.Errorf("Year() = %d, want %d", rm.Year(), tc.date.Year())
			}
			if rm.Month() != tc.date.Month() {
				t.Errorf("Month() = %v, want %v", rm.Month(), tc.date.Month())
			}
		})
	}
}

func TestReferenceMonthEqual(t *testing.T) {
	t.Parallel()

	a, _ := vos.NewReferenceMonth("2024-06")
	b, _ := vos.NewReferenceMonth("2024-06")
	c, _ := vos.NewReferenceMonth("2024-07")

	if !a.Equal(b) {
		t.Error("Equal: same month should be equal")
	}
	if a.Equal(c) {
		t.Error("Equal: different months should not be equal")
	}
}

func TestReferenceMonthFirstAndLastDay(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		rm           string
		wantFirst    string
		wantLastDay  int
	}{
		{name: "jan 31 days", rm: "2024-01", wantFirst: "2024-01-01", wantLastDay: 31},
		{name: "feb leap 29 days", rm: "2024-02", wantFirst: "2024-02-01", wantLastDay: 29},
		{name: "feb non-leap 28 days", rm: "2023-02", wantFirst: "2023-02-01", wantLastDay: 28},
		{name: "apr 30 days", rm: "2024-04", wantFirst: "2024-04-01", wantLastDay: 30},
		{name: "dec 31 days", rm: "2024-12", wantFirst: "2024-12-01", wantLastDay: 31},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rm, _ := vos.NewReferenceMonth(tc.rm)

			first := rm.FirstDay()
			if first.Format("2006-01-02") != tc.wantFirst {
				t.Errorf("FirstDay() = %s, want %s", first.Format("2006-01-02"), tc.wantFirst)
			}

			last := rm.LastDay()
			// LastDay returns last second of the month; check the day
			if last.Day() != tc.wantLastDay {
				t.Errorf("LastDay().Day() = %d, want %d", last.Day(), tc.wantLastDay)
			}
		})
	}
}

func TestReferenceMonthAddMonths(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		start  string
		add    int
		want   string
	}{
		{name: "+1 month", start: "2024-01", add: 1, want: "2024-02"},
		{name: "+11 months (year boundary)", start: "2024-02", add: 11, want: "2025-01"},
		{name: "+12 months", start: "2024-01", add: 12, want: "2025-01"},
		{name: "+0 months", start: "2024-06", add: 0, want: "2024-06"},
		{name: "-1 month", start: "2024-03", add: -1, want: "2024-02"},
		{name: "-2 months (year boundary)", start: "2024-01", add: -2, want: "2023-11"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rm, _ := vos.NewReferenceMonth(tc.start)
			got := rm.AddMonths(tc.add)
			if got.String() != tc.want {
				t.Errorf("AddMonths(%d): got %s, want %s", tc.add, got.String(), tc.want)
			}
		})
	}
}

func TestReferenceMonthToTime(t *testing.T) {
	t.Parallel()

	rm, _ := vos.NewReferenceMonth("2024-06")
	tt := rm.ToTime()
	first := rm.FirstDay()

	if !tt.Equal(first) {
		t.Errorf("ToTime() = %v, want %v (same as FirstDay)", tt, first)
	}
}
