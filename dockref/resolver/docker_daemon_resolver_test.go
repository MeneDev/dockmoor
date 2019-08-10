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
	resolver := DockerDaemonResolverNew()
	daemonResolver := resolver.(*dockerDaemonResolver)
	return daemonResolver
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

func (m *mockDockerAPIClient) ImageList(ctx context.Context, reference string) ([]types.ImageSummary, error) {
	called := m.Called(ctx, reference)
	iss := called.Get(0)
	var imageSummaries []types.ImageSummary
	if iss != nil {
		imageSummaries = iss.([]types.ImageSummary)
	}

	err := called.Error(1)

	return imageSummaries, err
}

func (m *mockDockerAPIClient) ImageInspectWithRaw(ctx context.Context, reference string) (types.ImageInspect, []byte, error) {
	called := m.Called(ctx, reference)
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

func TestDockerDaemonRegistry_FindAllTags(t *testing.T) {
	// in resolver_test.go
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
		resolver := dockerDaemonResolverNewTest()
		resolver.osGetenv = func(key string) string {
			switch key {
			case "DOCKER_HOST":
				return "the host is not valid!"
			}
			return ""
		}

		apiClient, e := resolver.newClient()
		assert.Error(t, e)
		assert.Nil(t, apiClient)
	})

	t.Run("unresolvable host returns error", func(t *testing.T) {
		resolver := dockerDaemonResolverNewTest()
		resolver.osGetenv = func(key string) string {
			switch key {
			case "DOCKER_TLS":
				return "1"
			}
			return ""
		}

		client := &mockDockerCliInterface{}
		resolver.NewCli = func(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface {
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
		_, e := resolver.newClient()
		assert.Nil(t, e)
	})

}

func TestDockerDaemonRegistry_Resolve_Error_in_Initialize(t *testing.T) {
	reser := dockerDaemonResolverNewTest()
	mockCli := &mockDockerCli{}
	reser.NewCli = func(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface {
		return mockCli
	}

	mockClient := &mockDockerAPIClient{}

	expected := errors.New("initialize error")
	mockCli.On("Initialize", mock.Anything).Return(expected)
	mockCli.On("Client").Return(mockClient)

	references, e := reser.FindAllTags(dockref.MustParse("nginx"))
	assert.Error(t, e)
	assert.Equal(t, expected, e)
	assert.Empty(t, references)
}
