package main

import (
	"errors"
	"fmt"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockproc"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/jessevdk/go-flags"
	"io"
)

type listOptions struct {
	MatchingOptions
	matches bool
}

func listOptionsNew(mainOptions *mainOptions) *listOptions {
	return &listOptions{
		MatchingOptions: MatchingOptions{
			mainOpts: mainOptions,
		},
		matches: false,
	}
}

func addListCommand(mainOptions *mainOptions, adder func(opts *mainOptions, command string, shortDescription string, longDescription string, data interface{}) (*flags.Command, error)) (*flags.Command, error) {
	lo := listOptionsNew(mainOptions)

	return adder(mainOptions, "list",
		"List image references with matching predicates.",
		"List image references with matching predicates. Returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions and is of valid format, non-null otherwise",
		lo)
}

func (lo *listOptions) Execute(args []string) error {
	return errors.New("use ExecuteWithExitCode instead")
}

func (lo *listOptions) ExecuteWithExitCode(args []string) (exitCode ExitCode, err error) {
	// TODO code is redundant to other commands
	mopts := lo.MatchingOptions

	exitCode, err = mopts.Verify()
	if err != nil {
		return
	}

	predicate, err := mopts.getPredicate()
	if err != nil {
		return ExitPredicateInvalid, err
	}

	err = mopts.WithInputDo(func(inputPath string, inputReader io.Reader) error {
		return mopts.WithFormatProcessorDo(inputReader, func(processor dockfmt.FormatProcessor) error {
			return lo.applyFormatProcessor(predicate, processor)
		})
	})

	if errExitCode, ok := exitCodeFromError(err); ok {
		return errExitCode, err
	}

	if lo.matches {
		exitCode = ExitSuccess
	} else {
		exitCode = ExitNotFound
	}

	return exitCode, err
}

func (lo *listOptions) applyFormatProcessor(predicate dockproc.Predicate, processor dockfmt.FormatProcessor) error {

	return processor.Process(func(r dockref.Reference) (dockref.Reference, error) {
		if predicate.Matches(r) {
			lo.matches = true
			_, err := fmt.Fprintf(lo.Stdout(), "%s\n", r.Original())
			return r, err
		}
		return r, nil
	})
}
