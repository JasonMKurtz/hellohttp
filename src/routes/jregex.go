package main

import (
	"regexp"
)

type JRegex struct {
	exp, haystack string
}

func (e *JRegex) compAndMatch() (*regexp.Regexp, []string) {
	var r = regexp.MustCompile(e.exp)
	match := r.FindStringSubmatch(e.haystack)

	return r, match
}

func (e *JRegex) GetNamedGroups() map[string]string {
	r, match := e.compAndMatch()

	paramsMap := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	return paramsMap
}

func (e *JRegex) GetGroups() []string {
	r, match := e.compAndMatch()

	var params []string
	for i, _ := range r.SubexpNames() {
		if i > 0 && i <= len(match) {
			params = append(params, match[i])
		}
	}

	return params
}
