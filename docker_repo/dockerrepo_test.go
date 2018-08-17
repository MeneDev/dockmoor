package docker_repo

import (
	"context"
	"github.com/gianarb/testcontainer-go"
	"testing"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"io/ioutil"
	"archive/tar"
	"bytes"
	"log"
	"net/http"
	regcli "github.com/docker/distribution/registry/client"
	"time"
	"crypto/tls"
	"github.com/stretchr/testify/assert"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/docker/registry"
)

//func TestEmptyRegistryIsEmpty(t *testing.T) {
//	ctx := context.Background()
//	registryC, err := testcontainer.RunContainer(ctx, "menedev/testable-registry:2.6.2", testcontainer.RequestContainer{
//		ExportedPort: []string{
//			"5000/tpc",
//		},
//	})
//	if err != nil {
//		t.Error(err)
//	}
//	defer registryC.Terminate(ctx, t)
//
//	ip, err := registryC.GetIPAddress(ctx)
//	if err != nil {
//		t.Error(err)
//	}
//
//
//	rgs, _ := regcli.NewRegistry("https://"+ip+":5000", tripper)
//
//	repos := make([]string, 10)
//	n, err := rgs.Repositories(ctx, repos, "")
//
//	assert.Equal(t, 0, n)
//	assert.Equal(t, "EOF", err.Error())
//}

func TestEmptyTestImages(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	//rgs, _ := regcli.NewRegistry("docker.io", http.DefaultTransport)

	named, _ := reference.WithName("menedev/testimagea")
	challengeManager1 := challenge.NewSimpleManager()

	authConfig := types.AuthConfig{
		Username: "menedev",
		Password: "gxm823p7iyD4Axzk4vOv",
	}

	auth.NewTokenHandlerWithOptions(auth.TokenHandlerOptions{})
	transport.NewTransport(nil, auth.NewAuthorizer(challengeManager1, auth.NewBasicHandler(registry.NewStaticCredentialStore(&authConfig))))
	rep, _ := regcli.NewRepository(named, "https://registry.hub.docker.com/", http.DefaultTransport)


	tags := rep.Tags(ctx)
	tagList, e := tags.All(ctx)
	if e != nil {
		t.Error(e)
	}
	assert.Equal(t, 0, len(tagList))
}

func TestNginxLatestReturn(t *testing.T) {
	ctx := context.Background()
	registryC, err := testcontainer.RunContainer(ctx, "menedev/testable-registry:2.6.2", testcontainer.RequestContainer{
		ExportedPort: []string{
			"5000/tpc",
		},
	})
	if err != nil {
		t.Error(err)
	}
	time.Sleep(5 * 1000 * 1000 * 1000)
	defer registryC.Terminate(ctx, t)
	ip, err := registryC.GetIPAddress(ctx)
	if err != nil {
		t.Error(err)
	}

	var tripper = http.DefaultTransport.(*http.Transport)
	tripper.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	rgs, _ := regcli.NewRegistry("https://"+ip+":5000", tripper)

	repos := make([]string, 10)
	n, err := rgs.Repositories(ctx, repos, "")

	if err.Error() != "EOF" {
		t.Fail()
	}

	return
	print(n)
	//httpClient, err := defaultHTTPClient(t)
	//if err != nil {
	//	t.Fatal(err)
	//}

	cli, e := client.NewClientWithOpts(client.WithVersion("1.30"))
	if e != nil {
		panic(e)
	}
	s := `FROM alpine
RUN echo 1`

	tarHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(s)),
	}

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()
	err = tw.WriteHeader(tarHeader)
	if err != nil {
		log.Fatal(err, " :unable to write tar header")
	}

	_, err = tw.Write([]byte(s))
	if err != nil {
		log.Fatal(err, " :unable to write tar body")
	}
	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	ref := ip + ":5000/a"

	response, e := cli.ImageBuild(context.Background(), dockerFileTarReader, types.ImageBuildOptions{
		Tags:       []string{ref, ref + ":a", ref + ":b", ref + ":latest"},
		Dockerfile: "Dockerfile",
		Context:    dockerFileTarReader,
	})
	if e != nil {
		panic(e)
	}

	defer response.Body.Close()

	ioutil.ReadAll(response.Body)

	cli, e = client.NewClientWithOpts(client.WithVersion("1.30"))
	if e != nil {
		panic(e)
	}

	closer, e := cli.ImagePush(context.Background(), ref, types.ImagePushOptions{
		RegistryAuth: "required",
	})
	if e != nil {
		panic(e)
	}
	ioutil.ReadAll(closer)
	closer.Close()

	//rgs, _ := regcli.NewRegistry()
	//rgs.Repositories()
	//rgs.
	//named, _ := reference.WithName("a")
	//rep, _ := regcli.NewRepository(named, ip+":5000", http.DefaultTransport)
	//
	//blobWriter, _ := rep.Blobs(ctx).Create(ctx)
	//blobWriter.Commit(ctx, )
	//rep.Tags(ctx)
	//
	//_, _ := rep.Manifests()
	//
	//descriptor, _ := rep.Tags(ctx).Get()
	//descriptor.Digest

	readCloser, e := cli.ImagePull(context.Background(), ref, types.ImagePullOptions{
		RegistryAuth: "required",
	})
	if e != nil {
		panic(e)
	}
	ioutil.ReadAll(readCloser)
	readCloser.Close()

	repo, e := DockerRepoNew()
	if e != nil {
		panic(e)
	}

	inspect, e := repo.DistributionInspect(ref)
	if e != nil {
		panic(e)
	}

	print(inspect.Descriptor.Annotations)
}
