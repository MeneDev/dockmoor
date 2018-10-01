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
	"strings"
)

type MatchingMode int

const (
	matchOnly     MatchingMode = iota
	matchAndPrint MatchingMode = iota
)


var domainPredicateNames = []string{"domains"}
var namePredicateNames = []string{"names"}
var tagPredicateNames = []string{"latest", "outdated", "untagged", "tags"}
var digestPredicateNames = []string{"digests", "unpinned"}

var predicateGroups = map[string][]string{
	"domain": domainPredicateNames,
	"name": namePredicateNames,
	"tag": tagPredicateNames,
	"digest": digestPredicateNames,
}

var predicateNames = append(
	append(
		append(
			domainPredicateNames,
			namePredicateNames...),
		tagPredicateNames...),
	digestPredicateNames...)

var (
	ErrAtMostOneDomainPredicate = errors.Errorf("Provide at most one of --" + strings.Join(domainPredicateNames, ", --"))
	ErrAtMostOneNamePredicate = errors.Errorf("Provide at most one of --" + strings.Join(namePredicateNames, ", --"))
	ErrAtMostOneTagPredicate = errors.Errorf("Provide at most one of --" + strings.Join(tagPredicateNames, ", --"))
	ErrAtMostOneDigestPredicate = errors.Errorf("Provide at most one of --" + strings.Join(digestPredicateNames, ", --"))
)

var ErrAtMostOnePredicate = map[string]error {
	"domain": ErrAtMostOneDomainPredicate,
	"name": ErrAtMostOneNamePredicate,
	"tag": ErrAtMostOneTagPredicate,
	"digest": ErrAtMostOneDigestPredicate,
}

type MatchingOptions struct {
	DomainPredicates struct {
		Domains []string `required:"no" long:"domain" description:"Matches all images matching one of the specified domains" hidden:"true"`
	} `group:"Domain Predicates" description:"Limit matched image references depending on their domain"`

	NamePredicates struct {
		Names []string `required:"no" long:"name" description:"Matches all images matching one of the specified names" hidden:"true"`
	} `group:"Name Predicates" description:"Limit matched image references depending on their name"`

	TagPredicates struct {
		Untagged bool     `required:"no" long:"untagged" description:"Matches images with no tag"`
		Latest   bool     `required:"no" long:"latest" description:"Matches images with latest or no tag"`
		Outdated bool     `required:"no" long:"outdated" description:"Matches all images with newer versions available" hidden:"true"`
		Tags     []string `required:"no" long:"tag" description:"Matches all images matching one of the specified tags" hidden:"true"`
	} `group:"Tag Predicates" description:"Limit matched image references depending on their tag"`

	DigestPredicates struct {
		Unpinned bool     `required:"no" long:"unpinned" description:"Matches unpinned images"`
		Digests  []string `required:"no" long:"digest" description:"Matches all digests matching one of the specified digests" hidden:"true"`
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

type GroupCount struct {
	countDomain, countName, countTag, countDigest int
}

func calculateCounts(fo *MatchingOptions) GroupCount {
	setDomain := calculateDomainCounts(fo)
	setName := calculateNameCounts(fo)
	setTag := calculateTagCounts(fo)
	setDigest := calculateDigestCounts(fo)
	count := GroupCount{countDomain: setDomain, countName: setName, countTag: setTag, countDigest: setDigest}
	return count
}

func calculateDomainCounts(options *MatchingOptions) (count int) {
	if options.DomainPredicates.Domains != nil {
		count++
	}
	return
}

func calculateNameCounts(options *MatchingOptions) (count int) {
	if options.NamePredicates.Names != nil {
		count++
	}
	return
}

func calculateTagCounts(options *MatchingOptions) (count int) {
	if options.TagPredicates.Tags != nil {
		count++
	}
	if options.TagPredicates.Untagged {
		count++
	}
	if options.TagPredicates.Outdated {
		count++
	}
	if options.TagPredicates.Latest {
		count++
	}
	return
}

func calculateDigestCounts(options *MatchingOptions) (count int) {
	if options.DigestPredicates.Digests != nil {
		count++
	}
	if options.DigestPredicates.Unpinned {
		count++
	}
	return
}

func verifyMatchOptionsAtMostOnePredicatePerGroup(fo *MatchingOptions) error {

	counts := calculateCounts(fo)

	// domain and name cannot have more than one predicate set
	// as there is only one

	//if counts.countDomain > 1 {
	//	return ErrAtMostOneDomainPredicate
	//}
	//if counts.countName > 1 {
	//	return ErrAtMostOneNamePredicate
	//}

	if counts.countTag > 1 {
		return ErrAtMostOneTagPredicate
	}
	if counts.countDigest > 1 {
		return ErrAtMostOneDigestPredicate
	}

	return nil
}

func verifyMatchOptions(fo *MatchingOptions) error {
	err := verifyMatchOptionsAtMostOnePredicatePerGroup(fo)
	return err
}

func (mopts *MatchingOptions) Execute(args []string) error {
	return errors.New("Use ExecuteWithExitCode instead")
}

func (mopts *MatchingOptions) ExecuteWithExitCode(args []string) (ExitCode, error) {
	var result *multierror.Error

	errVerify := verifyMatchOptions(mopts)
	if errVerify != nil {
		mopts.Log().Errorf("Invalid options: %s\n", errVerify.Error())

		parser := flags.NewParser(&struct{}{}, flags.HelpFlag)
		command, err := addContainsCommand(mopts.mainOpts, AddCommand)
		result = multierror.Append(err, err)

		_, err = parser.ParseArgs([]string{command.Name, "--help"})
		result = multierror.Append(err, err)

		parser.WriteHelp(mopts.mainOpts.stdout)
		return ExitInvalidParams, errVerify
	}

	exitCode, err := mopts.match()
	result = multierror.Append(err, err)

	return exitCode, result.ErrorOrNil()
}

var latestPredicateFactory = func () dockproc.Predicate {
	return dockproc.LatestPredicateNew()
}

var latestUnpinnedFactory = func () dockproc.Predicate {
	return dockproc.UnpinnedPredicateNew()
}

var anyPredicateFactory = func () dockproc.Predicate {
	return dockproc.AnyPredicateNew()
}

var domainsPredicateFactory = func(domains []string) dockproc.Predicate {
	return dockproc.DomainsPredicateNew(domains)
}
var namePredicateFactory = func(names []string) dockproc.Predicate {
	return dockproc.NamesPredicateNew(names)
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
