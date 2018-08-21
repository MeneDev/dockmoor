package dockerfile

import (
	"testing"
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"github.com/opencontainers/go-digest"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/distribution/reference"
	"github.com/stretchr/testify/suite"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/MeneDev/dockfix/dockfmt"
	"github.com/MeneDev/dockfix"
)

type PinSuite struct {
	suite.Suite
	repo *MockDockerRepo
}

type MockDockerRepo struct {
	mock.Mock
}

func (MockDockerRepo) Close() error {
	panic("implement me")
}

func (m *MockDockerRepo) FindByNameAndTag(name string, tag string) (reference.Canonical, error) {
	called := m.Called(name, tag)
	return getCanonical(called, 0), called.Error(1)
}

func getCanonical(args mock.Arguments, index int) reference.Canonical {
	obj := args.Get(index)
	var s reference.Canonical
	var ok bool
	if obj == nil {
		return nil
	}
	if s, ok = obj.(reference.Canonical); !ok {
		panic(fmt.Sprintf("assert: arguments: Error(%d) failed because object wasn't correct type: %v", index, args.Get(index)))
	}
	return s
}

func getNameTagged(args mock.Arguments, index int) reference.NamedTagged {
	obj := args.Get(index)
	var s reference.NamedTagged
	var ok bool
	if obj == nil {
		return nil
	}
	if s, ok = obj.(reference.NamedTagged); !ok {
		panic(fmt.Sprintf("assert: arguments: Error(%d) failed because object wasn't correct type: %v", index, args.Get(index)))
	}
	return s
}

func (m *MockDockerRepo) FindDigest(dig digest.Digest) (reference.Digested, error) {
	called := m.Called(dig)
	get := called.Get(0)
	return get.(reference.Digested), called.Error(1)
}

func (m *MockDockerRepo) Pull(tagged reference.NamedTagged) (reference.Canonical, error) {
	called := m.Called(tagged)
	return getCanonical(called, 0), called.Error(1)
}

func (m *MockDockerRepo) DistributionInspect(ref string) (registry.DistributionInspect, error) {
	panic("implement me")
}

func nameTagged(name string, tag string) reference.NamedTagged {
	named, _ := reference.WithName(name)
	tagged, _ := reference.WithTag(named, tag)
	return tagged
}

func canon(name string, tag string, dig string) reference.Canonical {
	named, _ := reference.WithName(name)
	tagged, _ := reference.WithTag(named, tag)
	canonical, _ := reference.WithDigest(tagged, digest.Digest(dig))
	return canonical
}

func (suite *PinSuite) SetupTest() {
	suite.repo = new(MockDockerRepo)

	suite.repo.On("FindByNameAndTag", "docker.io/library/alpine", "3.8").Return(canon("docker.io/library/alpine", "3.8", "sha256:7043076348bf5040220df6ad703798fd8593a0918d06d3ce30c6c93be117e430"), nil)
	suite.repo.On("FindByNameAndTag", "docker.io/library/alpine", "3.7").Return(canon("docker.io/library/alpine", "3.7", "sha256:56e2f91ef15847a2b02a5a03cbfa483949d67a242c37e33ea178e3e7e01e0dfd"), nil)
	suite.repo.On("FindByNameAndTag", "docker.io/library/busybox", "latest").Return(canon("docker.io/library/busybox", "latest", "sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"), nil)
	suite.repo.On("FindByNameAndTag", "docker.io/library/busybox", "1").Return(canon("docker.io/library/busybox", "1", "sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"), nil)
	suite.repo.On("FindByNameAndTag", "docker.io/library/busybox", "1.29").Return(canon("docker.io/library/busybox", "1.29", "sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"), nil)
	suite.repo.On("FindByNameAndTag", "docker.io/library/busybox", "1.29.1").Return(canon("docker.io/library/busybox", "1.29.1", "sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"), nil)

	suite.repo.On("FindDigest", digest.Digest("sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")).Return(canon("docker.io/library/busybox", "1.29.1", "sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"), nil)

	suite.repo.On("FindByNameAndTag", "docker.io/library/locally_unknown", "1").Return(nil, nil)
	suite.repo.On("Pull", nameTagged("docker.io/library/locally_unknown", "1")).Return(canon("docker.io/library/locally_unknown", "1", "sha256:1111111111111111111111111111111111111111111111111111111111111111"), nil)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDockerfileTestSuite(t *testing.T) {
	t.SkipNow()
	suite.Run(t, new(PinSuite))
}

var _ dockfmt.FormatProvider = (*Provider)(nil)
type Provider struct {}

func (Provider) Formats() []dockfmt.Format {
	return []dockfmt.Format{
		DockerfileFormatNew(),
	}
}

func doPin(repo *MockDockerRepo, reader *strings.Reader, buffer *bytes.Buffer) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	formatProvider := Provider {}

	dockfix.Pin(log, formatProvider, repo, reader, "Dockerfile", buffer)
}

