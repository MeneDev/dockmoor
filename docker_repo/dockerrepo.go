package docker_repo

import (
	"github.com/opencontainers/go-digest"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"context"
	"io"
	"github.com/pkg/errors"
	"io/ioutil"
	"github.com/docker/docker/api/types/registry"
	)

type DockerRepo interface {
	io.Closer

	FindByNameAndTag(name string, tag string) (reference.Canonical, error)
	FindDigest(dig digest.Digest) (reference.Digested, error)
	Pull(tagged reference.NamedTagged) (reference.Canonical, error)
	DistributionInspect(ref string) (registry.DistributionInspect, error)
}

type dockerRepo struct {
	DockerRepo

	cli *client.Client
	maps map[string]map[string]string
}

func (repo dockerRepo) Close() error {
	return repo.cli.Close()
}

func DockerRepoNew() (DockerRepo, error) {
	cli, e := client.NewClientWithOpts(client.WithVersion("1.30"))
	if e != nil {
		return nil, e
	}
	repo := dockerRepo{
		cli : cli,
		maps:     make(map[string]map[string]string),
	}

	e = repo.collect()
	if e != nil {
		return nil, e
	}

	return repo, nil
}

func (repo dockerRepo) FindByNameAndTag(name string, tag string) (reference.Canonical, error) {
	named, e := reference.WithName(name)
	if e != nil {
		return nil, e
	}

	tagged, e := reference.WithTag(named, tag)
	if e != nil {
		return nil, e
	}

	rep := repo.maps[named.String()]
	if rep == nil {
		return nil, errors.Errorf("Unknown Image name '%s' (resolved as '%s')", name, named)
	}

	dig := rep[tag]
	if dig == "" {
		return nil, errors.Errorf("Unknown Tag '%s' for Image '%s' (resolved as '%s')", tag, name, tagged)
	}

	canonical, e := reference.WithDigest(tagged, digest.Digest(dig))
	if e != nil {
		return nil, e
	}

	return canonical, nil
}

func (repo dockerRepo) FindDigest(dig digest.Digest) (reference.Digested, error) {
	str := string(dig)
	tagName := repo.maps[str]
	if tagName == nil {
		return nil, errors.Errorf("Unknown digest: %s", dig)
	}
	for tag, name := range tagName {
		ref, err := reference.WithName(name)
		if err != nil {
			panic(err)
		}

		if tag != "" {
			ref, err = reference.WithTag(ref, tag)
			if err != nil {
				panic(err)
			}
		}

		digested, err := reference.WithDigest(ref, dig)

		return digested, err
	}

	return nil, errors.Errorf("Unknown digest: %s", dig)
}

func (repo dockerRepo) Pull(tagged reference.NamedTagged) (reference.Canonical, error) {
	options := types.ImagePullOptions{}

	//domain := reference.Domain(named)
	//

	//auth := types.auths[domain]
	//if auth != "" {
	//	options.RegistryAuth = auth
	//}

	resp, err := repo.cli.ImagePull(context.Background(), tagged.String(), options)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	//println("No known " + named.String() + " locally. Pulling...")
	_, err = ioutil.ReadAll(resp)
	if err != nil {
		return nil, err
	}

	repo.collect()

	canonical, err := repo.FindByNameAndTag(tagged.Name(), tagged.String())
	if err != nil {
		return nil, err
	}

	return canonical, nil
}

func (repo dockerRepo) DistributionInspect(ref string) (registry.DistributionInspect, error) {
	inspect, e := repo.cli.DistributionInspect(context.Background(), ref, "")
	return inspect, e
}

func (repo dockerRepo) collect() error {

	cli := repo.cli
	repositories := repo.maps

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return err
	}

	for _, image := range images {

		if len(image.RepoDigests) == 0 {
			println("INFO: Skipping " + image.ID + ": no RepoDigests")
			continue
		} else {
			normalizedNamed, err := reference.ParseNormalizedNamed(image.RepoDigests[0])
			if err != nil {
				if image.RepoDigests[0] == "<none>@<none>" {
					// special-case: image without a name and without a tag
					// just ignored for now
					continue
				}
				return err
			}
			name := normalizedNamed.Name()
			digRef, ok := normalizedNamed.(reference.Digested)
			if !ok {
				return errors.Errorf("RepoDigest without Digest")
			}

			dig := string(digRef.Digest())

			repo := repositories[name]
			if repo == nil {
				repo = make(map[string]string)
				repositories[name] = repo
			}

			repositories[dig] = make(map[string]string)

			for _, tag := range image.RepoTags {
				ref, err := reference.ParseAnyReference(tag)
				if err != nil {
					panic(err)
				}
				if ref, ok := ref.(reference.Tagged); ok {
					tag = ref.Tag()
					repo[tag] = dig
				}
				repositories[dig][tag] = name
			}

			if len(image.RepoTags) == 0 {
				repositories[dig][""] = name
			}
		}

	}

	return nil
}