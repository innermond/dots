package dots

import (
	"database/sql/driver"
	"errors"
	"regexp"
	"strconv"
	"time"
)

type PartialTime time.Time

func (pt PartialTime) Value() (driver.Value, error) {
	return driver.Value(time.Time(pt)), nil
}

func (pt *PartialTime) Scan(v interface{}) error {
	if v == nil {
		pt = (*PartialTime)(nil)
		return nil
	}

	t, err := time.Parse(time.RFC3339, v.(string))
	if err != nil {
		return err
	}

	*pt = PartialTime(t)
	return nil
}

func (pt *PartialTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" || s == `""` {
		return nil
	}

	t, err := parseTimeString(s)
	if err != nil {
		return err
	}

	*pt = PartialTime(*t)

	return nil
}

func parseTimeString(inputTimeStr string) (*time.Time, error) {
	// Define the input time string
	//inputTimeStr := "2023-04-13 14:01:45" or variations,
	//except year, which is mandatory

	pattern := `(\d{4})(-(\d{2}))?(-(\d{2}))?( (\d{2}))?(:(\d{2}))?(:(\d{2}))?`
	r := regexp.MustCompile(pattern)

	matches := r.FindStringSubmatch(inputTimeStr)
	if len(matches) == 0 {
		return nil, errors.New("error matching time string with regex pattern")
	}

	year, _ := strconv.Atoi(matches[1])
	// check for time fragments
	month := 1
	if matches[3] != "" {
		month, _ = strconv.Atoi(matches[3])
	}
	day := 1
	if matches[5] != "" {
		day, _ = strconv.Atoi(matches[5])
	}
	hour := 0
	if matches[7] != "" {
		hour, _ = strconv.Atoi(matches[7])
	}
	minute := 0
	if matches[9] != "" {
		minute, _ = strconv.Atoi(matches[9])
	}
	second := 0
	if matches[11] != "" {
		second, _ = strconv.Atoi(matches[11])
	}

	// Create a time value from the parsed components
	parsedTime := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)

	return &parsedTime, nil
}
