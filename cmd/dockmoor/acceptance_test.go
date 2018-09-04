package main

import (
	"testing"
	"os"
	"io/ioutil"
	"log"
	"github.com/stretchr/testify/assert"
	"bytes"
	"github.com/jessevdk/go-flags"
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

	assert.Equal(t, 0, exitCode)
}

func TestFindAnyNoMatch(t *testing.T) {
	df1 := dockerfile(`invalid`)
	defer os.Remove(df1)

	os.Args = []string {"exe", "find", "--any", df1}
	mainOptions := MainOptionsTestNew(addFindCommand)
	exitCode := doMain(mainOptions)

	assert.NotEqual(t, 0, exitCode)
}

func TestFindInvalidOptions(t *testing.T) {
	df1 := dockerfile(`FROM nginx`)
	defer os.Remove(df1)

	os.Args = []string {"exe", "find", "--any", "--latest", df1}

	mainOptions := MainOptionsTestNew(addFindCommand)
	exitCode := doMain(mainOptions)

	assert.NotEqual(t, 0, exitCode)
}

func TestMainVersion(t *testing.T) {

	os.Args = []string {"exe", "--version"}

	mainOptions := MainOptionsTestNew(addFindCommand)
	exitCode := doMain(mainOptions)

	assert.Equal(t, 0, exitCode)
}

func TestMainManpage(t *testing.T) {

	os.Args = []string {"exe", "--manpage"}
	mainOptions := MainOptionsTestNew(addFindCommand)
	exitCode := doMain(mainOptions)

	assert.Equal(t, 0, exitCode)
}

func TestMainMarkdown(t *testing.T) {

	os.Args = []string {"exe", "--markdown"}

	mainOptions := MainOptionsTestNew()
	exitCode := doMain(mainOptions)


	assert.Equal(t, 0, exitCode)
}

func TestMainLoglevelNone(t *testing.T) {

	os.Args = []string {"exe", "--log-level=NONE", "find", "--any", "/notExistingFile"}
	buffer := bytes.NewBuffer(nil)
	mainOptions := MainOptionsTestNew(addFindCommand)
	exitCode := doMain(mainOptions)

	assert.Empty(t, buffer.String())
	assert.NotEqual(t, 0, exitCode)
}
