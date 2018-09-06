package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"fmt"
	"github.com/jessevdk/go-flags"
	"io"
	"bytes"
	"github.com/MeneDev/dockmoor/dockfmt"
	_ "github.com/MeneDev/dockmoor/dockfmt/dockerfile"
	"path"
	"strings"
)

type ExitCodeCommander interface {
	flags.Commander
	ExecuteWithExitCode(args []string) (exitCode ExitCode, err error)
}

type mainOptions struct {
	LogLevel      string `required:"no" short:"l" long:"log-level" description:"Sets the log-level" choice:"NONE" choice:"ERROR" choice:"WARN" choice:"INFO" choice:"DEBUG" default:"ERROR"`
	ShowVersion   bool   `required:"no" long:"version" description:"Show version and exit"`
	Manpage       bool   `required:"no" long:"manpage" description:"Show man page and exit"`
	Markdown      bool   `required:"no" long:"markdown" description:"Show usage as markdown and exit"`
	AsciiDocUsage bool   `required:"no" long:"asciidoc-usage" description:"Show usage as asciidoc and exit"`

	readableOpener func(string) (io.ReadCloser, error)
	parser         *flags.Parser
	log            *logrus.Logger
	formatProvider dockfmt.FormatProvider
	stdout         io.Writer
	stdin          io.ReadCloser
}

var osStdout io.Writer = os.Stdout
var osStdin io.ReadCloser = os.Stdin

func MainOptionsNew() *mainOptions {
	mainOptions := &mainOptions{}

	parser := flags.NewParser(mainOptions, flags.HelpFlag|flags.PassDoubleDash)
	mainOptions.parser = parser
	mainOptions.readableOpener = defaultReadableOpener(mainOptions)
	mainOptions.log = logrus.New()
	mainOptions.formatProvider = dockfmt.DefaultFormatProvider()
	mainOptions.stdout = osStdout
	mainOptions.stdin = osStdin

	return mainOptions
}

func (options *mainOptions) Parser() *flags.Parser {
	return options.parser
}
func (options *mainOptions) Log() *logrus.Logger {
	return options.log
}
func (options *mainOptions) FormatProvider() dockfmt.FormatProvider {
	return options.formatProvider
}
func (options *mainOptions) SetStdout(writer io.Writer) {
	options.stdout = writer
	options.log.SetOutput(writer)
}

func defaultReadableOpener(options *mainOptions) func(filename string) (io.ReadCloser, error) {
	return func(filename string) (io.ReadCloser, error) {
		if filename == "-" {
			return options.stdin, nil
		}
		return os.Open(filename)
	}
}

func doMain(mainOptions *mainOptions) (exitCode ExitCode) {
	readableOpener := defaultReadableOpener(mainOptions)
	mainOptions.readableOpener = readableOpener

	cmd, cmdArgs, exitCode := CommandFromArgs(mainOptions, os.Args[1:])

	if cmd != nil {
		commander := cmd.(ExitCodeCommander)
		exitCode, _ = commander.ExecuteWithExitCode(cmdArgs)
	}

	return
}

var osExit = func(exitCode ExitCode) { os.Exit(int(exitCode)) }

func main() {
	mainOptions := MainOptionsNew()
	addFindCommand(mainOptions)

	exitCode := doMain(mainOptions)
	osExit(exitCode)
}

func CommandFromArgs(mainOptions *mainOptions, args []string) (theCommand flags.Commander, cmdArgs []string, exitCode ExitCode) {
	parser := mainOptions.parser

	name := path.Base(os.Args[0])
	name = strings.Replace(name, "-linux_amd64", "", 1)
	parser.Name = name

	log := mainOptions.log

	parser.ShortDescription = "Manage docker image references."
	parser.LongDescription = "Manage docker image references."
	//parser.Usage = "Usage here"

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

	if mainOptions.Markdown {
		WriteMarkdown(parser, mainOptions.stdout)
		exitCode = EXIT_SUCCESS
		return
	}

	if mainOptions.AsciiDocUsage {
		WriteAsciiDoc(parser, mainOptions.stdout)
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

var Version string = "<unknown Version>"
var BuildDate string = "<unknown BuildDate>"
var BuildNumber string = "<unknown BuildNumber>"
var BuildCommit string = "<unknown BuildCommit>"

func WriteVersion(writer io.Writer) {
	format := "%-13s%s\n"
	fmt.Fprintf(writer, format, "Version:", Version)
	fmt.Fprintf(writer, format, "BuildDate:", BuildDate)
	fmt.Fprintf(writer, format, "BuildNumber:", BuildNumber)
	fmt.Fprintf(writer, format, "BuildCommit:", BuildCommit)
}
