package dockproc

import "github.com/MeneDev/dockfix/dockref"

type Predicate interface {
	Matches(ref dockref.Reference) bool
}

var _ Predicate = (*anyAccumulator)(nil)

type anyAccumulator struct {

}

func (anyAccumulator) Matches(ref dockref.Reference) bool {
	return true
}

func AnyPredicateNew() Predicate {
	return anyAccumulator{}
}