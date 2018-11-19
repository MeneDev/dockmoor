package dockref

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDockerDaemonRegistry_Resolve(t *testing.T) {
	repo := DockerDaemonRepositoryNew()

	type T struct {
		name, digest string
	}

	tests := []T{
		{name: "nginx:1.15.5", digest: "nginx@sha256:b73f527d86e3461fd652f62cf47e7b375196063bbbd503e853af5be16597cb2e"},
		{name: "nginx:1.15.6", digest: "nginx@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"},
		{name: "nginx:latest", digest: "nginx@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"},
		{name: "nginx:1.15.5-perl", digest: "nginx@sha256:01c45fbd335b5fcfbfe95777508cc16044e0d6a929f5d531f48ab53ca4556578"},
		{name: "nginx:1.15.5-alpine", digest: "nginx@sha256:ae5da813f8ad7fa785d7668f0b018ecc8c3a87331527a61d83b3b5e816a0f03c"},
		{name: "nginx:1.15.5-alpine-perl", digest: "nginx@sha256:9c632b0423d3ceba7e94a6744a127b694caacb6117238aff033ab6bdc88c1fae"},
		{name: "nginx:1.14.0", digest: "nginx@sha256:8b600a4d029481cc5b459f1380b30ff6cb98e27544fc02370de836e397e34030"},
		{name: "nginx:1.14.0-perl", digest: "nginx@sha256:032acb6025fa581888812e79f4efcd32d008e0ce3dfe56c65f9c1011d93ce920"},
		{name: "nginx:1.14.0-alpine", digest: "nginx@sha256:8976218be775f4244df2a60a169d44606b6978bac4375192074cefc0c7824ddf"},
		{name: "nginx:1.14.0-alpine-perl", digest: "nginx@sha256:c3d6f9a179ba365ab4b41e176623a6fc9cfc2121567131127e43f5660e0c4767"},
	}

	for _, tst := range tests {
		t.Run("Resolves name "+tst.name, func(t *testing.T) {
			ref, e := FromOriginal(tst.name)
			assert.Nil(t, e)

			resolve, e := repo.Resolve(ref)
			assert.Nil(t, e)

			assert.NotNil(t, resolve)
			assert.Equal(t, tst.digest, resolve[0].Formatted(FormatHasName|FormatHasDigest))
		})
	}

	for _, tst := range tests {
		t.Run("Resolves digest "+tst.digest, func(t *testing.T) {
			ref, e := FromOriginal(tst.name)
			assert.Nil(t, e)

			resolve, e := repo.Resolve(ref)
			assert.Nil(t, e)

			assert.NotNil(t, resolve)

			matches := 0
			for _, res := range resolve {
				formatted := res.Formatted(FormatHasName | FormatHasTag)

				if formatted == tst.name {
					matches++
				}
			}

			assert.Equal(t, 1, matches)
		})
	}
}
