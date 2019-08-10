package main

import (
	"bytes"
	"fmt"
	"github.com/MeneDev/dockmoor/dockmoor"
	"github.com/jessevdk/go-flags"
	"html"
	"io"
	"strings"
)

func visibleCommands(commands []*flags.Command) []*flags.Command {
	visible := make([]*flags.Command, 0)
	for _, c := range commands {
		if !c.Hidden {
			visible = append(visible, c)
		}
	}
	return visible
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
	writer.Write(bytes.NewBufferString("[![CircleCI](https://img.shields.io/circleci/project/github/MeneDev/dockmoor/master.svg)](https://circleci.com/gh/MeneDev/dockmoor) [![Coverage Status](https://img.shields.io/coveralls/github/MeneDev/dockmoor/master.svg)](https://coveralls.io/github/MeneDev/dockmoor) [![Follow MeneDev on Twitter](https://img.shields.io/twitter/follow/MeneDev.svg?style=social&label=%40MeneDev)](https://twitter.com/MeneDev)\n").Bytes())
	mdPrintf(writer, "# %s ", parser.Name)
	mdPrintf(writer, "%s Version %s\n\n", parser.Name, dockmoor.Version)
	mdPrintf(writer, "%s\n\n", parser.LongDescription)
	mdPrintf(writer, "## Usage\n")
	commands := visibleCommands([]*flags.Command{parser.Command})
	WriteMarkDownUsage(commands, writer)

	WriteMarkdownGroups(writer, parser.Command.Groups(), 2)

	mdPrintf(writer, "## Commands\n\n")

	for _, cmd := range visibleCommands(parser.Commands()) {
		mdPrintf(writer, " * [%s](#%s)\n", cmd.Name, strings.ToLower(cmd.Name)+"-command")
	}
	mdPrintf(writer, "\n")

	for _, cmd := range visibleCommands(parser.Commands()) {
		mdPrintf(writer, "## %s command\n", cmd.Name)
		WriteMarkDownUsage(append(commands, cmd), writer)
		mdPrintf(writer, "%s\n\n", cmd.LongDescription)
		WriteMarkdownOptions(writer, cmd.Options(), 3)
		WriteMarkdownGroups(writer, cmd.Groups(), 3)
	}
}

func WriteMarkDownUsage(commands []*flags.Command, writer io.Writer) {

	commands = visibleCommands(commands)

	mdPrintf(writer, "> ")
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
		} else if len(visibleCommands(command.Commands())) > 0 {
			var cmds []string
			for _, cmd := range visibleCommands(command.Commands()) {
				cmds = append(cmds, fmt.Sprintf("[%s](#%s)", cmd.Name, strings.ToLower(cmd.Name)+"-command"))
			}

			var fmt string
			if command.SubcommandsOptional {
				fmt = " [%s]"
			} else {
				fmt = " &lt;%s&gt;"
			}
			mdPrintf(writer, fmt, strings.Join(cmds, " | "))
			mdPrintf(writer, " \\[command-OPTIONS\\]")
		}
	}

	mdPrintf(writer, "\n\n")
}

func WriteMarkdownGroups(writer io.Writer, groups []*flags.Group, level int) {

	for _, group := range groups {
		if group.Hidden {
			continue
		}
		mdPrintf(writer, strings.Repeat("#", level)+" %s\n", group.ShortDescription)
		if group.LongDescription != "" {
			mdPrintf(writer, "%s\n\n", group.LongDescription)
		}
		WriteMarkdownOptions(writer, group.Options(), level+1)
	}
}

func WriteMarkdownOptions(writer io.Writer, options []*flags.Option, level int) {
	for _, opt := range options {
		if opt.Hidden {
			continue
		}

		if opt.OptionalArgument {
			continue
		}

		mdPrintf(writer, "**")
		var names []string
		if opt.ShortName != 0 {
			names = append(names, "-"+string(opt.ShortName))
		}
		if opt.LongNameWithNamespace() != "" {
			names = append(names, "--"+opt.LongNameWithNamespace())
		}

		mdPrintf(writer, strings.Join(names, "**, **"))

		mdPrintf(writer, "**  \n%s", opt.Description)
		if opt.Choices != nil {
			mdPrintf(writer, " (one of `%s`)", strings.Join(opt.Choices, "`, `"))
		}
		mdPrintf(writer, "\n\n")
	}
}
