package dockfmt

import (
	"io"
	"github.com/sirupsen/logrus"
	"github.com/MeneDev/dockfix/dockref"
	"bytes"
)

type Format interface {
	Name() string
	ValidateInput(log logrus.FieldLogger, reader io.Reader, filename string) error
	Process(log logrus.FieldLogger, reader io.Reader, writer io.Writer, imageNameProcessor ImageNameProcessor) error
}
type ImageNameProcessor func(r dockref.Reference) (string, error)

type FormatProcessor interface {
	Process(imageNameProcessor ImageNameProcessor) error
}

var _ FormatProcessor = (*formatProcessor)(nil)
type formatProcessor struct {
	format             Format
	log                logrus.FieldLogger
	reader             io.Reader
	writer             io.Writer
}

func (fp *formatProcessor) Process(imageNameProcessor ImageNameProcessor) error {
	return fp.format.Process(fp.log, fp.reader, fp.writer, imageNameProcessor)
}

func (fp *formatProcessor) WithWriter(writer io.Writer) *formatProcessor {
	fp.writer = writer
	return fp
}

func FormatProcessorNew(format Format,
	log logrus.FieldLogger,
	reader io.Reader) *formatProcessor {
	return &formatProcessor{
		format:             format,
		log:                log,
		reader:             reader,
		writer:             bytes.NewBuffer(nil),
	}
}
