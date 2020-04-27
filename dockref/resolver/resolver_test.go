package resolver

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/MeneDev/dockmoor/dockref"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func runForDockerHub(t *testing.T, runAll func(t *testing.T, regAddr string, resolver dockref.Resolver)) {
	resolver := DockerRegistryResolverNew().(*dockerRegistryResolver)

	runAll(t, "", resolver)
}

func runForDockerRegistryWithoutAuth(t *testing.T, runAll func(t *testing.T, regAddr string, resolver dockref.Resolver)) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "registry_test",
		ExposedPorts: []string{"5000/tcp"},
		//Env: map[string]string{
		//	"REGISTRY_AUTH_HTPASSWD_PATH":  "/auth/.htpasswd",
		//	"REGISTRY_AUTH_HTPASSWD_REALM": "registry",
		//},
		WaitingFor: wait.NewHTTPStrategy("/").
			WithStartupTimeout(60 * time.Second).
			WithPort("5000/tcp").
			WithTLS(true).
			WithAllowInsecure(true).
			WithStatusCodeMatcher(func(status int) bool {
				return status == 200
			}),
		SkipReaper: true,
	}

	registryServer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		terminationErr := registryServer.Terminate(ctx)
		if terminationErr != nil {
			log.Fatal(terminationErr)
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

	resolver := DockerRegistryResolverNew().(*dockerRegistryResolver)

	runAll(t, registryAddress, resolver)
}

func runForDockerDeamon(t *testing.T, runAll func(t *testing.T, regAddr string, resolver dockref.Resolver)) {
	resolver := DockerDaemonResolverNew()
	runAll(t, "", resolver)
}

func TestResolvers_FindAllTags(t *testing.T) {
	type TestCaseData struct {
		ref  string
		tags []string
	}

	runWith := func(tcd TestCaseData, resolver dockref.Resolver) func(t *testing.T) {
		return func(t *testing.T) {
			result, e := resolver.FindAllTags(dockref.MustParse(tcd.ref))
			assert.Nil(t, e)
			assert.NotNil(t, result)
			if result != nil {
				foundTags := make([]string, 0)
				for _, ref := range result {
					foundTags = append(foundTags, ref.Tag())
				}
				for _, tag := range foundTags {
					assert.Contains(t, tcd.tags, tag)
				}
				for _, tag := range tcd.tags {
					assert.Contains(t, foundTags, tag)
				}
				assert.Equal(t, len(tcd.tags), len(foundTags), "Results and expected tag list have different length.")
			}
		}
	}

	runAllTestCasesForResolver := func(t *testing.T, regAddr string, resolver dockref.Resolver) {
		tags := []string{"1", "1.1.1", "1.1", "1.1.0", "1.0", "1.0.0", "1.0.1", "2", "2.0", "2.0.0", "edge", "mainline", "latest"}

		if strings.Contains(t.Name(), "registry") || strings.Contains(t.Name(), "hub") {
			tags = append(tags, "registry-only")
		}

		t.Run("unchanged menedev/testimagea", runWith(TestCaseData{
			regAddr + "menedev/testimagea",
			tags,
		}, resolver))
	}

	t.Run("Docker registry without auth", func(t *testing.T) {
		runForDockerRegistryWithoutAuth(t, runAllTestCasesForResolver)
	})

	t.Run("Docker hub", func(t *testing.T) {
		runForDockerHub(t, runAllTestCasesForResolver)
	})

	t.Run("Docker daemon", func(t *testing.T) {
		runForDockerDeamon(t, runAllTestCasesForResolver)
	})
}

func TestResolvers_Resolve(t *testing.T) {
	type TestCaseData struct {
		ref  string
		tag  string
		dig  string
		mode dockref.ResolveMode
	}

	runWith := func(tcd TestCaseData, resolver dockref.Resolver) func(t *testing.T) {
		return func(t *testing.T) {
			result, e := resolver.Resolve(dockref.MustParse(tcd.ref))
			assert.Nil(t, e)
			assert.NotNil(t, result)
			if result != nil {
				assert.Equal(t, tcd.tag, result.Tag())
				assert.Equal(t, tcd.dig, result.DigestString())
			}
		}
	}

	runAllTestCasesForResolver := func(t *testing.T, regAddr string, resolver dockref.Resolver) {
		t.Run("unchanged menedev/testimagea:2.0.0", runWith(TestCaseData{
			regAddr + "menedev/testimagea:2.0.0",
			"2.0.0",
			"sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1",
			dockref.ResolveModeUnchanged}, resolver))

		t.Run("unchanged_menedev/testimagea:2.0", runWith(TestCaseData{
			regAddr + "menedev/testimagea:2.0",
			"2.0",
			"sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1",
			dockref.ResolveModeUnchanged}, resolver))

		t.Run("unchanged menedev/testimagea:2", runWith(TestCaseData{
			regAddr + "menedev/testimagea:2",
			"2",
			"sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1",
			dockref.ResolveModeUnchanged}, resolver))

		t.Run("unchanged menedev/testimagea", runWith(TestCaseData{
			regAddr + "menedev/testimagea",
			"",
			"sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1",
			dockref.ResolveModeUnchanged}, resolver))

		t.Run("unchanged menedev/testimagea:mainline", runWith(TestCaseData{
			regAddr + "menedev/testimagea:mainline",
			"mainline", "sha256:1e2b1cc7d366650a93620ca3cc8691338ed600ababf90a0e5803e1ee32486624",
			dockref.ResolveModeUnchanged}, resolver))

		t.Run("unchanged menedev/testimagea:edge", runWith(TestCaseData{
			regAddr + "menedev/testimagea:edge",
			"edge",
			"sha256:3d4d88675636f0fdf7899e3d3c6f8d5a9cae768e8b7f38f05505d6a88497e7a1",
			dockref.ResolveModeUnchanged}, resolver))
	}

	t.Run("Docker registry without auth", func(t *testing.T) {
		runForDockerRegistryWithoutAuth(t, runAllTestCasesForResolver)
	})

	t.Run("Docker hub", func(t *testing.T) {
		runForDockerHub(t, runAllTestCasesForResolver)
	})

	t.Run("Docker daemon", func(t *testing.T) {
		runForDockerDeamon(t, runAllTestCasesForResolver)
	})
}
