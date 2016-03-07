package main

import (
	"strconv"
	"time"
)

type DurationValue struct {
	time.Duration
}

func NewDurationValue(d time.Duration) *DurationValue {
	return &DurationValue{d}
}

func (d *DurationValue) Set(s string) (err error) {
	// Interpret unit-less numbers as seconds
	if i, err2 := strconv.Atoi(s); err2 == nil {
		d.Duration = time.Duration(i) * time.Second
		return
	}

	x, err := time.ParseDuration(s)
	if err != nil {
		return
	}

	// Rounded at the second
	d.Duration = x - (x % time.Second)

	return
}

func (d *DurationValue) String() string {
	return d.Duration.String()
}

func (d *DurationValue) Value() time.Duration {
	return d.Duration
}
