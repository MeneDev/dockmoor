package dockproc

import (
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockref"
	"errors"
)

type Accumulator interface {
	Accumulate(format dockfmt.FormatProcessor)
}

var _ Accumulator = (*containsAccumulator)(nil)

type containsAccumulator struct {
	result    bool
	predicate Predicate
}

func (accumulator *containsAccumulator) Accumulate(format dockfmt.FormatProcessor) {

	found := false
	var processor dockfmt.ImageNameProcessor = func(r dockref.Reference) (string, error) {
		if accumulator.predicate.Matches(r) {
			found = true
		}
		return "", nil
	}

	format.Process(processor)

	accumulator.result = found
}

func ContainsAccumulatorNew(predicate Predicate) (*containsAccumulator, error) {

	if predicate == nil {
		return nil, errors.New("Parameter predicate must not be null")
	}

	return &containsAccumulator{
		predicate: predicate,
	}, nil
}

func (accumulator *containsAccumulator) Result() bool {
	return accumulator.result
}
