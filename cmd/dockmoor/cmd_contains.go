package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/MeneDev/dockmoor/dockproc"
	"io"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/sirupsen/logrus"
	"fmt"
)

type MatchingMode int

const (
	MATCH_ONLY MatchingMode = iota
	MATCH_AND_PRINT MatchingMode = iota
)

type MatchingOptions struct {
	Predicates struct{
		Any      bool     `required:"no" long:"any" description:"Matches all images"`
		Latest   bool     `required:"no" long:"latest" description:"Matches images with latest or no tag"`
		Unpinned bool     `required:"no" long:"unpinned" description:"Matches unpinned images"`
		Outdated bool     `required:"no" long:"outdated" description:"Matches all images with newer versions available" hidden:"true"`
	} `group:"Predicates" description:"Specify which kind of image references should be selected. Exactly one must be specified"`

	Filters struct {
		Name   []string `required:"no" long:"name" description:"Matches all images matching one of the specified names" hidden:"true"`
		Domain []string `required:"no" long:"domain" description:"Matches all images matching one of the specified domains" hidden:"true"`
	} `group:"Filters" description:"Optional additional filters. Specifying each kind of filter must be matched at least once" hidden:"true"`
	
	Positional struct {
		InputFile flags.Filename `required:"yes"`
	} `positional-args:"yes"`

	mainOptions *mainOptions
	mode MatchingMode
}

func (fo *MatchingOptions) MainOptions() *mainOptions {
	return fo.mainOptions
}

func addContainsCommand(mainOptions *mainOptions) (*flags.Command, error) {
	parser := mainOptions.Parser()
	var containsOptions MatchingOptions
	containsOptions.mainOptions = mainOptions
	containsOptions.mode = MATCH_ONLY

	return parser.AddCommand("contains",
		"Test if a file contains image references with matching predicates.",
		"Test if a file contains image references with matching predicates. Returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions and is of valid format, non-null otherwise",
		&containsOptions)
}

var (
	ERR_AT_LEAST_ONE_PREDICATE = errors.Errorf("Provide at least one predicate")
	ERR_AT_MOST_ONE_PREDICATE  = errors.Errorf("Provide at most one of --any, --latest, --unpinned, --outdated")
)

func verifyContainsOptions(fo *MatchingOptions) error {

	p := fo.Predicates
	f := fo.Filters
	if !p.Any &&
		!p.Latest &&
		!p.Unpinned &&
		len(f.Name) == 0 &&
		len(f.Domain) == 0 &&
		!p.Outdated {

		return ERR_AT_LEAST_ONE_PREDICATE
	}

	set := 0
	if p.Any {
		set++
	}
	if p.Latest {
		set++
	}
	if p.Unpinned {
		set++
	}
	if p.Outdated {
		set++
	}
	if set > 1 {
		return ERR_AT_MOST_ONE_PREDICATE
	}

	return nil
}

func (opts *MatchingOptions) Execute(args []string) error {
	return errors.New("Use ExecuteWithExitCode instead")
}

func (opts *MatchingOptions) ExecuteWithExitCode(args []string) (exitCode ExitCode, err error) {
	err = verifyContainsOptions(opts)
	if err != nil {
		opts.Log().Errorf("Invalid options: %s\n", err.Error())

		parser := flags.NewParser(&struct{}{}, flags.HelpFlag)
		command, _ := addContainsCommand(opts.mainOptions)
		parser.ParseArgs([]string{command.Name, "--help"})

		parser.WriteHelp(opts.mainOptions.stdout)
		exitCode = EXIT_INVALID_PARAMS
		return
	}

	exitCode, err = opts.match()
	return
}

func (opts *MatchingOptions) getPredicate() dockproc.Predicate {
	predicates := opts.Predicates

	switch {
	case predicates.Any:
		return dockproc.AnyPredicateNew()
	case predicates.Latest:
		return dockproc.LatestPredicateNew()
	case predicates.Unpinned:
		return dockproc.UnpinnedPredicateNew()
	}

	return nil
}

func (opts *MatchingOptions) open(readable string) (io.ReadCloser, error) {
	return opts.mainOptions.readableOpener(readable)
}

func saveClose(readCloser io.ReadCloser) {
	if readCloser != nil {
		readCloser.Close()
	}
}

func (opts *MatchingOptions) match() (exitCode ExitCode, err error) {
	log := opts.Log()

	filePathInput := string(opts.Positional.InputFile)

	fpInput, err := opts.open(filePathInput)
	defer saveClose(fpInput)

	if err != nil {
		log.Errorf("Could not open file: %s", err.Error())
		exitCode = EXIT_COULD_NOT_OPEN_FILE
		return
	}

	formatProvider := opts.MainOptions().FormatProvider()
	fileFormat, formatError := dockfmt.IdentifyFormat(log, formatProvider, fpInput, filePathInput)
	if fileFormat == nil {
		return EXIT_INVALID_FORMAT, formatError
	}

	formatProcessor := dockfmt.FormatProcessorNew(fileFormat, log, fpInput)
	exitCode, err = opts.matchFormatProcessor(formatProcessor)
	return
}

func (opts *MatchingOptions) matchFormatProcessor(formatProcessor dockfmt.FormatProcessor) (exitCode ExitCode, err error) {
	log := opts.Log()

	predicate := opts.getPredicate()
	accumulator, err := dockproc.MatchesAccumulatorNew(predicate, log, opts.Stdout())

	errAcc := accumulator.Accumulate(formatProcessor)
	if errAcc != nil {
		log.Errorf("Error during accumulation: %s", errAcc.Error())
	}

	matches := accumulator.Matches()

	if len(matches) > 0 {
		exitCode = EXIT_SUCCESS
	} else {
		exitCode = EXIT_NOT_FOUND
	}

	if opts.mode == MATCH_AND_PRINT {
		for _, r := range matches {
			fmt.Fprintf(opts.Stdout(), "%s\n", r.Original())
		}
	}
	return
}

func (opts *MatchingOptions) Log() *logrus.Logger {
	return opts.mainOptions.Log()
}

func (fo *MatchingOptions) Stdout() io.Writer {
	return fo.MainOptions().stdout
}
