package main

import (
	"errors"
	"strings"
)

// Command-line Value that returns a map[string]string

type TagsValue map[string]string

func NewTagsValue(hostname string) TagsValue {
	return TagsValue(map[string]string{"hostname": hostname})
}

func (self TagsValue) Set(str string) error {
	s := strings.SplitN(str, "=", 2)
	if len(s) != 2 {
		return errors.New("Invalid format, must be k=v")
	}
	self[s[0]] = s[1]
	return nil
}

func (self TagsValue) String() string {
	return ""
}

func (self TagsValue) Value() map[string]string {
	return map[string]string(self)
}
