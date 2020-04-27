package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockproc"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type containsOptionsTest struct {
	*containsOptions

	mainOptionsTest *mainOptionsTest
}

func (fo *containsOptionsTest) MainOptions() *mainOptionsTest {
	return fo.mainOptionsTest
}

func containsOptionsTestNew() *containsOptionsTest {
	mainOptions := mainOptionsTestNew()
	containsOptions := containsOptionsTest{
		containsOptions: containsOptionsNew(mainOptions.mainOptions),
		mainOptionsTest: mainOptions,
	}

	containsOptions.mainOpts = mainOptions.mainOptions

	return &containsOptions
}

type ReadableOpenerMock struct {
	mock.Mock
}

func (m *ReadableOpenerMock) Open(str string) (io.ReadCloser, error) {
	called := m.Called(str)
	return getReadCloser(called, 0), called.Error(1)
}

func getReadCloser(args mock.Arguments, index int) io.ReadCloser {
	obj := args.Get(index)
	var v io.ReadCloser
	var ok bool
	if obj == nil {
		return nil
	}
	if v, ok = obj.(io.ReadCloser); !ok {
		panic(fmt.Sprintf("assert: arguments: Error(%d) failed because object wasn't correct type: %v", index, args.Get(index)))
	}
	return v
}

func makeReadCloser(str string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBufferString(str))
}

func TestInvalidDockerfileWithContains(t *testing.T) {
	// given
	mainOptions := mainOptionsTestNew()

	formatProvider := mainOptions.FormatProvider()

	format := new(FormatMock)
	format.OnName().Return("mock")
	format.OnValidateInput(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not my department"))

	formatProvider.OnFormats().Return([]dockfmt.Format{format})

	mainOptions.formatProvider = formatProvider

	mo := &MatchingOptions{
		mainOpts: mainOptions.mainOptions,
	}

	mo.Positional.InputFile = flags.Filename(NotADockerfile)

	// when
	err := mo.WithFormatProcessorDo(nil, func(processor dockfmt.FormatProcessor) error {
		return nil
	})

	// then
	assert.NotNil(t, err)

	_, ok := err.(dockfmt.UnknownFormatError)
	assert.True(t, ok)
}

func TestReportInvalidPredicateWithContains(t *testing.T) {
	// given
	mainOptions := mainOptionsTestNew()
	stdout := bytes.NewBuffer(nil)
	mainOptions.SetStdout(stdout)

	formatProvider := mainOptions.FormatProvider()

	format := new(FormatMock)
	format.OnName().Return("mock")
	format.OnValidateInput(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	expected := errors.New("process Error")
	format.OnProcess(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expected)

	formatProvider.OnFormats().Return([]dockfmt.Format{format})

	mainOptions.formatProvider = formatProvider

	mopts := &MatchingOptions{
		mainOpts: mainOptions.mainOptions,
	}

	mopts.Positional.InputFile = flags.Filename(NotADockerfile)

	// when
	err := mopts.WithInputDo(func(filePathInput string, fpInput io.Reader) error {
		return mopts.WithFormatProcessorDo(fpInput, func(processor dockfmt.FormatProcessor) error {
			return processor.Process(func(r dockref.Reference) (dockref.Reference, error) {
				return r, nil
			})
		})
	})

	assert.Error(t, err)
}

func TestFilenameRequiredWithContains(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{"contains"}, addContainsCommand)
	assert.NotEqual(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "level=error")
	assert.Contains(t, stdout.String(), "the required argument `InputFile` was not provided")
}

func TestContainsCallsFindExecuteWithContains(t *testing.T) {
	cmd, _, _, _ := testMain([]string{"contains", "fileName"}, addContainsCommand)

	_, ok := cmd.(*containsOptions)
	assert.True(t, ok)
}

func TestOpenErrorsArePropagatedWithContains(t *testing.T) {
	fo := containsOptionsTestNew()
	fo.TagPredicates.Latest = true
	expectedError := errors.New("could not open")
	fo.MainOptions().openerMock.On("Open", mock.Anything).Return(nil, expectedError)

	//exitCode, err := fo.matchAndProcess()

	//assert.NotEqual(t, 0, exitCode)
	//assert.Equal(t, expectedError, err)
}

func TestExecuteReturnsErrorWithContains(t *testing.T) {
	fo := containsOptionsTestNew()
	expected := "use ExecuteWithExitCode instead"
	err := fo.Execute(nil)

	assert.Equal(t, expected, err.Error())
}

func TestMainMarkdownWithContains(t *testing.T) {
	os.Args = []string{"exe", "--markdown"}

	mainOptions := mainOptionsACNew(addContainsCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	assert.Contains(t, buffer.String(), "contains command")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestMainAsciiDocWithContains(t *testing.T) {
	os.Args = []string{"exe", "--asciidoc-usage"}

	mainOptions := mainOptionsACNew(addContainsCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	assert.Contains(t, buffer.String(), "contains command")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestContainsHelpIsNotAnError(t *testing.T) {
	os.Args = []string{"exe", "contains", "--help"}

	mainOptions := mainOptionsACNew(addContainsCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	assert.Contains(t, buffer.String(), "contains command")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestContainsCommandDoesntPrint(t *testing.T) {
	test := containsOptionsTestNew()
	stdout := test.MainOptions().Stdout()

	processorMock := &FormatProcessorMock{}
	processorMock.process = func(imageNameProcessor dockfmt.ImageNameProcessor) error {
		r, _ := dockref.Parse("nginx")
		imageNameProcessor(r)
		r, _ = dockref.Parse("nginx:latest")
		imageNameProcessor(r)
		r, _ = dockref.Parse("nginx:1.2")
		imageNameProcessor(r)
		return nil
	}

	predicate, e := dockproc.AnyPredicateNew()
	assert.Nil(t, e)

	test.applyFormatProcessor(predicate, processorMock)
	s := stdout.String()
	assert.Empty(t, s)
}

func equalsAnyString(needle string, values ...string) bool {
	for _, v := range values {
		if needle == v {
			return true
		}
	}

	return false
}
