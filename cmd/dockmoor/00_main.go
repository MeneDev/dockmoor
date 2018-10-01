package main

import (
	"bytes"
	"fmt"
	"github.com/MeneDev/dockmoor/dockfmt"
	_ "github.com/MeneDev/dockmoor/dockfmt/dockerfile"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ExitCodeCommander interface {
	flags.Commander
	ExecuteWithExitCode(args []string) (exitCode ExitCode, err error)
}

type mainOptions struct {
	LogLevel    string `required:"no" short:"l" long:"log-level" description:"Sets the log-level" choice:"NONE" choice:"ERROR" choice:"WARN" choice:"INFO" choice:"DEBUG" default:"WARN"`
	ShowVersion bool   `required:"no" long:"version" description:"Show version and exit"`

	Help struct {
		Help          bool `short:"h" long:"help" description:"Show help and exit"`
		Manpage       bool `required:"no" long:"manpage" description:"Show man page and exit"`
		Markdown      bool `required:"no" long:"markdown" description:"Show usage as markdown and exit"`
		ASCIIDocUsage bool `required:"no" long:"asciidoc-usage" description:"Show usage as AsciiDoc and exit"`
	} `group:"Help Options" description:"Help Options"`

	readableOpener func(string) (io.ReadCloser, error)
	parser         *flags.Parser
	log            *logrus.Logger
	formatProvider dockfmt.FormatProvider
	stdout         io.Writer
	stdin          io.ReadCloser
}

var osStdout io.Writer = os.Stdout
var osStdin io.ReadCloser = os.Stdin

func mainOptionsNew() *mainOptions {
	mainOptions := &mainOptions{}

	parser := flags.NewParser(mainOptions, flags.PassDoubleDash)
	mainOptions.parser = parser
	mainOptions.readableOpener = defaultReadableOpener(mainOptions)
	log := logrus.New()
	mainOptions.log = log
	mainOptions.formatProvider = dockfmt.DefaultFormatProvider()
	mainOptions.stdout = osStdout
	mainOptions.stdin = osStdin
	log.SetOutput(osStdout)

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
		return os.Open(filepath.Clean(filename))
	}
}

func doMain(mainOptions *mainOptions) (exitCode ExitCode) {
	readableOpener := defaultReadableOpener(mainOptions)
	mainOptions.readableOpener = readableOpener

	cmd, cmdArgs, exitCode := CommandFromArgs(mainOptions, os.Args[1:])

	if cmd != nil {
		commander := cmd.(ExitCodeCommander)
		eC, err := commander.ExecuteWithExitCode(cmdArgs)
		exitCode = eC
		if err != nil {
			// This is only debug level because the loging should take place at a more informed place
			mainOptions.Log().Debug("Error: %s", err.Error())
		}
	}

	return
}

var osExitInternal = os.Exit

func osExit(exitCode ExitCode) { osExitInternal(int(exitCode)) }


var AddCommand = func (opts *mainOptions, command string, shortDescription string, longDescription string, data interface{}) (*flags.Command, error) {
	return opts.Parser().AddCommand(command, shortDescription, longDescription, data)
}

func main() {
	mainOptions := mainOptionsNew()
	log := mainOptions.Log()
	if _, err := addContainsCommand(mainOptions, AddCommand); err != nil {
		log.Errorf("Could not add contains command: %s", err)
	}

	if _, err := addListCommand(mainOptions, AddCommand); err != nil {
		log.Errorf("Could not add list command: %s", err)
	}

	exitCode := doMain(mainOptions)
	osExit(exitCode)
}

func handleHelp(mainOptions *mainOptions) bool {
	parser := mainOptions.parser
	if mainOptions.Help.Help {
		parser.WriteHelp(mainOptions.stdout)
		return true
	}

	if mainOptions.Help.Markdown {
		WriteMarkdown(parser, mainOptions.stdout)
		return true
	}

	if mainOptions.Help.ASCIIDocUsage {
		WriteASCIIDoc(parser, mainOptions.stdout)
		return true
	}

	if mainOptions.Help.Manpage {
		parser.WriteManPage(mainOptions.stdout)
		return true
	}

	return false
}

func handleVersion(mainOptions *mainOptions) bool {
	log := mainOptions.Log()
	if mainOptions.ShowVersion {
		WriteVersion(log, mainOptions.stdout)
		return true
	}

	return false
}

func CommandFromArgs(mainOptions *mainOptions, args []string) (theCommand flags.Commander, cmdArgs []string, exitCode ExitCode) {
	parser := mainOptions.parser

	name := path.Base(os.Args[0])
	name = strings.Replace(name, "-linux_amd64", "", 1)
	parser.Name = name

	log := mainOptions.log

	parser.ShortDescription = "Manage docker image references."
	parser.LongDescription = "Manage docker image references."

	parser.CommandHandler = func(command flags.Commander, args []string) error {
		theCommand = command
		return nil
	}

	cmdArgs, optsErr := parser.ParseArgs(args)

	if handleHelp(mainOptions) {
		exitCode = ExitSuccess
		return
	}

	if handleVersion(mainOptions) {
		exitCode = ExitSuccess
		return
	}

	level := logrus.WarnLevel
	log.SetLevel(level)
	if mainOptions.LogLevel == "NONE" {
		log.SetOutput(bytes.NewBuffer(nil))
	} else {
		level, err := logrus.ParseLevel(mainOptions.LogLevel)
		if err != nil {
			log.Errorf("Error parsing log-level '%s': %s", mainOptions.LogLevel, err.Error())
			exitCode = ExitInvalidParams
			return
		} else {
			log.SetLevel(level)
		}
	}

	if optsErr != nil {
		log.Errorf("Error in parameters: %s", optsErr)
		exitCode = ExitInvalidParams
		return
	}

	if len(parser.Commands()) == 0 {
		log.Error("No Command registered")
	}

	if theCommand == nil {
		log.Error("No Command specified")
		exitCode = ExitInvalidParams
		return
	}

	exitCode = ExitSuccess
	return
}

var Version = "<unknown Version>"
var BuildDate = "<unknown BuildDate>"
var BuildNumber = "<unknown BuildNumber>"
var BuildCommit = "<unknown BuildCommit>"

func WriteVersion(log *logrus.Logger, writer io.Writer) {
	format := "%-13s%s\n"
	fmtFprintf(log, writer, format, "Version:", Version)
	fmtFprintf(log, writer, format, "BuildDate:", BuildDate)
	fmtFprintf(log, writer, format, "BuildNumber:", BuildNumber)
	fmtFprintf(log, writer, format, "BuildCommit:", BuildCommit)
}

func fmtFprintf(log *logrus.Logger, w io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		log.Errorf("Error printing: %s", err.Error())
	}
}
