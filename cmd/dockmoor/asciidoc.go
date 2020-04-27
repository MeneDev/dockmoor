package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/jessevdk/go-flags"
)

func WriteASCIIDoc(parser *flags.Parser, writer io.Writer) {
	mdPrintf(writer, "== Usage\n")
	commands := []*flags.Command{parser.Command}
	commands = visibleCommands(commands)
	WriteASCIIDocUsage(commands, writer)

	WriteASCIIDocGroups(writer, parser.Command.Groups(), 2)

	mdPrintf(writer, "=== Commands\n\n")

	parserCommands := parser.Commands()
	parserCommands = visibleCommands(parserCommands)
	for _, cmd := range parserCommands {
		mdPrintf(writer, " * <<%[2]s,%[1]s>>\n", cmd.Name, strings.ToLower(cmd.Name)+"-command")
	}
	mdPrintf(writer, "\n")

	for _, cmd := range parserCommands {
		mdPrintf(writer, "==== %s command\n", cmd.Name)
		WriteASCIIDocUsage(append(commands, cmd), writer)
		mdPrintf(writer, "%s\n\n", cmd.LongDescription)
		WriteASCIIDocOptions(writer, cmd.Options(), 5)
		WriteASCIIDocGroups(writer, cmd.Groups(), 5)
	}
}

func WriteASCIIDocUsage(commands []*flags.Command, writer io.Writer) {
	commands = visibleCommands(commands)

	mdPrintf(writer, "> ")
	for idxCommand, command := range commands {
		isFirstCommand := idxCommand == 0
		isLastCommand := idxCommand+1 == len(commands)

		mdPrintf(writer, "%s", command.Name)
		if len(command.Options()) > 0 || len(command.Groups()) > 0 {
			if isFirstCommand {
				mdPrintf(writer, " [OPTIONS]")
			} else {
				mdPrintf(writer, " [%s-OPTIONS]", command.Name)
			}
		}

		if len(command.Args()) > 0 {
			for _, v := range command.Args() {
				var format string
				if v.Required == 0 {
					format = " [%s]"
				} else {
					format = " %s"
				}

				mdPrintf(writer, format, v.Name)
			}
		}

		if !isLastCommand {
			mdPrintf(writer, " ")
		} else {
			commandCommands := command.Commands()
			commandCommands = visibleCommands(commandCommands)
			if len(commandCommands) > 0 {
				var cmds []string
				for _, cmd := range commandCommands {
					cmds = append(cmds, fmt.Sprintf("<<%[2]s,%[1]s>>", cmd.Name, strings.ToLower(cmd.Name)+"-command"))
				}

				var format string
				if command.SubcommandsOptional {
					format = " [%s]"
				} else {
					format = " &lt;%s&gt;"
				}
				fmt.Fprintf(writer, format, strings.Join(cmds, " | "))
				mdPrintf(writer, " [command-OPTIONS]")
			}
		}
	}

	mdPrintf(writer, "\n\n")
}

func WriteASCIIDocGroups(writer io.Writer, groups []*flags.Group, level int) {
	for _, group := range groups {
		if group.Hidden {
			continue
		}
		mdPrintf(writer, strings.Repeat("#", level)+" %s\n", group.ShortDescription)
		if group.LongDescription != "" {
			mdPrintf(writer, "%s\n\n", group.LongDescription)
		}
		WriteASCIIDocOptions(writer, group.Options(), level+1)
	}
}

func WriteASCIIDocOptions(writer io.Writer, options []*flags.Option, level int) {
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
