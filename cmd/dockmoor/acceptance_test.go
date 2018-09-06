package main

import (
	"testing"
	"os"
	"io/ioutil"
	"log"
	"github.com/stretchr/testify/assert"
	"bytes"
	"github.com/jessevdk/go-flags"
	"strings"
	"github.com/mattn/go-shellwords"
	"path/filepath"
	"html/template"
)

func dockerfile(content string) (fileName string) {
	contentBytes := []byte(content)
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		log.Fatal(err)
	}


	if _, err := tmpfile.Write(contentBytes); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	return tmpfile.Name()
}

func MainOptionsTestNew(commandAdders ...func(mainOptions *mainOptions) (*flags.Command, error)) *mainOptions {
	mainOptions := MainOptionsNew()

	mainOptions.SetStdout(bytes.NewBuffer(nil))

	for _, adder := range commandAdders {
		adder(mainOptions)
	}

	return mainOptions
}

func TestFindAnyMatches(t *testing.T) {
	df1 := dockerfile(`FROM nginx`)
	defer os.Remove(df1)

	os.Args = []string {"exe", "find", "--any", df1}
	mainOptions := MainOptionsTestNew(addFindCommand)

	exitCode := doMain(mainOptions)

	assert.Equal(t, EXIT_SUCCESS, exitCode)
}

func TestFindAnyNoMatch(t *testing.T) {
	df1 := dockerfile(`invalid`)
	defer os.Remove(df1)

	os.Args = []string {"exe", "find", "--any", df1}
	mainOptions := MainOptionsTestNew(addFindCommand)
	exitCode := doMain(mainOptions)

	assert.NotEqual(t, EXIT_SUCCESS, exitCode)
}

func TestFindInvalidOptions(t *testing.T) {
	df1 := dockerfile(`FROM nginx`)
	defer os.Remove(df1)

	os.Args = []string {"exe", "find", "--any", "--latest", df1}

	mainOptions := MainOptionsTestNew(addFindCommand)
	exitCode := doMain(mainOptions)

	assert.NotEqual(t, EXIT_SUCCESS, exitCode)
}

func TestMainVersion(t *testing.T) {

	os.Args = []string {"exe", "--version"}

	mainOptions := MainOptionsTestNew(addFindCommand)
	exitCode := doMain(mainOptions)

	assert.Equal(t, EXIT_SUCCESS, exitCode)
}

func TestMainManpage(t *testing.T) {

	os.Args = []string {"exe", "--manpage"}
	mainOptions := MainOptionsTestNew(addFindCommand)
	exitCode := doMain(mainOptions)

	assert.Equal(t, EXIT_SUCCESS, exitCode)
}

func TestMainMarkdown(t *testing.T) {

	os.Args = []string {"exe", "--markdown"}

	mainOptions := MainOptionsTestNew()
	exitCode := doMain(mainOptions)


	assert.Equal(t, EXIT_SUCCESS, exitCode)
}

func TestMainLoglevelNone(t *testing.T) {

	os.Args = []string {"exe", "--log-level=NONE", "find", "--any", "/notExistingFile"}
	buffer := bytes.NewBuffer(nil)
	mainOptions := MainOptionsTestNew(addFindCommand)
	exitCode := doMain(mainOptions)

	assert.Empty(t, buffer.String())
	assert.NotEqual(t, EXIT_SUCCESS, exitCode)
}

func TestExitCodeIsZeroForDockerfile(t *testing.T) {
	dir, _ := ioutil.TempDir("", "dockmoor")
	defer os.RemoveAll(dir)

	tmpfn := filepath.Join(dir, "Dockerfile")
	dockerfile :=
		`FROM nginx`

	if err := ioutil.WriteFile(tmpfn, []byte(dockerfile), 0666); err != nil {
		log.Fatal(err)
	}

	stdout, code := shell(t, `dockmoor find --any {{.Dockerfile}}`, struct {
		Dockerfile string
	}{tmpfn})

	assert.Empty(t, stdout)
	assert.Equal(t, EXIT_SUCCESS, code, "Exits with code 0")
}

func TestExitCodeIsZeroForInvalidDockerfile(t *testing.T) {
	dir, _ := ioutil.TempDir("", "dockmoor")
	defer os.RemoveAll(dir)

	tmpfn := filepath.Join(dir, "Dockerfile")
	dockerfile :=
		`Not from nginx`

	if err := ioutil.WriteFile(tmpfn, []byte(dockerfile), 0666); err != nil {
		log.Fatal(err)
	}

	stdout, code := shell(t, `dockmoor find --any {{.Dockerfile}}`, struct {
		Dockerfile string
	}{tmpfn})

	assert.Empty(t, stdout)
	assert.Equal(t, EXIT_INVALID_FORMAT, code, "Exits with code 1")
}

func shell(t *testing.T, argsLine string, values interface{}) (stdout string, exitCode ExitCode) {
	tpl, _ := template.New("name").Parse(argsLine)
	shellBuf := bytes.NewBuffer(nil)
	tpl.Execute(shellBuf, values)

	argsLine = shellBuf.String()
	args, _ := shellwords.Parse(argsLine)
	exitCodeSet := false
	oldOsExit := osExit
	osExit = func(code ExitCode) {
		exitCode = code
		exitCodeSet = true
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
		osExit = oldOsExit
		os.Args = oldArgs
	}()

	main()

	buffer := bytes.NewBuffer(nil)
	buffer.ReadFrom(stdoutBuf)
	stdout = buffer.String()

	assert.True(t, exitCodeSet, "Expected exitCode to be set (no call to osExit)")
	return
}