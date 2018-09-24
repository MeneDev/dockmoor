package main

import (
	"bytes"
	"fmt"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockproc"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

type containsOptionsTest struct {
	*MatchingOptions

	mainOptionsTest *mainOptionsTest
}

func (fo *containsOptionsTest) MainOptions() *mainOptionsTest {
	return fo.mainOptionsTest
}

func containsOptionsTestNew() *containsOptionsTest {
	mainOptions := mainOptionsTestNew()
	containsOptions := containsOptionsTest{
		MatchingOptions: &MatchingOptions{},
		mainOptionsTest: mainOptions,
	}

	containsOptions.mainOpts = mainOptions.mainOptions
	containsOptions.mode = matchOnly

	return &containsOptions
}

func TestEmptyPredicates(t *testing.T) {
	fo := &MatchingOptions{}
	err := verifyContainsOptions(fo)
	assert.Equal(t, ErrAtLeastOnePredicate, err)
}

func TestSingleExclusivePredicatesFail(t *testing.T) {
	strings := []string{"any", "latest", "unpinned", "outdated"}
	for _, a := range strings {
		t.Run(a, func(t *testing.T) {
			fo := &MatchingOptions{}
			fo.Predicates.Any = equalsAnyString("any", a)
			fo.Predicates.Outdated = equalsAnyString("outdated", a)
			fo.Predicates.Unpinned = equalsAnyString("unpinned", a)
			fo.Predicates.Latest = equalsAnyString("latest", a)
			err := verifyContainsOptions(fo)
			assert.Nil(t, err)
		})
	}
}

func TestMultipleExclusivePredicatesFail(t *testing.T) {

	strings := []string{"any", "latest", "unpinned", "outdated"}
	for _, a := range strings {
		for _, b := range strings {
			if a == b {
				continue
			}

			t.Run(a+" and "+b, func(t *testing.T) {
				fo := &MatchingOptions{}
				fo.Predicates.Any = equalsAnyString("any", a, b)
				fo.Predicates.Outdated = equalsAnyString("outdated", a, b)
				fo.Predicates.Unpinned = equalsAnyString("unpinned", a, b)
				fo.Predicates.Latest = equalsAnyString("latest", a, b)
				err := verifyContainsOptions(fo)
				assert.Equal(t, ErrAtMostOnePredicate, err)
			})
		}
	}

	for _, a := range strings {
		for _, b := range strings {
			if a == b {
				continue
			}

			for _, c := range strings {
				if a == c {
					continue
				}

				if b == c {
					continue
				}

				t.Run(a+" and "+b+" and "+c, func(t *testing.T) {
					fo := &MatchingOptions{}
					fo.Predicates.Any = equalsAnyString("any", a, b, c)
					fo.Predicates.Outdated = equalsAnyString("outdated", a, b, c)
					fo.Predicates.Unpinned = equalsAnyString("unpinned", a, b, c)
					fo.Predicates.Latest = equalsAnyString("latest", a, b, c)
					err := verifyContainsOptions(fo)
					assert.Equal(t, ErrAtMostOnePredicate, err)
				})
			}
		}
	}

}

func TestAllExclusivePredicatesAtOnceFail(t *testing.T) {
	fo := &MatchingOptions{}
	fo.Predicates.Any = true
	fo.Predicates.Outdated = true
	fo.Predicates.Unpinned = true
	fo.Predicates.Latest = true
	err := verifyContainsOptions(fo)
	assert.Equal(t, ErrAtMostOnePredicate, err)
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
	format.OnValidateInput(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("Not my department"))

	formatProvider.OnFormats().Return([]dockfmt.Format{format})

	mainOptions.formatProvider = formatProvider

	fo := &MatchingOptions{
		mainOpts: mainOptions.mainOptions,
	}

	fo.Predicates.Any = true
	fo.Positional.InputFile = flags.Filename(NotADockerfile)

	// when
	_, err := fo.match()

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
	expected := errors.New("Process Error")
	format.OnProcess(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expected)

	formatProvider.OnFormats().Return([]dockfmt.Format{format})

	mainOptions.formatProvider = formatProvider

	fo := &MatchingOptions{
		mainOpts: mainOptions.mainOptions,
	}

	fo.Predicates.Any = true
	fo.Positional.InputFile = flags.Filename(NotADockerfile)

	// when
	_, err := fo.match()

	s := stdout.String()

	// then
	assert.Contains(t, s, `level=error`)
	assert.Contains(t, s, expected.Error())
	// and: no error is returned
	assert.Nil(t, err)
}

func TestNoPredicateForNoFlagsWithContains(t *testing.T) {
	fo := &MatchingOptions{}

	predicate := fo.getPredicate()

	assert.Nil(t, predicate)
}

func TestAnyPredicateWhenAnyFlagWithContains(t *testing.T) {
	fo := &MatchingOptions{}
	fo.Predicates.Any = true

	predicate := fo.getPredicate()

	expected := reflect.TypeOf(dockproc.AnyPredicateNew())
	actual := reflect.TypeOf(predicate)
	assert.Equal(t, expected, actual)
}

func TestLatestPredicateWhenLatestFlagWithContains(t *testing.T) {
	fo := &MatchingOptions{}
	fo.Predicates.Latest = true

	predicate := fo.getPredicate()

	expected := reflect.TypeOf(dockproc.LatestPredicateNew())
	actual := reflect.TypeOf(predicate)
	assert.Equal(t, expected, actual)
}

func TestUnpinnedPredicateWhenLatestFlagWithContains(t *testing.T) {
	fo := &MatchingOptions{}
	fo.Predicates.Unpinned = true

	predicate := fo.getPredicate()

	expected := reflect.TypeOf(dockproc.UnpinnedPredicateNew())
	actual := reflect.TypeOf(predicate)
	assert.Equal(t, expected, actual)
}

func TestFilenameRequiredWithContains(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{"contains"}, addContainsCommand)
	assert.NotEqual(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "level=error")
	assert.Contains(t, stdout.String(), "the required argument `InputFile` was not provided")
}

