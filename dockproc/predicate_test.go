package dockproc

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/MeneDev/dockfix/dockref"
)

func TestAnyPredicate(t *testing.T) {

	predicateNew := AnyPredicateNew()

	originals := []string{"nginx", "nginx:latest", "nginx:1.15.2-alpine-perl",
		"mongo:3.4.16-windowsservercore-ltsc2016", "d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"}
	for _, original := range originals {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicateNew.Matches(ref))
		})
	}
}
