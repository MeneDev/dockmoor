package dockref

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWellknownNames(t *testing.T) {
	originals := []string{"nginx", "alpine", "httpd"}
	for _, original := range originals {
		t.Run("Parses " + original, func(t *testing.T) {
			ref, e := FromOriginal(original)
			assert.Nil(t, e)
			assert.Equal(t, "", ref.Tag())

			expected := "docker.io/library/" + original
			assert.Equal(t, expected, ref.Name())
		})
	}
}

func TestWellknownTaggedNames(t *testing.T) {
	originals := []string{"nginx:latest", "nginx:1.15.2-alpine-perl", "mongo:3.4.16-windowsservercore-ltsc2016"}
	for _, original := range originals {
		t.Run("Parses " + original, func(t *testing.T) {
			ref, e := FromOriginal(original)
			assert.Nil(t, e)

			expected := "docker.io/library/" + original
			assert.Equal(t, expected, ref.Name() + ":" + ref.Tag())
		})
	}
}

func TestDigestOnly(t *testing.T) {
	originals := []string{"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"}
	for _, original := range originals {
		t.Run("Parses " + original, func(t *testing.T) {
			ref, e := FromOriginal(original)
			assert.Nil(t, e)
			assert.Empty(t, ref.Name())
			assert.Empty(t, ref.Tag())

			expected := "sha256:" + original
			assert.Equal(t, expected, ref.DigestString())
		})
	}
}