func TestContainsCallsFindExecuteWithContains(t *testing.T) {
	cmd, _, _, _ := testMain([]string{"contains", "fileName"}, addContainsCommand)

	_, ok := cmd.(*MatchingOptions)
	assert.True(t, ok)
}

func TestOpenErrorsArePropagatedWithContains(t *testing.T) {
	fo := containsOptionsTestNew()
	fo.Predicates.Latest = true
	expectedError := errors.New("Could not open")
	fo.MainOptions().openerMock.On("Open", mock.Anything).Return(nil, expectedError)

	exitCode, err := fo.match()

	assert.NotEqual(t, 0, exitCode)
	assert.Equal(t, expectedError, err)
}

func TestExecuteReturnsErrorWithContains(t *testing.T) {
	fo := containsOptionsTestNew()
	expected := "Use ExecuteWithExitCode instead"
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

func TestContainsHelpContainsImplementedPredicates(t *testing.T) {

	os.Args = []string{"exe", "contains", "--help"}

	mainOptions := mainOptionsACNew(addContainsCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	assert.Contains(t, buffer.String(), "--any")
	assert.Contains(t, buffer.String(), "--latest")
	assert.Contains(t, buffer.String(), "--unpinned")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestFindHelpHidesUnimplementedPredicates(t *testing.T) {

	os.Args = []string{"exe", "contains", "--help"}

	mainOptions := mainOptionsACNew(addContainsCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	assert.NotContains(t, buffer.String(), "--outdated")
	assert.NotContains(t, buffer.String(), "--name")
	assert.NotContains(t, buffer.String(), "--domain")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestContainsCommandDoesntPrint(t *testing.T) {
	test := containsOptionsTestNew()
	stdout := test.MainOptions().Stdout()

	test.Predicates.Any = true

	processorMock := &FormatProcessorMock{}
	processorMock.process = func(imageNameProcessor dockfmt.ImageNameProcessor) error {
		r, _ := dockref.FromOriginal("nginx")
		imageNameProcessor(r)
		r, _ = dockref.FromOriginal("nginx:latest")
		imageNameProcessor(r)
		r, _ = dockref.FromOriginal("nginx:1.2")
		imageNameProcessor(r)
		return nil
	}

	test.matchFormatProcessor(processorMock)

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
