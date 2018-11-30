package main

import (
	"bytes"
	"github.com/MeneDev/dockmoor/dockfmt"
	"github.com/MeneDev/dockmoor/dockproc"
	"github.com/MeneDev/dockmoor/dockref"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"os"
	"testing"
)

type listOptionsTest struct {
	*listOptions

	mainOptionsTest *mainOptionsTest
}

func listOptionsTestNew() *listOptionsTest {
	mainOptions := mainOptionsTestNew()
	lot := listOptionsTest{
		listOptions:     listOptionsNew(mainOptions.mainOptions),
		mainOptionsTest: mainOptions,
	}

	lot.mainOpts = mainOptions.mainOptions

	return &lot
}

func (lo *listOptionsTest) MainOptions() *mainOptionsTest {
	return lo.mainOptionsTest
}

func TestFilenameRequiredWithList(t *testing.T) {
	_, _, exitCode, stdout := testMain([]string{"list"}, addListCommand)
	assert.NotEqual(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "level=error")
	assert.Contains(t, stdout.String(), "the required argument `InputFile` was not provided")
}

func TestListCallsFindExecute(t *testing.T) {
	cmd, _, _, _ := testMain([]string{"list", "fileName"}, addListCommand)

	_, ok := cmd.(*listOptions)
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
	ran := false
	processorMock.process = func(imageNameProcessor dockfmt.ImageNameProcessor) error {
		ran = true
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

	assert.True(t, ran)
	s := stdout.String()
	assert.Contains(t, s, "nginx")
	assert.Contains(t, s, "nginx:latest")
	assert.Contains(t, s, "nginx:1.2")
}

func TestListOptions_ExecuteWithExitCode(t *testing.T) {
	lo := listOptionsTestNew()
	lo.NamePredicates.Names = []string{"/a(b/"}

	predicate, e := lo.getPredicate()

	assert.Error(t, e)
	assert.Nil(t, predicate)

	exitCode, err := lo.ExecuteWithExitCode(nil)

	assert.Error(t, err)
	assert.Equal(t, ExitPredicateInvalid, exitCode)
}
