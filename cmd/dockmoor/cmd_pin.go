package main

import (
	"bytes"
	"errors"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockproc"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/jessevdk/go-flags"
	"io"
	"io/ioutil"
	"os"
)

type pinOptions struct {
	MatchingOptions

	ReferenceFormat struct {
		ForceDomain bool `required:"no" long:"force-domain" description:"Includes domain even in well-known references"`
		NoName      bool `required:"no" long:"no-name" description:"Formats well-known references as digest only"`
		NoTag       bool `required:"no" long:"no-tag" description:"Don't include the tag in the reference'"`
		NoDigest    bool `required:"no" long:"no-digest" description:"Don't include the digest in the reference'"`
	} `group:"Reference format" description:"Control the format of references, defaults are sensible, changes are not recommended"`

	Output struct {
		OutputFile flags.Filename `required:"no" short:"o" long:"output" description:"Output file to write to. If empty, input file will be used."`
	} `group:"Output parameters" description:"Output parameters"`

	repoFactory func() dockref.Repository
	matches     bool
}

func (po *pinOptions) Execute(args []string) error {
	return errors.New("Use ExecuteWithExitCode instead")
}

func (po *pinOptions) ExecuteWithExitCode(args []string) (exitCode ExitCode, err error) {
	// TODO code is redundant to other commands
	mopts := po.MatchingOptions

	exitCode, err = mopts.Verify()
	if err != nil {
		return
	}

	predicate, err := mopts.getPredicate()
	buffer := bytes.NewBuffer(nil)

	err = mopts.WithInputDo(func(inputPath string, inputReader io.Reader) error {

		err := mopts.WithFormatProcessorDo(inputReader, func(processor dockfmt.FormatProcessor) error {
			processor = processor.WithWriter(buffer)
			return po.applyFormatProcessor(predicate, processor)
		})

		if err != nil {
			exitCode = ExitInvalidFormat
			return err
		}
		return nil
	})

	if err != nil {
		switch {
		case contains(err, func(err error) bool { _, ok := err.(dockfmt.UnknownFormatError); return ok }) ||
			contains(err, func(err error) bool { _, ok := err.(dockfmt.AmbiguousFormatError); return ok }) ||
			contains(err, func(err error) bool { _, ok := err.(dockfmt.FormatError); return ok }):
			exitCode = ExitInvalidFormat
		case contains(err, func(err error) bool { _, ok := err.(error); return ok }):
			exitCode = ExitCouldNotOpenFile
		}
		return
	} else {
		err = mopts.WithOutputDo(func (outputPath string) error {

			mode := os.FileMode(0660)

			info, e := os.Stat(outputPath)
			if e == nil {
				mode = info.Mode()
			}

			err := ioutil.WriteFile(outputPath, buffer.Bytes(), mode)
			return err
		})
	}

	if po.matches {
		exitCode = ExitSuccess
	} else {
		exitCode = ExitNotFound
	}

	return exitCode, err
}

func (po *pinOptions) applyFormatProcessor(predicate dockproc.Predicate, processor dockfmt.FormatProcessor) error {

	processor.Process(func(original dockref.Reference) (dockref.Reference, error) {
		if predicate.Matches(original) {
			repo := po.Repo()
			rs, err := repo.Resolve(original)
			if err != nil {
				return nil, err
			}
			mostPrecise, err := dockref.MostPreciseTag(rs, po.Log())
			if err != nil {
				return nil, err
			}

			po.matches = true
			if err != nil {
				return mostPrecise, err
			}
			return mostPrecise, err
		}
		return original, nil
	})

	return nil
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
		matches: false,
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
