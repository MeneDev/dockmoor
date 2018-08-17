package dockproc

import (
	"github.com/MeneDev/dockfix/dockfmt"
	"github.com/MeneDev/dockfix/dockref"
	"errors"
)

type Accumulator interface {
	Accumulate(format dockfmt.FormatProcessor) bool
}

var _ Accumulator = (*findAccumulator)(nil)

type findAccumulator struct {
	result    bool
	predicate Predicate
}

func (accumulator *findAccumulator) Accumulate(format dockfmt.FormatProcessor) bool {

	found := false
	var processor dockfmt.ImageNameProcessor = func(r dockref.Reference) (string, error) {
		if accumulator.predicate.Matches(r) {
			found = true
		}
		return "", nil
	}

	format.Process(processor)
	return found
}

func FindAccumulatorNew(predicate Predicate) (*findAccumulator, error) {

	if predicate == nil {
		return nil, errors.New("Parameter predicate must not be null")
	}

	return &findAccumulator{
		predicate: predicate,
	}, nil
}

func (accumulator *findAccumulator) Result() bool {
	return accumulator.result
}
