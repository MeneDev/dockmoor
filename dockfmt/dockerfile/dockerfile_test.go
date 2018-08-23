package dockerfile

import (
	"testing"
	"github.com/sirupsen/logrus"
	"bytes"
	"strings"
	"github.com/stretchr/testify/assert"
	"github.com/MeneDev/dockfix/dockref"
	"github.com/pkg/errors"
	"io"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

var log = logrus.New()

func init()  {
	log.SetOutput(bytes.NewBuffer(nil))
}

func TestDockerfileName(t *testing.T) {
	format := DockerfileFormatNew()
	name := format.Name()
	assert.Equal(t, "Dockerfile", name)
}

func TestDockerfileFormatEmptyIsInvalid(t *testing.T) {
	file := ``
	format := DockerfileFormatNew()
	valid := format.ValidateInput(log, strings.NewReader(file), "anything")

	assert.Error(t, valid)
}

func TestDockerfileFormatMissingFromIsInvalid(t *testing.T) {
	file := `RUN command`
	format := DockerfileFormatNew()
	valid := format.ValidateInput(log, strings.NewReader(file), "anything")

	assert.Error(t, valid)
}

func TestDockerfileFormatOtherIsInvalid(t *testing.T) {
	file := `other stuff`
	format := DockerfileFormatNew()
	valid := format.ValidateInput(log, strings.NewReader(file), "anything")

	assert.Error(t, valid)
}

func TestDockerfileFromScratchIsValid(t *testing.T) {
	file := `FROM scratch`
	format := DockerfileFormatNew()
	valid := format.ValidateInput(log, strings.NewReader(file), "anything")

	assert.Nil(t, valid)
}

func TestDockerfileFromScratchPlusInvalidIsInvalid(t *testing.T) {
	file := `FROM scratch
Invalid thing`
	format := DockerfileFormatNew()
	valid := format.ValidateInput(log, strings.NewReader(file), "anything")

	assert.Error(t, valid)
}

func TestDockerfileFromNginxIsValid(t *testing.T) {
	file := `FROM nginx`
	format := DockerfileFormatNew()
	valid := format.ValidateInput(log, strings.NewReader(file), "anything")

	assert.Nil(t, valid)
}

func TestDockerfileFromNginxWithTagIsValid(t *testing.T) {
	file := `FROM nginx:tag`
	format := DockerfileFormatNew()
	valid := format.ValidateInput(log, strings.NewReader(file), "anything")

	assert.Nil(t, valid)
}

func TestDockerfileMultiFromIsValid(t *testing.T) {
	file := `FROM nginx:tag
FROM something:tag`
	format := DockerfileFormatNew()
	valid := format.ValidateInput(log, strings.NewReader(file), "anything")

	assert.Nil(t, valid)
}
func TestMultilineCommandIsValid(t *testing.T) {
	file := `FROM nginx:tag
RUN some \
	command`

	format := DockerfileFormatNew()
	valid := format.ValidateInput(log, strings.NewReader(file), "anything")

	assert.Nil(t, valid)
}

func TestDockerfileMultiFromIsCalls(t *testing.T) {
	file := `FROM nginx:tag
RUN some \
	command

FROM something:tag`
	format := DockerfileFormatNew()
	calls := 0
	format.ValidateInput(log, strings.NewReader(file), "anything")

	err := format.Process(log, strings.NewReader(file), bytes.NewBuffer(nil), func(r dockref.Reference) (string, error) {
		calls++
		return "", nil
	})

	assert.Nil(t, err)
	assert.Equal(t, 2, calls)
}

func TestDockerfilePassProcessorErrors(t *testing.T) {
	file := `FROM valid`
	format := DockerfileFormatNew()
	format.ValidateInput(log, strings.NewReader(file), "anything")

	expected := errors.New("Expected")
	err := format.Process(log, strings.NewReader(file), bytes.NewBuffer(nil), func(r dockref.Reference) (string, error) {
		return "", expected
	})

	assert.Equal(t, expected, err)
}

func TestDockerfilePassMultiLineAndMultistage(t *testing.T) {
	file := `FROM nginx:tag
RUN some \
	command

FROM something:tag
RUN something \
	in the end

# And a comment`
	format := DockerfileFormatNew()
	format.ValidateInput(log, strings.NewReader(file), "anything")

	calls := 0
	err := format.Process(log, strings.NewReader(file), bytes.NewBuffer(nil), func(r dockref.Reference) (string, error) {
		calls++
		return "", nil
	})

	assert.Nil(t, err)
	assert.Equal(t, 2, calls)
}

func TestDockerfileInvalidFromReported(t *testing.T) {
	file := `FROM nginx:a:b`
	format := DockerfileFormatNew()
	format.ValidateInput(log, strings.NewReader(file), "anything")

	processErr := format.Process(log, strings.NewReader(file), bytes.NewBuffer(nil), func(r dockref.Reference) (string, error) {
		return "", nil
	})

	assert.Error(t, processErr)
}

func TestParserErrorsAreReported(t *testing.T) {
	file := `FROM nginx:a:b`
	format := DockerfileFormatNew()

	expected := errors.New("expected")
	format.parseFunction = func(rwc io.Reader) (*parser.Result, error) {
		return nil, expected
	}

	err := format.ValidateInput(log, strings.NewReader(file), "anything")

	assert.Equal(t, expected, err)
}

