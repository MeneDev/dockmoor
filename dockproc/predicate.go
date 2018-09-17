package dockproc

import "github.com/MeneDev/dockmoor/dockref"

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
