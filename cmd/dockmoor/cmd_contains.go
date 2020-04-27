package main

import (
	"errors"
	"io"

	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockproc"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/jessevdk/go-flags"
)

type containsOptions struct {
	MatchingOptions
	matches bool
}

func containsOptionsNew(mainOptions *mainOptions) *containsOptions {
	return &containsOptions{
		MatchingOptions: MatchingOptions{
			mainOpts: mainOptions,
		},
		matches: false,
	}
}

func addContainsCommand(mainOptions *mainOptions, adder func(opts *mainOptions, command string, shortDescription string, longDescription string, data interface{}) (*flags.Command, error)) (*flags.Command, error) {
	co := containsOptionsNew(mainOptions)

	return adder(mainOptions, "contains",
		"Test if a file contains image references with matching predicates.",
		"Test if a file contains image references with matching predicates. Returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions and is of valid format, non-null otherwise",
		co)
}

func (co *containsOptions) Execute(args []string) error {
	return errors.New("use ExecuteWithExitCode instead")
}

func (co *containsOptions) ExecuteWithExitCode(args []string) (exitCode ExitCode, err error) {
	// TODO code is redundant to other commands
	mopts := co.MatchingOptions

	exitCode, err = mopts.Verify()
	if err != nil {
		return
	}

	predicate, err := mopts.getPredicate()
	if err != nil {
		return ExitPredicateInvalid, err
	}

	err = mopts.WithInputDo(func(inputPath string, inputReader io.Reader) error {
		errFormat := mopts.WithFormatProcessorDo(inputReader, func(processor dockfmt.FormatProcessor) error {
			return co.applyFormatProcessor(predicate, processor)
		})
		return errFormat
	})

	if errExitCode, ok := exitCodeFromError(err); ok {
		return errExitCode, err
	}

	if co.matches {
		exitCode = ExitSuccess
	} else {
		exitCode = ExitNotFound
	}

	return exitCode, err
}

func (co *containsOptions) applyFormatProcessor(predicate dockproc.Predicate, processor dockfmt.FormatProcessor) error {
	return processor.Process(func(r dockref.Reference) (dockref.Reference, error) {
		if predicate.Matches(r) {
			co.matches = true
		}
		return r, nil
	})
}
