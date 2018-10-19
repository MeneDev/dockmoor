package dockproc

import (
	"errors"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/sirupsen/logrus"
	"io"
)

func NullHandler(r dockref.Reference) error { return nil }

type Accumulator interface {
	Accumulate(format dockfmt.FormatProcessor, handler func(r dockref.Reference) error) error
	Matches() []dockref.Reference
	Results() []string
}

var _ Accumulator = (*matchesAccumulator)(nil)

type matchesAccumulator struct {
	matches   []dockref.Reference
	results   []string
	predicate Predicate
	log       *logrus.Logger
	stdout    io.Writer
}

func MatchesAccumulatorNew(predicate Predicate, log *logrus.Logger, stdout io.Writer) (Accumulator, error) {

	if predicate == nil {
		return nil, errors.New("Parameter predicate must not be null")
	}

	return &matchesAccumulator{
		predicate: predicate,
		log:       log,
		stdout:    stdout,
	}, nil
}

func (ma *matchesAccumulator) Accumulate(format dockfmt.FormatProcessor, handler func(r dockref.Reference) error) (err error) {

	matches := make([]dockref.Reference, 0)

	var processor dockfmt.ImageNameProcessor = func(r dockref.Reference) (string, error) {
		if ma.predicate.Matches(r) {
			handler(r)
			matches = append(matches, r)
		}
		return "", nil
	}

	err = format.Process(processor)

	ma.matches = matches

	return
}

func (ma *matchesAccumulator) Matches() []dockref.Reference {
	return ma.matches
}
func (ma *matchesAccumulator) Results() []string {
	return ma.results
}

type pinAccumulator struct {
	*matchesAccumulator
	repo   dockref.Repository
	format dockref.Format
}

func PinAccumulatorNew(predicate Predicate, log *logrus.Logger, stdout io.Writer, repo dockref.Repository, format dockref.Format) (Accumulator, error) {
	accumulator, e := MatchesAccumulatorNew(predicate, log, stdout)
	if e != nil {
		return accumulator, e
	}

	ma, _ := accumulator.(*matchesAccumulator)

	return &pinAccumulator{
		matchesAccumulator: ma,
		repo:               repo,
		format:             format,
	}, nil
}

func (pa *pinAccumulator) Accumulate(format dockfmt.FormatProcessor, handler func(r dockref.Reference) error) (err error) {

	e := pa.matchesAccumulator.Accumulate(format, NullHandler)
	if e != nil {
		return e
	}

	matches := make([]dockref.Reference, 0)
	results := make([]string, 0)

	var processor dockfmt.ImageNameProcessor = func(r dockref.Reference) (string, error) {
		if pa.predicate.Matches(r) {
			matches = append(matches, r)
			res, e := pa.repo.Resolve(r)
			if e != nil {
				return "", e
			}
			result := res.Formatted(r.Format() | pa.format)
			results = append(results, result)
			return result, nil
		}
		return "", nil
	}

	err = format.Process(processor)

	pa.matches = matches
	pa.results = results

	return
}

func (pa *pinAccumulator) Matches() []dockref.Reference {
	return pa.matches
}

func (pa *pinAccumulator) Results() []string {
	return pa.results
}
