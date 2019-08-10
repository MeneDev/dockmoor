package main

import (
	"bytes"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockproc"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/MeneDev/dockmoor/dockref/resolver"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
)

type pinOptions struct {
	MatchingOptions

	ReferenceFormat struct {
		ForceDomain bool `required:"no" long:"force-domain" description:"Includes domain even in well-known references"`
		NoName      bool `required:"no" long:"no-name" description:"Formats well-known references as digest only"`
		NoTag       bool `required:"no" long:"no-tag" description:"Don't include the tag in the reference"`
		NoDigest    bool `required:"no" long:"no-digest" description:"Don't include the digest in the reference"`
	} `group:"Reference format" description:"Control the format of references, defaults are sensible, changes are not recommended"`

	PinOptions struct {
		Resolver string `required:"no" short:"r" long:"resolver" description:"Strategy to resolve image references" choice:"dockerd" choice:"registry" default:"dockerd"`
		TagMode  string `required:"no" long:"tag-mode" description:"Strategy to resolve image references" choice:"unchanged" default:"unchanged"`
	} `group:"Pin Options" description:"Control how the image references are resolved"`

	Output struct {
		OutputFile flags.Filename `required:"no" short:"o" long:"output" description:"Output file to write to. If empty, input file will be used."`
	} `group:"Output parameters" description:"Output parameters"`

	resolverFactory func(name string) dockref.Resolver
	matches         bool
}

func (po *pinOptions) Execute(args []string) error {
	return errors.New("use ExecuteWithExitCode instead")
}

func (po *pinOptions) ExecuteWithExitCode(args []string) (exitCode ExitCode, err error) {
	// TODO code is redundant to other commands
	mopts := po.MatchingOptions

	exitCode, err = mopts.Verify()
	if err != nil {
		return
	}

	predicate, err := mopts.getPredicate()
	if err != nil {
		return ExitPredicateInvalid, err
	}

	buffer := bytes.NewBuffer(nil)

	err = mopts.WithInputDo(func(inputPath string, inputReader io.Reader) error {

		errFormat := mopts.WithFormatProcessorDo(inputReader, func(processor dockfmt.FormatProcessor) error {
			processor = processor.WithWriter(buffer)
			return po.applyFormatProcessor(predicate, processor)
		})

		if errFormat != nil {
			exitCode = ExitInvalidFormat
			return errFormat
		}
		return nil
	})

	if errExitCode, ok := exitCodeFromError(err); ok {
		return errExitCode, err
	}

	err = po.WithOutputDo(func(outputPath string) error {

		mode := os.FileMode(0660)

		info, e := os.Stat(outputPath)
		if e == nil {
			mode = info.Mode()
		}

		errWriteFile := ioutil.WriteFile(outputPath, buffer.Bytes(), mode)
		return errWriteFile
	})

	if po.matches {
		exitCode = ExitSuccess
	} else {
		exitCode = ExitNotFound
	}

	return exitCode, err
}

func (po *pinOptions) applyFormatProcessor(predicate dockproc.Predicate, processor dockfmt.FormatProcessor) error {

	return processor.Process(func(original dockref.Reference) (dockref.Reference, error) {
		if predicate.Matches(original) {
			po.matches = true
			repo := po.Resolver()

			mode, e := tagMode(po.PinOptions.TagMode)
			if e != nil {
				return nil, e
			}

			switch mode {
			case dockref.ResolveModeUnchanged:
				resolved, err := repo.Resolve(original)
				if err != nil {
					po.Log().WithField("error", err.Error()).Errorf("Could not resolve %s", original.Original())
					return nil, err
				}

				format, err := po.RefFormat()
				if err != nil {
					return nil, err
				}

				formatted, err := resolved.WithRequestedFormat(format)
				if err != nil {
					return nil, err
				}

				return formatted, nil

			case dockref.ResolveModeMostPreciseVersion:
				return nil, errors.Errorf("VersionMode %s not yet implemented", po.PinOptions.TagMode)
			}

		}
		return original, nil
	})
}

func tagMode(modeString string) (dockref.ResolveMode, error) {
	switch modeString {
	case "unchanged":
		return dockref.ResolveModeUnchanged, nil
	case "most-precise-version":
		return dockref.ResolveModeMostPreciseVersion, nil
	}

	return -1, errors.Errorf("Invalid VersionMode '%s'", modeString)
}

func (po *pinOptions) Resolver() dockref.Resolver {
	return po.resolverFactory(po.PinOptions.Resolver)
}

func pinOptionsNew(mainOptions *mainOptions) *pinOptions {
	po := pinOptions{
		MatchingOptions: MatchingOptions{
			mainOpts: mainOptions,
		},
		matches: false,
	}

	po.PinOptions.TagMode = "unchanged"
	po.resolverFactory = defaultResolverFactory

	return &po
}

func addPinCommand(
	mainOptions *mainOptions,
	adder func(opts *mainOptions, command string, shortDescription string, longDescription string, data interface{}) (*flags.Command, error)) (*flags.Command, error) {

	return addPinCommandWith(pinOptionsNew)(mainOptions, adder)
}

func addPinCommandWith(pinOptionsFactory func(mainOptions *mainOptions) *pinOptions) func(
	mainOptions *mainOptions,
	adder func(opts *mainOptions, command string, shortDescription string, longDescription string, data interface{}) (*flags.Command, error)) (*flags.Command, error) {

	return func(
		mainOptions *mainOptions,
		adder func(opts *mainOptions, command string, shortDescription string, longDescription string, data interface{}) (*flags.Command, error)) (*flags.Command, error) {
		pinOptions := pinOptionsFactory(mainOptions)

		command, e := adder(mainOptions, "pin",
			"Change image references to a more reproducible format",
			"Change image references to a more reproducible format",
			pinOptions)
		return command, e
	}
}

func defaultResolverFactory(resolverName string) dockref.Resolver {
	switch resolverName {
	case "dockerd":
		return resolver.DockerDaemonResolverNew()
	case "registry":
		return resolver.DockerRegistryResolverNew()
	}

	return nil
}

func (po *pinOptions) RefFormat() (dockref.Format, error) {
	format := dockref.FormatHasName | dockref.FormatHasTag | dockref.FormatHasDigest

	rf := po.ReferenceFormat

	if rf.NoDigest && rf.NoName {
		return 0, errors.New("invalid Reference Format: --no-name and --no-digest are mutually exclusive")
	}

	if rf.ForceDomain && rf.NoName {
		return 0, errors.New("invalid Reference Format: --force-domain and --no-name are mutually exclusive")
	}

	if rf.ForceDomain {
		format |= dockref.FormatHasDomain
	}
	if rf.NoName {
		format &= ^dockref.FormatHasName
	}
	if rf.NoTag {
		format &= ^dockref.FormatHasTag
	}
	if rf.NoDigest {
		format &= ^dockref.FormatHasDigest
	}

	return format, nil
}

func (po *pinOptions) WithOutputDo(action func(outputPath string) error) error {
	filename := string(po.Output.OutputFile)
	if filename != "" {
		return action(filename)
	}

	return po.MatchingOptions.WithOutputDo(action)
}
