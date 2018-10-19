package dockproc

import (
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAnyPredicate(t *testing.T) {

	predicate, e := AnyPredicateNew()
	assert.Nil(t, e)

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

	predicate, e := LatestPredicateNew()
	assert.Nil(t, e)

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

	predicate, e := UnpinnedPredicateNew()
	assert.Nil(t, e)

	shouldMatches := []string{
		"nginx",
		"nginx:latest",
		"nginx:1.15.2-alpine-perl",
		"mongo:3.4.16-windowsservercore-ltsc2016",
		"example.com/image-name:latest",
	}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{
		"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
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

func TestUntaggedPredicate(t *testing.T) {

	predicate, e := UntaggedPredicateNew()
	assert.Nil(t, e)

	shouldMatches := []string{"nginx", "d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{"nginx:latest", "nginx:1.15.2-alpine-perl",
		"mongo:3.4.16-windowsservercore-ltsc2016",
		"nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}

}

func TestDomainsPredicate(t *testing.T) {

	predicate, e := DomainsPredicateNew([]string{"my.com", "my2.com"})
	assert.Nil(t, e)

	shouldMatches := []string{
		"my.com/nginx",
		"my2.com/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my.com/nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my2.com/nginx:latest", "my.com/nginx:1.15.2-alpine-perl",
		"my2.com/mongo:3.4.16-windowsservercore-ltsc2016",
		"my.com/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{"nginx", "d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240", "nginx:latest",
		"nginx:1.15.2-alpine-perl",
		"mongo:3.4.16-windowsservercore-ltsc2016",
		"nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my3.org/nginx", "my3.org/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my3.org/nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240", "my3.org/nginx:latest",
		"my3.org/nginx:1.15.2-alpine-perl",
		"my3.org/mongo:3.4.16-windowsservercore-ltsc2016",
		"my3.org/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}
}

func TestDomainsPredicateWithRegExp(t *testing.T) {

	predicate, e := DomainsPredicateNew([]string{"/my\\./", "/my2\\./"})
	assert.Nil(t, e)

	shouldMatches := []string{
		"my.com/nginx",
		"my2.com/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my.com/nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my2.com/nginx:latest", "my.com/nginx:1.15.2-alpine-perl",
		"my2.com/mongo:3.4.16-windowsservercore-ltsc2016",
		"my.com/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldMatches {
		t.Run("Matches"+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{"nginx", "d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240", "nginx:latest",
		"nginx:1.15.2-alpine-perl",
		"mongo:3.4.16-windowsservercore-ltsc2016",
		"nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my3.org/nginx", "my3.org/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my3.org/nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240", "my3.org/nginx:latest",
		"my3.org/nginx:1.15.2-alpine-perl",
		"my3.org/mongo:3.4.16-windowsservercore-ltsc2016",
		"my3.org/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}
}

func TestNamesPredicate(t *testing.T) {

	predicate, e := NamesPredicateNew([]string{"nginx", "mongo"})
	assert.Nil(t, e)

	shouldMatches := []string{
		"nginx",
		"docker.io/library/nginx",
		"docker.io/mongo:3.4.16-windowsservercore-ltsc2016",
		"mongo:3.4.16-windowsservercore-ltsc2016",
	}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{
		"my.com/nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my2.com/nginx:latest", "my.com/nginx:1.15.2-alpine-perl",
		"my2.com/mongo:3.4.16-windowsservercore-ltsc2016",
		"my.com/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my.com/nginx",
		"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my3.org/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}

}

func TestNamesPredicateWithRegExp(t *testing.T) {

	predicate, e := NamesPredicateNew([]string{"/ngin/", "/mon/"})
	assert.Nil(t, e)

	shouldMatches := []string{
		"nginx",
		"docker.io/library/nginx",
		"docker.io/mongo:3.4.16-windowsservercore-ltsc2016",
		"mongo:3.4.16-windowsservercore-ltsc2016",
	}

	for _, original := range shouldMatches {
		t.Run("Matches"+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{
		"my.com/ngnix@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my2.com/ngnix:latest", "my.com/ngnix:1.15.2-alpine-perl",
		"my2.com/mnogo:3.4.16-windowsservercore-ltsc2016",
		"my.com/ngnix:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my.com/ngnix",
		"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my3.org/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}
}

func TestFamiliarNamesPredicate(t *testing.T) {

	predicate, e := FamiliarNamesPredicateNew([]string{"nginx", "mongo"})
	assert.Nil(t, e)

	shouldMatches := []string{
		"nginx",
		"docker.io/library/nginx",
		"mongo:3.4.16-windowsservercore-ltsc2016",
		"docker.io/library/mongo",
		"docker.io/library/mongo:3.4.16-windowsservercore-ltsc2016",
	}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{
		"my.com/nginx",
		"my.com/nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my2.com/nginx:latest", "my.com/nginx:1.15.2-alpine-perl",
		"my2.com/mongo:3.4.16-windowsservercore-ltsc2016",
		"my.com/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my3.org/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}

}

func TestFamiliarNamesPredicateWithRegExp(t *testing.T) {

	predicate, e := FamiliarNamesPredicateNew([]string{"/ngin/", "/mon/"})
	assert.Nil(t, e)

	shouldMatches := []string{
		"nginx",
		"docker.io/library/nginx",
		"mongo:3.4.16-windowsservercore-ltsc2016",
		"docker.io/library/mongo",
		"docker.io/library/mongo:3.4.16-windowsservercore-ltsc2016",
	}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{
		"my.com/ngnix",
		"my.com/ngnix@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my2.com/ngnix:latest", "my.com/ngnix:1.15.2-alpine-perl",
		"my2.com/mnogo:3.4.16-windowsservercore-ltsc2016",
		"my.com/ngnix:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my3.org/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}
}

func TestTagsPredicate(t *testing.T) {

	predicate, e := TagsPredicateNew([]string{"1.2", "3.4.16-windowsservercore-ltsc2016"})
	assert.Nil(t, e)

	shouldMatches := []string{
		"mongo:3.4.16-windowsservercore-ltsc2016",
		"my2.com/mongo:3.4.16-windowsservercore-ltsc2016",
		"my.com/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{
		"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my3.org/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"nginx",
		"docker.io/library/nginx",
		"my.com/nginx",
		"my.com/nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my2.com/nginx:latest",
		"my.com/nginx:1.15.2-alpine-perl",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}

}

func TestTagsPredicateWithRegExp(t *testing.T) {

	predicate, e := TagsPredicateNew([]string{"/1.2/", "/3.4.16-windowsservercore/"})
	assert.Nil(t, e)

	shouldMatches := []string{
		"mongo:3.4.16-windowsservercore-ltsc2016",
		"my2.com/mongo:3.4.16-windowsservercore-ltsc2016",
		"my.com/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{
		"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my3.org/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"nginx",
		"docker.io/library/nginx",
		"my.com/nginx",
		"my.com/nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my2.com/nginx:latest",
		"my.com/nginx:1.15.2-alpine-perl",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}
}

func TestDigestsPredicate(t *testing.T) {

	predicate, e := DigestsPredicateNew([]string{
		"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b241",
	})
	assert.Nil(t, e)

	shouldMatches := []string{
		"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my.com/nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my.com/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{
		"nginx",
		"docker.io/library/nginx",
		"my.com/nginx",
		"my2.com/nginx:latest",
		"my.com/nginx:1.15.2-alpine-perl",
		"mongo:3.4.16-windowsservercore-ltsc2016",
		"my2.com/mongo:3.4.16-windowsservercore-ltsc2016",
		// That's a name!
		"my3.org/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}
}

func TestPathsPredicate(t *testing.T) {

	predicate, _ := PathsPredicateNew([]string{
		"library/nginx",
		"mongo",
	})

	shouldMatches := []string{
		"nginx",
		"docker.io/library/nginx",
		"my2.com/mongo:3.4.16-windowsservercore-ltsc2016",
	}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{
		"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my.com/nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my.com/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my.com/nginx",
		"my2.com/nginx:latest",
		"my.com/nginx:1.15.2-alpine-perl",
		"mongo:3.4.16-windowsservercore-ltsc2016",
		// That's a name!
		"my3.org/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}
}

func TestPathsPredicateWithRegExp(t *testing.T) {

	predicate, e := PathsPredicateNew([]string{
		"/library/ngin/",
		"/mon/",
	})
	assert.Nil(t, e)

	shouldMatches := []string{
		"nginx",
		"docker.io/library/nginx",
		"my2.com/mongo:3.4.16-windowsservercore-ltsc2016",
	}

	for _, original := range shouldMatches {
		t.Run("Matches "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.True(t, predicate.Matches(ref))
		})
	}

	shouldNotMatches := []string{
		"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my.com/ngnix@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my.com/ngnix:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		"my.com/ngnix",
		"my2.com/ngnix:latest",
		"my.com/ngnix:1.15.2-alpine-perl",
		"mnogo:3.4.16-windowsservercore-ltsc2016",
		// That's a name!
		"my3.org/d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
	}

	for _, original := range shouldNotMatches {
		t.Run("Not matching "+original, func(t *testing.T) {
			ref, e := dockref.FromOriginal(original)

			assert.Nil(t, e)
			assert.False(t, predicate.Matches(ref))
		})
	}
}

var _ Predicate = (*mockPredicate)(nil)

type mockPredicate struct {
	matches bool
}

func (p mockPredicate) Matches(ref dockref.Reference) bool {
	return p.matches
}

func TestAndPredicate_Matches(t *testing.T) {
	ref, _ := dockref.FromOriginal("a")
	and := func(matches ...bool) Predicate {
		predicates := make([]Predicate, 0)

		for _, v := range matches {
			predicates = append(predicates, mockPredicate{v})
		}

		predicate, e := AndPredicateNew(predicates)
		assert.Nil(t, e)

		return predicate
	}

	t.Run("Matches when only predicate matches", func(t *testing.T) {
		assert.True(t, and(true).Matches(ref))
	})

	t.Run("Not matching when only predicate not matching", func(t *testing.T) {
		assert.False(t, and(false).Matches(ref))
	})

	t.Run("Matches when both predicate matches", func(t *testing.T) {
		assert.True(t, and(true, true).Matches(ref))
	})

	t.Run("Not matching one of two predicates not matching", func(t *testing.T) {
		assert.False(t, and(true, false).Matches(ref))
		assert.False(t, and(false, true).Matches(ref))
	})

	t.Run("Not matching one of 5 predicates not matching", func(t *testing.T) {
		assert.False(t, and(true, false, true, true, true).Matches(ref))
		assert.False(t, and(true, true, true, false, true).Matches(ref))
	})

	t.Run("Matches when 5 of 5 predicate matches", func(t *testing.T) {
		assert.True(t, and(true, true, true, true, true).Matches(ref))
	})
}

func TestAndPredicate_Predicates(t *testing.T) {
	p1 := mockPredicate{true}
	p2 := mockPredicate{false}
	p3 := mockPredicate{true}

	p, e := AndPredicateNew([]Predicate{p1, p2, p3})
	assert.Nil(t, e)

	a, ok := p.(andPredicate)
	assert.True(t, ok)

	ps := a.Predicates()

	assert.Contains(t, ps, p1)
	assert.Contains(t, ps, p2)
	assert.Contains(t, ps, p3)
}

func TestDomainsPredicate_returnsErrorForInvalidRegex(t *testing.T) {
	predicateNew, err := DomainsPredicateNew([]string{"/", "/a(b/"})
	assert.Nil(t, predicateNew)
	assert.Error(t, err)
}

func TestNamesPredicate_returnsErrorForInvalidRegex(t *testing.T) {
	predicateNew, err := NamesPredicateNew([]string{"/a(b/"})
	assert.Nil(t, predicateNew)
	assert.Error(t, err)
}

func TestFamiliarNamesPredicate_returnsErrorForInvalidRegex(t *testing.T) {
	predicateNew, err := FamiliarNamesPredicateNew([]string{"/a(b/"})
	assert.Nil(t, predicateNew)
	assert.Error(t, err)
}

func TestPathPredicate_returnsErrorForInvalidRegex(t *testing.T) {
	predicateNew, err := PathsPredicateNew([]string{"/a(b/"})
	assert.Nil(t, predicateNew)
	assert.Error(t, err)
}

func TestTagsPredicate_returnsErrorForInvalidRegex(t *testing.T) {
	predicateNew, err := TagsPredicateNew([]string{"/a(b/"})
	assert.Nil(t, predicateNew)
	assert.Error(t, err)
}
