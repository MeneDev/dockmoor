package resolver

import (
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDockerRegistryResolver_Resolve(t *testing.T) {
	resolver := DockerRegistryResolverNew()

	references, e := resolver.Resolve(dockref.MustParse("nginx"))
	assert.Nil(t, e)
	assert.NotNil(t, references)
	lenOfRefs := len(references)
	assert.True(t, lenOfRefs > 0)
}
