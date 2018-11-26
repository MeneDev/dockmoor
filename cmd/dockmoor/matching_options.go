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
		Domains []string `required:"no" long:"domain" description:"Matches all images matching one of the specified domains. Surround with '/' for regex i.e. /regex/."`
	} `group:"Domain Predicates" description:"Limit matched image references depending on their domain"`

	NamePredicates struct {
		Names         []string `required:"no" long:"name" description:"Matches all images matching one of the specified names (e.g. \"docker.io/library/nginx\"). Surround with '/' for regex i.e. /regex/."`
		FamiliarNames []string `required:"no" long:"familiar-name" short:"f" description:"Matches all images matching one of the specified familiar names (e.g. \"nginx\"). Surround with '/' for regex i.e. /regex/."`
		Paths         []string `required:"no" long:"path" description:"Matches all images matching one of the specified paths (e.g. \"library/nginx\"). Surround with '/' for regex i.e. /regex/."`
	} `group:"Name Predicates" description:"Limit matched image references depending on their name"`

	TagPredicates struct {
		Untagged bool     `required:"no" long:"untagged" description:"Matches images with no tag"`
		Latest   bool     `required:"no" long:"latest" description:"Matches images with latest or no tag. References with digest are only matched when explicit latest tag is present."`
		Outdated bool     `required:"no" long:"outdated" description:"Matches all images with newer versions available" hidden:"true"`
		Tags     []string `required:"no" long:"tag" description:"Matches all images matching one of the specified tag. Surround with '/' for regex i.e. /regex/."`
	} `group:"Tag Predicates" description:"Limit matched image references depending on their tag"`

	DigestPredicates struct {
		Unpinned bool     `required:"no" long:"unpinned" description:"Matches unpinned image references, i.e. image references without digest."`
		Digests  []string `required:"no" long:"digest" description:"Matches all image references with one of the provided digests."`
	} `group:"Digest Predicates" description:"Limit matched image references depending on their digest"`

	Positional struct {
		InputFile flags.Filename `required:"yes"`
	} `positional-args:"yes"`

	mainOpts     *mainOptions
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

func (mopts *MatchingOptions) Verify() (ExitCode, error) {
	var result *multierror.Error

	errVerify := verifyMatchOptions(mopts)
	if errVerify != nil {
		result = multierror.Append(result, errVerify)
		mopts.Log().Errorf("Invalid options: %s\n", errVerify.Error())
		parser := flags.NewParser(mopts, flags.HelpFlag)
		parser.WriteHelp(mopts.Stdout())
		return ExitInvalidParams, result.ErrorOrNil()
	}

	return ExitSuccess, result.ErrorOrNil()
}

var latestPredicateFactory = func() (dockproc.Predicate, error) {
	return dockproc.LatestPredicateNew()
}

var latestUnpinnedFactory = func() (dockproc.Predicate, error) {
	return dockproc.UnpinnedPredicateNew()
}

var anyPredicateFactory = func() (dockproc.Predicate, error) {
	return dockproc.AnyPredicateNew()
}

var domainsPredicateFactory = func(domains []string) (dockproc.Predicate, error) {
	return dockproc.DomainsPredicateNew(domains)
}
var namePredicateFactory = func(names []string) (dockproc.Predicate, error) {
	return dockproc.NamesPredicateNew(names)
}
var familiarNamePredicateFactory = func(familiarNames []string) (dockproc.Predicate, error) {
	return dockproc.FamiliarNamesPredicateNew(familiarNames)
}
var pathsPredicateFactory = func(paths []string) (dockproc.Predicate, error) {
	return dockproc.PathsPredicateNew(paths)
}

//var outdatedPredicateFactory = func() (dockproc.Predicate, error) {
//	return dockproc.OutdatedPredicateNew()
//}
var untaggedPredicateFactory = func() (dockproc.Predicate, error) {
	return dockproc.UntaggedPredicateNew()
}
var tagsPredicateFactory = func(tags []string) (dockproc.Predicate, error) {
	return dockproc.TagsPredicateNew(tags)
}
var digestsPredicateFactory = func(digests []string) (dockproc.Predicate, error) {
	return dockproc.DigestsPredicateNew(digests)
}
var andPredicateFactory = func(predicates []dockproc.Predicate) (dockproc.Predicate, error) {
	return dockproc.AndPredicateNew(predicates)
}

