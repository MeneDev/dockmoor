package main

import (
	"fmt"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/jessevdk/go-flags"
)

type listOptions struct {
	MatchingOptions
}

func ListOptionsNew(mainOptions *mainOptions) *listOptions {
	return &listOptions{
		MatchingOptions{
			mainOpts: mainOptions,
			matchHandler: func(r dockref.Reference) (string, error) {
				fmt.Fprintf(mainOptions.stdout, "%s\n", r.Original())
				return "", nil
			},
		},
	}
}

func addListCommand(mainOptions *mainOptions, adder func(opts *mainOptions, command string, shortDescription string, longDescription string, data interface{}) (*flags.Command, error)) (*flags.Command, error) {
	lo := ListOptionsNew(mainOptions)

	return adder(mainOptions, "list",
		"List image references with matching predicates.",
		"List image references with matching predicates. Returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions and is of valid format, non-null otherwise",
		lo)
}
