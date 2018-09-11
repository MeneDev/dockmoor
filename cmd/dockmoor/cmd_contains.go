package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/MeneDev/dockmoor/dockproc"
	"io"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/sirupsen/logrus"
)


type ContainsOptions struct {
	Predicates struct{
		Any      bool     `required:"no" long:"any" description:"Matches all images"`
		Latest   bool     `required:"no" long:"latest" description:"Matches images with latest or no tag"`
		Unpinned bool     `required:"no" long:"unpinned" description:"Using unpinned images" hidden:"true"`
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
}

func (fo *ContainsOptions) MainOptions() *mainOptions {
	return fo.mainOptions
}

func addContainsCommand(mainOptions *mainOptions) (*flags.Command, error) {
	parser := mainOptions.Parser()
	var containsOptions ContainsOptions
	containsOptions.mainOptions = mainOptions

	return parser.AddCommand("contains",
		"Test if a file contains image references with certain predicates.",
		"Returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions, non-null otherwise",
		&containsOptions)
}

var (
	ERR_AT_LEAST_ONE_PREDICATE = errors.Errorf("Provide at least one predicate")
	ERR_AT_MOST_ONE_PREDICATE  = errors.Errorf("Provide at most one of --any, --latest, --unpinned, --outdated")
)

func verifyContainsOptions(fo *ContainsOptions) error {

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

func (opts *ContainsOptions) Execute(args []string) error {
	return errors.New("Use ExecuteWithExitCode instead")
}

func (opts *ContainsOptions) ExecuteWithExitCode(args []string) (exitCode ExitCode, err error) {
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

	exitCode, err = opts.find()
	return
}

func (opts *ContainsOptions) getPredicate() dockproc.Predicate {
	predicates := opts.Predicates
	if predicates.Any {
		return dockproc.AnyPredicateNew()
	}
	if predicates.Latest {
		return dockproc.LatestPredicateNew()
	}

	return nil
}

func (opts *ContainsOptions) open(readable string) (io.ReadCloser, error) {
	return opts.mainOptions.readableOpener(readable)
}

func (opts *ContainsOptions) find() (exitCode ExitCode, err error) {
	log := opts.Log()
	formatProvider := opts.MainOptions().FormatProvider()

	filePathInput := string(opts.Positional.InputFile)
	fpInput, err := opts.open(filePathInput)
	defer func() {
		if fpInput != nil {
			fpInput.Close()
		}
	}()
	if err != nil {
		log.Errorf("Could not open file: %s", err.Error())
		exitCode = EXIT_COULD_NOT_OPEN_FILE
		return
	}

	predicate := opts.getPredicate()

	fileFormat, formatError := dockfmt.IdentifyFormat(log, formatProvider, fpInput, filePathInput)
	if fileFormat == nil {
		return EXIT_INVALID_FORMAT, formatError
	}

	formatProcessor := dockfmt.FormatProcessorNew(fileFormat, log, fpInput)

	accumulator, err := dockproc.ContainsAccumulatorNew(predicate)

	errAcc := accumulator.Accumulate(formatProcessor)
	if errAcc != nil {
		log.Errorf("Error during accumulation: %s", errAcc.Error())
	}

	found := accumulator.Result()

	if found {
		exitCode = EXIT_SUCCESS
	} else {
		exitCode = EXIT_NOT_FOUND
	}

	return
}
func (opts *ContainsOptions) Log() *logrus.Logger {
	return opts.mainOptions.Log()
}
