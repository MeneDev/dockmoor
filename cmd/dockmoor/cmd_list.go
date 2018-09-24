package main

import (
	"github.com/jessevdk/go-flags"
)

func addListCommand(mainOptions *mainOptions) (*flags.Command, error) {
	parser := mainOptions.Parser()
	var containsOptions MatchingOptions
	containsOptions.mainOpts = mainOptions
	containsOptions.mode = matchAndPrint

	return parser.AddCommand("list",
		"List image references with matching predicates.",
		"List image references with matching predicates. Returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions and is of valid format, non-null otherwise",
		&containsOptions)
}