func (mopts *MatchingOptions) getPredicate() (dockproc.Predicate, error) {

	anyPredicate, e := anyPredicateFactory()
	if e != nil {
		return nil, e
	}

	var err *multierror.Error
	var predicates []dockproc.Predicate

	if mopts.DomainPredicates.Domains != nil {
		p, e := domainsPredicateFactory(mopts.DomainPredicates.Domains)
		err = multierror.Append(err, e)
		predicates = append(predicates, p)
	}

	if mopts.NamePredicates.Names != nil {
		p, e := namePredicateFactory(mopts.NamePredicates.Names)
		err = multierror.Append(err, e)
		predicates = append(predicates, p)
	}

	if mopts.NamePredicates.FamiliarNames != nil {
		p, e := familiarNamePredicateFactory(mopts.NamePredicates.FamiliarNames)
		err = multierror.Append(err, e)
		predicates = append(predicates, p)
	}

	if mopts.NamePredicates.Paths != nil {
		p, e := pathsPredicateFactory(mopts.NamePredicates.Paths)
		err = multierror.Append(err, e)
		predicates = append(predicates, p)
	}

	//if mopts.TagPredicates.Outdated {
	//	p := outdatedPredicateFactory()
	//	predicates = append(predicates, p)
	//}

	if mopts.TagPredicates.Untagged {
		p, e := untaggedPredicateFactory()
		err = multierror.Append(err, e)
		predicates = append(predicates, p)
	}

	if mopts.TagPredicates.Tags != nil {
		p, e := tagsPredicateFactory(mopts.TagPredicates.Tags)
		err = multierror.Append(err, e)
		predicates = append(predicates, p)
	}

	if mopts.TagPredicates.Latest {
		p, e := latestPredicateFactory()
		err = multierror.Append(err, e)
		predicates = append(predicates, p)
	}

	if mopts.DigestPredicates.Unpinned {
		p, e := latestUnpinnedFactory()
		err = multierror.Append(err, e)
		predicates = append(predicates, p)
	}

	if mopts.DigestPredicates.Digests != nil {
		p, e := digestsPredicateFactory(mopts.DigestPredicates.Digests)
		err = multierror.Append(err, e)
		predicates = append(predicates, p)
	}

	switch len(predicates) {
	case 0:
		return anyPredicate, err.ErrorOrNil()
	case 1:
		return predicates[0], err.ErrorOrNil()
	default:
		predicate, e := andPredicateFactory(predicates)
		err = multierror.Append(err, e)
		return predicate, err.ErrorOrNil()
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

func (mopts *MatchingOptions) WithInputDo(action func(filePathInput string, fpInput io.Reader) error) error {
	log := mopts.Log()

	filePathInput := string(mopts.Positional.InputFile)
	fpInput, err := mopts.open(filePathInput)
	defer saveClose(log, fpInput)

	if err != nil {
		return err
	}

	err = action(filePathInput, fpInput)

	return err
}

func (mopts *MatchingOptions) WithFormatProcessorDo(fpInput io.Reader, action func(processor dockfmt.FormatProcessor) error) error {
	log := mopts.Log()

	formatProvider := mopts.mainOptions().FormatProvider()
	filename := string(mopts.Positional.InputFile)
	fileFormat, formatError := dockfmt.IdentifyFormat(log, formatProvider, fpInput, filename)

	if formatError != nil {
		return formatError
	}

	formatProcessor := dockfmt.FormatProcessorNew(fileFormat, log, fpInput)

	return action(formatProcessor)
}

func (mopts *MatchingOptions) WithOutputDo(action func(outputPath string) error) error {
	filename := string(mopts.Positional.InputFile)
	return action(filename)
}
