package main

import (
	"errors"
	"github.com/sirupsen/logrus"
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
	logrus.Infoln("Provided bucket id, was faulty. Returned Error.")
	return errors.New("bucket id doesn't matches the needed pattern")
}

func KeyIdMatchingError() error {
	logrus.Infoln("Provided key id, was faulty. Returned Error.")
	return errors.New("key id doesn't matches the needed pattern")
}

func UrlKeyMatchingError() error {
	logrus.Infoln("Provided Url Key, was faulty. Returned Error.")
	return errors.New("url key doesn't matches the needed pattern")
}
