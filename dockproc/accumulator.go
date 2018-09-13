package dockproc

import (
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockref"
	"errors"
	"github.com/sirupsen/logrus"
	"io"
)

type Accumulator interface {
	Accumulate(format dockfmt.FormatProcessor) error
}


var _ Accumulator = (*matchesAccumulator)(nil)

type matchesAccumulator struct {
	matches   []dockref.Reference
	predicate Predicate
	log       *logrus.Logger
	stdout    io.Writer
}

func MatchesAccumulatorNew(predicate Predicate, log *logrus.Logger, stdout io.Writer) (*matchesAccumulator, error) {

	if predicate == nil {
		return nil, errors.New("Parameter predicate must not be null")
	}

	return &matchesAccumulator{
		predicate: predicate,
		log: log,
		stdout: stdout,
	}, nil
}

func (accumulator *matchesAccumulator) Accumulate(format dockfmt.FormatProcessor) (err error) {

	matches := make([]dockref.Reference, 0)

	var processor dockfmt.ImageNameProcessor = func(r dockref.Reference) (string, error) {
		if accumulator.predicate.Matches(r) {
			matches = append(matches, r)
		}
		return "", nil
	}

	err = format.Process(processor)

	accumulator.matches = matches

	return
}

func (accumulator *matchesAccumulator) Matches() []dockref.Reference {
	return accumulator.matches
}
