package dockref

import (
	"bytes"
	"context"
	"io/ioutil"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
)

type Repository interface {
	Resolve(reference Reference) ([]Reference, error)
}

type dockerDaemonRepository struct {
}

var _ Repository = (*dockerDaemonRepository)(nil)

func DockerDaemonRepositoryNew() Repository {
	repo := &dockerDaemonRepository{}
	return repo
}

func (dockerDaemonRepository) Resolve(reference Reference) ([]Reference, error) {
	in := ioutil.NopCloser(bytes.NewBuffer(nil))
	out := bytes.NewBuffer(nil)
	errWriter := bytes.NewBuffer(nil)
	isTrusted := false
	cli := command.NewDockerCli(in, out, errWriter, isTrusted, nil)
	ctx := context.Background()
	opts := flags.NewClientOptions()
	cli.Initialize(opts)
	client := cli.Client()
	imageInspect, _, err := client.ImageInspectWithRaw(ctx, reference.Original())
	if err != nil {
		return nil, err
	}
	digs := imageInspect.RepoDigests
	tags := imageInspect.RepoTags
	refs := make([]Reference, 0)
	// TODO why can there more than one digest?
	for _, dig := range digs {
		digRef := FromOriginalNoError(dig)
		r := reference.WithDigest(digRef.DigestString())
		for _, tag := range tags {
			tagRef := FromOriginalNoError(tag)
			r = r.WithTag(tagRef.Tag())
			refs = append(refs, r)
		}

		if len(tags) == 0 {
			refs = append(refs, r)
		}
	}

	return refs, nil
}
