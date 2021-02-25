package utils

import (
	"fmt"
	"time"
)

const (
	periodLayout = "2006-01-02"
)

func ParsePeriod(period []string) (*time.Time, *time.Time, error) {
	if len(period) != 2 {
		return nil, nil, fmt.Errorf("expected `from` and `to` period dates")
	}

	from, err := time.Parse(periodLayout, period[0])
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing `from` date: %s", err)
	}

	to, err := time.Parse(periodLayout, period[1])
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing `to` date: %s", err)
	}

	fromMin := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toMax := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 1e9-1, time.UTC)
	return &fromMin, &toMax, nil
}
