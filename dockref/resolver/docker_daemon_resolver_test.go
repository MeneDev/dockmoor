package resolver

import (
	"bytes"
	"context"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
)

func dockerDaemonResolverNewTest() *dockerDaemonResolver {
	reser := DockerDaemonResolverNew()
	resolver := reser.(*dockerDaemonResolver)
	return resolver
}

var _ dockerCliInterface = (*mockDockerCli)(nil)

type mockDockerCli struct {
	mock.Mock
}

func (m *mockDockerCli) Initialize(options *flags.ClientOptions) error {
	called := m.Called(options)
	return called.Error(0)
}

func (m *mockDockerCli) Client() dockerAPIClient {
	called := m.Called()
	obj := called.Get(0)
	apiClient := obj.(dockerAPIClient)
	return apiClient
}

var _ dockerAPIClient = (*mockDockerAPIClient)(nil)

type mockDockerAPIClient struct {
	mock.Mock
}

func (m *mockDockerAPIClient) ImageInspectWithRaw(ctx context.Context, image string) (types.ImageInspect, []byte, error) {
	called := m.Called(ctx, image)
	ii := called.Get(0)
	imageInspect := ii.(types.ImageInspect)

	bytesObj := called.Get(1)
	var bytes []byte
	if bytesObj != nil {
		bytes = bytesObj.([]byte)
	}

	err := called.Error(2)

	return imageInspect, bytes, err
}

