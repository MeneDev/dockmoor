package resolver

import (
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDockerRegistryResolver_FindAllTags(t *testing.T) {
	resolver := DockerRegistryResolverNew(dockref.ResolverOptions{Mode: dockref.ResolveModeUnchanged})

	references, e := resolver.FindAllTags(dockref.MustParse("nginx"))
	assert.Nil(t, e)
	assert.NotNil(t, references)
	lenOfRefs := len(references)
	assert.True(t, lenOfRefs > 0)
}

func TestDockerRegistryResolver_Resolve_resolves_versions_to_most_exact_version(t *testing.T) {
	resolver := DockerRegistryResolverNew(dockref.ResolverOptions{Mode: dockref.ResolveModeUnchanged})

	type TestCaseResult struct {
		tag string
		dig string
	}

	t.Run("SemVer versions", func(t *testing.T) {
		parentTestCase := t.Name()
		results := map[string]TestCaseResult{
			"menedev/testimagea:2.0.0": {"2.0.0", "sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1"},
			"menedev/testimagea:2.0":   {"2.0.0", "sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1"},
			"menedev/testimagea":       {"2.0.0", "sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1"},
		}

		run := func(t *testing.T) {
			testCase := t.Name()[len(parentTestCase)+1:]
			expected := results[testCase]

			result, e := resolver.Resolve(dockref.MustParse(testCase))
			assert.Nil(t, e)
			assert.NotNil(t, result)
			if result != nil {
				assert.Equal(t, expected.tag, result.Tag())
				assert.Equal(t, expected.dig, result.DigestString())
			}
		}
		t.Run("menedev/testimagea:2.0.0", run)
		t.Run("menedev/testimagea", run)
		t.Run("menedev/testimagea:2.0", run)
	})

	t.Run("Named versions", func(t *testing.T) {
		parentTestCase := t.Name()
		results := map[string]TestCaseResult{
			"menedev/testimagea:mainline": {"mainline", "sha256:1e2b1cc7d366650a93620ca3cc8691338ed600ababf90a0e5803e1ee32486624"},
			"menedev/testimagea:edge":     {"edge", "sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1"},
		}

		run := func(t *testing.T) {
			testCase := t.Name()[len(parentTestCase)+1:]
			expected := results[testCase]

			result, e := resolver.Resolve(dockref.MustParse(testCase))
			assert.Nil(t, e)
			assert.NotNil(t, result)
			if result != nil {
				assert.Equal(t, expected.tag, result.Tag())
				assert.Equal(t, expected.dig, result.DigestString())
			}
		}

		t.Run("menedev/testimagea:mainline", run)
		t.Run("menedev/testimagea:edge", run)
	})
}
