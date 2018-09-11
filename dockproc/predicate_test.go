package dockproc

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/MeneDev/dockmoor/dockref"
)

func TestAnyPredicate(t *testing.T) {

	predicate := AnyPredicateNew()

	originals := []string{"nginx", "nginx:latest", "nginx:1.15.2-alpine-perl",
		"mongo:3.4.16-windowsservercore-ltsc2016", "d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"}
	for _, original := range originals {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}
}

func TestLatestPredicate(t *testing.T) {

	predicate := LatestPredicateNew()

	shouldMatches := []string{"nginx", "nginx:latest"}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{"nginx:1.15.2-alpine-perl",
		"mongo:3.4.16-windowsservercore-ltsc2016", "d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}

}

func TestUnpinnedPredicate(t *testing.T) {

	predicate := UnpinnedPredicateNew()

	shouldMatches := []string{"nginx", "nginx:latest", "nginx:1.15.2-alpine-perl",
		"mongo:3.4.16-windowsservercore-ltsc2016"}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}

}
