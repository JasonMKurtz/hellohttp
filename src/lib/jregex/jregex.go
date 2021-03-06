package jregex

import (
	"regexp"
)

type JRegex struct {
	Exp, Haystack string
}

func IsMatch(exp, search string) bool {
	r := &JRegex{Exp: exp, Haystack: search}
	_, match := r.compAndMatch()
	return len(match) >= 1
}

func (e *JRegex) compAndMatch() (*regexp.Regexp, []string) {
	var r = regexp.MustCompile(e.Exp)
	match := r.FindStringSubmatch(e.Haystack)

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
