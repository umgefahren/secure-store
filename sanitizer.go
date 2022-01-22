package main

import (
	"errors"
	"regexp"
)

const regexExp = `^[a-z]+([-][a-z0-9]+)*[a-z]+$`

type Matcher struct {
	reg *regexp.Regexp
}

func NewMatcher() *Matcher {
	ret := new(Matcher)
	ret.reg = regexp.MustCompile(regexExp)
	return ret
}

func (m *Matcher) MatchString(inp string) bool {
	return m.reg.MatchString(inp)
}

func BucketIdMatchingError() error {
	return errors.New("bucket id doesn't matches the needed pattern")
}

func KeyIdMatchingError() error {
	return errors.New("key id doesn't matches the needed pattern")
}

func UrlKeyMatchingError() error {
	return errors.New("url key doesn't matches the needed pattern")
}
