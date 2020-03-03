package resolver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/credentials"
	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/registry"
	"github.com/pkg/errors"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
)

func DockerRegistryResolverNew() dockref.Resolver {
	resolver := &dockerRegistryResolver{
		NewCli:   newCli,
		osGetenv: os.Getenv,
	}
	resolver.credentialsStoreFactory = resolver.defaultCredentialsStore
	return resolver
}

var _ dockref.Resolver = (*dockerRegistryResolver)(nil)

type dockerRegistryResolver struct {
	NewCli func(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface

	osGetenv func(key string) string

	credentialsStoreFactory func(ref dockref.Reference) (credentials.Store, error)
}

var _ reference.Named = (*lookupReference)(nil)

type lookupReference struct {
	r dockref.Reference
}

func (lr lookupReference) String() string {
	panic("implement me")
}

func (lr lookupReference) Name() string {
	return lr.r.Path()
}

func (repo *dockerRegistryResolver) FindAllTags(ref dockref.Reference) ([]dockref.Reference, error) {
	ctx := context.Background()
	tagService, err := repo.tagService(ctx, ref)
	if err != nil {
		return nil, err
	}

	strings, err := tagService.All(ctx)
	if err != nil {
		return nil, err
	}

	refs := make([]dockref.Reference, 0)
	println(strings)
	for _, tag := range strings {
		r := ref.WithTag(tag).WithDigest("")

		refs = append(refs, r)
	}

	return refs, nil
}

func (repo *dockerRegistryResolver) Resolve(ref dockref.Reference) (dockref.Reference, error) {
	ctx := context.Background()
	tagService, err := repo.tagService(ctx, ref)
	if err != nil {
		return nil, err
	}

	tag := ref.Tag()
	if tag == "" {
		tag = "latest"
	}

	descriptor, err := tagService.Get(ctx, tag)
	if err != nil {
		return nil, err
	}

	ref = ref.WithDigest(string(descriptor.Digest))

	return ref, nil
}

func (repo *dockerRegistryResolver) defaultCredentialsStore(ref dockref.Reference) (credentials.Store, error) {
	errOut := bytes.NewBuffer(nil)
	configFile := config.LoadDefaultConfigFile(errOut)
	errStr := errOut.String()
	if errStr != "" {
		return nil, errors.New(errStr)
	}
	store := configFile.GetCredentialsStore(ref.Domain())
	return store, nil
}

func (repo *dockerRegistryResolver) tagService(ctx context.Context, ref dockref.Reference) (distribution.TagService, error) {
	store, err := repo.credentialsStoreFactory(ref)
	if err != nil {
		return nil, err
	}

	options := registry.ServiceOptions{}
	defaultService, err := registry.NewService(options)
	if err != nil {
		return nil, err
	}
	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return nil, err
	}

	endpoints, err := defaultService.LookupPullEndpoints(reference.Domain(repoInfo.Name))
	if err != nil {
		return nil, err
	}

	authConfig, err := store.Get(ref.Domain())
	if err != nil {
		return nil, err
	}
	lrr := lookupReference{ref}
	authConfig2 := types.AuthConfig{authConfig.Username, authConfig.Password, authConfig.Auth, authConfig.Email, authConfig.ServerAddress, authConfig.IdentityToken, authConfig.RegistryToken}
	roundTripper, err := getHTTPTransport(authConfig2, endpoints[0], lrr.Name(), UserAgent())
	if err != nil {
		return nil, err
	}

	repository, err := client.NewRepository(lrr, endpoints[0].URL.String(), roundTripper)
	if err != nil {
		return nil, err
	}

	tagService := repository.Tags(ctx)
	return tagService, nil
}

// getHTTPTransport builds a transport for use in communicating with a registry
func getHTTPTransport(authConfig types.AuthConfig, endpoint registry.APIEndpoint, repoName string, userAgent string) (http.RoundTripper, error) {
	// get the http transport, this will be used in a client to upload manifest
	base := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
			return (&net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext(ctx, network, addr)
		},
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     endpoint.TLSConfig,
		DisableKeepAlives:   true,
	}

	modifiers := registry.Headers(userAgent, http.Header{})
	authTransport := transport.NewTransport(base, modifiers...)
	challengeManager, confirmedV2, err := registry.PingV2Registry(endpoint.URL, authTransport)
	if err != nil {
		return nil, errors.Wrap(err, "error pinging v2 registry")
	}
	if !confirmedV2 {
		return nil, fmt.Errorf("unsupported registry version")
	}
	if authConfig.RegistryToken != "" {
		passThruTokenHandler := &existingTokenHandler{token: authConfig.RegistryToken}
		modifiers = append(modifiers, auth.NewAuthorizer(challengeManager, passThruTokenHandler))
	} else {
		creds := registry.NewStaticCredentialStore(&authConfig)
		tokenHandler := auth.NewTokenHandler(authTransport, creds, repoName, "push", "pull")
		basicHandler := auth.NewBasicHandler(creds)
		modifiers = append(modifiers, auth.NewAuthorizer(challengeManager, tokenHandler, basicHandler))
	}
	return transport.NewTransport(base, modifiers...), nil
}

type existingTokenHandler struct {
	token string
}

func (th *existingTokenHandler) AuthorizeRequest(req *http.Request, params map[string]string) error {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", th.token))
	return nil
}

func (th *existingTokenHandler) Scheme() string {
	return "bearer"
}

// UserAgent returns the user agent string used for making API requests
func UserAgent() string {
	return "dockmoor/0 (" + runtime.GOOS + ")"
}
