package dockproc

import (
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/docker/distribution/reference"
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

func AnyPredicateNew() Predicate {
	return anyPredicate{}
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

func LatestPredicateNew() Predicate {
	return latestPredicate{}
}

var _ Predicate = (*unpinnedPredicate)(nil)

type unpinnedPredicate struct {
}

func (unpinnedPredicate) Matches(ref dockref.Reference) bool {
	return ref.DigestString() == ""
}

func UnpinnedPredicateNew() Predicate {
	return unpinnedPredicate{}
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

func UntaggedPredicateNew() Predicate {
	return untaggedPredicate{}
}

var _ Predicate = (*domainsPredicate)(nil)

type domainsPredicate struct {
	domains []string
}

func (p domainsPredicate) Matches(ref dockref.Reference) bool {
	for _, v := range p.domains {
		if v == ref.Domain() {
			return true
		}
	}
	return false
}

func DomainsPredicateNew(domains []string) Predicate {
	return domainsPredicate{domains: domains}
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
		ref2, _ := dockref.FromOriginal(v)

		if ref.Domain() == "docker.io" && ref2.Domain() == "docker.io" {
			if ref.Name() == ref2.Name() {
				return true
			}
		} else {
			// Ignore everything that is not considered the name,
			// e.g. docker.io/library/ or the domain-name
			fam1 := reference.FamiliarName(ref.Named())
			fam2 := reference.FamiliarName(ref2.Named())

			if ref.Path() == ref2.Path() ||
				fam2 == ref.Path() ||
				fam1 == ref2.Path() {
				return true
			}
		}
	}
	return false
}

func NamesPredicateNew(names []string) Predicate {
	return namesPredicate{names: names}
}

var _ Predicate = (*tagsPredicate)(nil)

type tagsPredicate struct {
	tags []string
}

func (p tagsPredicate) Matches(ref dockref.Reference) bool {
	for _, tag := range p.tags {
		if tag == ref.Tag() {
			return true
		}
	}
	return false
}

func TagsPredicateNew(tags []string) Predicate {
	return tagsPredicate{tags: tags}
}

var _ Predicate = (*digestsPredicate)(nil)

type digestsPredicate struct {
	digests []string
}

func (p digestsPredicate) Matches(ref dockref.Reference) bool {
	for _, digest := range p.digests {
		if digest == ref.DigestString() || "sha256:" + digest == ref.DigestString() {
			return true
		}
	}
	return false
}

func DigestsPredicateNew(digests []string) Predicate {
	return digestsPredicate{digests: digests}
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

func AndPredicateNew(predicates []Predicate) Predicate {
	return andPredicate{predicates: predicates}
}
