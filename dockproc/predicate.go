package dockproc

import (
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/docker/distribution/reference"
	"github.com/hashicorp/go-multierror"
	"regexp"
	"strings"
)

type Predicate interface {
	Matches(ref dockref.Reference) bool
}

var _ Predicate = (*anyPredicate)(nil)

type anyPredicate struct {
}

func (anyPredicate) Matches(ref dockref.Reference) bool {
	return true
}

func AnyPredicateNew() (Predicate, error) {
	return anyPredicate{}, nil
}

var _ Predicate = (*latestPredicate)(nil)

type latestPredicate struct {
}

func (latestPredicate) Matches(ref dockref.Reference) bool {
	if ref.Tag() == "latest" {
		return true
	}

	if ref.DigestString() != "" {
		return false
	}

	// Note the edge-case: given only a digest, tag *and* name is empty.
	return ref.Tag() == ""
}

func LatestPredicateNew() (Predicate, error) {
	return latestPredicate{}, nil
}

var _ Predicate = (*unpinnedPredicate)(nil)

type unpinnedPredicate struct {
}

func (unpinnedPredicate) Matches(ref dockref.Reference) bool {
	return ref.DigestString() == ""
}

func UnpinnedPredicateNew() (Predicate, error) {
	return unpinnedPredicate{}, nil
}

//var _ Predicate = (*outdatedPredicate)(nil)
//
//type outdatedPredicate struct {
//}
//
//func (outdatedPredicate) Matches(ref dockref.Reference) bool {
//	panic("not implemented")
//}
//
//func OutdatedPredicateNew() Predicate {
//	return outdatedPredicate{}
//}

var _ Predicate = (*untaggedPredicate)(nil)

type untaggedPredicate struct {
}

func (untaggedPredicate) Matches(ref dockref.Reference) bool {
	return ref.Tag() == ""
}

func UntaggedPredicateNew() (Predicate, error) {
	return untaggedPredicate{}, nil
}

var _ Predicate = (*domainsPredicate)(nil)

type domainsPredicate struct {
	domains []string
}

func (p domainsPredicate) Matches(ref dockref.Reference) bool {
	for _, v := range p.domains {
		if isRegex(v) {
			if regExpMatches(v, ref.Domain()) {
				return true
			}
		} else if v == ref.Domain() {
			return true
		}
	}
	return false
}

func vaildateRegex(regexs []string) error {
	var result *multierror.Error
	for _, v := range regexs {
		if isRegex(v) {
			_, e := regexp.Compile(trimRegexMarkers(v))
			result = multierror.Append(result, e)
		}
	}
	return result.ErrorOrNil()
}

func DomainsPredicateNew(domains []string) (Predicate, error) {
	e := vaildateRegex(domains)
	var predicate Predicate
	if e == nil {
		predicate = domainsPredicate{domains: domains}
	}
	return predicate, e
}

var _ Predicate = (*namesPredicate)(nil)

type namesPredicate struct {
	names []string
}

func (p namesPredicate) Matches(ref dockref.Reference) bool {
	named := ref.Named()
	if named == nil {
		return false
	}

	for _, v := range p.names {
		ref2 := dockref.FromOriginalNoError(v)

		if isRegex(v) {
			if regExpMatches(v, ref.Name()) {
				return true
			}
		} else if ref.Name() == ref2.Name() {
			return true
		}
	}

	return false
}

func NamesPredicateNew(names []string) (Predicate, error) {
	e := vaildateRegex(names)
	var predicate Predicate
	if e == nil {
		predicate = namesPredicate{names: names}
	}
	return predicate, e
}

var _ Predicate = (*familiarNamesPredicate)(nil)

type familiarNamesPredicate struct {
	familiarNames []string
}

func (p familiarNamesPredicate) Matches(ref dockref.Reference) bool {
	named := ref.Named()
	if named == nil {
		return false
	}

	fam1 := reference.FamiliarName(ref.Named())

	for _, v := range p.familiarNames {

		if isRegex(v) {
			if regExpMatches(v, fam1) {
				return true
			}
		} else {
			ref2 := dockref.FromOriginalNoError(v)
			fam2 := reference.FamiliarName(ref2.Named())

			if fam1 == fam2 {
				return true
			}
		}
	}

	return false
}

func FamiliarNamesPredicateNew(familiarNames []string) (Predicate, error) {
	e := vaildateRegex(familiarNames)
	var predicate Predicate
	if e == nil {
		predicate = familiarNamesPredicate{familiarNames: familiarNames}
	}
	return predicate, e
}

var _ Predicate = (*pathsPredicate)(nil)

type pathsPredicate struct {
	paths []string
}

func (p pathsPredicate) Matches(ref dockref.Reference) bool {
	named := ref.Named()
	if named == nil {
		return false
	}

	path := reference.Path(named)
	for _, v := range p.paths {
		if isRegex(v) {
			if regExpMatches(v, path) {
				return true
			}
		} else if path == v {
			return true
		}
	}

	return false
}

func PathsPredicateNew(paths []string) (Predicate, error) {
	e := vaildateRegex(paths)
	var predicate Predicate
	if e == nil {
		predicate = pathsPredicate{paths: paths}
	}
	return predicate, e
}

var _ Predicate = (*tagsPredicate)(nil)

type tagsPredicate struct {
	tags []string
}

func (p tagsPredicate) Matches(ref dockref.Reference) bool {
	for _, tag := range p.tags {
		if isRegex(tag) {
			if regExpMatches(tag, ref.Tag()) {
				return true
			}
		} else if tag == ref.Tag() {
			return true
		}
	}
	return false
}

func TagsPredicateNew(tags []string) (Predicate, error) {
	e := vaildateRegex(tags)
	var predicate Predicate
	if e == nil {
		predicate = tagsPredicate{tags: tags}
	}
	return predicate, e
}

var _ Predicate = (*digestsPredicate)(nil)

type digestsPredicate struct {
	digests []string
}

func (p digestsPredicate) Matches(ref dockref.Reference) bool {
	for _, digest := range p.digests {
		if digest == ref.DigestString() || "sha256:"+digest == ref.DigestString() {
			return true
		}
	}
	return false
}

func DigestsPredicateNew(digests []string) (Predicate, error) {
	return digestsPredicate{digests: digests}, nil
}

type AndPredicate interface {
	Predicate
	Predicates() []Predicate
}

var _ AndPredicate = (*andPredicate)(nil)

type andPredicate struct {
	predicates []Predicate
}

func (a andPredicate) Predicates() []Predicate {
	return a.predicates
}

func (a andPredicate) Matches(ref dockref.Reference) bool {
	for _, p := range a.predicates {
		if !p.Matches(ref) {
			return false
		}
	}
	return true
}

func AndPredicateNew(predicates []Predicate) (Predicate, error) {
	return andPredicate{predicates: predicates}, nil
}

func isRegex(pattern string) bool {
	if len(pattern) < 2 {
		return false
	}
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
		return true
	}
	return false
}

func trimRegexMarkers(pattern string) string {
	return pattern[1 : len(pattern)-1]
}

func regExpMatches(pattern string, ref string) bool {
	compiled := regexp.MustCompile(trimRegexMarkers(pattern))
	matched := compiled.MatchString(ref)
	return matched
}
