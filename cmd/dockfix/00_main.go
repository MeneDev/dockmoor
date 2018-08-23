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
	"path"
	"strings"
	"html"
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

type mainOptions struct {
	LogLevel    string `required:"no" short:"l" long:"log-level" description:"Sets the log-level" choice:"NONE" choice:"ERROR" choice:"WARN" choice:"INFO" choice:"DEBUG" default:"ERROR"`
	ShowVersion bool   `required:"no" long:"version" description:"Show version and exit"`
	Manpage     bool   `required:"no" long:"manpage" description:"Show man page and exit"`
	Markdown     bool   `required:"no" long:"markdown" description:"Show usage as markdown and exit"`

	readableOpener func(string) (io.ReadCloser, error)
	parser         *flags.Parser
	log            *logrus.Logger
	formatProvider dockfmt.FormatProvider
	stdout         io.Writer
	stdin          io.ReadCloser
}

func MainOptionsNew() *mainOptions {
	mainOptions := &mainOptions{}

	parser := flags.NewParser(mainOptions, flags.HelpFlag|flags.PassDoubleDash)
	mainOptions.parser = parser
	mainOptions.readableOpener = defaultReadableOpener(mainOptions)
	mainOptions.log = logrus.New()
	mainOptions.formatProvider = dockfmt.DefaultFormatProvider()
	mainOptions.stdout = os.Stdout
	mainOptions.stdin = os.Stdin

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

func doMain(mainOptions *mainOptions) (exitCode int) {
	readableOpener := defaultReadableOpener(mainOptions)
	mainOptions.readableOpener = readableOpener

	cmd, cmdArgs, exitCode := CommandFromArgs(mainOptions, os.Args[1:])

	if cmd != nil {
		commander := cmd.(ExitCodeCommander)
		exitCode, _ = commander.ExecuteWithExitCode(cmdArgs)
	}

	return
}

func main() {
	mainOptions := MainOptionsNew()
	addFindCommand(mainOptions)

	exitCode := doMain(mainOptions)
	os.Exit(exitCode)
}

func CommandFromArgs(mainOptions *mainOptions, args []string) (theCommand flags.Commander, cmdArgs []string, exitCode int) {
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

func mdPrintf(writer io.Writer, format string, a ...interface{}) (n int, err error) {
	for i, e := range a {
		if s, ok := e.(string); ok {
			s = html.EscapeString(s)
			a[i] = s
		}
	}

	return fmt.Fprintf(writer, format, a...)
}

func WriteMarkdown(parser *flags.Parser, writer io.Writer) {
	mdPrintf(writer, "# %s [![CircleCI](https://circleci.com/gh/MeneDev/dockfix.svg?style=shield)](https://circleci.com/gh/MeneDev/dockfix) [![Coverage Status](https://coveralls.io/repos/github/MeneDev/dockfix/badge.svg)](https://coveralls.io/github/MeneDev/dockfix)\n", parser.Name)
	mdPrintf(writer, "%s Version %s\n\n",  parser.Name, Version)
	mdPrintf(writer, "%s\n\n", parser.LongDescription)
	mdPrintf(writer, "## Usage\n")
	commands := []*flags.Command{parser.Command}
	WriteUsage(commands, writer)

	WriteGroups(writer, parser.Command.Groups(), 2)

	mdPrintf(writer, "## Commands\n\n")

	for _, cmd := range parser.Commands() {
		mdPrintf(writer, " * [%s](#%s)\n", cmd.Name, strings.ToLower(cmd.Name) + "-command")
	}
	mdPrintf(writer, "\n")

	for _, cmd := range parser.Commands() {
		mdPrintf(writer, "## %s command\n", cmd.Name)
		WriteUsage(append(commands, cmd), writer)
		mdPrintf(writer, "%s\n\n", cmd.LongDescription)
		WriteOptions(writer, cmd.Options(), 3)
		WriteGroups(writer, cmd.Groups(), 3)
	}
}

func WriteUsage(commands []*flags.Command, writer io.Writer) {

	mdPrintf(writer, "> ",)
	for idxCommand, command := range commands {

		isFirstCommand := idxCommand == 0
		isLastCommand := idxCommand+1 == len(commands)

		mdPrintf(writer, "%s", command.Name)
		if len(command.Options()) > 0 || len(command.Groups()) > 0 {
			if isFirstCommand {
				mdPrintf(writer, " \\[OPTIONS\\]")
			} else {
				mdPrintf(writer, " \\[%s-OPTIONS\\]", command.Name)
			}
		}

		if len(command.Args()) > 0 {
			for _, v := range command.Args() {
				var fmt string
				if v.Required == 0 {
					fmt = " \\[%s\\]"
				} else {
					fmt = " %s"
				}

				mdPrintf(writer, fmt, v.Name)
			}
		}

		if !isLastCommand {
			mdPrintf(writer, " ")
		} else {
			if len(command.Commands()) > 0 {
				var cmds []string
				for _, cmd := range command.Commands() {
					cmds = append(cmds, fmt.Sprintf("[%s](#%s)", cmd.Name, strings.ToLower(cmd.Name) + "-command"))
				}

				var fmt string
				if command.SubcommandsOptional {
					fmt = " \\[%s\\]"
				} else {
					fmt = " &lt;%s&gt;"
				}
				mdPrintf(writer, fmt, strings.Join(cmds, " | "))
				mdPrintf(writer, " \\[command-OPTIONS\\]")
			}
		}
	}

	mdPrintf(writer, "\n\n")
}

func WriteGroups(writer io.Writer, groups []*flags.Group, level int) {

	for _, group := range groups {
		mdPrintf(writer,  strings.Repeat("#", level) + " %s\n", group.ShortDescription)
		if group.LongDescription != "" {
			mdPrintf(writer,  "%s\n\n", group.LongDescription)
		}
		WriteOptions(writer, group.Options(), level + 1)
	}
}

func WriteOptions(writer io.Writer, options []*flags.Option, level int) {
	for _, opt := range options {
		if opt.Hidden { continue }

		if opt.OptionalArgument { continue }

		mdPrintf(writer, "**")
		var names []string
		if opt.ShortName != 0 {
			names = append(names, "-" + string(opt.ShortName))
		}
		if opt.LongNameWithNamespace() != "" {
			names = append(names, "--" + string(opt.LongNameWithNamespace()))
		}

		mdPrintf(writer, strings.Join(names, "**, **"))

		mdPrintf(writer, "**  \n%s", opt.Description)
		if opt.Choices != nil {
			mdPrintf(writer, " (one of `%s`)", strings.Join(opt.Choices, "`, `"))
		}
		mdPrintf(writer, "\n\n")
	}
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
