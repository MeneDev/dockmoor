package main

import (
	"bytes"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"os"
	"testing"
)

func listOptionsTestNew() *containsOptionsTest {
	mainOptions := mainOptionsTestNew()
	containsOptions := containsOptionsTest{
		MatchingOptions: &MatchingOptions{},
		mainOptionsTest: mainOptions,
	}

	containsOptions.mainOpts = mainOptions.mainOptions
	containsOptions.mode = matchAndPrint

	return &containsOptions
}

func TestFilenameRequiredWithList(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{"list"}, addListCommand)
	assert.NotEqual(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "level=error")
	assert.Contains(t, stdout.String(), "the required argument `InputFile` was not provided")
}

func TestListCallsFindExecute(t *testing.T) {
	cmd, _, _, _ := testMain([]string{"list", "fileName"}, addListCommand)

	_, ok := cmd.(*MatchingOptions)
	assert.True(t, ok)
}

func TestMainMarkdownWithList(t *testing.T) {

	os.Args = []string{"exe", "--markdown"}

	mainOptions := mainOptionsACNew(addListCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	assert.Contains(t, buffer.String(), "list command")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestMainAsciiDocWithList(t *testing.T) {

	os.Args = []string{"exe", "--asciidoc-usage"}

	mainOptions := mainOptionsACNew(addListCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	assert.Contains(t, buffer.String(), "list command")

	assert.Equal(t, ExitSuccess, exitCode)
}

func TestListHelpIsNotAnError(t *testing.T) {

	os.Args = []string{"exe", "list", "--help"}

	mainOptions := mainOptionsACNew(addListCommand)
	buffer := bytes.NewBuffer(nil)
	mainOptions.SetStdout(buffer)
	exitCode := doMain(mainOptions)

	assert.Contains(t, buffer.String(), "list command")

	assert.Equal(t, ExitSuccess, exitCode)
}

var _ dockfmt.FormatProcessor = (*FormatProcessorMock)(nil)

type FormatProcessorMock struct {
	*mock.Mock

	process func(imageNameProcessor dockfmt.ImageNameProcessor) error
}

func (d *FormatProcessorMock) WithWriter(writer io.Writer) dockfmt.FormatProcessor {
	panic("implement me")
}

func (d *FormatProcessorMock) Process(imageNameProcessor dockfmt.ImageNameProcessor) error {
	return d.process(imageNameProcessor)
}

func TestListCommandPrints(t *testing.T) {
	test := listOptionsTestNew()
	stdout := test.MainOptions().Stdout()

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

	test.matchAndProcessFormatProcessor(processorMock)

	s := stdout.String()
	assert.Contains(t, s, "nginx")
	assert.Contains(t, s, "nginx:latest")
	assert.Contains(t, s, "nginx:1.2")
}
