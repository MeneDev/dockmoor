package main

import (
	"io"
	"github.com/jessevdk/go-flags"
	"strings"
	"fmt"
)

func WriteAsciiDoc(parser *flags.Parser, writer io.Writer) {
	mdPrintf(writer, "== Usage\n")
	commands := []*flags.Command{parser.Command}
	WriteAsciiDocUsage(commands, writer)

	WriteAsciiDocGroups(writer, parser.Command.Groups(), 2)

	mdPrintf(writer, "=== Commands\n\n")

	for _, cmd := range parser.Commands() {
		mdPrintf(writer, " * <<%[2]s,%[1]s>>\n", cmd.Name, strings.ToLower(cmd.Name)+"-command")
	}
	mdPrintf(writer, "\n")

	for _, cmd := range parser.Commands() {
		mdPrintf(writer, "==== %s command\n", cmd.Name)
		WriteAsciiDocUsage(append(commands, cmd), writer)
		mdPrintf(writer, "%s\n\n", cmd.LongDescription)
		WriteAsciiDocOptions(writer, cmd.Options(), 3)
		WriteAsciiDocGroups(writer, cmd.Groups(), 3)
	}
}

func WriteAsciiDocUsage(commands []*flags.Command, writer io.Writer) {

	mdPrintf(writer, "> ", )
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
			if len(command.Commands()) > 0 {
				var cmds []string
				for _, cmd := range command.Commands() {
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

func WriteAsciiDocGroups(writer io.Writer, groups []*flags.Group, level int) {

	for _, group := range groups {
		if group.Hidden {
			continue
		}
		mdPrintf(writer, strings.Repeat("#", level)+" %s\n", group.ShortDescription)
		if group.LongDescription != "" {
			mdPrintf(writer, "%s\n\n", group.LongDescription)
		}
		WriteAsciiDocOptions(writer, group.Options(), level+1)
	}
}

func WriteAsciiDocOptions(writer io.Writer, options []*flags.Option, level int) {
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
			names = append(names, "--"+string(opt.LongNameWithNamespace()))
		}

		mdPrintf(writer, strings.Join(names, "**, **"))

		mdPrintf(writer, "**  \n%s", opt.Description)
		if opt.Choices != nil {
			mdPrintf(writer, " (one of `%s`)", strings.Join(opt.Choices, "`, `"))
		}
		mdPrintf(writer, "\n\n")
	}
}
