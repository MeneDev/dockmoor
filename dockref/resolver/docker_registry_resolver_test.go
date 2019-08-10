package resolver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MeneDev/dockmoor/dockref"
	"github.com/MeneDev/dockmoor/dockref/resolver/mocks"
	"github.com/docker/cli/cli/config/credentials"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDockerRegistryResolver_FindAllTags(t *testing.T) {
	withRegistry(t, "registry_test", func(regAddr string) {

		resolver := DockerRegistryResolverNew().(*dockerRegistryResolver)
		resolver.credentialsStoreFactory = func(ref dockref.Reference) (credentials.Store, error) {
			store := &mocks.Store{}
			store.On("Get", mock.AnythingOfType("string")).Return(types.AuthConfig{
				Username: "testuser",
				Password: "testpassword",
			}, nil)
			return store, nil
		}

		references, e := resolver.FindAllTags(dockref.MustParse(regAddr + "menedev/testimagea"))
		assert.Nil(t, e)
		assert.NotNil(t, references)
		lenOfRefs := len(references)
		assert.True(t, lenOfRefs > 0)
	})
}

func TestDockerRegistryResolver_Resolve_resolves_versions_to_most_exact_version(t *testing.T) {

	type TestCaseData struct {
		ref  string
		tag  string
		dig  string
		mode dockref.ResolveMode
	}

	runWith := func(tcd TestCaseData) func(t *testing.T) {

		return func(t *testing.T) {
			resolver := DockerRegistryResolverNew().(*dockerRegistryResolver)
			resolver.credentialsStoreFactory = func(ref dockref.Reference) (credentials.Store, error) {
				store := &mocks.Store{}
				store.On("Get", mock.AnythingOfType("string")).Return(types.AuthConfig{
					Username: "testuser",
					Password: "testpassword",
				}, nil)
				return store, nil
			}

			result, e := resolver.Resolve(dockref.MustParse(tcd.ref))
			assert.Nil(t, e)
			assert.NotNil(t, result)
			if result != nil {
				assert.Equal(t, tcd.tag, result.Tag())
				assert.Equal(t, tcd.dig, result.DigestString())
			}
		}

	}

	runAll := func(regAddr string) {
		t.Run("unchanged menedev/testimagea:2.0.0", runWith(TestCaseData{
			regAddr + "menedev/testimagea:2.0.0",
			"2.0.0",
			"sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1",
			dockref.ResolveModeUnchanged}))

		t.Run("unchanged_menedev/testimagea:2.0", runWith(TestCaseData{
			regAddr + "menedev/testimagea:2.0",
			"2.0",
			"sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1",
			dockref.ResolveModeUnchanged}))

		t.Run("unchanged menedev/testimagea:2", runWith(TestCaseData{
			regAddr + "menedev/testimagea:2",
			"2",
			"sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1",
			dockref.ResolveModeUnchanged}))

		t.Run("unchanged menedev/testimagea", runWith(TestCaseData{
			regAddr + "menedev/testimagea",
			"",
			"sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1",
			dockref.ResolveModeUnchanged}))

		t.Run("unchanged menedev/testimagea", runWith(TestCaseData{
			regAddr + "menedev/testimagea:mainline",
			"mainline", "sha256:1e2b1cc7d366650a93620ca3cc8691338ed600ababf90a0e5803e1ee32486624",
			dockref.ResolveModeUnchanged}))

		t.Run("unchanged menedev/testimagea", runWith(TestCaseData{
			regAddr + "menedev/testimagea:edge",
			"edge",
			"sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1",
			dockref.ResolveModeUnchanged}))
	}

	withRegistry(t, "registry_test", runAll)
}

func withRegistry(t *testing.T, containerName string, callback func(registryAddress string)) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        containerName,
		ExposedPorts: []string{"5000/tcp"},
		Env: map[string]string{
			"REGISTRY_AUTH_HTPASSWD_PATH":  "/auth/.htpasswd",
			"REGISTRY_AUTH_HTPASSWD_REALM": "registry",
		},
		WaitingFor: wait.NewHTTPStrategy("/").
			WithStartupTimeout(2 * time.Second).
			WithPort("5000/tcp").
			WithTLS(true).
			WithAllowInsecure(true).
			WithStatusCodeMatcher(func(status int) bool {
				return status == 200
			}),
	}

	registryServer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := registryServer.Terminate(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	ip, err := registryServer.Host(ctx)
	if err != nil {
		t.Error(err)
	}
	port, err := registryServer.MappedPort(ctx, "5000/tcp")
	if err != nil {
		t.Error(err)
	}

	registryAddress := fmt.Sprintf("%s:%s/", ip, port.Port())

	callback(registryAddress)
}
