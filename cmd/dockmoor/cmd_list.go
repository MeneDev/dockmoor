package main

import (
	"github.com/jessevdk/go-flags"
)

func addListCommand(mainOptions *mainOptions, adder func (opts *mainOptions, command string, shortDescription string, longDescription string, data interface{}) (*flags.Command, error)) (*flags.Command, error) {
	var containsOptions MatchingOptions
	containsOptions.mainOpts = mainOptions
	containsOptions.mode = matchAndPrint

	return adder(mainOptions, "list",
		"List image references with matching predicates.",
		"List image references with matching predicates. Returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions and is of valid format, non-null otherwise",
		&containsOptions)
}
