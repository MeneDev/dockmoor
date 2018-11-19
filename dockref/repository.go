package dockref

import (
	"bytes"
	"context"
	"github.com/docker/docker/api/types"
	"io"
	"io/ioutil"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
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

func (repo dockerDaemonRepository) newClient() (dockerApiClient, error) {
	in := ioutil.NopCloser(bytes.NewBuffer(nil))
	out := bytes.NewBuffer(nil)
	errWriter := bytes.NewBuffer(nil)
	isTrusted := false
	cli := repo.NewCli(in, out, errWriter, isTrusted)
	opts := flags.NewClientOptions()
	err := cli.Initialize(opts)
	if err != nil {
		return nil, err
	}
	client := cli.Client()
	return client, nil
}

type dockerApiClient interface {
	ImageInspectWithRaw(ctx context.Context, image string) (types.ImageInspect, []byte, error)
}

type dockerCliInterface interface {
	Initialize(options *flags.ClientOptions) error
	Client() dockerApiClient
}

type dockerCli struct {
	cli *command.DockerCli
}

func (d dockerCli) Initialize(options *flags.ClientOptions) error {
	return d.cli.Initialize(options)
}

func (d dockerCli) Client() dockerApiClient {
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
