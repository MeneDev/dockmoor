package main

import (
	"github.com/jessevdk/go-flags"
)

func addListCommand(mainOptions *mainOptions) (*flags.Command, error) {
	parser := mainOptions.Parser()
	var containsOptions MatchingOptions
	containsOptions.mainOptions = mainOptions
	containsOptions.mode = MATCH_AND_PRINT

	return parser.AddCommand("list",
		"List image references with matching predicates.",
		"Returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions, non-null otherwise",
		&containsOptions)
}
