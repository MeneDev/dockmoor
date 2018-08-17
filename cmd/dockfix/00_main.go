package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"fmt"
	"github.com/jessevdk/go-flags"
	"io"
	"bytes"
	"github.com/MeneDev/dockfix/dockfmt"
	_ "github.com/MeneDev/dockfix/dockfmt/dockerfile"
)

const (
	EXIT_SUCCESS int = iota
	EXIT_INVALID_PARAMS
	EXIT_UNKNOWN_ERROR
)

type ExitCodeCommander interface {
	flags.Commander
	ExecuteWithExitCode(args []string) (exitCode int, err error)
}

type OutputFile struct {
	Filename flags.Filename `description:"Input file" default:"-"`
}

type MainOptions struct {
	LogLevel    string `required:"no" short:"l" long:"log-level" description:"Sets the log-level" choice:"NONE" choice:"ERROR" choice:"WARN" choice:"INFO" choice:"DEBUG" default:"ERROR"`
	ShowVersion bool   `required:"no" long:"version" description:"Show version and exit"`
	Manpage     bool   `required:"no" long:"manpage" description:"Show man page and exit"`

	readableOpener func(string) (io.ReadCloser, error)
	parser         *flags.Parser
	log            *logrus.Logger
	formatProvider dockfmt.FormatProvider
	stdout         io.Writer
	stdin          io.ReadCloser
}

func (options *MainOptions) Parser() *flags.Parser {
	return options.parser
}
func (options *MainOptions) Log() *logrus.Logger {
	return options.log
}
func (options *MainOptions) FormatProvider() dockfmt.FormatProvider {
	return options.formatProvider
}
func (options *MainOptions) SetStdout(writer io.Writer) {
	options.stdout = writer
	options.log.SetOutput(writer)
}

var globalMainOptions MainOptions
var globalParser = flags.NewParser(&globalMainOptions, flags.HelpFlag|flags.PassDoubleDash)

func init() {
	globalMainOptions.parser = globalParser
	globalMainOptions.formatProvider = dockfmt.DefaultFormatProvider()
}

func defaultReadableOpener(options *MainOptions) func(filename string) (io.ReadCloser, error) {
	return func(filename string) (io.ReadCloser, error) {
		if filename == "-" {
			return options.stdin, nil
		}
		return os.Open(filename)
	}
}

func main() {

	exitCode := func() (exitCode int) {
		// defers will not be run when using os.Exit, so wrap in function to ensure writer.Flush()
		log := logrus.New()

		globalMainOptions.log = log
		writer := os.Stdout
		globalMainOptions.SetStdout(writer)
		globalMainOptions.stdin = os.Stdin

		readableOpener := defaultReadableOpener(&globalMainOptions)
		globalMainOptions.readableOpener = readableOpener

		cmd, cmdArgs, exitCode := doMain(&globalMainOptions, os.Args[1:])

		if cmd != nil {
			commander := cmd.(ExitCodeCommander)
			exitCode, _ = commander.ExecuteWithExitCode(cmdArgs)
		}

		return
	}()

	os.Exit(exitCode)
}

func doMain(mainOptions *MainOptions, args []string) (theCommand flags.Commander, cmdArgs []string, exitCode int) {

	log := mainOptions.log

	parser := mainOptions.parser

	parser.CommandHandler = func(command flags.Commander, args []string) error {
		theCommand = command
		return nil
	}

	cmdArgs, optsErr := parser.ParseArgs(args)

	if flagsErr, ok := optsErr.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
		parser.WriteHelp(mainOptions.stdout)
		exitCode = EXIT_SUCCESS
		return
	}

	if mainOptions.Manpage {
		parser.WriteManPage(mainOptions.stdout)
		exitCode = EXIT_SUCCESS
		return
	}

	if mainOptions.ShowVersion {
		WriteVersion(mainOptions.stdout)
		exitCode = EXIT_SUCCESS
		return
	}

	if mainOptions.LogLevel == "NONE" {
		log.SetOutput(bytes.NewBuffer(nil))
	} else {
		level := logrus.ErrorLevel
		level, _ = logrus.ParseLevel(mainOptions.LogLevel)
		log.SetLevel(level)
	}

	if optsErr != nil {
		log.Errorf("Error in parameters: %s", optsErr)
		exitCode = EXIT_INVALID_PARAMS
		return
	}

	if len(parser.Commands()) == 0 {
		log.Error("No Command registered")
	}

	if theCommand == nil {
		log.Error("No Command specified")
		exitCode = EXIT_INVALID_PARAMS
		return
	}

	exitCode = EXIT_SUCCESS
	return
}

var Version string = "<unknown>"
var BuildDate string = "<unknown>"
var BuildNumber string = "<unknown>"
var BuildCommit string = "<unknown>"

func WriteVersion(writer io.Writer) {
	format := "%-13s%s\n"
	fmt.Fprintf(writer, format, "Version:", Version)
	fmt.Fprintf(writer, format, "BuildDate:", BuildDate)
	fmt.Fprintf(writer, format, "BuildNumber:", BuildNumber)
	fmt.Fprintf(writer, format, "BuildCommit:", BuildCommit)
}