func TestDockerDaemonRegistry_Resolve(t *testing.T) {
	reser := dockerDaemonResolverNewTest()
	mockCli := &mockDockerCli{}
	reser.NewCli = func(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface {
		return mockCli
	}

	mockClient := &mockDockerAPIClient{}

	mockCli.On("Initialize", mock.Anything).Return(nil)
	mockCli.On("Client").Return(mockClient)

	mockClient.On("ImageInspectWithRaw", mock.Anything, "nginx:1.15.5").
		Return(types.ImageInspect{
			RepoDigests: []string{"nginx@sha256:b73f527d86e3461fd652f62cf47e7b375196063bbbd503e853af5be16597cb2e"},
			RepoTags:    []string{"nginx:1.15.5"},
		}, nil, nil)

	mockClient.On("ImageInspectWithRaw", mock.Anything, "nginx@sha256:b73f527d86e3461fd652f62cf47e7b375196063bbbd503e853af5be16597cb2e").
		Return(types.ImageInspect{
			RepoDigests: []string{"nginx@sha256:b73f527d86e3461fd652f62cf47e7b375196063bbbd503e853af5be16597cb2e"},
			RepoTags:    []string{"nginx:1.15.5"},
		}, nil, nil)

	mockClient.On("ImageInspectWithRaw", mock.Anything, "nginx:1.15.6").
		Return(types.ImageInspect{
			RepoDigests: []string{"nginx@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"},
			RepoTags:    []string{"nginx:1.15.6", "nginx:latest"},
		}, nil, nil)

	mockClient.On("ImageInspectWithRaw", mock.Anything, "nginx:latest").
		Return(types.ImageInspect{
			RepoDigests: []string{"nginx@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"},
			RepoTags:    []string{"nginx:1.15.6", "nginx:latest"},
		}, nil, nil)

	mockClient.On("ImageInspectWithRaw", mock.Anything, "nginx@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991").
		Return(types.ImageInspect{
			RepoDigests: []string{"nginx@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"},
			RepoTags:    []string{"nginx:1.15.6", "nginx:latest"},
		}, nil, nil)

	mockClient.On("ImageInspectWithRaw", mock.Anything, "unknown:unknown").
		Return(types.ImageInspect{}, nil, errors.New("Error"))

	t.Run("invalid", func(t *testing.T) {
		references, e := reser.FindAllTags(dockref.MustParse("unknown:unknown"))
		assert.Error(t, e)
		assert.Nil(t, references)
	})

	type T struct {
		name, digest string
	}

	tests := []T{
		{name: "nginx:1.15.5", digest: "nginx@sha256:b73f527d86e3461fd652f62cf47e7b375196063bbbd503e853af5be16597cb2e"},
		{name: "nginx:1.15.6", digest: "nginx@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"},
		{name: "nginx:latest", digest: "nginx@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"},
		//{name: "nginx:1.15.5-perl", digest: "nginx@sha256:01c45fbd335b5fcfbfe95777508cc16044e0d6a929f5d531f48ab53ca4556578"},
		//{name: "nginx:1.15.5-alpine", digest: "nginx@sha256:ae5da813f8ad7fa785d7668f0b018ecc8c3a87331527a61d83b3b5e816a0f03c"},
		//{name: "nginx:1.15.5-alpine-perl", digest: "nginx@sha256:9c632b0423d3ceba7e94a6744a127b694caacb6117238aff033ab6bdc88c1fae"},
		//{name: "nginx:1.14.0", digest: "nginx@sha256:8b600a4d029481cc5b459f1380b30ff6cb98e27544fc02370de836e397e34030"},
		//{name: "nginx:1.14.0-perl", digest: "nginx@sha256:032acb6025fa581888812e79f4efcd32d008e0ce3dfe56c65f9c1011d93ce920"},
		//{name: "nginx:1.14.0-alpine", digest: "nginx@sha256:8976218be775f4244df2a60a169d44606b6978bac4375192074cefc0c7824ddf"},
		//{name: "nginx:1.14.0-alpine-perl", digest: "nginx@sha256:c3d6f9a179ba365ab4b41e176623a6fc9cfc2121567131127e43f5660e0c4767"},
	}

	for _, tst := range tests {
		t.Run("Resolves name "+tst.name, func(t *testing.T) {
			ref, e := dockref.Parse(tst.name)
			assert.Nil(t, e)

			resolve, e := reser.FindAllTags(ref)
			assert.Nil(t, e)

			assert.NotNil(t, resolve)
			if resolve != nil {
				reference, err := resolve[0].WithRequestedFormat(dockref.FormatHasName | dockref.FormatHasDigest)
				assert.Nil(t, err)
				assert.Equal(t, tst.digest, reference.Formatted())
			}
		})
	}

	for _, tst := range tests {
		dig := dockref.MustParse(tst.digest).Formatted()
		//dig = strings.SplitAfter(dig, ":")[1]
		t.Run("Resolves digest "+dig, func(t *testing.T) {
			ref := dockref.MustParse(dig)

			resolve, e := reser.FindAllTags(ref)
			assert.Nil(t, e)

			assert.NotNil(t, resolve)

			formatted := make([]string, 0)

			for _, res := range resolve {
				reference, err := res.WithRequestedFormat(dockref.FormatHasName | dockref.FormatHasTag)
				assert.Nil(t, err)
				f := reference.Formatted()
				formatted = append(formatted, f)
			}

			assert.Contains(t, formatted, tst.name)
		})
	}
}

type mockDockerCliInterface struct {
	mock.Mock
}

func (m *mockDockerCliInterface) Initialize(options *flags.ClientOptions) error {
	called := m.Called(options)
	e := called.Error(0)
	return e
}

func (m *mockDockerCliInterface) Client() dockerAPIClient {
	called := m.Called()
	client := called.Get(0)
	if client != nil {
		return client.(dockerAPIClient)
	}
	return nil
}

func TestDockerDaemonRegistry_newClient(t *testing.T) {
	t.Run("unresolvable host returns error", func(t *testing.T) {
		reser := dockerDaemonResolverNewTest()
		reser.osGetenv = func(key string) string {
			switch key {
			case "DOCKER_HOST":
				return "the host is not valid!"
			}
			return ""
		}

		apiClient, e := reser.newClient()
		assert.Error(t, e)
		assert.Nil(t, apiClient)
	})

	t.Run("unresolvable host returns error", func(t *testing.T) {
		reser := dockerDaemonResolverNewTest()
		reser.osGetenv = func(key string) string {
			switch key {
			case "DOCKER_TLS":
				return "1"
			}
			return ""
		}

		client := &mockDockerCliInterface{}
		reser.NewCli = func(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface {
			return client
		}

		client.On("Initialize", mock.Anything).Run(func(args mock.Arguments) {
			get := args.Get(0)
			assert.NotNil(t, get)
			if get != nil {
				options := get.(*flags.ClientOptions)
				tlsOpts := options.Common.TLSOptions

				assert.NotNil(t, tlsOpts)
			}
		}).Return(nil)
		client.On("Client").Return(nil)
		_, e := reser.newClient()
		assert.Nil(t, e)
	})

}

func TestDockerDaemonRegistry_Resolve_IT(t *testing.T) {
	// integration tests
	// require pulled nginx images
	reser := DockerDaemonResolverNew()

	t.Run("invalid", func(t *testing.T) {
		references, e := reser.FindAllTags(dockref.MustParse("unknown:unknown"))
		assert.Error(t, e)
		assert.Nil(t, references)
	})

	type T struct {
		name, digest string
	}

	tests := []T{
		{name: "nginx:1.15.5", digest: "nginx@sha256:b73f527d86e3461fd652f62cf47e7b375196063bbbd503e853af5be16597cb2e"},
		{name: "nginx:1.15.6", digest: "nginx@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"},
		{name: "nginx:latest", digest: "nginx@sha256:31b8e90a349d1fce7621f5a5a08e4fc519b634f7d3feb09d53fac9b12aa4d991"},
	}

	for _, tst := range tests {
		t.Run("Resolves name "+tst.name, func(t *testing.T) {
			ref, e := dockref.Parse(tst.name)
			assert.Nil(t, e)

			resolve, e := reser.FindAllTags(ref)
			assert.Nil(t, e)

			assert.NotNil(t, resolve)
			if resolve != nil {
				reference, err := resolve[0].WithRequestedFormat(dockref.FormatHasName | dockref.FormatHasDigest)
				assert.Nil(t, err)
				assert.Equal(t, tst.digest, reference.Formatted())
			}
		})
	}

	for _, tst := range tests {
		dig := dockref.MustParse(tst.digest).Formatted()
		t.Run("Resolves digest "+dig, func(t *testing.T) {
			ref := dockref.MustParse(dig)

			resolve, e := reser.FindAllTags(ref)
			assert.Nil(t, e)

			assert.NotNil(t, resolve)

			formatted := make([]string, 0)

			for _, res := range resolve {
				reference, err := res.WithRequestedFormat(dockref.FormatHasName | dockref.FormatHasTag)
				assert.Nil(t, err)
				f := reference.Formatted()
				formatted = append(formatted, f)
			}

			assert.Contains(t, formatted, tst.name)
		})
	}
}

func TestDockerDaemonRegistry_Resolve_Error_in_Initialize(t *testing.T) {
	reser := dockerDaemonResolverNewTest()
	mockCli := &mockDockerCli{}
	reser.NewCli = func(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface {
		return mockCli
	}

	mockClient := &mockDockerAPIClient{}

	expected := errors.New("Initialize error")
	mockCli.On("Initialize", mock.Anything).Return(expected)
	mockCli.On("Client").Return(mockClient)

	references, e := reser.FindAllTags(dockref.MustParse("nginx"))
	assert.Error(t, e)
	assert.Equal(t, expected, e)
	assert.Empty(t, references)
}

func TestDockerDaemonResolver_Resolve_DigestOnly(t *testing.T) {
	reser := dockerDaemonResolverNewTest()
	mockCli := &mockDockerCli{}
	reser.NewCli = func(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface {
		return mockCli
	}

	mockClient := &mockDockerAPIClient{}

	mockCli.On("Initialize", mock.Anything).Return(nil)
	mockCli.On("Client").Return(mockClient)

	mockClient.On("ImageInspectWithRaw", mock.Anything, "3247732819d6cd7af0c45a05b30d0b147f05a25ee2e83d7b9707ee25fcdd0f58").
		Return(types.ImageInspect{
			ID: "sha256:3247732819d6cd7af0c45a05b30d0b147f05a25ee2e83d7b9707ee25fcdd0f58",
		}, nil, nil)

	references, e := reser.FindAllTags(dockref.MustParse("3247732819d6cd7af0c45a05b30d0b147f05a25ee2e83d7b9707ee25fcdd0f58"))
	assert.Nil(t, e)
	assert.NotEmpty(t, references)

	if len(references) > 0 {
		reference := references[0]
		assert.Equal(t, reference.Formatted(), "3247732819d6cd7af0c45a05b30d0b147f05a25ee2e83d7b9707ee25fcdd0f58")
	}
}

func TestDockerDaemonResolver_Resolve_LocalOnly_but_tagged(t *testing.T) {
	reser := dockerDaemonResolverNewTest()
	mockCli := &mockDockerCli{}
	reser.NewCli = func(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface {
		return mockCli
	}

	mockClient := &mockDockerAPIClient{}

	mockCli.On("Initialize", mock.Anything).Return(nil)
	mockCli.On("Client").Return(mockClient)

	mockClient.On("ImageInspectWithRaw", mock.Anything, "3247732819d6cd7af0c45a05b30d0b147f05a25ee2e83d7b9707ee25fcdd0f58").
		Return(types.ImageInspect{
			ID:       "sha256:3247732819d6cd7af0c45a05b30d0b147f05a25ee2e83d7b9707ee25fcdd0f58",
			RepoTags: []string{"test:tagged"},
		}, nil, nil)

	references, e := reser.FindAllTags(dockref.MustParse("3247732819d6cd7af0c45a05b30d0b147f05a25ee2e83d7b9707ee25fcdd0f58"))
	assert.Nil(t, e)
	assert.NotEmpty(t, references)

	if len(references) > 0 {
		reference := references[0]
		assert.Equal(t, reference.Formatted(), "3247732819d6cd7af0c45a05b30d0b147f05a25ee2e83d7b9707ee25fcdd0f58")
	}
}