func (suite *PinSuite) TestPinAlpine() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM alpine:3.8`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/alpine:3.8@sha256:7043076348bf5040220df6ad703798fd8593a0918d06d3ce30c6c93be117e430`

	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}

func (suite *PinSuite) TestPinBusyboxNoTag() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM busybox`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/busybox:latest@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240`

	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}

func (suite *PinSuite) TestPinBusyboxLatestTag() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM busybox:latest`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/busybox:latest@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240`

	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}

func (suite *PinSuite) TestPinBusybox1Tag() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM busybox:1`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/busybox:1@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240`

	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}

func (suite *PinSuite) TestPinBusybox1_29Tag() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM busybox:1.29`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/busybox:1.29@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240`

	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}

func (suite *PinSuite) TestPinBusybox1_29_1Tag() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM busybox:1.29.1`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/busybox:1.29.1@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240`

	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}

func (suite *PinSuite) TestPinBusybox1_29_1TagAndDomain() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM docker.io/library/busybox:1.29.1`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/busybox:1.29.1@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240`

	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}

func (suite *PinSuite) TestPinFromSha() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/busybox:1.29.1@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240`

	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}

func (suite *PinSuite) TestPinAlpineMultiple() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM alpine:3.8
FROM busybox:1.29.1
FROM busybox:1`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/alpine:3.8@sha256:7043076348bf5040220df6ad703798fd8593a0918d06d3ce30c6c93be117e430
FROM docker.io/library/busybox:1.29.1@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240
FROM docker.io/library/busybox:1@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240`

	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}


func (suite *PinSuite) TestPinAlpineAsSomething() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM alpine:3.8 AS something`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/alpine:3.8@sha256:7043076348bf5040220df6ad703798fd8593a0918d06d3ce30c6c93be117e430 AS something`

	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}


func (suite *PinSuite) TestIgnoreOtherDirectives() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM alpine:3.8
RUN something
WORKDIR somewhere
USER someome`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/alpine:3.8@sha256:7043076348bf5040220df6ad703798fd8593a0918d06d3ce30c6c93be117e430
RUN something
WORKDIR somewhere
USER someome`
	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}


func (suite *PinSuite) TestLocallyUnknown() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM locally_unknown:1`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/locally_unknown:1@sha256:1111111111111111111111111111111111111111111111111111111111111111`
	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	dockerRepo.AssertCalled(suite.T(), "Pull", nameTagged("docker.io/library/locally_unknown", "1"))
	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}


func (suite *PinSuite) TestKeepEmptyLinesAtBeginning() {
	dockerRepo := suite.repo

	dockerfileIn := `

FROM alpine:3.8
RUN something`
	var buffer bytes.Buffer

	expected := `

FROM docker.io/library/alpine:3.8@sha256:7043076348bf5040220df6ad703798fd8593a0918d06d3ce30c6c93be117e430
RUN something`
	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}


func (suite *PinSuite) TestKeepEmptyLinesInBetween() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM alpine:3.8


RUN something`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/alpine:3.8@sha256:7043076348bf5040220df6ad703798fd8593a0918d06d3ce30c6c93be117e430


RUN something`
	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}


func (suite *PinSuite) TestKeepEmptyLinesAtEnd() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM alpine:3.8

`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/alpine:3.8@sha256:7043076348bf5040220df6ad703798fd8593a0918d06d3ce30c6c93be117e430

`
	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}

func (suite *PinSuite) TestKeepCommentsAtBeginning() {
	dockerRepo := suite.repo

	dockerfileIn := `# A comment
FROM alpine:3.8`
	var buffer bytes.Buffer

	expected := `# A comment
FROM docker.io/library/alpine:3.8@sha256:7043076348bf5040220df6ad703798fd8593a0918d06d3ce30c6c93be117e430`
	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}

func (suite *PinSuite) TestKeepCommentsAtEndOfLine() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM alpine:3.8 # A comment`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/alpine:3.8@sha256:7043076348bf5040220df6ad703798fd8593a0918d06d3ce30c6c93be117e430 # A comment`
	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}

func (suite *PinSuite) TestMultiLineIdeal() {
	dockerRepo := suite.repo

	dockerfileIn := `FROM alpine:3.8 \
AS something
RUN something`
	var buffer bytes.Buffer

	expected := `FROM docker.io/library/alpine:3.8@sha256:7043076348bf5040220df6ad703798fd8593a0918d06d3ce30c6c93be117e430 \
AS something
RUN something`
	doPin(dockerRepo, strings.NewReader(dockerfileIn), &buffer)

	assert.Equal(suite.T(), expected, buffer.String(), "they should be equal")
}

