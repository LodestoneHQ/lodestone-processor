package model

import (
	"github.com/sabhiram/go-gitignore"
)

type Filter struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`

	incMatcher *ignore.GitIgnore
	excMatcher *ignore.GitIgnore
}

func (flt *Filter) ValidPath(pathToTest string) bool {

	if flt.incMatcher == nil {
		incMatcher, err := ignore.CompileIgnoreLines(flt.Include...)
		if err != nil {
			return true
		}
		flt.incMatcher = incMatcher
	}
	if flt.excMatcher == nil {
		excMatcher, err := ignore.CompileIgnoreLines(flt.Exclude...)
		if err != nil {
			return true
		}
		flt.excMatcher = excMatcher
	}

	//includes always override excludes
	matched := flt.incMatcher.MatchesPath(pathToTest)
	if matched {
		return true //if the file matches a pattern in "include", we need to tell the processor to continue
	}

	//test excludes next
	matched = flt.excMatcher.MatchesPath(pathToTest)
	if matched {
		return false //if the file matches a pattern in "exclude", we need to tell the processor to skip it
	}

	//by default, we'll include all files
	return true
}
