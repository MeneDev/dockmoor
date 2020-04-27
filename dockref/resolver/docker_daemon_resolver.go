package resolver

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/MeneDev/dockmoor/dockref"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

var _ dockref.Resolver = (*dockerDaemonResolver)(nil)

type dockerDaemonResolver struct {
	ImageInspect func(reference dockref.Reference) (types.ImageInspect, error)
	NewCli       func(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface

	osGetenv func(key string) string
}

func DockerDaemonResolverNew() dockref.Resolver {
	repo := &dockerDaemonResolver{
		NewCli: newCli,

		osGetenv: os.Getenv,
	}
	return repo
}

func (reser dockerDaemonResolver) imageInspect(reference dockref.Reference) (types.ImageInspect, error) {
	ctx := context.Background()

	client, err := reser.newClient()
	if err != nil {
		return types.ImageInspect{}, err
	}

	imageInspect, _, err := client.ImageInspectWithRaw(ctx, reference.Original())

	return imageInspect, err
}
func (reser dockerDaemonResolver) imageList(reference dockref.Reference) ([]types.ImageSummary, error) {
	ctx := context.Background()

	client, err := reser.newClient()
	if err != nil {
		return nil, err
	}

	summaries, err := client.ImageList(ctx, reference.Original())

	return summaries, err
}

func (reser dockerDaemonResolver) newClient() (dockerAPIClient, error) {
	dockerTLSVerify := reser.osGetenv("DOCKER_TLS_VERIFY") != ""
	dockerTLS := reser.osGetenv("DOCKER_TLS") != ""

	in := ioutil.NopCloser(bytes.NewBuffer(nil))
	out := bytes.NewBuffer(nil)
	errWriter := bytes.NewBuffer(nil)
	isTrusted := false
	cli := reser.NewCli(in, out, errWriter, isTrusted)
	cliOpts := flags.NewClientOptions()

	tls := dockerTLS || dockerTLSVerify
	host, e := opts.ParseHost(tls, reser.osGetenv("DOCKER_HOST"))
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

type dockerAPIClient interface {
	ImageInspectWithRaw(ctx context.Context, reference string) (types.ImageInspect, []byte, error)
	ImageList(ctx context.Context, reference string) ([]types.ImageSummary, error)
}

type dockerCliInterface interface {
	Initialize(options *flags.ClientOptions) error
	Client() dockerAPIClient
}

type dockerCli struct {
	cli *command.DockerCli
}

func (d dockerCli) Initialize(options *flags.ClientOptions) error {
	return d.cli.Initialize(options)
}

var _ dockerAPIClient = (*dockerCliAdapter)(nil)

type dockerCliAdapter struct {
	client client.APIClient
}

func (d *dockerCliAdapter) ImageInspectWithRaw(ctx context.Context, reference string) (types.ImageInspect, []byte, error) {
	return d.client.ImageInspectWithRaw(ctx, reference)
}

func (d *dockerCliAdapter) ImageList(ctx context.Context, reference string) ([]types.ImageSummary, error) {
	filterOpts := opts.NewFilterOpt()
	filters := filterOpts.Value()

	filters.Add("reference", reference)

	listOptions := types.ImageListOptions{
		All:     true,
		Filters: filters,
	}

	return d.client.ImageList(ctx, listOptions)
}

func dockerCliAdapterNew(client client.APIClient) dockerAPIClient {
	adapter := &dockerCliAdapter{
		client: client,
	}

	return adapter
}

func (d dockerCli) Client() dockerAPIClient {
	cli := d.cli.Client()
	cli.NegotiateAPIVersion(context.Background())
	return dockerCliAdapterNew(cli)
}

func newCli(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface {
	cli, _ := command.NewDockerCli(command.WithInputStream(in),
		command.WithOutputStream(out),
		command.WithErrorStream(errWriter),
		command.WithContentTrust(isTrusted))

	return &dockerCli{cli}
}

func (reser dockerDaemonResolver) FindAllTags(reference dockref.Reference) ([]dockref.Reference, error) {
	summaries, err := reser.imageList(reference)

	if err != nil {
		return nil, err
	}

	refs := make([]dockref.Reference, 0)
	for _, summary := range summaries {
		digs := summary.RepoDigests
		tags := summary.RepoTags

		for _, tag := range tags {
			tagRef := dockref.MustParse(tag)
			r := reference.WithTag(tagRef.Tag())
			for _, dig := range digs {
				digRef := dockref.MustParse(dig)
				r = r.WithDigest(digRef.DigestString())
				refs = append(refs, r)
			}

			if len(digs) == 0 {
				refs = append(refs, r)
			}
		}

		if len(digs) == 0 && len(tags) == 0 {
			r := dockref.MustParseAlgoDigest(summary.ID)
			refs = append(refs, r)
		}
	}
	return refs, nil
}

//func (reser dockerDaemonResolver) FindMatchingTags(reference dockref.Reference) ([]dockref.Reference, error) {
//	// TODO upcoming release
//	imageInspect, err := reser.imageInspect(reference)
//
//	if err != nil {
//		return nil, err
//	}
//
//	digs := imageInspect.RepoDigests
//	tags := imageInspect.RepoTags
//
//	refs := make([]dockref.Reference, 0)
//	// TODO why can there more than one digest?
//	for _, tag := range tags {
//		tagRef := dockref.MustParse(tag)
//		r := reference.WithTag(tagRef.Tag())
//		for _, dig := range digs {
//			digRef := dockref.MustParse(dig)
//			r = r.WithDigest(digRef.DigestString())
//			refs = append(refs, r)
//		}
//
//		if len(digs) == 0 {
//			refs = append(refs, r)
//		}
//	}
//
//	if len(digs) == 0 && len(tags) == 0 {
//		r := dockref.MustParseAlgoDigest(imageInspect.ID)
//		refs = append(refs, r)
//	}
//
//	return refs, nil
//}

func (reser dockerDaemonResolver) Resolve(reference dockref.Reference) (dockref.Reference, error) {
	imageInspect, err := reser.imageInspect(reference)

	if err != nil {
		return nil, err
	}

	digs := imageInspect.RepoDigests
	//tags := imageInspect.RepoTags

	// NOTE RepoDigests include the repo name
	// e.g.
	// menedev/testimagea@sha256:1e2b1cc7d366650a93620ca3cc8691338ed600ababf90a0e5803e1ee32486624
	// localhost:5000/menedev/testimagea@sha256:1e2b1cc7d366650a93620ca3cc8691338ed600ababf90a0e5803e1ee32486624

	digFilters := []func(dig string) bool{
		func(dig string) bool {
			return true
		},

		func(dig string) bool {
			digRef := dockref.MustParse(dig)

			return digRef.Domain() == reference.Domain()
		},
	}

	var dig dockref.Reference
	for _, filter := range digFilters {
		digs = filterStrings(digs, filter)
		if len(digs) == 1 {
			dig = dockref.MustParse(digs[0])
			break
		}
	}

	if len(digs) == 0 {
		return nil, errors.Errorf(
			"no RepoDigests for reference %s. The Docker Daemon Resolver can only resolve pulled images.",
			reference.Original())
	}

	if len(digs) > 1 {
		return nil, errors.Errorf(
			"ambigious RepoDigests [%s] for reference %s",
			strings.Join(imageInspect.RepoDigests, ", "),
			reference.Original())
	}
	reference = reference.WithDigest(dig.DigestString())

	return reference, nil
}

func filterStrings(strings []string, predicate func(s string) bool) []string {
	filteredStrings := make([]string, 0)
	for _, s := range strings {
		if predicate(s) {
			filteredStrings = append(filteredStrings, s)
		}
	}

	return filteredStrings
}
