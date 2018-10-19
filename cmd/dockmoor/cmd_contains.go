package main

import (
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/jessevdk/go-flags"
)

type containsOptions struct {
	MatchingOptions
}

func containsOptionsNew(mainOptions *mainOptions) *containsOptions {
	return &containsOptions{
		MatchingOptions{
			mainOpts: mainOptions,
			matchHandler: func(r dockref.Reference) (dockref.Reference, error) {
				return r, nil
			},
		},
	}
}

func addContainsCommand(mainOptions *mainOptions, adder func(opts *mainOptions, command string, shortDescription string, longDescription string, data interface{}) (*flags.Command, error)) (*flags.Command, error) {
	co := containsOptionsNew(mainOptions)

	return adder(mainOptions, "contains",
		"Test if a file contains image references with matching predicates.",
		"Test if a file contains image references with matching predicates. Returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions and is of valid format, non-null otherwise",
		co)
}
