package composeparser

import (
	"strings"
)

type serviceOptions struct {
	ignore    bool
	onlyMinor bool
	onlyPatch bool
	warnMajor bool
	warnMinor bool
	warnPatch bool
	warnAll   bool
}

func newServiceOptions(headComment string, lineComment string) *serviceOptions {
	comment := headComment + lineComment

	return &serviceOptions{
		ignore:    containsOption(comment, "ignore"),
		onlyMinor: containsOption(comment, "minor"),
		onlyPatch: containsOption(comment, "patch"),
		warnMajor: containsOption(comment, "warnMajor"),
		warnMinor: containsOption(comment, "warnMinor"),
		warnPatch: containsOption(comment, "warnPatch"),
		warnAll:   containsOption(comment, "warnAll"),
	}
}

func containsOption(comment string, option string) bool {
	optionStr := "impose:" + option
	return strings.Contains(comment, optionStr)
}
