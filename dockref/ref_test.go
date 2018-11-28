package dockref

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestInvalid(t *testing.T) {
	ref, e := FromOriginal("invalid:reference:format")
	assert.Nil(t, ref)
	assert.Error(t, e)
}

func TestWellknownNames(t *testing.T) {
	originals := []string{"nginx", "alpine", "httpd"}
	for _, original := range originals {
		t.Run("Parses "+original, func(t *testing.T) {
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
		t.Run("Parses "+original, func(t *testing.T) {
			ref, e := FromOriginal(original)
			assert.Nil(t, e)

			expected := "docker.io/library/" + original
			assert.Equal(t, expected, ref.Name()+":"+ref.Tag())
		})
	}
}

func TestOriginalsAreUnchanged(t *testing.T) {
	originals := []string{"nginx", "nginx:latest", "nginx:1.15.2-alpine-perl"}
	for _, original := range originals {
		t.Run(original+" remains "+original, func(t *testing.T) {
			ref, e := FromOriginal(original)
			assert.Nil(t, e)

			expected := original
			assert.Equal(t, expected, ref.Original())
		})
	}
}

func TestDigestOnly(t *testing.T) {
	originals := []string{"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"}
	for _, original := range originals {
		t.Run("Parses "+original, func(t *testing.T) {
			ref, e := FromOriginal(original)
			assert.Nil(t, e)
			assert.Empty(t, ref.Name())
			assert.Empty(t, ref.Tag())

			expected := "sha256:" + original
			assert.Equal(t, expected, ref.DigestString())
			assert.Equal(t, expected, string(ref.Digest()))
		})
	}
}

func TestNameAndDigest(t *testing.T) {
	originals := []string{"nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"}
	for _, original := range originals {
		t.Run("Parses "+original, func(t *testing.T) {
			ref, e := FromOriginal(original)
			assert.Nil(t, e)
			assert.Equal(t, "docker.io/library/nginx", ref.Name())
			assert.Empty(t, ref.Tag())

			assert.Equal(t, "sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240", ref.DigestString())
			assert.Equal(t, "sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240", string(ref.Digest()))
		})
	}
}

func TestNameAndTagAndDigest(t *testing.T) {
	originals := []string{"nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"}
	for _, original := range originals {
		t.Run("Parses "+original, func(t *testing.T) {
			ref, e := FromOriginal(original)
			assert.Nil(t, e)
			assert.Equal(t, "docker.io/library/nginx", ref.Name())
			assert.Equal(t, "1.2", ref.Tag())

			assert.Equal(t, "sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240", ref.DigestString())
			assert.Equal(t, "sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240", string(ref.Digest()))
		})
	}
}

func TestDomainAndNameAndTagAndDigest(t *testing.T) {
	originals := []string{"nginx"}
	for _, original := range originals {
		t.Run("Parses "+original, func(t *testing.T) {
			ref, e := FromOriginal(original)
			assert.Nil(t, e)
			assert.Equal(t, "docker.io", ref.Domain())
			assert.Equal(t, "library/nginx", ref.Path())
		})
	}
	originals = []string{"my.com/nginx"}
	for _, original := range originals {
		t.Run("Parses "+original, func(t *testing.T) {
			ref, e := FromOriginal(original)
			assert.Nil(t, e)
			assert.Equal(t, "my.com", ref.Domain())
			assert.Equal(t, "nginx", ref.Path())
		})
	}
}

func TestDockref_Named(t *testing.T) {
	t.Run("Returns Named for named references", func(t *testing.T) {
		ref, e := FromOriginal("nginx")
		assert.Nil(t, e)
		assert.NotNil(t, ref.Named())
	})
	t.Run("Returns nil for unnamed references", func(t *testing.T) {
		ref, e := FromOriginal("d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
		assert.Nil(t, e)
		assert.Nil(t, ref.Named())
	})
}

func TestFromOriginalNoError(t *testing.T) {
	reference := FromOriginalNoError("nginx")
	assert.NotNil(t, reference)
}

func TestDeliberatelyUnsued(t *testing.T) {
	deliberatelyUnsued(errors.New("error"))
	deliberatelyUnsued(nil)
	// didn't crash
	assert.True(t, true)
}

func TestDockref_Format(t *testing.T) {
	func() {
		original := "nginx"

		t.Run(original, func(t *testing.T) {
			nginx := FromOriginalNoError(original)
			format := nginx.Format()

			t.Run("has name", func(t *testing.T) {
				assert.True(t, format.hasName())
			})
			t.Run("has no tag", func(t *testing.T) {
				assert.False(t, format.hasTag())
			})
			t.Run("has no domain", func(t *testing.T) {
				assert.False(t, format.hasDomain())
			})
			t.Run("has no digest", func(t *testing.T) {
				assert.False(t, format.hasDigest())
			})
		})
	}()

	func() {
		original := "nginx:latest"

		t.Run(original, func(t *testing.T) {
			nginx := FromOriginalNoError(original)
			format := nginx.Format()

			t.Run("has name", func(t *testing.T) {
				assert.True(t, format.hasName())
			})
			t.Run("has tag", func(t *testing.T) {
				assert.True(t, format.hasTag())
			})
			t.Run("has no domain", func(t *testing.T) {
				assert.False(t, format.hasDomain())
			})
			t.Run("has no digest", func(t *testing.T) {
				assert.False(t, format.hasDigest())
			})
		})
	}()

	func() {
		original := "docker.io/library/nginx"

		t.Run(original, func(t *testing.T) {
			nginx := FromOriginalNoError(original)
			format := nginx.Format()

			t.Run("has name", func(t *testing.T) {
				assert.True(t, format.hasName())
			})
			t.Run("has no tag", func(t *testing.T) {
				assert.False(t, format.hasTag())
			})
			t.Run("has domain", func(t *testing.T) {
				assert.True(t, format.hasDomain())
			})
			t.Run("has no digest", func(t *testing.T) {
				assert.False(t, format.hasDigest())
			})
		})
	}()

	func() {
		original := "docker.io/library/nginx:latest"

		t.Run(original, func(t *testing.T) {
			nginx := FromOriginalNoError(original)
			format := nginx.Format()

			t.Run("has name", func(t *testing.T) {
				assert.True(t, format.hasName())
			})
			t.Run("has tag", func(t *testing.T) {
				assert.True(t, format.hasTag())
			})
			t.Run("has domain", func(t *testing.T) {
				assert.True(t, format.hasDomain())
			})
			t.Run("has no digest", func(t *testing.T) {
				assert.False(t, format.hasDigest())
			})
		})
	}()

	func() {
		original := "docker.io/library/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"

		t.Run(original, func(t *testing.T) {
			nginx := FromOriginalNoError(original)
			format := nginx.Format()

			t.Run("has name", func(t *testing.T) {
				assert.True(t, format.hasName())
			})
			t.Run("has tag", func(t *testing.T) {
				assert.True(t, format.hasTag())
			})
			t.Run("has domain", func(t *testing.T) {
				assert.True(t, format.hasDomain())
			})
			t.Run("has no digest", func(t *testing.T) {
				assert.True(t, format.hasDigest())
			})
		})
	}()

	func() {
		original := "d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"

		t.Run(original, func(t *testing.T) {
			nginx := FromOriginalNoError(original)
			format := nginx.Format()

			t.Run("has no name", func(t *testing.T) {
				assert.False(t, format.hasName())
			})
			t.Run("has no tag", func(t *testing.T) {
				assert.False(t, format.hasTag())
			})
			t.Run("has no domain", func(t *testing.T) {
				assert.False(t, format.hasDomain())
			})
			t.Run("has digest", func(t *testing.T) {
				assert.True(t, format.hasDigest())
			})
		})
	}()
}

func TestDockref_Formatted(t *testing.T) {

	t.Run("reformatting with same format is equal", func(t *testing.T) {
		originals := []string{
			"nginx",
			"nginx:latest",
			"docker.io/library/nginx",
			"docker.io/library/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
			"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		}

		for _, original := range originals {
			t.Run(original, func(t *testing.T) {
				ref := FromOriginalNoError(original)
				format := ref.Format()

				reference, err := ref.WithRequestedFormat(format)
				assert.Nil(t, err)

				formatted := reference.Formatted()
				assert.Equal(t, original, formatted)
			})
		}
	})
}

func TestDockref_String(t *testing.T) {

	t.Run("String is equal to Original", func(t *testing.T) {
		originals := []string{
			"nginx",
			"nginx:latest",
			"docker.io/library/nginx",
			"docker.io/library/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
			"d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240",
		}

		for _, original := range originals {
			t.Run(original, func(t *testing.T) {
				ref := FromOriginalNoError(original)

				formatted := ref.String()
				assert.Equal(t, original, formatted)
			})
		}
	})
}

func TestDockref_WithRequestedFormat(t *testing.T) {
	t.Run("invalid returns error", func(t *testing.T) {
		r := FromOriginalNoError("docker.io/library/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
		_, e := r.WithRequestedFormat(FormatHasName | FormatHasDigest | FormatHasTag | FormatHasDomain + 1)
		assert.Error(t, e)
	})

	t.Run("well known name only", func(t *testing.T) {
		r := FromOriginalNoError("docker.io/library/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
		rName, e := r.WithRequestedFormat(FormatHasName)
		assert.Nil(t, e)
		assert.Equal(t, "nginx", rName.String())
	})

	t.Run("well known digest only", func(t *testing.T) {
		r := FromOriginalNoError("docker.io/library/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
		rName, e := r.WithRequestedFormat(FormatHasDigest)
		assert.Nil(t, e)
		assert.Equal(t, "d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240", rName.String())
	})

	t.Run("well known tag and digest only", func(t *testing.T) {
		r := FromOriginalNoError("docker.io/library/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
		rName, e := r.WithRequestedFormat(FormatHasTag | FormatHasDigest)
		assert.Nil(t, e)
		assert.Equal(t, "nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240", rName.String())
	})

	t.Run("own-domain", func(t *testing.T) {
		r := FromOriginalNoError("example.com/library/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")

		rNameOwn, e := r.WithRequestedFormat(FormatHasName)
		assert.Nil(t, e)
		assert.Equal(t, "example.com/library/nginx", rNameOwn.String())
	})

	t.Run("own-domain digest-only", func(t *testing.T) {
		r := FromOriginalNoError("example.com/library/nginx:1.2@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")

		rNameOwn, e := r.WithRequestedFormat(FormatHasDigest)
		assert.Nil(t, e)
		assert.Equal(t, "example.com/library/nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240", rNameOwn.String())
	})

	t.Run("digest-only remains digest-only", func(t *testing.T) {
		r := FromOriginalNoError("d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")

		r, e := r.WithRequestedFormat(FormatHasName)
		assert.Nil(t, e)
		assert.Equal(t, r.Format(), FormatHasDigest)
		//assert.Equal(t, "d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240", r.String())

	})
}

func TestFromAlgoDigest(t *testing.T) {
	t.Run("Invalid algorithm", func(t *testing.T) {
		ref, e := ParseAlgoDigest("invalid:3247732819d6cd7af0c45a05b30d0b147f05a25ee2e83d7b9707ee25fcdd0f58")
		assert.Nil(t, ref)
		assert.Error(t, e)
	})
	t.Run("Invalid hex", func(t *testing.T) {
		ref, e := ParseAlgoDigest("sha256:0000")
		assert.Nil(t, ref)
		assert.Error(t, e)
	})
	t.Run("Valid algo digest", func(t *testing.T) {
		ref, e := ParseAlgoDigest("sha256:3247732819d6cd7af0c45a05b30d0b147f05a25ee2e83d7b9707ee25fcdd0f58")
		assert.NotNil(t, ref)
		assert.Nil(t, e)
	})

}

func TestMostPreciseTag(t *testing.T) {
	type TestCase struct {
		list     []string
		expected string
	}

	t.Run("Nil slice returns nil", func(t *testing.T) {
		result, err := MostPreciseTag(nil, nil)
		assert.Nil(t, result)
		assert.Error(t, err)
	})
	t.Run("Nil element returns nil", func(t *testing.T) {
		result, err := MostPreciseTag([]Reference{nil, FromOriginalNoError("nginx")}, nil)
		assert.Nil(t, result)
		assert.Error(t, err)
	})

	cases := []TestCase{
		{list: []string{"nginx:latest"}, expected: "nginx:latest"},
		{list: []string{"nginx:latest", "nginx"}, expected: "nginx:latest"},
		{list: []string{"nginx", "nginx:latest"}, expected: "nginx:latest"},
		{list: []string{"nginx:latest", "nginx:latest"}, expected: "nginx:latest"},
		{list: []string{"nginx:latest", "nginx:notlatest", "nginx:latest"}, expected: "nginx:notlatest"},
		{list: []string{"nginx:latest", "nginx:notlatest", "nginx:1", "nginx:latest"}, expected: "nginx:1"},
		{list: []string{"nginx:1", "nginx:1.1"}, expected: "nginx:1.1"},
		{list: []string{"nginx:1", "nginx:1.1", "nginx:2"}, expected: "nginx:2"},
		{list: []string{"nginx:1", "nginx:1.1", "nginx:2", "nginx:2.1"}, expected: "nginx:2.1"},
		{list: []string{"nginx:1", "nginx:1.1", "nginx:1.1-rc1"}, expected: "nginx:1.1"},
		{list: []string{"nginx:1.1-rc1", "nginx:1.1-beta"}, expected: "nginx:1.1-rc1"},
		{list: []string{"nginx:1.1-beta", "nginx:1.1-alpha"}, expected: "nginx:1.1-beta"},
		{list: []string{"nginx:latest", "nginx:1.1-alpha"}, expected: "nginx:1.1-alpha"},
		{list: []string{"img:20181120", "img:20181121", "img:20181119"}, expected: "img:20181121"},
		{list: []string{"img:a", "img:aaa", "img:bb"}, expected: "img:aaa"},
		{list: []string{"img:aaa", "img:aab"}, expected: "img:aab"},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.list, ", ")+" to "+c.expected, func(t *testing.T) {
			refs := make([]Reference, 0)
			for _, refStr := range c.list {
				ref := FromOriginalNoError(refStr)
				refs = append(refs, ref)
			}
			expected := FromOriginalNoError(c.expected)
			result, err := MostPreciseTag(refs, nil)
			assert.Equal(t, expected, result)
			assert.Nil(t, err)
		})

	}

	cases = []TestCase{
		{list: []string{"nginx:latest"}, expected: "nginx:latest"},
		{list: []string{"nginx:latest", "nginx"}, expected: "nginx:latest"},
		{list: []string{"nginx", "nginx:latest"}, expected: "nginx:latest"},
		{list: []string{"nginx:latest", "nginx:latest"}, expected: "nginx:latest"},
		{list: []string{"nginx:latest", "nginx:notlatest", "nginx:latest"}, expected: "nginx:notlatest"},
		{list: []string{"nginx:notlatest", "nginx:1"}, expected: "nginx:1"},
		{list: []string{"nginx:1", "nginx:1.1"}, expected: "nginx:1.1"},
	}

	for _, c := range cases {
		t.Run("Not warning for "+strings.Join(c.list, ", "), func(t *testing.T) {
			refs := make([]Reference, 0)
			for _, refStr := range c.list {
				ref := FromOriginalNoError(refStr)
				refs = append(refs, ref)
			}

			log := logrus.New()
			stdout := bytes.NewBuffer(nil)
			log.SetOutput(stdout)

			reference, err := MostPreciseTag(refs, log)

			assert.NotNil(t, reference)
			assert.Nil(t, err)

			str := stdout.String()
			assert.Empty(t, str)
		})
	}

	cases = []TestCase{
		{list: []string{"img:a", "img:aaa", "img:bb"}, expected: "img:aaa"},
		{list: []string{"img:aaa", "img:aab"}, expected: "img:aab"},
	}

	for _, c := range cases {
		t.Run("Warning for "+list(c.list), func(t *testing.T) {
			log := logrus.New()
			stdout := bytes.NewBuffer(nil)
			log.SetOutput(stdout)

			refs := toRefs(c.list)
			reference, err := MostPreciseTag(refs, log)

			assert.NotNil(t, reference)
			assert.Nil(t, err)

			str := stdout.String()
			assert.NotEmpty(t, str)
		})
	}
}

func toRefs(strs []string) []Reference {
	refs := make([]Reference, 0)
	for _, refStr := range strs {
		ref := FromOriginalNoError(refStr)
		refs = append(refs, ref)
	}

	return refs
}

func TestFindRelevantTagsForReference(t *testing.T) {
	type TestCase struct {
		ref      string
		list     []string
		expected []string
	}

	run := func(c TestCase) func(t *testing.T) {
		return func(t *testing.T) {
			ref := FromOriginalNoError(c.ref)
			refs := toRefs(c.list)
			expected := toRefs(c.expected)

			t.Name()

			found, e := FindRelevantTagsForReference(ref, refs, nil)
			assert.Nil(t, e)

			assertSameSet(t, expected, found)
		}
	}

	t.Run("empty list return empty list", run(TestCase{
		ref:      "img",
		list:     []string{},
		expected: []string{},
	}))

	t.Run("untagged, all same name", run(TestCase{
		ref:      "img",
		list:     []string{"img:latest", "img:1", "img:2"},
		expected: []string{"img:latest", "img:1", "img:2"},
	}))

	t.Run("different name", run(TestCase{
		ref:      "img",
		list:     []string{"other:latest", "img:latest"},
		expected: []string{"img:latest"},
	}))

	t.Run("different domain", run(TestCase{
		ref:      "img",
		list:     []string{"example.com/img:latest", "img:latest"},
		expected: []string{"img:latest"},
	}))

	t.Run("different other name", run(TestCase{
		ref:      "other",
		list:     []string{"other:latest", "img:latest"},
		expected: []string{"other:latest"},
	}))

	t.Run("same name, same variant", run(TestCase{
		ref:      "img:1-something",
		list:     []string{"img:1-something", "img:1.2-something", "img:1.2.1-something", "img:2.2.1-something"},
		expected: []string{"img:1-something", "img:1.2-something", "img:1.2.1-something", "img:2.2.1-something"},
	}))

	t.Run("same name, different variant", run(TestCase{
		ref:      "img:1-something",
		list:     []string{"img:1-something", "img:1.2-something", "img:2.2-something", "img:1.2.1-other", "img:2.2.1-other"},
		expected: []string{"img:1-something", "img:1.2-something", "img:2.2-something"},
	}))

	t.Run("same name, different variant with common post-fix", run(TestCase{
		ref:      "img:1-something",
		list:     []string{"img:1-something", "img:1.2-something", "img:2.2-something", "img:1.2.1-other-something", "img:2.2.1-other-something"},
		expected: []string{"img:1-something", "img:1.2-something", "img:2.2-something"},
	}))

	t.Run("same name, different variant with common pre-fix", run(TestCase{
		ref:      "img:1-something",
		list:     []string{"img:1-something", "img:1.2-something", "img:2.2-something", "img:1.2.1-something-other", "img:2.2.1-something-other"},
		expected: []string{"img:1-something", "img:1.2-something", "img:2.2-something"},
	}))

}

func assertSameSet(t *testing.T, expected []Reference, slice []Reference) {
	t.Helper()
	for _, item := range slice {
		assert.Contains(t, expected, item)
	}
}

func list(ss []string) string {
	join := strings.Join(ss, ", ")
	return join
}

func TestSplitVersionAndVariant(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		tag := "1"
		expVersion := "1"
		expVar := ""
		version, variant := splitVersionAndVariant(tag)
		assert.Equal(t, expVersion, version)
		assert.Equal(t, expVar, variant)
	})
	t.Run("1.1.1", func(t *testing.T) {
		tag := "1.1.1"
		expVersion := "1.1.1"
		expVar := ""
		version, variant := splitVersionAndVariant(tag)
		assert.Equal(t, expVersion, version)
		assert.Equal(t, expVar, variant)
	})
	t.Run("something", func(t *testing.T) {
		tag := "something"
		expVersion := ""
		expVar := "something"
		version, variant := splitVersionAndVariant(tag)
		assert.Equal(t, expVersion, version)
		assert.Equal(t, expVar, variant)
	})
	t.Run("1-something", func(t *testing.T) {
		tag := "1-something"
		expVersion := "1"
		expVar := "something"
		version, variant := splitVersionAndVariant(tag)
		assert.Equal(t, expVersion, version)
		assert.Equal(t, expVar, variant)
	})
	t.Run("1.15.6-alpine-perl", func(t *testing.T) {
		tag := "1.15.6-alpine-perl"
		expVersion := "1.15.6"
		expVar := "alpine-perl"
		version, variant := splitVersionAndVariant(tag)
		assert.Equal(t, expVersion, version)
		assert.Equal(t, expVar, variant)
	})
}
