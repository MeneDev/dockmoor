package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"io"
	"github.com/stretchr/testify/mock"
	"fmt"
	"bytes"
	"github.com/MeneDev/dockfix/dockfmt"
	"github.com/MeneDev/dockfix/dockproc"
	"reflect"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
)

type findOptionsTest struct {
	*FindOptions

	mainOptionsTest *mainOptionsTest
}

func (fo *findOptionsTest) MainOptions() *mainOptionsTest {
	return fo.mainOptionsTest
}

func FindOptionsTest() *findOptionsTest {
	mainOptions := MainOptionsTest()
	findOptions := findOptionsTest{
		FindOptions: &FindOptions{},
		mainOptionsTest: mainOptions,
	}

	findOptions.mainOptions = mainOptions.mainOptions

	return &findOptions
}

func TestEmptyPredicates(t *testing.T) {
	fo := &FindOptions{}
	err := verifyFindOptions(fo)
	assert.Equal(t, ERR_AT_LEAST_ONE_PREDICATE, err)
}

func TestSingleExclusivePredicatesFail(t *testing.T) {
	strings := []string{"any", "latest", "unpinned", "outdated"}
	for _, a := range strings {
		t.Run(a, func(t *testing.T) {
			fo := &FindOptions{}
			fo.Predicates.Any = equalsAnyString("any", a)
			fo.Predicates.Outdated = equalsAnyString("outdated", a)
			fo.Predicates.Unpinned = equalsAnyString("unpinned", a)
			fo.Predicates.Latest = equalsAnyString("latest", a)
			err := verifyFindOptions(fo)
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
				fo := &FindOptions{}
				fo.Predicates.Any = equalsAnyString("any", a, b)
				fo.Predicates.Outdated = equalsAnyString("outdated", a, b)
				fo.Predicates.Unpinned = equalsAnyString("unpinned", a, b)
				fo.Predicates.Latest = equalsAnyString("latest", a, b)
				err := verifyFindOptions(fo)
				assert.Equal(t, ERR_AT_MOST_ONE_PREDICATE, err)
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
					fo := &FindOptions{}
					fo.Predicates.Any = equalsAnyString("any", a, b, c)
					fo.Predicates.Outdated = equalsAnyString("outdated", a, b, c)
					fo.Predicates.Unpinned = equalsAnyString("unpinned", a, b, c)
					fo.Predicates.Latest = equalsAnyString("latest", a, b, c)
					err := verifyFindOptions(fo)
					assert.Equal(t, ERR_AT_MOST_ONE_PREDICATE, err)
				})
			}
		}
	}

}

func TestAllExclusivePredicatesAtOnceFail(t *testing.T) {
	fo := &FindOptions{}
	fo.Predicates.Any = true
	fo.Predicates.Outdated = true
	fo.Predicates.Unpinned = true
	fo.Predicates.Latest = true
	err := verifyFindOptions(fo)
	assert.Equal(t, ERR_AT_MOST_ONE_PREDICATE, err)
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

func TestInvalidDockerfile(t *testing.T) {
	// given
	mainOptions := MainOptionsTest()

	formatProvider := mainOptions.FormatProvider()

	format := new(FormatMock)
	format.OnName().Return("mock")
	format.OnValidateInput(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("Not my department"))

	formatProvider.OnFormats().Return([]dockfmt.Format{format})

	mainOptions.formatProvider = formatProvider

	fo := &FindOptions{
		mainOptions: mainOptions.mainOptions,
	}

	fo.Predicates.Any = true
	fo.Positional.InputFile = flags.Filename(NotADockerfile)

	// when
	_, err := fo.find()

	// then
	assert.NotNil(t, err)

	_, ok := err.(dockfmt.UnknownFormatError)
	assert.True(t, ok)
}

func TestNoPredicateForNoFlags(t *testing.T) {
	fo := &FindOptions{}

	predicate := fo.getPredicate()

	assert.Nil(t, predicate)
}

func TestAnyPredicateWhenAnyFlag(t *testing.T) {
	fo := &FindOptions{}
	fo.Predicates.Any = true

	predicate := fo.getPredicate()

	expected := reflect.TypeOf(dockproc.AnyPredicateNew())
	actual := reflect.TypeOf(predicate)
	assert.Equal(t, expected, actual)
}

func TestLatestPredicateWhenLatestFlag(t *testing.T) {
	t.Skip()
	fo := &FindOptions{}
	fo.Predicates.Latest = true

	predicate := fo.getPredicate()

	expected := reflect.TypeOf(dockproc.AnyPredicateNew())
	actual := reflect.TypeOf(predicate)
	assert.Equal(t, expected, actual)
}

func TestFilenameRequired(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{"find"}, addFindCommand)
	assert.NotEqual(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "level=error")
	assert.Contains(t, stdout.String(), "the required argument `InputFile` was not provided")
}


func TestFindCallsFindExecute(t *testing.T) {
	cmd, _, _, _ := testMain([]string{"find", "fileName"}, addFindCommand)

	_, ok := cmd.(*FindOptions)
	assert.True(t, ok)
}

func TestOpenErrorsArePropagated(t *testing.T) {
	fo := FindOptionsTest()
	fo.Predicates.Latest = true
	expectedError := errors.New("Could not open")
	fo.MainOptions().openerMock.On("Open", mock.Anything).Return(nil, expectedError)

	exitCode, err := fo.find()

	assert.NotEqual(t, 0, exitCode)
	assert.Equal(t, expectedError, err)
}

func TestExecuteReturnsError(t *testing.T) {
	fo := FindOptionsTest()
	expected := "Use ExecuteWithExitCode instead"
	err := fo.Execute(nil)

	assert.Equal(t, expected, err.Error())
}

func TestMainMarkdownWithFind(t *testing.T) {

	os.Args = []string {"exe", "--markdown"}

	mainOptions := MainOptionsTestNew(addFindCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	assert.Contains(t, buffer.String(), "find command")

	assert.Equal(t, 0, exitCode)
}

func TestFindHelpIsNotAnError(t *testing.T) {

	os.Args = []string {"exe", "find", "--help"}

	mainOptions := MainOptionsTestNew(addFindCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	assert.Contains(t, buffer.String(), "find command")

	assert.Equal(t, 0, exitCode)
}

func TestFindHelpContainsAny(t *testing.T) {

	os.Args = []string {"exe", "find", "--help"}

	mainOptions := MainOptionsTestNew(addFindCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	assert.Contains(t, buffer.String(), "--any")

	assert.Equal(t, 0, exitCode)
}

func TestFindHelpHidesUnimplemented(t *testing.T) {

	os.Args = []string {"exe", "find", "--help"}

	mainOptions := MainOptionsTestNew(addFindCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	assert.NotContains(t, buffer.String(), "--latest")
	assert.NotContains(t, buffer.String(), "--unpinned")
	assert.NotContains(t, buffer.String(), "--outdated")
	assert.NotContains(t, buffer.String(), "--name")
	assert.NotContains(t, buffer.String(), "--domain")

	assert.Equal(t, 0, exitCode)
}

func equalsAnyString(needle string, values ...string) bool {
	for _, v := range values {
		if needle == v {
			return true
		}
	}

	return false
}
