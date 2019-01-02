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
	"github.com/docker/cli/opts"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	types2 "github.com/docker/docker/api/types"
	"github.com/docker/docker/registry"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

func DockerRegistryResolverNew(opts dockref.ResolverOptions) dockref.Resolver {
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

func (repo *dockerRegistryResolver) FindAllTags(rfrnce dockref.Reference) ([]dockref.Reference, error) {
	ctx := context.Background()

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

	err := cli.Initialize(cliOpts)
	if err != nil {
		return nil, err
	}

	errOut := bytes.NewBuffer(nil)
	configFile := cliconfig.LoadDefaultConfigFile(errOut)
	errStr := errOut.String()
	if errStr != "" {
		return nil, errors.New(errStr)
	}

	store := configFile.GetCredentialsStore(rfrnce.Domain())

	options := registry.ServiceOptions{}
	defaultService, err := registry.NewService(options)

	repoInfo, err := registry.ParseRepositoryInfo(rfrnce)

	endpoints, err := defaultService.LookupPullEndpoints(reference.Domain(repoInfo.Name))

	authConfig, err := store.Get(rfrnce.Domain())
	if err != nil {
		return nil, err
	}

	lrr := lookupReference{rfrnce}

	roundTripper, err := getHTTPTransport(authConfig, endpoints[0], lrr.Name(), UserAgent())
	println(endpoints)

	//fmtd, err := rfrnce.WithRequestedFormat(dockref.FormatHasDigest | dockref.FormatHasDigest | dockref.FormatHasDomain | dockref.FormatHasTag)
	repository, err := client.NewRepository(lrr, endpoints[0].URL.String(), roundTripper)
	tags := repository.Tags(ctx)
	strings, err := tags.All(ctx)
	strings, err = tags.All(ctx)
	if err != nil {
		return nil, err
	}

	refs := make([]dockref.Reference, 0)
	println(strings)
	for _, tag := range strings {
		r := rfrnce.WithTag(tag).WithDigest("")

		//descriptor, err := tags.Get(ctx, tag)
		//if err != nil {
		//	return nil, err
		//}

		//r = r.WithDigest(descriptor.Digest.String())
		refs = append(refs, r)
	}

	return refs, nil
}

func (repo *dockerRegistryResolver) Resolve(rfrnce dockref.Reference) (dockref.Reference, error) {
	ctx := context.Background()

	errOut := bytes.NewBuffer(nil)
	configFile := cliconfig.LoadDefaultConfigFile(errOut)
	errStr := errOut.String()
	if errStr != "" {
		return nil, errors.New(errStr)
	}

	store := configFile.GetCredentialsStore(rfrnce.Domain())

	options := registry.ServiceOptions{}
	defaultService, err := registry.NewService(options)

	repoInfo, err := registry.ParseRepositoryInfo(rfrnce)

	endpoints, err := defaultService.LookupPullEndpoints(reference.Domain(repoInfo.Name))

	authConfig, err := store.Get(rfrnce.Domain())
	if err != nil {
		return nil, err
	}

	lrr := lookupReference{rfrnce}

	roundTripper, err := getHTTPTransport(authConfig, endpoints[0], lrr.Name(), UserAgent())

	repository, err := client.NewRepository(lrr, endpoints[0].URL.String(), roundTripper)
	tagService := repository.Tags(ctx)

	tag := rfrnce.Tag()
	if tag == "" {
		tag = "latest"
		rfrnce = rfrnce.WithTag(tag)
	}

	descriptor, err := tagService.Get(ctx, tag)
	if err != nil {
		return nil, err
	}

	rfrnce = rfrnce.WithDigest(string(descriptor.Digest))

	rfrnce, err = findTag(ctx, rfrnce, tagService)
	if err != nil {
		return nil, err
	}

	return rfrnce, nil
}

func findTag(ctx context.Context, ref dockref.Reference, tagService distribution.TagService) (dockref.Reference, error) {
	if ref.Digest() == "" {
		return nil, errors.New("Expected digest")
	}

	tags, err := tagService.All(ctx)
	if err != nil {
		return nil, err
	}

	tagged := make([]dockref.Reference, 0)
	for _, tag := range tags {
		tagged = append(tagged, ref.WithTag(tag))
	}

	relevant, err := dockref.MatchingDomainNameAndVariant(ref, tagged, nil)
	if err != nil {
		return nil, err
	}

	relevant, err = dockref.TagVersionsGreaterOrEqualOrNotAVersion(ref, relevant, nil)
	if err != nil {
		return nil, err
	}

	relevant, err = dockref.TagVersionsEqualOrNotAVersion(ref, relevant, nil)
	if err != nil {
		return nil, err
	}

	sameDigestChan := make(chan dockref.Reference, len(relevant))
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(relevant))

	sameDigest := make([]dockref.Reference, 0)
	for _, tagged := range relevant {
		go func(tagged dockref.Reference) {
			defer func() {
				waitGroup.Done()
			}()

			descriptor, err := tagService.Get(ctx, tagged.Tag())
			if err != nil {
				return
			}
			if descriptor.Digest == ref.Digest() {
				sameDigestChan <- tagged
			}
		}(tagged)
	}

	waitGroup.Wait()
	close(sameDigestChan)

	for v := range sameDigestChan {
		sameDigest = append(sameDigest, v)
	}

	if len(sameDigest) == 0 {
		// TODO better error
		return nil, errors.New("not found")
	}

	mostPreciseTag, err := dockref.MostPreciseTag(sameDigest, nil)
	return mostPreciseTag, nil
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
