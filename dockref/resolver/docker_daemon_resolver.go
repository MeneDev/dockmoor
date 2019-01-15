package resolver

import (
	"bytes"
	"context"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"io"
	"io/ioutil"
	"os"
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

func (repo dockerDaemonResolver) imageInspect(reference dockref.Reference) (types.ImageInspect, error) {
	ctx := context.Background()

	client, err := repo.newClient()
	if err != nil {
		return types.ImageInspect{}, err
	}

	imageInspect, _, err := client.ImageInspectWithRaw(ctx, reference.Original())

	return imageInspect, err
}

func (repo dockerDaemonResolver) newClient() (dockerAPIClient, error) {

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

type dockerAPIClient interface {
	ImageInspectWithRaw(ctx context.Context, image string) (types.ImageInspect, []byte, error)
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

func (d dockerCli) Client() dockerAPIClient {
	return d.cli.Client()
}

func newCli(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface {
	return &dockerCli{command.NewDockerCli(in, out, errWriter, isTrusted, nil)}
}

func (repo dockerDaemonResolver) FindAllTags(reference dockref.Reference) ([]dockref.Reference, error) {
	imageInspect, err := repo.imageInspect(reference)

	if err != nil {
		return nil, err
	}

	digs := imageInspect.RepoDigests
	tags := imageInspect.RepoTags

	refs := make([]dockref.Reference, 0)
	// TODO why can there more than one digest?
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
		r := dockref.MustParseAlgoDigest(imageInspect.ID)
		refs = append(refs, r)
	}

	return refs, nil
}

func (repo dockerDaemonResolver) Resolve(reference dockref.Reference) (dockref.Reference, error) {
	imageInspect, err := repo.imageInspect(reference)

	if err != nil {
		return nil, err
	}

	digs := imageInspect.RepoDigests
	tags := imageInspect.RepoTags

	// TODO why can there more than one digest?
	for _, tag := range tags {
		tagRef := dockref.MustParse(tag)
		tag = tagRef.Tag()
		if tag == "latest" && reference.Tag() == "" || reference.Tag() == tag {
			r := reference.WithTag(tagRef.Tag())
			for _, dig := range digs {
				digRef := dockref.MustParse(dig)
				r = r.WithDigest(digRef.DigestString())
				return r, nil
			}
		}
	}

	return nil, errors.Errorf("Could not find a image matching %s via docker daemon. Maybe you need to pull the image?", reference.Formatted())
}