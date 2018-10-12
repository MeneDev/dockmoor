package main

import (
	"fmt"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockproc"
	"github.com/hashicorp/go-multierror"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
)

type MatchingMode int

const (
	matchOnly     MatchingMode = iota
	matchAndPrint MatchingMode = iota
)

const (
	domainPred       = "domain"
	namePred         = "name"
	pathPred         = "path"
	familiarNamePred = "familiar-name"

	latestPred   = "latest"
	outdatedPred = "outdated"
	untaggedPred = "untagged"
	tagPred      = "tag"

	digestsPred  = "digest"
	unpinnedPred = "unpinned"
)

var namePredicateNames = []string{domainPred, namePred, pathPred, familiarNamePred}
var tagPredicateNames = []string{latestPred, outdatedPred, untaggedPred, tagPred}
var digestPredicateNames = []string{digestsPred, unpinnedPred}

var predicateNames = append(
	append(
		namePredicateNames,
		tagPredicateNames...),
	digestPredicateNames...)

func indexOf(item string, slice []string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

func deliberatelyUnhandled(err error) {
	// noop
}

func withoutNoError(item string, slice []string) []string {
	strings, err := without(item, slice)
	deliberatelyUnhandled(err)
	return strings
}

func without(item string, slice []string) ([]string, error) {
	idx := indexOf(item, slice)
	if idx < 0 {
		return slice, errors.Errorf("Cannot find item '%s' in slice", item)
	}
	wo := make([]string, 0)
	wo = append(wo, slice[:idx]...)
	wo = append(wo, slice[idx+1:]...)

	return wo, nil
}

var exclusives = map[string][]string{
	domainPred: {
		namePred, familiarNamePred,
	},
	namePred: {
		domainPred, pathPred, familiarNamePred,
	},
	familiarNamePred: {
		domainPred, pathPred, namePred,
	},
	pathPred: {
		namePred, familiarNamePred,
	},
	latestPred:   withoutNoError(latestPred, tagPredicateNames),
	outdatedPred: withoutNoError(outdatedPred, tagPredicateNames),
	untaggedPred: withoutNoError(untaggedPred, tagPredicateNames),
	tagPred:      withoutNoError(tagPred, tagPredicateNames),

	digestsPred:  withoutNoError(digestsPred, digestPredicateNames),
	unpinnedPred: withoutNoError(unpinnedPred, digestPredicateNames),
}

func errorFor(a, b string) error {
	return errors.Errorf("Cannot combine --%s and --%s", a, b)
}

type MatchingOptions struct {
	DomainPredicates struct {
		Domains []string `required:"no" long:"domain" description:"Matches all images matching one of the specified domains"`
	} `group:"Domain Predicates" description:"Limit matched image references depending on their domain"`

	NamePredicates struct {
		Names         []string `required:"no" long:"name" description:"Matches all images matching one of the specified names (e.g. \"docker.io/library/nginx\")"`
		FamiliarNames []string `required:"no" long:"familiar-name" short:"f" description:"Matches all images matching one of the specified familiar names (e.g. \"nginx\")"`
		Paths         []string `required:"no" long:"path" description:"Matches all images matching one of the specified paths (e.g. \"library/nginx\")"`
	} `group:"Name Predicates" description:"Limit matched image references depending on their name"`

	TagPredicates struct {
		Untagged bool     `required:"no" long:"untagged" description:"Matches images with no tag"`
		Latest   bool     `required:"no" long:"latest" description:"Matches images with latest or no tag. References with digest are only matched when explicit latest tag is present."`
		Outdated bool     `required:"no" long:"outdated" description:"Matches all images with newer versions available" hidden:"true"`
		Tags     []string `required:"no" long:"tag" description:"Matches all images matching one of the specified tag"`
	} `group:"Tag Predicates" description:"Limit matched image references depending on their tag"`

	DigestPredicates struct {
		Unpinned bool     `required:"no" long:"unpinned" description:"Matches unpinned image references, i.e. image references without digest."`
		Digests  []string `required:"no" long:"digest" description:"Matches all image references with one of the provided digests"`
	} `group:"Digest Predicates" description:"Limit matched image references depending on their digest"`

	Positional struct {
		InputFile flags.Filename `required:"yes"`
	} `positional-args:"yes"`

	mainOpts *mainOptions
	mode     MatchingMode
}

func (mopts *MatchingOptions) mainOptions() *mainOptions {
	return mopts.mainOpts
}

func (mopts *MatchingOptions) Log() *logrus.Logger {
	return mopts.mainOpts.Log()
}

func (mopts *MatchingOptions) Stdout() io.Writer {
	return mopts.mainOptions().stdout
}

func (mopts *MatchingOptions) isSetPredicateByName(name string) bool {
	switch name {
	case domainPred:
		return mopts.DomainPredicates.Domains != nil
	case namePred:
		return mopts.NamePredicates.Names != nil
	case pathPred:
		return mopts.NamePredicates.Paths != nil
	case familiarNamePred:
		return mopts.NamePredicates.FamiliarNames != nil
	case latestPred:
		return mopts.TagPredicates.Latest
	case outdatedPred:
		return mopts.TagPredicates.Outdated
	case untaggedPred:
		return mopts.TagPredicates.Untagged
	case tagPred:
		return mopts.TagPredicates.Tags != nil
	case digestsPred:
		return mopts.DigestPredicates.Digests != nil
	case unpinnedPred:
		return mopts.DigestPredicates.Unpinned
	}

	panic(fmt.Sprintf("Unknown predicate name %s", name))
}

func verifyMatchOptions(mo *MatchingOptions) error {

	var err error

	for i1, p1 := range predicateNames {
		if !mo.isSetPredicateByName(p1) {
			continue
		}

		for i2, p2 := range predicateNames {
			if i1 >= i2 {
				continue
			}

			if !mo.isSetPredicateByName(p2) {
				continue
			}

			strings := exclusives[p1]
			idx := indexOf(p2, strings)

			if idx >= 0 {
				err = multierror.Append(err, errorFor(p1, p2))
			}
		}
	}

	return err
}

func (mopts *MatchingOptions) Execute(args []string) error {
	return errors.New("Use ExecuteWithExitCode instead")
}

func (mopts *MatchingOptions) ExecuteWithExitCode(args []string) (ExitCode, error) {
	var result *multierror.Error

	errVerify := verifyMatchOptions(mopts)
	if errVerify != nil {
		result = multierror.Append(result, errVerify)
		mopts.Log().Errorf("Invalid options: %s\n", errVerify.Error())

		parser := flags.NewParser(&struct{}{}, flags.HelpFlag)
		command, err := addContainsCommand(mopts.mainOpts, AddCommand)
		result = multierror.Append(result, err)

		_, err = parser.ParseArgs([]string{command.Name, "--help"})
		result = multierror.Append(result, err)

		parser.WriteHelp(mopts.mainOpts.stdout)
		return ExitInvalidParams, result.ErrorOrNil()
	}

	exitCode, err := mopts.match()
	result = multierror.Append(result, err)

	return exitCode, result.ErrorOrNil()
}

var latestPredicateFactory = func() dockproc.Predicate {
	return dockproc.LatestPredicateNew()
}

var latestUnpinnedFactory = func() dockproc.Predicate {
	return dockproc.UnpinnedPredicateNew()
}

var anyPredicateFactory = func() dockproc.Predicate {
	return dockproc.AnyPredicateNew()
}

var domainsPredicateFactory = func(domains []string) dockproc.Predicate {
	return dockproc.DomainsPredicateNew(domains)
}
var namePredicateFactory = func(names []string) dockproc.Predicate {
	return dockproc.NamesPredicateNew(names)
}
var familiarNamePredicateFactory = func(familiarNames []string) dockproc.Predicate {
	return dockproc.FamiliarNamesPredicateNew(familiarNames)
}
var pathsPredicateFactory = func(paths []string) dockproc.Predicate {
	return dockproc.PathsPredicateNew(paths)
}

//var outdatedPredicateFactory = func() dockproc.Predicate {
//	return dockproc.OutdatedPredicateNew()
//}
var untaggedPredicateFactory = func() dockproc.Predicate {
	return dockproc.UntaggedPredicateNew()
}
var tagsPredicateFactory = func(tags []string) dockproc.Predicate {
	return dockproc.TagsPredicateNew(tags)
}
var digestsPredicateFactory = func(digests []string) dockproc.Predicate {
	return dockproc.DigestsPredicateNew(digests)
}
var andPredicateFactory = func(predicates []dockproc.Predicate) dockproc.Predicate {
	return dockproc.AndPredicateNew(predicates)
}

func (mopts *MatchingOptions) getPredicate() dockproc.Predicate {
	anyPredicate := anyPredicateFactory()
	var predicates []dockproc.Predicate

	if mopts.DomainPredicates.Domains != nil {
		p := domainsPredicateFactory(mopts.DomainPredicates.Domains)
		predicates = append(predicates, p)
	}

	if mopts.NamePredicates.Names != nil {
		p := namePredicateFactory(mopts.NamePredicates.Names)
		predicates = append(predicates, p)
	}

	if mopts.NamePredicates.FamiliarNames != nil {
		p := familiarNamePredicateFactory(mopts.NamePredicates.FamiliarNames)
		predicates = append(predicates, p)
	}

	if mopts.NamePredicates.Paths != nil {
		p := pathsPredicateFactory(mopts.NamePredicates.Paths)
		predicates = append(predicates, p)
	}

	//if mopts.TagPredicates.Outdated {
	//	p := outdatedPredicateFactory()
	//	predicates = append(predicates, p)
	//}

	if mopts.TagPredicates.Untagged {
		p := untaggedPredicateFactory()
		predicates = append(predicates, p)
	}

	if mopts.TagPredicates.Tags != nil {
		p := tagsPredicateFactory(mopts.TagPredicates.Tags)
		predicates = append(predicates, p)
	}

	if mopts.TagPredicates.Latest {
		p := latestPredicateFactory()
		predicates = append(predicates, p)
	}

	if mopts.DigestPredicates.Unpinned {
		p := latestUnpinnedFactory()
		predicates = append(predicates, p)
	}

	if mopts.DigestPredicates.Digests != nil {
		p := digestsPredicateFactory(mopts.DigestPredicates.Digests)
		predicates = append(predicates, p)
	}

	switch len(predicates) {
	case 0:
		return anyPredicate
	case 1:
		return predicates[0]
	default:
		return andPredicateFactory(predicates)
	}
}

func (mopts *MatchingOptions) open(readable string) (io.ReadCloser, error) {
	return mopts.mainOpts.readableOpener(readable)
}

func saveClose(log *logrus.Logger, readCloser io.Closer) {
	if readCloser != nil {
		err := readCloser.Close()
		if err != nil {
			log.Errorf("Error closing: %s", err.Error())
		}
	}
}

func (mopts *MatchingOptions) match() (exitCode ExitCode, err error) {
	log := mopts.Log()

	filePathInput := string(mopts.Positional.InputFile)

	fpInput, err := mopts.open(filePathInput)
	defer saveClose(log, fpInput)

	if err != nil {
		log.Errorf("Could not open file: %s", err.Error())
		exitCode = ExitCouldNotOpenFile
		return
	}

	formatProvider := mopts.mainOptions().FormatProvider()
	fileFormat, formatError := dockfmt.IdentifyFormat(log, formatProvider, fpInput, filePathInput)
	if fileFormat == nil {
		return ExitInvalidFormat, formatError
	}

	formatProcessor := dockfmt.FormatProcessorNew(fileFormat, log, fpInput)
	exitCode, err = mopts.matchFormatProcessor(formatProcessor)
	return
}

func (mopts *MatchingOptions) matchFormatProcessor(formatProcessor dockfmt.FormatProcessor) (exitCode ExitCode, err error) {
	log := mopts.Log()

	predicate := mopts.getPredicate()
	accumulator, err := dockproc.MatchesAccumulatorNew(predicate, log, mopts.Stdout())

	if err != nil {
		return ExitUnknownError, err
	}

	errAcc := accumulator.Accumulate(formatProcessor)
	if errAcc != nil {
		log.Errorf("Error during accumulation: %s", errAcc.Error())
	}

	matches := accumulator.Matches()

	if len(matches) > 0 {
		exitCode = ExitSuccess
	} else {
		exitCode = ExitNotFound
	}

	var results *multierror.Error

	if mopts.mode == matchAndPrint {
		for _, r := range matches {
			_, err = fmt.Fprintf(mopts.Stdout(), "%s\n", r.Original())
			results = multierror.Append(results, err)
		}
	}
	return exitCode, results.ErrorOrNil()
}
