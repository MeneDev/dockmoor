package dockfmt

import (
	"bytes"
	"io"

	"github.com/MeneDev/dockmoor/dockref"
	"github.com/sirupsen/logrus"
)

type Format interface {
	Name() string
	ValidateInput(log logrus.FieldLogger, reader io.Reader, filename string) error
	Process(log logrus.FieldLogger, reader io.Reader, writer io.Writer, imageNameProcessor ImageNameProcessor) error
}
type ImageNameProcessor func(r dockref.Reference) (dockref.Reference, error)

type FormatProcessor interface {
	Process(imageNameProcessor ImageNameProcessor) error
	WithWriter(writer io.Writer) FormatProcessor
}

var _ FormatProcessor = (*formatProcessor)(nil)

type formatProcessor struct {
	format Format
	log    logrus.FieldLogger
	reader io.Reader
	writer io.Writer
}

func (fp *formatProcessor) Process(imageNameProcessor ImageNameProcessor) error {
	return fp.format.Process(fp.log, fp.reader, fp.writer, imageNameProcessor)
}

func (fp *formatProcessor) WithWriter(writer io.Writer) FormatProcessor {
	fp.writer = writer
	return fp
}

func FormatProcessorNew(format Format,
	log logrus.FieldLogger,
	reader io.Reader) FormatProcessor {
	return &formatProcessor{
		format: format,
		log:    log,
		reader: reader,
		writer: bytes.NewBuffer(nil),
	}
}

type FormatError struct {
	reason error
}

func (e FormatError) Error() string {
	return "FormatError: " + e.reason.Error()
}

func FormatErrorNew(err error) FormatError {
	return FormatError{
		reason: err,
	}
}
