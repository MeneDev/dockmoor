package main

import "github.com/jessevdk/go-flags"

type UpdateOptions struct {
	Version     string     `required:"no" long:"version" description:"The version to update to" choice:"latest" choice:"tag" choice:"major" choice:"minor" choice:"patch" default:"tag"`
	Positional  OutputFile `positional-args:"yes" required:"yes"`
	mainOptions *MainOptions
}

func init() {
	addUpdateCommand(&globalMainOptions)
}


func addUpdateCommand(mainOptions *MainOptions) (*flags.Command, error) {

	parser := mainOptions.Parser()
	var updateOptions UpdateOptions
	updateOptions.mainOptions = mainOptions

	return 	parser.AddCommand("update",
		"Replace image references with a latest reference from repository",
		"Replace image references with a latest reference from repository",
		&updateOptions)
}
