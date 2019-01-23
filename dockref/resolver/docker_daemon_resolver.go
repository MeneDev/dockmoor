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
	"strings"
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

func (reser dockerDaemonResolver) FindAllTags(reference dockref.Reference) ([]dockref.Reference, error) {
	imageInspect, err := reser.imageInspect(reference)

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

func (reser dockerDaemonResolver) Resolve(reference dockref.Reference) (dockref.Reference, error) {
	imageInspect, err := reser.imageInspect(reference)

	if err != nil {
		return nil, err
	}

	digs := imageInspect.RepoDigests
	//tags := imageInspect.RepoTags

	// TODO there can be multiple repo digests (with the same digest) when the image has been pulled from multiple
	// registries
	// eg
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

	if dig == nil {
		return nil, errors.Errorf(
			"ambigious RepoDigests [%s] for reference %s",
			strings.Join(imageInspect.RepoDigests, ", "),
			reference.Original())
	}
	//
	//digDomain := dig.Domain()
	//
	//tagsFilters := []func(tag string) bool{
	//	func(tag string) bool {
	//		return true
	//	},
	//
	//	func(dig string) bool {
	//		digRef := dockref.MustParse(dig)
	//
	//		return digRef.Domain() == digDomain
	//	},
	//}
	//
	//
	//var tag dockref.Reference
	//for _, filter := range tagsFilters {
	//	tags = filterStrings(tags, filter)
	//	for _, t := range tags {
	//		tagRef := dockref.MustParse(t)
	//		t = tagRef.Tag()
	//		implicitLatestTag := t == "latest" && reference.Tag() == ""
	//		tagsEqual := reference.Tag() == t
	//		onlyTag := reference.Tag() == "" && len(tags) == 1
	//		if implicitLatestTag || tagsEqual || onlyTag {
	//			tag = tagRef
	//			break
	//		}
	//	}
	//}

	//if tag == "" {
	//	return nil, errors.Errorf("non of the tags [%s] matched for %s on %s",
	//		strings.Join(imageInspect.RepoTags, ","),
	//		reference.Original(),
	//		digDomain)
	//}

	reference = reference.WithDigest(dig.DigestString())
	//if tag != nil {
	//	reference = reference.WithTag(tag.Tag())
	//}

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
