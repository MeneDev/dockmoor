package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/MeneDev/dockfix/dockproc"
	"io"
	"github.com/MeneDev/dockfix/dockfmt"
	"github.com/sirupsen/logrus"
)


type FindOptions struct {
	Any      bool     `required:"no" long:"any" description:"Find all images"`
	Latest   bool     `required:"no" long:"latest" description:"Using latest tag"`
	Unpinned bool     `required:"no" long:"unpinned" description:"Using unpinned images"`
	Outdated bool     `required:"no" long:"outdated" description:"Find all images with newer versions available"`
	Name     []string `required:"no" long:"name" description:"Find all images matching one of the specified names"`
	Domain   []string `required:"no" long:"domain" description:"Find all images matching one of the specified domains"`

	Positional struct {
		InputFile flags.Filename `required:"yes"`
	} `positional-args:"yes"`

	mainOptions *MainOptions
}

func (fo *FindOptions) MainOptions() *MainOptions {
	return fo.mainOptions
}

func init() {
	addFindCommand(&globalMainOptions)
}

func addFindCommand(mainOptions *MainOptions) (*flags.Command, error) {
	parser := mainOptions.Parser()
	var findOptions FindOptions
	findOptions.mainOptions = mainOptions

	return parser.AddCommand("find",
		"Test if a file contains image references with certain predicates.",
		"The find returns exit code 0 when the given input contains at least one image reference that satisfy the given conditions, non-null otherwise",
		&findOptions)
}

var (
	ERR_AT_LEAST_ONE_PREDICATE = errors.Errorf("Provide at least one predicate")
	ERR_AT_MOST_ONE_PREDICATE  = errors.Errorf("Provide at most one of --any, --latest, --unpinned, --outdated")
)

func verifyFindOptions(fo *FindOptions) error {

	if !fo.Any &&
		!fo.Latest &&
		!fo.Unpinned &&
		len(fo.Name) == 0 &&
		len(fo.Domain) == 0 &&
		!fo.Outdated {

		return ERR_AT_LEAST_ONE_PREDICATE
	}

	set := 0
	if fo.Any {
		set++
	}
	if fo.Latest {
		set++
	}
	if fo.Unpinned {
		set++
	}
	if fo.Outdated {
		set++
	}
	if set > 1 {
		return ERR_AT_MOST_ONE_PREDICATE
	}

	return nil
}

func (opts *FindOptions) Execute(args []string) error {
	panic("Use ExecuteWithExitCode instead")
}

func (opts *FindOptions) ExecuteWithExitCode(args []string) (exitCode int, err error) {
	err = verifyFindOptions(opts)
	if err != nil {
		opts.Log().Errorf("Invalid options: %s\n", err.Error())

		parser := flags.NewParser(&struct{}{}, flags.HelpFlag)
		command, _ := addFindCommand(opts.mainOptions)
		parser.ParseArgs([]string{command.Name, "--help"})

		parser.WriteHelp(opts.mainOptions.stdout)
		exitCode = EXIT_INVALID_PARAMS
		return
	}

	exitCode, err = opts.find()
	return
}

func (opts *FindOptions) getPredicate() dockproc.Predicate {
	if opts.Any {
		return dockproc.AnyPredicateNew()
	}

	return nil
}

func (opts *FindOptions) open(readable string) (io.ReadCloser, error) {
	return opts.mainOptions.readableOpener(readable)
}

func (opts *FindOptions) find() (exitCode int, err error) {
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
		exitCode = 1
		return
	}

	predicate := opts.getPredicate()

	fileFormat, formatError := dockfmt.IdentifyFormat(log, formatProvider, fpInput, filePathInput)
	if fileFormat == nil {
		return 1, formatError
	}

	formatProcessor := dockfmt.FormatProcessorNew(fileFormat, log, fpInput)

	accumulator, err := dockproc.FindAccumulatorNew(predicate)

	accumulator.Accumulate(formatProcessor)

	found := accumulator.Result()

	if found {
		exitCode = 0
	} else {
		exitCode = 1
	}

	return
}
func (opts *FindOptions) Log() *logrus.Logger {
	return opts.mainOptions.log
}
