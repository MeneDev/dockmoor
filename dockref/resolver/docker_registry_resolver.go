package resolver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/cli/cli/manifest/types"
	registryclient "github.com/docker/cli/cli/registry/client"
	"github.com/docker/cli/opts"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	types2 "github.com/docker/docker/api/types"
	registry2 "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/registry"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
)

func DockerRegistryResolverNew() dockref.Resolver {
	return &dockerRegistryResolver{
		NewCli:   newCli,
		osGetenv: os.Getenv,
	}
}

var _ dockref.Resolver = (*dockerRegistryResolver)(nil)

type dockerRegistryResolver struct {
	NewCli func(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface

	osGetenv func(key string) string
}

type searchOptions struct {
	format  string
	term    string
	noTrunc bool
	limit   int
	filter  opts.FilterOpt

	// Deprecated
	stars     uint
	automated bool
}

func (repo *dockerRegistryResolver) Resolve(reference dockref.Reference) ([]dockref.Reference, error) {
	ctx := context.Background()

	ref, err := normalizeReference(reference.Original())
	if err != nil {
		return nil, err
	}

	dockerTLSVerify := repo.osGetenv("DOCKER_TLS_VERIFY") != ""
	dockerTLS := repo.osGetenv("DOCKER_TLS") != ""

	in := ioutil.NopCloser(bytes.NewBuffer(nil))
	out := bytes.NewBuffer(nil)
	errWriter := bytes.NewBuffer(nil)
	isTrusted := false
	cli := command.NewDockerCli(in, out, errWriter, isTrusted, nil)
	cliOpts := flags.NewClientOptions()

	tls := dockerTLS || dockerTLSVerify
	host, e := opts.ParseHost(tls, repo.osGetenv("DOCKER_HOST"))
	if e != nil {
		return nil, e
	}
	cliOpts.Common.TLS = tls
	cliOpts.Common.TLSVerify = dockerTLSVerify
	cliOpts.Common.Hosts = []string{host}

	if tls {
		flgs := pflag.NewFlagSet("testing", pflag.ContinueOnError)
		cliOpts.Common.InstallFlags(flgs)
	}

	err = cli.Initialize(cliOpts)
	if err != nil {
		return nil, err
	}

	tripper := cli.Client().HTTPClient().Transport

	if tr, ok := tripper.(*http.Transport); ok {
		tr.TLSClientConfig = tlsconfig.ClientDefault()
	}

	tripper = http.DefaultTransport

	errOut := bytes.NewBuffer(nil)
	configFile := cliconfig.LoadDefaultConfigFile(errOut)
	errStr := errOut.String()
	if errStr != "" {
		return nil, errors.New(errStr)
	}

	authConfig := configFile.AuthConfigs["https://index.docker.io/v1/"]
	print(authConfig.Auth)

	options := registry.ServiceOptions{}
	defaultService, err := registry.NewService(options)
	endpoints, err := defaultService.LookupPullEndpoints("index.docker.io")

	roundTripper, err := getHTTPTransport(authConfig, endpoints[0], "ngix", UserAgent())
	println(endpoints)

	repository, err := client.NewRepository(ref, "https://registry-1.docker.io/", roundTripper)
	tags := repository.Tags(ctx)
	descriptor, err := tags.Get(ctx, "latest")
	dig := descriptor.Digest.Encoded()
	print(dig)

	resolver := func(ctx context.Context, index *registry2.IndexInfo) types2.AuthConfig {
		return command.ResolveAuthConfig(ctx, cli, index)
	}

	registryClient := registryclient.NewRegistryClient(resolver, UserAgent(), false)

	if e != nil {
		return nil, e
	}

	named, e := normalizeReference("nginx:latest")

	imageManifest, e := registryClient.GetManifest(ctx, named)

	manifestList, e := registryClient.GetManifestList(ctx, named)
	if e != nil {
		return nil, e
	}

	print(tags)
	print(manifestList)
	print(imageManifest.Descriptor.Annotations)

	targetRepo, err := registry.ParseRepositoryInfo(named)
	if err != nil {
		return nil, err
	}

	manifests := []manifestlist.ManifestDescriptor{}
	// More than one response. This is a manifest list.
	for _, img := range manifestList {
		mfd, err := buildManifestDescriptor(targetRepo, img)
		if err != nil {
			return nil, errors.Wrap(err, "failed to assemble ManifestDescriptor")
		}
		manifests = append(manifests, mfd)
	}
	deserializedML, err := manifestlist.FromDescriptors(manifests)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := deserializedML.MarshalJSON()

	str := string(jsonBytes)
	println(str)

	return nil, nil
}

// getHTTPTransport builds a transport for use in communicating with a registry
func getHTTPTransport(authConfig types2.AuthConfig, endpoint registry.APIEndpoint, repoName string, userAgent string) (http.RoundTripper, error) {
	// get the http transport, this will be used in a client to upload manifest
	base := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).Dial,
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

func buildManifestDescriptor(targetRepo *registry.RepositoryInfo, imageManifest types.ImageManifest) (manifestlist.ManifestDescriptor, error) {
	repoInfo, err := registry.ParseRepositoryInfo(imageManifest.Ref)
	if err != nil {
		return manifestlist.ManifestDescriptor{}, err
	}

	manifestRepoHostname := reference.Domain(repoInfo.Name)
	targetRepoHostname := reference.Domain(targetRepo.Name)
	if manifestRepoHostname != targetRepoHostname {
		return manifestlist.ManifestDescriptor{}, errors.Errorf("cannot use source images from a different registry than the target image: %s != %s", manifestRepoHostname, targetRepoHostname)
	}

	manifest := manifestlist.ManifestDescriptor{
		Descriptor: distribution.Descriptor{
			Digest:    imageManifest.Descriptor.Digest,
			Size:      imageManifest.Descriptor.Size,
			MediaType: imageManifest.Descriptor.MediaType,
		},
	}

	platform := types.PlatformSpecFromOCI(imageManifest.Descriptor.Platform)
	if platform != nil {
		manifest.Platform = *platform
	}

	if err = manifest.Descriptor.Digest.Validate(); err != nil {
		return manifestlist.ManifestDescriptor{}, errors.Wrapf(err,
			"digest parse of image %q failed", imageManifest.Ref)
	}

	return manifest, nil
}
func normalizeReference(ref string) (reference.Named, error) {
	namedRef, err := reference.ParseNormalizedNamed(ref)
	if err != nil {
		return nil, err
	}
	if _, isDigested := namedRef.(reference.Canonical); !isDigested {
		return reference.TagNameOnly(namedRef), nil
	}
	return namedRef, nil
}

func (repo *dockerRegistryResolver) newClient() (dockerAPIClient, error) {

	dockerTLSVerify := repo.osGetenv("DOCKER_TLS_VERIFY") != ""
	dockerTLS := repo.osGetenv("DOCKER_TLS") != ""

	in := ioutil.NopCloser(bytes.NewBuffer(nil))
	out := bytes.NewBuffer(nil)
	errWriter := bytes.NewBuffer(nil)
	isTrusted := false
	cli := repo.NewCli(in, out, errWriter, isTrusted)
	cliOpts := flags.NewClientOptions()

	tls := dockerTLS || dockerTLSVerify
	host, e := opts.ParseHost(tls, repo.osGetenv("DOCKER_HOST"))
	if e != nil {
		return nil, e
	}
	cliOpts.Common.TLS = tls
	cliOpts.Common.TLSVerify = dockerTLSVerify
	cliOpts.Common.Hosts = []string{host}

	if tls {
		flgs := pflag.NewFlagSet("testing", pflag.ContinueOnError)
		cliOpts.Common.InstallFlags(flgs)
	}

	err := cli.Initialize(cliOpts)
	if err != nil {
		return nil, err
	}
	client := cli.Client()
	return client, nil
}

// UserAgent returns the user agent string used for making API requests
func UserAgent() string {
	return "dockmoor/" + cli.Version + " (" + runtime.GOOS + ")"
}
