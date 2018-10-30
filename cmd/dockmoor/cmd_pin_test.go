package main

import (
	"bytes"
	"errors"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/MeneDev/dockmoor/docktst/dockreftst"
	"github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

type pinOptionsTest struct {
	*pinOptions

	mainOptionsTest *mainOptionsTest
	mockRepo        *dockreftst.MockRepository
}

func (fo *pinOptionsTest) MainOptions() *mainOptionsTest {
	return fo.mainOptionsTest
}

func pinOptionsTestNew() *pinOptionsTest {
	mainOptions := mainOptionsTestNew()

	repo := dockreftst.MockRepositoryNew()

	pinOptions := &pinOptionsTest{
		pinOptions:      pinOptionsNew(mainOptions.mainOptions, repo),
		mainOptionsTest: mainOptions,
		mockRepo:        repo,
	}
	return pinOptions
}

func TestMainAsciiDocWithPin(t *testing.T) {
	t.SkipNow()
	os.Args = []string{"exe", "--asciidoc-usage"}

	mainOptions := mainOptionsACNew(addPinCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	helpText := buffer.String()
	assert.Contains(t, helpText, "pin command")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestMainMarkdownWithPin(t *testing.T) {
	t.SkipNow()
	os.Args = []string{"exe", "--markdown"}

	mainOptions := mainOptionsACNew(addPinCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	helpText := buffer.String()
	assert.Contains(t, helpText, "pin command")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestPinHelpIsNotAnError(t *testing.T) {
	t.SkipNow()

	os.Args = []string{"exe", "pin", "--help"}

	mainOptions := mainOptionsACNew(addPinCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	helpText := buffer.String()
	assert.Contains(t, helpText, "pin command")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestInvalidDockerfileWithPin(t *testing.T) {
	// given
	po := pinOptionsTestNew()
	mainOptions := po.MainOptions()

	formatProvider := mainOptions.FormatProvider()

	format := new(FormatMock)
	format.OnName().Return("mock")
	format.OnValidateInput(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("Not my department"))

	formatProvider.OnFormats().Return([]dockfmt.Format{format})

	mainOptions.formatProvider = formatProvider

	po.Positional.InputFile = flags.Filename(NotADockerfile)

	// when
	_, err := po.matchAndProcess()

	// then
	assert.NotNil(t, err)

	_, ok := err.(dockfmt.UnknownFormatError)
	assert.True(t, ok)
}

func TestPinOptions_RefFormat(t *testing.T) {
	po := pinOptionsTestNew()
	po.ReferenceFormat.ForceDomain = false
	po.ReferenceFormat.NoName = false
	po.ReferenceFormat.NoTag = false
	po.ReferenceFormat.NoDigest = false
	t.Run("all unset", func(t *testing.T) {
		format, e := po.RefFormat()
		assert.Nil(t, e)
		assert.False(t, (format&dockref.FormatHasDomain) != 0)
		assert.True(t, (format&dockref.FormatHasName) != 0)
		assert.True(t, (format&dockref.FormatHasTag) != 0)
		assert.True(t, (format&dockref.FormatHasDigest) != 0)
	})
	po.ReferenceFormat.ForceDomain = true
	po.ReferenceFormat.NoName = true
	po.ReferenceFormat.NoTag = true
	po.ReferenceFormat.NoDigest = true
	t.Run("all set", func(t *testing.T) {
		format, e := po.RefFormat()
		assert.Nil(t, e)
		assert.True(t, (format&dockref.FormatHasDomain) != 0)
		assert.False(t, (format&dockref.FormatHasName) != 0)
		assert.False(t, (format&dockref.FormatHasTag) != 0)
		assert.False(t, (format&dockref.FormatHasDigest) != 0)
	})
}

func TestPinCommandPins(t *testing.T) {
	po := pinOptionsTestNew()
	po.mockRepo.OnResolve(dockref.FromOriginalNoError("nginx")).
		Return(dockref.FromOriginalNoError("nginx:tag@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240"), nil)

	processorMock := &FormatProcessorMock{}

	pin := func(expected string) {
		processorMock.process = func(imageNameProcessor dockfmt.ImageNameProcessor) error {
			ref, e := imageNameProcessor(dockref.FromOriginalNoError("nginx"))
			assert.Nil(t, e)
			str := ref.String()
			assert.Equal(t, expected, str)
			return nil
		}
		po.matchAndProcessFormatProcessor(processorMock)
	}

	t.Run("tag and sha", func(t *testing.T) {
		po.ReferenceFormat.ForceDomain = false
		po.ReferenceFormat.NoName = false
		po.ReferenceFormat.NoTag = false
		po.ReferenceFormat.NoDigest = false

		pin("nginx:tag@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
	})
	t.Run("tag only", func(t *testing.T) {
		po.ReferenceFormat.ForceDomain = false
		po.ReferenceFormat.NoName = false
		po.ReferenceFormat.NoTag = false
		po.ReferenceFormat.NoDigest = true

		pin("nginx:tag")
	})
	t.Run("sha and name", func(t *testing.T) {
		po.ReferenceFormat.ForceDomain = false
		po.ReferenceFormat.NoName = false
		po.ReferenceFormat.NoTag = true
		po.ReferenceFormat.NoDigest = false

		pin("nginx@sha256:d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
	})
	t.Run("sha only", func(t *testing.T) {
		po.ReferenceFormat.ForceDomain = false
		po.ReferenceFormat.NoName = true
		po.ReferenceFormat.NoTag = true
		po.ReferenceFormat.NoDigest = false

		pin("d21b79794850b4b15d8d332b451d95351d14c951542942a816eea69c9e04b240")
	})

}

func TestFilenameRequiredWithPin(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{"pin"}, addPinCommand)
	assert.NotEqual(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "level=error")
	assert.Contains(t, stdout.String(), "the required argument `InputFile` was not provided")
}

func TestPinCallsFindExecuteWithPin(t *testing.T) {
	cmd, _, _, _ := testMain([]string{"pin", "fileName"}, addPinCommand)

	_, ok := cmd.(*pinOptions)
	assert.True(t, ok)
}
