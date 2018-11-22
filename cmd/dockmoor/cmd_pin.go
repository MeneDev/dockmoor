package main

import (
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/hashicorp/go-multierror"
	"github.com/jessevdk/go-flags"
)

type pinOptions struct {
	MatchingOptions

	ReferenceFormat struct {
		ForceDomain bool `required:"no" long:"force-domain" description:"Includes domain even in well-known references"`
		NoName      bool `required:"no" long:"no-name" description:"Formats well-known references as digest only"`
		NoTag       bool `required:"no" long:"no-tag" description:"Don't include the tag in the reference'"`
		NoDigest    bool `required:"no" long:"no-digest" description:"Don't include the digest in the reference'"`
	} `group:"Reference format" description:"Control the format of references, defaults are sensible, changes are not recommended"`

	repoFactory func() dockref.Repository
}

func (po *pinOptions) ExecuteWithExitCode(args []string) (ExitCode, error) {
	var result *multierror.Error

	errVerify := verifyMatchOptions(&po.MatchingOptions)
	if errVerify != nil {
		result = multierror.Append(result, errVerify)
		po.Log().Errorf("Invalid options: %s\n", errVerify.Error())

		parser := flags.NewParser(&struct{}{}, flags.HelpFlag)
		command, err := addContainsCommand(po.mainOpts, AddCommand)
		result = multierror.Append(result, err)

		_, err = parser.ParseArgs([]string{command.Name, "--help"})
		result = multierror.Append(result, err)

		parser.WriteHelp(po.Stdout())
		return ExitInvalidParams, result.ErrorOrNil()
	}

	exitCode, err := po.matchAndProcess()
	result = multierror.Append(result, err)

	return exitCode, result.ErrorOrNil()
}

func (po *pinOptions) Repo() dockref.Repository {
	return po.repoFactory()
}

func pinOptionsNew(mainOptions *mainOptions, repositoryFactory func() dockref.Repository) *pinOptions {
	log := mainOptions.Log()

	po := pinOptions{
		MatchingOptions: MatchingOptions{
			mainOpts: mainOptions,
		},
		repoFactory: repositoryFactory,
	}

	po.matchHandler = func(r dockref.Reference) (dockref.Reference, error) {
		repo := po.Repo()
		resolvedArr, e := repo.Resolve(r)
		if e != nil {
			return nil, e
		}
		resolved, e := dockref.MostPreciseTag(resolvedArr, log)
		if e != nil {
			return resolved, e
		}

		format, err := po.RefFormat()
		if err != nil {
			return nil, err
		}

		formattedAndResolved, e := resolved.WithRequestedFormat(format)
		return formattedAndResolved, e
	}

	return &po
}

func addPinCommand(mainOptions *mainOptions, adder func(opts *mainOptions, command string, shortDescription string, longDescription string, data interface{}) (*flags.Command, error)) (*flags.Command, error) {
	repoFactory := mainOptions.repositoryFactory()
	pinOptions := pinOptionsNew(mainOptions, repoFactory)

	command, e := adder(mainOptions, "pin",
		"Change image references to a more reproducible format",
		"Change image references to a more reproducible format by adding version tags or digest",
		pinOptions)
	if e != nil {
		return nil, e

	}
	command.Hidden = true
	return command, e
}

func (po *pinOptions) RefFormat() (dockref.Format, error) {
	format := dockref.FormatHasName | dockref.FormatHasTag | dockref.FormatHasDigest

	rf := po.ReferenceFormat
	if rf.ForceDomain {
		format |= dockref.FormatHasDomain
	}
	if rf.NoName {
		format &= ^dockref.FormatHasName
	}
	if rf.NoTag {
		format = format & ^dockref.FormatHasTag
	}
	if rf.NoDigest {
		format &= ^dockref.FormatHasDigest
	}

	return format, nil
}
