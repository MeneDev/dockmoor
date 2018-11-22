package main

import (
	"bytes"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var NotADockerfile = "notDocker"

type mainOptionsTest struct {
	*mainOptions
	openerMock *ReadableOpenerMock
}

func mainOptionsTestNew() *mainOptionsTest {
	mopts := mainOptionsNew()
	mainOptions := mainOptionsTest{mainOptions: mopts}

	stdout := bytes.NewBuffer(nil)

	mainOptions.log = logrus.New()
	mainOptions.SetStdout(stdout)

	mainOptions.formatProvider = new(FormatProviderMock)

	mainOptions.openerMock = new(ReadableOpenerMock)
	mainOptions.openerMock.On("Open", NotADockerfile).Return(makeReadCloser("not a dockerfile"), nil)

	mainOptions.readableOpener = func(s string) (io.ReadCloser, error) {
		return mainOptions.openerMock.Open(s)
	}

	return &mainOptions
}

func (options *mainOptionsTest) FormatProvider() *FormatProviderMock {
	return options.formatProvider.(*FormatProviderMock)
}
func (options *mainOptionsTest) Stdout() *bytes.Buffer {
	return options.stdout.(*bytes.Buffer)
}

func testMain(args []string, registerOptions ...func(mainOptions *mainOptions, adder func(opts *mainOptions, command string, shortDescription string, longDescription string, data interface{}) (*flags.Command, error)) (*flags.Command, error)) (theCommand flags.Commander, cmdArgs []string, exitCode ExitCode, buffer *bytes.Buffer) {
	mainOptions := mainOptionsTestNew()

	for _, reg := range registerOptions {
		reg(mainOptions.mainOptions, AddCommand)
	}

	cmd, args, exitCode := CommandFromArgs(mainOptions.mainOptions, args)

	return cmd, args, exitCode, mainOptions.Stdout()
}

func TestNoCommandKnownIsError(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{})
	s := stdout.String()
	assert.NotEqual(t, 0, exitCode)
	assert.Contains(t, s, "level=error")
	assert.Contains(t, s, "No Command registered")
}

func TestHelpIsNotError(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{"--help"})
	s := stdout.String()
	assert.Equal(t, ExitSuccess, exitCode)
	assert.NotContains(t, s, "level=error")
	assert.Contains(t, s, "Usage")
	assert.Contains(t, s, "Application Options")
	assert.Contains(t, s, "Help Options")
}

func TestManIsNotError(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{"--manpage"})
	s := stdout.String()
	assert.Equal(t, ExitSuccess, exitCode)
	assert.NotContains(t, s, "level=error")
	assert.Contains(t, s, "NAME")
	assert.Contains(t, s, "SYNOPSIS")
	assert.Contains(t, s, "OPTIONS")
}

func TestMarkdownIsNotError(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{"--markdown"})
	s := stdout.String()
	assert.Equal(t, ExitSuccess, exitCode)
	assert.NotEmpty(t, s)
}

func TestOpensStdin(t *testing.T) {

	optionsTest := mainOptionsTestNew()
	opener := defaultReadableOpener(optionsTest.mainOptions)

	readCloser, e := opener("-")

	assert.Nil(t, e)
	assert.Equal(t, optionsTest.stdin, readCloser)
}

var _ io.Writer = (*failingWriter)(nil)

type failingWriter struct {
}

func (failingWriter) Write(p []byte) (n int, err error) {
	return 0, errors.Errorf("Error")
}

func Test_fmtFprintf_reportsWriteErrors(t *testing.T) {
	logger := logrus.New()
	buffer := bytes.NewBuffer(nil)
	logger.SetOutput(buffer)

	fmtFprintf(logger, failingWriter{}, "foo")

	assert.Contains(t, buffer.String(), "level=error")
}

func TestMainReportsAddingListCommandErrors(t *testing.T) {
	org := AddCommand
	defer func() {
		AddCommand = org
	}()

	AddCommand = func(opts *mainOptions, command string, shortDescription string, longDescription string, data interface{}) (*flags.Command, error) {
		return nil, errors.Errorf("A Panda, rolling around")
	}

	args := []string{"exe"}
	oldOsExit := osExitInternal
	osExitInternal = func(code int) {
		// don't actually exit
	}

	oldStdout := osStdout
	stdoutBuf := bytes.NewBuffer(nil)
	osStdout = stdoutBuf
	oldStdin := os.Stdin
	osStdin = ioutil.NopCloser(strings.NewReader(""))
	oldArgs := os.Args
	os.Args = args

	defer func() {
		osStdout = oldStdout
		osStdin = oldStdin
		osExitInternal = oldOsExit
		os.Args = oldArgs
	}()

	main()

	assert.Contains(t, stdoutBuf.String(), "level=error")
	assert.Contains(t, stdoutBuf.String(), "Could not add list command")
	assert.Contains(t, stdoutBuf.String(), "Could not add contains command")
}

func TestInvalidFlagIsReportedByName(t *testing.T) {
	_, _, exitCode, buf := testMain([]string{"--myInvalidFlag", "pin", "fileName"}, addPinCommand)

	s := buf.String()
	assert.Equal(t, ExitInvalidParams, exitCode)
	assert.Contains(t, s, "level=error")
	assert.Contains(t, s, "myInvalidFlag")
}

func TestInvalidSolverIsNil(t *testing.T) {
	cmd, _, _, _ := testMain([]string{"--resolver", "dockerd", "pin", "fileName"}, addPinCommand)

	po, _ := cmd.(*pinOptions)

	po.mainOptions().Resolver = "Invalid"

	solver := po.mainOptions().repositoryFactory()()
	assert.Nil(t, solver)
}

func TestUsesDockerdSolver(t *testing.T) {
	cmd, _, _, _ := testMain([]string{"--resolver", "dockerd", "pin", "fileName"}, addPinCommand)

	po, _ := cmd.(*pinOptions)
	assert.Equal(t, po.mainOptions().Resolver, "dockerd")
	assert.IsType(t, dockref.DockerDaemonRepositoryNew(), po.mainOptions().repositoryFactory()())
}
