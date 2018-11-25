package dockref

import (
	"bytes"
	"context"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/spf13/pflag"
	"io"
	"io/ioutil"
	"os"
)

type Repository interface {
	Resolve(reference Reference) ([]Reference, error)
}

type dockerDaemonRepository struct {
	ImageInspect func(reference Reference) (types.ImageInspect, error)
	NewCli       func(in io.ReadCloser, out *bytes.Buffer, errWriter *bytes.Buffer, isTrusted bool) dockerCliInterface
}

var _ Repository = (*dockerDaemonRepository)(nil)

func DockerDaemonRepositoryNew() Repository {
	repo := &dockerDaemonRepository{
		NewCli: newCli,
	}
	return repo
}

func (repo dockerDaemonRepository) imageInspect(reference Reference) (types.ImageInspect, error) {
	ctx := context.Background()

	client, err := repo.newClient()
	if err != nil {
		return types.ImageInspect{}, err
	}

	imageInspect, _, err := client.ImageInspectWithRaw(ctx, reference.Original())

	return imageInspect, err
}

var (
	dockerTLSVerify = os.Getenv("DOCKER_TLS_VERIFY") != ""
	dockerTLS       = os.Getenv("DOCKER_TLS") != ""
)

func (repo dockerDaemonRepository) newClient() (dockerAPIClient, error) {
	in := ioutil.NopCloser(bytes.NewBuffer(nil))
	out := bytes.NewBuffer(nil)
	errWriter := bytes.NewBuffer(nil)
	isTrusted := false
	cli := repo.NewCli(in, out, errWriter, isTrusted)
	cliOpts := flags.NewClientOptions()

	tls := dockerTLS || dockerTLSVerify
	host, e := opts.ParseHost(tls, os.Getenv("DOCKER_HOST"))
	if e != nil {
		return nil, e
	}
	cliOpts.Common.TLS = tls
	cliOpts.Common.TLSVerify = dockerTLSVerify
	cliOpts.Common.Hosts = []string{host}

	if tls {
		flags := pflag.NewFlagSet("testing", pflag.ContinueOnError)
		cliOpts.Common.InstallFlags(flags)
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

func (repo dockerDaemonRepository) Resolve(reference Reference) ([]Reference, error) {
	imageInspect, err := repo.imageInspect(reference)

	if err != nil {
		return nil, err
	}

	digs := imageInspect.RepoDigests
	tags := imageInspect.RepoTags

	refs := make([]Reference, 0)
	// TODO why can there more than one digest?
	for _, tag := range tags {
		tagRef := FromOriginalNoError(tag)
		r := reference.WithTag(tagRef.Tag())
		for _, dig := range digs {
			digRef := FromOriginalNoError(dig)
			r = r.WithDigest(digRef.DigestString())
			refs = append(refs, r)
		}

		if len(digs) == 0 {
			refs = append(refs, r)
		}
	}

	if len(digs) == 0 && len(tags) == 0 {
		r := MustParseAlgoDigest(imageInspect.ID)
		refs = append(refs, r)
	}

	return refs, nil
}
